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
	userService          *users.UserService
	authService          *auth.AuthService
	passwordResetService *auth.PasswordResetService
}

func NewAuthHandler(
	userService *users.UserService,
	authService *auth.AuthService,
	passwordResetService *auth.PasswordResetService,
) *AuthHandler {
	return &AuthHandler{
		userService:          userService,
		authService:          authService,
		passwordResetService: passwordResetService,
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

	tokenResult, err := handler.authService.Authenticate(user, input.Password, input.RememberMe, ctx)
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

	httputils.SetAuthCookie(httputils.AuthCookieName, tokenResult.Token, tokenResult.TokenTime, input.RememberMe, writer)
	writer.Header().Set("HX-Redirect", "/")
	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) RefreshToken(writer http.ResponseWriter, req *http.Request) {
	return
}

func (handler *AuthHandler) GetLogin(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if httputils.IsHTMXRequest(req) {
		util.Must(templates.Login().Render(ctx, writer))
		return
	}

	util.Must(templates.SimpleLayout(
		templates.LoginRegister(templates.Login()),
		"Вход",
		"Влезте в профила си и продължете към здравословен начин на живот.",
		ctxutils.GetCSRF(ctx),
	).Render(ctx, writer))
}

func (handler *AuthHandler) GetRegister(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if httputils.IsHTMXRequest(req) {
		util.Must(templates.Register().Render(ctx, writer))
		return
	}

	util.Must(templates.SimpleLayout(
		templates.LoginRegister(templates.Register()),
		"Регистрация",
		"Създайте профил и започнете пътя към по-здравословен живот.",
		ctxutils.GetCSRF(ctx),
	).Render(ctx, writer))
}

func (handler *AuthHandler) GetForgotPassword(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	util.Must(templates.SimpleLayout(
		templates.ForgotPassword(),
		"Забравена парола",
		"Възстановете достъпа до вашия акаунт.",
		ctxutils.GetCSRF(ctx),
	).Render(ctx, writer))
}

func (handler *AuthHandler) HandleForgotPassword(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	input := new(models.ForgotPasswordResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.Error(result.ParsingError.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if result.ValidationErrors != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		util.Must(templates.FormErrors(result.ValidationErrors).Render(ctx, writer))
		return
	}

	// Always return success to prevent email enumeration
	err := handler.passwordResetService.RequestReset(ctx, input.Email)
	if err != nil {
		slog.Error("Failed to process password reset request", "error", err)
		// Still return success to prevent enumeration
	}

	util.Must(templates.ForgotPasswordSuccess().Render(ctx, writer))
}

func (handler *AuthHandler) GetResetPassword(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	token := req.URL.Query().Get("token")

	if token == "" {
		http.Redirect(writer, req, "/forgot-password", http.StatusSeeOther)
		return
	}

	// Validate token without using it
	valid, err := handler.passwordResetService.ValidateToken(ctx, token)
	if err != nil || !valid {
		util.Must(templates.SimpleLayout(
			templates.ResetPasswordInvalid(),
			"Невалиден линк",
			"Линкът за смяна на парола е невалиден или изтекъл.",
			ctxutils.GetCSRF(ctx),
		).Render(ctx, writer))
		return
	}

	util.Must(templates.SimpleLayout(
		templates.ResetPassword(token),
		"Нова парола",
		"Задайте нова парола за вашия акаунт.",
		ctxutils.GetCSRF(ctx),
	).Render(ctx, writer))
}

func (handler *AuthHandler) HandleResetPassword(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	input := new(models.ResetPasswordResource)
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
		util.Must(templates.FormErrors(result.ValidationErrors).Render(ctx, writer))
		return
	}

	err := handler.passwordResetService.ResetPassword(ctx, input.Token, input.Password)
	if err == auth.ErrInvalidToken {
		writer.WriteHeader(http.StatusBadRequest)
		util.Must(templates.InvalidMessage("Линкът е невалиден или изтекъл. Моля, заявете нов.", "error-token").Render(ctx, writer))
		return
	}
	if err == auth.ErrPasswordWeak {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		util.Must(templates.InvalidMessage("Паролата трябва да е поне 8 символа", "error-password").Render(ctx, writer))
		return
	}
	if err != nil {
		slog.Error("Failed to reset password", "error", err)
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	writer.Header().Set("HX-Redirect", "/login")
	writer.WriteHeader(http.StatusOK)
}
