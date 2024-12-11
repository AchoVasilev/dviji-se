package routes

import (
	"net/http"
	"server/web/ports/rest"
)

func CategoriesRoutes(mux *http.ServeMux) {
	prefix := "/categories"
	// @Tags categories v1
	// @Description Get all categories
	// @Produce json
	// @Success 200 {array} categories.CategoryResponseResource
	// @Router /v1/categories [get]
	mux.HandleFunc("GET "+prefix, rest.CategoriesCtrl.GetCategories)
}
