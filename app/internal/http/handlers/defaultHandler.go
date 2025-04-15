package handlers

import (
	"net/http"
	"server/internal/application/categories"
	"server/util"
	"server/util/ctxutils"
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

	util.Must(templates.Layout(templates.Home(), "Home", "/", ctxutils.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleNotFound(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	util.Must(templates.Layout(templates.NotFound(), "Not found", "/not-found", ctxutils.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleError(writer http.ResponseWriter, req *http.Request) {
	requestId := ctxutils.RequestIdFromContext(req.Context())
	writer.WriteHeader(http.StatusInternalServerError)
	util.Must(templates.Layout(templates.Error(requestId), "Error", "/error", ctxutils.GetCSRF(req.Context())).Render(req.Context(), writer))
}
