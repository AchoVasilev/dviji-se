package categories

import (
	"context"
	"log"
	"server/domain/category"
	"time"

	"github.com/google/uuid"
)

type CategoryService struct {
	CategoryRepository *category.CategoryRepository
}

var Service = instance()

func instance() *CategoryService {
	return &CategoryService{CategoryRepository: category.Repository}
}

func (categoryService *CategoryService) GetCategories(ctx context.Context) ([]category.Category, error) {
	return categoryService.CategoryRepository.FindAll(ctx)
}

func (categoryService *CategoryService) Create(ctx context.Context, resource *CreateCategoryResource) (category.Category, error) {
	log.Println("Creating a new category")
	toCreate := category.Category{
		Id:        uuid.New(),
		Name:      resource.Name,
		ImageUrl:  resource.ImageUrl,
		CreatedAt: time.Now(),
	}

	err := categoryService.CategoryRepository.Create(ctx, toCreate)
	if err == nil {
		log.Printf("Created category. [id=%s]", toCreate.Id)
	}

	return toCreate, err
}
