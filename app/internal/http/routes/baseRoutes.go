package routes

import (
	"database/sql"
	"net/http"
	"server/internal/application/categories"
	"server/internal/domain/category"
	"server/internal/http/handlers"
)

func BaseRoutes(mux *http.ServeMux, db *sql.DB) {
	categoriesRepo := category.NewCategoryRepository(db)
	categoryService := categories.NewCategoryService(categoriesRepo)
	handler := handlers.NewDefaultHandler(categoryService)

	mux.HandleFunc("GET /", handler.HandleHomePage)
	mux.HandleFunc("GET /not-found", handler.HandleNotFound)
}
