package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	platformhttp "github.com/gabrielnakaema/project-chat/internal/platform/http"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
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

	secret, ok := bearerSecretFromHeader(r.Header.Get("Authorization"))
	if !ok {
		writeHTTPError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	authenticated, err := h.authService.Authenticate(r.Context(), secret)
	if err != nil {
		writeDomainHTTPError(w, err)
		return
	}

	principal := principal{
		APIKeyID: authenticated.ID,
		UserID:   authenticated.UserID,
		Scopes:   authenticated.Scopes,
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
	status := "success"
	toolName := req.Method
	resp := response{JSONRPC: "2.0", ID: req.ID}

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
		result, callErr := h.callTool(r.Context(), principal, params)
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

	log := logger.FromContext(r.Context())
	log.Info("mcp_request",
		"tool_name", toolName,
		"status", status,
		"latency_ms", time.Since(start).Milliseconds(),
		"key_id", principal.APIKeyID,
		"owner_user_id", principal.UserID,
	)

	_ = platformhttp.WriteJSON(w, http.StatusOK, resp, nil)
}

func (h *Handler) handleNotification(w http.ResponseWriter, r *http.Request, principal principal, req request) {
	start := time.Now()
	status := "accepted"
	toolName := req.Method

	switch req.Method {
	case "notifications/initialized":
		// Streamable HTTP notifications are acknowledged with 202 and no body.
	default:
		status = "ignored"
	}

	log := logger.FromContext(r.Context())
	log.Info("mcp_notification",
		"tool_name", toolName,
		"status", status,
		"latency_ms", time.Since(start).Milliseconds(),
		"key_id", principal.APIKeyID,
		"owner_user_id", principal.UserID,
	)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) callTool(ctx context.Context, principal principal, params toolCallParams) (map[string]any, error) {
	switch params.Name {
	case "list_projects":
		if err := requireScope(principal, domain.MCPAPIScopeProjectsRead); err != nil {
			return nil, err
		}
		projects, err := h.projectService.ListByUserID(ctx, ListProjectsRequest{
			UserID: principal.UserID,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"projects": projects}, nil
	case "list_project_board":
		if err := requireScope(principal, domain.MCPAPIScopeProjectsBoardRead); err != nil {
			return nil, err
		}
		projectID, err := requiredUUIDArg(params.Arguments, "project_id")
		if err != nil {
			return nil, err
		}
		project, err := h.projectService.GetByID(ctx, projectID, principal.UserID)
		if err != nil {
			return nil, err
		}
		projectColumnIDs, err := optionalUUIDSliceArg(params.Arguments, "project_column_ids")
		if err != nil {
			return nil, err
		}
		limitPerColumn := optionalIntArg(params.Arguments, "limit_per_column", 15)
		includeArchived := optionalBoolArg(params.Arguments, "include_archived", false)
		grouped, err := h.taskService.GroupByColumn(ctx, GroupByColumnRequest{
			ProjectID:        projectID,
			UserID:           principal.UserID,
			ProjectColumnIDs: projectColumnIDs,
			Archived:         includeArchived,
			Limit:            limitPerColumn,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"project":         project,
			"columns":         project.Columns,
			"tasks_by_column": grouped,
		}, nil
	case "create_task":
		if err := requireScope(principal, domain.MCPAPIScopeTasksCreate); err != nil {
			return nil, err
		}
		projectID, err := requiredUUIDArg(params.Arguments, "project_id")
		if err != nil {
			return nil, err
		}
		projectColumnID, err := requiredUUIDArg(params.Arguments, "project_column_id")
		if err != nil {
			return nil, err
		}
		title, err := requiredStringArg(params.Arguments, "title")
		if err != nil {
			return nil, err
		}
		description, err := requiredStringArg(params.Arguments, "description")
		if err != nil {
			return nil, err
		}
		priority, err := requiredTaskPriorityArg(params.Arguments, "priority")
		if err != nil {
			return nil, err
		}
		responsibleID, err := optionalUUIDArg(params.Arguments, "responsible_id")
		if err != nil {
			return nil, err
		}
		dueDate, err := optionalTimeArg(params.Arguments, "due_date")
		if err != nil {
			return nil, err
		}
		tags, err := optionalStringSliceArg(params.Arguments, "tags")
		if err != nil {
			return nil, err
		}
		dependsOnTaskIDs, err := optionalUUIDSliceArg(params.Arguments, "depends_on_task_ids")
		if err != nil {
			return nil, err
		}
		code, err := optionalTrimmedStringArg(params.Arguments, "code")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		task, err := h.taskService.Create(ctx, CreateTaskRequest{
			ProjectID:        projectID,
			ProjectColumnID:  projectColumnID,
			Title:            title,
			Description:      description,
			Code:             code,
			RequestUserID:    principal.UserID,
			Priority:         string(priority),
			ResponsibleID:    responsibleID,
			DueDate:          dueDate,
			Tags:             tags,
			DependsOnTaskIDs: dependsOnTaskIDs,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"task": task}, nil
	case "get_task":
		if err := requireScope(principal, domain.MCPAPIScopeTasksRead); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		task, err := h.taskService.GetByID(ctx, taskID, principal.UserID)
		if err != nil {
			return nil, err
		}
		includeComments := optionalBoolArg(params.Arguments, "include_comments", false)
		if includeComments {
			if err := requireScope(principal, domain.MCPAPIScopeTasksCommentsRead); err != nil {
				return nil, err
			}
			commentsLimit := optionalIntArg(params.Arguments, "comments_limit", 10)
			comments, err := h.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
				TaskID:        taskID,
				RequestUserID: principal.UserID,
				Limit:         commentsLimit,
			})
			if err != nil {
				return nil, err
			}
			task.Comments = comments.Data
		}
		return map[string]any{"task": task}, nil
	case "find_task_by_code":
		if err := requireScope(principal, domain.MCPAPIScopeTasksRead); err != nil {
			return nil, err
		}
		projectID, err := requiredUUIDArg(params.Arguments, "project_id")
		if err != nil {
			return nil, err
		}
		code, err := requiredStringArg(params.Arguments, "code")
		if err != nil {
			return nil, err
		}
		task, err := h.taskService.FindTaskByCode(ctx, FindTaskByCodeRequest{
			ProjectID: projectID,
			UserID:    principal.UserID,
			Code:      code,
		})
		if err != nil {
			return nil, err
		}
		includeComments := optionalBoolArg(params.Arguments, "include_comments", false)
		if includeComments {
			if err := requireScope(principal, domain.MCPAPIScopeTasksCommentsRead); err != nil {
				return nil, err
			}
			commentsLimit := optionalIntArg(params.Arguments, "comments_limit", 10)
			comments, err := h.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
				TaskID:        task.Id,
				RequestUserID: principal.UserID,
				Limit:         commentsLimit,
			})
			if err != nil {
				return nil, err
			}
			task.Comments = comments.Data
		}
		return map[string]any{"task": task}, nil
	case "list_task_comments":
		if err := requireScope(principal, domain.MCPAPIScopeTasksCommentsRead); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		commentsLimit := optionalIntArg(params.Arguments, "limit", 10)
		comments, err := h.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
			TaskID:        taskID,
			RequestUserID: principal.UserID,
			Limit:         commentsLimit,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"comments": comments.Data}, nil
	case "update_task":
		if err := requireScope(principal, domain.MCPAPIScopeTasksUpdate); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		projectColumnID, err := requiredUUIDArg(params.Arguments, "project_column_id")
		if err != nil {
			return nil, err
		}
		title, err := requiredStringArg(params.Arguments, "title")
		if err != nil {
			return nil, err
		}
		description, err := requiredStringArg(params.Arguments, "description")
		if err != nil {
			return nil, err
		}
		priority, err := requiredTaskPriorityArg(params.Arguments, "priority")
		if err != nil {
			return nil, err
		}
		responsibleID, err := optionalUUIDArg(params.Arguments, "responsible_id")
		if err != nil {
			return nil, err
		}
		dueDate, err := optionalTimeArg(params.Arguments, "due_date")
		if err != nil {
			return nil, err
		}
		tags, err := optionalStringSliceArg(params.Arguments, "tags")
		if err != nil {
			return nil, err
		}
		dependsOnTaskIDs, err := optionalUUIDSliceArg(params.Arguments, "depends_on_task_ids")
		if err != nil {
			return nil, err
		}
		code, err := optionalTrimmedStringArgPointer(params.Arguments, "code")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		task, err := h.taskService.Update(ctx, UpdateTaskRequest{
			TaskID:           taskID,
			Title:            title,
			Description:      description,
			Code:             code,
			ProjectColumnID:  projectColumnID,
			RequestUserID:    principal.UserID,
			Priority:         priority,
			ResponsibleID:    responsibleID,
			DueDate:          dueDate,
			Tags:             tags,
			DependsOnTaskIDs: dependsOnTaskIDs,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"task": task}, nil
	case "move_task":
		if err := requireScope(principal, domain.MCPAPIScopeTasksMove); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		projectID, err := requiredUUIDArg(params.Arguments, "project_id")
		if err != nil {
			return nil, err
		}
		targetColumnID, err := requiredUUIDArg(params.Arguments, "target_project_column_id")
		if err != nil {
			return nil, err
		}
		afterTaskID, err := optionalUUIDArg(params.Arguments, "after_task_id")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		task, err := h.taskService.Move(ctx, MoveTaskRequest{
			TaskID:          taskID,
			RequestUserID:   principal.UserID,
			AfterTaskID:     afterTaskID,
			ProjectID:       projectID,
			ProjectColumnID: targetColumnID,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"task": task}, nil
	case "add_task_comment":
		if err := requireScope(principal, domain.MCPAPIScopeTasksComment); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		parentCommentID, err := optionalUUIDArg(params.Arguments, "parent_comment_id")
		if err != nil {
			return nil, err
		}
		content, err := requiredStringArg(params.Arguments, "content")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		comment, err := h.taskCommentService.Create(ctx, CreateTaskCommentRequest{
			TaskID:          taskID,
			RequestUserID:   principal.UserID,
			Content:         content,
			ParentCommentID: parentCommentID,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"comment": comment}, nil
	case "mark_task_done":
		if err := requireScope(principal, domain.MCPAPIScopeTasksMarkDone); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		task, err := h.taskService.MarkTaskDone(ctx, MarkTaskDoneRequest{
			TaskID:        taskID,
			RequestUserID: principal.UserID,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"task": task}, nil
	case "assign_task_to_self":
		if err := requireScope(principal, domain.MCPAPIScopeTasksAssignSelf); err != nil {
			return nil, err
		}
		taskID, err := requiredUUIDArg(params.Arguments, "task_id")
		if err != nil {
			return nil, err
		}
		ctx = domain.WithActionOrigin(ctx, domain.ActionOriginMCPAgent)
		task, err := h.taskService.AssignTaskToSelf(ctx, AssignTaskToSelfRequest{
			TaskID:        taskID,
			RequestUserID: principal.UserID,
		})
		if err != nil {
			return nil, err
		}
		return map[string]any{"task": task}, nil
	default:
		return nil, apperr.NotFoundError("tool not found")
	}
}

func requireScope(principal principal, scope domain.MCPAPIScope) error {
	if principal.HasScope(scope) {
		return nil
	}

	return apperr.ForbiddenError("missing required scope")
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

func toolErrorResult(err error) map[string]any {
	var domainErr apperr.DomainError
	if !errors.As(err, &domainErr) {
		return buildToolErrorResult("internal server error", "server_error", "internal server error")
	}

	errorType := "server_error"
	switch domainErr.Code {
	case apperr.UnauthorizedErrorCode:
		errorType = "authentication_failed"
	case apperr.ForbiddenErrorCode:
		if domainErr.Message == "missing required scope" {
			errorType = "missing_scope"
		} else {
			errorType = "forbidden"
		}
	case apperr.NotFoundErrorCode:
		errorType = "not_found"
	case apperr.BusinessValidationErrorCode:
		errorType = "business_validation"
	}

	return buildToolErrorResult(domainErr.Message, errorType, domainErr.Message)
}

func buildToolErrorResult(summary string, errorType string, message string) map[string]any {
	structuredContent := map[string]any{
		"error": map[string]any{
			"type":    errorType,
			"message": message,
		},
	}

	content, err := toolResultContent(summary, structuredContent)
	if err != nil {
		return map[string]any{
			"content": []map[string]any{{"type": "text", "text": summary}},
			"isError": true,
			"structuredContent": map[string]any{
				"error": map[string]any{
					"type":    "server_error",
					"message": summary,
				},
			},
		}
	}

	return map[string]any{
		"content":           content,
		"isError":           true,
		"structuredContent": structuredContent,
	}
}

func requiredUUIDArg(args map[string]any, key string) (uuid.UUID, error) {
	raw, err := requiredStringArg(args, key)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	return id, nil
}

func optionalUUIDArg(args map[string]any, key string) (*uuid.UUID, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid uuid", key))
	}

	return &id, nil
}

func optionalUUIDSliceArg(args map[string]any, key string) ([]uuid.UUID, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
	}

	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
		}
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of uuids", key))
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func paramsFromRaw(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}

	var params map[string]any
	if err := json.Unmarshal(raw, &params); err != nil {
		return map[string]any{}
	}

	return params
}

func requiredStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s is required", key))
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s is required", key))
	}

	return strings.TrimSpace(value), nil
}

