package securityutil

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"server/internal/domain/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id          string            `json:"id"`
	Username    string            `json:"username"`
	Roles       []user.Role       `json:"roles"`
	Permissions []user.Permission `json:"permissions"`
	jwt.Claims
}

type LoggedInUser struct {
	Id          string
	Username    string
	Roles       []user.Role
	Permissions []user.Permission
}

func GenerateAccessToken(user user.User) (string, time.Time) {
	secret := os.Getenv("JWT_KEY")

	return generateToken(user, time.Now().UTC(), []byte(secret))
}

func GenerateRefreshToken(user user.User) (string, time.Time) {
	secret := os.Getenv("JWT_REFRESH_KEY")

	return generateToken(user, time.Now().UTC(), []byte(secret))
}

func UserFromToken(tokenStr string) (*LoggedInUser, error) {
	token, err := validateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid user claims")
	}

	id, ok := claims["id"].(string)
	if !ok {
		return nil, errors.New("Invalid id claim")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("Invalid username claim")
	}

	roles, ok := claims["roles"].([]user.Role)
	if !ok {
		roles = []user.Role{}
	}

	permissions, ok := claims["permissions"].([]user.Permission)
	if !ok {
		permissions = []user.Permission{}
	}

	return &LoggedInUser{
		Id:          id,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func ValidateRefreshToken(tokenStr string) (*jwt.Token, error) {
	token, err := parseToken(tokenStr, os.Getenv("JWT_REFRESH_KEY"))
	if err != nil {
		return nil, err
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return token, nil
	}

	return nil, errors.New("Invalid refresh token provided")
}

func validateToken(tokenStr string) (*jwt.Token, error) {
	token, err := parseToken(tokenStr, os.Getenv("JWT_KEY"))
	if err != nil {
		return nil, err
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return token, nil
	}

	return nil, errors.New("Invalid token provided")
}

func parseToken(tokenStr string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	return token, err
}

func generateToken(user user.User, expirationTime time.Time, secret []byte) (string, time.Time) {
	claims := &Claims{
		Id:       user.Id.String(),
		Username: user.Email,
		Claims: jwt.MapClaims{
			"exp": expirationTime,
			"iat": time.Now().UTC(),
			"iss": "dviji-se",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	return tokenStr, expirationTime
}
