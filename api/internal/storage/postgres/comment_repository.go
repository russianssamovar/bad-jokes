package postgres

import (
	"badJokes/internal/lib/sl"
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

type CommentsRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewCommentsRepository(db *sql.DB, log *slog.Logger) *CommentsRepository {
	return &CommentsRepository{
		db:  db,
		log: log.With(slog.String("component", "comments_repository")),
	}
}

func (r *CommentsRepository) AddComment(jokeID, userID int64, body string, parentID *int64) (int64, error) {
	r.log.Debug("Adding comment",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID),
		slog.String("body_length", fmt.Sprintf("%d chars", len(body))),
		slog.Int64("parent_id", func() int64 { if parentID != nil { return *parentID } else { return 0 } }()))

	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM jokes WHERE id = $1)", jokeID).Scan(&exists)
	if err != nil {
		r.log.Error("Failed to check joke existence", sl.Err(err))
		return 0, err
	}
	if !exists {
		r.log.Info("Joke not found", slog.Int64("joke_id", jokeID))
		return 0, ErrJokeNotFound
	}

	if parentID != nil {
		err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND joke_id = $2)",
			*parentID, jokeID).Scan(&exists)
		if err != nil {
			r.log.Error("Failed to check parent comment existence", sl.Err(err))
			return 0, err
		}
		if !exists {
			r.log.Info("Parent comment not found", slog.Int64("parent_id", *parentID))
			return 0, ErrCommentNotFound
		}
	}

	var id int64
	query := `
        INSERT INTO comments (joke_id, parent_id, body, user_id, created_at, modified_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id
    `
	err = r.db.QueryRow(query, jokeID, parentID, body, userID).Scan(&id)
	if err != nil {
		r.log.Error("Failed to insert comment", sl.Err(err))
		return 0, err
	}

	r.log.Info("Comment added successfully", slog.Int64("comment_id", id))
	return id, nil
}

func (r *CommentsRepository) GetComments(jokeID int64) ([]models.Comment, error) {
	r.log.Debug("Fetching comments", slog.Int64("joke_id", jokeID))

	rows, err := r.db.Query(`
		SELECT id, joke_id, user_id, body, created_at, modified_at
		FROM comments
		WHERE joke_id = $1
	`, jokeID)
	if err != nil {
		r.log.Error("Failed to fetch comments", sl.Err(err))
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
			r.log.Error("Failed to scan comment", sl.Err(err))
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error iterating comment rows", sl.Err(err))
		return nil, fmt.Errorf("error iterating comment rows: %w", err)
	}

	r.log.Debug("Comments fetched successfully", slog.Int("count", len(comments)))
	return comments, nil
}

func (r *CommentsRepository) GetCommentsByJokeID(jokeID, currentUserID int64) ([]models.Comment, error) {
	r.log.Debug("Fetching comments by joke ID",
		slog.Int64("joke_id", jokeID),
		slog.Int64("current_user_id", currentUserID))

	query := `
        SELECT 
            c.id,
            c.joke_id,
            c.parent_id,
            c.body,
            c.user_id,
            u.username AS author_username,
            c.created_at,
            c.is_deleted,
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
        JOIN users u ON c.user_id = u.id
        LEFT JOIN votes uv ON c.id = uv.entity_id AND uv.entity_type = 'comment' AND uv.user_id = $2
        WHERE c.joke_id = $3
        ORDER BY 
            CASE WHEN c.parent_id IS NULL THEN c.id ELSE c.parent_id END ASC,
            c.parent_id IS NOT NULL ASC,
            c.created_at ASC
    `

	rows, err := r.db.Query(query, currentUserID, currentUserID, jokeID)
	if err != nil {
		r.log.Error("Failed to fetch comments by joke ID", sl.Err(err))
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
			&comment.IsDeleted,
			&comment.ModifiedAt,
			&comment.Social.Pluses,
			&reactionsJSON,
			&userVote,
			&userReactions,
		); err != nil {
			r.log.Error("Failed to scan comment", sl.Err(err))
			return nil, err
		}

		if parentID.Valid {
			comment.ParentID = parentID.Int64
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

		if comment.IsDeleted {
			comment.Body = ""
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("Error iterating comment rows", sl.Err(err))
		return nil, err
	}

	r.log.Debug("Comments fetched successfully", slog.Int("count", len(comments)))
	return comments, nil
}

func (r *CommentsRepository) DeleteComment(commentID int64) error {
	r.log.Info("Deleting comment", slog.Int64("comment_id", commentID))

	result, err := r.db.Exec("UPDATE comments SET is_deleted = TRUE WHERE id = $1", commentID)
	if err != nil {
		r.log.Error("Failed to delete comment", sl.Err(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected", sl.Err(err))
		return err
	}

	if rowsAffected == 0 {
		r.log.Info("Comment not found", slog.Int64("comment_id", commentID))
		return ErrCommentNotFound
	}

	r.log.Info("Comment deleted successfully", slog.Int64("comment_id", commentID))
	return nil
}

func (r *CommentsRepository) GetCommentByID(commentID int64) (models.Comment, error) {
	r.log.Debug("Fetching comment by ID", slog.Int64("comment_id", commentID))

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
			r.log.Info("Comment not found", slog.Int64("comment_id", commentID))
			return comment, ErrCommentNotFound
		}
		r.log.Error("Failed to fetch comment by ID", sl.Err(err))
		return comment, err
	}

	if parentID.Valid {
		comment.ParentID = parentID.Int64
	}

	r.log.Debug("Comment fetched successfully", slog.Int64("comment_id", commentID))
	return comment, nil
}