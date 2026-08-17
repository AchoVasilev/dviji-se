package users

import (
	"context"
	"server/internal/domain/user"
	"server/internal/http/handlers/models"
	"server/util/securityutil"
	"time"

	"github.com/google/uuid"
)

type userRepository interface {
	Create(ctx context.Context, u user.User) error
	FindByEmail(ctx context.Context, email string) (user.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type UserService struct {
	userRepository userRepository
}

func NewUserService(userRepository userRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) GetUserByEmail(ctx context.Context, email string) (user.User, error) {
	return userService.userRepository.FindByEmail(ctx, email)
}

func (userService *UserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return userService.userRepository.ExistsByEmail(ctx, email)
}

func (userService *UserService) RegisterUser(ctx context.Context, input *models.CreateUserResource) (uuid.UUID, error) {
	// Hashing and id generation can fail; returning the error keeps a bad
	// password hash from taking the request down with a panic.
	hashed, err := securityutil.HashPassword(input.Password)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	newUser := user.User{
		Id:        id,
		Email:     input.Email,
		Password:  hashed,
		CreatedAt: time.Now().UTC(),
		Status:    "Active",
	}

	if err := userService.userRepository.Create(ctx, newUser); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
