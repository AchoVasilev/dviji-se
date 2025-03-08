package models

type CreateUserResource struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	RepeatPassword string `json:"repeatPassword" validate:"required,min=6"`
}

type LoginResource struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required.min=6"`
}
