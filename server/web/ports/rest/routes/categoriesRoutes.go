package routes

import (
	"github.com/gin-gonic/gin"
	"server/web/ports/rest"
)

func CategoriesRoutes(incommingRoutes *gin.RouterGroup) {
	categoriesRoutes := incommingRoutes.Group("/categories")
	{
		// @Tags categories v1
		// @Description Get all categories
		// @Produce json
		// @Success 200 {array} categories.CategoryResponseResource
		// @Router /v1/categories [get]
		categoriesRoutes.GET("/", rest.CategoriesCtrl.GetCategories)
	}
}
