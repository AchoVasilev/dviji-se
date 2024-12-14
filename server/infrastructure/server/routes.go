package server

import (
	"net/http"
	"server/web/ports/rest/routes"
)

func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	routes.CategoriesRoutes(mux)

	return mux
}
