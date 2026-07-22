package mcp

import (
	"errors"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
)

func writeHTTPError(w http.ResponseWriter, status int, message string) {
	_ = platformhttp.WriteJSON(w, status, map[string]any{
		"status":  status,
		"message": message,
	}, nil)
}

func writeDomainHTTPError(w http.ResponseWriter, err error) {
	var domainErr apperr.DomainError
	if !errors.As(err, &domainErr) {
		writeHTTPError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	switch domainErr.Code {
	case apperr.UnauthorizedErrorCode:
		writeHTTPError(w, http.StatusUnauthorized, domainErr.Message)
	case apperr.ForbiddenErrorCode:
		writeHTTPError(w, http.StatusForbidden, domainErr.Message)
	case apperr.NotFoundErrorCode:
		writeHTTPError(w, http.StatusNotFound, domainErr.Message)
	case apperr.BusinessValidationErrorCode, apperr.ValidationFailedErrorCode, apperr.DuplicateEntryErrorCode:
		writeHTTPError(w, http.StatusUnprocessableEntity, domainErr.Message)
	default:
		writeHTTPError(w, http.StatusInternalServerError, "Internal server error")
	}
}
