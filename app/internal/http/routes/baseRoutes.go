package routes

import (
	"net/http"
	"server/internal/http/handlers"
)

func BaseRoutes(mux *http.ServeMux) {
	handler := handlers.NewDefaultHandler()

	mux.HandleFunc("GET /", handler.HandleHomePage)
}
