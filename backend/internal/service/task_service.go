package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type taskRepository interface {
	Create(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	FindTaskRefsByProjectAndCode(ctx context.Context, projectId uuid.UUID, code string) ([]domain.TaskDependencyRef, error)
	SuggestTaskCodesByProjectPrefix(ctx context.Context, projectId uuid.UUID, prefix string, limit int) ([]domain.TaskCodeSuggestion, error)
	ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	Update(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error
	CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error
	GetFirstTaskInColumn(ctx context.Context, projectId uuid.UUID, projectColumnID uuid.UUID) (*domain.Task, error)
	GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error)
	MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID, buildEvents func(*domain.Task) []outbox.Message) (*domain.Task, error)
	WithProjectColumnMoveLock(ctx context.Context, projectColumnID uuid.UUID, fn func(context.Context) error) error
	CountTasksByProjectIdAndColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID) (map[string]int, error)
	ListUserDueTasks(ctx context.Context, userId uuid.UUID, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	SearchTasksForUser(ctx context.Context, userId uuid.UUID, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	SearchProjectTasksForDependencies(ctx context.Context, projectId uuid.UUID, query string, excludeTaskId *uuid.UUID, limit int) ([]domain.TaskDependencyRef, error)
	GetTaskDependencyRefsByProjectAndIds(ctx context.Context, projectId uuid.UUID, taskIds []uuid.UUID) ([]domain.TaskDependencyRef, error)
	ListTaskDependenciesByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.TaskDependencyEdge, error)
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
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("project not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		if !project.IsMember(*request.ResponsibleId) {
			return nil, domain.BusinessValidationError("responsible is not a member of the project")
		}
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		return nil, domain.ServerError("failed to get user", err)
	}

	var responsible *domain.User
	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		responsible, err = ts.userRepository.GetById(ctx, *request.ResponsibleId)
		if err != nil {
			return nil, domain.ServerError("failed to get responsible user", err)
		}
	}

	firstTask, err := ts.taskRepository.GetFirstTaskInColumn(ctx, request.ProjectId, request.ProjectColumnId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code != domain.NotFoundErrorCode {
				return nil, domain.ServerError("failed to get first task in column", err)
			}
		} else {
			return nil, domain.ServerError("failed to get first task in column", err)
		}
	}

	var nextOrder string
	if firstTask != nil {
		nextOrder = firstTask.Order
	}

	order, err := fracindex.GenerateKeyBetween("", nextOrder)
	if err != nil {
		return nil, domain.ServerError("failed to generate task order", err)
	}

	formattedTags := []string{}
	for _, tag := range request.Tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}

		formattedTags = append(formattedTags, tag)
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
		Status:           domain.TaskStatus(strings.ToLower(projectColumn.Name)),
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
		Tags:             formattedTags,
		DependsOnTaskIds: domain.TaskDependencyRefsToIDs(dependsOnRefs),
		DependsOnTasks:   dependsOnRefs,
		Updates:          []domain.TaskUpdate{},
		ProjectColumn:    projectColumn,
	}

	if projectColumn.IsDoneColumn {
		now := time.Now()
		task.DoneAt = &now
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
		return nil, domain.ServerError("failed to create task", err)
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
		return nil, domain.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, domain.BusinessValidationError("project_id is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	if err := validateProjectColumnIDs(project, request.ProjectColumnIDs); err != nil {
		return nil, err
	}

	tasks, err := ts.taskRepository.ListByProjectId(ctx, request.ProjectId, request.ProjectColumnIDs, request.Archived, request.TaskOrder, request.CursorUpdatedAt, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to list tasks", err)
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
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, domain.ForbiddenError("forbidden")
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
			return nil, domain.ServerError("failed to list tasks", err)
		}
		results[statusID.String()] = *result
	}

	return results, nil
}

