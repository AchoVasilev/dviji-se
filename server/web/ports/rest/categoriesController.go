package rest

import (
	"context"
	"log"
	"net/http"
	"server/application/categories"
	"server/common/api"
	"time"
)

type CategoriesController struct {
	CategoryService *categories.CategoryService
}

var CategoriesCtrl = instance()
var cancelTime = 10 * time.Second

func instance() *CategoriesController {
	return &CategoriesController{CategoryService: categories.Service}
}

func (controller *CategoriesController) GetCategories(writer http.ResponseWriter, req *http.Request) {
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)

	allCategories, err := controller.CategoryService.GetCategories(ctx)

	defer cancel()

	if err != nil {
		log.Println(err.Error())
		api.SendInternalServerResponse(writer)

		return
	}

	var response []categories.CategoryResponseResource
	for _, category := range allCategories {
		var resource categories.CategoryResponseResource
		response = append(response, resource.CreateCategoryResponseFrom(category))
	}

	api.SendOkWithBody(writer, response)
}

func (controller *CategoriesController) Create(writer http.ResponseWriter, req *http.Request, input *categories.CreateCategoryResource) {
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)

	result, err := controller.CategoryService.Create(ctx, input)
	defer cancel()

	if err != nil {
		api.SendInternalServerResponse(writer)
		return
	}

	var response categories.CategoryResponseResource
	response = response.CreateCategoryResponseFrom(result)
	api.SendOkWithBody(writer, response)
}
