package users

import (
	"context"
	"errors"
	"testing"

	"server/internal/domain/user"
)

type stubBootstrapRepo struct {
	adminExists bool
	emailTaken  bool
	created     []user.User
	err         error
}

func (s *stubBootstrapRepo) AdminExists(context.Context) (bool, error) {
	return s.adminExists, s.err
}

func (s *stubBootstrapRepo) ExistsByEmail(context.Context, string) (bool, error) {
	return s.emailTaken, nil
}

func (s *stubBootstrapRepo) CreateAdmin(_ context.Context, u user.User) error {
	s.created = append(s.created, u)
	return nil
}

const strongPassword = "Str0ng!Passw0rd"

func TestEnsureAdmin_CreatesTheFirstAdministrator(t *testing.T) {
	repo := &stubBootstrapRepo{}

	if err := EnsureAdmin(context.Background(), repo, "admin@example.com", strongPassword); err != nil {
		t.Fatalf("EnsureAdmin() = %v", err)
	}

	if len(repo.created) != 1 {
		t.Fatalf("created %d administrators, want 1", len(repo.created))
	}

	created := repo.created[0]
	if created.Email != "admin@example.com" {
		t.Errorf("email = %q", created.Email)
	}

	if created.Password == strongPassword {
		t.Error("the password was stored without hashing")
	}
}

// This is a bootstrap, not a reset: leaving the variables set must never
// overwrite an existing administrator or re-grant access.
func TestEnsureAdmin_DoesNothingWhenAnAdminExists(t *testing.T) {
	repo := &stubBootstrapRepo{adminExists: true}

	if err := EnsureAdmin(context.Background(), repo, "admin@example.com", strongPassword); err != nil {
		t.Fatalf("EnsureAdmin() = %v", err)
	}

	if len(repo.created) != 0 {
		t.Error("an administrator was created even though one already exists")
	}
}

func TestEnsureAdmin_SkipsWhenUnconfigured(t *testing.T) {
	repo := &stubBootstrapRepo{}

	if err := EnsureAdmin(context.Background(), repo, "", ""); err != nil {
		t.Fatalf("EnsureAdmin() = %v", err)
	}

	if len(repo.created) != 0 {
		t.Error("an administrator was created without credentials")
	}
}

// An address the login form rejects would create an account nobody can use.
func TestEnsureAdmin_RejectsAnUnusableEmail(t *testing.T) {
	repo := &stubBootstrapRepo{}

	err := EnsureAdmin(context.Background(), repo, "admin@localhost", strongPassword)
	if !errors.Is(err, ErrInvalidBootstrapEmail) {
		t.Fatalf("EnsureAdmin() = %v, want ErrInvalidBootstrapEmail", err)
	}

	if len(repo.created) != 0 {
		t.Error("an administrator was created with an unusable email")
	}
}

func TestEnsureAdmin_RejectsAWeakPassword(t *testing.T) {
	repo := &stubBootstrapRepo{}

	err := EnsureAdmin(context.Background(), repo, "admin@example.com", "password")
	if !errors.Is(err, ErrWeakBootstrapPassword) {
		t.Fatalf("EnsureAdmin() = %v, want ErrWeakBootstrapPassword", err)
	}
}

func TestEnsureAdmin_RefusesWhenTheEmailIsTaken(t *testing.T) {
	repo := &stubBootstrapRepo{emailTaken: true}

	if err := EnsureAdmin(context.Background(), repo, "admin@example.com", strongPassword); err == nil {
		t.Fatal("expected an error when the address already belongs to an account")
	}

	if len(repo.created) != 0 {
		t.Error("an administrator was created for an address already in use")
	}
}
