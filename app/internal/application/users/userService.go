package users

import (
	"context"
	"server/internal/domain/user"
	"server/internal/http/handlers/models"
	"server/util"
	"server/util/securityutil"
	"time"

	"github.com/google/uuid"
)

type userRepository interface {
	Create(user.User) error
	FindByEmail(ctx context.Context, email string) (user.User, error)
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

func (userService *UserService) RegisterUser(input *models.CreateUserResource) (uuid.UUID, error) {
	hashed := util.MustProduce(securityutil.HashPassword(input.Password))

	id := util.MustProduce(uuid.NewRandom())

	user := user.User{
		Id:        id,
		Email:     input.Email,
		Password:  hashed,
		CreatedAt: time.Now(),
		Status:    "Active",
	}

	err := userService.userRepository.Create(user)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (userService *UserService) LoginUser(ctx context.Context, email string, password string) (user.User, error) {
	user, err := userService.GetUserByEmail(ctx, email)
	if err != nil {
		return user, err
	}

	return user, nil
}
