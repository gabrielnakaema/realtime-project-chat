package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenProvider interface {
	Verify(token string) (*jwt.Token, error)
}

type Middleware struct {
	tokenProvider tokenProvider
}

func NewMiddleware(tokenProvider tokenProvider) *Middleware {
	return &Middleware{
		tokenProvider: tokenProvider,
	}
}

type userIdContextKey string

const UserIdContextKey userIdContextKey = "user_id"

func UserIdFromContext(ctx context.Context) uuid.UUID {
	userId, ok := ctx.Value(UserIdContextKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userId
}

func WithAnonymousUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, UserIdContextKey, uuid.Nil)
}

func (m *Middleware) IdentifyUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const INVALID_TOKEN_ERROR_MESSAGE = "invalid token"

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			req := r.WithContext(WithAnonymousUser(r.Context()))
			next.ServeHTTP(w, req)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		if strings.TrimSpace(token) == "" {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		jwt, err := m.tokenProvider.Verify(token)
		if err != nil {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		claims := jwt.Claims

		exp, err := claims.GetExpirationTime()
		if err != nil {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		if exp.Time.Before(time.Now()) {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		userId, err := claims.GetSubject()
		if err != nil {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		tokenUserId, err := uuid.Parse(userId)
		if err != nil {
			unauthorizedResponse(w, INVALID_TOKEN_ERROR_MESSAGE)
			return
		}

		req := r.WithContext(context.WithValue(r.Context(), UserIdContextKey, tokenUserId))

		next.ServeHTTP(w, req)
	})
}

func (m *Middleware) ProtectRoutes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := UserIdFromContext(r.Context())
		if userId == uuid.Nil {
			unauthorizedResponse(w, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

type apiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func unauthorizedResponse(w http.ResponseWriter, message string) {
	err := utils.WriteJSON(w, http.StatusUnauthorized, apiError{
		Status:  http.StatusUnauthorized,
		Message: message,
	}, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
