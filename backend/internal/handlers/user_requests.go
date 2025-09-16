package handlers

import "github.com/gabrielnakaema/project-chat/internal/validator"

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *CreateUserRequest) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(req.Name))
	v.Check("email", "email is required", validator.NotBlank(req.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(req.Email))
	v.Check("password", "password is required", validator.NotBlank(req.Password))
	v.Check("password", "password must be at least 6 characters", validator.MinLength(req.Password, 6))
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginRequest) Validate(v *validator.Validator) {
	v.Check("email", "email is required", validator.NotBlank(req.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(req.Email))
	v.Check("password", "password is required", validator.NotBlank(req.Password))
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (req *RefreshTokenRequest) Validate(v *validator.Validator) {
	v.Check("refresh_token", "refresh token is required", validator.NotBlank(req.RefreshToken))
}
