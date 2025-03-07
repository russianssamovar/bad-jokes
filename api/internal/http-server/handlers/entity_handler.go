package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
)

type EntityHandler struct {
	entityRepo storage.EntityRepository
	log        *slog.Logger
}

func NewEntityHandler(repo storage.EntityRepository, log *slog.Logger) *EntityHandler {
	return &EntityHandler{
		entityRepo: repo,
		log:        log.With(slog.String("component", "entity_handler")),
	}
}

func (h *EntityHandler) Vote(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Vote request received")

	var input struct {
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		VoteType   string `json:"vote_type,omitempty"`
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Unauthorized voting attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode vote request body",
			sl.Err(err),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.log.Debug("Processing vote request",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("vote_type", input.VoteType),
		slog.Int64("user_id", userID))

	if input.EntityType != "joke" && input.EntityType != "comment" {
		h.log.Warn("Invalid entity type in vote request",
			slog.String("entity_type", input.EntityType),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	if input.VoteType != "" && input.VoteType != "plus" && input.VoteType != "minus" {
		h.log.Warn("Invalid vote type in request",
			slog.String("vote_type", input.VoteType),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	if input.VoteType == "" {
		h.log.Debug("Removing vote",
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.Int64("user_id", userID))

		if err := h.entityRepo.RemoveVote(input.EntityType, input.EntityID, userID); err != nil {
			h.log.Error("Failed to remove vote",
				sl.Err(err),
				slog.String("entity_type", input.EntityType),
				slog.Int64("entity_id", input.EntityID),
				slog.Int64("user_id", userID))
			http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
			return
		}

		h.log.Info("Vote removed successfully",
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.Int64("user_id", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.log.Debug("Adding vote",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("vote_type", input.VoteType),
		slog.Int64("user_id", userID))

	if err := h.entityRepo.AddVote(input.EntityType, input.EntityID, userID, input.VoteType); err != nil {
		h.log.Error("Failed to add vote",
			sl.Err(err),
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.String("vote_type", input.VoteType),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to process vote", http.StatusInternalServerError)
		return
	}

	h.log.Info("Vote added successfully",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("vote_type", input.VoteType),
		slog.Int64("user_id", userID))
	w.WriteHeader(http.StatusNoContent)
}

func (h *EntityHandler) HandleReaction(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Reaction request received")

	var input struct {
		EntityType   string `json:"entity_type"`
		EntityID     int64  `json:"entity_id"`
		ReactionType string `json:"reaction_type"`
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Unauthorized reaction attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode reaction request body",
			sl.Err(err),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.log.Debug("Processing reaction request",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("reaction_type", input.ReactionType),
		slog.Int64("user_id", userID))

	validReactions := map[string]bool{
		"laugh":       true,
		"heart":       true,
		"neutral":     true,
		"surprised":   true,
		"fire":        true,
		"poop":        true,
		"thumbs_up":   true,
		"thumbs_down": true,
		"angry":       true,
		"monkey":      true,
	}

	if !validReactions[input.ReactionType] {
		h.log.Warn("Invalid reaction type in request",
			slog.String("reaction_type", input.ReactionType),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	existingReaction, err := h.entityRepo.GetReaction(input.EntityType, input.EntityID, userID, input.ReactionType)
	if err != nil {
		h.log.Error("Failed to check existing reaction",
			sl.Err(err),
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.String("reaction_type", input.ReactionType),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to check reaction", http.StatusInternalServerError)
		return
	}

	if existingReaction {
		h.log.Debug("Removing existing reaction",
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.String("reaction_type", input.ReactionType),
			slog.Int64("user_id", userID))

		if err := h.entityRepo.RemoveReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
			h.log.Error("Failed to remove reaction",
				sl.Err(err),
				slog.String("entity_type", input.EntityType),
				slog.Int64("entity_id", input.EntityID),
				slog.String("reaction_type", input.ReactionType),
				slog.Int64("user_id", userID))
			http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
			return
		}

		h.log.Info("Reaction removed successfully",
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.String("reaction_type", input.ReactionType),
			slog.Int64("user_id", userID))
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.log.Debug("Adding new reaction",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("reaction_type", input.ReactionType),
		slog.Int64("user_id", userID))

	if err := h.entityRepo.AddReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
		h.log.Error("Failed to add reaction",
			sl.Err(err),
			slog.String("entity_type", input.EntityType),
			slog.Int64("entity_id", input.EntityID),
			slog.String("reaction_type", input.ReactionType),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	h.log.Info("Reaction added successfully",
		slog.String("entity_type", input.EntityType),
		slog.Int64("entity_id", input.EntityID),
		slog.String("reaction_type", input.ReactionType),
		slog.Int64("user_id", userID))
	w.WriteHeader(http.StatusNoContent)
}
