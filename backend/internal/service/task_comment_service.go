package service

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type taskCommentRepository interface {
	Create(ctx context.Context, comment *domain.TaskComment, parentCommentID *uuid.UUID) error
	ListByTaskID(ctx context.Context, taskID uuid.UUID, before time.Time, beforeID uuid.UUID, limit int) (*utils.CursorPaginated[domain.TaskComment], error)
}

type TaskCommentService struct {
	taskCommentRepository taskCommentRepository
	taskRepository        taskRepository
	projectRepository     taskServiceProjectRepository
	userRepository        taskServiceUserRepository
	publisher             taskServicePublisher
}

func NewTaskCommentService(taskCommentRepository taskCommentRepository, taskRepository taskRepository, projectRepository taskServiceProjectRepository, userRepository taskServiceUserRepository, publisher taskServicePublisher) *TaskCommentService {
	return &TaskCommentService{
		taskCommentRepository: taskCommentRepository,
		taskRepository:        taskRepository,
		projectRepository:     projectRepository,
		userRepository:        userRepository,
		publisher:             publisher,
	}
}

type CreateTaskCommentRequest struct {
	TaskID          uuid.UUID
	RequestUserID   uuid.UUID
	Content         string
	ParentCommentID *uuid.UUID
}

func (s *TaskCommentService) Create(ctx context.Context, request CreateTaskCommentRequest) (*domain.TaskComment, error) {
	if request.RequestUserID == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := s.taskRepository.GetById(ctx, request.TaskID)
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

	project, err := s.projectRepository.GetById(ctx, task.ProjectId)
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

	if !project.IsMember(request.RequestUserID) {
		return nil, domain.ForbiddenError("forbidden")
	}

	user, err := s.userRepository.GetById(ctx, request.RequestUserID)
	if err != nil {
		return nil, domain.ServerError("failed to get user", err)
	}

	now := time.Now()
	comment := domain.TaskComment{
		Task:      task,
		User:      user,
		Content:   request.Content,
		CreatedAt: now,
		UpdatedAt: now,
		Replies:   []domain.TaskComment{},
	}

	err = s.taskCommentRepository.Create(ctx, &comment, request.ParentCommentID)
	if err != nil {
		return nil, domain.ServerError("failed to create task comment", err)
	}

	err = s.publisher.Publish(ctx, events.TaskCommentCreated, &events.TaskCommentCreatedPayload{
		TaskComment: comment,
		User: domain.User{
			Id: request.RequestUserID,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish task comment created event", err)
	}

	return &comment, nil
}

type ListTaskCommentsRequest struct {
	TaskID          uuid.UUID
	RequestUserID   uuid.UUID
	Limit           int
	Before          time.Time
	BeforeCommentID uuid.UUID
}

func (s *TaskCommentService) ListByTaskID(ctx context.Context, request ListTaskCommentsRequest) (*utils.CursorPaginated[domain.TaskComment], error) {
	if request.RequestUserID == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	task, err := s.taskRepository.GetById(ctx, request.TaskID)
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

	project, err := s.projectRepository.GetById(ctx, task.ProjectId)
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

	if !project.IsMember(request.RequestUserID) {
		return nil, domain.ForbiddenError("forbidden")
	}

	comments, err := s.taskCommentRepository.ListByTaskID(ctx, request.TaskID, request.Before, request.BeforeCommentID, request.Limit)
	if err != nil {
		return nil, domain.ServerError("failed to list task comments", err)
	}

	return comments, nil
}
