package project

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
)

var defaultProjectColumnColors = []string{
	"#64748B",
	"#2563EB",
	"#059669",
	"#D97706",
	"#DC2626",
	"#0891B2",
}

var projectColumnColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type projectRepository interface {
	Create(ctx context.Context, project *domain.Project, buildEvents func() []outbox.Message) error
	GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string, searchQuery string) ([]domain.Project, error)
	UpdateWithColumns(ctx context.Context, params UpdateProjectWithColumnsParams, buildEvents func() []outbox.Message) error
	MarkUpdatedAt(ctx context.Context, projectId uuid.UUID, msgs ...outbox.Message) error
	UpdateColumn(ctx context.Context, status *domain.ProjectColumn) error
	GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error)
	CreateMember(ctx context.Context, member *domain.ProjectMember, buildEvents func() []outbox.Message) error
	RemoveMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, msgs ...outbox.Message) error
	GetMemberByUserIdAndProjectId(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*domain.ProjectMember, error)
	ListMembersByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.ProjectMember, error)
	Delete(ctx context.Context, projectId uuid.UUID, msgs ...outbox.Message) error
}

type projectServiceActivityRepository interface {
	List(ctx context.Context, params ListProjectActivityLogsParams) (*utils.CursorPaginated[domain.ProjectActivity], error)
}

type projectServiceUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type ProjectService struct {
	activityRepository projectServiceActivityRepository
	projectRepository  projectRepository
	userRepository     projectServiceUserRepository
}

func NewProjectService(projectRepository projectRepository, userRepository projectServiceUserRepository, activityRepository projectServiceActivityRepository) *ProjectService {
	return &ProjectService{
		projectRepository:  projectRepository,
		userRepository:     userRepository,
		activityRepository: activityRepository,
	}
}

type CreateProjectRequest struct {
	Name             string
	Description      string
	RepositoryURL    string
	RepositoryOwner  string
	RepositoryName   string
	DefaultBranch    string
	BranchNamePrefix string
	UserId           uuid.UUID
	Columns          []domain.ProjectColumn
}

