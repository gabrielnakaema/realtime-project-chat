package tasks

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type taskRepository interface {
	CountTasksByProjectIdAndColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID) (map[string]int, error)
	Create(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error
	CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error
	FindTaskRefsByProjectAndCode(ctx context.Context, projectId uuid.UUID, code string) ([]domain.TaskDependencyRef, error)
	GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	GetFirstTaskInColumn(ctx context.Context, projectId uuid.UUID, projectColumnID uuid.UUID) (*domain.Task, error)
	GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error)
	GetTaskDependencyRefsByProjectAndIds(ctx context.Context, projectId uuid.UUID, taskIds []uuid.UUID) ([]domain.TaskDependencyRef, error)
	ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	ListTaskDependenciesByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.TaskDependencyEdge, error)
	ListUserDueTasks(ctx context.Context, userId uuid.UUID, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	MoveTask(ctx context.Context, placement *domain.TaskPlacement, userId uuid.UUID, buildEvents func(*domain.Task) []outbox.Message) (*domain.Task, error)
	SearchProjectTasksForDependencies(ctx context.Context, projectId uuid.UUID, query string, excludeTaskId *uuid.UUID, limit int) ([]domain.TaskDependencyRef, error)
	SearchTasks(ctx context.Context, request SearchTasksRequest) (*utils.CursorPaginated[domain.Task], error)
	SuggestTaskCodesByProjectPrefix(ctx context.Context, projectId uuid.UUID, prefix string, limit int) ([]domain.TaskCodeSuggestion, error)
	Update(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error
	WithProjectColumnMoveLock(ctx context.Context, projectColumnID uuid.UUID, fn func(context.Context) error) error
}

type taskServiceProjectRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error)
}

type taskServiceUserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type TaskService struct {
	taskRepository    taskRepository
	projectRepository taskServiceProjectRepository
	userRepository    taskServiceUserRepository
}

func NewTaskService(taskRepository taskRepository, projectRepository taskServiceProjectRepository, userRepository taskServiceUserRepository) *TaskService {
	return &TaskService{
		taskRepository:    taskRepository,
		projectRepository: projectRepository,
		userRepository:    userRepository,
	}
}

type CreateTaskRequest struct {
	ProjectId        uuid.UUID
	ProjectColumnId  uuid.UUID
	Title            string
	Description      string
	Code             string
	RequestUserId    uuid.UUID
	Priority         string
	DueDate          *time.Time
	ResponsibleId    *uuid.UUID
	Tags             []string
	DependsOnTaskIds []uuid.UUID
}

