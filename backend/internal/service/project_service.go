package service

import (
	"context"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

type projectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string, searchQuery string) ([]domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	CreateMember(ctx context.Context, member *domain.ProjectMember) error
	RemoveMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error
	GetMemberByUserIdAndProjectId(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*domain.ProjectMember, error)
	ListMembersByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.ProjectMember, error)
}

type projectServiceActivityRepository interface {
	List(ctx context.Context, params repository.ListProjectActivityLogsParams) (*utils.CursorPaginated[domain.ProjectActivity], error)
}

type projectServiceUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type projectServicePublisher interface {
	Publish(ctx context.Context, topic events.Topic, payload events.Payload) error
}

type ProjectService struct {
	activityRepository projectServiceActivityRepository
	projectRepository  projectRepository
	userRepository     projectServiceUserRepository
	publisher          projectServicePublisher
}

func NewProjectService(projectRepository projectRepository, userRepository projectServiceUserRepository, publisher projectServicePublisher, activityRepository projectServiceActivityRepository) *ProjectService {
	return &ProjectService{
		projectRepository:  projectRepository,
		userRepository:     userRepository,
		publisher:          publisher,
		activityRepository: activityRepository,
	}
}

type CreateProjectRequest struct {
	Name        string
	Description string
	UserId      uuid.UUID
}

func (ps *ProjectService) Create(ctx context.Context, request CreateProjectRequest) (*domain.Project, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project := domain.Project{
		Name:        request.Name,
		Description: request.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: request.UserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
		UserId: request.UserId,
	}

	err := ps.projectRepository.Create(ctx, &project)
	if err != nil {
		return nil, domain.ServerError("failed to create project", err)
	}

	err = ps.publisher.Publish(ctx, events.ProjectCreated, &events.ProjectCreatedPayload{
		Project: project,
		User: domain.User{
			Id: request.UserId,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish project created event", err)
	}

	return &project, nil
}

func (ps *ProjectService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error) {
	if userId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, id)
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

	return project, nil
}

type ListProjectsByUserIdRequest struct {
	UserId             uuid.UUID
	MemberRole         domain.ProjectMemberRole
	ShouldFilterByRole bool
	SearchQuery        string
}

func (ps *ProjectService) ListByUserId(ctx context.Context, request ListProjectsByUserIdRequest) ([]domain.Project, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	strRole := ""
	if request.ShouldFilterByRole {
		strRole = string(request.MemberRole)
	}

	projects, err := ps.projectRepository.ListByUserId(ctx, request.UserId, strRole, request.SearchQuery)
	if err != nil {
		return nil, domain.ServerError("failed to list projects", err)
	}

	return projects, nil
}

type UpdateProjectRequest struct {
	Id          uuid.UUID
	Name        string
	Description string
	UserId      uuid.UUID
}

func (ps *ProjectService) Update(ctx context.Context, request UpdateProjectRequest) (*domain.Project, error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.Id)
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

	if project.UserId != request.UserId {
		return nil, domain.ForbiddenError("forbidden")
	}

	project.Name = request.Name
	project.Description = request.Description
	project.UpdatedAt = time.Now()

	err = ps.projectRepository.Update(ctx, project)
	if err != nil {
		return nil, domain.ServerError("failed to update project", err)
	}

	err = ps.publisher.Publish(ctx, events.ProjectUpdated, &events.ProjectUpdatedPayload{
		Project: *project,
		User: domain.User{
			Id: request.UserId,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish project updated event", err)
	}

	return project, nil
}

type CreateMemberRequest struct {
	ProjectId     uuid.UUID
	Email         string
	RequestUserId uuid.UUID
}

func (ps *ProjectService) CreateMember(ctx context.Context, request CreateMemberRequest) (*domain.ProjectMember, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	user, err := ps.userRepository.GetByEmail(ctx, request.Email)
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

	if user.Id == request.RequestUserId {
		return nil, domain.BusinessValidationError("you cannot add yourself as a member")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
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

	if project.UserId != request.RequestUserId {
		return nil, domain.ForbiddenError("forbidden")
	}

	alreadyMember := false
	for _, member := range project.Members {
		if member.UserId == user.Id {
			alreadyMember = true
			break
		}
	}
	if alreadyMember {
		return nil, domain.DuplicateEntryError("member already exists")
	}

	member := domain.ProjectMember{
		ProjectId: request.ProjectId,
		UserId:    user.Id,
		Role:      domain.ProjectMemberRoleMember,
	}

	err = ps.projectRepository.CreateMember(ctx, &member)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.DuplicateEntryErrorCode {
				return nil, domain.DuplicateEntryError("member already exists")
			}
			return nil, domainErr
		}
		return nil, domain.ServerError("failed to create member", err)
	}

	err = ps.publisher.Publish(ctx, events.ProjectMemberCreated, &events.ProjectMemberCreatedPayload{
		ProjectMember: member,
		User: domain.User{
			Id: request.RequestUserId,
		},
	})
	if err != nil {
		return nil, domain.ServerError("failed to publish project member created event", err)
	}

	return &member, nil
}

type ListActivitiesByProjectRequest struct {
	ProjectId       uuid.UUID
	UserId          uuid.UUID
	BeforeCreatedAt time.Time
	BeforeId        uuid.UUID
	Limit           int32
}

func (ps *ProjectService) ListActivitiesByProject(ctx context.Context, request ListActivitiesByProjectRequest) (*utils.CursorPaginated[domain.ProjectActivity], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, domain.ServerError("failed to get project", err)
	}

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	params := repository.ListProjectActivityLogsParams{
		ProjectId:       request.ProjectId,
		BeforeCreatedAt: request.BeforeCreatedAt,
		BeforeId:        request.BeforeId,
		Limit:           request.Limit,
		UserId:          uuid.Nil,
	}

	activities, err := ps.activityRepository.List(ctx, params)
	if err != nil {
		return nil, domain.ServerError("failed to list activities", err)
	}

	return activities, nil
}

