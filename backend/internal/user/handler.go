package user

import (
	"context"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/httperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/google/uuid"
)

var (
	RefreshTokenCookieName = "project_chat_refresh_token"
)

type userService interface {
	Create(context.Context, CreateUserRequest) (*domain.User, error)
	Login(context.Context, LoginRequest) (*LoginResult, error)
	RefreshToken(context.Context, RefreshTokenRequest) (*LoginResult, error)
	GetMe(context.Context, uuid.UUID) (*domain.User, error)
	ChangePassword(context.Context, ChangePasswordRequest) error
	Logout(context.Context, uuid.UUID, string) error
	ListUsers(context.Context, uuid.UUID) ([]domain.User, error)
}

type UserHandler struct {
	userService userService
	config      *config.Config
}

func NewUserHandler(userService userService, config *config.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		config:      config,
	}
}

func (uh *UserHandler) isProduction() bool {
	return uh.config.Environment == "production"
}

func (uh *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateUserBody

	err := platformhttp.ReadJSON(w, r, &request)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)

	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := CreateUserRequest{
		Name:     request.Name,
		Email:    request.Email,
		Password: request.Password,
	}

	user, err := uh.userService.Create(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, user, nil)
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        domain.User `json:"user"`
}

func loginResultToResponse(result *LoginResult) LoginResponse {
	response := LoginResponse{
		AccessToken: result.AccessToken,
		User:        *result.User,
	}

	return response
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginBody

	err := platformhttp.ReadJSON(w, r, &request)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	serviceRequest := LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	result, err := uh.userService.Login(r.Context(), serviceRequest)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	refreshToken := result.RefreshToken

	setRefreshTokenCookie(w, refreshToken, uh.isProduction())

	platformhttp.WriteJSON(w, http.StatusOK, loginResultToResponse(result), nil)
}

func (uh *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		httperr.UnauthorizedResponse(w, "refresh token not found")
		return
	}

	if refreshToken == nil {
		httperr.UnauthorizedResponse(w, "refresh token not found")
		return
	}

	serviceRequest := RefreshTokenRequest{
		Token: refreshToken.Value,
	}

	result, err := uh.userService.RefreshToken(r.Context(), serviceRequest)
	if err != nil {
		setRefreshTokenCookie(w, "", uh.isProduction())
		httperr.ErrorResponse(w, r, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken, uh.isProduction())

	platformhttp.WriteJSON(w, http.StatusOK, loginResultToResponse(result), nil)
}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		clearRefreshTokenCookie(w, uh.isProduction())
		platformhttp.WriteJSON(w, http.StatusOK, nil, nil)
		return
	}

	refreshToken, err := r.Cookie(RefreshTokenCookieName)
	if err != nil {
		clearRefreshTokenCookie(w, uh.isProduction())
		platformhttp.WriteJSON(w, http.StatusOK, nil, nil)
		return
	}

	if refreshToken == nil {
		clearRefreshTokenCookie(w, uh.isProduction())
		platformhttp.WriteJSON(w, http.StatusOK, nil, nil)
		return
	}

	log := logger.FromContext(r.Context())

	err = uh.userService.Logout(r.Context(), userId, refreshToken.Value)
	if err != nil {
		log.Error("error while logging out", "error", err, "user_id", userId, "refresh_token", refreshToken.Value)
	}

	clearRefreshTokenCookie(w, uh.isProduction())
	platformhttp.WriteJSON(w, http.StatusOK, nil, nil)
}

func clearRefreshTokenCookie(w http.ResponseWriter, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string, isProduction bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(domain.RefreshTokenTTL),
	})
}

func (uh *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	user, err := uh.userService.GetMe(r.Context(), userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, user, nil)
}

func (uh *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	var request ChangePasswordBody

	err := platformhttp.ReadJSON(w, r, &request)
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	err = uh.userService.ChangePassword(r.Context(), ChangePasswordRequest{
		UserID:                  userId,
		OldPassword:             request.OldPassword,
		NewPassword:             request.NewPassword,
		NewPasswordConfirmation: request.NewPasswordConfirmation,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, nil, nil)
}

func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	userId := auth.UserIdFromContext(r.Context())
	if userId == uuid.Nil {
		httperr.UnauthorizedResponse(w, "unauthorized")
		return
	}

	users, err := uh.userService.ListUsers(r.Context(), userId)
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, users, nil)
}
