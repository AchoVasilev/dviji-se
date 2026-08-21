package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// RoleAdmin is the only role that may reach the admin panel. Names are
// compared exactly: the seed data stores them upper case.
const RoleAdmin = "ADMIN"

// HasRole reports whether the roles contain one with the given name.
func HasRole(roles []Role, name string) bool {
	for _, role := range roles {
		if role.Name == name {
			return true
		}
	}

	return false
}

type Role struct {
	Id          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Permissions []Permission   `json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
	UpdatedBy   sql.NullString `json:"updated_by"`
	IsDeleted   bool           `json:"is_deleted"`
}
