package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/application/users"
	"server/internal/http/handlers/models"
	"server/internal/http/middleware"
	"server/util"
	"server/util/httputils"
	"server/web/templates"
)

type AuthHandler struct {
	userService *users.UserService
}

func NewAuthHandler(userService *users.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (handler *AuthHandler) HandleRegister(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Registering new user")

	ctx, cancel := context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	input := new(models.CreateUserResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.Error(result.ParsingError.Error())
		http.Error(writer, "internal.server.error", http.StatusInternalServerError)
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if input.Password != input.RepeatPassword {
		result.ValidationErrors = append(result.ValidationErrors, &httputils.ValidationError{
			Value: "",
			Field: "repeatPassword",
			Error: "Паролите не съвпадат",
		})
	}

	if result.ValidationErrors != nil {
		util.Must(templates.FormErrors(result.ValidationErrors).Render(req.Context(), writer))
		return
	}

	user, err := handler.userService.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.Error(err.Error())
		httputils.SendInternalServerResponse(writer, req)
		return
	}
	slog.Info(user.Email)
	//	if user != nil {
	//		httputils.SendConflictResponse(writer, "User exists")
	//		return
	//	}

	id, err := handler.userService.RegisterUser(input)
	if err != nil {
		httputils.SendInternalServerResponse(writer, req)
		return
	}

	slog.Info(fmt.Sprintf("User successfully created. [id=%s]", id.String()))
	httputils.SendCreatedAt(writer, fmt.Sprintf("/users/%s", id.String()))
}

func (handler *AuthHandler) HandleLogin(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Handling user login")

	ctx, cancel := context.WithTimeout(context.Background(), cancelTime)
	defer cancel()

	input := new(models.LoginResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.Error(result.ParsingError.Error())
		http.Error(writer, "internal.server.error", http.StatusInternalServerError)
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if result.ValidationErrors != nil {
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	user, err := handler.userService.GetUserByEmail(ctx, input.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.Error(err.Error())
		httputils.SendInternalServerResponse(writer, req)
		return
	}

	slog.Info(user.Email)

	//	if user == nil {
	//		slog.Info("Success")
	//	}
}

func (handler *AuthHandler) RefreshToken(writer http.ResponseWriter, req *http.Request) {
	return
}

func (handler *AuthHandler) GetLoginLayout(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.SimpleLayout(templates.LoginRegister(templates.Login()), "Вход", middleware.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *AuthHandler) GetLogin(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Login().Render(req.Context(), writer))
}

func (handler *AuthHandler) GetRegisterLayour(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.SimpleLayout(templates.LoginRegister(templates.Register()), "Регистрация", middleware.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *AuthHandler) GetRegister(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Register().Render(req.Context(), writer))
}
