package handlers

import (
	"badJokes/internal/config"
	"badJokes/internal/lib/sl"
	"badJokes/internal/storage"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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

	if err := validateUsername(input.Username); err != nil {
		h.log.Info("Invalid username",
			sl.Err(err),
			slog.String("username", input.Username))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := validatePassword(input.Password); err != nil {
		h.log.Info("Invalid password", sl.Err(err))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

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
		"is_admin": user.IsAdmin,
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

func validateUsername(username string) error {
	const (
		minLength = 3
		maxLength = 20
	)

	length := utf8.RuneCountInString(username)
	if length < minLength {
		return fmt.Errorf("username must be at least %d characters long", minLength)
	}

	if length > maxLength {
		return fmt.Errorf("username must be at most %d characters long", maxLength)
	}

	if strings.Contains(username, " ") {
		return errors.New("username cannot contain spaces")
	}

	validUsernamePattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validUsernamePattern.MatchString(username) {
		return errors.New("username can only contain letters, numbers, dots, underscores, and hyphens")
	}

	forbiddenWords := []string{
		"admin", "administrator", "mod", "moderator", "system", "support",
		"staff", "official", "root", "superuser", "fuck", "shit", "ass",
	}

	lowerUsername := strings.ToLower(username)
	for _, word := range forbiddenWords {
		if strings.Contains(lowerUsername, word) {
			return fmt.Errorf("username contains forbidden word: %s", word)
		}
	}

	return nil
}

func validatePassword(password string) error {
	const (
		minLength = 8
		maxLength = 72
	)

	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}

	if len(password) > maxLength {
		return fmt.Errorf("password must be at most %d characters long", maxLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:,.<>?/~", char):
			hasSpecial = true
		}
	}

	var missing []string
	if !hasUpper {
		missing = append(missing, "an uppercase letter")
	}
	if !hasLower {
		missing = append(missing, "a lowercase letter")
	}
	if !hasNumber {
		missing = append(missing, "a number")
	}
	if !hasSpecial {
		missing = append(missing, "a special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain %s", formatRequirements(missing))
	}

	commonPasswords := map[string]bool{
		"password":    true,
		"123456":      true,
		"qwerty":      true,
		"12345678":    true,
		"111111":      true,
		"1234567890":  true,
		"password123": true,
		"admin":       true,
		"welcome":     true,
		"abc123":      true,
	}

	if commonPasswords[strings.ToLower(password)] {
		return errors.New("password is too common, please choose a more secure password")
	}

	return nil
}

func formatRequirements(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}
