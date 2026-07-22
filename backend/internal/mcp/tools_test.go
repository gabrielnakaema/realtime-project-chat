package mcp

import (
	"context"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallToolMissingRequiredArgs(t *testing.T) {
	handler := NewHandler(nil,
		&stubProjectService{project: &domain.Project{}},
		&stubTaskService{task: &domain.Task{}},
		&stubTaskCommentService{},
	)
	principal := principal{
		APIKeyID: uuid.New(),
		UserID:   uuid.New(),
		Scopes:   domain.AllowedMCPAPIScopes,
	}

	tests := []struct {
		name     string
		toolName string
		args     map[string]any
	}{
		{
			name:     "list_project_board missing project_id",
			toolName: "list_project_board",
			args:     map[string]any{},
		},
		{
			name:     "create_task missing project_id",
			toolName: "create_task",
			args: map[string]any{
				"project_column_id": uuid.NewString(),
				"title":             "Title",
				"description":       "Description",
				"priority":          "medium",
			},
		},
		{
			name:     "get_task missing task_id",
			toolName: "get_task",
			args:     map[string]any{},
		},
		{
			name:     "find_task_by_code missing project_id",
			toolName: "find_task_by_code",
			args:     map[string]any{"code": "TASK-1"},
		},
		{
			name:     "list_task_comments missing task_id",
			toolName: "list_task_comments",
			args:     map[string]any{},
		},
		{
			name:     "update_task missing task_id",
			toolName: "update_task",
			args: map[string]any{
				"project_column_id": uuid.NewString(),
				"title":             "Title",
				"description":       "Description",
				"priority":          "medium",
			},
		},
		{
			name:     "move_task missing task_id",
			toolName: "move_task",
			args: map[string]any{
				"project_id":               uuid.NewString(),
				"target_project_column_id": uuid.NewString(),
			},
		},
		{
			name:     "add_task_comment missing task_id",
			toolName: "add_task_comment",
			args:     map[string]any{"content": "Hi"},
		},
		{
			name:     "mark_task_done missing task_id",
			toolName: "mark_task_done",
			args:     map[string]any{},
		},
		{
			name:     "assign_task_to_self missing task_id",
			toolName: "assign_task_to_self",
			args:     map[string]any{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := handler.callTool(context.Background(), principal, toolCallParams{
				Name:      tc.toolName,
				Arguments: tc.args,
			})
			require.Error(t, err)

			result := toolErrorResult(err)
			structured := result["structuredContent"].(map[string]any)
			assert.Equal(t, "business_validation", structured["error"].(map[string]any)["type"])
		})
	}
}

func TestCallToolUnknownToolReturnsNotFound(t *testing.T) {
	handler := NewHandler(nil, &stubProjectService{}, &stubTaskService{}, &stubTaskCommentService{})
	principal := principal{UserID: uuid.New(), Scopes: domain.AllowedMCPAPIScopes}

	_, err := handler.callTool(context.Background(), principal, toolCallParams{Name: "does_not_exist"})
	require.Error(t, err)

	result := toolErrorResult(err)
	structured := result["structuredContent"].(map[string]any)
	assert.Equal(t, "not_found", structured["error"].(map[string]any)["type"])
}

func TestRegisteredToolsHaveUniqueCompleteExecutionMetadata(t *testing.T) {
	seenNames := make(map[string]struct{}, len(registeredTools))

	for _, spec := range registeredTools {
		assert.NotEmpty(t, spec.Name)
		assert.NotEmpty(t, spec.RequiredScope, "tool %q must declare a scope", spec.Name)
		assert.NotNil(t, spec.Handle, "tool %q must declare a handler", spec.Name)
		assert.NotNil(t, spec.SuccessText, "tool %q must declare a success formatter", spec.Name)

		_, duplicate := seenNames[spec.Name]
		assert.False(t, duplicate, "tool %q is registered more than once", spec.Name)
		seenNames[spec.Name] = struct{}{}
	}

	assert.Len(t, registeredTools, 11)
}
