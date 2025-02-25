package database

import (
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"

	"go-server/internal/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations runs database migrations
func RunMigrations(dbURL string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// CreateMigration creates a new migration file
func CreateMigration(name string) error {
	timestamp := time.Now().Format("20060102150405")
	upFileName := fmt.Sprintf("migrations/%s_%s.up.sql", timestamp, name)
	downFileName := fmt.Sprintf("migrations/%s_%s.down.sql", timestamp, name)

	if err := createFile(upFileName, "-- Add up migration here"); err != nil {
		return err
	}

	if err := createFile(downFileName, "-- Add down migration here"); err != nil {
		return err
	}

	logger.Info("Created new migration files",
		zap.String("up", upFileName),
		zap.String("down", downFileName),
	)
	return nil
}

func createFile(filename, content string) error {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file %s: %w", filename, err)
	}
	return nil
}
