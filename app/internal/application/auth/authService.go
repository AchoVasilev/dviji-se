package auth

import (
	"context"
	"errors"
	"server/internal/domain/user"
	"server/util/securityutil"
	"time"
)

var ErrHashNotMatch = errors.New("Hashes didn't match")

type TokenResult struct {
	Token            string
	TokenTime        time.Time
	RefreshToken     string
	RefreshTokenTime time.Time
}

type AuthService struct{}

func (auth *AuthService) Authenticate(user user.User, password string, ctx context.Context) (*TokenResult, error) {
	hashMatch := securityutil.CompareHash(user.Password, password)
	if !hashMatch {
		return nil, ErrHashNotMatch
	}

	token, tokenTime := securityutil.GenerateAccessToken(user.Id.String(), user.Email)
	refreshToken, refreshTokenTime := securityutil.GenerateRefreshToken(user.Id.String(), user.Email)
	return &TokenResult{
		Token:            token,
		TokenTime:        tokenTime,
		RefreshToken:     refreshToken,
		RefreshTokenTime: refreshTokenTime,
	}, nil
}
