package sqlite

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var ErrJokeNotFound = fmt.Errorf("joke not found")

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

func (r *JokesRepository) GetJokeByID(jokeID int64) (*models.Joke, error) {
	var joke models.Joke
	err := r.db.QueryRow("SELECT id, body, author_id, created_at, modified_at FROM jokes WHERE id = ?", jokeID).
		Scan(&joke.ID, &joke.Body, &joke.AuthorID, &joke.CreatedAt, &joke.ModifiedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("joke not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query joke: %w", err)
	}

	return &joke, nil
}

func (r *JokesRepository) AddComment(jokeID, userID int64, body string) (int64, error) {
	stmt, err := r.db.Prepare("INSERT INTO comments(joke_id, user_id, body, created_at) VALUES(?, ?, ?, datetime('now'))")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(jokeID, userID, body)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
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
