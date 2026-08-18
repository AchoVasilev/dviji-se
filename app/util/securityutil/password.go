package securityutil

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// MinPasswordLength applies to passwords being set. It is deliberately not
// enforced at login: existing accounts may hold shorter passwords and must
// still be able to sign in (and then change it).
const MinPasswordLength = 12

// ErrPasswordTooWeak is returned when a new password fails the policy.
var ErrPasswordTooWeak = errors.New("password does not meet the strength requirements")

// IsPasswordStrong reports whether a new password satisfies the policy:
// at least MinPasswordLength characters with an upper case letter, a lower
// case letter, a digit and a symbol.
func IsPasswordStrong(password string) bool {
	if len([]rune(password)) < MinPasswordLength {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CompareHash(hashedPass string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(password))

	return err == nil
}
