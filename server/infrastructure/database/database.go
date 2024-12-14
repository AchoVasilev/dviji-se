package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/lib/pq"
)

var Db *sql.DB

func ConnectDatabase() {
	host := os.Getenv("HOST")
	port := os.Getenv("DBPORT")
	user := os.Getenv("DBUSER")
	password := os.Getenv("DBPASSWORD")
	dbName := os.Getenv("DBNAME")

	dbSetup := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbName, password)
	db, err := sql.Open("postgres", dbSetup)

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	maxConnections := 25
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	Db = db
	slog.Info(fmt.Sprintf("Successfully connected to database. PORT=%v", port))
}

func RunMigrations(db *sql.DB) {
	slog.Info("Running database migrations..")

	migrationsPath := "file://db/migrations"

	applyMigrations(db, migrationsPath)

	slog.Info("Successfully applied migrations")
}

func applyMigrations(db *sql.DB, path string) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not execute migrations: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		log.Fatalf("Could not instanciate migrations. Path=%s Error=%v", path, err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error applying migrations. Path=%s, Error=%v", path, err)
	}
}
