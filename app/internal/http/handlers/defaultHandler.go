package handlers

import (
	"net/http"
	"server/util"
	"server/web/templates"
)

type DefaultHandler struct {
	// categoryService *categories.CategoryService
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (handler *DefaultHandler) HandleHomePage(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Layout(templates.Home(), "Home", "/").Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleNotFound(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	util.Must(templates.Layout(templates.NotFound(), "Not found", "/not-found").Render(req.Context(), writer))
}
