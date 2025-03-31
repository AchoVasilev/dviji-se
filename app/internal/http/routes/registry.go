package routes

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"server/internal/http/middleware"
)

func RegisterRoutes(db *sql.DB) *http.ServeMux {
	slog.Info("Registering routes")

	mux := http.NewServeMux()
	rootDir, err := getRootDir()
	if err != nil {
		panic(err)
	}

	staticDir := filepath.Join(rootDir, "web/static")
	fmt.Println("Serving static files from: ", staticDir)
	fileServer := http.FileServer(http.Dir(staticDir))

	mux.Handle("GET /static/", http.StripPrefix("/static/", middleware.CacheStaticAssets(fileServer, staticDir)))

	BaseRoutes(mux, db)
	CategoriesRoutes(mux, db)

	return mux
}

func getRootDir() (string, error) {
	return filepath.Abs("./")
}
