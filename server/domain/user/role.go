package user

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	Id          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	UpdatedBy   string       `json:"updated_by"`
	IsDeleted   bool         `json:"is_deleted"`
}
