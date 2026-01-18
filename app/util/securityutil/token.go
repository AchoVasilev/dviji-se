package securityutil

import (
	"errors"
	"fmt"
	"log/slog"
	"server/internal/config"
	"server/internal/domain/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token durations
const (
	AccessTokenDuration     = 15 * time.Minute
	RefreshTokenDuration    = 24 * time.Hour
	RememberMeAccessDuration  = 7 * 24 * time.Hour  // 7 days
	RememberMeRefreshDuration = 30 * 24 * time.Hour // 30 days
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

func GenerateAccessToken(user user.User, rememberMe bool) (string, time.Time) {
	duration := AccessTokenDuration
	if rememberMe {
		duration = RememberMeAccessDuration
	}
	expiration := time.Now().UTC().Add(duration)

	return generateToken(user, expiration, []byte(config.JWTAccessKey()))
}

func GenerateRefreshToken(user user.User, rememberMe bool) (string, time.Time) {
	duration := RefreshTokenDuration
	if rememberMe {
		duration = RememberMeRefreshDuration
	}
	expiration := time.Now().UTC().Add(duration)

	return generateToken(user, expiration, []byte(config.JWTRefreshKey()))
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

	// Parse roles from JWT claims (they come as []interface{} of maps)
	var roles []user.Role
	if rolesRaw, ok := claims["roles"].([]interface{}); ok {
		for _, r := range rolesRaw {
			if roleMap, ok := r.(map[string]interface{}); ok {
				role := user.Role{}
				if name, ok := roleMap["name"].(string); ok {
					role.Name = name
				}
				roles = append(roles, role)
			}
		}
	}

	// Parse permissions from JWT claims
	var permissions []user.Permission
	if permsRaw, ok := claims["permissions"].([]interface{}); ok {
		for _, p := range permsRaw {
			if permMap, ok := p.(map[string]interface{}); ok {
				perm := user.Permission{}
				if name, ok := permMap["name"].(string); ok {
					perm.Name = name
				}
				permissions = append(permissions, perm)
			}
		}
	}

	return &LoggedInUser{
		Id:          id,
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

func ValidateRefreshToken(tokenStr string) (*jwt.Token, error) {
	token, err := parseToken(tokenStr, config.JWTRefreshKey())
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
	token, err := parseToken(tokenStr, config.JWTAccessKey())
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

func generateToken(u user.User, expirationTime time.Time, secret []byte) (string, time.Time) {
	// Build roles for JWT (only include name to keep token small)
	rolesForJwt := make([]map[string]string, len(u.Roles))
	for i, r := range u.Roles {
		rolesForJwt[i] = map[string]string{"name": r.Name}
	}

	// Build permissions for JWT
	permsForJwt := make([]map[string]string, len(u.Permissions))
	for i, p := range u.Permissions {
		permsForJwt[i] = map[string]string{"name": p.Name}
	}

	claims := jwt.MapClaims{
		"id":          u.Id.String(),
		"username":    u.Email,
		"roles":       rolesForJwt,
		"permissions": permsForJwt,
		"exp":         expirationTime.Unix(),
		"iat":         time.Now().UTC().Unix(),
		"iss":         "dviji-se",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	return tokenStr, expirationTime
}