func (ts *TaskService) Create(ctx context.Context, request CreateTaskRequest) (*domain.Task, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, apperr.NotFoundError("project not found")
			}
			return nil, domainErr
		}
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		if !project.IsMember(*request.ResponsibleId) {
			return nil, apperr.BusinessValidationError("responsible is not a member of the project")
		}
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		return nil, apperr.ServerError("failed to get user", err)
	}

	var responsible *domain.User
	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		responsible, err = ts.userRepository.GetById(ctx, *request.ResponsibleId)
		if err != nil {
			return nil, apperr.ServerError("failed to get responsible user", err)
		}
	}

	firstTask, err := ts.taskRepository.GetFirstTaskInColumn(ctx, request.ProjectId, request.ProjectColumnId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code != apperr.NotFoundErrorCode {
				return nil, apperr.ServerError("failed to get first task in column", err)
			}
		} else {
			return nil, apperr.ServerError("failed to get first task in column", err)
		}
	}

	var nextOrder string
	if firstTask != nil {
		nextOrder = firstTask.Order
	}

	order, err := fracindex.GenerateKeyBetween("", nextOrder)
	if err != nil {
		return nil, apperr.ServerError("failed to generate task order", err)
	}

	dependsOnRefs, err := ts.validateDependsOnTaskIds(ctx, request.ProjectId, uuid.Nil, request.DependsOnTaskIds)
	if err != nil {
		return nil, err
	}

	task := domain.Task{
		ProjectId:        request.ProjectId,
		Title:            request.Title,
		Description:      request.Description,
		Code:             strings.TrimSpace(request.Code),
		Status:           domain.TaskStatusIn(projectColumn, nil),
		AuthorId:         request.RequestUserId,
		ProjectColumnId:  request.ProjectColumnId,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Author:           user,
		Priority:         domain.TaskPriority(request.Priority),
		Order:            order,
		ResponsibleId:    request.ResponsibleId,
		Responsible:      responsible,
		DueDate:          request.DueDate,
		Tags:             domain.NormalizeTaskTags(request.Tags),
		DependsOnTaskIds: domain.TaskDependencyRefsToIDs(dependsOnRefs),
		DependsOnTasks:   dependsOnRefs,
		DoneAt:           projectColumn.CompletedAt(nil),
		Updates:          []domain.TaskUpdate{},
		ProjectColumn:    projectColumn,
	}

	err = ts.taskRepository.Create(ctx, &task, func(t *domain.Task) []outbox.Message {
		return []outbox.Message{{
			Topic:       events.TaskCreated,
			AggregateID: t.Id,
			Payload: &events.TaskCreatedPayload{
				Task:         *t,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.RequestUserId,
				},
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to create task", err)
	}

	return &task, nil
}

type UpdateTaskRequest struct {
	TaskId           uuid.UUID
	Title            string
	Description      string
	Code             *string
	ProjectColumnId  uuid.UUID
	RequestUserId    uuid.UUID
	Priority         domain.TaskPriority
	DueDate          *time.Time
	ResponsibleId    *uuid.UUID
	Tags             []string
	DependsOnTaskIds []uuid.UUID
}

func (ts *TaskService) Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error) {
	task, project, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	return ts.updateLoadedTask(ctx, task, project, request)
}

type ListTasksRequest struct {
	ProjectId        uuid.UUID
	RequestUserId    uuid.UUID
	ProjectColumnIDs []uuid.UUID
	Archived         bool
	TaskOrder        string
	Limit            int
	CursorUpdatedAt  *time.Time
}

func (ts *TaskService) List(ctx context.Context, request ListTasksRequest) (*utils.CursorPaginated[domain.Task], error) {
	if request.RequestUserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, apperr.BusinessValidationError("project_id is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	if request.Archived && !project.IsCreator(request.RequestUserId) {
		return nil, apperr.ForbiddenError("only the project creator can view archived tasks")
	}

	if err := validateProjectColumnIDs(project, request.ProjectColumnIDs); err != nil {
		return nil, err
	}

	tasks, err := ts.taskRepository.ListByProjectId(ctx, request.ProjectId, request.ProjectColumnIDs, request.Archived, request.TaskOrder, request.CursorUpdatedAt, request.Limit)
	if err != nil {
		return nil, apperr.ServerError("failed to list tasks", err)
	}

	return tasks, nil
}

func (ts *TaskService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error) {
	task, _, err := ts.getTaskAndProjectForUser(ctx, id, userId)
	if err != nil {
		return nil, err
	}
	return task, nil
}

type GroupByColumnRequest struct {
	ProjectId        uuid.UUID
	UserId           uuid.UUID
	ProjectColumnIDs []uuid.UUID
	Archived         bool
	TaskOrder        string
	CursorUpdatedAt  *time.Time
	Limit            int
}

func (ts *TaskService) GroupByColumn(ctx context.Context, request GroupByColumnRequest) (map[string]utils.CursorPaginated[domain.Task], error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	if request.Archived && !project.IsCreator(request.UserId) {
		return nil, apperr.ForbiddenError("only the project creator can view archived tasks")
	}

	statusIDs := request.ProjectColumnIDs
	if len(statusIDs) == 0 {
		for _, column := range project.Columns {
			statusIDs = append(statusIDs, column.Id)
		}
	}

	if err := validateProjectColumnIDs(project, statusIDs); err != nil {
		return nil, err
	}

	results := map[string]utils.CursorPaginated[domain.Task]{}

	for _, statusID := range statusIDs {
		result, err := ts.taskRepository.ListByProjectId(ctx, request.ProjectId, []uuid.UUID{statusID}, request.Archived, request.TaskOrder, request.CursorUpdatedAt, request.Limit)
		if err != nil {
			return nil, apperr.ServerError("failed to list tasks", err)
		}
		results[statusID.String()] = *result
	}

	return results, nil
}

func (ts *TaskService) CountByColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, requestUserId uuid.UUID) (map[string]int, error) {
	if requestUserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, projectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(requestUserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	if err := validateProjectColumnIDs(project, projectColumnIDs); err != nil {
		return nil, err
	}
	results, err := ts.taskRepository.CountTasksByProjectIdAndColumn(ctx, projectId, projectColumnIDs)
	if err != nil {
		return nil, apperr.ServerError("failed to count tasks", err)
	}
	return results, nil
}

type ArchiveTaskRequest struct {
	TaskId        uuid.UUID
	RequestUserId uuid.UUID
}

func (ts *TaskService) Archive(ctx context.Context, request ArchiveTaskRequest) (*domain.Task, error) {
	task, _, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	archivedTask, err := task.Archive()
	if err != nil {
		return nil, archiveTransitionError(err)
	}

	err = ts.taskRepository.Update(ctx, archivedTask, func(updated *domain.Task) []outbox.Message {
		return []outbox.Message{{
			Topic:       events.TaskUpdated,
			AggregateID: updated.Id,
			Payload: &events.TaskUpdatedPayload{
				Task:         *updated,
				PreviousTask: task,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.RequestUserId,
				},
				PreviousProjectColumnID: &task.ProjectColumnId,
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to archive task", err)
	}

	return archivedTask, nil
}

type RestoreTaskRequest struct {
	TaskId          uuid.UUID
	ProjectColumnId uuid.UUID
	RequestUserId   uuid.UUID
}

func (ts *TaskService) Restore(ctx context.Context, request RestoreTaskRequest) (*domain.Task, error) {
	task, project, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	restoredTask, err := task.Restore(project, request.ProjectColumnId)
	if err != nil {
		return nil, archiveTransitionError(err)
	}

	err = ts.taskRepository.Update(ctx, restoredTask, func(updated *domain.Task) []outbox.Message {
		return []outbox.Message{{
			Topic:       events.TaskUpdated,
			AggregateID: updated.Id,
			Payload: &events.TaskUpdatedPayload{
				Task:         *updated,
				PreviousTask: task,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.RequestUserId,
				},
				PreviousProjectColumnID: nil,
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to restore task", err)
	}

	return restoredTask, nil
}

type ListUserDueTasksRequest struct {
	UserId          uuid.UUID
	Limit           int
	CursorDueDate   *time.Time
	CursorUpdatedAt *time.Time
}

func (ts *TaskService) ListUserDueTasks(ctx context.Context, request ListUserDueTasksRequest) (*utils.CursorPaginated[domain.Task], error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	result, err := ts.taskRepository.ListUserDueTasks(ctx, request.UserId, request.CursorDueDate, request.CursorUpdatedAt, request.Limit)
	if err != nil {
		return nil, apperr.ServerError("failed to list user due tasks", err)
	}

	return result, nil
}

type MoveTaskRequest struct {
	TaskId          uuid.UUID
	RequestUserId   uuid.UUID
	AfterTaskId     *uuid.UUID
	ProjectId       uuid.UUID
	ProjectColumnId uuid.UUID
}

func (ts *TaskService) Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error) {
	oldTask, project, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	if oldTask.ProjectId != request.ProjectId {
		return nil, apperr.BusinessValidationError("task does not belong to the provided project_id")
	}

	return ts.moveLoadedTask(ctx, oldTask, project, request)
}

type MarkTaskDoneRequest struct {
	TaskId        uuid.UUID
	RequestUserId uuid.UUID
}

func (ts *TaskService) MarkTaskDone(ctx context.Context, request MarkTaskDoneRequest) (*domain.Task, error) {
	task, project, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	doneColumn, err := resolveDoneColumn(project)
	if err != nil {
		return nil, err
	}

	if task.IsCompletedIn(doneColumn) {
		return task, nil
	}

	return ts.moveLoadedTask(ctx, task, project, MoveTaskRequest{
		TaskId:          task.Id,
		RequestUserId:   request.RequestUserId,
		ProjectId:       task.ProjectId,
		ProjectColumnId: doneColumn.Id,
	})
}

type AssignTaskToSelfRequest struct {
	TaskId        uuid.UUID
	RequestUserId uuid.UUID
}

func (ts *TaskService) AssignTaskToSelf(ctx context.Context, request AssignTaskToSelfRequest) (*domain.Task, error) {
	task, project, err := ts.getTaskAndProjectForUser(ctx, request.TaskId, request.RequestUserId)
	if err != nil {
		return nil, err
	}

	if task.IsAssignedTo(request.RequestUserId) {
		return task, nil
	}

	return ts.updateLoadedTask(ctx, task, project, UpdateTaskRequest{
		TaskId:           task.Id,
		Title:            task.Title,
		Description:      task.Description,
		Code:             &task.Code,
		ProjectColumnId:  task.ProjectColumnId,
		RequestUserId:    request.RequestUserId,
		Priority:         task.Priority,
		DueDate:          task.DueDate,
		ResponsibleId:    &request.RequestUserId,
		Tags:             task.Tags,
		DependsOnTaskIds: task.DependsOnTaskIds,
	})
}

func (ts *TaskService) moveLoadedTask(ctx context.Context, oldTask *domain.Task, project *domain.Project, request MoveTaskRequest) (*domain.Task, error) {
	if oldTask.IsArchived() {
		return nil, apperr.BusinessValidationError(domain.ErrTaskArchived.Error())
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	var updatedTask *domain.Task
	err = ts.taskRepository.WithProjectColumnMoveLock(ctx, request.ProjectColumnId, func(ctx context.Context) error {
		newOrder, err := ts.calculateOrder(ctx, request)
		if err != nil {
			return err
		}

		placement := oldTask.MoveTo(projectColumn, newOrder)

		moved, err := ts.taskRepository.MoveTask(ctx, placement, request.RequestUserId, func(moved *domain.Task) []outbox.Message {
			var previousProjectColumnID *uuid.UUID
			if oldTask.ProjectColumnId != moved.ProjectColumnId {
				previousProjectColumnID = &oldTask.ProjectColumnId
			}

			return []outbox.Message{{
				Topic:       events.TaskUpdated,
				AggregateID: moved.Id,
				Payload: &events.TaskUpdatedPayload{
					Task:         *moved,
					PreviousTask: oldTask,
					ActionOrigin: domain.ActionOriginFromContext(ctx),
					User: domain.User{
						Id: request.RequestUserId,
					},
					PreviousProjectColumnID: previousProjectColumnID,
				},
			}}
		})
		if err != nil {
			return err
		}

		updatedTask = moved
		return nil
	})
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, apperr.NotFoundError("task not found")
			}
			return nil, domainErr
		}
		return nil, apperr.ServerError("failed to move task", err)
	}

	return updatedTask, nil
}

func (ts *TaskService) updateLoadedTask(ctx context.Context, task *domain.Task, project *domain.Project, request UpdateTaskRequest) (*domain.Task, error) {
	if task.IsArchived() {
		return nil, apperr.BusinessValidationError(domain.ErrTaskArchived.Error())
	}

	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		if !project.IsMember(*request.ResponsibleId) {
			return nil, apperr.BusinessValidationError("responsible is not a member of the project")
		}
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	responsible := task.Responsible
	if request.ResponsibleId == nil || *request.ResponsibleId == uuid.Nil {
		responsible = nil
	} else if task.ResponsibleId == nil || *task.ResponsibleId != *request.ResponsibleId || task.Responsible == nil {
		responsible, err = ts.userRepository.GetById(ctx, *request.ResponsibleId)
		if err != nil {
			return nil, apperr.ServerError("failed to get responsible user", err)
		}
	}

	oldProjectColumnID := task.ProjectColumnId

	dependsOnRefs, err := ts.validateDependsOnTaskIds(ctx, task.ProjectId, task.Id, request.DependsOnTaskIds)
	if err != nil {
		return nil, err
	}

	updatedTask := task.ApplyEdit(domain.TaskEdit{
		Title:          request.Title,
		Description:    request.Description,
		Code:           request.Code,
		Priority:       request.Priority,
		DueDate:        request.DueDate,
		ResponsibleId:  request.ResponsibleId,
		Responsible:    responsible,
		Tags:           request.Tags,
		DependsOnTasks: dependsOnRefs,
	}, projectColumn)

	var previousProjectColumnID *uuid.UUID
	if oldProjectColumnID != updatedTask.ProjectColumnId {
		previousProjectColumnID = &oldProjectColumnID
	}

	err = ts.taskRepository.Update(ctx, updatedTask, func(updated *domain.Task) []outbox.Message {
		return []outbox.Message{{
			Topic:       events.TaskUpdated,
			AggregateID: updated.Id,
			Payload: &events.TaskUpdatedPayload{
				Task:         *updated,
				PreviousTask: task,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.RequestUserId,
				},
				PreviousProjectColumnID: previousProjectColumnID,
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to update task", err)
	}

	return updatedTask, nil
}

func (ts *TaskService) getTaskAndProjectForUser(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.Task, *domain.Project, error) {
	if userID == uuid.Nil {
		return nil, nil, apperr.UnauthorizedError("unauthorized")
	}

	task, err := ts.taskRepository.GetById(ctx, taskID)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, nil, apperr.NotFoundError("task not found")
			}
			return nil, nil, domainErr
		}
		return nil, nil, apperr.ServerError("failed to get task", err)
	}

	project, err := ts.projectRepository.GetById(ctx, task.ProjectId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, nil, apperr.NotFoundError("project not found")
			}
			return nil, nil, domainErr
		}
		return nil, nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(userID) {
		return nil, nil, apperr.ForbiddenError("forbidden")
	}

	return task, project, nil
}

func (ts *TaskService) calculateOrder(ctx context.Context, request MoveTaskRequest) (string, error) {
	if request.AfterTaskId == nil {
		firstTask, err := ts.taskRepository.GetFirstTaskInColumn(ctx, request.ProjectId, request.ProjectColumnId)
		if err != nil {
			var domainErr apperr.DomainError
			if errors.As(err, &domainErr) && domainErr.Code == apperr.NotFoundErrorCode {
				return fracindex.GenerateKeyBetween("", "")
			}
			return "", err
		}

		if firstTask.Id == request.TaskId {
			nextTask, err := ts.taskRepository.GetProjectTaskAfterId(ctx, request.TaskId, request.ProjectId)
			if err != nil {
				var domainErr apperr.DomainError
				if errors.As(err, &domainErr) && domainErr.Code == apperr.NotFoundErrorCode {
					return fracindex.GenerateKeyBetween("", "")
				}
				return "", err
			}

			return fracindex.GenerateKeyBetween("", nextTask.Order)
		}

		return fracindex.GenerateKeyBetween("", firstTask.Order)
	}

	prevTask, err := ts.taskRepository.GetById(ctx, *request.AfterTaskId)
	if err != nil {
		return "", err
	}

	if prevTask.ProjectId != request.ProjectId || prevTask.ProjectColumnId != request.ProjectColumnId {
		return "", apperr.BusinessValidationError("after_task_id must belong to the target column")
	}

	nextTask, err := ts.taskRepository.GetProjectTaskAfterId(ctx, *request.AfterTaskId, request.ProjectId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return fracindex.GenerateKeyBetween(prevTask.Order, "")
			}
		}
		return "", err
	}

	if nextTask.Id == request.TaskId {
		nextTask, err = ts.taskRepository.GetProjectTaskAfterId(ctx, request.TaskId, request.ProjectId)
		if err != nil {
			var domainErr apperr.DomainError
			if errors.As(err, &domainErr) {
				if domainErr.Code == apperr.NotFoundErrorCode {
					return fracindex.GenerateKeyBetween(prevTask.Order, "")
				}
			}
			return "", err
		}
	}

	if nextTask.Order <= prevTask.Order {
		return fracindex.GenerateKeyBetween(prevTask.Order, "")
	}

	return fracindex.GenerateKeyBetween(prevTask.Order, nextTask.Order)
}

type SearchTasksRequest struct {
	UserId           uuid.UUID
	ProjectId        *uuid.UUID
	ProjectColumnIDs []uuid.UUID
	SearchQuery      string
	IncludeArchived  bool
	IncludeDone      bool
	Limit            int
	CursorDueDate    *time.Time
	CursorUpdatedAt  *time.Time
	CursorTaskId     *uuid.UUID
}

func (ts *TaskService) SearchTasks(ctx context.Context, request SearchTasksRequest) (*utils.CursorPaginated[domain.Task], error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	request.SearchQuery = strings.TrimSpace(request.SearchQuery)
	if request.SearchQuery == "" {
		return nil, apperr.BusinessValidationError("query is required")
	}

	if request.Limit < 1 || request.Limit > 100 {
		return nil, apperr.BusinessValidationError("limit must be between 1 and 100")
	}

	if (request.CursorUpdatedAt == nil) != (request.CursorTaskId == nil) {
		return nil, apperr.BusinessValidationError("cursor_updated_at and cursor_task_id must be provided together")
	}

	if request.ProjectId == nil {
		if len(request.ProjectColumnIDs) > 0 {
			return nil, apperr.BusinessValidationError("project_id is required when project_column_ids are provided")
		}
		if request.IncludeArchived {
			return nil, apperr.BusinessValidationError("project_id is required when archived tasks are included")
		}
	} else {
		if *request.ProjectId == uuid.Nil {
			return nil, apperr.BusinessValidationError("project_id is required")
		}

		project, err := ts.projectRepository.GetById(ctx, *request.ProjectId)
		if err != nil {
			return nil, apperr.ServerError("failed to get project", err)
		}
		if !project.IsMember(request.UserId) {
			return nil, apperr.ForbiddenError("forbidden")
		}
		if err := validateProjectColumnIDs(project, request.ProjectColumnIDs); err != nil {
			return nil, err
		}
		if request.IncludeArchived && !project.IsCreator(request.UserId) {
			return nil, apperr.ForbiddenError("only the project creator can view archived tasks")
		}
	}

	result, err := ts.taskRepository.SearchTasks(ctx, request)
	if err != nil {
		return nil, apperr.ServerError("failed to search tasks", err)
	}

	return result, nil
}

type FindTaskByCodeRequest struct {
	ProjectId uuid.UUID
	UserId    uuid.UUID
	Code      string
}

func (ts *TaskService) FindTaskByCode(ctx context.Context, request FindTaskByCodeRequest) (*domain.Task, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, apperr.BusinessValidationError("project_id is required")
	}

	code := strings.TrimSpace(request.Code)
	if code == "" {
		return nil, apperr.BusinessValidationError("code is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	matches, err := ts.taskRepository.FindTaskRefsByProjectAndCode(ctx, request.ProjectId, code)
	if err != nil {
		return nil, apperr.ServerError("failed to find task by code", err)
	}

	switch len(matches) {
	case 0:
		return nil, apperr.NotFoundError("task not found")
	case 1:
		task, err := ts.taskRepository.GetById(ctx, matches[0].Id)
		if err != nil {
			var domainErr apperr.DomainError
			if errors.As(err, &domainErr) {
				return nil, domainErr
			}
			return nil, apperr.ServerError("failed to get task", err)
		}
		return task, nil
	default:
		return nil, apperr.BusinessValidationError("task code matches multiple tasks in this project")
	}
}

type SearchProjectTasksForDependenciesRequest struct {
	ProjectId     uuid.UUID
	UserId        uuid.UUID
	Query         string
	ExcludeTaskId *uuid.UUID
	Limit         int
}

type SuggestTaskCodesRequest struct {
	ProjectId uuid.UUID
	UserId    uuid.UUID
	Prefix    string
	Limit     int
}

func (ts *TaskService) SuggestTaskCodes(ctx context.Context, request SuggestTaskCodesRequest) ([]domain.TaskCodeSuggestion, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, apperr.BusinessValidationError("project_id is required")
	}

	prefix := strings.TrimSpace(request.Prefix)
	if len(prefix) < 2 {
		return nil, apperr.BusinessValidationError("prefix must be at least 2 characters")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	results, err := ts.taskRepository.SuggestTaskCodesByProjectPrefix(ctx, request.ProjectId, prefix, request.Limit)
	if err != nil {
		return nil, apperr.ServerError("failed to suggest task codes", err)
	}

	return results, nil
}

func (ts *TaskService) SearchProjectTasksForDependencies(ctx context.Context, request SearchProjectTasksForDependenciesRequest) ([]domain.TaskDependencyRef, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, apperr.BusinessValidationError("project_id is required")
	}

	query := strings.TrimSpace(request.Query)
	if query == "" {
		return nil, apperr.BusinessValidationError("query is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, apperr.ForbiddenError("forbidden")
	}

	results, err := ts.taskRepository.SearchProjectTasksForDependencies(ctx, request.ProjectId, query, request.ExcludeTaskId, request.Limit)
	if err != nil {
		return nil, apperr.ServerError("failed to search project tasks for dependencies", err)
	}

	return results, nil
}

func archiveTransitionError(err error) error {
	switch {
	case errors.Is(err, domain.ErrTaskNotArchived),
		errors.Is(err, domain.ErrTaskAlreadyArchived),
		errors.Is(err, domain.ErrProjectColumnNotFound):
		return apperr.BusinessValidationError(err.Error())
	default:
		return apperr.ServerError("failed to change task archive state", err)
	}
}

func findProjectColumn(project *domain.Project, columnID uuid.UUID) (*domain.ProjectColumn, error) {
	column, err := project.Column(columnID)
	if err != nil {
		return nil, apperr.BusinessValidationError(err.Error())
	}

	return column, nil
}

func validateProjectColumnIDs(project *domain.Project, columnIDs []uuid.UUID) error {
	if err := project.ValidateColumnIDs(columnIDs); err != nil {
		return apperr.BusinessValidationError(err.Error())
	}

	return nil
}

func (ts *TaskService) validateDependsOnTaskIds(ctx context.Context, projectId uuid.UUID, taskId uuid.UUID, dependsOnTaskIds []uuid.UUID) ([]domain.TaskDependencyRef, error) {
	uniqueIDs := domain.DedupeTaskIDs(dependsOnTaskIds)
	if len(uniqueIDs) == 0 {
		return []domain.TaskDependencyRef{}, nil
	}

	if err := domain.ValidateDependencyIDs(taskId, uniqueIDs); err != nil {
		return nil, apperr.BusinessValidationError(err.Error())
	}

	refs, err := ts.taskRepository.GetTaskDependencyRefsByProjectAndIds(ctx, projectId, uniqueIDs)
	if err != nil {
		return nil, apperr.ServerError("failed to validate task dependencies", err)
	}

	orderedRefs, err := domain.ResolveDependencyRefs(uniqueIDs, refs)
	if err != nil {
		return nil, apperr.BusinessValidationError(err.Error())
	}

	if taskId == uuid.Nil {
		return orderedRefs, nil
	}

	edges, err := ts.taskRepository.ListTaskDependenciesByProjectId(ctx, projectId)
	if err != nil {
		return nil, apperr.ServerError("failed to validate task dependencies", err)
	}

	if domain.HasDependencyCycle(taskId, uniqueIDs, edges) {
		return nil, apperr.BusinessValidationError(domain.ErrDependencyCycle.Error())
	}

	return orderedRefs, nil
}

func resolveDoneColumn(project *domain.Project) (*domain.ProjectColumn, error) {
	column, err := project.DoneColumn()
	if err != nil {
		return nil, apperr.BusinessValidationError(err.Error())
	}

	return column, nil
}
