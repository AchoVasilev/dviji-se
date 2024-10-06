package routes

import (
	"github.com/gin-gonic/gin"
	"server/web/ports/rest"
)

func CategoriesRoutes(incommingRoutes *gin.Engine) {
	incommingRoutes.GET("/", rest.GetCategories())
}
