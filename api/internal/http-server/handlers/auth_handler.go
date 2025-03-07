package handlers

import (
	"badJokes/internal/config"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	repo      storage.UserRepository
	jwtSecret []byte
	log       *slog.Logger
}

func NewAuthHandler(repo storage.UserRepository, cfg *config.Config, log *slog.Logger) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		jwtSecret: []byte(cfg.JWTSecret),
		log:       log.With(slog.String("component", "auth_handler")),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Register request received")

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode registration request", sl.Err(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.log.Debug("Attempting to register user",
		slog.String("username", input.Username),
		slog.String("email", input.Email))

	id, err := h.repo.Register(input.Username, input.Email, input.Password)
	if err != nil {
		h.log.Error("Failed to register user",
			sl.Err(err),
			slog.String("username", input.Username),
			slog.String("email", input.Email))
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	h.log.Info("User registered successfully",
		slog.Int64("user_id", id),
		slog.String("username", input.Username))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  id,
		"username": input.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		h.log.Error("Failed to generate token",
			sl.Err(err),
			slog.Int64("user_id", id))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	h.log.Debug("JWT token generated successfully", slog.Int64("user_id", id))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Login request received")

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.log.Error("Failed to decode login request", sl.Err(err))
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.log.Debug("Attempting to authenticate user", slog.String("email", input.Email))

	user, err := h.repo.Authenticate(input.Email, input.Password)
	if err != nil {
		h.log.Info("Authentication failed",
			sl.Err(err),
			slog.String("email", input.Email))
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	h.log.Info("User authenticated successfully",
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		h.log.Error("Failed to generate token",
			sl.Err(err),
			slog.Int64("user_id", user.ID))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	h.log.Debug("JWT token generated successfully", slog.Int64("user_id", user.ID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
