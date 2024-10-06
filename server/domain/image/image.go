package image

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	Id        uuid.UUID `json:"id"`
	Url       string    `json:"url"`
	OwnerId   uuid.UUID `json:"owner_id"`
	OwnerType string    `json:"owner_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
	IsDeleted bool      `json:"is_deleted"`
}
