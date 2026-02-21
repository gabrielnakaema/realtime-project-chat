package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type taskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	ListByProjectId(ctx context.Context, projectId uuid.UUID, statuses []string, taskOrder int, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error)
	Update(ctx context.Context, task *domain.Task) error
	CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error
	GetSmallestOrderProjectTask(ctx context.Context, projectId uuid.UUID) (*domain.Task, error)
	GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error)
	MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID) (*domain.Task, error)
	NormalizeProjectTaskOrders(ctx context.Context, projectId uuid.UUID) error
	CountTasksByProjectIdAndStatus(ctx context.Context, projectId uuid.UUID, statuses []string) (map[string]int, error)
}

type taskServiceProjectRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
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
	ProjectId     uuid.UUID
	Title         string
	Description   string
	RequestUserId uuid.UUID
	Priority      string
	DueDate       *time.Time
	ResponsibleId *uuid.UUID
	Tags          []string
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

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		return nil, domain.ServerError("failed to get user", err)
	}

	orderTask, err := ts.taskRepository.GetSmallestOrderProjectTask(ctx, request.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code != domain.NotFoundErrorCode {
				return nil, domain.ServerError("failed to get smallest project task order", err)
			}
		} else {
			return nil, domain.ServerError("failed to get smallest project task order", err)
		}
	}

	order := 1000

	if orderTask == nil {
		order = 1000
	} else {
		order = order / 2
		if order < 50 {
			err = ts.taskRepository.NormalizeProjectTaskOrders(ctx, request.ProjectId)
			if err != nil {
				return nil, domain.ServerError("failed to normalize project task orders", err)
			}
		}
	}

	formattedTags := []string{}
	for _, tag := range request.Tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}

		formattedTags = append(formattedTags, tag)
	}

	task := domain.Task{
		ProjectId:     request.ProjectId,
		Title:         request.Title,
		Description:   request.Description,
		AuthorId:      request.RequestUserId,
		Status:        domain.TaskStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Author:        user,
		Priority:      domain.TaskPriority(request.Priority),
		Order:         int(order),
		ResponsibleId: request.ResponsibleId,
		DueDate:       request.DueDate,
		Tags:          formattedTags,
		Updates:       []domain.TaskUpdate{},
	}

	task.Updates = append(task.Updates, domain.NewTaskCreatedUpdate(&task, user))

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
	TaskId        uuid.UUID
	Title         string
	Description   string
	Status        domain.TaskStatus
	RequestUserId uuid.UUID
	Priority      domain.TaskPriority
	DueDate       *time.Time
	ResponsibleId *uuid.UUID
	Tags          []string
	DoneAt        *time.Time
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

	// Capture old status for event payload
	oldStatus := task.Status

	formattedTags := []string{}
	for _, tag := range request.Tags {
		if strings.TrimSpace(tag) == "" {
			continue
		}
		formattedTags = append(formattedTags, tag)
	}

	updatedTask := domain.Task{
		Id:            task.Id,
		ProjectId:     task.ProjectId,
		Status:        task.Status,
		AuthorId:      task.AuthorId,
		CreatedAt:     task.CreatedAt,
		Author:        task.Author,
		Order:         task.Order,
		Title:         request.Title,
		Description:   request.Description,
		Priority:      request.Priority,
		ResponsibleId: request.ResponsibleId,
		DueDate:       request.DueDate,
		DoneAt:        request.DoneAt,
		Tags:          formattedTags,
		UpdatedAt:     time.Now(),
		Updates:       []domain.TaskUpdate{},
	}

	err = updatedTask.ChangeStatus(request.Status)
	if err != nil {
		return nil, err
	}

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return nil, domain.NotFoundError("user not found")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to get user", err)
	}

	err = ts.taskRepository.Update(ctx, &updatedTask)
	if err != nil {
		return nil, domain.ServerError("failed to update task", err)
	}

	newTaskUpdate := domain.NewTaskUpdate(task, &updatedTask, user)
	updates := []domain.TaskUpdate{newTaskUpdate}

	if len(newTaskUpdate.Changes) > 0 {
		err = ts.taskRepository.CreateUpdates(ctx, &updatedTask, updates)
		if err != nil {
			return nil, domain.ServerError("failed to create task updates", err)
		}

		updatedTask.Updates = append(task.Updates, updates...)
	}

	// Only include previous status if it changed
	var previousStatus *domain.TaskStatus
	if oldStatus != updatedTask.Status {
		previousStatus = &oldStatus
	}

	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task: updatedTask,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousStatus: previousStatus,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task updated event", err)
	}

	return &updatedTask, nil
}

