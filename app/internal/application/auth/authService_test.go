package auth

import (
	"context"
	"os"
	"server/internal/domain/user"
	"server/util/securityutil"
	"testing"

	"github.com/google/uuid"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	os.Setenv("JWT_KEY", "test-jwt-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_KEY", "test-jwt-refresh-secret-key-for-testing")
	os.Setenv("XSRF", "test-xsrf-key-for-testing")

	return func() {
		os.Unsetenv("JWT_KEY")
		os.Unsetenv("JWT_REFRESH_KEY")
		os.Unsetenv("XSRF")
	}
}

func createTestUser(t *testing.T, password string) user.User {
	t.Helper()
	hashedPassword, err := securityutil.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	return user.User{
		Id:       uuid.New(),
		Email:    "test@example.com",
		Password: hashedPassword,
		Roles: []user.Role{
			{Id: uuid.New(), Name: "USER"},
		},
	}
}

func TestNewAuthService(t *testing.T) {
	service := NewAuthService()
	if service == nil {
		t.Error("NewAuthService() should return non-nil service")
	}
}

func TestAuthenticate_Success(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	password := "password123"
	testUser := createTestUser(t, password)

	service := NewAuthService()
	ctx := context.Background()

	result, err := service.Authenticate(testUser, password, false, ctx)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	if result == nil {
		t.Fatal("Authenticate() returned nil result")
	}

	if result.Token == "" {
		t.Error("Authenticate() Token should not be empty")
	}

	if result.RefreshToken == "" {
		t.Error("Authenticate() RefreshToken should not be empty")
	}

	if result.TokenTime.IsZero() {
		t.Error("Authenticate() TokenTime should not be zero")
	}

	if result.RefreshTokenTime.IsZero() {
		t.Error("Authenticate() RefreshTokenTime should not be zero")
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser(t, "password123")

	service := NewAuthService()
	ctx := context.Background()

	result, err := service.Authenticate(testUser, "wrongpassword", false, ctx)
	if err != ErrHashNotMatch {
		t.Errorf("Authenticate() error = %v, want ErrHashNotMatch", err)
	}

	if result != nil {
		t.Error("Authenticate() should return nil result on error")
	}
}

func TestAuthenticate_EmptyPassword(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser(t, "password123")

	service := NewAuthService()
	ctx := context.Background()

	result, err := service.Authenticate(testUser, "", false, ctx)
	if err != ErrHashNotMatch {
		t.Errorf("Authenticate() error = %v, want ErrHashNotMatch", err)
	}

	if result != nil {
		t.Error("Authenticate() should return nil result on error")
	}
}

func TestAuthenticate_RememberMe(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	password := "password123"
	testUser := createTestUser(t, password)

	service := NewAuthService()
	ctx := context.Background()

	// Without remember me
	resultNoRemember, err := service.Authenticate(testUser, password, false, ctx)
	if err != nil {
		t.Fatalf("Authenticate(rememberMe=false) error = %v", err)
	}

	// With remember me
	resultWithRemember, err := service.Authenticate(testUser, password, true, ctx)
	if err != nil {
		t.Fatalf("Authenticate(rememberMe=true) error = %v", err)
	}

	// RememberMe should have longer expiration
	if !resultWithRemember.TokenTime.After(resultNoRemember.TokenTime) {
		t.Error("Authenticate() with rememberMe should have longer token expiration")
	}

	if !resultWithRemember.RefreshTokenTime.After(resultNoRemember.RefreshTokenTime) {
		t.Error("Authenticate() with rememberMe should have longer refresh token expiration")
	}
}

func TestAuthenticate_TokensAreValid(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	password := "password123"
	testUser := createTestUser(t, password)

	service := NewAuthService()
	ctx := context.Background()

	result, err := service.Authenticate(testUser, password, false, ctx)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	// Verify access token can be parsed
	loggedInUser, err := securityutil.UserFromToken(result.Token)
	if err != nil {
		t.Fatalf("Access token is invalid: %v", err)
	}

	if loggedInUser.Id != testUser.Id.String() {
		t.Errorf("Token user ID = %v, want %v", loggedInUser.Id, testUser.Id.String())
	}

	if loggedInUser.Username != testUser.Email {
		t.Errorf("Token username = %v, want %v", loggedInUser.Username, testUser.Email)
	}

	// Verify refresh token can be validated
	_, err = securityutil.ValidateRefreshToken(result.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh token is invalid: %v", err)
	}
}

func TestAuthenticate_TokenContainsRoles(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	password := "password123"
	hashedPassword, _ := securityutil.HashPassword(password)

	testUser := user.User{
		Id:       uuid.New(),
		Email:    "admin@example.com",
		Password: hashedPassword,
		Roles: []user.Role{
			{Id: uuid.New(), Name: "ADMIN"},
			{Id: uuid.New(), Name: "USER"},
		},
		Permissions: []user.Permission{
			{Id: uuid.New(), Name: "read:posts"},
			{Id: uuid.New(), Name: "write:posts"},
		},
	}

	service := NewAuthService()
	ctx := context.Background()

	result, err := service.Authenticate(testUser, password, false, ctx)
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}

	loggedInUser, err := securityutil.UserFromToken(result.Token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if len(loggedInUser.Roles) != 2 {
		t.Errorf("Token roles count = %d, want 2", len(loggedInUser.Roles))
	}

	roleNames := make(map[string]bool)
	for _, r := range loggedInUser.Roles {
		roleNames[r.Name] = true
	}

	if !roleNames["ADMIN"] {
		t.Error("Token should contain ADMIN role")
	}

	if !roleNames["USER"] {
		t.Error("Token should contain USER role")
	}

	if len(loggedInUser.Permissions) != 2 {
		t.Errorf("Token permissions count = %d, want 2", len(loggedInUser.Permissions))
	}
}

func TestErrHashNotMatch(t *testing.T) {
	if ErrHashNotMatch.Error() != "Hashes didn't match" {
		t.Errorf("ErrHashNotMatch = %v, want 'Hashes didn't match'", ErrHashNotMatch.Error())
	}
}
