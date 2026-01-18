package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"server/internal/http/routes"
	"server/tests/integration/testdb"
	"testing"
)

func setupAuthTestEnv(t *testing.T) func() {
	t.Helper()
	os.Setenv("JWT_KEY", "test-jwt-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_KEY", "test-jwt-refresh-secret-key-for-testing")
	os.Setenv("XSRF", "test-xsrf-key")
	os.Setenv("ENVIRONMENT", "test")

	return func() {
		os.Unsetenv("JWT_KEY")
		os.Unsetenv("JWT_REFRESH_KEY")
		os.Unsetenv("XSRF")
		os.Unsetenv("ENVIRONMENT")
	}
}

func TestAuthAPI_Register(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("successful registration", func(t *testing.T) {
		payload := map[string]string{
			"email":          "newuser@example.com",
			"password":       "password123",
			"repeatPassword": "password123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Handler returns 200 OK on success
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("duplicate email returns conflict", func(t *testing.T) {
		// First registration
		payload := map[string]string{
			"email":          "duplicate@example.com",
			"password":       "password123",
			"repeatPassword": "password123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()

		// Second registration with same email
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusConflict)
		}
	})

	t.Run("invalid email returns validation error", func(t *testing.T) {
		payload := map[string]string{
			"email":          "not-an-email",
			"password":       "password123",
			"repeatPassword": "password123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("short password returns validation error", func(t *testing.T) {
		payload := map[string]string{
			"email":          "shortpass@example.com",
			"password":       "short",
			"repeatPassword": "short",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

}

func TestAuthAPI_Login(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Register a user first
	registerPayload := map[string]string{
		"email":          "logintest@example.com",
		"password":       "password123",
		"repeatPassword": "password123",
	}
	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	t.Run("successful login", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":      "logintest@example.com",
			"password":   "password123",
			"rememberMe": false,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		// Check for auth cookies
		cookies := resp.Cookies()
		hasAuthCookie := false
		for _, c := range cookies {
			if c.Name == "X-LOGIN-TOKEN" {
				hasAuthCookie = true
			}
		}

		if !hasAuthCookie {
			t.Error("Response should have X-LOGIN-TOKEN cookie")
		}
	})

	t.Run("wrong password returns not found", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":      "logintest@example.com",
			"password":   "wrongpassword",
			"rememberMe": false,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Handler returns 404 for invalid credentials (to prevent enumeration)
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})

	t.Run("non-existent email returns not found", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":      "nonexistent@example.com",
			"password":   "password123",
			"rememberMe": false,
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestAuthAPI_PasswordReset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cleanup := setupAuthTestEnv(t)
	defer cleanup()

	tdb := testdb.SetupTestDB(t)
	tdb.CleanupTables(t)

	handler := routes.RegisterRoutes(tdb.DB)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Register a user first
	registerPayload := map[string]string{
		"email":          "resettest@example.com",
		"password":       "password123",
		"repeatPassword": "password123",
	}
	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	t.Run("request password reset - existing email", func(t *testing.T) {
		payload := map[string]string{
			"email": "resettest@example.com",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/forgot-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return OK regardless of whether email exists (prevents enumeration)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("request password reset - non-existent email still returns OK", func(t *testing.T) {
		payload := map[string]string{
			"email": "nonexistent@example.com",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/forgot-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Should return OK even for non-existent email (prevents enumeration)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("reset password with invalid token returns bad request", func(t *testing.T) {
		payload := map[string]string{
			"token":          "invalid-token-12345678901234567890123456789012345678901234567890123456",
			"password":       "newpassword123",
			"repeatPassword": "newpassword123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

}
