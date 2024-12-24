package handlers

import (
	"net/http"
	"server/internal/application/users"
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

}

func (handler *AuthHandler) HandleLogin(writer http.ResponseWriter, req *http.Request) {

}

func (handler *AuthHandler) RefreshToken(writer http.ResponseWriter, req *http.Request) {

}
