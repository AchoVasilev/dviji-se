package handlers

import (
	"net/http"
	"server/internal/application/categories"
	"server/util"
	"server/web/templates"
)

type DefaultHandler struct {
	categoryService *categories.CategoryService
}

func NewDefaultHandler(categoryService *categories.CategoryService) *DefaultHandler {
	return &DefaultHandler{
		categoryService: categoryService,
	}
}

func (handler *DefaultHandler) HandleHomePage(writer http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		handler.HandleNotFound(writer, req)
		return
	}

	util.Must(templates.Layout(templates.Home(), "Home", "/").Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleNotFound(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	util.Must(templates.Layout(templates.NotFound(), "Not found", "/not-found").Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleError(writer http.ResponseWriter, req *http.Request) {
	requestId := req.Context().Value("requestId").(string)
	writer.WriteHeader(http.StatusInternalServerError)
	util.Must(templates.Layout(templates.Error(requestId), "Error", "/error").Render(req.Context(), writer))
}
