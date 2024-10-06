package category

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ImageUrl   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	IsDeleted  bool      `json:"is_deleted"`
}
