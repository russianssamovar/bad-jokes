package postgres

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

var ErrJokeNotFound = fmt.Errorf("joke not found")

type JokesRepository struct {
	db *sql.DB
}

func NewJokesRepository(db *sql.DB) *JokesRepository {
	return &JokesRepository{db: db}
}

func (r *JokesRepository) Insert(body string, authorID int64) (int64, error) {
	query := `
		INSERT INTO jokes (body, author_id, created_at, modified_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(query, body, authorID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert joke: %w", err)
	}
	return id, nil
}

func (r *JokesRepository) ListPage(page, pageSize int, sortField, order string, currentUserID int64) ([]models.Joke, error) {
    offset := (page - 1) * pageSize
    
    baseQuery := `
        SELECT 
            j.id,
            j.body,
            j.author_id,
            j.created_at,
            j.modified_at,
            (
                SELECT COALESCE(SUM(CASE WHEN vote_type = 'plus' THEN 1 WHEN vote_type = 'minus' THEN -1 ELSE 0 END), 0)
                FROM votes WHERE entity_id = j.id AND entity_type = 'joke'
            ) AS vote_count,
            COUNT(DISTINCT c.id) AS comment_count,
            (
                SELECT json_object_agg(type, count) 
                FROM (
                    SELECT type, COUNT(*) as count 
                    FROM interactions 
                    WHERE entity_id = j.id AND entity_type = 'joke' 
                    GROUP BY type
                ) reaction_counts
            ) AS reactions_json,
            COALESCE(uv.vote_type, '') AS user_vote,
            COALESCE(
                (SELECT array_to_string(array_agg(type), ',')
                 FROM interactions 
                 WHERE entity_id = j.id AND entity_type = 'joke' AND user_id = $2), 
                ''
            ) AS user_reactions
        FROM jokes j
        LEFT JOIN comments c ON j.id = c.joke_id
        LEFT JOIN votes uv ON j.id = uv.entity_id AND uv.entity_type = 'joke' AND uv.user_id = $1
        GROUP BY j.id, j.body, j.author_id, j.created_at, j.modified_at, uv.vote_type
    `
    
    var query string
    var args []interface{}
    args = append(args, currentUserID, currentUserID, pageSize, offset)
    
    // Safely construct the ORDER BY clause based on validated input
    switch sortField {
    case "created_at":
        if order == "asc" {
            query = baseQuery + " ORDER BY j.created_at ASC LIMIT $3 OFFSET $4"
        } else {
            query = baseQuery + " ORDER BY j.created_at DESC LIMIT $3 OFFSET $4"
        }
    case "modified_at":
        if order == "asc" {
            query = baseQuery + " ORDER BY j.modified_at ASC LIMIT $3 OFFSET $4"
        } else {
            query = baseQuery + " ORDER BY j.modified_at DESC LIMIT $3 OFFSET $4"
        }
    case "id":
        if order == "asc" {
            query = baseQuery + " ORDER BY j.id ASC LIMIT $3 OFFSET $4"
        } else {
            query = baseQuery + " ORDER BY j.id DESC LIMIT $3 OFFSET $4"
        }
    case "score":
        if order == "asc" {
            query = baseQuery + " ORDER BY vote_count ASC LIMIT $3 OFFSET $4"
        } else {
            query = baseQuery + " ORDER BY vote_count DESC LIMIT $3 OFFSET $4"
        }
    case "reactions_count":
        // Count the total number of reactions
        if order == "asc" {
            query = baseQuery + ` 
                ORDER BY (
                    SELECT COUNT(*) 
                    FROM interactions 
                    WHERE entity_id = j.id AND entity_type = 'joke'
                ) ASC LIMIT $3 OFFSET $4`
        } else {
            query = baseQuery + ` 
                ORDER BY (
                    SELECT COUNT(*) 
                    FROM interactions 
                    WHERE entity_id = j.id AND entity_type = 'joke'
                ) DESC LIMIT $3 OFFSET $4`
        }
    default:
        // Default to created_at DESC
        query = baseQuery + " ORDER BY j.created_at DESC LIMIT $3 OFFSET $4"
    }
    
    rows, err := r.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to list jokes: %w", err)
    }
    defer rows.Close()
    
    // Rest of the function remains unchanged
    var jokes []models.Joke
    for rows.Next() {
        var joke models.Joke
        var reactionsJSON sql.NullString
        var userVote sql.NullString
        var userReactions sql.NullString
        if err := rows.Scan(
            &joke.ID,
            &joke.Body,
            &joke.AuthorID,
            &joke.CreatedAt,
            &joke.ModifiedAt,
            &joke.Social.Pluses,
            &joke.CommentCount,
            &reactionsJSON,
            &userVote,
            &userReactions,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan joke: %w", err)
        }
        
        // Processing reactionsJSON, userVote, and userReactions remains unchanged
        reactionMap := map[string]int{}
        if reactionsJSON.Valid && reactionsJSON.String != "" && reactionsJSON.String != "null" {
            jsonStr := strings.Trim(reactionsJSON.String, "{}")
            if jsonStr != "" {
                pairs := strings.Split(jsonStr, ",")
                for _, pair := range pairs {
                    kv := strings.Split(pair, ":")
                    if len(kv) == 2 {
                        key := strings.Trim(kv[0], "\" ")
                        val := strings.Trim(kv[1], " ")
                        count := 0
                        fmt.Sscanf(val, "%d", &count)
                        reactionMap[key] = count
                    }
                }
            }
        }
        joke.Social.Reactions = reactionMap
        
        if userVote.Valid && userVote.String != "" {
            joke.Social.User = &models.UserInteraction{VoteType: userVote.String}
        }
        
        if userReactions.Valid && userReactions.String != "" {
            userReactionsArray := strings.Split(userReactions.String, ",")
            for i, r := range userReactionsArray {
                userReactionsArray[i] = strings.TrimSpace(r)
            }
            if joke.Social.User == nil {
                joke.Social.User = &models.UserInteraction{}
            }
            joke.Social.User.Reactions = userReactionsArray
        }
        jokes = append(jokes, joke)
    }
    
    return jokes, nil
}

func (r *JokesRepository) AddVote(entityType string, entityID, userID int64, voteType string) error {
	_, err := r.db.Exec(`
		INSERT INTO votes (entity_type, entity_id, user_id, vote_type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id) DO UPDATE SET vote_type = $5, modified_at = NOW()`,
		entityType, entityID, userID, voteType, voteType)
	return err
}

func (r *JokesRepository) RemoveVote(entityType string, entityID, userID int64) error {
	_, err := r.db.Exec("DELETE FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", entityType, entityID, userID)
	return err
}

func (r *JokesRepository) GetVote(entityType string, entityID, userID int64) (string, error) {
	var voteType sql.NullString
	err := r.db.QueryRow("SELECT vote_type FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", entityType, entityID, userID).Scan(&voteType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return voteType.String, err
}

func (r *JokesRepository) AddReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec(`
		INSERT INTO interactions (entity_type, entity_id, user_id, type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id, type) DO NOTHING`,
		entityType, entityID, userID, reactionType)
	return err
}

func (r *JokesRepository) RemoveReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec("DELETE FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", entityType, entityID, userID, reactionType)
	return err
}

func (r *JokesRepository) GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", entityType, entityID, userID, reactionType).Scan(&count)
	return count > 0, err
}

func (r *JokesRepository) DeleteJoke(jokeID int64) error {
	_, err := r.db.Exec("DELETE FROM jokes WHERE id = $1", jokeID)
	return err
}

func (r *JokesRepository) GetJokeByID(jokeID int64) (*models.Joke, error) {
	var joke models.Joke
	err := r.db.QueryRow("SELECT id, body, author_id, created_at, modified_at FROM jokes WHERE id = $1", jokeID).
		Scan(&joke.ID, &joke.Body, &joke.AuthorID, &joke.CreatedAt, &joke.ModifiedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("joke not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query joke: %w", err)
	}

	return &joke, nil
}

func (r *JokesRepository) AddComment(jokeID, userID int64, body string) (int64, error) {
	query := `
		INSERT INTO comments (joke_id, user_id, body, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(query, jokeID, userID, body).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert comment: %w", err)
	}
	return id, nil
}

func (r *JokesRepository) GetComments(jokeID int64) ([]models.Comment, error) {
	rows, err := r.db.Query(`
		SELECT id, joke_id, user_id, body, created_at, modified_at
		FROM comments
		WHERE joke_id = $1
	`, jokeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.JokeID,
			&comment.UserID,
			&comment.Body,
			&comment.CreatedAt,
			&comment.ModifiedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
