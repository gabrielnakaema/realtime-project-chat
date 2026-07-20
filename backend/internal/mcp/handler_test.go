package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProjectService struct {
	projects []domain.Project
	project  *domain.Project
}

func (s *stubProjectService) ListByUserID(ctx context.Context, request ListProjectsRequest) ([]domain.Project, error) {
	return s.projects, nil
}

func (s *stubProjectService) GetByID(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error) {
	return s.project, nil
}

type stubTaskService struct {
	createRequest     *CreateTaskRequest
	task              *domain.Task
	grouped           map[string]utils.CursorPaginated[domain.Task]
	findByCodeRequest *FindTaskByCodeRequest
	updateRequest     *UpdateTaskRequest
	moveRequest       *MoveTaskRequest
	markDoneRequest   *MarkTaskDoneRequest
	assignSelfRequest *AssignTaskToSelfRequest
	createOrigin      domain.ActionOrigin
	updateOrigin      domain.ActionOrigin
	moveOrigin        domain.ActionOrigin
	markDoneOrigin    domain.ActionOrigin
	assignSelfOrigin  domain.ActionOrigin
}

func (s *stubTaskService) Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error) {
	s.createRequest = &request
	s.createOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error) {
	return s.grouped, nil
}

func (s *stubTaskService) GetByID(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error) {
	return s.task, nil
}

func (s *stubTaskService) FindTaskByCode(ctx context.Context, request FindTaskByCodeRequest) (*domain.Task, error) {
	s.findByCodeRequest = &request
	return s.task, nil
}

func (s *stubTaskService) Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error) {
	s.moveRequest = &request
	s.moveOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error) {
	s.updateRequest = &request
	s.updateOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) MarkTaskDone(ctx context.Context, request MarkTaskDoneRequest) (*domain.Task, error) {
	s.markDoneRequest = &request
	s.markDoneOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) AssignTaskToSelf(ctx context.Context, request AssignTaskToSelfRequest) (*domain.Task, error) {
	s.assignSelfRequest = &request
	s.assignSelfOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

type stubTaskCommentService struct {
	comment       *domain.TaskComment
	comments      *utils.CursorPaginated[domain.TaskComment]
	createRequest *CreateTaskCommentRequest
	createOrigin  domain.ActionOrigin
	listRequest   *ListTaskCommentsRequest
}

func (s *stubTaskCommentService) Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error) {
	s.createRequest = &request
	s.createOrigin = domain.ActionOriginFromContext(ctx)
	return s.comment, nil
}

func (s *stubTaskCommentService) ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error) {
	s.listRequest = &request
	return s.comments, nil
}

