package routes

import (
	"database/sql"
	"net/http"
	"server/internal/application/users"
	"server/internal/domain/user"
	"server/internal/http/handlers"
)

func AuthRoutes(mux *http.ServeMux, db *sql.DB) {
	userRepository := user.NewUserRepository(db)
	userService := users.NewUserService(userRepository)
	authHandler := handlers.NewAuthHandler(userService)

	mux.HandleFunc("POST /register", authHandler.HandleRegister)

	mux.HandleFunc("POST /login", authHandler.HandleLogin)

	mux.HandleFunc("POST /refresh-token", authHandler.RefreshToken)
}