package securityutil

import (
	"os"
	"server/internal/domain/user"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	testJWTKey        = "test-jwt-secret-key-for-testing-only"
	testJWTRefreshKey = "test-jwt-refresh-secret-key-for-testing"
)

func setupTestEnv(t *testing.T) func() {
	t.Helper()
	os.Setenv("JWT_KEY", testJWTKey)
	os.Setenv("JWT_REFRESH_KEY", testJWTRefreshKey)
	os.Setenv("XSRF", "test-xsrf-key-for-testing")

	return func() {
		os.Unsetenv("JWT_KEY")
		os.Unsetenv("JWT_REFRESH_KEY")
		os.Unsetenv("XSRF")
	}
}

func createTestUser() user.User {
	return user.User{
		Id:    uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Email: "test@example.com",
		Roles: []user.Role{
			{Id: uuid.New(), Name: "USER"},
		},
		Permissions: []user.Permission{
			{Id: uuid.New(), Name: "read:posts"},
		},
	}
}

func createAdminUser() user.User {
	return user.User{
		Id:    uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Email: "admin@example.com",
		Roles: []user.Role{
			{Id: uuid.New(), Name: "ADMIN"},
			{Id: uuid.New(), Name: "USER"},
		},
		Permissions: []user.Permission{
			{Id: uuid.New(), Name: "read:posts"},
			{Id: uuid.New(), Name: "write:posts"},
			{Id: uuid.New(), Name: "delete:posts"},
		},
	}
}

func TestGenerateAccessToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()

	tests := []struct {
		name       string
		rememberMe bool
		minExp     time.Duration
		maxExp     time.Duration
	}{
		{"without remember me", false, 14 * time.Minute, 16 * time.Minute},
		{"with remember me", true, 6 * 24 * time.Hour, 8 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiration := GenerateAccessToken(testUser, tt.rememberMe)

			if token == "" {
				t.Error("GenerateAccessToken() returned empty token")
			}

			// Check expiration is within expected range
			now := time.Now().UTC()
			minExp := now.Add(tt.minExp)
			maxExp := now.Add(tt.maxExp)

			if expiration.Before(minExp) || expiration.After(maxExp) {
				t.Errorf("GenerateAccessToken() expiration = %v, want between %v and %v", expiration, minExp, maxExp)
			}

			// Verify token is valid and can be parsed
			loggedInUser, err := UserFromToken(token)
			if err != nil {
				t.Fatalf("UserFromToken() error = %v", err)
			}

			if loggedInUser.Id != testUser.Id.String() {
				t.Errorf("UserFromToken() Id = %v, want %v", loggedInUser.Id, testUser.Id.String())
			}

			if loggedInUser.Username != testUser.Email {
				t.Errorf("UserFromToken() Username = %v, want %v", loggedInUser.Username, testUser.Email)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()

	tests := []struct {
		name       string
		rememberMe bool
		minExp     time.Duration
		maxExp     time.Duration
	}{
		{"without remember me", false, 23 * time.Hour, 25 * time.Hour},
		{"with remember me", true, 29 * 24 * time.Hour, 31 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, expiration := GenerateRefreshToken(testUser, tt.rememberMe)

			if token == "" {
				t.Error("GenerateRefreshToken() returned empty token")
			}

			// Check expiration is within expected range
			now := time.Now().UTC()
			minExp := now.Add(tt.minExp)
			maxExp := now.Add(tt.maxExp)

			if expiration.Before(minExp) || expiration.After(maxExp) {
				t.Errorf("GenerateRefreshToken() expiration = %v, want between %v and %v", expiration, minExp, maxExp)
			}

			// Verify token is valid
			validatedToken, err := ValidateRefreshToken(token)
			if err != nil {
				t.Fatalf("ValidateRefreshToken() error = %v", err)
			}

			if !validatedToken.Valid {
				t.Error("ValidateRefreshToken() token should be valid")
			}
		})
	}
}

func TestUserFromToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createAdminUser()
	token, _ := GenerateAccessToken(testUser, false)

	loggedInUser, err := UserFromToken(token)
	if err != nil {
		t.Fatalf("UserFromToken() error = %v", err)
	}

	if loggedInUser.Id != testUser.Id.String() {
		t.Errorf("UserFromToken() Id = %v, want %v", loggedInUser.Id, testUser.Id.String())
	}

	if loggedInUser.Username != testUser.Email {
		t.Errorf("UserFromToken() Username = %v, want %v", loggedInUser.Username, testUser.Email)
	}

	// Check roles
	if len(loggedInUser.Roles) != len(testUser.Roles) {
		t.Errorf("UserFromToken() Roles count = %d, want %d", len(loggedInUser.Roles), len(testUser.Roles))
	}

	roleNames := make(map[string]bool)
	for _, r := range loggedInUser.Roles {
		roleNames[r.Name] = true
	}

	for _, expectedRole := range testUser.Roles {
		if !roleNames[expectedRole.Name] {
			t.Errorf("UserFromToken() missing role %s", expectedRole.Name)
		}
	}

	// Check permissions
	if len(loggedInUser.Permissions) != len(testUser.Permissions) {
		t.Errorf("UserFromToken() Permissions count = %d, want %d", len(loggedInUser.Permissions), len(testUser.Permissions))
	}

	permNames := make(map[string]bool)
	for _, p := range loggedInUser.Permissions {
		permNames[p.Name] = true
	}

	for _, expectedPerm := range testUser.Permissions {
		if !permNames[expectedPerm.Name] {
			t.Errorf("UserFromToken() missing permission %s", expectedPerm.Name)
		}
	}
}

