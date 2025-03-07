package postgres

import (
	"badJokes/internal/lib/sl"
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
	r.log.Debug("Adding vote",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("vote_type", voteType))

	_, err := r.db.Exec(`
		INSERT INTO votes (entity_type, entity_id, user_id, vote_type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id) DO UPDATE SET vote_type = $5, modified_at = NOW()`,
		entityType, entityID, userID, voteType, voteType)
	
	if err != nil {
		r.log.Error("Failed to add vote",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
		return err
	}

	r.log.Info("Vote added successfully",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("vote_type", voteType))
	return nil
}

func (r *EntityRepository) RemoveVote(entityType string, entityID, userID int64) error {
	r.log.Debug("Removing vote",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID))

	result, err := r.db.Exec("DELETE FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", 
		entityType, entityID, userID)
	
	if err != nil {
		r.log.Error("Failed to remove vote",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.log.Info("Vote removed successfully",
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
	} else {
		r.log.Debug("No vote found to remove",
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
	}
	
	return nil
}

func (r *EntityRepository) GetVote(entityType string, entityID, userID int64) (string, error) {
	r.log.Debug("Getting vote",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID))

	var voteType sql.NullString
	err := r.db.QueryRow("SELECT vote_type FROM votes WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3", 
		entityType, entityID, userID).Scan(&voteType)
	
	if err == sql.ErrNoRows {
		r.log.Debug("No vote found",
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
		return "", nil
	}
	
	if err != nil {
		r.log.Error("Failed to get vote",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID))
		return "", err
	}

	r.log.Debug("Vote retrieved successfully",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("vote_type", voteType.String))
	return voteType.String, nil
}

func (r *EntityRepository) AddReaction(entityType string, entityID, userID int64, reactionType string) error {
	r.log.Debug("Adding reaction",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("reaction_type", reactionType))

	_, err := r.db.Exec(`
		INSERT INTO interactions (entity_type, entity_id, user_id, type, created_at, modified_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT(entity_type, entity_id, user_id, type) DO NOTHING`,
		entityType, entityID, userID, reactionType)
	
	if err != nil {
		r.log.Error("Failed to add reaction",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID),
			slog.String("reaction_type", reactionType))
		return err
	}

	r.log.Info("Reaction added successfully",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("reaction_type", reactionType))
	return nil
}

func (r *EntityRepository) RemoveReaction(entityType string, entityID, userID int64, reactionType string) error {
	r.log.Debug("Removing reaction",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("reaction_type", reactionType))

	result, err := r.db.Exec("DELETE FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", 
		entityType, entityID, userID, reactionType)
	
	if err != nil {
		r.log.Error("Failed to remove reaction",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID),
			slog.String("reaction_type", reactionType))
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.log.Info("Reaction removed successfully",
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID),
			slog.String("reaction_type", reactionType))
	} else {
		r.log.Debug("No reaction found to remove",
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID),
			slog.String("reaction_type", reactionType))
	}
	
	return nil
}

func (r *EntityRepository) GetReaction(entityType string, entityID, userID int64, reactionType string) (bool, error) {
	r.log.Debug("Checking for reaction",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("reaction_type", reactionType))

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM interactions WHERE entity_type = $1 AND entity_id = $2 AND user_id = $3 AND type = $4", 
		entityType, entityID, userID, reactionType).Scan(&count)
	
	if err != nil {
		r.log.Error("Failed to check reaction",
			sl.Err(err),
			slog.String("entity_type", entityType),
			slog.Int64("entity_id", entityID),
			slog.Int64("user_id", userID),
			slog.String("reaction_type", reactionType))
		return false, err
	}

	r.log.Debug("Reaction check completed",
		slog.String("entity_type", entityType),
		slog.Int64("entity_id", entityID),
		slog.Int64("user_id", userID),
		slog.String("reaction_type", reactionType),
		slog.Bool("exists", count > 0))
	return count > 0, nil
}