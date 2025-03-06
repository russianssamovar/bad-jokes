package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

func Migrate(db *sql.DB, migrationsDir string) error {
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			if err := executeMigration(db, migrationsDir, file.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	return err
}

func executeMigration(db *sql.DB, dir, filename string) error {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM migrations WHERE name = $1)", filename).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check migration existence: %w", err)
	}

	if exists {
		log.Printf("Migration %s already executed", filename)
		return nil
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", filename, err)
	}

	_, err = db.Exec("INSERT INTO migrations (name) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", filename, err)
	}

	log.Printf("Migration %s executed successfully", filename)
	return nil
}
