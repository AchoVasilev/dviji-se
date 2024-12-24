package models

type CreateUserResource struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	RepeatPassword string `json:"repeatPassword" validate:"required"`
}
