package routes

import (
	"database/sql"
	"net/http"

	appPosts "server/internal/application/posts"
	"server/internal/domain/posts"
	"server/internal/http/handlers"
)

func FeedRoutes(mux *http.ServeMux, db *sql.DB) {
	postRepo := posts.NewPostRepository(db)
	postService := appPosts.NewPostService(postRepo)

	handler := handlers.NewFeedHandler(postService)

	mux.HandleFunc("GET /feed.xml", handler.GetRSSFeed)
	mux.HandleFunc("GET /sitemap.xml", handler.GetSitemap)
	mux.HandleFunc("GET /robots.txt", handler.GetRobotsTxt)
}
