package handlers

import (
	"net/http"
	"server/internal/application/categories"
	"server/internal/config"
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

	util.Must(templates.LayoutSEO(
		templates.Home(),
		templates.SEO{
			Title:       "Начало",
			Description: "Тренировки, хранителни режими и рецепти за здравословен начин на живот.",
			Path:        "/",
		},
		"/",
		ctxutils.GetCSRF(req.Context()),
		config.AllowRegistration(),
	).Render(req.Context(), writer))
}

// HandleAbout renders the about page.
func (handler *DefaultHandler) HandleAbout(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.About().Render(req.Context(), writer))
}

// HandlePrivacy renders the privacy policy.
func (handler *DefaultHandler) HandlePrivacy(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Privacy().Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleNotFound(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	util.Must(templates.Layout(
		templates.NotFound(),
		"Страницата не е намерена",
		"Страницата, която търсите, не съществува.",
		"/not-found",
		ctxutils.GetCSRF(req.Context()),
		config.AllowRegistration(),
	).Render(req.Context(), writer))
}

func (handler *DefaultHandler) HandleError(writer http.ResponseWriter, req *http.Request) {
	requestId := ctxutils.RequestIdFromContext(req.Context())
	writer.WriteHeader(http.StatusInternalServerError)
	util.Must(templates.Layout(
		templates.Error(requestId),
		"Грешка",
		"Възникна неочаквана грешка.",
		"/error",
		ctxutils.GetCSRF(req.Context()),
		config.AllowRegistration(),
	).Render(req.Context(), writer))
}