func (ps *ProjectService) Create(ctx context.Context, request CreateProjectRequest) (*domain.Project, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project := domain.Project{
		Name:             request.Name,
		Description:      request.Description,
		RepositoryURL:    request.RepositoryURL,
		RepositoryOwner:  request.RepositoryOwner,
		RepositoryName:   request.RepositoryName,
		DefaultBranch:    request.DefaultBranch,
		BranchNamePrefix: request.BranchNamePrefix,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: request.UserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
		UserId:  request.UserId,
		Columns: defaultProjectColumns(request.Columns),
	}

	err := ps.projectRepository.Create(ctx, &project, func() []outbox.Message {
		return []outbox.Message{{
			Topic:       events.ProjectCreated,
			AggregateID: project.Id,
			Payload: &events.ProjectCreatedPayload{
				Project:      project,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.UserId,
				},
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to create project", err)
	}

	return &project, nil
}

func (ps *ProjectService) GetById(ctx context.Context, id uuid.UUID, userId uuid.UUID) (*domain.Project, error) {
	if userId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, id)
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

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == userId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, apperr.ForbiddenError("forbidden")
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
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	strRole := ""
	if request.ShouldFilterByRole {
		strRole = string(request.MemberRole)
	}

	projects, err := ps.projectRepository.ListByUserId(ctx, request.UserId, strRole, request.SearchQuery)
	if err != nil {
		return nil, apperr.ServerError("failed to list projects", err)
	}

	return projects, nil
}

type UpdateProjectRequest struct {
	Id               uuid.UUID
	Name             string
	Description      string
	RepositoryURL    string
	RepositoryOwner  string
	RepositoryName   string
	DefaultBranch    string
	BranchNamePrefix string
	UserId           uuid.UUID
	Columns          []domain.ProjectColumn
	DeletedColumns   []DeletedProjectColumnRequest
}

type DeletedProjectColumnRequest struct {
	Id                  uuid.UUID
	MoveTasksToColumnId uuid.UUID
}

type UpdateProjectColumnRequest struct {
	ProjectId    uuid.UUID
	ColumnId     uuid.UUID
	UserId       uuid.UUID
	Name         string
	Description  string
	Color        string
	IsDoneColumn bool
}

func (ps *ProjectService) Update(ctx context.Context, request UpdateProjectRequest) (*domain.Project, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.Id)
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

	if project.UserId != request.UserId {
		return nil, apperr.ForbiddenError("forbidden")
	}

	project.Name = request.Name
	project.Description = request.Description
	project.RepositoryURL = request.RepositoryURL
	project.RepositoryOwner = request.RepositoryOwner
	project.RepositoryName = request.RepositoryName
	project.DefaultBranch = request.DefaultBranch
	project.BranchNamePrefix = request.BranchNamePrefix
	project.UpdatedAt = time.Now()
	request.Columns = restoreMissingProjectColumnIDs(request.Columns, project.Columns, request.DeletedColumns)

	if err := validateProjectColumns(request.Columns); err != nil {
		return nil, err
	}

	existingColumns := map[uuid.UUID]domain.ProjectColumn{}
	for _, column := range project.Columns {
		existingColumns[column.Id] = column
	}

	deletedColumns := map[uuid.UUID]DeletedProjectColumnRequest{}
	for _, deleted := range request.DeletedColumns {
		deletedColumns[deleted.Id] = deleted
	}

	for _, column := range request.Columns {
		if column.Id == uuid.Nil {
			continue
		}

		if _, ok := existingColumns[column.Id]; !ok {
			return nil, apperr.BusinessValidationError("invalid project column")
		}
	}

	finalColumnsByID := map[uuid.UUID]domain.ProjectColumn{}
	for i, column := range request.Columns {
		column.Position = i
		column.ProjectId = project.Id
		finalColumnsByID[column.Id] = column
	}

	deletions := []ProjectColumnDeletion{}
	for existingID := range existingColumns {
		if _, ok := finalColumnsByID[existingID]; ok {
			continue
		}

		deleted, ok := deletedColumns[existingID]
		if !ok {
			return nil, apperr.BusinessValidationError("deleted column reassignment is required")
		}

		targetColumn, ok := finalColumnsByID[deleted.MoveTasksToColumnId]
		if !ok {
			return nil, apperr.BusinessValidationError("deleted column target is invalid")
		}

		deletions = append(deletions, ProjectColumnDeletion{
			ColumnID:     existingID,
			TargetColumn: targetColumn,
		})
	}

	project.Columns = request.Columns

	err = ps.projectRepository.UpdateWithColumns(ctx, UpdateProjectWithColumnsParams{
		Project:   project,
		Columns:   request.Columns,
		Deletions: deletions,
	}, func() []outbox.Message {
		return []outbox.Message{{
			Topic:       events.ProjectUpdated,
			AggregateID: project.Id,
			Payload: &events.ProjectUpdatedPayload{
				Project:      *project,
				ActionOrigin: domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.UserId,
				},
			},
		}}
	})
	if err != nil {
		return nil, apperr.ServerError("failed to update project", err)
	}

	return project, nil
}

func (ps *ProjectService) UpdateColumn(ctx context.Context, request UpdateProjectColumnRequest) (*domain.ProjectColumn, error) {
	if request.UserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
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

	if project.UserId != request.UserId {
		return nil, apperr.ForbiddenError("forbidden")
	}

	column, err := ps.projectRepository.GetColumnById(ctx, request.ColumnId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, apperr.NotFoundError("project column not found")
			}
			return nil, domainErr
		}

		return nil, apperr.ServerError("failed to get project column", err)
	}

	if column.ProjectId != project.Id {
		return nil, apperr.NotFoundError("project column not found")
	}

	switchingToDoneColumn := request.IsDoneColumn && !column.IsDoneColumn

	column.Name = request.Name
	column.Description = request.Description
	column.Color = request.Color

	if err := validateSingleProjectColumn(*column); err != nil {
		return nil, err
	}

	if request.IsDoneColumn != column.IsDoneColumn {
		if !request.IsDoneColumn {
			return nil, apperr.BusinessValidationError("project must have exactly one done column")
		}

		for i := range project.Columns {
			if project.Columns[i].Id == column.Id {
				continue
			}
			if !project.Columns[i].IsDoneColumn {
				continue
			}

			project.Columns[i].IsDoneColumn = false
			if err := ps.projectRepository.UpdateColumn(ctx, &project.Columns[i]); err != nil {
				return nil, apperr.ServerError("failed to update previous done project column", err)
			}
			break
		}
	}

	column.IsDoneColumn = request.IsDoneColumn
	column.ProjectId = project.Id
	now := time.Now()
	column.UpdatedAt = now

	if err := ps.projectRepository.UpdateColumn(ctx, column); err != nil {
		return nil, apperr.ServerError("failed to update project column", err)
	}

	project.UpdatedAt = now
	for i := range project.Columns {
		if project.Columns[i].Id == column.Id {
			project.Columns[i] = *column
			continue
		}
		if switchingToDoneColumn && project.Columns[i].IsDoneColumn {
			project.Columns[i].IsDoneColumn = false
		}
	}

	if err := ps.projectRepository.MarkUpdatedAt(ctx, project.Id, outbox.Message{
		Topic:       events.ProjectUpdated,
		AggregateID: project.Id,
		Payload: &events.ProjectUpdatedPayload{
			Project:      *project,
			ActionOrigin: domain.ActionOriginFromContext(ctx),
			User: domain.User{
				Id: request.UserId,
			},
		},
	}); err != nil {
		return nil, apperr.ServerError("failed to mark project updated at", err)
	}

	return column, nil
}