func TestUserFromToken_InvalidToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"malformed token", "notavalidjwt"},
		{"invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InRlc3QiLCJ1c2VybmFtZSI6InRlc3RAZXhhbXBsZS5jb20ifQ.invalidsignature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UserFromToken(tt.token)
			if err == nil {
				t.Error("UserFromToken() should return error for invalid token")
			}
		})
	}
}

func TestUserFromToken_ExpiredToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()

	// Create an expired token manually
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	claims := jwt.MapClaims{
		"id":          testUser.Id.String(),
		"username":    testUser.Email,
		"roles":       []map[string]string{{"name": "USER"}},
		"permissions": []map[string]string{},
		"exp":         expiredTime.Unix(),
		"iat":         time.Now().UTC().Add(-2 * time.Hour).Unix(),
		"iss":         "dviji-se",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testJWTKey))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	_, err = UserFromToken(tokenStr)
	if err == nil {
		t.Error("UserFromToken() should return error for expired token")
	}
}

func TestUserFromToken_WrongSecret(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()

	// Create token with different secret
	expiration := time.Now().UTC().Add(15 * time.Minute)
	claims := jwt.MapClaims{
		"id":          testUser.Id.String(),
		"username":    testUser.Email,
		"roles":       []map[string]string{{"name": "USER"}},
		"permissions": []map[string]string{},
		"exp":         expiration.Unix(),
		"iat":         time.Now().UTC().Unix(),
		"iss":         "dviji-se",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("wrong-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	_, err = UserFromToken(tokenStr)
	if err == nil {
		t.Error("UserFromToken() should return error for token with wrong secret")
	}
}

func TestValidateRefreshToken(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()
	refreshToken, _ := GenerateRefreshToken(testUser, false)

	t.Run("valid refresh token", func(t *testing.T) {
		token, err := ValidateRefreshToken(refreshToken)
		if err != nil {
			t.Fatalf("ValidateRefreshToken() error = %v", err)
		}

		if !token.Valid {
			t.Error("ValidateRefreshToken() should return valid token")
		}
	})

	t.Run("access token used as refresh token", func(t *testing.T) {
		accessToken, _ := GenerateAccessToken(testUser, false)
		_, err := ValidateRefreshToken(accessToken)
		if err == nil {
			t.Error("ValidateRefreshToken() should reject access token (different secret)")
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		_, err := ValidateRefreshToken("invalid-token")
		if err == nil {
			t.Error("ValidateRefreshToken() should return error for invalid token")
		}
	})
}

func TestGenerateToken_UniqueTokens(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := createTestUser()

	token1, _ := GenerateAccessToken(testUser, false)
	time.Sleep(1100 * time.Millisecond) // Wait > 1 second to ensure different iat (Unix timestamp)
	token2, _ := GenerateAccessToken(testUser, false)

	if token1 == token2 {
		t.Error("GenerateAccessToken() should produce unique tokens (different iat)")
	}
}

func TestUserFromToken_EmptyRolesAndPermissions(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	testUser := user.User{
		Id:          uuid.New(),
		Email:       "noroles@example.com",
		Roles:       []user.Role{},
		Permissions: []user.Permission{},
	}

	token, _ := GenerateAccessToken(testUser, false)
	loggedInUser, err := UserFromToken(token)
	if err != nil {
		t.Fatalf("UserFromToken() error = %v", err)
	}

	if loggedInUser.Roles == nil {
		// It's okay if it's nil, but let's check length
		if len(loggedInUser.Roles) != 0 {
			t.Errorf("UserFromToken() Roles should be empty, got %d", len(loggedInUser.Roles))
		}
	}

	if loggedInUser.Permissions == nil {
		if len(loggedInUser.Permissions) != 0 {
			t.Errorf("UserFromToken() Permissions should be empty, got %d", len(loggedInUser.Permissions))
		}
	}
}

func TestTokenDurations(t *testing.T) {
	if AccessTokenDuration != 15*time.Minute {
		t.Errorf("AccessTokenDuration = %v, want %v", AccessTokenDuration, 15*time.Minute)
	}

	if RefreshTokenDuration != 24*time.Hour {
		t.Errorf("RefreshTokenDuration = %v, want %v", RefreshTokenDuration, 24*time.Hour)
	}

	if RememberMeAccessDuration != 7*24*time.Hour {
		t.Errorf("RememberMeAccessDuration = %v, want %v", RememberMeAccessDuration, 7*24*time.Hour)
	}

	if RememberMeRefreshDuration != 30*24*time.Hour {
		t.Errorf("RememberMeRefreshDuration = %v, want %v", RememberMeRefreshDuration, 30*24*time.Hour)
	}
}
