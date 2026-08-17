package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"server/internal/config"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDatabase() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.DBHost(), config.DBPort(), config.DBUser(),
		config.DBName(), config.DBPassword(), config.DBSSLMode())

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	db.SetMaxOpenConns(config.DBMaxConns())
	db.SetMaxIdleConns(config.DBMaxConns() / 2)
	db.SetConnMaxLifetime(1 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to database. PORT=%v", config.DBPort()))

	return db
}

func RunMigrations(db *sql.DB) {
	slog.Info("Running database migrations..")

	migrationsPath := getMigrationsPath()

	applyMigrations(db, migrationsPath)

	slog.Info("Successfully applied migrations")
}

func applyMigrations(db *sql.DB, path string) {
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		log.Fatalf("Could not execute migrations: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(path, "pgx5", driver)
	if err != nil {
		log.Fatalf("Could not instantiate migrations. Path=%s Error=%v", path, err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("Error applying migrations. Path=%s, Error=%v", path, err)
	}
}

// getMigrationsPath resolves the migrations directory relative to the working
// directory: ./app locally, /app in the container. It must not be derived from
// runtime.Caller, which yields the build machine's source path.
func getMigrationsPath() string {
	migrationsPath, err := filepath.Abs("cmd/db/migrations")
	if err != nil {
		log.Fatalf("Could not resolve migrations path: %v", err)
	}

	if _, err := os.Stat(migrationsPath); err != nil {
		log.Fatalf("Migrations folder not found at path: %s", migrationsPath)
	}

	return fmt.Sprintf("file://%s", migrationsPath)
}
