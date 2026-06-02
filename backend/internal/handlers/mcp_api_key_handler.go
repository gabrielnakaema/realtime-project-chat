package handlers

import (
	"context"
	"net/http"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/gabrielnakaema/project-chat/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mcpAPIKeyService interface {
	ListAvailableScopes() []domain.MCPAPIScopeDefinition
	Create(ctx context.Context, request service.CreateMCPAPIKeyRequest) (*service.CreateMCPAPIKeyResult, error)
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
	var request CreateMCPAPIKeyRequest
	if err := utils.ReadJSON(w, r, &request); err != nil {
		BadRequestResponse(w, err)
		return
	}

	v := validator.New()
	request.Validate(v)
	if !v.Valid() {
		ValidationFailedResponse(w, v)
		return
	}

	result, err := h.service.Create(r.Context(), service.CreateMCPAPIKeyRequest{
		UserID: UserIdFromContext(r.Context()),
		Name:   request.Name,
		Scopes: request.Scopes,
	})
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusCreated, result, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) ListAvailableScopes(w http.ResponseWriter, r *http.Request) {
	if err := utils.WriteJSON(w, http.StatusOK, h.service.ListAvailableScopes(), nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.service.ListByUserID(r.Context(), UserIdFromContext(r.Context()))
	if err != nil {
		ErrorResponse(w, r, err)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, keys, nil); err != nil {
		ErrorResponse(w, r, err)
	}
}

func (h *MCPAPIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		BadRequestResponse(w, err)
		return
	}

	if err := h.service.Revoke(r.Context(), id, UserIdFromContext(r.Context())); err != nil {
		ErrorResponse(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
