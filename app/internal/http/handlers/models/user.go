package models

type CreateUserResource struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,strongpassword"`
	RepeatPassword string `json:"repeatPassword" validate:"required"`
}

// LoginResource intentionally applies no strength rules: accounts created
// before the current policy must still be able to sign in.
type LoginResource struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"rememberMe"`
}

type ForgotPasswordResource struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordResource struct {
	Token          string `json:"token" validate:"required"`
	Password       string `json:"password" validate:"required,strongpassword"`
	RepeatPassword string `json:"repeatPassword" validate:"required"`
}
