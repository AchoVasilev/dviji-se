package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

const (
	TokenExpirationDuration = 1 * time.Hour
	TokenLength             = 32
)

type PasswordResetToken struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

type PasswordResetTokenRepository struct {
	db *sql.DB
}

func NewPasswordResetTokenRepository(db *sql.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// GenerateToken creates a new random token and returns both the plain token (to send to user) and the hash (to store)
func GenerateToken() (plainToken string, tokenHash string, err error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	plainToken = hex.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash = hex.EncodeToString(hash[:])

	return plainToken, tokenHash, nil
}

// HashToken hashes a plain token for comparison
func HashToken(plainToken string) string {
	hash := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(hash[:])
}

// Create stores a new password reset token
func (r *PasswordResetTokenRepository) Create(ctx context.Context, userId uuid.UUID, tokenHash string) (*PasswordResetToken, error) {
	token := &PasswordResetToken{
		Id:        uuid.New(),
		UserId:    userId,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(TokenExpirationDuration),
		CreatedAt: time.Now().UTC(),
	}

	query := `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query, token.Id, token.UserId, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// FindValidByHash finds a valid (not used, not expired) token by its hash
func (r *PasswordResetTokenRepository) FindValidByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error) {
	var token PasswordResetToken

	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.Id, &token.UserId, &token.TokenHash, &token.ExpiresAt, &token.UsedAt, &token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// MarkAsUsed marks a token as used
func (r *PasswordResetTokenRepository) MarkAsUsed(ctx context.Context, tokenId uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, tokenId)
	return err
}

// InvalidateAllForUser marks all tokens for a user as used (e.g., after password change)
func (r *PasswordResetTokenRepository) InvalidateAllForUser(ctx context.Context, userId uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, userId)
	return err
}

// DeleteExpired removes expired tokens (for cleanup job)
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM password_reset_tokens WHERE expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
