package postgres

import (
	"database/sql"
)

type EntityRepository struct {
	db *sql.DB
}

func NewEntityRepository(db *sql.DB) *EntityRepository {
	return &EntityRepository{db: db}
}

func (r *EntityRepository) AddVote(entityType string, entityID, userID int64, voteType string) error {
	_, err := r.db.Exec(`
		INSERT INTO votes (entity_type, entity_id, user_id, vote_type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id) DO UPDATE SET vote_type = $5, modified_at = NOW()`,
		entityType, entityID, userID, voteType, voteType)
	return err
}

func (r *EntityRepository) RemoveVote(entityType string, entityID, userID int64) error {
	_, err := r.db.Exec("DELETE FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", entityType, entityID, userID)
	return err
}

func (r *EntityRepository) GetVote(entityType string, entityID, userID int64) (string, error) {
	var voteType sql.NullString
	err := r.db.QueryRow("SELECT vote_type FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", entityType, entityID, userID).Scan(&voteType)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return voteType.String, err
}

func (r *EntityRepository) AddReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec(`
		INSERT INTO interactions (entity_type, entity_id, user_id, type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id, type) DO NOTHING`,
		entityType, entityID, userID, reactionType)
	return err
}

func (r *EntityRepository) RemoveReaction(entityType string, entityID, userID int64, reactionType string) error {
	_, err := r.db.Exec("DELETE FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", entityType, entityID, userID, reactionType)
	return err
}

func (r *EntityRepository) GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", entityType, entityID, userID, reactionType).Scan(&count)
	return count > 0, err
}