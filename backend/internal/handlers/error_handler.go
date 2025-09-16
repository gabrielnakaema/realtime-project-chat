package handlers

import (
	"errors"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/logger"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
)

type ApiError struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta,omitempty"`
}

func (a ApiError) Error() string {
	return a.Message
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr ApiError
	var domainErr domain.DomainError

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
	err := utils.WriteJSON(w, apiErr.Status, apiErr, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func mapDomainErrors(err domain.DomainError) ApiError {
	switch err.Code {
	case domain.NotFoundErrorCode:
		return ApiError{
			Status:  http.StatusNotFound,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case domain.UnauthorizedErrorCode:
		return ApiError{
			Status:  http.StatusUnauthorized,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case domain.DuplicateEntryErrorCode:
		return ApiError{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case domain.ForbiddenErrorCode:
		return ApiError{
			Status:  http.StatusForbidden,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case domain.ValidationFailedErrorCode:
		return ApiError{
			Status:  http.StatusUnprocessableEntity,
			Message: err.Message,
			Meta:    err.Meta,
		}
	case domain.ServerErrorCode:
		return ApiError{
			Status:  http.StatusInternalServerError,
			Message: "Internal server error",
		}
	case domain.BusinessValidationErrorCode:
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
