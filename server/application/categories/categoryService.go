package categories

import (
	"context"
	"log"
	"server/domain/category"
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

func (categoryService *CategoryService) Create(ctx context.Context, resource CreateCategoryResource) (category.Category, error) {
	log.Println("Creating a new category")
	toCreate := category.Category{
		Id:        uuid.New(),
		Name:      resource.Name,
		ImageUrl:  resource.ImageUrl,
		CreatedAt: time.Now(),
	}

	err := categoryService.categoryRepository.Create(ctx, toCreate)
	if err == nil {
		log.Printf("Created category. [id=%s]", toCreate.Id)
	}

	return toCreate, err
}
