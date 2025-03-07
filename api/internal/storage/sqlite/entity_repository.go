package sqlite

import (
	"database/sql"
	"log/slog"
)

type EntityRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewEntityRepository(db *sql.DB, log *slog.Logger) *EntityRepository {
	return &EntityRepository{
		db:  db,
		log: log.With(slog.String("component", "entity_repository")),
	}
}

func (r *EntityRepository) AddVote(entityType string, entityID, userID int64, voteType string) error {
	_, err := r.db.Exec(`
		INSERT INTO votes (entity_type, entity_id, user_id, vote_type, created_at, modified_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(entity_type, entity_id, user_id) DO UPDATE SET vote_type = ?, modified_at = datetime('now')`,
		entityType, entityID, userID, voteType, voteType)
	return err
}

func (r *EntityRepository) RemoveVote(entityType string, entityID, userID int64) error {
	_, err := r.db.Exec("DELETE FROM votes WHERE entity_type = ? AND entity_id = ? AND user_id = ?", entityType, entityID, userID)
	return err
}

func (r *EntityRepository) GetVote(entityType string, entityID, userID int64) (string, error) {
	var voteType sql.NullString
	err := r.db.QueryRow("SELECT vote_type FROM votes WHERE entity_type = ? AND entity_id = ? AND user_id = ?", entityType, entityID, userID).Scan(&voteType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return voteType.String, err
}

func (r *EntityRepository) AddReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec(`
		INSERT INTO interactions (entity_type, entity_id, user_id, type, created_at, modified_at)
		VALUES (?, ?, ?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(entity_type, entity_id, user_id, type) DO NOTHING`,
		entityType, entityID, userID, reactionType)
	return err
}

func (r *EntityRepository) RemoveReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec("DELETE FROM interactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND type = ?", entityType, entityID, userID, reactionType)
	return err
}

func (r *EntityRepository) GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM interactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND type = ?", entityType, entityID, userID, reactionType).Scan(&count)
	return count > 0, err
}
