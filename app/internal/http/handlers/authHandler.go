package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/application/auth"
	"server/internal/application/users"
	"server/internal/http/handlers/models"
	"server/util"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/web/templates"
)

type AuthHandler struct {
	userService *users.UserService
	authService *auth.AuthService
}

func NewAuthHandler(userService *users.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

func (handler *AuthHandler) HandleRegister(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Registering new user")

	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	input := new(models.CreateUserResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.Error(result.ParsingError.Error())
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
		writer.WriteHeader(http.StatusUnprocessableEntity)
		util.Must(templates.FormErrors(result.ValidationErrors).Render(req.Context(), writer))
		return
	}

	exists, err := handler.userService.ExistsByEmail(ctx, input.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.Error(err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if exists {
		slog.Info(fmt.Sprintf("Attempt to register with existing user. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusConflict)
		util.Must(templates.InvalidMessage("Потребител с този имейл съществува", "error-email").Render(req.Context(), writer))
		return
	}

	id, err := handler.userService.RegisterUser(input)
	if err != nil {
		slog.Error(err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	slog.Info(fmt.Sprintf("User successfully created. [id=%s]", id.String()))
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("HX-Redirect", "/")
}

func (handler *AuthHandler) HandleLogin(writer http.ResponseWriter, req *http.Request) {
	slog.Info("Handling user login")

	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
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
		writer.WriteHeader(http.StatusUnprocessableEntity)
		util.Must(templates.FormErrors(result.ValidationErrors).Render(ctx, writer))
		return
	}

	user, err := handler.userService.GetUserByEmail(ctx, input.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.Error(err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if err == sql.ErrNoRows || user.Email == "" {
		slog.Info(fmt.Sprintf("Attempt to login with invalid credentials. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusNotFound)
		util.Must(templates.InvalidMessage("Невалиден имейл или парола", "error-email").Render(ctx, writer))
		return
	}

	// remember me func

	tokenResult, err := handler.authService.Authenticate(user, input.Password, ctx)
	if err == auth.ErrHashNotMatch {
		slog.Info(fmt.Sprintf("Attempt to login with invalid credentials. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusNotFound)
		util.Must(templates.InvalidMessage("Невалиден имейл или парола", "error-email").Render(ctx, writer))
		return
	}

	if err != nil {
		slog.Error(err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	httputils.SetHttpOnlyCookie(httputils.AuthCookieName, tokenResult.Token, tokenResult.TokenTime, writer)
	writer.Header().Set("HX-Redirect", "/")
	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) RefreshToken(writer http.ResponseWriter, req *http.Request) {
	return
}

func (handler *AuthHandler) GetLoginLayout(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.SimpleLayout(templates.LoginRegister(templates.Login()), "Вход", ctxutils.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *AuthHandler) GetLogin(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Login().Render(req.Context(), writer))
}

func (handler *AuthHandler) GetRegisterLayout(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.SimpleLayout(templates.LoginRegister(templates.Register()), "Регистрация", ctxutils.GetCSRF(req.Context())).Render(req.Context(), writer))
}

func (handler *AuthHandler) GetRegister(writer http.ResponseWriter, req *http.Request) {
	util.Must(templates.Register().Render(req.Context(), writer))
}
