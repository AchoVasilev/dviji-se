package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"server/internal/application/auth"
	"server/internal/application/users"
	"server/internal/http/handlers/models"
	"server/util"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"server/web/templates"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	slog.InfoContext(ctx, "Registering new user")

	input := new(models.CreateUserResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.ErrorContext(ctx, result.ParsingError.Error())
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.ErrorContext(ctx, err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if exists {
		slog.InfoContext(ctx, fmt.Sprintf("Attempt to register with existing user. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusConflict)
		util.Must(templates.InvalidMessage("Потребител с този имейл съществува", "error-email").Render(req.Context(), writer))
		return
	}

	id, err := handler.userService.RegisterUser(ctx, input)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	slog.InfoContext(ctx, fmt.Sprintf("User successfully created. [id=%s]", id.String()))
	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("HX-Redirect", "/")
}

func (handler *AuthHandler) HandleLogin(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	slog.InfoContext(ctx, "Handling user login")

	input := new(models.LoginResource)
	result := httputils.ProcessBody(writer, req, input)
	if result.ParsingError != nil {
		slog.ErrorContext(ctx, result.ParsingError.Error())
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.ErrorContext(ctx, err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	if errors.Is(err, sql.ErrNoRows) || user.Email == "" {
		slog.InfoContext(ctx, fmt.Sprintf("Attempt to login with invalid credentials. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusNotFound)
		util.Must(templates.InvalidMessage("Невалиден имейл или парола", "error-email").Render(ctx, writer))
		return
	}

	tokenResult, err := handler.authService.Authenticate(user, input.Password, input.RememberMe, ctx)
	if errors.Is(err, auth.ErrHashNotMatch) {
		slog.InfoContext(ctx, fmt.Sprintf("Attempt to login with invalid credentials. [email=%s]", input.Email))
		writer.WriteHeader(http.StatusNotFound)
		util.Must(templates.InvalidMessage("Невалиден имейл или парола", "error-email").Render(ctx, writer))
		return
	}

	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	httputils.SetAuthCookie(httputils.AuthCookieName, tokenResult.Token, tokenResult.TokenTime, input.RememberMe, writer)
	httputils.SetRefreshCookie(tokenResult.RefreshToken, tokenResult.RefreshTokenTime, writer)

	redirect := "/"
	if strings.HasPrefix(req.URL.Path, "/admin") {
		redirect = "/admin"
	}
	writer.Header().Set("HX-Redirect", redirect)
	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) HandleLogout(writer http.ResponseWriter, req *http.Request) {
	handler.clearSession(writer)
	httputils.ClearCookie(httputils.XSRFCookieName, writer)

	writer.Header().Set("HX-Redirect", "/")
	writer.WriteHeader(http.StatusOK)
}

// RefreshToken exchanges a valid refresh token for a fresh access token. The
// user is reloaded from the database rather than trusted from the token, so
// role changes, deletions and revocations take effect on refresh.
func (handler *AuthHandler) RefreshToken(writer http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), cancelTime)
	defer cancel()

	refreshCookie, err := req.Cookie(string(httputils.RefreshCookieName))
	if err != nil || refreshCookie.Value == "" {
		httputils.SendErrorResponse(ctx, writer, "refresh.token.missing", http.StatusUnauthorized)
		return
	}

	token, err := securityutil.ValidateRefreshToken(refreshCookie.Value)
	if err != nil {
		slog.InfoContext(ctx, "Rejected an invalid refresh token", "error", err)
		handler.clearSession(writer)
		httputils.SendErrorResponse(ctx, writer, "refresh.token.invalid", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		handler.clearSession(writer)
		httputils.SendErrorResponse(ctx, writer, "refresh.token.invalid", http.StatusUnauthorized)
		return
	}

	userId, ok := claims["id"].(string)
	if !ok || userId == "" {
		handler.clearSession(writer)
		httputils.SendErrorResponse(ctx, writer, "refresh.token.invalid", http.StatusUnauthorized)
		return
	}

	// A refresh token is only as good as the account behind it.
	currentUser, err := handler.userService.GetUserById(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			handler.clearSession(writer)
			httputils.SendErrorResponse(ctx, writer, "refresh.token.invalid", http.StatusUnauthorized)
			return
		}

		slog.ErrorContext(ctx, "Could not load the user while refreshing", "error", err, "userId", userId)
		httputils.SendInternalServerResponse(writer, req)
		return
	}

	issuedAt := time.Unix(int64(claimFloat(claims, "iat")), 0).UTC()
	if !handler.userService.IsSessionValid(ctx, userId, issuedAt) {
		slog.InfoContext(ctx, "Rejected a revoked refresh token", "userId", userId)
		handler.clearSession(writer)
		httputils.SendErrorResponse(ctx, writer, "refresh.token.revoked", http.StatusUnauthorized)
		return
	}

	accessToken, accessExpiry := securityutil.GenerateAccessToken(currentUser, false)
	httputils.SetAuthCookie(httputils.AuthCookieName, accessToken, accessExpiry, false, writer)

	slog.InfoContext(ctx, "Refreshed an access token", "userId", userId)
	httputils.SendSuccessResponse(ctx, writer, "Token refreshed", nil, http.StatusOK)
}

// clearSession drops both session cookies. The refresh cookie is path scoped,
// so it has to be cleared on that same path.
func (handler *AuthHandler) clearSession(writer http.ResponseWriter) {
	httputils.ClearCookie(httputils.AuthCookieName, writer)
	httputils.ClearCookieAtPath(httputils.RefreshCookieName, httputils.RefreshTokenPath, writer)
}

func claimFloat(claims jwt.MapClaims, key string) float64 {
	value, _ := claims[key].(float64)

	return value
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

func (handler *AuthHandler) GetAdminLogin(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	util.Must(templates.SimpleLayout(
		templates.LoginRegister(templates.AdminLogin()),
		"Вход",
		"Администраторски вход.",
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
		slog.ErrorContext(ctx, result.ParsingError.Error())
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
		slog.ErrorContext(ctx, "Failed to process password reset request", "error", err)
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
		slog.ErrorContext(ctx, result.ParsingError.Error())
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
	if errors.Is(err, auth.ErrInvalidToken) {
		writer.WriteHeader(http.StatusBadRequest)
		util.Must(templates.InvalidMessage("Линкът е невалиден или изтекъл. Моля, заявете нов.", "error-token").Render(ctx, writer))
		return
	}
	if errors.Is(err, auth.ErrPasswordWeak) {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		util.Must(templates.InvalidMessage("Паролата трябва да е поне 12 символа и да съдържа главна и малка буква, цифра и символ", "error-password").Render(ctx, writer))
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "Failed to reset password", "error", err)
		writer.Header().Add("HX-Redirect", "/error")
		return
	}

	writer.Header().Set("HX-Redirect", "/login")
	writer.WriteHeader(http.StatusOK)
}
