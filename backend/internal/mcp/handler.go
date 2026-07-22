package mcp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/google/uuid"
)

const protocolVersion = "2025-06-18"

type Handler struct {
	authService        authService
	projectService     projectService
	taskService        taskService
	taskCommentService taskCommentService
}

func NewHandler(authService authService, projectService projectService, taskService taskService, taskCommentService taskCommentService) *Handler {
	return &Handler{
		authService:        authService,
		projectService:     projectService,
		taskService:        taskService,
		taskCommentService: taskCommentService,
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Result  map[string]any `json:"result,omitempty"`
	Error   *rpcError      `json:"error,omitempty"`
}

type rpcError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type principal struct {
	APIKeyID uuid.UUID
	UserID   uuid.UUID
	Scopes   []domain.MCPAPIScope
}

func (p principal) HasScope(scope domain.MCPAPIScope) bool {
	for _, granted := range p.Scopes {
		if granted == scope {
			return true
		}
	}

	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("MCP-Protocol-Version", protocolVersion)

	principal, err := h.authenticateRequest(r)
	if err != nil {
		writeDomainHTTPError(w, err)
		return
	}

	var req request
	if err := platformhttp.ReadJSON(w, r, &req); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ID == nil {
		h.handleNotification(w, r, principal, req)
		return
	}

	start := time.Now()
	toolName, status, resp := h.dispatchRPCMethod(r.Context(), principal, req)
	logMCPEvent(r.Context(), "mcp_request", toolName, status, start, principal)

	_ = platformhttp.WriteJSON(w, http.StatusOK, resp, nil)
}
