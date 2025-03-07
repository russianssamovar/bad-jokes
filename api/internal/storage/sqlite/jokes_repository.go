package sqlite

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type JokesRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewJokesRepository(db *sql.DB, log *slog.Logger) *JokesRepository {
	return &JokesRepository{
		db:  db,
		log: log.With(slog.String("component", "jokes_repository")),
	}
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
            GROUP_CONCAT(DISTINCT i.type) as reactions,
            COALESCE(uv.vote_type, '') AS user_vote,
            GROUP_CONCAT(DISTINCT uiv.type) as user_reactions,
            u.username AS author_username
        FROM jokes j
        LEFT JOIN comments c ON j.id = c.joke_id
        LEFT JOIN votes uv ON j.id = uv.entity_id AND uv.entity_type = 'joke' AND uv.user_id = ?
        LEFT JOIN interactions i ON j.id = i.entity_id AND i.entity_type = 'joke'
        LEFT JOIN interactions uiv ON j.id = uiv.entity_id AND uiv.entity_type = 'joke' AND uiv.user_id = ?
        JOIN users u ON j.author_id = u.id
        WHERE j.id = ?
        GROUP BY j.id, j.body, j.author_id, j.created_at, j.modified_at, uv.vote_type, u.username
    `

	var joke models.Joke
	var reactions sql.NullString
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
		&reactions,
		&userVote,
		&userReactions,
		&authorUsername,
	)
	if err != nil {
		return joke, err
	}

	reactionMap := map[string]int{}
	if reactions.Valid && reactions.String != "" {
		for _, reaction := range strings.Split(strings.TrimSpace(reactions.String), ",") {
			if reaction != "" {
				reactionMap[reaction]++
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
	_, err := r.db.Exec("DELETE FROM jokes WHERE id = ?", jokeID)
	return err
}