type ListTasksRequest struct {
	ProjectId       uuid.UUID
	RequestUserId   uuid.UUID
	Statuses        []string
	TaskOrder       int
	Limit           int
	CursorUpdatedAt *time.Time
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

	for _, s := range request.Statuses {
		if s == string(domain.TaskStatusArchived) {
			if !project.IsCreator(request.RequestUserId) {
				return nil, domain.ForbiddenError("only project owners can view archived tasks")
			}
			break
		}
	}

	tasks, err := ts.taskRepository.ListByProjectId(ctx, request.ProjectId, request.Statuses, request.TaskOrder, request.CursorUpdatedAt, request.Limit)
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

type GroupByStatusRequest struct {
	ProjectId       uuid.UUID
	UserId          uuid.UUID
	Statuses        []domain.TaskStatus
	TaskOrder       int
	CursorUpdatedAt *time.Time
	Limit           int
}

func (ts *TaskService) GroupByStatus(ctx context.Context, request GroupByStatusRequest) (map[domain.TaskStatus]utils.CursorPaginated[domain.Task], error) {
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

	stringStatuses := []string{}
	for _, status := range request.Statuses {
		if status == domain.TaskStatusArchived {
			if !project.IsCreator(request.UserId) {
				return nil, domain.ForbiddenError("only project owners can view archived tasks")
			}
		}
		stringStatuses = append(stringStatuses, string(status))
	}

	results := map[domain.TaskStatus]utils.CursorPaginated[domain.Task]{}

	for _, status := range stringStatuses {
		result, err := ts.taskRepository.ListByProjectId(ctx, request.ProjectId, []string{status}, request.TaskOrder, request.CursorUpdatedAt, request.Limit)
		if err != nil {
			return nil, domain.ServerError("failed to list tasks", err)
		}
		results[domain.TaskStatus(status)] = *result
	}

	return results, nil
}

func (ts *TaskService) CountByStatus(ctx context.Context, projectId uuid.UUID, statuses []domain.TaskStatus, requestUserId uuid.UUID) (map[domain.TaskStatus]int, error) {
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

	stringStatuses := []string{}
	for _, status := range statuses {
		if !slices.Contains(domain.AllowedTaskStatuses, status) {
			return nil, domain.BusinessValidationError("invalid status")
		}
		stringStatuses = append(stringStatuses, string(status))
	}
	results, err := ts.taskRepository.CountTasksByProjectIdAndStatus(ctx, projectId, stringStatuses)
	if err != nil {
		return nil, domain.ServerError("failed to count tasks", err)
	}

	result := map[domain.TaskStatus]int{}
	for status, count := range results {
		result[domain.TaskStatus(status)] = count
	}
	return result, nil
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

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		return nil, domain.ServerError("failed to get user", err)
	}

	oldStatus := task.Status

	archivedTask := *task
	archivedTask.Status = domain.TaskStatusArchived
	archivedTask.UpdatedAt = time.Now()
	archivedTask.Updates = []domain.TaskUpdate{}

	err = ts.taskRepository.Update(ctx, &archivedTask)
	if err != nil {
		return nil, domain.ServerError("failed to archive task", err)
	}

	newTaskUpdate := domain.NewTaskUpdate(task, &archivedTask, user)
	if len(newTaskUpdate.Changes) > 0 {
		err = ts.taskRepository.CreateUpdates(ctx, &archivedTask, []domain.TaskUpdate{newTaskUpdate})
		if err != nil {
			return nil, domain.ServerError("failed to create task updates", err)
		}
		archivedTask.Updates = append(task.Updates, newTaskUpdate)
	}

	previousStatus := &oldStatus
	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task: archivedTask,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousStatus: previousStatus,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task archived event", err)
	}

	return &archivedTask, nil
}

type MoveTaskRequest struct {
	TaskId        uuid.UUID
	RequestUserId uuid.UUID
	AfterTaskId   *uuid.UUID
	ProjectId     uuid.UUID
	Status        domain.TaskStatus
}

func (ts *TaskService) Move(ctx context.Context, request MoveTaskRequest) (*domain.Task, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	// Get the task before moving to capture old status
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
	oldStatus := oldTask.Status

	newOrder, prevOrder, err := ts.calculateOrder(ctx, request)
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

	needsNormalization := false
	if request.AfterTaskId == nil {
		if newOrder < 1 {
			needsNormalization = true
		}
	} else {
		if newOrder <= prevOrder {
			needsNormalization = true
		}
	}

	if needsNormalization {
		err = ts.taskRepository.NormalizeProjectTaskOrders(ctx, request.ProjectId)
		if err != nil {
			return nil, domain.ServerError("failed to normalize project task orders", err)
		}

		newOrder, _, err = ts.calculateOrder(ctx, request)
		if err != nil {
			return nil, domain.ServerError("failed to recalculate order", err)
		}
	}

	task := domain.Task{
		Id:        request.TaskId,
		ProjectId: request.ProjectId,
		Order:     newOrder,
		Status:    request.Status,
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
	var previousStatus *domain.TaskStatus
	if oldStatus != updatedTask.Status {
		previousStatus = &oldStatus
	}

	err = ts.publisher.Publish(ctx, events.TaskUpdated, &events.TaskUpdatedPayload{
		Task: *updatedTask,
		User: domain.User{
			Id: request.RequestUserId,
		},
		PreviousStatus: previousStatus,
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task moved event", err)
	}

	return updatedTask, nil
}

// Returns the new, previous order and an error if any
func (ts *TaskService) calculateOrder(ctx context.Context, request MoveTaskRequest) (int, int, error) {
	if request.AfterTaskId == nil {
		smallestTask, err := ts.taskRepository.GetSmallestOrderProjectTask(ctx, request.ProjectId)
		if err != nil {
			var domainErr domain.DomainError
			if errors.As(err, &domainErr) && domainErr.Code == domain.NotFoundErrorCode {
				return 1000, 0, nil
			}
			return 0, 0, err
		}
		return int(smallestTask.Order / 2), 0, nil
	}

	prevTask, err := ts.taskRepository.GetById(ctx, *request.AfterTaskId)
	if err != nil {
		return 0, 0, err
	}

	nextTask, err := ts.taskRepository.GetProjectTaskAfterId(ctx, *request.AfterTaskId, request.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) && domainErr.Code == domain.NotFoundErrorCode {
			return prevTask.Order + 1000, prevTask.Order, nil
		}
		return 0, 0, err
	}

	return int((prevTask.Order + nextTask.Order) / 2), prevTask.Order, nil
}
