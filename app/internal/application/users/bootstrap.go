package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"server/internal/domain/user"
	"server/util/httputils"
	"server/util/securityutil"

	"github.com/google/uuid"
)

// ErrWeakBootstrapPassword is returned when ADMIN_PASSWORD does not meet the
// password policy. Failing loudly beats seeding a weak administrator.
var ErrWeakBootstrapPassword = errors.New("ADMIN_PASSWORD does not meet the password policy")

// ErrInvalidBootstrapEmail is returned when ADMIN_EMAIL would not pass the
// login form's validation, which would leave the account unusable.
var ErrInvalidBootstrapEmail = errors.New("ADMIN_EMAIL is not a valid email address")

// adminBootstrapRepository is the slice of the repository this needs.
type adminBootstrapRepository interface {
	AdminExists(ctx context.Context) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateAdmin(ctx context.Context, u user.User) error
}

// EnsureAdmin creates the first administrator when the database has none.
//
// Registration only ever grants the USER role, so without this a fresh
// deployment has no way to reach the admin panel short of writing SQL against
// production.
//
// It is a bootstrap, not a reset: once any administrator exists this does
// nothing, so leaving the variables set cannot overwrite a password or
// re-grant access.
func EnsureAdmin(ctx context.Context, repo adminBootstrapRepository, email, password string) error {
	adminExists, err := repo.AdminExists(ctx)
	if err != nil {
		return fmt.Errorf("could not check for an existing administrator: %w", err)
	}

	if adminExists {
		return nil
	}

	if email == "" || password == "" {
		slog.WarnContext(ctx, "No administrator exists and ADMIN_EMAIL/ADMIN_PASSWORD are unset; the admin panel cannot be used until one is created")
		return nil
	}

	if !httputils.IsValidEmail(email) {
		return fmt.Errorf("%w: %q", ErrInvalidBootstrapEmail, email)
	}

	if !securityutil.IsPasswordStrong(password) {
		return ErrWeakBootstrapPassword
	}

	// Guard against an account already holding the address without the role.
	emailTaken, err := repo.ExistsByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("could not check the administrator email: %w", err)
	}

	if emailTaken {
		return fmt.Errorf("cannot bootstrap the administrator: %q is already registered", email)
	}

	hashed, err := securityutil.HashPassword(password)
	if err != nil {
		return fmt.Errorf("could not hash the administrator password: %w", err)
	}

	admin := user.User{
		Id:        uuid.New(),
		Email:     email,
		Password:  hashed,
		CreatedAt: time.Now().UTC(),
		Status:    "Active",
	}

	if err := repo.CreateAdmin(ctx, admin); err != nil {
		return fmt.Errorf("could not create the administrator: %w", err)
	}

	slog.InfoContext(ctx, "Created the initial administrator", "email", email)

	return nil
}
