package routes

import (
	"net/http"
	"server/application/categories"
	"server/domain/category"
	"server/web/ports/rest"
)

func CategoriesRoutes(mux *http.ServeMux) {
	repo := category.NewCategoryRepository()
	service := categories.NewCategoryService(repo)
	controller := rest.NewCategoriesController(service)

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
