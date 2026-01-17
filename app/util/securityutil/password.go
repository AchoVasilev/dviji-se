package securityutil

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CompareHash(hashedPass string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))

	return err == nil
}
