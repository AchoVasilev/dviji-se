package users

import (
	"server/internal/domain/user"
)

type UserService struct {
	userRepository *user.UserRepository
}

func NewUserService(userRepository *user.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}
