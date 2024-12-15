package routes

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	CategoriesRoutes(mux, db)

	return mux
}
