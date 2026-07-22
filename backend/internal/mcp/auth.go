package mcp

import (
	"net/http"
	"strings"

	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
)

func (h *Handler) authenticateRequest(r *http.Request) (principal, error) {
	secret, ok := bearerSecretFromHeader(r.Header.Get("Authorization"))
	if !ok {
		return principal{}, apperr.UnauthorizedError("unauthorized")
	}

	authenticated, err := h.authService.Authenticate(r.Context(), secret)
	if err != nil {
		return principal{}, err
	}

	return principal{
		APIKeyID: authenticated.ID,
		UserID:   authenticated.UserID,
		Scopes:   authenticated.Scopes,
	}, nil
}

func bearerSecretFromHeader(header string) (string, bool) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}

	secret := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if secret == "" {
		return "", false
	}

	return secret, true
}
