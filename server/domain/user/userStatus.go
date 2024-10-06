package user

import "errors"

type UserStatus string

const (
	Active    UserStatus = "Active"
	Inactive  UserStatus = "Inactive"
	Suspended UserStatus = "Suspended"
)

func ValidateUserStatus(status UserStatus) error {
	switch status {
	case Active, Inactive, Suspended:
		return nil
	default:
		return errors.New("Invalid user status!")
	}
}
