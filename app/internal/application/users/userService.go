package users

import (
	"context"
	"database/sql"
	"log/slog"
	"server/internal/domain/user"
	"server/internal/http/handlers/models"
	"server/util/securityutil"
	"time"

	"github.com/google/uuid"
)

type userRepository interface {
	Create(ctx context.Context, u user.User) error
	FindByEmail(ctx context.Context, email string) (user.User, error)
	FindById(ctx context.Context, userId string) (user.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	TokensValidAfter(ctx context.Context, userId string) (sql.NullTime, error)
	RevokeTokensIssuedBefore(ctx context.Context, userId string, cutoff time.Time) error
}

type UserService struct {
	userRepository userRepository
}

func NewUserService(userRepository userRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) GetUserByEmail(ctx context.Context, email string) (user.User, error) {
	return userService.userRepository.FindByEmail(ctx, email)
}

func (userService *UserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return userService.userRepository.ExistsByEmail(ctx, email)
}

func (userService *UserService) GetUserById(ctx context.Context, userId string) (user.User, error) {
	return userService.userRepository.FindById(ctx, userId)
}

// IsSessionValid reports whether a token minted at issuedAt still authenticates
// its user. A token issued at or before the user's revocation cutoff - set on a
// password change - is rejected even though its signature and expiry are fine.
//
// Errors fail closed: if the cutoff cannot be read, the session is refused.
func (userService *UserService) IsSessionValid(ctx context.Context, userId string, issuedAt time.Time) bool {
	validAfter, err := userService.userRepository.TokensValidAfter(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, "Could not read the token revocation cutoff", "error", err, "userId", userId)
		return false
	}

	return sessionValidAt(validAfter, issuedAt)
}

// sessionValidAt compares a token's issue time against a revocation cutoff.
// Shared by the cached and uncached validators so the two cannot drift.
//
// Second resolution on iat means a token minted in the same second as the
// revocation is treated as revoked, which errs towards refusing access.
func sessionValidAt(cutoff sql.NullTime, issuedAt time.Time) bool {
	if !cutoff.Valid {
		return true
	}

	return issuedAt.After(cutoff.Time)
}

// RevokeSessions invalidates every access token issued for the user so far.
func (userService *UserService) RevokeSessions(ctx context.Context, userId string) error {
	return userService.userRepository.RevokeTokensIssuedBefore(ctx, userId, time.Now().UTC())
}

func (userService *UserService) RegisterUser(ctx context.Context, input *models.CreateUserResource) (uuid.UUID, error) {
	// Hashing and id generation can fail; returning the error keeps a bad
	// password hash from taking the request down with a panic.
	hashed, err := securityutil.HashPassword(input.Password)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	newUser := user.User{
		Id:        id,
		Email:     input.Email,
		Password:  hashed,
		CreatedAt: time.Now().UTC(),
		Status:    "Active",
	}

	if err := userService.userRepository.Create(ctx, newUser); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
