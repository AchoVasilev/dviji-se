package routes

import (
	"net/http"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	CategoriesRoutes(mux)

	return mux
}
