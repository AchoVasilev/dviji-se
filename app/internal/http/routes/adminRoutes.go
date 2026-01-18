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
	"server/internal/infrastructure/cloudinary"
)

func AdminRoutes(mux *http.ServeMux, db *sql.DB) {
	postRepo := posts.NewPostRepository(db)
	postService := appPosts.NewPostService(postRepo)

	categoryRepo := category.NewCategoryRepository(db)
	categoryService := categories.NewCategoryService(categoryRepo)

	cloudinaryService, _ := cloudinary.NewCloudinaryService()

	handler := handlers.NewAdminHandler(postService, categoryService, cloudinaryService)

	// Wrap all admin routes with auth and admin middleware
	adminAuth := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(h)))
	}

	// Dashboard
	mux.Handle("GET /admin", adminAuth(handler.GetDashboard))
	mux.Handle("GET /admin/", adminAuth(handler.GetDashboard))

	// Posts management
	mux.Handle("GET /admin/posts", adminAuth(handler.GetPosts))
	mux.Handle("GET /admin/posts/new", adminAuth(handler.GetPostForm))
	mux.Handle("GET /admin/posts/{id}", adminAuth(handler.GetPostForm))
	mux.Handle("POST /admin/posts", adminAuth(handler.CreatePost))
	mux.Handle("PUT /admin/posts/{id}", adminAuth(handler.UpdatePost))
	mux.Handle("DELETE /admin/posts/{id}", adminAuth(handler.DeletePost))

	// Image upload
	mux.Handle("POST /admin/upload", adminAuth(handler.UploadImage))
}
