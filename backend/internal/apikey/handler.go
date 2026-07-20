package apikey

import (
	"context"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/auth"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/httperr"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mcpAPIKeyService interface {
	ListAvailableScopes() []domain.MCPAPIScopeDefinition
	Create(ctx context.Context, request CreateMCPAPIKeyRequest) (*CreateMCPAPIKeyResult, error)
	Update(ctx context.Context, request UpdateMCPAPIKeyRequest) (*domain.MCPAPIKey, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MCPAPIKey, error)
	Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type MCPAPIKeyHandler struct {
	service mcpAPIKeyService
}

func NewMCPAPIKeyHandler(service mcpAPIKeyService) *MCPAPIKeyHandler {
	return &MCPAPIKeyHandler{service: service}
}

func (h *MCPAPIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateMCPAPIKeyBody
	if err := platformhttp.ReadJSON(w, r, &request); err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	result, err := h.service.Create(r.Context(), CreateMCPAPIKeyRequest{
		UserID: auth.UserIdFromContext(r.Context()),
		Name:   request.Name,
		Scopes: request.Scopes,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusCreated, result, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) ListAvailableScopes(w http.ResponseWriter, r *http.Request) {
	if err := platformhttp.WriteJSON(w, http.StatusOK, h.service.ListAvailableScopes(), nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.service.ListByUserID(r.Context(), auth.UserIdFromContext(r.Context()))
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, keys, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	var request UpdateMCPAPIKeyBody
	if err := platformhttp.ReadJSON(w, r, &request); err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		httperr.ValidationFailedResponse(w, v)
		return
	}

	key, err := h.service.Update(r.Context(), UpdateMCPAPIKeyRequest{
		ID:     id,
		UserID: auth.UserIdFromContext(r.Context()),
		Name:   request.Name,
		Scopes: request.Scopes,
	})
	if err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	if err := platformhttp.WriteJSON(w, http.StatusOK, key, nil); err != nil {
		httperr.ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httperr.BadRequestResponse(w, err)
		return
	}

	if err := h.service.Revoke(r.Context(), id, auth.UserIdFromContext(r.Context())); err != nil {
		httperr.ErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
