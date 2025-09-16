package handlers

import (
	"context"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

var (
	RefreshTokenCookieName = "project_chat_refresh_token"
)

type userService interface {
	Create(context.Context, service.CreateUserRequest) (*domain.User, error)
	Login(context.Context, service.LoginRequest) (*service.LoginResult, error)
	RefreshToken(context.Context, service.RefreshTokenRequest) (*service.LoginResult, error)
	GetMe(context.Context, uuid.UUID) (*domain.User, error)
}

type UserHandler struct {
	userService userService
}

func NewUserHandler(userService userService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (uh *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateUserRequest

	err := utils.ReadJSON(w, r, &request)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)

	if !v.Valid() {
		ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := service.CreateUserRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	}

	user, err := uh.userService.Create(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, user, nil)
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        domain.User `json:"user"`
}

func loginResultToResponse(result *service.LoginResult) LoginResponse {
	response := LoginResponse{
		AccessToken: result.AccessToken,
		User:        *result.User,
	}

	return response
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	err := utils.ReadJSON(w, r, &request)
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := service.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	result, err := uh.userService.Login(r.Context(), serviceRequest)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	refreshToken := result.RefreshToken

	setRefreshTokenCookie(w, refreshToken)

	utils.WriteJSON(w, http.StatusOK, loginResultToResponse(result), nil)
}

func (uh *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		UnauthorizedResponse(w, "refresh token not found")
		return
	}

	if refreshToken == nil {
		UnauthorizedResponse(w, "refresh token not found")
		return
	}

	serviceRequest := service.RefreshTokenRequest{
		Token: refreshToken.Value,
	}

	result, err := uh.userService.RefreshToken(r.Context(), serviceRequest)
	if err != nil {
		setRefreshTokenCookie(w, "")
		ErrorResponse(w, r, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken)

	utils.WriteJSON(w, http.StatusOK, loginResultToResponse(result), nil)
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {
	// TODO: Manage production domains and secure cookies
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (uh *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId := UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		UnauthorizedResponse(w, "unauthorized")
		return
	}

	user, err := uh.userService.GetMe(r.Context(), userId)
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, user, nil)
}
