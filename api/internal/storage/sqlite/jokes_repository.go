package sqlite

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var ErrJokeNotFound = fmt.Errorf("joke not found")
var ErrCommentNotFound = fmt.Errorf("comment not found")

type JokesRepository struct {
	db *sql.DB
}

func NewJokesRepository(db *sql.DB) *JokesRepository {
	return &JokesRepository{db: db}
}

func (r *JokesRepository) Insert(body string, authorID int64) (int64, error) {
	stmt, err := r.db.Prepare("INSERT INTO jokes(body, author_id, created_at, modified_at) VALUES(?, ?, datetime('now'), datetime('now'))")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(body, authorID)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

func (r *JokesRepository) ListPage(page, pageSize int, sortField, order string, currentUserID int64) ([]models.Joke, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT j.id, j.body, j.author_id, j.created_at, j.modified_at, 
		       COUNT(DISTINCT v.id) as vote_count,
		       COUNT(DISTINCT c.id) as comment_count,
		       GROUP_CONCAT(DISTINCT i.type) as reactions,
		       COALESCE(uv.vote_type, '') as user_vote,
		       GROUP_CONCAT(DISTINCT uiv.type) as user_reactions
		FROM jokes j
		LEFT JOIN votes v ON j.id = v.entity_id AND v.entity_type = 'joke'
		LEFT JOIN comments c ON j.id = c.joke_id
		LEFT JOIN interactions i ON j.id = i.entity_id AND i.entity_type = 'joke'
		LEFT JOIN votes uv ON j.id = uv.entity_id AND uv.entity_type = 'joke' AND uv.user_id = ?
		LEFT JOIN interactions uiv ON j.id = uiv.entity_id AND uiv.entity_type = 'joke' AND uiv.user_id = ?
		GROUP BY j.id, j.body, j.author_id, j.created_at, j.modified_at, uv.vote_type
		ORDER BY j.` + sortField + ` ` + order + `
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, currentUserID, currentUserID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list jokes: %w", err)
	}
	defer rows.Close()

	var jokes []models.Joke
	for rows.Next() {
		var joke models.Joke
		var reactions sql.NullString
		var userVote sql.NullString
		var userReactions sql.NullString

		if err := rows.Scan(&joke.ID, &joke.Body, &joke.AuthorID, &joke.CreatedAt, &joke.ModifiedAt,
			&joke.Social.Pluses, &joke.CommentCount, &reactions, &userVote, &userReactions); err != nil {
			return nil, fmt.Errorf("failed to scan joke: %w", err)
		}

		reactionMap := map[string]int{}
		if reactions.Valid && reactions.String != "[]" {
			for _, reaction := range strings.Split(strings.Trim(reactions.String, "[]"), ",") {
				reaction = strings.TrimSpace(reaction)
				reactionMap[reaction]++
			}
		}
		joke.Social.Reactions = reactionMap

		if userVote.Valid && userVote.String != "" {
			joke.Social.User = &models.UserInteraction{VoteType: userVote.String}
		}

		if userReactions.Valid && userReactions.String != "[]" {
			userReactionsArray := strings.Split(strings.Trim(userReactions.String, "[]"), ",")
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
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(entity_type, entity_id, user_id) DO UPDATE SET vote_type = ?, modified_at = datetime('now')`,
		entityType, entityID, userID, voteType, voteType)
	return err
}

func (r *JokesRepository) RemoveVote(entityType string, entityID, userID int64) error {
	_, err := r.db.Exec("DELETE FROM votes WHERE entity_type = ? AND entity_id = ? AND user_id = ?", entityType, entityID, userID)
	return err
}

func (r *JokesRepository) GetVote(entityType string, entityID, userID int64) (string, error) {
	var voteType sql.NullString
	err := r.db.QueryRow("SELECT vote_type FROM votes WHERE entity_type = ? AND entity_id = ? AND user_id = ?", entityType, entityID, userID).Scan(&voteType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return voteType.String, err
}

func (r *JokesRepository) AddReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec(`
		INSERT INTO interactions (entity_type, entity_id, user_id, type, created_at, modified_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(entity_type, entity_id, user_id, type) DO NOTHING`,
		entityType, entityID, userID, reactionType)
	return err
}

func (r *JokesRepository) RemoveReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec("DELETE FROM interactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND type = ?", entityType, entityID, userID, reactionType)
	return err
}

func (r *JokesRepository) GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM interactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND type = ?", entityType, entityID, userID, reactionType).Scan(&count)
	return count > 0, err
}

func (r *JokesRepository) DeleteJoke(jokeID int64) error {
	_, err := r.db.Exec("DELETE FROM jokes WHERE id = ?", jokeID)
	return err
}

func (r *JokesRepository) GetJokeByID(jokeID, currentUserID int64) (models.Joke, error) {
    query := `
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
            ) AS user_reactions,
            u.username AS author_username
        FROM jokes j
        LEFT JOIN comments c ON j.id = c.joke_id
        LEFT JOIN votes uv ON j.id = uv.entity_id AND uv.entity_type = 'joke' AND uv.user_id = $1
        JOIN users u ON j.author_id = u.id
        WHERE j.id = $3
        GROUP BY j.id, j.body, j.author_id, j.created_at, j.modified_at, uv.vote_type, u.username
    `
    
    var joke models.Joke
    var reactionsJSON sql.NullString
    var userVote sql.NullString
    var userReactions sql.NullString
    var authorUsername string
    
    err := r.db.QueryRow(query, currentUserID, currentUserID, jokeID).Scan(
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
        &authorUsername,
    )
    if err != nil {
        return joke, err
    }
    
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
    
    joke.AuthorUsername = authorUsername
    return joke, nil
}

func (r *JokesRepository) GetComments(jokeID int64) ([]models.Comment, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.joke_id, c.user_id, c.body, c.created_at, c.modified_at
		FROM comments c
		WHERE c.joke_id = ?`, jokeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.JokeID, &comment.UserID, &comment.Body, &comment.CreatedAt, &comment.ModifiedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *JokesRepository) AddComment(jokeID, userID int64, body string, parentID *int64) (int64, error) {
    var exists bool
    err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM jokes WHERE id = $1)", jokeID).Scan(&exists)
    if err != nil {
        return 0, err
    }
    if !exists {
        return 0, ErrJokeNotFound
    }

    if parentID != nil {
        err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND joke_id = $2)", 
            *parentID, jokeID).Scan(&exists)
        if err != nil {
            return 0, err
        }
        if !exists {
            return 0, ErrCommentNotFound
        }
    }

    var id int64
    query := `
        INSERT INTO comments (joke_id, parent_id, body, author_id, created_at, modified_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id
    `
    err = r.db.QueryRow(query, jokeID, parentID, body, userID).Scan(&id)
    if err != nil {
        return 0, err
    }
    
    return id, nil
}


