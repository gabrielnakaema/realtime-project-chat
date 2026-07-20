package httperr

import (
	"errors"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/gabrielnakaema/project-chat/internal/validator"
)

type ApiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Meta    any    `json:"meta,omitempty"`
}

func (a ApiError) Error() string {
	return a.Message
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr ApiError
	var domainErr apperr.DomainError

	log := logger.FromContext(r.Context())

	if errors.As(err, &apiErr) {
		writeError(w, apiErr)
		return
	}

	if errors.As(err, &domainErr) {
		apiErr = mapDomainErrors(domainErr)
		if apiErr.Status == http.StatusInternalServerError {
			if domainErr.Cause != nil {
				log.Error("internal_server_error", "message", domainErr.Message, "error", domainErr.Cause.Error())
			} else {
				log.Error("internal_server_error", "message", domainErr.Message)
			}
		}
		writeError(w, apiErr)
		return
	}

	apiErr = ApiError{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
	log.Error("internal_server_error", "message", "unknown error", "error", err.Error())

	writeError(w, apiErr)
}

func ValidationFailedResponse(w http.ResponseWriter, v *validator.Validator) {
	apiErr := ApiError{
		Status:  http.StatusUnprocessableEntity,
		Message: "Validation Failed",
		Meta:    v.Errors,
	}

	writeError(w, apiErr)
}

func BadRequestResponse(w http.ResponseWriter, err error) {
	apiErr := ApiError{
		Status:  http.StatusBadRequest,
		Message: err.Error(),
	}

	writeError(w, apiErr)
}

func UnauthorizedResponse(w http.ResponseWriter, message string) {
	apiErr := ApiError{
		Status:  http.StatusUnauthorized,
		Message: message,
	}
	writeError(w, apiErr)
}

func writeError(w http.ResponseWriter, apiErr ApiError) {
	err := platformhttp.WriteJSON(w, apiErr.Status, apiErr, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func mapDomainErrors(err apperr.DomainError) ApiError {
	switch err.Code {
	case apperr.NotFoundErrorCode:
		return ApiError{
			Status:  http.StatusNotFound,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case apperr.UnauthorizedErrorCode:
		return ApiError{
			Status:  http.StatusUnauthorized,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case apperr.DuplicateEntryErrorCode:
		return ApiError{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case apperr.ForbiddenErrorCode:
		return ApiError{
			Status:  http.StatusForbidden,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case apperr.ValidationFailedErrorCode:
		return ApiError{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case apperr.ServerErrorCode:
		return ApiError{
			Status:  http.StatusInternalServerError,
			Message: "Internal server error",
		}
	case apperr.BusinessValidationErrorCode:
		return ApiError{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Message,
			Meta:    err.Meta,
		}
	default:
		return ApiError{
			Status:  http.StatusInternalServerError,
			Message: "Internal server error",
		}
	}
}
