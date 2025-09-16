package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockTokenProvider struct {
	mock.Mock
}

func (m *mockTokenProvider) Verify(token string) (*jwt.Token, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func createValidJwt(userId uuid.UUID, validExpirationTime time.Time) *jwt.Token {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"exp": float64(validExpirationTime.Unix()),
		"iat": float64(time.Now().Unix()),
		"iss": "projectmanagementapi",
	})

	token.Valid = true

	return token
}

func TestAuthMiddleware(t *testing.T) {
	type testCase struct {
		name        string
		authHeader  string
		status      int
		checkUserId uuid.UUID
		mockSetup   func(*mockTokenProvider)
	}

	validUserId := uuid.New()
	validExpirationTime := time.Now().Add(time.Hour)

	tests := []testCase{
		{
			name:        "valid token",
			authHeader:  "Bearer valid-token",
			status:      http.StatusOK,
			checkUserId: validUserId,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
				token := createValidJwt(validUserId, validExpirationTime)

				mockTokenProvider.On("Verify", "valid-token").Return(token, nil)
			},
		},
		{
			name:        "empty header",
			authHeader:  "",
			status:      http.StatusOK,
			checkUserId: uuid.Nil,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
			},
		},
		{
			name:        "empty token",
			authHeader:  "Bearer ",
			status:      http.StatusUnauthorized,
			checkUserId: uuid.Nil,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
			},
		},
		{
			name:        "token spaces",
			authHeader:  "Bearer   ",
			status:      http.StatusUnauthorized,
			checkUserId: uuid.Nil,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
			},
		},
		{
			name:        "invalid token",
			authHeader:  "Bearer invalid-token",
			status:      http.StatusUnauthorized,
			checkUserId: uuid.Nil,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
				mockTokenProvider.On("Verify", "invalid-token").Return(nil, errors.New("invalid token"))
			},
		},
		{
			name:        "expired token",
			authHeader:  "Bearer expired-token",
			status:      http.StatusUnauthorized,
			checkUserId: uuid.Nil,
			mockSetup: func(mockTokenProvider *mockTokenProvider) {
				token := createValidJwt(validUserId, time.Now().Add(-time.Hour))
				mockTokenProvider.On("Verify", "expired-token").Return(token, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTokenProvider := &mockTokenProvider{}
			tt.mockSetup(mockTokenProvider)

			authMiddleware := handlers.NewAuthMiddleware(mockTokenProvider)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			recorder := httptest.NewRecorder()

			var userId uuid.UUID
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userId = handlers.UserIdFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := authMiddleware.IdentifyUser(nextHandler)
			handler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.checkUserId, userId)
			assert.Equal(t, tt.status, recorder.Code)

			mockTokenProvider.AssertExpectations(t)
		})
	}
}