type ListUsersProjectActivitiesRequest struct {
	UserId          uuid.UUID
	BeforeCreatedAt time.Time
	BeforeId        uuid.UUID
	Limit           int32
}

func (ps *ProjectService) ListUsersProjectActivities(ctx context.Context, request ListUsersProjectActivitiesRequest) (*utils.CursorPaginated[domain.ProjectActivity], error) {
	if request.UserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	params := repository.ListProjectActivityLogsParams{
		UserId:          request.UserId,
		BeforeCreatedAt: request.BeforeCreatedAt,
		BeforeId:        request.BeforeId,
		Limit:           request.Limit,
		ProjectId:       uuid.Nil,
	}

	activities, err := ps.activityRepository.List(ctx, params)
	if err != nil {
		return nil, domain.ServerError("failed to list activities", err)
	}

	return activities, nil
}

func (ps *ProjectService) ListMembersByProjectId(ctx context.Context, projectId uuid.UUID, requestUserId uuid.UUID) ([]domain.ProjectMember, error) {
	if requestUserId == uuid.Nil {
		return nil, domain.UnauthorizedError("unauthorized")
	}

	members, err := ps.projectRepository.ListMembersByProjectId(ctx, projectId)
	if err != nil {
		return nil, domain.ServerError("failed to list members", err)
	}

	hasPermission := false
	for _, member := range members {
		if member.UserId == requestUserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, domain.ForbiddenError("forbidden")
	}

	return members, nil
}

type RemoveMemberRequest struct {
	ProjectId     uuid.UUID
	UserId        uuid.UUID
	RequestUserId uuid.UUID
}

func (ps *ProjectService) RemoveMember(ctx context.Context, request RemoveMemberRequest) error {
	if request.RequestUserId == uuid.Nil {
		return domain.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		var domainErr domain.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == domain.NotFoundErrorCode {
				return domain.NotFoundError("project not found")
			}
			return domainErr
		}

		return domain.ServerError("failed to get project", err)
	}

	member := domain.ProjectMember{}
	var requestUserRole domain.ProjectMemberRole

	for _, m := range project.Members {
		if m.UserId == request.UserId {
			if m.Role == domain.ProjectMemberRoleCreator {
				return domain.BusinessValidationError("cannot remove creator from project")
			}

			member = m
		}

		if m.UserId == request.RequestUserId {
			requestUserRole = m.Role
		}
	}

	isCreator := requestUserRole == domain.ProjectMemberRoleCreator
	isSelfRemoval := request.UserId == request.RequestUserId && requestUserRole == domain.ProjectMemberRoleMember

	if !isCreator && !isSelfRemoval {
		return domain.ForbiddenError("forbidden")
	}

	err = ps.projectRepository.RemoveMember(ctx, request.ProjectId, request.UserId)
	if err != nil {
		return domain.ServerError("failed to remove member", err)
	}

	err = ps.publisher.Publish(ctx, events.ProjectMemberRemoved, &events.ProjectMemberRemovedPayload{
		ProjectMember: member,
		User: domain.User{
			Id: request.RequestUserId,
		},
	})
	if err != nil {
		return domain.ServerError("failed to publish project member removed event", err)
	}

	return nil
}
