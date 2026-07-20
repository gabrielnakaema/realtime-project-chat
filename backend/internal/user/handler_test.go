package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	"github.com/gabrielnakaema/project-chat/internal/platform/config"
	"github.com/gabrielnakaema/project-chat/internal/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, request user.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, request user.LoginRequest) (*user.LoginResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.LoginResult), args.Error(1)
}

func (m *mockUserService) RefreshToken(ctx context.Context, request user.RefreshTokenRequest) (*user.LoginResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.LoginResult), args.Error(1)
}

func (m *mockUserService) GetMe(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) ChangePassword(ctx context.Context, request user.ChangePasswordRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *mockUserService) Logout(ctx context.Context, userId uuid.UUID, token string) error {
	args := m.Called(ctx, userId, token)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockUserService) ListUsers(ctx context.Context, excludeId uuid.UUID) ([]domain.User, error) {
	args := m.Called(ctx, excludeId)
	return args.Get(0).([]domain.User), args.Error(1)
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mockUserService)
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful user creation",
			requestBody: user.CreateUserBody{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.CreateUserRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				}
				expectedUser := &domain.User{
					Id:        uuid.New(),
					Name:      "John Doe",
					Email:     "john@example.com",
					Password:  "hashedpassword",
					CreatedAt: time.Now(),
				}
				mockService.On("Create", mock.Anything, expectedRequest).Return(expectedUser, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var user domain.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				assert.NoError(t, err)
				assert.Equal(t, "John Doe", user.Name)
				assert.Equal(t, "john@example.com", user.Email)
				assert.Empty(t, user.Password)
				assert.NotEmpty(t, user.Id)
			},
		},
		{
			name:        "invalid JSON",
			requestBody: `{"name":"John","email":}`,
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - missing name",
			requestBody: user.CreateUserBody{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - invalid email",
			requestBody: user.CreateUserBody{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - short password",
			requestBody: user.CreateUserBody{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - duplicate email",
			requestBody: user.CreateUserBody{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.CreateUserRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Create", mock.Anything, expectedRequest).Return(nil, apperr.DuplicateEntryError("user email is already taken"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - internal server error",
			requestBody: user.CreateUserBody{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.CreateUserRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Create", mock.Anything, expectedRequest).Return(nil, apperr.ServerError("database error", errors.New("connection failed")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := user.NewUserHandler(mockService, &config.Config{Environment: "development"})

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/users", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	validUser := &domain.User{
		Id:        uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	validLoginResult := &user.LoginResult{
		AccessToken:  "jwt-access-token",
		RefreshToken: "refresh-token",
		User:         validUser,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: user.LoginBody{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.LoginRequest{
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Login", mock.Anything, expectedRequest).Return(validLoginResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response user.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "jwt-access-token", response.AccessToken)
				assert.Equal(t, validUser.Id, response.User.Id)
				assert.Equal(t, validUser.Name, response.User.Name)
				assert.Equal(t, validUser.Email, response.User.Email)
			},
		},
		{
			name:        "invalid JSON",
			requestBody: `{"email":"john@example.com","password":}`,
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - missing email",
			requestBody: user.LoginBody{
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - missing password",
			requestBody: user.LoginBody{
				Email: "john@example.com",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - invalid credentials",
			requestBody: user.LoginBody{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.LoginRequest{
					Email:    "john@example.com",
					Password: "wrongpassword",
				}
				mockService.On("Login", mock.Anything, expectedRequest).Return(nil, apperr.UnauthorizedError("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := user.NewUserHandler(mockService, &config.Config{Environment: "development"})

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/auth/login", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    interface{}
		withUser       bool
		mockSetup      func(*mockUserService)
		expectedStatus int
	}{
		{
			name: "successful password change",
			requestBody: user.ChangePasswordBody{
				OldPassword:             "password123",
				NewPassword:             "newpassword123",
				NewPasswordConfirmation: "newpassword123",
			},
			withUser: true,
			mockSetup: func(mockService *mockUserService) {
				mockService.On("ChangePassword", mock.Anything, user.ChangePasswordRequest{
					UserID:                  userID,
					OldPassword:             "password123",
					NewPassword:             "newpassword123",
					NewPasswordConfirmation: "newpassword123",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized without authenticated user",
			requestBody:    user.ChangePasswordBody{},
			withUser:       false,
			mockSetup:      func(mockService *mockUserService) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{"old_password":"password123","new_password":}`,
			withUser:       true,
			mockSetup:      func(mockService *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - missing old password",
			requestBody: user.ChangePasswordBody{
				NewPassword:             "newpassword123",
				NewPasswordConfirmation: "newpassword123",
			},
			withUser:       true,
			mockSetup:      func(mockService *mockUserService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - short new password",
			requestBody: user.ChangePasswordBody{
				OldPassword:             "password123",
				NewPassword:             "123",
				NewPasswordConfirmation: "123",
			},
			withUser:       true,
			mockSetup:      func(mockService *mockUserService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - confirmation mismatch",
			requestBody: user.ChangePasswordBody{
				OldPassword:             "password123",
				NewPassword:             "newpassword123",
				NewPasswordConfirmation: "differentpassword123",
			},
			withUser:       true,
			mockSetup:      func(mockService *mockUserService) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - incorrect old password",
			requestBody: user.ChangePasswordBody{
				OldPassword:             "wrongpassword",
				NewPassword:             "newpassword123",
				NewPasswordConfirmation: "newpassword123",
			},
			withUser: true,
			mockSetup: func(mockService *mockUserService) {
				mockService.On("ChangePassword", mock.Anything, user.ChangePasswordRequest{
					UserID:                  userID,
					OldPassword:             "wrongpassword",
					NewPassword:             "newpassword123",
					NewPasswordConfirmation: "newpassword123",
				}).Return(apperr.BusinessValidationError("old password is incorrect"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := user.NewUserHandler(mockService, &config.Config{Environment: "development"})

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("PUT", "/users/me/password", &body)
			req.Header.Set("Content-Type", "application/json")
			if tt.withUser {
				req = req.WithContext(context.WithValue(req.Context(), auth.UserIdContextKey, userID))
			}

			w := httptest.NewRecorder()

			handler.ChangePassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

/* func TestUserHandler_RefreshToken(t *testing.T) {
	validUser := &domain.User{
		Id:        uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	validLoginResult := &user.LoginResult{
		AccessToken:  "new-jwt-access-token",
		RefreshToken: "new-refresh-token",
		User:         validUser,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mockUserService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful token refresh",
			requestBody: user.RefreshTokenBody{
				RefreshToken: "valid-refresh-token",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.RefreshTokenRequest{
					Token: "valid-refresh-token",
				}
				mockService.On("RefreshToken", mock.Anything, expectedRequest).Return(validLoginResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response user.LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "new-jwt-access-token", response.AccessToken)
				assert.Equal(t, "new-refresh-token", response.RefreshToken)
				assert.Equal(t, validUser.Id, response.User.Id)
			},
		},
		{
			name:        "invalid JSON",
			requestBody: `{"refresh_token":}`,
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "validation error - missing refresh token",
			requestBody: user.RefreshTokenBody{},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - invalid refresh token",
			requestBody: user.RefreshTokenBody{
				RefreshToken: "invalid-refresh-token",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := user.RefreshTokenRequest{
					Token: "invalid-refresh-token",
				}
				mockService.On("RefreshToken", mock.Anything, expectedRequest).Return(nil, apperr.UnauthorizedError("invalid refresh token"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := user.NewUserHandler(mockService)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/auth/refresh-token", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RefreshToken(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockService.AssertExpectations(t)
		})
	}
}
*/
