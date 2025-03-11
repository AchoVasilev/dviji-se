package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/application/categories"
	"server/internal/http/handlers/models"
	"server/util/httputils"
	"time"
)

type CategoriesHandler struct {
	categoryService *categories.CategoryService
}

var cancelTime = 10 * time.Second

func NewCategoriesHandler(categoryService *categories.CategoryService) *CategoriesHandler {
	return &CategoriesHandler{categoryService: categoryService}
}

func (controller *CategoriesHandler) GetCategories(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	allCategories, err := controller.categoryService.GetCategories(ctx)
	if err != nil {
		slog.Error(err.Error())
		httputils.SendInternalServerResponse(writer, req)

		return
	}

	var response []models.CategoryResponseResource
	for _, category := range allCategories {
		var resource models.CategoryResponseResource
		response = append(response, resource.CreateCategoryResponseFrom(&category))
	}

	httputils.SendOkWithBody(writer, response)
}

func (controller *CategoriesHandler) Create(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Creating a new category")
	ctx, cancel := context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	var input models.CreateCategoryResource
	success := httputils.ProcessRequestBody(writer, req, &input)
	if !success {
		return
	}

	result, err := controller.categoryService.Create(ctx, input)
	if err != nil {
		slog.Error(err.Error())
		httputils.SendInternalServerResponse(writer, req)
		return
	}

	slog.Info(fmt.Sprintf("Successfully created a new category. [id=%s]", result.Id.String()))

	var response models.CategoryResponseResource
	response = response.CreateCategoryResponseFrom(result)
	httputils.SendOkWithBody(writer, response)
}
