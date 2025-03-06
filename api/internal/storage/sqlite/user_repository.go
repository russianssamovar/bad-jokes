package sqlite

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

	stmt, err := r.db.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, email, hashedPassword)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *UserRepository) Authenticate(email, password string) (*models.User, error) {
	var user models.User
	var storedPassword string
	var isPasswordHashed int

	err := r.db.QueryRow(`
		SELECT id, username, email, password, is_password_hashed
		FROM users
		WHERE email = ?
	`, email).Scan(&user.ID, &user.Username, &user.Email, &storedPassword, &isPasswordHashed)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if isPasswordHashed == 1 {
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
			SET password = ?, is_password_hashed = 1, modified_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, hashedPassword, user.ID)

		if err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	return &user, nil
}