func defaultProjectColumns(columns []domain.ProjectColumn) []domain.ProjectColumn {
	if len(columns) > 0 {
		for i := range columns {
			columns[i].Position = i
			if columns[i].Color == "" {
				columns[i].Color = defaultProjectColumnColors[i%len(defaultProjectColumnColors)]
			}
		}
		return columns
	}

	return []domain.ProjectColumn{
		{Name: "Pending", Color: defaultProjectColumnColors[0], Position: 0, IsDoneColumn: false},
		{Name: "Doing", Color: defaultProjectColumnColors[1], Position: 1, IsDoneColumn: false},
		{Name: "Done", Color: defaultProjectColumnColors[2], Position: 2, IsDoneColumn: true},
	}
}

func validateProjectColumns(columns []domain.ProjectColumn) error {
	if len(columns) == 0 {
		return apperr.BusinessValidationError("project must have at least one column")
	}

	doneColumns := 0
	for _, column := range columns {
		if column.Name == "" {
			return apperr.BusinessValidationError("project column name is required")
		}
		if !projectColumnColorPattern.MatchString(column.Color) {
			return apperr.BusinessValidationError("project column color must be a valid hex value")
		}
		if column.IsDoneColumn {
			doneColumns++
		}
	}

	if doneColumns != 1 {
		return apperr.BusinessValidationError("project must have exactly one done column")
	}

	return nil
}

func restoreMissingProjectColumnIDs(
	columns []domain.ProjectColumn,
	existingColumns []domain.ProjectColumn,
	deletedColumns []DeletedProjectColumnRequest,
) []domain.ProjectColumn {
	if len(columns) != len(existingColumns) || len(deletedColumns) > 0 {
		return columns
	}

	missingIDs := 0
	for _, column := range columns {
		if column.Id == uuid.Nil {
			missingIDs++
			continue
		}

		return columns
	}

	if missingIDs == 0 {
		return columns
	}

	restored := make([]domain.ProjectColumn, len(columns))
	copy(restored, columns)

	for i := range restored {
		restored[i].Id = existingColumns[i].Id
	}

	return restored
}

func validateSingleProjectColumn(column domain.ProjectColumn) error {
	if column.Name == "" {
		return apperr.BusinessValidationError("project column name is required")
	}
	if !projectColumnColorPattern.MatchString(column.Color) {
		return apperr.BusinessValidationError("project column color must be a valid hex value")
	}

	return nil
}

type CreateMemberRequest struct {
	ProjectId     uuid.UUID
	Email         string
	RequestUserId uuid.UUID
}

