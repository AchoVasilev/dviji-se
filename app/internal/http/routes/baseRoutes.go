package routes

import (
	"database/sql"
	"log/slog"
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
	mux.HandleFunc("GET /error", handler.HandleError)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := []byte(`{"status":"ok"}`)
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			body = []byte(`{"status":"unhealthy","reason":"database unreachable"}`)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if _, err := w.Write(body); err != nil {
			slog.WarnContext(r.Context(), "Failed to write health check response", "error", err)
		}
	})
}
