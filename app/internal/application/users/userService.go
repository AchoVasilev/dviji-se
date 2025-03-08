package users

import (
	"context"
	"server/internal/domain/user"
	"server/internal/http/handlers/models"
	"server/util/securityutil"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	userRepository *user.UserRepository
}

func NewUserService(userRepository *user.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return userService.userRepository.FindByEmail(ctx, email)
}

func (userService *UserService) RegisterUser(input *models.CreateUserResource) (uuid.UUID, error) {
	hashed, err := securityutil.HashPassword(input.Password)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	user := user.User{
		Id:        id,
		Email:     input.Email,
		Password:  hashed,
		CreatedAt: time.Now(),
		Status:    "Active",
	}

	err = userService.userRepository.Create(user)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (userService *UserService) LoginUser(ctx context.Context, email string, password string) (*user.User, error) {
	user, err := userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
