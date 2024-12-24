package categories

import (
	"context"
	"server/internal/domain/category"
	"server/internal/http/handlers/models"
	"time"

	"github.com/google/uuid"
)

type CategoryService struct {
	categoryRepository *category.CategoryRepository
}

func NewCategoryService(categoryRepository *category.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepository: categoryRepository}
}

func (categoryService *CategoryService) GetCategories(ctx context.Context) ([]category.Category, error) {
	return categoryService.categoryRepository.FindAll(ctx)
}

func (categoryService *CategoryService) Create(ctx context.Context, resource models.CreateCategoryResource) (*category.Category, error) {
	toCreate := category.Category{
		Id:        uuid.New(),
		Name:      resource.Name,
		ImageUrl:  resource.ImageUrl,
		CreatedAt: time.Now(),
	}

	result, err := categoryService.categoryRepository.Create(ctx, toCreate)

	return result, err
}
