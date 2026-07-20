package tasks

import (
	"context"
	"encoding/json"
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

// TaskServiceClient adapts the tasks gRPC client to the task-service interface
// the MCP handler depends on, so api can call tasks-service without
// instantiating the tasks slice in-process.
type TaskServiceClient struct {
	client tasksv1.TaskServiceClient
}

func NewTaskServiceClient(client tasksv1.TaskServiceClient) *TaskServiceClient {
	return &TaskServiceClient{client: client}
}

func (c *TaskServiceClient) Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error) {
	resp, err := c.client.CreateTask(withOutgoingActionOrigin(ctx), &tasksv1.CreateTaskRequest{
		ProjectId:        request.ProjectId.String(),
		ProjectColumnId:  request.ProjectColumnId.String(),
		Title:            request.Title,
		Description:      request.Description,
		Code:             request.Code,
		RequestUserId:    request.RequestUserId.String(),
		Priority:         request.Priority,
		DueDate:          optionalTimeString(request.DueDate),
		ResponsibleId:    optionalUUIDString(request.ResponsibleId),
		Tags:             request.Tags,
		DependsOnTaskIds: uuidStrings(request.DependsOnTaskIds),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error) {
	resp, err := c.client.GetTaskById(withOutgoingActionOrigin(ctx), &tasksv1.GetTaskByIdRequest{
		TaskId: id.String(),
		UserId: userId.String(),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error) {
	resp, err := c.client.GroupTasksByColumn(withOutgoingActionOrigin(ctx), &tasksv1.GroupTasksByColumnRequest{
		ProjectId:        request.ProjectId.String(),
		UserId:           request.UserId.String(),
		ProjectColumnIds: uuidStrings(request.ProjectColumnIDs),
		Archived:         request.Archived,
		TaskOrder:        request.TaskOrder,
		CursorUpdatedAt:  optionalTimeString(request.CursorUpdatedAt),
		Limit:            int32(request.Limit),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}

	var grouped map[string]utils.CursorPaginated[domain.Task]
	if err := json.Unmarshal(resp.GetTasksByColumnJson(), &grouped); err != nil {
		return nil, apperr.ServerError("failed to decode tasks", err)
	}
	return grouped, nil
}

func (c *TaskServiceClient) FindTaskByCode(ctx context.Context, request FindTaskByCodeRequest) (*domain.Task, error) {
	resp, err := c.client.FindTaskByCode(withOutgoingActionOrigin(ctx), &tasksv1.FindTaskByCodeRequest{
		ProjectId: request.ProjectId.String(),
		UserId:    request.UserId.String(),
		Code:      request.Code,
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error) {
	resp, err := c.client.MoveTask(withOutgoingActionOrigin(ctx), &tasksv1.MoveTaskRequest{
		TaskId:          request.TaskId.String(),
		RequestUserId:   request.RequestUserId.String(),
		AfterTaskId:     optionalUUIDString(request.AfterTaskId),
		ProjectId:       request.ProjectId.String(),
		ProjectColumnId: request.ProjectColumnId.String(),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error) {
	resp, err := c.client.UpdateTask(withOutgoingActionOrigin(ctx), &tasksv1.UpdateTaskRequest{
		TaskId:           request.TaskId.String(),
		Title:            request.Title,
		Description:      request.Description,
		Code:             request.Code,
		ProjectColumnId:  request.ProjectColumnId.String(),
		RequestUserId:    request.RequestUserId.String(),
		Priority:         string(request.Priority),
		DueDate:          optionalTimeString(request.DueDate),
		ResponsibleId:    optionalUUIDString(request.ResponsibleId),
		Tags:             request.Tags,
		DependsOnTaskIds: uuidStrings(request.DependsOnTaskIds),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) MarkTaskDone(ctx context.Context, request MarkTaskDoneRequest) (*domain.Task, error) {
	resp, err := c.client.MarkTaskDone(withOutgoingActionOrigin(ctx), &tasksv1.MarkTaskDoneRequest{
		TaskId:        request.TaskId.String(),
		RequestUserId: request.RequestUserId.String(),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

func (c *TaskServiceClient) AssignTaskToSelf(ctx context.Context, request AssignTaskToSelfRequest) (*domain.Task, error) {
	resp, err := c.client.AssignTaskToSelf(withOutgoingActionOrigin(ctx), &tasksv1.AssignTaskToSelfRequest{
		TaskId:        request.TaskId.String(),
		RequestUserId: request.RequestUserId.String(),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}
	return decodeTask(resp.GetTaskJson())
}

// TaskCommentServiceClient adapts the tasks gRPC client to the task-comment
// service interface the MCP handler depends on.
type TaskCommentServiceClient struct {
	client tasksv1.TaskServiceClient
}

func NewTaskCommentServiceClient(client tasksv1.TaskServiceClient) *TaskCommentServiceClient {
	return &TaskCommentServiceClient{client: client}
}

func (c *TaskCommentServiceClient) Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error) {
	resp, err := c.client.CreateTaskComment(withOutgoingActionOrigin(ctx), &tasksv1.CreateTaskCommentRequest{
		TaskId:          request.TaskID.String(),
		RequestUserId:   request.RequestUserID.String(),
		Content:         request.Content,
		ParentCommentId: optionalUUIDString(request.ParentCommentID),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}

	var comment domain.TaskComment
	if err := json.Unmarshal(resp.GetCommentJson(), &comment); err != nil {
		return nil, apperr.ServerError("failed to decode task comment", err)
	}
	return &comment, nil
}

func (c *TaskCommentServiceClient) ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error) {
	resp, err := c.client.ListTaskComments(withOutgoingActionOrigin(ctx), &tasksv1.ListTaskCommentsRequest{
		TaskId:          request.TaskID.String(),
		RequestUserId:   request.RequestUserID.String(),
		Limit:           int32(request.Limit),
		Before:          optionalTimeString(request.Before),
		BeforeCommentId: optionalUUIDString(request.BeforeCommentID),
		After:           optionalTimeString(request.After),
		AfterCommentId:  optionalUUIDString(request.AfterCommentID),
	})
	if err != nil {
		return nil, statusToDomainError(err)
	}

	var comments utils.CursorPaginated[domain.TaskComment]
	if err := json.Unmarshal(resp.GetCommentsJson(), &comments); err != nil {
		return nil, apperr.ServerError("failed to decode task comments", err)
	}
	return &comments, nil
}

func decodeTask(payload []byte) (*domain.Task, error) {
	var task domain.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		return nil, apperr.ServerError("failed to decode task", err)
	}
	return &task, nil
}

func withOutgoingActionOrigin(ctx context.Context) context.Context {
	origin := domain.ActionOriginFromContext(ctx)
	return metadata.AppendToOutgoingContext(ctx, actionOriginMetadataKey, string(origin))
}

func optionalUUIDString(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	value := id.String()
	return &value
}

func optionalTimeString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	value := t.Format(time.RFC3339)
	return &value
}

func uuidStrings(ids []uuid.UUID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}
	return values
}

func statusToDomainError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return apperr.ServerError("task service call failed", err)
	}
	message := st.Message()
	switch st.Code() {
	case codes.NotFound:
		return apperr.NotFoundError(message)
	case codes.Unauthenticated:
		return apperr.UnauthorizedError(message)
	case codes.PermissionDenied:
		return apperr.ForbiddenError(message)
	case codes.InvalidArgument:
		return apperr.BusinessValidationError(message)
	case codes.FailedPrecondition:
		return apperr.BusinessValidationError(message)
	case codes.AlreadyExists:
		return apperr.DuplicateEntryError(message)
	default:
		return apperr.ServerError(message, err)
	}
}
