package app

import (
	"context"

	"github.com/gabrielnakaema/project-chat/internal/apikey"
	apikeyv1 "github.com/gabrielnakaema/project-chat/internal/apikey/v1"
	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/mcp"
	"github.com/gabrielnakaema/project-chat/internal/project"
	projectv1 "github.com/gabrielnakaema/project-chat/internal/project/v1"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	tasksv1 "github.com/gabrielnakaema/project-chat/internal/tasks/v1"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type apiKeyService struct {
	client *apikey.Client
}

func newAPIKeyService(client apikeyv1.APIKeyServiceClient) *apiKeyService {
	return &apiKeyService{client: apikey.NewClient(client)}
}

func (s *apiKeyService) Authenticate(ctx context.Context, bearerSecret string) (*mcp.AuthenticatedAPIKey, error) {
	key, err := s.client.Authenticate(ctx, bearerSecret)
	if err != nil {
		return nil, err
	}
	return &mcp.AuthenticatedAPIKey{ID: key.ID, UserID: key.UserID, Scopes: key.Scopes}, nil
}

type projectService struct {
	client *project.Client
}

func newProjectService(client projectv1.ProjectServiceClient) *projectService {
	return &projectService{client: project.NewClient(client)}
}

func (s *projectService) ListByUserID(ctx context.Context, request mcp.ListProjectsRequest) ([]domain.Project, error) {
	return s.client.List(ctx, project.ListRequest{
		UserID:             request.UserID,
		MemberRole:         request.MemberRole,
		ShouldFilterByRole: request.ShouldFilterByRole,
		SearchQuery:        request.SearchQuery,
	})
}

func (s *projectService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Project, error) {
	return s.client.Get(ctx, id, userID)
}

type taskService struct {
	client *tasks.TaskServiceClient
}

func newTaskService(client tasksv1.TaskServiceClient) *taskService {
	return &taskService{client: tasks.NewTaskServiceClient(client)}
}

func (s *taskService) Create(ctx context.Context, request mcp.CreateTaskRequest) (*domain.Task, error) {
	return s.client.Create(ctx, tasks.CreateTaskRequest{
		ProjectId: request.ProjectID, ProjectColumnId: request.ProjectColumnID,
		Title: request.Title, Description: request.Description, Code: request.Code,
		RequestUserId: request.RequestUserID, Priority: request.Priority,
		DueDate: request.DueDate, ResponsibleId: request.ResponsibleID,
		Tags: request.Tags, DependsOnTaskIds: request.DependsOnTaskIDs,
	})
}

func (s *taskService) GroupByColumn(ctx context.Context, request mcp.GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error) {
	return s.client.GroupByColumn(ctx, tasks.GroupByColumnRequest{
		ProjectId: request.ProjectID, UserId: request.UserID,
		ProjectColumnIDs: request.ProjectColumnIDs, Archived: request.Archived,
		TaskOrder: request.TaskOrder, CursorUpdatedAt: request.CursorUpdatedAt, Limit: request.Limit,
	})
}

func (s *taskService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
	return s.client.GetById(ctx, id, userID)
}

func (s *taskService) FindTaskByCode(ctx context.Context, request mcp.FindTaskByCodeRequest) (*domain.Task, error) {
	return s.client.FindTaskByCode(ctx, tasks.FindTaskByCodeRequest{ProjectId: request.ProjectID, UserId: request.UserID, Code: request.Code})
}

func (s *taskService) Move(ctx context.Context, request mcp.MoveTaskRequest) (*domain.Task, error) {
	return s.client.Move(ctx, tasks.MoveTaskRequest{
		TaskId: request.TaskID, RequestUserId: request.RequestUserID, AfterTaskId: request.AfterTaskID,
		ProjectId: request.ProjectID, ProjectColumnId: request.ProjectColumnID,
	})
}

func (s *taskService) Update(ctx context.Context, request mcp.UpdateTaskRequest) (*domain.Task, error) {
	return s.client.Update(ctx, tasks.UpdateTaskRequest{
		TaskId: request.TaskID, Title: request.Title, Description: request.Description, Code: request.Code,
		ProjectColumnId: request.ProjectColumnID, RequestUserId: request.RequestUserID, Priority: request.Priority,
		DueDate: request.DueDate, ResponsibleId: request.ResponsibleID, Tags: request.Tags,
		DependsOnTaskIds: request.DependsOnTaskIDs,
	})
}

func (s *taskService) MarkTaskDone(ctx context.Context, request mcp.MarkTaskDoneRequest) (*domain.Task, error) {
	return s.client.MarkTaskDone(ctx, tasks.MarkTaskDoneRequest{TaskId: request.TaskID, RequestUserId: request.RequestUserID})
}

func (s *taskService) AssignTaskToSelf(ctx context.Context, request mcp.AssignTaskToSelfRequest) (*domain.Task, error) {
	return s.client.AssignTaskToSelf(ctx, tasks.AssignTaskToSelfRequest{TaskId: request.TaskID, RequestUserId: request.RequestUserID})
}

type taskCommentService struct {
	client *tasks.TaskCommentServiceClient
}

func newTaskCommentService(client tasksv1.TaskServiceClient) *taskCommentService {
	return &taskCommentService{client: tasks.NewTaskCommentServiceClient(client)}
}

func (s *taskCommentService) Create(ctx context.Context, request mcp.CreateTaskCommentRequest) (*domain.TaskComment, error) {
	return s.client.Create(ctx, tasks.CreateTaskCommentRequest{
		TaskID: request.TaskID, RequestUserID: request.RequestUserID,
		Content: request.Content, ParentCommentID: request.ParentCommentID,
	})
}

func (s *taskCommentService) ListByTaskID(ctx context.Context, request mcp.ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error) {
	return s.client.ListByTaskID(ctx, tasks.ListTaskCommentsRequest{
		TaskID: request.TaskID, RequestUserID: request.RequestUserID, Limit: request.Limit,
		Before: request.Before, BeforeCommentID: request.BeforeCommentID,
		After: request.After, AfterCommentID: request.AfterCommentID,
	})
}
