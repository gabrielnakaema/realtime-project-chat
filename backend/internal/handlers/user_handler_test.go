package handlers_test

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
	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, request service.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, request service.LoginRequest) (*service.LoginResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResult), args.Error(1)
}

func (m *mockUserService) RefreshToken(ctx context.Context, request service.RefreshTokenRequest) (*service.LoginResult, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResult), args.Error(1)
}

func (m *mockUserService) GetMe(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
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
			requestBody: handlers.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.CreateUserRequest{
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
			requestBody: handlers.CreateUserRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - invalid email",
			requestBody: handlers.CreateUserRequest{
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
			requestBody: handlers.CreateUserRequest{
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
			requestBody: handlers.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.CreateUserRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Create", mock.Anything, expectedRequest).Return(nil, domain.DuplicateEntryError("user email is already taken"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - internal server error",
			requestBody: handlers.CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.CreateUserRequest{
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Create", mock.Anything, expectedRequest).Return(nil, domain.ServerError("database error", errors.New("connection failed")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := handlers.NewUserHandler(mockService)

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

	validLoginResult := &service.LoginResult{
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
			requestBody: handlers.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.LoginRequest{
					Email:    "john@example.com",
					Password: "password123",
				}
				mockService.On("Login", mock.Anything, expectedRequest).Return(validLoginResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handlers.LoginResponse
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
			requestBody: handlers.LoginRequest{
				Password: "password123",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "validation error - missing password",
			requestBody: handlers.LoginRequest{
				Email: "john@example.com",
			},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - invalid credentials",
			requestBody: handlers.LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.LoginRequest{
					Email:    "john@example.com",
					Password: "wrongpassword",
				}
				mockService.On("Login", mock.Anything, expectedRequest).Return(nil, domain.UnauthorizedError("invalid credentials"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := handlers.NewUserHandler(mockService)

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

/* func TestUserHandler_RefreshToken(t *testing.T) {
	validUser := &domain.User{
		Id:        uuid.New(),
		Name:      "John Doe",
		Email:     "john@example.com",
		Password:  "hashedpassword",
		CreatedAt: time.Now(),
	}

	validLoginResult := &service.LoginResult{
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
			requestBody: handlers.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.RefreshTokenRequest{
					Token: "valid-refresh-token",
				}
				mockService.On("RefreshToken", mock.Anything, expectedRequest).Return(validLoginResult, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handlers.LoginResponse
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
			requestBody: handlers.RefreshTokenRequest{},
			mockSetup: func(mockService *mockUserService) {
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "service error - invalid refresh token",
			requestBody: handlers.RefreshTokenRequest{
				RefreshToken: "invalid-refresh-token",
			},
			mockSetup: func(mockService *mockUserService) {
				expectedRequest := service.RefreshTokenRequest{
					Token: "invalid-refresh-token",
				}
				mockService.On("RefreshToken", mock.Anything, expectedRequest).Return(nil, domain.UnauthorizedError("invalid refresh token"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockUserService{}
			tt.mockSetup(mockService)

			handler := handlers.NewUserHandler(mockService)

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
