package models

type CreateUserResource struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	RepeatPassword string `json:"repeatPassword" validate:"required,min=8"`
}

type LoginResource struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	RememberMe bool   `json:"rememberMe"`
}

type ForgotPasswordResource struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordResource struct {
	Token          string `json:"token" validate:"required"`
	Password       string `json:"password" validate:"required,min=8"`
	RepeatPassword string `json:"repeatPassword" validate:"required,min=8"`
}