func TestCallToolSuccessPaths(t *testing.T) {
	projectID := uuid.New()
	columnID := uuid.New()
	taskID := uuid.New()
	userID := uuid.New()

	task := &domain.Task{
		Id:              taskID,
		ProjectId:       projectID,
		ProjectColumnId: columnID,
		Title:           "Task",
		Description:     "Desc",
		Code:            "TASK-1",
		Priority:        domain.TaskPriorityMedium,
		Tags:            []string{"backend"},
	}
	project := &domain.Project{
		Id: projectID,
		Columns: []domain.ProjectColumn{
			{Id: columnID, Name: "Doing"},
			{Id: uuid.New(), Name: "Done", IsDoneColumn: true},
		},
	}
	comment := domain.TaskComment{ID: uuid.NewString(), Content: "Hi", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	projectSvc := &stubProjectService{
		projects: []domain.Project{{Id: projectID}},
		project:  project,
	}
	taskSvc := &stubTaskService{
		task: task,
		grouped: map[string]utils.CursorPaginated[domain.Task]{
			columnID.String(): {Data: []domain.Task{*task}},
		},
	}
	commentSvc := &stubTaskCommentService{
		comment:  &comment,
		comments: &utils.CursorPaginated[domain.TaskComment]{Data: []domain.TaskComment{comment}},
	}

	handler := NewHandler(nil, projectSvc, taskSvc, commentSvc)
	principal := principal{
		APIKeyID: uuid.New(),
		UserID:   userID,
		Scopes:   domain.AllowedMCPAPIScopes,
	}

	_, err := handler.callTool(context.Background(), principal, toolCallParams{Name: "list_projects", Arguments: map[string]any{}})
	require.NoError(t, err)

	boardResult, err := handler.callTool(context.Background(), principal, toolCallParams{
		Name: "list_project_board",
		Arguments: map[string]any{
			"project_id": projectID.String(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, projectID.String(), boardResult["project"].(*domain.Project).Id.String())

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "create_task",
		Arguments: map[string]any{
			"project_id":          projectID.String(),
			"project_column_id":   columnID.String(),
			"title":               "New Task",
			"description":         "Created via MCP",
			"code":                "  TASK-123  ",
			"priority":            "medium",
			"tags":                []any{"backend", "mcp"},
			"depends_on_task_ids": []any{taskID.String()},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.createOrigin)
	require.NotNil(t, taskSvc.createRequest)
	assert.Equal(t, projectID, taskSvc.createRequest.ProjectID)
	assert.Equal(t, "TASK-123", taskSvc.createRequest.Code)
	assert.Equal(t, []uuid.UUID{taskID}, taskSvc.createRequest.DependsOnTaskIDs)

	taskResult, err := handler.callTool(context.Background(), principal, toolCallParams{
		Name: "get_task",
		Arguments: map[string]any{
			"task_id":          taskID.String(),
			"include_comments": true,
			"comments_limit":   float64(5),
		},
	})
	require.NoError(t, err)
	assert.Len(t, taskResult["task"].(*domain.Task).Comments, 1)
	require.NotNil(t, commentSvc.listRequest)
	assert.Equal(t, 5, commentSvc.listRequest.Limit)

	findTaskResult, err := handler.callTool(context.Background(), principal, toolCallParams{
		Name: "find_task_by_code",
		Arguments: map[string]any{
			"project_id":       projectID.String(),
			"code":             "  TASK-1  ",
			"include_comments": true,
			"comments_limit":   float64(4),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, taskID.String(), findTaskResult["task"].(*domain.Task).Id.String())
	require.NotNil(t, taskSvc.findByCodeRequest)
	assert.Equal(t, projectID, taskSvc.findByCodeRequest.ProjectID)
	assert.Equal(t, "TASK-1", taskSvc.findByCodeRequest.Code)
	require.NotNil(t, commentSvc.listRequest)
	assert.Equal(t, 4, commentSvc.listRequest.Limit)

	listCommentsResult, err := handler.callTool(context.Background(), principal, toolCallParams{
		Name: "list_task_comments",
		Arguments: map[string]any{
			"task_id": taskID.String(),
			"limit":   float64(3),
		},
	})
	require.NoError(t, err)
	assert.Len(t, listCommentsResult["comments"].([]domain.TaskComment), 1)
	require.NotNil(t, commentSvc.listRequest)
	assert.Equal(t, 3, commentSvc.listRequest.Limit)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "update_task",
		Arguments: map[string]any{
			"task_id":             taskID.String(),
			"project_column_id":   columnID.String(),
			"title":               "Task Updated",
			"description":         "Updated via MCP",
			"code":                "TASK-456",
			"priority":            "high",
			"tags":                []any{"backend", "updated"},
			"depends_on_task_ids": []any{taskID.String()},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.updateOrigin)
	require.NotNil(t, taskSvc.updateRequest)
	assert.Equal(t, taskID, taskSvc.updateRequest.TaskID)
	require.NotNil(t, taskSvc.updateRequest.Code)
	assert.Equal(t, "TASK-456", *taskSvc.updateRequest.Code)
	assert.Equal(t, []uuid.UUID{taskID}, taskSvc.updateRequest.DependsOnTaskIDs)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "move_task",
		Arguments: map[string]any{
			"task_id":                  taskID.String(),
			"project_id":               projectID.String(),
			"target_project_column_id": columnID.String(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.moveOrigin)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "add_task_comment",
		Arguments: map[string]any{
			"task_id": taskID.String(),
			"content": "Ship it",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, commentSvc.createOrigin)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "mark_task_done",
		Arguments: map[string]any{
			"task_id": taskID.String(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.markDoneOrigin)
	require.NotNil(t, taskSvc.markDoneRequest)
	assert.Equal(t, taskID, taskSvc.markDoneRequest.TaskID)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "assign_task_to_self",
		Arguments: map[string]any{
			"task_id": taskID.String(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.assignSelfOrigin)
	require.NotNil(t, taskSvc.assignSelfRequest)
	assert.Equal(t, taskID, taskSvc.assignSelfRequest.TaskID)
}

func TestUpdateTaskCodeArgHandling(t *testing.T) {
	projectID := uuid.New()
	columnID := uuid.New()
	taskID := uuid.New()

	task := &domain.Task{
		Id:              taskID,
		ProjectId:       projectID,
		ProjectColumnId: columnID,
		Title:           "Task",
		Description:     "Desc",
		Code:            "EXISTING-1",
		Priority:        domain.TaskPriorityMedium,
		Tags:            []string{},
	}
	project := &domain.Project{
		Id:      projectID,
		Columns: []domain.ProjectColumn{{Id: columnID, Name: "Doing"}},
	}

	baseArgs := func() map[string]any {
		return map[string]any{
			"task_id":           taskID.String(),
			"project_column_id": columnID.String(),
			"title":             "Task",
			"description":       "Desc",
			"priority":          "medium",
		}
	}

	newHandler := func() (*Handler, *stubTaskService) {
		svc := &stubTaskService{
			task: task,
			grouped: map[string]utils.CursorPaginated[domain.Task]{
				columnID.String(): {Data: []domain.Task{*task}},
			},
		}
		projectSvc := &stubProjectService{project: project}
		return NewHandler(nil, projectSvc, svc, &stubTaskCommentService{}), svc
	}

	principal := principal{
		APIKeyID: uuid.New(),
		UserID:   uuid.New(),
		Scopes:   domain.AllowedMCPAPIScopes,
	}

	t.Run("code absent: Code field is nil (let service preserve existing)", func(t *testing.T) {
		h, svc := newHandler()
		args := baseArgs()

		_, err := h.callTool(context.Background(), principal, toolCallParams{Name: "update_task", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, svc.updateRequest)
		assert.Nil(t, svc.updateRequest.Code, "absent code key should produce nil, signalling service to keep existing value")
	})

	t.Run("code null: Code field is nil (treated same as absent)", func(t *testing.T) {
		h, svc := newHandler()
		args := baseArgs()
		args["code"] = nil

		_, err := h.callTool(context.Background(), principal, toolCallParams{Name: "update_task", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, svc.updateRequest)
		assert.Nil(t, svc.updateRequest.Code, "null code key should produce nil, signalling service to keep existing value")
	})

	t.Run("code empty string: Code field is pointer to empty string (clear it)", func(t *testing.T) {
		h, svc := newHandler()
		args := baseArgs()
		args["code"] = ""

		_, err := h.callTool(context.Background(), principal, toolCallParams{Name: "update_task", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, svc.updateRequest)
		require.NotNil(t, svc.updateRequest.Code, "explicit empty string should produce a non-nil pointer so the service clears the code")
		assert.Equal(t, "", *svc.updateRequest.Code)
	})

	t.Run("code with value: Code field is pointer to trimmed string", func(t *testing.T) {
		h, svc := newHandler()
		args := baseArgs()
		args["code"] = "  NEW-99  "

		_, err := h.callTool(context.Background(), principal, toolCallParams{Name: "update_task", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, svc.updateRequest)
		require.NotNil(t, svc.updateRequest.Code)
		assert.Equal(t, "NEW-99", *svc.updateRequest.Code, "value should be trimmed")
	})
}

func TestCallToolMissingScope(t *testing.T) {
	handler := NewHandler(nil, &stubProjectService{}, &stubTaskService{}, &stubTaskCommentService{})

	_, err := handler.callTool(context.Background(), principal{
		UserID: uuid.New(),
		Scopes: []domain.MCPAPIScope{domain.MCPAPIScopeTasksRead},
	}, toolCallParams{
		Name: "move_task",
		Arguments: map[string]any{
			"task_id":                  uuid.NewString(),
			"project_id":               uuid.NewString(),
			"target_project_column_id": uuid.NewString(),
		},
	})
	require.Error(t, err)

	result := toolErrorResult(err)
	assert.Equal(t, "missing_scope", result["structuredContent"].(map[string]any)["error"].(map[string]any)["type"])
}

func TestToolSuccessResultIncludesSummaryAndStructuredJSON(t *testing.T) {
	projectID := uuid.New()
	projects := []domain.Project{
		{
			Id:   projectID,
			Name: "Alpha",
		},
	}

	result, err := toolSuccessResult("list_projects", map[string]any{
		"projects": projects,
	})
	require.NoError(t, err)

	content := result["content"].([]map[string]any)
	require.Len(t, content, 2)
	assert.Equal(t, "Listed 1 visible project(s).", content[0]["text"])
	assert.Contains(t, content[1]["text"], projectID.String())

	structured := result["structuredContent"].(map[string]any)
	projectsPayload := structured["projects"].([]domain.Project)
	require.Len(t, projectsPayload, 1)
	assert.Equal(t, "Alpha", projectsPayload[0].Name)

	var decoded map[string]any
	err = json.Unmarshal([]byte(content[1]["text"].(string)), &decoded)
	require.NoError(t, err)
	assert.Contains(t, decoded, "projects")
}

func TestToolErrorResultIncludesSummaryAndStructuredJSON(t *testing.T) {
	result := toolErrorResult(apperr.ForbiddenError("missing required scope"))

	content := result["content"].([]map[string]any)
	require.Len(t, content, 2)
	assert.Equal(t, "missing required scope", content[0]["text"])
	assert.Contains(t, content[1]["text"], `"type":"missing_scope"`)

	structured := result["structuredContent"].(map[string]any)
	assert.Equal(t, "missing_scope", structured["error"].(map[string]any)["type"])
}

func TestToolDefinitionsForPrincipalFiltersScopes(t *testing.T) {
	principal := principal{
		Scopes: []domain.MCPAPIScope{
			domain.MCPAPIScopeProjectsRead,
			domain.MCPAPIScopeTasksRead,
		},
	}

	tools := toolDefinitionsForPrincipal(principal)
	toolNames := make([]string, 0, len(tools))
	for _, tool := range tools {
		toolNames = append(toolNames, tool["name"].(string))
		assert.NotEmpty(t, tool["title"])
		assert.NotNil(t, tool["outputSchema"])
	}

	assert.Contains(t, toolNames, "list_projects")
	assert.Contains(t, toolNames, "find_task_by_code")
	assert.Contains(t, toolNames, "get_task")
	assert.NotContains(t, toolNames, "create_task")
	assert.NotContains(t, toolNames, "list_task_comments")
}

func TestReadResourceBuildsScopeAwareGuide(t *testing.T) {
	principal := principal{
		Scopes: []domain.MCPAPIScope{
			domain.MCPAPIScopeProjectsRead,
			domain.MCPAPIScopeTasksRead,
		},
	}

	result, err := readResource(serverGuideURI, principal)
	require.NoError(t, err)

	contents := result["contents"].([]map[string]any)
	require.Len(t, contents, 1)
	text := contents[0]["text"].(string)

	assert.Contains(t, text, "Recommended workflow")
	assert.Contains(t, text, "list_projects")
	assert.Contains(t, text, "find_task_by_code")
	assert.NotContains(t, text, "create_task")
}