func (ts *TaskService) CountByColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, requestUserId uuid.UUID) (map[string]int, error) {
	if requestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, projectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(requestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	if err := validateProjectColumnIDs(project, projectColumnIDs); err != nil {
		return nil, err
	}
	results, err := ts.taskRepository.CountTasksByProjectIdAndColumn(ctx, projectId, projectColumnIDs)
	if err != nil {
		return nil, domain.ServerError("failed to count tasks", err)
	}
	return results, nil
}

type ArchiveTaskRequest struct {
	TaskId        uuid.UUID
	RequestUserId uuid.UUID
}

func (ts *TaskService) Archive(ctx context.Context, request ArchiveTaskRequest) (*domain.Task, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := ts.taskRepository.GetById(ctx, request.TaskId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("task not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get task", err)
	}

	project, err := ts.projectRepository.GetById(ctx, task.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("project not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	archivedTask := *task
	archivedTask.Status = domain.TaskStatusArchived
	now := time.Now()
	archivedTask.ArchivedAt = &now
	archivedTask.UpdatedAt = time.Now()
	archivedTask.Updates = []domain.TaskUpdate{}

	err = ts.taskRepository.Update(ctx, &archivedTask, func(updated *domain.Task) []outbox.Message {
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
		return nil, domain.ServerError("failed to archive task", err)
	}

	return &archivedTask, nil
}

type RestoreTaskRequest struct {
	TaskId          uuid.UUID
	ProjectColumnId uuid.UUID
	RequestUserId   uuid.UUID
}

func (ts *TaskService) Restore(ctx context.Context, request RestoreTaskRequest) (*domain.Task, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := ts.taskRepository.GetById(ctx, request.TaskId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("task not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get task", err)
	}

	project, err := ts.projectRepository.GetById(ctx, task.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("project not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	if task.ArchivedAt == nil {
		return nil, domain.BusinessValidationError("task is not archived")
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	restoredTask := *task
	restoredTask.ProjectColumnId = request.ProjectColumnId
	restoredTask.ProjectColumn = projectColumn
	restoredTask.Status = domain.TaskStatus(strings.ToLower(projectColumn.Name))
	restoredTask.ArchivedAt = nil
	restoredTask.UpdatedAt = time.Now()
	restoredTask.Updates = []domain.TaskUpdate{}

	if projectColumn.IsDoneColumn {
		now := time.Now()
		restoredTask.DoneAt = &now
	} else {
		restoredTask.DoneAt = nil
	}

	err = ts.taskRepository.Update(ctx, &restoredTask, func(updated *domain.Task) []outbox.Message {
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
		return nil, domain.ServerError("failed to restore task", err)
	}

	return &restoredTask, nil
}

type ListUserDueTasksRequest struct {
	UserId          uuid.UUID
	Limit           int
	CursorDueDate   *time.Time
	CursorUpdatedAt *time.Time
}

func (ts *TaskService) ListUserDueTasks(ctx context.Context, request ListUserDueTasksRequest) (*utils.CursorPaginated[domain.Task], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	result, err := ts.taskRepository.ListUserDueTasks(ctx, request.UserId, request.CursorDueDate, request.CursorUpdatedAt, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to list user due tasks", err)
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
		return nil, domain.BusinessValidationError("task does not belong to the provided project_id")
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

	if task.ProjectColumnId == doneColumn.Id && task.DoneAt != nil {
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

	if task.ResponsibleId != nil && *task.ResponsibleId == request.RequestUserId {
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

		task := domain.Task{
			Id:              request.TaskId,
			ProjectId:       request.ProjectId,
			Order:           newOrder,
			Status:          domain.TaskStatus(strings.ToLower(projectColumn.Name)),
			ProjectColumnId: request.ProjectColumnId,
			ProjectColumn:   projectColumn,
		}

		if projectColumn.IsDoneColumn {
			now := time.Now()
			task.DoneAt = &now
		}

		moved, err := ts.taskRepository.MoveTask(ctx, &task, request.RequestUserId, func(moved *domain.Task) []outbox.Message {
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
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("task not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to move task", err)
	}

	return updatedTask, nil
}

func (ts *TaskService) updateLoadedTask(ctx context.Context, task *domain.Task, project *domain.Project, request UpdateTaskRequest) (*domain.Task, error) {
	if request.ResponsibleId != nil && *request.ResponsibleId != uuid.Nil {
		if !project.IsMember(*request.ResponsibleId) {
			return nil, domain.BusinessValidationError("responsible is not a member of the project")
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
			return nil, domain.ServerError("failed to get responsible user", err)
		}
	}

	oldProjectColumnID := task.ProjectColumnId

	formattedTags := []string{}
	for _, tag := range request.Tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}
		formattedTags = append(formattedTags, tag)
	}

	dependsOnRefs, err := ts.validateDependsOnTaskIds(ctx, task.ProjectId, task.Id, request.DependsOnTaskIds)
	if err != nil {
		return nil, err
	}

	code := task.Code
	if request.Code != nil {
		code = strings.TrimSpace(*request.Code)
	}

	updatedTask := domain.Task{
		Id:               task.Id,
		ProjectId:        task.ProjectId,
		ProjectColumnId:  request.ProjectColumnId,
		AuthorId:         task.AuthorId,
		CreatedAt:        task.CreatedAt,
		Author:           task.Author,
		Order:            task.Order,
		Title:            request.Title,
		Description:      request.Description,
		Code:             code,
		Status:           domain.TaskStatus(strings.ToLower(projectColumn.Name)),
		Priority:         request.Priority,
		ResponsibleId:    request.ResponsibleId,
		Responsible:      responsible,
		DueDate:          request.DueDate,
		Tags:             formattedTags,
		DependsOnTaskIds: domain.TaskDependencyRefsToIDs(dependsOnRefs),
		DependsOnTasks:   dependsOnRefs,
		UpdatedAt:        time.Now(),
		Updates:          []domain.TaskUpdate{},
		ProjectColumn:    projectColumn,
		ArchivedAt:       task.ArchivedAt,
	}

	if projectColumn.IsDoneColumn {
		now := time.Now()
		updatedTask.DoneAt = &now
	}

	var previousProjectColumnID *uuid.UUID
	if oldProjectColumnID != updatedTask.ProjectColumnId {
		previousProjectColumnID = &oldProjectColumnID
	}

	err = ts.taskRepository.Update(ctx, &updatedTask, func(updated *domain.Task) []outbox.Message {
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
		return nil, domain.ServerError("failed to update task", err)
	}

	return &updatedTask, nil
}

func (ts *TaskService) getTaskAndProjectForUser(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.Task, *domain.Project, error) {
	if userID == uuid.Nil {
		return nil, nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := ts.taskRepository.GetById(ctx, taskID)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, nil, domain.NotFoundError("task not found")
			}
			return nil, nil, domainErr
		}
		return nil, nil, domain.ServerError("failed to get task", err)
	}

	project, err := ts.projectRepository.GetById(ctx, task.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, nil, domain.NotFoundError("project not found")
			}
			return nil, nil, domainErr
		}
		return nil, nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(userID) {
		return nil, nil, domain.ForbiddenError("forbidden")
	}

	return task, project, nil
}

func (ts *TaskService) calculateOrder(ctx context.Context, request MoveTaskRequest) (string, error) {
	if request.AfterTaskId == nil {
		firstTask, err := ts.taskRepository.GetFirstTaskInColumn(ctx, request.ProjectId, request.ProjectColumnId)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) && domainErr.Code == domain.NotFoundErrorCode {
				return fracindex.GenerateKeyBetween("", "")
			}
			return "", err
		}

		if firstTask.Id == request.TaskId {
			nextTask, err := ts.taskRepository.GetProjectTaskAfterId(ctx, request.TaskId, request.ProjectId)
			if err != nil {
				var domainErr domain.DomainError
				if errors.As(err, &domainErr) && domainErr.Code == domain.NotFoundErrorCode {
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
		return "", domain.BusinessValidationError("after_task_id must belong to the target column")
	}

	nextTask, err := ts.taskRepository.GetProjectTaskAfterId(ctx, *request.AfterTaskId, request.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return fracindex.GenerateKeyBetween(prevTask.Order, "")
			}
		}
		return "", err
	}

	if nextTask.Id == request.TaskId {
		nextTask, err = ts.taskRepository.GetProjectTaskAfterId(ctx, request.TaskId, request.ProjectId)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) {
				if domainErr.Code == domain.NotFoundErrorCode {
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

type SearchTasksForUserRequest struct {
	UserId          uuid.UUID
	Limit           int
	CursorDueDate   *time.Time
	CursorUpdatedAt *time.Time
	SearchQuery     string
}

func (ts *TaskService) SearchTasksForUser(ctx context.Context, request SearchTasksForUserRequest) (*utils.CursorPaginated[domain.Task], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	result, err := ts.taskRepository.SearchTasksForUser(ctx, request.UserId, request.SearchQuery, request.CursorDueDate, request.CursorUpdatedAt, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to search tasks for user", err)
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
		return nil, domain.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, domain.BusinessValidationError("project_id is required")
	}

	code := strings.TrimSpace(request.Code)
	if code == "" {
		return nil, domain.BusinessValidationError("code is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	matches, err := ts.taskRepository.FindTaskRefsByProjectAndCode(ctx, request.ProjectId, code)
	if err != nil {
		return nil, domain.ServerError("failed to find task by code", err)
	}

	switch len(matches) {
	case 0:
		return nil, domain.NotFoundError("task not found")
	case 1:
		task, err := ts.taskRepository.GetById(ctx, matches[0].Id)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) {
				return nil, domainErr
			}
			return nil, domain.ServerError("failed to get task", err)
		}
		return task, nil
	default:
		return nil, domain.BusinessValidationError("task code matches multiple tasks in this project")
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
		return nil, domain.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, domain.BusinessValidationError("project_id is required")
	}

	prefix := strings.TrimSpace(request.Prefix)
	if len(prefix) < 2 {
		return nil, domain.BusinessValidationError("prefix must be at least 2 characters")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	results, err := ts.taskRepository.SuggestTaskCodesByProjectPrefix(ctx, request.ProjectId, prefix, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to suggest task codes", err)
	}

	return results, nil
}

func (ts *TaskService) SearchProjectTasksForDependencies(ctx context.Context, request SearchProjectTasksForDependenciesRequest) ([]domain.TaskDependencyRef, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	if request.ProjectId == uuid.Nil {
		return nil, domain.BusinessValidationError("project_id is required")
	}

	query := strings.TrimSpace(request.Query)
	if query == "" {
		return nil, domain.BusinessValidationError("query is required")
	}

	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.UserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	results, err := ts.taskRepository.SearchProjectTasksForDependencies(ctx, request.ProjectId, query, request.ExcludeTaskId, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to search project tasks for dependencies", err)
	}

	return results, nil
}

func findProjectColumn(project *domain.Project, statusID uuid.UUID) (*domain.ProjectColumn, error) {
	for _, status := range project.Columns {
		if status.Id == statusID {
			statusCopy := status
			return &statusCopy, nil
		}
	}

	return nil, domain.BusinessValidationError("invalid project_column_id")
}

func validateProjectColumnIDs(project *domain.Project, statusIDs []uuid.UUID) error {
	if len(statusIDs) == 0 {
		return nil
	}

	valid := map[uuid.UUID]struct{}{}
	for _, column := range project.Columns {
		valid[column.Id] = struct{}{}
	}

	for _, statusID := range statusIDs {
		if _, ok := valid[statusID]; !ok {
			return domain.BusinessValidationError("invalid project_column_id")
		}
	}

	return nil
}

func (ts *TaskService) validateDependsOnTaskIds(ctx context.Context, projectId uuid.UUID, taskId uuid.UUID, dependsOnTaskIds []uuid.UUID) ([]domain.TaskDependencyRef, error) {
	uniqueIDs := dedupeDependsOnTaskIds(dependsOnTaskIds)
	if len(uniqueIDs) == 0 {
		return []domain.TaskDependencyRef{}, nil
	}

	for _, dependsOnTaskID := range uniqueIDs {
		if dependsOnTaskID == uuid.Nil {
			return nil, domain.BusinessValidationError("depends_on_task_ids contains an invalid task id")
		}
		if taskId != uuid.Nil && dependsOnTaskID == taskId {
			return nil, domain.BusinessValidationError("task cannot depend on itself")
		}
	}

	refs, err := ts.taskRepository.GetTaskDependencyRefsByProjectAndIds(ctx, projectId, uniqueIDs)
	if err != nil {
		return nil, domain.ServerError("failed to validate task dependencies", err)
	}

	refByID := make(map[uuid.UUID]domain.TaskDependencyRef, len(refs))
	for _, ref := range refs {
		refByID[ref.Id] = ref
	}

	orderedRefs := make([]domain.TaskDependencyRef, 0, len(uniqueIDs))
	for _, id := range uniqueIDs {
		ref, ok := refByID[id]
		if !ok {
			return nil, domain.BusinessValidationError("depends_on_task_ids contains unknown tasks")
		}
		orderedRefs = append(orderedRefs, ref)
	}

	if taskId == uuid.Nil {
		return orderedRefs, nil
	}

	edges, err := ts.taskRepository.ListTaskDependenciesByProjectId(ctx, projectId)
	if err != nil {
		return nil, domain.ServerError("failed to validate task dependencies", err)
	}

	if hasDependencyCycle(taskId, uniqueIDs, edges) {
		return nil, domain.BusinessValidationError("depends_on_task_ids would create a dependency cycle")
	}

	return orderedRefs, nil
}

func dedupeDependsOnTaskIds(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return []uuid.UUID{}
	}

	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}

func hasDependencyCycle(taskId uuid.UUID, dependsOnTaskIds []uuid.UUID, edges []domain.TaskDependencyEdge) bool {
	adjacency := make(map[uuid.UUID][]uuid.UUID)
	for _, edge := range edges {
		if edge.TaskId == taskId {
			continue
		}
		adjacency[edge.TaskId] = append(adjacency[edge.TaskId], edge.DependsOnTaskId)
	}
	adjacency[taskId] = dependsOnTaskIds

	for _, dependsOnTaskID := range dependsOnTaskIds {
		if dependencyPathReachesTask(dependsOnTaskID, taskId, adjacency, map[uuid.UUID]bool{}) {
			return true
		}
	}

	return false
}

func dependencyPathReachesTask(current, target uuid.UUID, adjacency map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]bool) bool {
	if current == target {
		return true
	}
	if visited[current] {
		return false
	}

	visited[current] = true
	for _, next := range adjacency[current] {
		if dependencyPathReachesTask(next, target, adjacency, visited) {
			return true
		}
	}

	return false
}

func resolveDoneColumn(project *domain.Project) (*domain.ProjectColumn, error) {
	doneColumns := make([]domain.ProjectColumn, 0, 1)
	for _, column := range project.Columns {
		if column.IsDoneColumn {
			doneColumns = append(doneColumns, column)
		}
	}

	if len(doneColumns) != 1 {
		return nil, domain.BusinessValidationError("project must have exactly one done column")
	}

	column := doneColumns[0]
	return &column, nil
}