func (ps *ProjectService) CreateMember(ctx context.Context, request CreateMemberRequest) (*domain.ProjectMember, error) {
	if request.RequestUserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	user, err := ps.userRepository.GetByEmail(ctx, request.Email)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return nil, apperr.NotFoundError("user not found")
			}
			return nil, domainErr
		}
		return nil, apperr.ServerError("failed to get user", err)
	}

	if user.Id == request.RequestUserId {
		return nil, apperr.BusinessValidationError("you cannot add yourself as a member")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
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

	if project.UserId != request.RequestUserId {
		return nil, apperr.ForbiddenError("forbidden")
	}

	alreadyMember := false
	for _, member := range project.Members {
		if member.UserId == user.Id {
			alreadyMember = true
			break
		}
	}
	if alreadyMember {
		return nil, apperr.DuplicateEntryError("member already exists")
	}

	member := domain.ProjectMember{
		ProjectId: request.ProjectId,
		UserId:    user.Id,
		Role:      domain.ProjectMemberRoleMember,
	}

	err = ps.projectRepository.CreateMember(ctx, &member, func() []outbox.Message {
		return []outbox.Message{{
			Topic:       events.ProjectMemberCreated,
			AggregateID: member.ProjectId,
			Payload: &events.ProjectMemberCreatedPayload{
				ProjectMember: member,
				ActionOrigin:  domain.ActionOriginFromContext(ctx),
				User: domain.User{
					Id: request.RequestUserId,
				},
			},
		}}
	})
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.DuplicateEntryErrorCode {
				return nil, apperr.DuplicateEntryError("member already exists")
			}
			return nil, domainErr
		}
		return nil, apperr.ServerError("failed to create member", err)
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
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, apperr.ServerError("failed to get project", err)
	}

	hasPermission := false
	for _, member := range project.Members {
		if member.UserId == request.UserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, apperr.ForbiddenError("forbidden")
	}

	params := ListProjectActivityLogsParams{
		ProjectId:       request.ProjectId,
		BeforeCreatedAt: request.BeforeCreatedAt,
		BeforeId:        request.BeforeId,
		Limit:           request.Limit,
		UserId:          uuid.Nil,
	}

	activities, err := ps.activityRepository.List(ctx, params)
	if err != nil {
		return nil, apperr.ServerError("failed to list activities", err)
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
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	params := ListProjectActivityLogsParams{
		UserId:          request.UserId,
		BeforeCreatedAt: request.BeforeCreatedAt,
		BeforeId:        request.BeforeId,
		Limit:           request.Limit,
		ProjectId:       uuid.Nil,
	}

	activities, err := ps.activityRepository.List(ctx, params)
	if err != nil {
		return nil, apperr.ServerError("failed to list activities", err)
	}

	return activities, nil
}

func (ps *ProjectService) ListMembersByProjectId(ctx context.Context, projectId uuid.UUID, requestUserId uuid.UUID) ([]domain.ProjectMember, error) {
	if requestUserId == uuid.Nil {
		return nil, apperr.UnauthorizedError("unauthorized")
	}

	members, err := ps.projectRepository.ListMembersByProjectId(ctx, projectId)
	if err != nil {
		return nil, apperr.ServerError("failed to list members", err)
	}

	hasPermission := false
	for _, member := range members {
		if member.UserId == requestUserId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return nil, apperr.ForbiddenError("forbidden")
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
		return apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return apperr.NotFoundError("project not found")
			}
			return domainErr
		}

		return apperr.ServerError("failed to get project", err)
	}

	member := domain.ProjectMember{}
	var requestUserRole domain.ProjectMemberRole

	for _, m := range project.Members {
		if m.UserId == request.UserId {
			if m.Role == domain.ProjectMemberRoleCreator {
				return apperr.BusinessValidationError("cannot remove creator from project")
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
		return apperr.ForbiddenError("forbidden")
	}

	err = ps.projectRepository.RemoveMember(ctx, request.ProjectId, request.UserId, outbox.Message{
		Topic:       events.ProjectMemberRemoved,
		AggregateID: request.ProjectId,
		Payload: &events.ProjectMemberRemovedPayload{
			ProjectMember: member,
			ActionOrigin:  domain.ActionOriginFromContext(ctx),
			User: domain.User{
				Id: request.RequestUserId,
			},
		},
	})
	if err != nil {
		return apperr.ServerError("failed to remove member", err)
	}

	return nil
}

type DeleteProjectRequest struct {
	ProjectId     uuid.UUID
	RequestUserId uuid.UUID
}

func (ps *ProjectService) Delete(ctx context.Context, request DeleteProjectRequest) error {
	if request.RequestUserId == uuid.Nil {
		return apperr.UnauthorizedError("unauthorized")
	}

	project, err := ps.projectRepository.GetById(ctx, request.ProjectId)
	if err != nil {
		var domainErr apperr.DomainError
		if errors.As(err, &domainErr) {
			if domainErr.Code == apperr.NotFoundErrorCode {
				return apperr.NotFoundError("project not found")
			}
			return domainErr
		}

		return apperr.ServerError("failed to get project", err)
	}

	if project.UserId != request.RequestUserId {
		return apperr.ForbiddenError("forbidden")
	}

	err = ps.projectRepository.Delete(ctx, request.ProjectId, outbox.Message{
		Topic:       events.ProjectDeleted,
		AggregateID: project.Id,
		Payload: &events.ProjectDeletedPayload{
			Project:      *project,
			ActionOrigin: domain.ActionOriginFromContext(ctx),
			User: domain.User{
				Id: request.RequestUserId,
			},
		},
	})
	if err != nil {
		return apperr.ServerError("failed to delete project", err)
	}

	return nil
}
