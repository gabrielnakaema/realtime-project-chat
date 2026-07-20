package user

import "github.com/gabrielnakaema/project-chat/internal/validator"

type CreateUserBody struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *CreateUserBody) Validate(v *validator.Validator) {
	v.Check("name", "name is required", validator.NotBlank(req.Name))
	v.Check("email", "email is required", validator.NotBlank(req.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(req.Email))
	v.Check("password", "password is required", validator.NotBlank(req.Password))
	v.Check("password", "password must be at least 6 characters", validator.MinLength(req.Password, 6))
}

type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req *LoginBody) Validate(v *validator.Validator) {
	v.Check("email", "email is required", validator.NotBlank(req.Email))
	v.Check("email", "email is invalid", validator.ValidEmail(req.Email))
	v.Check("password", "password is required", validator.NotBlank(req.Password))
}

type RefreshTokenBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (req *RefreshTokenBody) Validate(v *validator.Validator) {
	v.Check("refresh_token", "refresh token is required", validator.NotBlank(req.RefreshToken))
}

type ChangePasswordBody struct {
	OldPassword             string `json:"old_password"`
	NewPassword             string `json:"new_password"`
	NewPasswordConfirmation string `json:"new_password_confirmation"`
}

func (req *ChangePasswordBody) Validate(v *validator.Validator) {
	v.Check("old_password", "old password is required", validator.NotBlank(req.OldPassword))
	v.Check("new_password", "new password is required", validator.NotBlank(req.NewPassword))
	v.Check("new_password", "password must be at least 6 characters", validator.MinLength(req.NewPassword, 6))
	v.Check(
		"new_password_confirmation",
		"new password confirmation must match new password",
		req.NewPassword == req.NewPasswordConfirmation,
	)
}
