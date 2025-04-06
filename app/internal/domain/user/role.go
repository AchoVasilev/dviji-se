package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	Id          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Permissions []Permission   `json:"permissions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
	UpdatedBy   sql.NullString `json:"updated_by"`
	IsDeleted   bool           `json:"is_deleted"`
}
