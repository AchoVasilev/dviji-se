package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type dbConfig struct {
	host                  string
	port                  string
	user                  string
	password              string
	dbName                string
	sslMode               string
	maxConnections        int
	maxConnectionLifeTime time.Duration
}

func ConnectDatabase() *sql.DB {
	dbConfig := initDbConfig()
	db, err := sql.Open("pgx", dbConfig.connectionString())

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	db.SetMaxOpenConns(dbConfig.maxConnections)
	db.SetConnMaxLifetime(dbConfig.maxConnectionLifeTime)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	slog.Info(fmt.Sprintf("Successfully connected to database. PORT=%v", dbConfig.port))

	return db
}

func RunMigrations(db *sql.DB) {
	slog.Info("Running database migrations..")

	migrationsPath := getMigrationsPath()

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

func initDbConfig() *dbConfig {
	connections, err := strconv.Atoi(os.Getenv("DBCONNECTIONS"))
	if err != nil {
		connections = 10
	}

	return &dbConfig{
		host:                  os.Getenv("DBHOST"),
		port:                  os.Getenv("DBPORT"),
		user:                  os.Getenv("DBUSER"),
		password:              os.Getenv("DBPASSWORD"),
		dbName:                os.Getenv("DBNAME"),
		sslMode:               os.Getenv("DBSSLMODE"),
		maxConnections:        connections,
		maxConnectionLifeTime: 1 * time.Minute,
	}
}

func (dbConfig *dbConfig) connectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		dbConfig.host, dbConfig.port, dbConfig.user, dbConfig.dbName, dbConfig.password, dbConfig.sslMode)
}
