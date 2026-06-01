package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProjectService struct {
	projects []domain.Project
	project  *domain.Project
}

func (s *stubProjectService) ListByUserId(ctx context.Context, request service.ListProjectsByUserIdRequest) ([]domain.Project, error) {
	return s.projects, nil
}

func (s *stubProjectService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error) {
	return s.project, nil
}

type stubTaskService struct {
	task              *domain.Task
	grouped           map[string]utils.CursorPaginated[domain.Task]
	moveRequest       *service.MoveTaskRequest
	markDoneRequest   *service.MarkTaskDoneRequest
	assignSelfRequest *service.AssignTaskToSelfRequest
	moveOrigin        domain.ActionOrigin
	markDoneOrigin    domain.ActionOrigin
	assignSelfOrigin  domain.ActionOrigin
}

func (s *stubTaskService) GroupByColumn(ctx context.Context, request service.GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error) {
	return s.grouped, nil
}

func (s *stubTaskService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error) {
	return s.task, nil
}

func (s *stubTaskService) Move(ctx context.Context, request service.MoveTaskRequest) (*domain.Task, error) {
	s.moveRequest = &request
	s.moveOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) Update(ctx context.Context, request service.UpdateTaskRequest) (*domain.Task, error) {
	return s.task, nil
}

func (s *stubTaskService) MarkTaskDone(ctx context.Context, request service.MarkTaskDoneRequest) (*domain.Task, error) {
	s.markDoneRequest = &request
	s.markDoneOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

func (s *stubTaskService) AssignTaskToSelf(ctx context.Context, request service.AssignTaskToSelfRequest) (*domain.Task, error) {
	s.assignSelfRequest = &request
	s.assignSelfOrigin = domain.ActionOriginFromContext(ctx)
	return s.task, nil
}

type stubTaskCommentService struct {
	comment       *domain.TaskComment
	comments      *utils.CursorPaginated[domain.TaskComment]
	createRequest *service.CreateTaskCommentRequest
	createOrigin  domain.ActionOrigin
	listRequest   *service.ListTaskCommentsRequest
}

func (s *stubTaskCommentService) Create(ctx context.Context, request service.CreateTaskCommentRequest) (*domain.TaskComment, error) {
	s.createRequest = &request
	s.createOrigin = domain.ActionOriginFromContext(ctx)
	return s.comment, nil
}

func (s *stubTaskCommentService) ListByTaskID(ctx context.Context, request service.ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error) {
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
	assert.Equal(t, taskID, taskSvc.markDoneRequest.TaskId)

	_, err = handler.callTool(context.Background(), principal, toolCallParams{
		Name: "assign_task_to_self",
		Arguments: map[string]any{
			"task_id": taskID.String(),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ActionOriginMCPAgent, taskSvc.assignSelfOrigin)
	require.NotNil(t, taskSvc.assignSelfRequest)
	assert.Equal(t, taskID, taskSvc.assignSelfRequest.TaskId)
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