func optionalIntArg(args map[string]any, key string, defaultValue int) int {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue
	}

	number, ok := raw.(float64)
	if !ok || number < 1 {
		return defaultValue
	}

	return int(number)
}

func optionalBoolArg(args map[string]any, key string, defaultValue bool) bool {
	raw, ok := args[key]
	if !ok || raw == nil {
		return defaultValue
	}

	value, ok := raw.(bool)
	if !ok {
		return defaultValue
	}

	return value
}

func requiredTaskPriorityArg(args map[string]any, key string) (domain.TaskPriority, error) {
	value, err := requiredStringArg(args, key)
	if err != nil {
		return "", err
	}

	priority := domain.TaskPriority(strings.ToLower(value))
	for _, allowed := range domain.AllowedTaskPriorities {
		if priority == allowed {
			return priority, nil
		}
	}

	return "", apperr.BusinessValidationError(fmt.Sprintf("%s must be one of: low, medium, high", key))
}

func optionalTimeArg(args map[string]any, key string) (*time.Time, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid RFC3339 datetime", key))
	}

	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a valid RFC3339 datetime", key))
	}

	return &parsed, nil
}

func optionalStringSliceArg(args map[string]any, key string) ([]string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of strings", key))
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be an array of strings", key))
		}
		values = append(values, strings.TrimSpace(value))
	}

	return values, nil
}

func optionalTrimmedStringArg(args map[string]any, key string) (string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return "", nil
	}

	value, ok := raw.(string)
	if !ok {
		return "", apperr.BusinessValidationError(fmt.Sprintf("%s must be a string", key))
	}

	return strings.TrimSpace(value), nil
}

func optionalTrimmedStringArgPointer(args map[string]any, key string) (*string, error) {
	raw, ok := args[key]
	if !ok || raw == nil {
		return nil, nil
	}

	value, ok := raw.(string)
	if !ok {
		return nil, apperr.BusinessValidationError(fmt.Sprintf("%s must be a string", key))
	}

	trimmed := strings.TrimSpace(value)
	return &trimmed, nil
}

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
