package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/models"
	"badJokes/internal/storage"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type JokesHandler struct {
	jokeRepo    storage.JokesRepository
	commentRepo storage.CommentsRepository
	log         *slog.Logger
}

func NewJokesHandler(jokeRepo storage.JokesRepository, commentRepo storage.CommentsRepository, log *slog.Logger) *JokesHandler {
	return &JokesHandler{
		jokeRepo:    jokeRepo,
		commentRepo: commentRepo,
		log:         log.With(slog.String("component", "jokes_handler")),
	}
}

func (h *JokesHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Create joke request received")

	var input struct {
		Body string `json:"body"`
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Unauthorized attempt to create joke")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode joke creation request body",
			sl.Err(err),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := validateJokeContent(input.Body); err != nil {
		h.log.Warn("Invalid joke content",
			sl.Err(err),
			slog.Int64("user_id", userID),
			slog.String("body_length", strconv.Itoa(len(input.Body))))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.log.Debug("Creating joke",
		slog.Int64("user_id", userID),
		slog.String("body_length", strconv.Itoa(len(input.Body))))

	id, err := h.jokeRepo.Insert(input.Body, userID)
	if err != nil {
		h.log.Error("Failed to insert joke",
			sl.Err(err),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to create joke", http.StatusInternalServerError)
		return
	}

	h.log.Info("Joke created successfully",
		slog.Int64("joke_id", id),
		slog.Int64("user_id", userID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func validateJokeContent(content string) error {
	if len(strings.TrimSpace(content)) == 0 {
		return errors.New("joke content cannot be empty")
	}

	const (
		minLength = 5
	)

	contentLength := len(content)
	if contentLength < minLength {
		return fmt.Errorf("joke content must be at least %d characters", minLength)
	}

	if strings.Contains(strings.ToLower(content), "<script") {
		return errors.New("joke content cannot contain script tags")
	}

	dangerousPatterns := []string{
		"javascript:",
		"data:text/html",
		"vbscript:",
		"onclick=",
		"onerror=",
		"onload=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
	}

	lowercaseContent := strings.ToLower(content)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowercaseContent, pattern) {
			return fmt.Errorf("joke content cannot contain potentially harmful code: %s", pattern)
		}
	}

	return nil
}

func (h *JokesHandler) List(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("List jokes request received")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")
	sortField := r.URL.Query().Get("sort_field")
	order := r.URL.Query().Get("order")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			h.log.Debug("Invalid page parameter, using default",
				slog.String("page_str", pageStr),
				slog.Int("default_page", 1))
			page = 1
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			h.log.Debug("Invalid page size parameter, using default",
				slog.String("page_size_str", pageSizeStr),
				slog.Int("default_page_size", 10))
			pageSize = 10
		}
	}

	allowedSortFields := map[string]bool{
		"created_at":      true,
		"modified_at":     true,
		"id":              true,
		"score":           true,
		"reactions_count": true,
		"comments_count":  true,
	}

	if !allowedSortFields[sortField] {
		h.log.Debug("Invalid sort field, using default",
			slog.String("requested_sort", sortField),
			slog.String("default_sort", "created_at"))
		sortField = "created_at"
	}

	if order != "asc" && order != "desc" {
		h.log.Debug("Invalid order parameter, using default",
			slog.String("requested_order", order),
			slog.String("default_order", "desc"))
		order = "desc"
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	h.log.Debug("Fetching jokes list",
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
		slog.String("sort_field", sortField),
		slog.String("order", order),
		slog.Int64("user_id", userID))

	jokesList, err := h.jokeRepo.ListPage(page, pageSize, sortField, order, userID)
	if err != nil {
		h.log.Error("Failed to fetch jokes list",
			sl.Err(err),
			slog.Int("page", page),
			slog.Int("page_size", pageSize))
		http.Error(w, "Failed to fetch jokes", http.StatusInternalServerError)
		return
	}

	h.log.Info("Jokes list fetched successfully",
		slog.Int("count", len(jokesList)),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jokesList)
}

func (h *JokesHandler) GetJoke(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Get joke request received")

	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		h.log.Warn("Invalid joke ID in context")
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		h.log.Error("Failed to parse joke ID",
			sl.Err(err),
			slog.String("joke_id_str", jokeIDStr))
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	h.log.Debug("Fetching joke by ID",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Info("Joke not found", slog.Int64("joke_id", jokeID))
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		h.log.Error("Failed to fetch joke by ID",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	h.log.Debug("Joke fetched successfully",
		slog.Int64("joke_id", jokeID),
		slog.Int64("author_id", joke.AuthorID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(joke)
}

func (h *JokesHandler) DeleteJoke(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Delete joke request received")

	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		h.log.Warn("Invalid joke ID in context")
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		h.log.Error("Failed to parse joke ID",
			sl.Err(err),
			slog.String("joke_id_str", jokeIDStr))
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Unauthorized attempt to delete joke",
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Debug("Fetching joke to verify ownership",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Info("Joke not found", slog.Int64("joke_id", jokeID))
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		h.log.Error("Failed to fetch joke by ID",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	if joke.AuthorID != userID {
		h.log.Warn("Permission denied: User attempted to delete another user's joke",
			slog.Int64("joke_id", jokeID),
			slog.Int64("requesting_user_id", userID),
			slog.Int64("joke_author_id", joke.AuthorID))
		http.Error(w, "Forbidden: You can only delete your own jokes", http.StatusForbidden)
		return
	}

	h.log.Debug("Deleting joke",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	if err := h.jokeRepo.DeleteJoke(jokeID); err != nil {
		h.log.Error("Failed to delete joke",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to delete joke", http.StatusInternalServerError)
		return
	}

	h.log.Info("Joke deleted successfully",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	w.WriteHeader(http.StatusNoContent)
}

func (h *JokesHandler) GetJokeWithComments(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Get joke with comments request received")

	jokeIDStr, ok := r.Context().Value("jokeId").(string)
	if !ok {
		h.log.Warn("Invalid joke ID in context")
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil {
		h.log.Error("Failed to parse joke ID",
			sl.Err(err),
			slog.String("joke_id_str", jokeIDStr))
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	h.log.Debug("Fetching joke with comments",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	joke, err := h.jokeRepo.GetJokeByID(jokeID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Info("Joke not found", slog.Int64("joke_id", jokeID))
			http.Error(w, "Joke not found", http.StatusNotFound)
			return
		}
		h.log.Error("Failed to fetch joke by ID",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to get joke", http.StatusInternalServerError)
		return
	}

	h.log.Debug("Fetching comments for joke",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	comments, err := h.commentRepo.GetCommentsByJokeID(jokeID, userID)
	if err != nil {
		h.log.Error("Failed to fetch comments for joke",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	response := models.JokeWithComments{
		Joke:     joke,
		Comments: comments,
	}

	h.log.Info("Joke with comments fetched successfully",
		slog.Int64("joke_id", jokeID),
		slog.Int("comment_count", len(comments)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
