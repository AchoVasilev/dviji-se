package rest

import (
	"context"
	"log"
	"net/http"
	"server/application/categories"
	"server/common/api"
	"time"

	"github.com/gin-gonic/gin"
)

type CategoriesController struct {
	CategoryService *categories.CategoryService
}

var CategoriesCtrl = instance()
var cancelTime = 100 * time.Second

func instance() *CategoriesController {
	return &CategoriesController{CategoryService: categories.Service}
}

func (controller *CategoriesController) GetCategories(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)

	allCategories, err := controller.CategoryService.GetCategories(ctx)

	defer cancel()

	if err != nil {
		log.Println(err.Error())
		common.SendInternalServerResponse(c)

		return
	}

	var response []categories.CategoryResponseResource
	for _, category := range allCategories {
		var resource categories.CategoryResponseResource
		response = append(response, resource.CreateCategoryResponseFrom(category))
	}

	c.JSON(http.StatusOK, response)
}

func (controller *CategoriesController) Create(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)

	var input categories.CreateCategoryResource
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Println(err)
		defer cancel()

		return
	}

	result, err := controller.CategoryService.Create(ctx, input)
	defer cancel()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error.code.internal"})
		return
	}

	var response categories.CategoryResponseResource
	response = response.CreateCategoryResponseFrom(result)
	c.JSON(http.StatusOK, response)
}
