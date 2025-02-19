package handlers

import (
	"badJokes/api/internal/http-server/middleware"
	"badJokes/api/internal/storage/sqlite/jokes"
	"encoding/json"
	"net/http"
	"strconv"
)

type JokesHandler struct {
	repo *jokes.Repository
}

func NewJokesHandler(repo *jokes.Repository) *JokesHandler {
	return &JokesHandler{repo: repo}
}

func (h *JokesHandler) ListJokes(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	sortField := r.URL.Query().Get("sortField")
	order := r.URL.Query().Get("order")

	var currentUserID int64
	if userID, ok := r.Context().Value(middleware.UserIDKey).(int64); ok {
		currentUserID = userID
	}

	jokesList, err := h.repo.ListPage(page, pageSize, sortField, order, currentUserID)
	if err != nil {
		http.Error(w, "Failed to fetch jokes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jokesList)
}

func (h *JokesHandler) CreateJoke(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body string `json:"body"`
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

	id, err := h.repo.Insert(input.Body, userID)
	if err != nil {
		http.Error(w, "Failed to create joke", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *JokesHandler) VoteEntity(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EntityType string `json:"entity_type"`
		EntityID   int64  `json:"entity_id"`
		VoteType   string `json:"vote_type"`
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

	if input.VoteType != "plus" && input.VoteType != "minus" {
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	existingVote, err := h.repo.GetVote(input.EntityType, input.EntityID, userID)
	if err != nil {
		http.Error(w, "Failed to check vote", http.StatusInternalServerError)
		return
	}

	if existingVote == input.VoteType {
		if err := h.repo.RemoveVote(input.EntityType, input.EntityID, userID); err != nil {
			http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.repo.AddVote(input.EntityType, input.EntityID, userID, input.VoteType); err != nil {
		http.Error(w, "Failed to vote", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JokesHandler) ReactToEntity(w http.ResponseWriter, r *http.Request) {
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

	existingReaction, err := h.repo.GetReaction(input.EntityType, input.EntityID, userID, input.ReactionType)
	if err != nil {
		http.Error(w, "Failed to check reaction", http.StatusInternalServerError)
		return
	}

	if existingReaction {
		if err := h.repo.RemoveReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
			http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.repo.AddReaction(input.EntityType, input.EntityID, userID, input.ReactionType); err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JokesHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		JokeID int64  `json:"joke_id"`
		Body   string `json:"body"`
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

	id, err := h.repo.AddComment(input.JokeID, userID, input.Body)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *JokesHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	comments, err := h.repo.GetComments(jokeID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *JokesHandler) DeleteJoke(w http.ResponseWriter, r *http.Request) {
	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	joke, err := h.repo.GetJokeByID(jokeID)
	if err != nil {
		http.Error(w, "Joke not found", http.StatusNotFound)
		return
	}

	if joke.AuthorID != userID {
		http.Error(w, "Forbidden: You can only delete your own jokes", http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteJoke(jokeID); err != nil {
		http.Error(w, "Failed to delete joke", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
