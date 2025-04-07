package auth

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	jwt.Claims
}
