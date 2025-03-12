package securityutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
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

func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}

	return hex.EncodeToString(bytes)
}

func GenerateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error(err.Error())
		return "", err
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		slog.Error(err.Error())
		return "", err
	}

	hash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	return "sha256-" + hash, nil
}
