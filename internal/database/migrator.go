package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) RunMigrations(migrationsPath string) error {
	// Check if migrations directory exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsPath)
	}

	// Create database driver instance
	driver, err := sqlite3.WithInstance(m.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance
	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		migrator.Close() // Only close migrator on error
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Close migrator but don't close the underlying database connection
	migrator.Close()

	if err == migrate.ErrNoChange {
		log.Println("Database is up to date, no migrations applied")
	} else {
		log.Println("Database migrations completed successfully")
	}

	return nil
}
