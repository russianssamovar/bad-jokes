package postgres

import (
	"badJokes/internal/lib/sl"
	"badJokes/internal/models"
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db  *sql.DB
	log *slog.Logger
}

func NewUserRepository(db *sql.DB, log *slog.Logger) *UserRepository {
	return &UserRepository{
		db:  db,
		log: log.With(slog.String("component", "user_repository")),
	}
}

func (r *UserRepository) Register(username, email, password string) (int64, error) {
	r.log.Debug("Registering new user",
		slog.String("username", username),
		slog.String("email", email))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		r.log.Error("Failed to hash password", sl.Err(err))
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, email, password, is_password_hashed, created_at, modified_at)
		VALUES ($1, $2, $3, 1, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err = r.db.QueryRow(query, username, email, hashedPassword).Scan(&id)
	if err != nil {
		r.log.Error("Failed to insert user",
			sl.Err(err),
			slog.String("username", username),
			slog.String("email", email))
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	r.log.Info("User registered successfully",
		slog.Int64("user_id", id),
		slog.String("username", username))
	return id, nil
}

func (r *UserRepository) Authenticate(email, password string) (*models.User, error) {
	r.log.Debug("Authenticating user",
		slog.String("email", email))

	var user models.User
	var storedPassword string
	var isPasswordHashed bool

	err := r.db.QueryRow(`
		SELECT id, username, email, password, is_password_hashed
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Username, &user.Email, &storedPassword, &isPasswordHashed)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Info("Authentication failed: user not found",
				slog.String("email", email))
			return nil, fmt.Errorf("user not found")
		}
		r.log.Error("Failed to query user",
			sl.Err(err),
			slog.String("email", email))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if isPasswordHashed {
		if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
			r.log.Info("Authentication failed: invalid password",
				slog.String("email", email),
				slog.Int64("user_id", user.ID))
			return nil, fmt.Errorf("invalid password")
		}
	} else {
		if storedPassword != password {
			r.log.Info("Authentication failed: invalid password",
				slog.String("email", email),
				slog.Int64("user_id", user.ID))
			return nil, fmt.Errorf("invalid password")
		}

		r.log.Debug("Upgrading plaintext password to hashed",
			slog.Int64("user_id", user.ID))

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			r.log.Error("Failed to hash password during upgrade",
				sl.Err(err),
				slog.Int64("user_id", user.ID))
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = r.db.Exec(`
			UPDATE users
			SET password = $1, is_password_hashed = 1, modified_at = NOW()
			WHERE id = $2
		`, hashedPassword, user.ID)

		if err != nil {
			r.log.Error("Failed to update password hash",
				sl.Err(err),
				slog.Int64("user_id", user.ID))
			return nil, fmt.Errorf("failed to update password: %w", err)
		}

		r.log.Info("User password upgraded from plaintext to hash",
			slog.Int64("user_id", user.ID))
	}

	r.log.Info("User authenticated successfully",
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Username))
	return &user, nil
}