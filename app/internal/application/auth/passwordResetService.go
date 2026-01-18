package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"server/internal/domain/user"
	"server/internal/infrastructure/email"
	"server/util/securityutil"
)

var (
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrUserNotFound  = errors.New("user not found")
	ErrPasswordWeak  = errors.New("password too weak")
)

type PasswordResetService struct {
	userRepo      *user.UserRepository
	tokenRepo     *user.PasswordResetTokenRepository
	emailService  *email.EmailService
}

func NewPasswordResetService(
	userRepo *user.UserRepository,
	tokenRepo *user.PasswordResetTokenRepository,
	emailService *email.EmailService,
) *PasswordResetService {
	return &PasswordResetService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
	}
}

// RequestReset initiates a password reset for the given email
// Always returns nil to prevent email enumeration attacks
func (s *PasswordResetService) RequestReset(ctx context.Context, emailAddr string) error {
	// Find user by email
	u, err := s.userRepo.FindByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Don't reveal that the email doesn't exist
			slog.Info("Password reset requested for non-existent email", "email", emailAddr)
			return nil
		}
		return err
	}

	// Generate token
	plainToken, tokenHash, err := user.GenerateToken()
	if err != nil {
		return err
	}

	// Invalidate any existing tokens for this user
	if err := s.tokenRepo.InvalidateAllForUser(ctx, u.Id); err != nil {
		slog.Warn("Failed to invalidate existing tokens", "error", err)
	}

	// Store token hash
	_, err = s.tokenRepo.Create(ctx, u.Id, tokenHash)
	if err != nil {
		return err
	}

	// Send email with plain token
	err = s.emailService.SendPasswordResetEmail(u.Email, plainToken)
	if err != nil {
		slog.Error("Failed to send password reset email", "error", err, "email", emailAddr)
		// Don't return error to user - token is still valid if they got the email
	}

	slog.Info("Password reset requested", "email", emailAddr)
	return nil
}

// ResetPassword resets the password using the token
func (s *PasswordResetService) ResetPassword(ctx context.Context, plainToken, newPassword string) error {
	// Validate password strength
	if len(newPassword) < 8 {
		return ErrPasswordWeak
	}

	// Hash the token and find it
	tokenHash := user.HashToken(plainToken)
	token, err := s.tokenRepo.FindValidByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidToken
		}
		return err
	}

	// Hash the new password
	hashedPassword, err := securityutil.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user's password
	err = s.userRepo.UpdatePassword(ctx, token.UserId.String(), hashedPassword)
	if err != nil {
		return err
	}

	// Mark token as used
	err = s.tokenRepo.MarkAsUsed(ctx, token.Id)
	if err != nil {
		slog.Warn("Failed to mark token as used", "error", err)
	}

	// Invalidate all other tokens for this user
	err = s.tokenRepo.InvalidateAllForUser(ctx, token.UserId)
	if err != nil {
		slog.Warn("Failed to invalidate other tokens", "error", err)
	}

	slog.Info("Password reset successful", "userId", token.UserId)
	return nil
}

// ValidateToken checks if a token is valid without using it
func (s *PasswordResetService) ValidateToken(ctx context.Context, plainToken string) (bool, error) {
	tokenHash := user.HashToken(plainToken)
	_, err := s.tokenRepo.FindValidByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
