package rest

import (
	"context"
	"log/slog"
	"net/http"
	"server/application/categories"
	"server/common/api"
	"time"
)

type CategoriesController struct {
	categoryService *categories.CategoryService
}

var cancelTime = 10 * time.Second

func NewCategoriesController(categoryService *categories.CategoryService) *CategoriesController {
	return &CategoriesController{categoryService: categoryService}
}

func (controller *CategoriesController) GetCategories(writer http.ResponseWriter, req *http.Request) {
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	allCategories, err := controller.categoryService.GetCategories(ctx)

	if err != nil {
		slog.Error(err.Error())
		api.SendInternalServerResponse(writer, req)

		return
	}

	var response []categories.CategoryResponseResource
	for _, category := range allCategories {
		var resource categories.CategoryResponseResource
		response = append(response, resource.CreateCategoryResponseFrom(category))
	}

	api.SendOkWithBody(writer, response)
}

func (controller *CategoriesController) Create(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Creating a new category")
	var ctx, cancel = context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	var input categories.CreateCategoryResource
	success := api.ProcessRequestBody(writer, req, &input)
	if !success {
		return
	}

	result, err := controller.categoryService.Create(ctx, input)

	if err != nil {
		slog.Error(err.Error())
		api.SendInternalServerResponse(writer, req)
		return
	}

	slog.Info("Successfully created a new category", slog.String("id", result.Id.String()))

	var response categories.CategoryResponseResource
	response = response.CreateCategoryResponseFrom(result)
	api.SendOkWithBody(writer, response)
}
