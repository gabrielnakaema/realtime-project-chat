package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/google/uuid"
)

type taskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	ListByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error

	CreateChanges(ctx context.Context, task *domain.Task, changes []domain.TaskChange) error
}

type taskServiceProjectRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
}

type taskServiceUserRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type taskServicePublisher interface {
	Publish(ctx context.Context, topic events.Topic, payload interface{}) error
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

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == request.RequestUserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	user, err := ts.userRepository.GetById(ctx, request.RequestUserId)
	if err != nil {
		return nil, domain.ServerError("failed to get user", err)
	}

	task := domain.Task{
		ProjectId:   request.ProjectId,
		Title:       request.Title,
		Description: request.Description,
		AuthorId:    request.RequestUserId,
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Author:      user,
		Changes:     []domain.TaskChange{},
	}

	err = ts.taskRepository.Create(ctx, &task)
	if err != nil {
		return nil, domain.ServerError("failed to create task", err)
	}

	taskChange := domain.TaskChange{
		TaskId:            task.Id,
		AuthorId:          request.RequestUserId,
		CreatedAt:         time.Now(),
		ChangeDescription: fmt.Sprintf("Task created by %s", user.Name),
	}

	err = ts.taskRepository.CreateChanges(ctx, &task, []domain.TaskChange{taskChange})
	if err != nil {
		return nil, domain.ServerError("failed to create task changes", err)
	}

	task.Changes = append(task.Changes, taskChange)

	err = ts.publisher.Publish(ctx, events.TaskCreated, task)
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

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == request.RequestUserId {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	updatedTask := domain.Task{
		Id:          task.Id,
		ProjectId:   task.ProjectId,
		Title:       request.Title,
		Description: request.Description,
		Status:      task.Status,
		UpdatedAt:   time.Now(),
		Changes:     task.Changes,
		AuthorId:    task.AuthorId,
		CreatedAt:   task.CreatedAt,
		Author:      task.Author,
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

	newTaskChanges := domain.NewTaskChanges(task, &updatedTask, user)

	err = ts.taskRepository.CreateChanges(ctx, &updatedTask, newTaskChanges)
	if err != nil {
		return nil, domain.ServerError("failed to create task changes", err)
	}

	updatedTask.Changes = append(task.Changes, newTaskChanges...)

	err = ts.publisher.Publish(ctx, events.TaskUpdated, updatedTask)
	if err != nil {
		return nil, domain.ServerError("failed to publish task updated event", err)
	}

	return &updatedTask, nil
}

func (ts *TaskService) List(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) ([]domain.Task, error) {
	if projectId == uuid.Nil {
		return nil, domain.BusinessValidationError("project_id is required")
	}

	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, projectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == userId {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	tasks, err := ts.taskRepository.ListByProjectId(ctx, projectId)
	if err != nil {
		return nil, domain.ServerError("failed to list tasks", err)
	}

	return tasks, nil
}

func (ts *TaskService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Task, error) {
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ts.projectRepository.GetById(ctx, id)
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

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == userId {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
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

	return task, nil
}
