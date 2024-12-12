package routes

import (
	"net/http"
	"server/common/middleware"
	"server/web/ports/rest"
)

func CategoriesRoutes(mux *http.ServeMux) {
	prefix := "/categories"
	// @Description Get all categories
	// @Produce json
	// @Success 200 {array} categories.CategoryResponseResource
	// @Router /categories [get]
	mux.HandleFunc("GET "+prefix, rest.CategoriesCtrl.GetCategories)

	// @Description Create a category
	// @Produce json
	// @Success 201 categories.CategoryResponseResource
	// @Router /categories [post]
	mux.HandleFunc("POST "+prefix, middleware.ValidateBody(rest.CategoriesCtrl.Create))
}
