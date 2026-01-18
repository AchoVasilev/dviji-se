package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	container *postgres.PostgresContainer
	db        *sql.DB
	once      sync.Once
	initErr   error
)

// TestDB holds the database connection and container for integration tests
type TestDB struct {
	DB        *sql.DB
	Container *postgres.PostgresContainer
}

// TestServer wraps httptest.Server with the test database
type TestServer struct {
	Server *httptest.Server
	DB     *sql.DB
}

// SetupTestDB creates a PostgreSQL container and returns a connection
// The container is shared across all tests in the same test run
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	once.Do(func() {
		ctx := context.Background()

		// Disable reaper for environments without bridge network
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

		// Start PostgreSQL container
		container, initErr = postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("test"),
			postgres.WithPassword("test"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
		if initErr != nil {
			log.Printf("Failed to start postgres container: %v", initErr)
			return
		}

		// Get connection string
		connStr, err := container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			initErr = fmt.Errorf("failed to get connection string: %w", err)
			return
		}

		// Connect to database
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Verify connection
		if err := db.Ping(); err != nil {
			initErr = fmt.Errorf("failed to ping database: %w", err)
			return
		}

		// Run migrations
		if err := runMigrations(db); err != nil {
			initErr = fmt.Errorf("failed to run migrations: %w", err)
			return
		}

		// Seed initial data (roles, permissions)
		if err := seedInitialData(db); err != nil {
			initErr = fmt.Errorf("failed to seed initial data: %w", err)
			return
		}
	})

	if initErr != nil {
		t.Fatalf("Failed to setup test database: %v", initErr)
	}

	return &TestDB{
		DB:        db,
		Container: container,
	}
}

// SetupTestServer creates a test HTTP server with the given handler
func SetupTestServer(t *testing.T, handler http.Handler) *TestServer {
	t.Helper()

	tdb := SetupTestDB(t)
	server := httptest.NewServer(handler)

	return &TestServer{
		Server: server,
		DB:     tdb.DB,
	}
}

