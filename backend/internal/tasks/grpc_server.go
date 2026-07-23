package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	tasksv1 "github.com/gabrielnakaema/project-chat/internal/tasks/v1"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const actionOriginMetadataKey = "x-action-origin"

type grpcTaskService interface {
	Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error)
	GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error)
	GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error)
	SearchTasks(ctx context.Context, request SearchTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	FindTaskByCode(ctx context.Context, request FindTaskByCodeRequest) (*domain.Task, error)
	Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error)
	Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error)
	MarkTaskDone(ctx context.Context, request MarkTaskDoneRequest) (*domain.Task, error)
	AssignTaskToSelf(ctx context.Context, request AssignTaskToSelfRequest) (*domain.Task, error)
}

type grpcTaskCommentService interface {
	Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error)
	ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error)
}

type GRPCServer struct {
	tasksv1.UnimplementedTaskServiceServer
	taskService        grpcTaskService
	taskCommentService grpcTaskCommentService
}

func NewGRPCServer(taskService grpcTaskService, taskCommentService grpcTaskCommentService) *GRPCServer {
	return &GRPCServer{taskService: taskService, taskCommentService: taskCommentService}
}

func (s *GRPCServer) CreateTask(ctx context.Context, req *tasksv1.CreateTaskRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	projectID, err := parseGRPCUUID(req.GetProjectId(), "project_id")
	if err != nil {
		return nil, err
	}
	columnID, err := parseGRPCUUID(req.GetProjectColumnId(), "project_column_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}
	responsibleID, err := parseOptionalGRPCUUID(req.ResponsibleId, "responsible_id")
	if err != nil {
		return nil, err
	}
	dueDate, err := parseOptionalGRPCTime(req.DueDate, "due_date")
	if err != nil {
		return nil, err
	}
	dependsOn, err := parseGRPCUUIDs(req.GetDependsOnTaskIds(), "depends_on_task_ids")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.Create(ctx, CreateTaskRequest{
		ProjectId:        projectID,
		ProjectColumnId:  columnID,
		Title:            req.GetTitle(),
		Description:      req.GetDescription(),
		Code:             req.GetCode(),
		RequestUserId:    userID,
		Priority:         req.GetPriority(),
		DueDate:          dueDate,
		ResponsibleId:    responsibleID,
		Tags:             req.GetTags(),
		DependsOnTaskIds: dependsOn,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) GetTaskById(ctx context.Context, req *tasksv1.GetTaskByIdRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.GetById(ctx, taskID, userID)
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) GroupTasksByColumn(ctx context.Context, req *tasksv1.GroupTasksByColumnRequest) (*tasksv1.TasksByColumnResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	projectID, err := parseGRPCUUID(req.GetProjectId(), "project_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}
	columnIDs, err := parseGRPCUUIDs(req.GetProjectColumnIds(), "project_column_ids")
	if err != nil {
		return nil, err
	}
	cursorUpdatedAt, err := parseOptionalGRPCTime(req.CursorUpdatedAt, "cursor_updated_at")
	if err != nil {
		return nil, err
	}

	grouped, err := s.taskService.GroupByColumn(ctx, GroupByColumnRequest{
		ProjectId:        projectID,
		UserId:           userID,
		ProjectColumnIDs: columnIDs,
		Archived:         req.GetArchived(),
		TaskOrder:        req.GetTaskOrder(),
		CursorUpdatedAt:  cursorUpdatedAt,
		Limit:            int(req.GetLimit()),
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}

	payload, err := json.Marshal(grouped)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode tasks")
	}
	return &tasksv1.TasksByColumnResponse{TasksByColumnJson: payload}, nil
}

func (s *GRPCServer) SearchTasks(ctx context.Context, req *tasksv1.SearchTasksRequest) (*tasksv1.SearchTasksResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	userID, err := parseGRPCUUID(req.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}
	projectID, err := parseOptionalGRPCUUID(req.ProjectId, "project_id")
	if err != nil {
		return nil, err
	}
	columnIDs, err := parseGRPCUUIDs(req.GetProjectColumnIds(), "project_column_ids")
	if err != nil {
		return nil, err
	}
	cursorDueDate, err := parseOptionalGRPCTime(req.CursorDueDate, "cursor_due_date")
	if err != nil {
		return nil, err
	}
	cursorUpdatedAt, err := parseOptionalGRPCTime(req.CursorUpdatedAt, "cursor_updated_at")
	if err != nil {
		return nil, err
	}
	cursorTaskID, err := parseOptionalGRPCUUID(req.CursorTaskId, "cursor_task_id")
	if err != nil {
		return nil, err
	}

	result, err := s.taskService.SearchTasks(ctx, SearchTasksRequest{
		UserId:           userID,
		ProjectId:        projectID,
		ProjectColumnIDs: columnIDs,
		SearchQuery:      req.GetQuery(),
		IncludeArchived:  req.GetIncludeArchived(),
		IncludeDone:      req.GetIncludeDone(),
		Limit:            int(req.GetLimit()),
		CursorDueDate:    cursorDueDate,
		CursorUpdatedAt:  cursorUpdatedAt,
		CursorTaskId:     cursorTaskID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode task search results")
	}
	return &tasksv1.SearchTasksResponse{ResultJson: payload}, nil
}

func (s *GRPCServer) FindTaskByCode(ctx context.Context, req *tasksv1.FindTaskByCodeRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	projectID, err := parseGRPCUUID(req.GetProjectId(), "project_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetUserId(), "user_id")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.FindTaskByCode(ctx, FindTaskByCodeRequest{
		ProjectId: projectID,
		UserId:    userID,
		Code:      req.GetCode(),
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) MoveTask(ctx context.Context, req *tasksv1.MoveTaskRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}
	projectID, err := parseGRPCUUID(req.GetProjectId(), "project_id")
	if err != nil {
		return nil, err
	}
	columnID, err := parseGRPCUUID(req.GetProjectColumnId(), "project_column_id")
	if err != nil {
		return nil, err
	}
	afterTaskID, err := parseOptionalGRPCUUID(req.AfterTaskId, "after_task_id")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.Move(ctx, MoveTaskRequest{
		TaskId:          taskID,
		RequestUserId:   userID,
		AfterTaskId:     afterTaskID,
		ProjectId:       projectID,
		ProjectColumnId: columnID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) UpdateTask(ctx context.Context, req *tasksv1.UpdateTaskRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	columnID, err := parseGRPCUUID(req.GetProjectColumnId(), "project_column_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}
	responsibleID, err := parseOptionalGRPCUUID(req.ResponsibleId, "responsible_id")
	if err != nil {
		return nil, err
	}
	dueDate, err := parseOptionalGRPCTime(req.DueDate, "due_date")
	if err != nil {
		return nil, err
	}
	dependsOn, err := parseGRPCUUIDs(req.GetDependsOnTaskIds(), "depends_on_task_ids")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.Update(ctx, UpdateTaskRequest{
		TaskId:           taskID,
		Title:            req.GetTitle(),
		Description:      req.GetDescription(),
		Code:             req.Code,
		ProjectColumnId:  columnID,
		RequestUserId:    userID,
		Priority:         domain.TaskPriority(req.GetPriority()),
		DueDate:          dueDate,
		ResponsibleId:    responsibleID,
		Tags:             req.GetTags(),
		DependsOnTaskIds: dependsOn,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) MarkTaskDone(ctx context.Context, req *tasksv1.MarkTaskDoneRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.MarkTaskDone(ctx, MarkTaskDoneRequest{
		TaskId:        taskID,
		RequestUserId: userID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) AssignTaskToSelf(ctx context.Context, req *tasksv1.AssignTaskToSelfRequest) (*tasksv1.TaskResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}

	task, err := s.taskService.AssignTaskToSelf(ctx, AssignTaskToSelfRequest{
		TaskId:        taskID,
		RequestUserId: userID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}
	return taskResponse(task)
}

func (s *GRPCServer) CreateTaskComment(ctx context.Context, req *tasksv1.CreateTaskCommentRequest) (*tasksv1.TaskCommentResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}
	parentID, err := parseOptionalGRPCUUID(req.ParentCommentId, "parent_comment_id")
	if err != nil {
		return nil, err
	}

	comment, err := s.taskCommentService.Create(ctx, CreateTaskCommentRequest{
		TaskID:          taskID,
		RequestUserID:   userID,
		Content:         req.GetContent(),
		ParentCommentID: parentID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}

	payload, err := json.Marshal(comment)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode task comment")
	}
	return &tasksv1.TaskCommentResponse{CommentJson: payload}, nil
}

func (s *GRPCServer) ListTaskComments(ctx context.Context, req *tasksv1.ListTaskCommentsRequest) (*tasksv1.TaskCommentsResponse, error) {
	ctx = withActionOriginFromMetadata(ctx)

	taskID, err := parseGRPCUUID(req.GetTaskId(), "task_id")
	if err != nil {
		return nil, err
	}
	userID, err := parseGRPCUUID(req.GetRequestUserId(), "request_user_id")
	if err != nil {
		return nil, err
	}
	before, err := parseOptionalGRPCTime(req.Before, "before")
	if err != nil {
		return nil, err
	}
	beforeCommentID, err := parseOptionalGRPCUUID(req.BeforeCommentId, "before_comment_id")
	if err != nil {
		return nil, err
	}
	after, err := parseOptionalGRPCTime(req.After, "after")
	if err != nil {
		return nil, err
	}
	afterCommentID, err := parseOptionalGRPCUUID(req.AfterCommentId, "after_comment_id")
	if err != nil {
		return nil, err
	}

	comments, err := s.taskCommentService.ListByTaskID(ctx, ListTaskCommentsRequest{
		TaskID:          taskID,
		RequestUserID:   userID,
		Limit:           int(req.GetLimit()),
		Before:          before,
		BeforeCommentID: beforeCommentID,
		After:           after,
		AfterCommentID:  afterCommentID,
	})
	if err != nil {
		return nil, domainErrorToStatus(err)
	}

	payload, err := json.Marshal(comments)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode task comments")
	}
	return &tasksv1.TaskCommentsResponse{CommentsJson: payload}, nil
}

func taskResponse(task *domain.Task) (*tasksv1.TaskResponse, error) {
	payload, err := json.Marshal(task)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to encode task")
	}
	return &tasksv1.TaskResponse{TaskJson: payload}, nil
}

func withActionOriginFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	values := md.Get(actionOriginMetadataKey)
	if len(values) == 0 || values[0] == "" {
		return ctx
	}
	return domain.WithActionOrigin(ctx, domain.ActionOrigin(values[0]))
}

func parseGRPCUUID(value string, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "valid %s is required", field)
	}
	return id, nil
}

func parseOptionalGRPCUUID(value *string, field string) (*uuid.UUID, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s must be a valid uuid", field)
	}
	return &id, nil
}

func parseGRPCUUIDs(values []string, field string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s contains an invalid uuid", field)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func parseOptionalGRPCTime(value *string, field string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s must be an RFC3339 timestamp", field)
	}
	return &parsed, nil
}

func domainErrorToStatus(err error) error {
	var domainErr apperr.DomainError
	if !errors.As(err, &domainErr) {
		return status.Error(codes.Internal, "task operation failed")
	}
	return status.Error(grpcCodeForDomainCode(domainErr.Code), domainErr.Message)
}

func grpcCodeForDomainCode(code apperr.ErrorCode) codes.Code {
	switch code {
	case apperr.NotFoundErrorCode:
		return codes.NotFound
	case apperr.UnauthorizedErrorCode:
		return codes.Unauthenticated
	case apperr.ForbiddenErrorCode:
		return codes.PermissionDenied
	case apperr.ValidationFailedErrorCode:
		return codes.InvalidArgument
	case apperr.BusinessValidationErrorCode:
		return codes.FailedPrecondition
	case apperr.DuplicateEntryErrorCode:
		return codes.AlreadyExists
	default:
		return codes.Internal
	}
}
