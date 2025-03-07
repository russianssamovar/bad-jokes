package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/storage"
	"encoding/json"
	"net/http"
)

type EntityHandler struct {
	entityRepo storage.EntityRepository
}

func NewEntityHandler(repo storage.EntityRepository) *EntityHandler {
	return &EntityHandler{entityRepo: repo}
}

func (h *EntityHandler) Vote(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		VoteType   string `json:"vote_type,omitempty"`
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if input.EntityType != "joke" && input.EntityType != "comment" {
		http.Error(w, "Invalid entity type", http.StatusBadRequest)
		return
	}

	if input.VoteType != "" && input.VoteType != "plus" && input.VoteType != "minus" {
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	if input.VoteType == "" {
		if err := h.entityRepo.RemoveVote(input.EntityType, input.EntityID, userID); err != nil {
			http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.entityRepo.AddVote(input.EntityType, input.EntityID, userID, input.VoteType); err != nil {
		http.Error(w, "Failed to process vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EntityHandler) HandleReaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EntityType   string `json:"entity_type"`
		EntityID     int64  `json:"entity_id"`
		ReactionType string `json:"reaction_type"`
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

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
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	existingReaction, err := h.entityRepo.GetReaction(input.EntityType, input.EntityID, userID, input.ReactionType)
	if err != nil {
		http.Error(w, "Failed to check reaction", http.StatusInternalServerError)
		return
	}

	if existingReaction {
		if err := h.entityRepo.RemoveReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
			http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.entityRepo.AddReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}