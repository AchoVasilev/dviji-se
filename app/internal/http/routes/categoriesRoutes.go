package routes

import (
	"database/sql"
	"net/http"
	"server/internal/application/categories"
	"server/internal/domain/category"
	"server/internal/http/handlers"
	"server/internal/http/middleware"
)

func CategoriesRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := category.NewCategoryRepository(db)
	service := categories.NewCategoryService(repo)
	controller := handlers.NewCategoriesHandler(service)

	prefix := "/categories"
	// @Description Get all categories
	// @Produce json
	// @Success 200 {array} categories.CategoryResponseResource
	// @Router /categories [get]
	mux.HandleFunc("GET "+prefix, controller.GetCategories)

	// @Description Create a category (admin only)
	// @Produce json
	// @Success 201 categories.CategoryResponseResource
	// @Router /categories [post]
	mux.Handle("POST "+prefix, middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(controller.Create))))
}
