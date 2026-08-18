package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uuid.UUID
	Email       string
	Password    string
	FirstName   sql.NullString
	LastName    sql.NullString
	Status      UserStatus
	Roles       []Role
	Permissions []Permission
	CreatedAt   time.Time
	UpdatedAt   sql.NullTime
	IsDeleted   bool

	// TokensValidAfter invalidates every access token issued at or before it.
	TokensValidAfter sql.NullTime
}
