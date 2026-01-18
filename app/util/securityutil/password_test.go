package securityutil

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!#$%^&*()"},
		{"unicode password", "пароль123"},
		{"empty password", ""},
		{"long password", "thisisaverylongpasswordthatshouldalsobehashed12345678901234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword() error = %v", err)
			}

			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			if hash == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}

			// Verify it's a valid bcrypt hash
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			if err != nil {
				t.Errorf("HashPassword() produced invalid bcrypt hash: %v", err)
			}
		})
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "testpassword"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() should produce different hashes for the same password (due to salt)")
	}
}

func TestCompareHash(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{"correct password", hash, "password123", true},
		{"wrong password", hash, "wrongpassword", false},
		{"empty password against hash", hash, "", false},
		{"password with extra char", hash, "password1234", false},
		{"password missing char", hash, "password12", false},
		{"case sensitive", hash, "PASSWORD123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareHash(tt.hash, tt.password)
			if got != tt.want {
				t.Errorf("CompareHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareHash_InvalidHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		password string
	}{
		{"empty hash", "", "password123"},
		{"invalid hash format", "notavalidbcrypthash", "password123"},
		{"truncated hash", "$2a$12$LQv3c1yqBWVHxkd0", "password123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareHash(tt.hash, tt.password)
			if got {
				t.Error("CompareHash() should return false for invalid hash")
			}
		})
	}
}

func TestHashPassword_BcryptCost(t *testing.T) {
	password := "testpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// bcrypt hashes start with $2a$XX$ where XX is the cost
	// Our cost is 12, so it should be $2a$12$
	if len(hash) < 7 {
		t.Fatal("Hash too short to check cost")
	}

	expectedPrefix := "$2a$12$"
	if hash[:7] != expectedPrefix {
		t.Errorf("HashPassword() cost mismatch, got prefix %s, want %s", hash[:7], expectedPrefix)
	}
}
