package routes

import (
	"database/sql"
	"net/http"

	"server/internal/application/categories"
	appPosts "server/internal/application/posts"
	"server/internal/domain/category"
	"server/internal/domain/posts"
	"server/internal/http/handlers"
	"server/internal/http/middleware"
)

func BlogRoutes(mux *http.ServeMux, db *sql.DB) {
	postRepo := posts.NewPostRepository(db)
	postService := appPosts.NewPostService(postRepo)

	categoryRepo := category.NewCategoryRepository(db)
	categoryService := categories.NewCategoryService(categoryRepo)

	handler := handlers.NewBlogHandler(postService, categoryService)

	// Blog list page
	mux.HandleFunc("GET /blog", handler.GetBlogList)

	// Search
	mux.Handle("GET /blog/search/suggestions", middleware.FragmentOnly("/blog/search", http.HandlerFunc(handler.SearchSuggestions)))
	mux.HandleFunc("GET /blog/search", handler.SearchBlogPosts)

	// Recent posts (HTMX endpoint for home page)
	mux.Handle("GET /blog/recent", middleware.FragmentOnly("/blog", http.HandlerFunc(handler.GetRecentPosts)))

	// Blog by category
	mux.HandleFunc("GET /blog/category/{slug}", handler.GetBlogByCategory)

	// Single blog post (must be after /blog/category to not conflict)
	mux.HandleFunc("GET /blog/{slug}", handler.GetBlogPost)
}
