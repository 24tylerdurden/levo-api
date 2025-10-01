package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DBPath string
}

func NewSQLiteDB(cfg Config) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1) // SQLite has limitations with concurrent writes
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	db.SetConnMaxIdleTime(0) // Keep connections alive indefinitely

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better performance
	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		log.Printf("Warning: Failed to enable WAL mode: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database after configuration: %w", err)
	}

	log.Printf("Connected to database: %s", cfg.DBPath)
	return db, nil
}

func InitializeDatabase(dbPath, migrationsPath string) (*sql.DB, error) {
	// Create a separate connection for migrations
	migrationDB, err := NewSQLiteDB(Config{DBPath: dbPath})
	if err != nil {
		return nil, err
	}

	// Run migrations using the migration connection
	migrator := NewMigrator(migrationDB)
	if err := migrator.RunMigrations(migrationsPath); err != nil {
		migrationDB.Close() // Close the migration connection
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Close the migration connection
	migrationDB.Close()

	// Create a new connection for the application
	appDB, err := NewSQLiteDB(Config{DBPath: dbPath})
	if err != nil {
		return nil, err
	}

	return appDB, nil
}
