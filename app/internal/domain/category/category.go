package category

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	Id        uuid.UUID
	Name      string
	Slug      string
	ImageUrl  string
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	IsDeleted bool
}
