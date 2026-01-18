package models

import (
	"server/internal/domain/category"
	"time"

	"github.com/google/uuid"
)

type CreateCategoryResource struct {
	Name     string `json:"name" validate:"required"`
	Slug     string `json:"slug" validate:"required"`
	ImageUrl string `json:"imageUrl" validate:"required"`
}

type CategoryResponseResource struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	ImageUrl  string    `json:"imageUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

func (dto *CategoryResponseResource) CreateCategoryResponseFrom(cat *category.Category) CategoryResponseResource {
	return CategoryResponseResource{
		Id:        cat.Id,
		Name:      cat.Name,
		Slug:      cat.Slug,
		ImageUrl:  cat.ImageUrl,
		CreatedAt: cat.CreatedAt,
	}
}
