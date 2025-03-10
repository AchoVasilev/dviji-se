package handlers

import (
	"log/slog"
	"net/http"
	"server/web/templates"
)

type DefaultHandler struct {
	// categoryService *categories.CategoryService
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (handler *DefaultHandler) HandleHomePage(writer http.ResponseWriter, req *http.Request) {
  err := templates.Layout(nil, "Home").Render(req.Context(), writer)
  if err != nil {
    slog.Info(err.Error())
  }
}
