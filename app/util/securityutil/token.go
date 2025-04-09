package securityutil

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	jwt.Claims
}

func GenerateAccessToken(id string, username string) (string, time.Time) {
	secret := os.Getenv("JWT_KEY")

	return generateToken(id, username, time.Now().UTC(), []byte(secret))
}

func GenerateRefreshToken(id string, username string) (string, time.Time) {
	secret := os.Getenv("JWT_REFRESH_KEY")

	return generateToken(id, username, time.Now().UTC(), []byte(secret))
}

func generateToken(id string, username string, expirationTime time.Time, secret []byte) (string, time.Time) {
	claims := &Claims{
		Id:       id,
		Username: username,
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