func (r *JokesRepository) DeleteComment(commentID int64) error {
    result, err := r.db.Exec("DELETE FROM comments WHERE id = $1", commentID)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return ErrCommentNotFound
    }
    
    return nil
}

func (r *JokesRepository) GetCommentsByJokeID(jokeID, currentUserID int64) ([]models.Comment, error) {
    query := `
        SELECT 
            c.id,
            c.joke_id,
            c.parent_id,
            c.body,
            c.author_id,
            u.username AS author_username,
            c.created_at,
            c.modified_at,
            (
                SELECT COALESCE(SUM(CASE WHEN vote_type = 'plus' THEN 1 WHEN vote_type = 'minus' THEN -1 ELSE 0 END), 0)
                FROM votes WHERE entity_id = c.id AND entity_type = 'comment'
            ) AS vote_count,
            (
                SELECT json_object_agg(type, count) 
                FROM (
                    SELECT type, COUNT(*) as count 
                    FROM interactions 
                    WHERE entity_id = c.id AND entity_type = 'comment' 
                    GROUP BY type
                ) reaction_counts
            ) AS reactions_json,
            COALESCE(uv.vote_type, '') AS user_vote,
            COALESCE(
                (SELECT array_to_string(array_agg(type), ',')
                 FROM interactions 
                 WHERE entity_id = c.id AND entity_type = 'comment' AND user_id = $1), 
                ''
            ) AS user_reactions
        FROM comments c
        JOIN users u ON c.author_id = u.id
        LEFT JOIN votes uv ON c.id = uv.entity_id AND uv.entity_type = 'comment' AND uv.user_id = $2
        WHERE c.joke_id = $3
        ORDER BY 
            CASE WHEN c.parent_id IS NULL THEN c.id ELSE c.parent_id END ASC,
            c.parent_id IS NOT NULL ASC,
            c.created_at ASC
    `
    
    rows, err := r.db.Query(query, currentUserID, currentUserID, jokeID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var comments []models.Comment
    for rows.Next() {
        var comment models.Comment
        var parentID sql.NullInt64
        var reactionsJSON sql.NullString
        var userVote sql.NullString
        var userReactions sql.NullString
        
        if err := rows.Scan(
            &comment.ID,
            &comment.JokeID,
            &parentID,
            &comment.Body,
            &comment.AuthorID,
            &comment.AuthorUsername,
            &comment.CreatedAt,
            &comment.ModifiedAt,
            &comment.Social.Pluses,
            &reactionsJSON,
            &userVote,
            &userReactions,
        ); err != nil {
            return nil, err
        }
        
        if parentID.Valid {
            comment.ParentID = parentID.Int64
        }
        
        // Process reaction data
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
        comment.Social.Reactions = reactionMap
        
        if userVote.Valid && userVote.String != "" {
            comment.Social.User = &models.UserInteraction{VoteType: userVote.String}
        }
        
        if userReactions.Valid && userReactions.String != "" {
            userReactionsArray := strings.Split(userReactions.String, ",")
            for i, r := range userReactionsArray {
                userReactionsArray[i] = strings.TrimSpace(r)
            }
            if comment.Social.User == nil {
                comment.Social.User = &models.UserInteraction{}
            }
            comment.Social.User.Reactions = userReactionsArray
        }
        
        comments = append(comments, comment)
    }
    
    return comments, nil
}

func (r *JokesRepository) GetCommentByID(commentID int64) (models.Comment, error) {
	query := `
		SELECT 
			c.id, 
			c.joke_id, 
			c.parent_id, 
			c.body, 
			c.user_id,
			u.username AS author_username,
			c.created_at, 
			c.modified_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`
	
	var comment models.Comment
	var parentID sql.NullInt64
	
	err := r.db.QueryRow(query, commentID).Scan(
		&comment.ID,
		&comment.JokeID,
		&parentID,
		&comment.Body,
		&comment.AuthorID,
		&comment.AuthorUsername,
		&comment.CreatedAt,
		&comment.ModifiedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return comment, ErrCommentNotFound
		}
		return comment, err
	}
	
	if parentID.Valid {
		comment.ParentID = parentID.Int64
	}
	
	return comment, nil
}