// api/internal/storage/postgres/joke_repository.go
package postgres

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"strings"
)

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
		query = baseQuery + " ORDER BY j.created_at DESC LIMIT $3 OFFSET $4"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list jokes: %w", err)
	}
	defer rows.Close()

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

func (r *JokesRepository) DeleteJoke(jokeID int64) error {
	_, err := r.db.Exec("DELETE FROM jokes WHERE id = $1", jokeID)
	return err
}