package routes

import (
	"database/sql"
	"net/http"
	"server/internal/application/categories"
	"server/internal/domain/category"
	"server/internal/http/controllers"
)

func CategoriesRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := category.NewCategoryRepository(db)
	service := categories.NewCategoryService(repo)
	controller := controllers.NewCategoriesController(service)

	prefix := "/categories"
	// @Description Get all categories
	// @Produce json
	// @Success 200 {array} categories.CategoryResponseResource
	// @Router /categories [get]
	mux.HandleFunc("GET "+prefix, controller.GetCategories)

	// @Description Create a category
	// @Produce json
	// @Success 201 categories.CategoryResponseResource
	// @Router /categories [post]
	mux.HandleFunc("POST "+prefix, controller.Create)
}
