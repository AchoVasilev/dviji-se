package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"server/internal/http/middleware"
	"server/internal/http/routes"
	"server/tests/integration/testdb"
	"strings"
	"testing"
)

// createTestHandler creates a handler with middleware stack for testing
func createTestHandler(tdb *testdb.TestDB) http.Handler {
	router := routes.RegisterRoutes(tdb.DB)

	// Apply middleware stack similar to server.Initialize
	stack := middleware.CreateChain(
		middleware.CheckAuth, // Extract user from JWT cookies
	)

	return stack(router)
}

func TestBlogAPI_GetBlogList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("returns empty list when no posts", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("returns published posts only", func(t *testing.T) {
		// Create a user first
		userId := createTestUser(t, tdb)
		categoryId := "dddddddd-dddd-dddd-dddd-dddddddddddd" // recepti category

		// Create a published post
		tdb.SeedTestPost(t, "Published Post", "published-post", "Content here", categoryId, userId, "published")

		// Create a draft post (should not appear)
		tdb.SeedTestPost(t, "Draft Post", "draft-post", "Draft content", categoryId, userId, "draft")

		resp, err := http.Get(server.URL + "/blog")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		// Published post should appear
		if !strings.Contains(bodyStr, "Published Post") {
			t.Error("Response should contain published post")
		}

		// Draft post should not appear
		if strings.Contains(bodyStr, "Draft Post") {
			t.Error("Response should not contain draft post")
		}
	})
}

func TestBlogAPI_GetBlogPost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create test data
	userId := createTestUser(t, tdb)
	categoryId := "dddddddd-dddd-dddd-dddd-dddddddddddd"

	// Create a published post
	tdb.SeedTestPost(t, "My Test Post", "my-test-post", "This is the content of my test post.", categoryId, userId, "published")

	// Create a draft post
	tdb.SeedTestPost(t, "My Draft Post", "my-draft-post", "Draft content here", categoryId, userId, "draft")

	t.Run("returns published post by slug", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/my-test-post")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		if !strings.Contains(bodyStr, "My Test Post") {
			t.Error("Response should contain post title")
		}
	})

	t.Run("returns 404 for draft post", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/my-draft-post")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("returns 404 for non-existent slug", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/non-existent-post")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestBlogAPI_GetBlogByCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create test data
	userId := createTestUser(t, tdb)
	receptiCategoryId := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	fitnesCategoryId := "ffffffff-ffff-ffff-ffff-ffffffffffff"

	// Create posts in different categories
	tdb.SeedTestPost(t, "Recipe Post", "recipe-post", "Delicious recipe", receptiCategoryId, userId, "published")
	tdb.SeedTestPost(t, "Fitness Post", "fitness-post", "Great workout", fitnesCategoryId, userId, "published")

	t.Run("returns posts filtered by category", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/category/recepti")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		// Should contain recipe post
		if !strings.Contains(bodyStr, "Recipe Post") {
			t.Error("Response should contain post from requested category")
		}

		// Should not contain fitness post (different category)
		if strings.Contains(bodyStr, "Fitness Post") {
			t.Error("Response should not contain post from different category")
		}
	})

	t.Run("returns empty for category with no posts", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/category/uprazhneniya")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})
}

func TestBlogAPI_GetRecentPosts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create test data
	userId := createTestUser(t, tdb)
	categoryId := "dddddddd-dddd-dddd-dddd-dddddddddddd"

	// Create some published posts
	tdb.SeedTestPost(t, "Recent Post 1", "recent-post-1", "Content 1", categoryId, userId, "published")
	tdb.SeedTestPost(t, "Recent Post 2", "recent-post-2", "Content 2", categoryId, userId, "published")

	t.Run("returns recent published posts", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/recent")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		if !strings.Contains(bodyStr, "Recent Post") {
			t.Error("Response should contain recent posts")
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/blog/recent?limit=1")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})
}

