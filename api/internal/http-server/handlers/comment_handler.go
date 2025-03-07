package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type CommentHandler struct {
	commentRepo storage.CommentsRepository
	log         *slog.Logger
}

func NewCommentHandler(repo storage.CommentsRepository, log *slog.Logger) *CommentHandler {
	return &CommentHandler{
		commentRepo: repo,
		log:         log.With(slog.String("component", "comment_handler")),
	}
}

func (h *CommentHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Add comment request received")

	var input struct {
		Body     string `json:"body"`
		ParentID *int64 `json:"parent_id"`
	}

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
		h.log.Warn("Unauthorized access attempt to add comment",
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode comment request body",
			sl.Err(err),
			slog.Int64("user_id", userID))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.log.Debug("Adding comment",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID),
		slog.String("body_length", strconv.Itoa(len(input.Body))),
		slog.Any("parent_id", input.ParentID))

	id, err := h.commentRepo.AddComment(jokeID, userID, input.Body, input.ParentID)
	if err != nil {
		h.log.Error("Failed to add comment",
			sl.Err(err),
			slog.Int64("joke_id", jokeID),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	h.log.Info("Comment added successfully",
		slog.Int64("comment_id", id),
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("List comments request received")

	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		h.log.Warn("Invalid joke ID in query",
			slog.String("joke_id_str", jokeIDStr))
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	h.log.Debug("Fetching comments for joke",
		slog.Int64("joke_id", jokeID))

	comments, err := h.commentRepo.GetComments(jokeID)
	if err != nil {
		h.log.Error("Failed to fetch comments",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	h.log.Info("Comments fetched successfully",
		slog.Int64("joke_id", jokeID),
		slog.Int("comment_count", len(comments)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) GetCommentsByJokeID(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Get comments by joke ID request received")

	jokeIDStr := r.URL.Query().Get("joke_id")
	jokeID, err := strconv.ParseInt(jokeIDStr, 10, 64)
	if err != nil || jokeID < 1 {
		h.log.Warn("Invalid joke ID in query",
			slog.String("joke_id_str", jokeIDStr))
		http.Error(w, "Invalid joke ID", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int64)
	h.log.Debug("Fetching comments for joke with user context",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID))

	comments, err := h.commentRepo.GetCommentsByJokeID(jokeID, userID)
	if err != nil {
		h.log.Error("Failed to fetch comments by joke ID",
			sl.Err(err),
			slog.Int64("joke_id", jokeID),
			slog.Int64("user_id", userID))
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	h.log.Info("Comments fetched successfully with user context",
		slog.Int64("joke_id", jokeID),
		slog.Int64("user_id", userID),
		slog.Int("comment_count", len(comments)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Delete comment request received")

	commentIDStr, ok := r.Context().Value("commentId").(string)
	if !ok {
		h.log.Warn("Invalid comment ID in context")
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		h.log.Error("Failed to parse comment ID",
			sl.Err(err),
			slog.String("comment_id_str", commentIDStr))
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Unauthorized access attempt to delete comment",
			slog.Int64("comment_id", commentID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Debug("Fetching comment to verify ownership",
		slog.Int64("comment_id", commentID),
		slog.Int64("user_id", userID))

	comment, err := h.commentRepo.GetCommentByID(commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.log.Info("Comment not found",
				slog.Int64("comment_id", commentID))
			http.Error(w, "Comment not found", http.StatusNotFound)
		} else {
			h.log.Error("Failed to fetch comment",
				sl.Err(err),
				slog.Int64("comment_id", commentID))
			http.Error(w, "Failed to fetch comment", http.StatusInternalServerError)
		}
		return
	}

	if comment.AuthorID != userID {
		h.log.Warn("Permission denied: User attempted to delete another user's comment",
			slog.Int64("comment_id", commentID),
			slog.Int64("requesting_user_id", userID),
			slog.Int64("comment_author_id", comment.AuthorID))
		http.Error(w, "Forbidden: You can only delete your own comments", http.StatusForbidden)
		return
	}

	h.log.Debug("Deleting comment",
		slog.Int64("comment_id", commentID),
		slog.Int64("user_id", userID))

	if err := h.commentRepo.DeleteComment(commentID); err != nil {
		h.log.Error("Failed to delete comment",
			sl.Err(err),
			slog.Int64("comment_id", commentID))
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	h.log.Info("Comment deleted successfully",
		slog.Int64("comment_id", commentID),
		slog.Int64("user_id", userID))

	w.WriteHeader(http.StatusNoContent)
}
