package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/storage"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	commentRepo storage.CommentsRepository
}

func NewCommentHandler(repo storage.CommentsRepository) *CommentHandler {
	return &CommentHandler{commentRepo: repo}
}

func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Body     string `json:"body"`
		ParentID *int64 `json:"parent_id"`
	}

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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	id, err := h.commentRepo.AddComment(jokeID, userID, input.Body, input.ParentID)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	comments, err := h.commentRepo.GetComments(jokeID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) GetCommentsByJokeID(w http.ResponseWriter, r *http.Request) {
	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)

	comments, err := h.commentRepo.GetCommentsByJokeID(jokeID, userID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr, ok := r.Context().Value("commentId").(string)
	if !ok {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comment, err := h.commentRepo.GetCommentByID(commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch comment", http.StatusInternalServerError)
		}
		return
	}

	if comment.AuthorID != userID {
		http.Error(w, "Forbidden: You can only delete your own comments", http.StatusForbidden)
		return
	}

	if err := h.commentRepo.DeleteComment(commentID); err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}