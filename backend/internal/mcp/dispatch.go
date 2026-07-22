package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
)

func (h *Handler) dispatchRPCMethod(ctx context.Context, principal principal, req request) (toolName string, status string, resp response) {
	toolName = req.Method
	status = "success"
	resp = response{JSONRPC: "2.0", ID: req.ID}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "project-chat-mcp",
				"version": "1.0.0",
			},
			"instructions": initializeInstructions(principal),
		}
	case "tools/list":
		toolName = "tools/list"
		resp.Result = map[string]any{"tools": toolDefinitionsForPrincipal(principal)}
	case "resources/list":
		toolName = "resources/list"
		resp.Result = map[string]any{"resources": resourceDefinitionsForPrincipal(principal)}
	case "resources/read":
		toolName = "resources/read"
		uri, err := requiredStringArg(paramsFromRaw(req.Params), "uri")
		if err != nil {
			resp.Error = &rpcError{Code: -32602, Message: "invalid params"}
			status = "failure"
			break
		}
		result, readErr := readResource(uri, principal)
		if readErr != nil {
			status = "failure"
			resp.Result = toolErrorResult(readErr)
			break
		}
		resp.Result = result
	case "prompts/list":
		toolName = "prompts/list"
		resp.Result = map[string]any{"prompts": []map[string]any{}}
	case "tools/call":
		toolName = "tools/call"
		var params toolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &rpcError{Code: -32602, Message: "invalid params"}
			status = "failure"
			break
		}
		toolName = params.Name
		result, callErr := h.callTool(ctx, principal, params)
		if callErr != nil {
			status = "failure"
			resp.Result = toolErrorResult(callErr)
			break
		}
		successResult, buildErr := toolSuccessResult(params.Name, result)
		if buildErr != nil {
			status = "failure"
			resp.Result = toolErrorResult(buildErr)
			break
		}
		resp.Result = successResult
	default:
		status = "failure"
		resp.Error = &rpcError{Code: -32601, Message: "method not found"}
	}

	return toolName, status, resp
}

func (h *Handler) handleNotification(w http.ResponseWriter, r *http.Request, principal principal, req request) {
	start := time.Now()
	status := notificationStatus(req.Method)

	logMCPEvent(r.Context(), "mcp_notification", req.Method, status, start, principal)

	w.WriteHeader(http.StatusAccepted)
}

func notificationStatus(method string) string {
	switch method {
	case "notifications/initialized":
		return "accepted"
	default:
		return "ignored"
	}
}

func logMCPEvent(ctx context.Context, event string, toolName string, status string, start time.Time, principal principal) {
	log := logger.FromContext(ctx)
	log.Info(event,
		"tool_name", toolName,
		"status", status,
		"latency_ms", time.Since(start).Milliseconds(),
		"key_id", principal.APIKeyID,
		"owner_user_id", principal.UserID,
	)
}
