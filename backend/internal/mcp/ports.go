package mcp

import (
	"context"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type AuthenticatedAPIKey struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Scopes []domain.MCPAPIScope
}

type authService interface {
	Authenticate(ctx context.Context, bearerSecret string) (*AuthenticatedAPIKey, error)
}

type ListProjectsRequest struct {
	UserID             uuid.UUID
	MemberRole         domain.ProjectMemberRole
	ShouldFilterByRole bool
	SearchQuery        string
}

type projectService interface {
	ListByUserID(ctx context.Context, request ListProjectsRequest) ([]domain.Project, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Project, error)
}

type CreateTaskRequest struct {
	ProjectID        uuid.UUID
	ProjectColumnID  uuid.UUID
	Title            string
	Description      string
	Code             string
	RequestUserID    uuid.UUID
	Priority         string
	DueDate          *time.Time
	ResponsibleID    *uuid.UUID
	Tags             []string
	DependsOnTaskIDs []uuid.UUID
}

type GroupByColumnRequest struct {
	ProjectID        uuid.UUID
	UserID           uuid.UUID
	ProjectColumnIDs []uuid.UUID
	Archived         bool
	TaskOrder        string
	CursorUpdatedAt  *time.Time
	Limit            int
}

type FindTaskByCodeRequest struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	Code      string
}

type SearchTasksRequest struct {
	ProjectID        uuid.UUID
	UserID           uuid.UUID
	ProjectColumnIDs []uuid.UUID
	Query            string
	IncludeArchived  bool
	Limit            int
	CursorDueDate    *time.Time
	CursorUpdatedAt  *time.Time
	CursorTaskID     *uuid.UUID
}

type MoveTaskRequest struct {
	TaskID          uuid.UUID
	RequestUserID   uuid.UUID
	AfterTaskID     *uuid.UUID
	ProjectID       uuid.UUID
	ProjectColumnID uuid.UUID
}

type UpdateTaskRequest struct {
	TaskID           uuid.UUID
	Title            string
	Description      string
	Code             *string
	ProjectColumnID  uuid.UUID
	RequestUserID    uuid.UUID
	Priority         domain.TaskPriority
	DueDate          *time.Time
	ResponsibleID    *uuid.UUID
	Tags             []string
	DependsOnTaskIDs []uuid.UUID
}

type MarkTaskDoneRequest struct {
	TaskID        uuid.UUID
	RequestUserID uuid.UUID
}

type AssignTaskToSelfRequest struct {
	TaskID        uuid.UUID
	RequestUserID uuid.UUID
}

type CreateTaskCommentRequest struct {
	TaskID          uuid.UUID
	RequestUserID   uuid.UUID
	Content         string
	ParentCommentID *uuid.UUID
}

type ListTaskCommentsRequest struct {
	TaskID          uuid.UUID
	RequestUserID   uuid.UUID
	Limit           int
	Before          *time.Time
	BeforeCommentID *uuid.UUID
	After           *time.Time
	AfterCommentID  *uuid.UUID
}

type taskService interface {
	Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error)
	GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error)
	SearchTasks(ctx context.Context, request SearchTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
	FindTaskByCode(ctx context.Context, request FindTaskByCodeRequest) (*domain.Task, error)
	Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error)
	Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error)
	MarkTaskDone(ctx context.Context, request MarkTaskDoneRequest) (*domain.Task, error)
	AssignTaskToSelf(ctx context.Context, request AssignTaskToSelfRequest) (*domain.Task, error)
}

type taskCommentService interface {
	Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error)
	ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error)
}