// Close closes the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// runMigrations applies database migrations
func runMigrations(db *sql.DB) error {
	// Find migrations directory
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		return fmt.Errorf("could not find migrations directory")
	}

	// Read and execute migration file
	migrationFile := filepath.Join(migrationsDir, "00001_init.up.sql")
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// findMigrationsDir searches for the migrations directory
func findMigrationsDir() string {
	// Try various paths relative to where tests might run
	paths := []string{
		"../../../cmd/db/migrations",
		"../../cmd/db/migrations",
		"cmd/db/migrations",
		"./cmd/db/migrations",
	}

	// Also try from GOPATH or module root
	if wd, err := os.Getwd(); err == nil {
		for i := 0; i < 5; i++ {
			tryPath := filepath.Join(wd, "cmd", "db", "migrations")
			if _, err := os.Stat(tryPath); err == nil {
				return tryPath
			}
			wd = filepath.Dir(wd)
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// seedInitialData seeds roles, permissions, and categories needed for the app
func seedInitialData(db *sql.DB) error {
	// Seed roles
	_, err := db.Exec(`
		INSERT INTO roles (id, name, is_deleted) VALUES
		('11111111-1111-1111-1111-111111111111', 'USER', false),
		('22222222-2222-2222-2222-222222222222', 'ADMIN', false)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// Seed permissions
	_, err = db.Exec(`
		INSERT INTO permissions (id, name, is_deleted) VALUES
		('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'read:posts', false),
		('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'write:posts', false),
		('cccccccc-cccc-cccc-cccc-cccccccccccc', 'delete:posts', false)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// Link roles to permissions
	_, err = db.Exec(`
		INSERT INTO roles_permissions (role_id, permission_id, is_deleted) VALUES
		('11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', false),
		('22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', false),
		('22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', false),
		('22222222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccc', false)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed roles_permissions: %w", err)
	}

	// Seed categories
	_, err = db.Exec(`
		INSERT INTO categories (id, name, slug, image_url, is_deleted) VALUES
		('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Рецепти', 'recepti', 'https://example.com/recepti.jpg', false),
		('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Упражнения', 'uprazhneniya', 'https://example.com/uprazhneniya.jpg', false),
		('ffffffff-ffff-ffff-ffff-ffffffffffff', 'Фитнес зали', 'fitnes-zali', 'https://example.com/fitnes-zali.jpg', false)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}

	return nil
}

// CleanupTables truncates user data tables for test isolation (keeps roles/permissions)
func (tdb *TestDB) CleanupTables(t *testing.T) {
	t.Helper()

	tables := []string{
		"password_reset_tokens",
		"images",
		"posts",
		"categories",
		"users_permissions",
		"users_roles",
		"users",
	}

	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to clean %s: %v", table, err)
		}
	}
}

// Close closes the database connection
// Note: We don't terminate the container here as it's shared across tests
func (tdb *TestDB) Close() {
	// Connection is shared, don't close it
}

// Terminate terminates the container (call this in TestMain if needed)
func Terminate() {
	if container != nil {
		ctx := context.Background()
		container.Terminate(ctx)
	}
}

// SeedTestUser creates a test user and returns the user ID
func (tdb *TestDB) SeedTestUser(t *testing.T, email, hashedPassword string) string {
	t.Helper()

	id := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	_, err := tdb.DB.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, status, is_deleted)
		VALUES ($1, $2, $3, 'Test', 'User', 'ACTIVE', false)
	`, id, email, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}

	return id
}

// SeedTestRole creates a test role and returns the role ID
func (tdb *TestDB) SeedTestRole(t *testing.T, name string) string {
	t.Helper()

	id := fmt.Sprintf("%s-%s", name, "1111-1111-1111-111111111111")
	_, err := tdb.DB.Exec(`
		INSERT INTO roles (id, name, is_deleted)
		VALUES ($1, $2, false)
	`, id, name)
	if err != nil {
		t.Fatalf("Failed to seed test role: %v", err)
	}

	return id
}

// SeedTestCategory creates a test category and returns the category ID
func (tdb *TestDB) SeedTestCategory(t *testing.T, name, slug string) string {
	t.Helper()

	id := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	_, err := tdb.DB.Exec(`
		INSERT INTO categories (id, name, slug, image_url, is_deleted)
		VALUES ($1, $2, $3, 'https://example.com/image.jpg', false)
	`, id, name, slug)
	if err != nil {
		t.Fatalf("Failed to seed test category: %v", err)
	}

	return id
}

// AssignRoleToUser assigns a role to a user
func (tdb *TestDB) AssignRoleToUser(t *testing.T, userId, roleId string) {
	t.Helper()

	_, err := tdb.DB.Exec(`
		INSERT INTO users_roles (user_id, role_id, is_deleted)
		VALUES ($1, $2, false)
	`, userId, roleId)
	if err != nil {
		t.Fatalf("Failed to assign role to user: %v", err)
	}
}

// SeedTestPost creates a test post and returns the post ID
func (tdb *TestDB) SeedTestPost(t *testing.T, title, slug, content, categoryId, creatorId, status string) string {
	t.Helper()

	id := uuid.New().String()
	publishedAt := sql.NullTime{}
	if status == "published" {
		publishedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := tdb.DB.Exec(`
		INSERT INTO posts (id, title, slug, content, excerpt, cover_image_url, status, published_at,
			meta_description, reading_time_minutes, category_id, creator_user_id, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, false)
	`, id, title, slug, content, "Test excerpt", "https://example.com/cover.jpg",
		status, publishedAt, "Test meta description", 5, categoryId, creatorId)
	if err != nil {
		t.Fatalf("Failed to seed test post: %v", err)
	}

	return id
}

// GetCategoryId returns the ID of a seeded category by slug
func (tdb *TestDB) GetCategoryId(t *testing.T, slug string) string {
	t.Helper()

	var id string
	err := tdb.DB.QueryRow(`SELECT id FROM categories WHERE slug = $1`, slug).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to get category ID: %v", err)
	}

	return id
}

// EnsureCategories ensures categories exist in the database
func (tdb *TestDB) EnsureCategories(t *testing.T) {
	t.Helper()

	_, err := tdb.DB.Exec(`
		INSERT INTO categories (id, name, slug, image_url, is_deleted) VALUES
		('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Рецепти', 'recepti', 'https://example.com/recepti.jpg', false),
		('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Упражнения', 'uprazhneniya', 'https://example.com/uprazhneniya.jpg', false),
		('ffffffff-ffff-ffff-ffff-ffffffffffff', 'Фитнес зали', 'fitnes-zali', 'https://example.com/fitnes-zali.jpg', false)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("Failed to ensure categories: %v", err)
	}
}
