package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type taskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	Update(ctx context.Context, task *domain.Task) error
	CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error
	GetFirstTaskInColumn(ctx context.Context, projectId uuid.UUID, projectColumnID uuid.UUID) (*domain.Task, error)
	GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error)
	MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID) (*domain.Task, error)
	CountTasksByProjectIdAndColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID) (map[string]int, error)
	ListUserDueTasks(ctx context.Context, userId uuid.UUID, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	SearchTasksForUser(ctx context.Context, userId uuid.UUID, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
}

type taskServiceProjectRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error)
}

type taskServiceUserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type taskServicePublisher interface {
	Publish(ctx context.Context, topic events.Topic, payload events.Payload) error
}

type TaskService struct {
	taskRepository    taskRepository
	projectRepository taskServiceProjectRepository
	userRepository    taskServiceUserRepository
	publisher         taskServicePublisher
}

func NewTaskService(taskRepository taskRepository, projectRepository taskServiceProjectRepository, userRepository taskServiceUserRepository, publisher taskServicePublisher) *TaskService {
	return &TaskService{
		taskRepository:    taskRepository,
		projectRepository: projectRepository,
		userRepository:    userRepository,
		publisher:         publisher,
	}
}

type CreateTaskRequest struct {
	ProjectId       uuid.UUID
	ProjectColumnId uuid.UUID
	Title           string
	Description     string
	RequestUserId   uuid.UUID
	Priority        string
	DueDate         *time.Time
	ResponsibleId   *uuid.UUID
	Tags            []string
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

	task := domain.Task{
		ProjectId:       request.ProjectId,
		Title:           request.Title,
		Description:     request.Description,
		Status:          domain.TaskStatus(strings.ToLower(projectColumn.Name)),
		AuthorId:        request.RequestUserId,
		ProjectColumnId: request.ProjectColumnId,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Author:          user,
		Priority:        domain.TaskPriority(request.Priority),
		Order:           order,
		ResponsibleId:   request.ResponsibleId,
		Responsible:     responsible,
		DueDate:         request.DueDate,
		Tags:            formattedTags,
		Updates:         []domain.TaskUpdate{},
		ProjectColumn:   projectColumn,
	}

	if projectColumn.IsDoneColumn {
		now := time.Now()
		task.DoneAt = &now
	}

	err = ts.taskRepository.Create(ctx, &task)
	if err != nil {
		return nil, domain.ServerError("failed to create task", err)
	}

	err = ts.publisher.Publish(ctx, events.TaskCreated, &events.TaskCreatedPayload{
		Task: task,
		User: domain.User{
			Id: request.RequestUserId,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task created event", err)
	}

	return &task, nil
}

type UpdateTaskRequest struct {
	TaskId          uuid.UUID
	Title           string
	Description     string
	ProjectColumnId uuid.UUID
	RequestUserId   uuid.UUID
	Priority        domain.TaskPriority
	DueDate         *time.Time
	ResponsibleId   *uuid.UUID
	Tags            []string
}

func (ts *TaskService) Update(ctx context.Context, request UpdateTaskRequest) (*domain.Task, error) {
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

	updatedTask := domain.Task{
		Id:              task.Id,
		ProjectId:       task.ProjectId,
		ProjectColumnId: request.ProjectColumnId,
		AuthorId:        task.AuthorId,
		CreatedAt:       task.CreatedAt,
		Author:          task.Author,
		Order:           task.Order,
		Title:           request.Title,
		Description:     request.Description,
		Status:          domain.TaskStatus(strings.ToLower(projectColumn.Name)),
		Priority:        request.Priority,
		ResponsibleId:   request.ResponsibleId,
		Responsible:     responsible,
		DueDate:         request.DueDate,
		Tags:            formattedTags,
		UpdatedAt:       time.Now(),
		Updates:         []domain.TaskUpdate{},
		ProjectColumn:   projectColumn,
		ArchivedAt:      task.ArchivedAt,
	}

	if projectColumn.IsDoneColumn {
		now := time.Now()
		updatedTask.DoneAt = &now
	}

	err = ts.taskRepository.Update(ctx, &updatedTask)
	if err != nil {
		return nil, domain.ServerError("failed to update task", err)
	}

	// Only include previous status if it changed
	var previousProjectColumnID *uuid.UUID
	if oldProjectColumnID != updatedTask.ProjectColumnId {
		previousProjectColumnID = &oldProjectColumnID
	}

	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task:         updatedTask,
		PreviousTask: task,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousProjectColumnID: previousProjectColumnID,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task updated event", err)
	}

	return &updatedTask, nil
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
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := ts.taskRepository.GetById(ctx, id)
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

	if !project.IsMember(userId) {
		return nil, domain.ForbiddenError("forbidden")
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

	err = ts.taskRepository.Update(ctx, &archivedTask)
	if err != nil {
		return nil, domain.ServerError("failed to archive task", err)
	}

	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task:         archivedTask,
		PreviousTask: task,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousProjectColumnID: nil,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task archived event", err)
	}

	return &archivedTask, nil
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
	if request.RequestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	oldTask, err := ts.taskRepository.GetById(ctx, request.TaskId)
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
	project, err := ts.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	if !project.IsMember(request.RequestUserId) {
		return nil, domain.ForbiddenError("forbidden")
	}

	projectColumn, err := findProjectColumn(project, request.ProjectColumnId)
	if err != nil {
		return nil, err
	}

	newOrder, err := ts.calculateOrder(ctx, request)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("task not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to calculate order", err)
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

	updatedTask, err := ts.taskRepository.MoveTask(ctx, &task, request.RequestUserId)
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

	// Only include previous status if it changed
	var previousProjectColumnID *uuid.UUID
	if oldTask.ProjectColumnId != updatedTask.ProjectColumnId {
		previousProjectColumnID = &oldTask.ProjectColumnId
	}

	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task:         *updatedTask,
		PreviousTask: oldTask,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousProjectColumnID: previousProjectColumnID,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task moved event", err)
	}

	return updatedTask, nil
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
