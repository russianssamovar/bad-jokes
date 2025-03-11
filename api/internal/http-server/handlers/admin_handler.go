package handlers

import (
	"badJokes/internal/http-server/middleware"
	"badJokes/internal/lib/sl"
	"badJokes/internal/models"
	"badJokes/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	userRepo    storage.UserRepository
	jokeRepo    storage.JokesRepository
	commentRepo storage.CommentsRepository
	log         *slog.Logger
}

func NewAdminHandler(userRepo storage.UserRepository, jokeRepo storage.JokesRepository, commentRepo storage.CommentsRepository, log *slog.Logger) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		jokeRepo:    jokeRepo,
		commentRepo: commentRepo,
		log:         log.With(slog.String("component", "admin_handler")),
	}
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin get users request received")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	pageSize := 20
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}
	}

	adminID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Admin ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Info("Admin fetching users list",
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
		slog.Int64("admin_id", adminID))

	users, err := h.userRepo.GetUsers(page, pageSize)
	if err != nil {
		h.log.Error("Failed to fetch users",
			sl.Err(err),
			slog.Int("page", page),
			slog.Int("page_size", pageSize))
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	count, err := h.userRepo.GetUserCount()
	if err != nil {
		h.log.Error("Failed to get user count", sl.Err(err))
		http.Error(w, "Failed to get user count", http.StatusInternalServerError)
		return
	}

	response := struct {
		Users      []*models.User `json:"users"`
		Page       int            `json:"page"`
		PageSize   int            `json:"page_size"`
		TotalCount int            `json:"total_count"`
		TotalPages int            `json:"total_pages"`
	}{
		Users:      users,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: count,
		TotalPages: (count + pageSize - 1) / pageSize,
	}

	h.log.Info("Users list fetched successfully",
		slog.Int("user_count", len(users)),
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
		slog.Int("total_count", count),
		slog.Int64("admin_id", adminID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) DeleteJoke(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin delete joke request received")

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

	adminID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Admin ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Info("Admin deleting joke",
		slog.Int64("joke_id", jokeID),
		slog.Int64("admin_id", adminID))

	if err := h.jokeRepo.DeleteJoke(jokeID); err != nil {
		h.log.Error("Failed to delete joke",
			sl.Err(err),
			slog.Int64("joke_id", jokeID))
		http.Error(w, "Failed to delete joke", http.StatusInternalServerError)
		return
	}

	h.log.Info("Joke deleted by admin",
		slog.Int64("joke_id", jokeID),
		slog.Int64("admin_id", adminID))

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin delete comment request received")

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

	adminID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Admin ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.log.Info("Admin deleting comment",
		slog.Int64("comment_id", commentID),
		slog.Int64("admin_id", adminID))

	if err := h.commentRepo.DeleteComment(commentID); err != nil {
		h.log.Error("Failed to delete comment",
			sl.Err(err),
			slog.Int64("comment_id", commentID))
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	h.log.Info("Comment deleted by admin",
		slog.Int64("comment_id", commentID),
		slog.Int64("admin_id", adminID))

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) SetUserAdminStatus(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin set user status request received")

	var input struct {
		UserID  int64 `json:"user_id"`
		IsAdmin bool  `json:"is_admin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode request", sl.Err(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	adminID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Warn("Admin ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if input.UserID == adminID {
		h.log.Warn("Admin attempted to change own status",
			slog.Int64("admin_id", adminID))
		http.Error(w, "Cannot change your own admin status", http.StatusBadRequest)
		return
	}

	h.log.Info("Admin changing user status",
		slog.Int64("target_user_id", input.UserID),
		slog.Bool("new_status", input.IsAdmin),
		slog.Int64("admin_id", adminID))

	if err := h.userRepo.SetAdminStatus(input.UserID, input.IsAdmin); err != nil {
		h.log.Error("Failed to update user admin status",
			sl.Err(err),
			slog.Int64("user_id", input.UserID))
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	h.log.Info("User admin status changed successfully",
		slog.Int64("user_id", input.UserID),
		slog.Bool("is_admin", input.IsAdmin),
		slog.Int64("changed_by_admin_id", adminID))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *AdminHandler) GetModLogs(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin get moderation logs request received")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	pageSize := 50
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 50
		}
	}

	logs, err := h.userRepo.GetModerationLogs(page, pageSize)
	if err != nil {
		h.log.Error("Failed to fetch moderation logs", sl.Err(err))
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	h.log.Info("Admin fetched moderation logs",
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
		slog.Int("count", len(logs)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *AdminHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Admin get user stats request received")

	stats, err := h.userRepo.GetUserStats()
	if err != nil {
		h.log.Error("Failed to fetch user stats", sl.Err(err))
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	h.log.Info("Admin fetched user statistics")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
