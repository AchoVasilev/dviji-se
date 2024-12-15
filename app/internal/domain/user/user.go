package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID
	Email     string
	Password  string
	FirstName string
	LastName  string
	Status    UserStatus
	Roles     []Role
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	IsDeleted bool
}
