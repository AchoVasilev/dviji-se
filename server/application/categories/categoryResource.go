package categories

import (
	"server/domain/category"
	"time"

	"github.com/google/uuid"
)

type CategoryResponseResource struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ImageUrl  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}

func (dto *CategoryResponseResource) CreateCategoryResponseFrom(cat category.Category) CategoryResponseResource {
	return CategoryResponseResource{
		Id:        cat.Id,
		Name:      cat.Name,
		ImageUrl:  cat.ImageUrl,
		CreatedAt: cat.CreatedAt,
	}
}