func TestAdminAPI_CreatePost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := createTestHandler(tdb)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("requires authentication", func(t *testing.T) {
		// Create a client that doesn't follow redirects
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		payload := map[string]string{
			"title":      "Test Post",
			"content":    "Test content",
			"categoryId": "dddddddd-dddd-dddd-dddd-dddddddddddd",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/admin/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should redirect to login (303) or return forbidden (403)
		if resp.StatusCode != http.StatusSeeOther && resp.StatusCode != http.StatusForbidden {
			t.Errorf("Status = %d, want 303 or 403", resp.StatusCode)
		}
	})

	t.Run("requires admin role", func(t *testing.T) {
		// Create a client that doesn't follow redirects
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// Register a regular user and login
		registerPayload := map[string]string{
			"email":          "regularuser@example.com",
			"password":       "password123",
			"repeatPassword": "password123",
		}
		body, _ := json.Marshal(registerPayload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()

		// Login
		loginPayload := map[string]interface{}{
			"email":      "regularuser@example.com",
			"password":   "password123",
			"rememberMe": false,
		}
		body, _ = json.Marshal(loginPayload)
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ = client.Do(req)
		cookies := resp.Cookies()
		resp.Body.Close()

		// Try to create post
		postPayload := map[string]string{
			"title":      "Test Post",
			"content":    "Test content",
			"categoryId": "dddddddd-dddd-dddd-dddd-dddddddddddd",
		}
		body, _ = json.Marshal(postPayload)
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/admin/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return forbidden (403) for non-admin user
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusForbidden)
		}
	})

	t.Run("admin can create post", func(t *testing.T) {
		// Create a client that doesn't follow redirects
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// Register an admin user
		adminEmail := "admin@example.com"
		registerPayload := map[string]string{
			"email":          adminEmail,
			"password":       "adminpass123",
			"repeatPassword": "adminpass123",
		}
		body, _ := json.Marshal(registerPayload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()

		// Assign admin role to user
		var userId string
		err := tdb.DB.QueryRow("SELECT id FROM users WHERE email = $1", adminEmail).Scan(&userId)
		if err != nil {
			t.Fatalf("Failed to get user ID: %v", err)
		}
		tdb.AssignRoleToUser(t, userId, "22222222-2222-2222-2222-222222222222") // ADMIN role

		// Login
		loginPayload := map[string]interface{}{
			"email":      adminEmail,
			"password":   "adminpass123",
			"rememberMe": false,
		}
		body, _ = json.Marshal(loginPayload)
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ = client.Do(req)
		cookies := resp.Cookies()
		resp.Body.Close()

		if len(cookies) == 0 {
			t.Fatal("No cookies returned from login")
		}

		// Create post
		postPayload := map[string]string{
			"title":      "Admin Created Post",
			"content":    "This post was created by an admin.",
			"categoryId": "dddddddd-dddd-dddd-dddd-dddddddddddd",
			"status":     "published",
		}
		body, _ = json.Marshal(postPayload)
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/admin/posts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}

		resp, err = client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Status = %d, want %d. Body: %s", resp.StatusCode, http.StatusCreated, string(bodyBytes))
		}

		// Verify post was created
		var postCount int
		err = tdb.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE title = 'Admin Created Post'").Scan(&postCount)
		if err != nil {
			t.Fatalf("Failed to count posts: %v", err)
		}
		if postCount != 1 {
			t.Errorf("Post count = %d, want 1", postCount)
		}
	})
}

func TestAdminAPI_GetDashboard(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)
	tdb.EnsureCategories(t)

	handler := createTestHandler(tdb)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("redirects unauthenticated users to login", func(t *testing.T) {
		// Create a client that doesn't follow redirects
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Get(server.URL + "/admin")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should redirect to login
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
		}

		location := resp.Header.Get("Location")
		if location != "/login" {
			t.Errorf("Location = %q, want /login", location)
		}
	})
}

// createTestUser creates a user directly in the database and returns the user ID
func createTestUser(t *testing.T, tdb *testdb.TestDB) string {
	t.Helper()

	// Use bcrypt-hashed password for "password123"
	hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MQDaLKCKyQXqxQq5qXJV4xJmQXXqMCG"
	return tdb.SeedTestUser(t, "testuser@example.com", hashedPassword)
}
