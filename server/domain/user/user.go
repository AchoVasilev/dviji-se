package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Status    UserStatus `json:"status"`
	Roles     []Role     `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	IsDeleted bool       `json:"is_deleted"`
}
