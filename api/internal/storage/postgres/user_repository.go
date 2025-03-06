package postgres

import (
	"badJokes/internal/models"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Register(username, email, password string) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, email, password, is_password_hashed, created_at, modified_at)
		VALUES ($1, $2, $3, TRUE, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err = r.db.QueryRow(query, username, email, hashedPassword).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}
	return id, nil
}

func (r *UserRepository) Authenticate(email, password string) (*models.User, error) {
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if isPasswordHashed {
		if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
			return nil, fmt.Errorf("invalid password")
		}
	} else {
		if storedPassword != password {
			return nil, fmt.Errorf("invalid password")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}

		_, err = r.db.Exec(`
			UPDATE users
			SET password = $1, is_password_hashed = 1, modified_at = NOW()
			WHERE id = $2
		`, hashedPassword, user.ID)

		if err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	return &user, nil
}
