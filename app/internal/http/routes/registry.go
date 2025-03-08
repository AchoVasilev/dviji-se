package routes

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("../../../web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	CategoriesRoutes(mux, db)

	return mux
}
