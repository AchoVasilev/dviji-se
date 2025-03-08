package securityutil

import (
	"crypto/rsa"
	"log/slog"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func LoadRSAKeys() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	publicKeyBytes, err := os.ReadFile("config/app.pub")
	if err != nil {
		slog.Error(err.Error())
		return nil, nil, err
	}

	privateKeyBytes, err := os.ReadFile("config/app.key")
	if err != nil {
		slog.Error(err.Error())
		return nil, nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		slog.Error(err.Error())
		return nil, nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		slog.Error(err.Error())
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}
