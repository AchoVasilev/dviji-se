package routes

import (
	"database/sql"
	"net/http"
	"server/internal/application/auth"
	"server/internal/application/users"
	"server/internal/config"
	"server/internal/domain/user"
	"server/internal/http/handlers"
	"server/internal/http/middleware"
	"server/internal/infrastructure/email"
)

func AuthRoutes(mux *http.ServeMux, db *sql.DB) {
	userRepository := user.NewUserRepository(db)
	tokenRepository := user.NewPasswordResetTokenRepository(db)
	userService := users.NewUserService(userRepository)
	authService := auth.NewAuthService()
	emailService := email.NewEmailService()
	passwordResetService := auth.NewPasswordResetService(userRepository, tokenRepository, emailService)

	authHandler := handlers.NewAuthHandler(userService, authService, passwordResetService)

	// Rate limiters
	authLimiter := middleware.AuthRateLimiter()
	passwordResetLimiter := middleware.PasswordResetRateLimiter()

	// Page routes (no rate limiting needed)
	mux.HandleFunc("GET /login", authHandler.GetLogin)
	mux.HandleFunc("GET /forgot-password", authHandler.GetForgotPassword)
	mux.HandleFunc("GET /reset-password", authHandler.GetResetPassword)

	// Registration routes (only when enabled)
	if config.AllowRegistration() {
		mux.HandleFunc("GET /register", authHandler.GetRegister)
		mux.Handle("POST /register", authLimiter.Middleware(http.HandlerFunc(authHandler.HandleRegister)))
	}

	// API routes with rate limiting
	mux.Handle("POST /login", authLimiter.Middleware(http.HandlerFunc(authHandler.HandleLogin)))
	mux.Handle("POST /refresh-token", authLimiter.Middleware(http.HandlerFunc(authHandler.RefreshToken)))
	mux.Handle("POST /forgot-password", passwordResetLimiter.Middleware(http.HandlerFunc(authHandler.HandleForgotPassword)))
	mux.Handle("POST /reset-password", passwordResetLimiter.Middleware(http.HandlerFunc(authHandler.HandleResetPassword)))
}
