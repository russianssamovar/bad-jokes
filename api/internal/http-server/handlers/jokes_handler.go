package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/models"
	"badJokes/internal/storage"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

type JokesHandler struct {
	jokeRepo    storage.JokesRepository
	commentRepo storage.CommentsRepository
}

func NewJokesHandler(jokeRepo storage.JokesRepository, commentRepo storage.CommentsRepository) *JokesHandler {
	return &JokesHandler{
		jokeRepo:    jokeRepo,
		commentRepo: commentRepo,
	}
}

func (h *JokesHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.jokeRepo.Insert(input.Body, userID)
	if err != nil {
		http.Error(w, "Failed to create joke", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *JokesHandler) List(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	sortField := r.URL.Query().Get("sort_field")
	order := r.URL.Query().Get("order")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}
	}

	allowedSortFields := map[string]bool{
		"created_at":      true,
		"modified_at":     true,
		"id":              true,
		"score":           true,
		"reactions_count": true,
	}

	if !allowedSortFields[sortField] {
		sortField = "created_at"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	jokesList, err := h.jokeRepo.ListPage(page, pageSize, sortField, order, userID)
	if err != nil {
		http.Error(w, "Failed to fetch jokes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jokesList)
}

func (h *JokesHandler) GetJoke(w http.ResponseWriter, r *http.Request) {
	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(joke)
}

func (h *JokesHandler) DeleteJoke(w http.ResponseWriter, r *http.Request) {
	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	if joke.AuthorID != userID {
		http.Error(w, "Forbidden: You can only delete your own jokes", http.StatusForbidden)
		return
	}

	if err := h.jokeRepo.DeleteJoke(jokeID); err != nil {
		http.Error(w, "Failed to delete joke", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JokesHandler) GetJokeWithComments(w http.ResponseWriter, r *http.Request) {
	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	comments, err := h.commentRepo.GetCommentsByJokeID(jokeID, userID)
	if err != nil {
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	response := models.JokeWithComments{
		Joke:     joke,
		Comments: comments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}