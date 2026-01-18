package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
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
	db.SetConnMaxLifetime(1 * time.Minute)

	err = db.Ping()
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
		log.Fatalf("Could not instanciate migrations. Path=%s Error=%v", path, err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error applying migrations. Path=%s, Error=%v", path, err)
	}
}

func getMigrationsPath() string {
	_, currentFilePath, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get runtime caller information")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFilePath), "../..")

	migrationsPath := filepath.Join(projectRoot, "db/migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatalf("Migrations folder not found at path: %s", migrationsPath)
	}

	return fmt.Sprintf("file://%s", migrationsPath)
}

