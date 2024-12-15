package posts

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	Id            uuid.UUID    `json:"id"`
	Title         string       `json:"title"`
	Content       string       `json:"content"`
	CategoryId    uuid.UUID    `json:"category_id"`
	CreatorUserId uuid.UUID    `json:"creator_user_id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
	UpdatedBy     string       `json:"updated_by"`
	IsDeleted     bool         `json:"is_deleted"`
}
