package postgres

import (
	"badJokes/internal/lib/sl"
	"badJokes/internal/models"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

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
		SELECT id, username, email, password, is_password_hashed, is_admin, created_at, modified_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Username, &user.Email, &storedPassword, &isPasswordHashed, &user.IsAdmin, &user.CreatedAt, &user.ModifiedAt)

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

func (r *UserRepository) GetUsers(page, pageSize int) ([]*models.User, error) {
	r.log.Debug("Fetching users with pagination",
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	offset := (page - 1) * pageSize

	query := `
		SELECT id, username, email, is_admin, created_at, modified_at
		FROM users
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		r.log.Error("Failed to fetch users", sl.Err(err))
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var createdAt, modifiedAt time.Time

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.IsAdmin,
			&createdAt,
			&modifiedAt,
		)

		if err != nil {
			r.log.Error("Failed to scan user row", sl.Err(err))
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		user.CreatedAt = createdAt.Format(time.RFC3339)
		user.ModifiedAt = modifiedAt.Format(time.RFC3339)
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating user rows", sl.Err(err))
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	r.log.Info("Successfully fetched users",
		slog.Int("page", page),
		slog.Int("page_size", pageSize),
		slog.Int("count", len(users)))

	return users, nil
}

func (r *UserRepository) GetUserCount() (int, error) {
	r.log.Debug("Getting total user count")

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		r.log.Error("Failed to get user count", sl.Err(err))
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	r.log.Debug("User count retrieved", slog.Int("count", count))
	return count, nil
}

func (r *UserRepository) SetAdminStatus(userID int64, isAdmin bool) error {
	r.log.Debug("Setting user admin status",
		slog.Int64("user_id", userID),
		slog.Bool("is_admin", isAdmin))

	query := `
		UPDATE users 
		SET is_admin = $1, modified_at = NOW() 
		WHERE id = $2
	`

	result, err := r.db.Exec(query, isAdmin, userID)
	if err != nil {
		r.log.Error("Failed to update user admin status",
			sl.Err(err),
			slog.Int64("user_id", userID))
		return fmt.Errorf("failed to update user admin status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected", sl.Err(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.log.Warn("No user found with ID", slog.Int64("user_id", userID))
		return fmt.Errorf("no user found with ID %d", userID)
	}

	logQuery := `
		INSERT INTO moderation_logs 
		(action, target_id, target_type, performed_by, details, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	details := fmt.Sprintf("Changed admin status to %v", isAdmin)
	_, err = r.db.Exec(logQuery, "SET_ADMIN_STATUS", userID, "user", userID, details)
	if err != nil {
		r.log.Error("Failed to log admin status change",
			sl.Err(err),
			slog.Int64("user_id", userID))
	}

	r.log.Info("User admin status updated successfully",
		slog.Int64("user_id", userID),
		slog.Bool("new_status", isAdmin))

	return nil
}

func (r *UserRepository) GetModerationLogs(page, pageSize int) ([]*models.ModerationLog, error) {
	r.log.Debug("Fetching moderation logs",
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	offset := (page - 1) * pageSize

	query := `
		SELECT ml.id, ml.action, ml.target_id, ml.target_type, ml.performed_by, 
		       ml.details, ml.created_at, u.username as admin_username
		FROM moderation_logs ml
		LEFT JOIN users u ON ml.performed_by = u.id
		ORDER BY ml.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		r.log.Error("Failed to fetch moderation logs", sl.Err(err))
		return nil, fmt.Errorf("failed to fetch moderation logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.ModerationLog
	for rows.Next() {
		var log models.ModerationLog
		var createdAt time.Time

		err := rows.Scan(
			&log.ID,
			&log.Action,
			&log.TargetID,
			&log.TargetType,
			&log.PerformedBy,
			&log.Details,
			&createdAt,
			&log.AdminUsername,
		)

		if err != nil {
			r.log.Error("Failed to scan moderation log", sl.Err(err))
			return nil, fmt.Errorf("failed to scan moderation log: %w", err)
		}

		log.CreatedAt = createdAt.Format(time.RFC3339)
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating moderation logs", sl.Err(err))
		return nil, fmt.Errorf("error iterating moderation logs: %w", err)
	}

	r.log.Info("Successfully fetched moderation logs",
		slog.Int("count", len(logs)),
		slog.Int("page", page),
		slog.Int("page_size", pageSize))

	return logs, nil
}

func (r *UserRepository) GetUserStats() (*models.UserStats, error) {
	r.log.Debug("Getting user statistics")

	stats := &models.UserStats{}

	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		r.log.Error("Failed to get total users count", sl.Err(err))
		return nil, fmt.Errorf("failed to get total users count: %w", err)
	}

	err = r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE is_admin = true`).Scan(&stats.AdminCount)
	if err != nil {
		r.log.Error("Failed to get admin count", sl.Err(err))
		return nil, fmt.Errorf("failed to get admin count: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&stats.NewUsersToday)
	if err != nil {
		r.log.Error("Failed to get new users count", sl.Err(err))
		return nil, fmt.Errorf("failed to get new users count: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= NOW() - INTERVAL '7 days'
	`).Scan(&stats.NewUsersThisWeek)
	if err != nil {
		r.log.Error("Failed to get weekly new users count", sl.Err(err))
		return nil, fmt.Errorf("failed to get weekly new users count: %w", err)
	}

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`).Scan(&stats.NewUsersThisMonth)
	if err != nil {
		r.log.Error("Failed to get monthly new users count", sl.Err(err))
		return nil, fmt.Errorf("failed to get monthly new users count: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT u.id, u.username, 
		       COUNT(DISTINCT j.id) as jokes_count, 
		       COUNT(DISTINCT c.id) as comments_count
		FROM users u
		LEFT JOIN jokes j ON u.id = j.author_id
		LEFT JOIN comments c ON u.id = c.user_id
		GROUP BY u.id, u.username
		ORDER BY (COUNT(DISTINCT j.id) + COUNT(DISTINCT c.id)) DESC
		LIMIT 5
	`)
	if err != nil {
		r.log.Error("Failed to get active users", sl.Err(err))
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	defer rows.Close()

	var activeUsers []*models.ActiveUser
	for rows.Next() {
		var user models.ActiveUser

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.JokesCount,
			&user.CommentsCount,
		)

		if err != nil {
			r.log.Error("Failed to scan active user", sl.Err(err))
			return nil, fmt.Errorf("failed to scan active user: %w", err)
		}

		activeUsers = append(activeUsers, &user)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error iterating active users", sl.Err(err))
		return nil, fmt.Errorf("error iterating active users: %w", err)
	}

	stats.MostActiveUsers = activeUsers

	r.log.Info("User statistics retrieved successfully")
	return stats, nil
}

func (r *UserRepository) FindOrCreateOAuthUser(email, username, provider, providerID string) (*models.User, error) {
	r.log.Debug("Finding or creating OAuth user",
		slog.String("provider", provider),
		slog.String("provider_id", providerID),
		slog.String("email", email))

	tx, err := r.db.Begin()
	if err != nil {
		r.log.Error("Failed to begin transaction", sl.Err(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	var createdAt, modifiedAt time.Time

	err = tx.QueryRow(`
		SELECT id, username, email, is_admin, created_at, modified_at
		FROM users
		WHERE provider = $1 AND provider_id = $2
	`, provider, providerID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsAdmin,
		&createdAt,
		&modifiedAt,
	)

	if err == nil {
		user.CreatedAt = createdAt.Format(time.RFC3339)
		user.ModifiedAt = modifiedAt.Format(time.RFC3339)

		r.log.Info("Found existing OAuth user", slog.Int64("user_id", user.ID))

		if err = tx.Commit(); err != nil {
			r.log.Error("Failed to commit transaction", sl.Err(err))
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return &user, nil
	}

	if err != sql.ErrNoRows {
		r.log.Error("Database error when finding user", sl.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	err = tx.QueryRow(`
		SELECT id, username, email, is_admin, created_at, modified_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsAdmin,
		&createdAt,
		&modifiedAt,
	)

	if err == nil {
		_, err = tx.Exec(`
			UPDATE users
			SET provider = $1, provider_id = $2, modified_at = NOW()
			WHERE id = $3
		`, provider, providerID, user.ID)

		if err != nil {
			r.log.Error("Failed to update user with OAuth info", sl.Err(err))
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		user.CreatedAt = createdAt.Format(time.RFC3339)
		user.ModifiedAt = time.Now().Format(time.RFC3339)

		r.log.Info("Linked existing user to OAuth account",
			slog.Int64("user_id", user.ID),
			slog.String("provider", provider))

		if err = tx.Commit(); err != nil {
			r.log.Error("Failed to commit transaction", sl.Err(err))
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return &user, nil
	}

	if err != sql.ErrNoRows {
		r.log.Error("Database error when finding user by email", sl.Err(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		r.log.Error("Failed to generate random password", sl.Err(err))
		return nil, fmt.Errorf("failed to generate random password: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(randomBytes, bcrypt.DefaultCost)
	if err != nil {
		r.log.Error("Failed to hash random password", sl.Err(err))
		return nil, fmt.Errorf("failed to hash random password: %w", err)
	}

	err = tx.QueryRow(`
		INSERT INTO users (username, email, provider, provider_id, password, is_password_hashed, created_at, modified_at)
		VALUES ($1, $2, $3, $4, $5, 1, NOW(), NOW())
		RETURNING id, username, email, is_admin, created_at, modified_at
	`, username, email, provider, providerID, hashedPassword).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.IsAdmin,
		&createdAt,
		&modifiedAt,
	)

	if err != nil {
		r.log.Error("Failed to create new OAuth user", sl.Err(err))
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = createdAt.Format(time.RFC3339)
	user.ModifiedAt = modifiedAt.Format(time.RFC3339)

	r.log.Info("Created new user via OAuth",
		slog.Int64("user_id", user.ID),
		slog.String("provider", provider))

	if err = tx.Commit(); err != nil {
		r.log.Error("Failed to commit transaction", sl.Err(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}
