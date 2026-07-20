package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProjectRepository struct {
	mock.Mock
}

type mockActivityRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) Create(ctx context.Context, project *domain.Project, buildEvents func() []outbox.Message) error {
	args := m.Called(ctx, project)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) ListMembersByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.ProjectMember, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.ProjectMember), args.Error(1)
}

func (m *mockProjectRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectRepository) ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string, searchQuery string) ([]domain.Project, error) {
	args := m.Called(ctx, userId, memberRole, searchQuery)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *mockProjectRepository) UpdateWithColumns(ctx context.Context, params UpdateProjectWithColumnsParams, buildEvents func() []outbox.Message) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *mockProjectRepository) MarkUpdatedAt(ctx context.Context, projectId uuid.UUID, msgs ...outbox.Message) error {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) UpdateColumn(ctx context.Context, status *domain.ProjectColumn) error {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) GetColumnById(ctx context.Context, id uuid.UUID) (*domain.ProjectColumn, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectColumn), args.Error(1)
}

func (m *mockProjectRepository) CreateMember(ctx context.Context, member *domain.ProjectMember, buildEvents func() []outbox.Message) error {
	args := m.Called(ctx, member)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) RemoveMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID, msgs ...outbox.Message) error {
	args := m.Called(ctx, projectId, userId)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) GetMemberByUserIdAndProjectId(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) (*domain.ProjectMember, error) {
	args := m.Called(ctx, projectId, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProjectMember), args.Error(1)
}

func (m *mockActivityRepository) List(ctx context.Context, params ListProjectActivityLogsParams) (*utils.CursorPaginated[domain.ProjectActivity], error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.ProjectActivity]), args.Error(1)
}

func TestProjectService_Create(t *testing.T) {
	validUserId := uuid.New()

	type testCase struct {
		name              string
		request           CreateProjectRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedProject   *domain.Project
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project creation",
			request: CreateProjectRequest{
				Name:             "Test Project",
				Description:      "Test Description",
				RepositoryURL:    "https://github.com/acme/project-chat-pubsub",
				RepositoryOwner:  "acme",
				RepositoryName:   "project-chat-pubsub",
				DefaultBranch:    "main",
				BranchNamePrefix: "task/",
				UserId:           validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Project")).Return(nil)
			},
			expectedProject: &domain.Project{
				Id:          uuid.New(),
				Name:        "Test Project",
				Description: "Test Description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Members: []domain.ProjectMember{
					{
						UserId: validUserId,
						Role:   domain.ProjectMemberRoleCreator,
					},
				},
				UserId: validUserId,
			},
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "unauthorized error",
			request: CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				UserId:      uuid.Nil,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.UnauthorizedErrorCode),
			expectedError:     apperr.UnauthorizedError("unauthorized"),
		},
		{
			name: "server error",
			request: CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				UserId:      validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Project")).Return(errors.New("error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to create project", errors.New("error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)
			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			project, err := service.Create(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.request.RepositoryURL, project.RepositoryURL)
				assert.Equal(t, tt.request.RepositoryOwner, project.RepositoryOwner)
				assert.Equal(t, tt.request.RepositoryName, project.RepositoryName)
				assert.Equal(t, tt.request.DefaultBranch, project.DefaultBranch)
				assert.Equal(t, tt.request.BranchNamePrefix, project.BranchNamePrefix)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}
		})
	}

}

func TestProjectService_GetById(t *testing.T) {
	validUserId := uuid.New()
	validMemberUserId := uuid.New()
	validProjectId := uuid.New()

	validProject := &domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: validMemberUserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
			{
				UserId: validUserId,
				Role:   domain.ProjectMemberRoleMember,
			},
		},
		UserId: validUserId,
	}

	type testCase struct {
		name              string
		id                uuid.UUID
		userId            uuid.UUID
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedProject   *domain.Project
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name:   "successful project retrieval by owner",
			id:     validProjectId,
			userId: validUserId,
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			expectedProject: validProject,
			shouldSucceed:   true,
			expectedError:   nil,
		},
		{
			name:   "successful project retrieval by member",
			id:     validProjectId,
			userId: validMemberUserId,
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			expectedProject: validProject,
			shouldSucceed:   true,
			expectedError:   nil,
		},
		{
			name:   "project not found",
			id:     uuid.New(),
			userId: validUserId,
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
			expectedError:     apperr.NotFoundError("project not found"),
		},
		{
			name:   "forbidden",
			id:     validProjectId,
			userId: uuid.New(),
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
			expectedError:     apperr.ForbiddenError("forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)
			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			project, err := service.GetById(ctx, tt.id, tt.userId)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.id, project.Id)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_ListByUserId(t *testing.T) {
	validUserId := uuid.New()
	validMemberUserId := uuid.New()
	validProjectId := uuid.New()

	validProject := domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: validMemberUserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
		UserId: validUserId,
	}

	type testCase struct {
		name              string
		request           ListProjectsByUserIdRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project list by user id",
			request: ListProjectsByUserIdRequest{
				UserId:             validUserId,
				MemberRole:         domain.ProjectMemberRoleCreator,
				ShouldFilterByRole: true,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("ListByUserId", mock.Anything, validUserId, "creator", "").Return([]domain.Project{validProject}, nil)
			},
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "throws server error",
			request: ListProjectsByUserIdRequest{
				UserId:             validUserId,
				MemberRole:         domain.ProjectMemberRoleCreator,
				ShouldFilterByRole: true,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("ListByUserId", mock.Anything, validUserId, "creator", "").Return(nil, errors.New("server error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to list projects", errors.New("server error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			projects, err := service.ListByUserId(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, projects)
			} else {
				require.Error(t, err)
				require.Nil(t, projects)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_Update(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()
	doneStatusID := uuid.New()

	validProject := domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: validUserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
		UserId: validUserId,
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}

	updatedColumns := []domain.ProjectColumn{
		{Id: pendingStatusID, ProjectId: validProjectId, Name: "Backlog", Color: "#64748B", Position: 0},
		{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
		{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
	}

	type testCase struct {
		name              string
		request           UpdateProjectRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedProject   *domain.Project
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project update",
			request: UpdateProjectRequest{
				Id:               validProjectId,
				Name:             "Updated Project",
				Description:      "Updated Description",
				RepositoryURL:    "https://github.com/acme/project-chat-pubsub",
				RepositoryOwner:  "acme",
				RepositoryName:   "project-chat-pubsub",
				DefaultBranch:    "main",
				BranchNamePrefix: "task/",
				UserId:           validUserId,
				Columns:          updatedColumns,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("UpdateWithColumns", mock.Anything, mock.AnythingOfType("UpdateProjectWithColumnsParams")).Return(nil)
			},
			expectedProject: &validProject,
			shouldSucceed:   true,
			expectedError:   nil,
		},
		{
			name: "restores missing column ids for unchanged project updates",
			request: UpdateProjectRequest{
				Id:               validProjectId,
				Name:             "Updated Project",
				Description:      "Updated Description",
				RepositoryURL:    "https://github.com/acme/project-chat-pubsub",
				RepositoryOwner:  "acme",
				RepositoryName:   "project-chat-pubsub",
				DefaultBranch:    "main",
				BranchNamePrefix: "task/",
				UserId:           validUserId,
				Columns: []domain.ProjectColumn{
					{Name: "Backlog", Color: "#64748B", Position: 0},
					{Name: "Doing", Color: "#2563EB", Position: 1},
					{Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
				},
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("UpdateWithColumns", mock.Anything, mock.AnythingOfType("UpdateProjectWithColumnsParams")).Return(nil)
			},
			expectedProject: &validProject,
			shouldSucceed:   true,
			expectedError:   nil,
		},
		{
			name: "throws server error",
			request: UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
				Columns:     updatedColumns,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("UpdateWithColumns", mock.Anything, mock.AnythingOfType("UpdateProjectWithColumnsParams")).Return(errors.New("server error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to update project", errors.New("server error")),
		},
		{
			name: "throws validation error without writing anything when a column is invalid",
			request: UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
				Columns: []domain.ProjectColumn{
					{Id: uuid.New(), ProjectId: validProjectId, Name: "Unknown", Color: "#64748B", Position: 0},
					{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
					{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
				},
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
			expectedError:     apperr.BusinessValidationError("invalid project column"),
		},
		{
			name: "throws validation error without writing anything when a deleted column has no reassignment",
			request: UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
				Columns: []domain.ProjectColumn{
					{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 0},
					{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 1, IsDoneColumn: true},
				},
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
			expectedError:     apperr.BusinessValidationError("deleted column reassignment is required"),
		},
		{
			name: "throws validation error without writing anything when a deleted column target is invalid",
			request: UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
				Columns: []domain.ProjectColumn{
					{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 0},
					{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 1, IsDoneColumn: true},
				},
				DeletedColumns: []DeletedProjectColumnRequest{
					{Id: pendingStatusID, MoveTasksToColumnId: uuid.New()},
				},
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
			expectedError:     apperr.BusinessValidationError("deleted column target is invalid"),
		},
		{
			name: "throws forbidden error",
			request: UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      uuid.New(),
				Columns:     updatedColumns,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
			expectedError:     apperr.ForbiddenError("forbidden"),
		},
		{
			name: "throws not found error",
			request: UpdateProjectRequest{
				Id:          uuid.New(),
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
				Columns:     updatedColumns,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
			expectedError:     apperr.NotFoundError("project not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			project, err := service.Update(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)

				assert.Equal(t, tt.request.Id, project.Id)
				assert.Equal(t, tt.request.Name, project.Name)
				assert.Equal(t, tt.request.Description, project.Description)
				assert.Equal(t, tt.request.RepositoryURL, project.RepositoryURL)
				assert.Equal(t, tt.request.RepositoryOwner, project.RepositoryOwner)
				assert.Equal(t, tt.request.RepositoryName, project.RepositoryName)
				assert.Equal(t, tt.request.DefaultBranch, project.DefaultBranch)
				assert.Equal(t, tt.request.BranchNamePrefix, project.BranchNamePrefix)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}

				if tt.expectedErrorCode == string(apperr.BusinessValidationErrorCode) {
					mockRepo.AssertNotCalled(t, "UpdateWithColumns", mock.Anything, mock.Anything)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_UpdateColumn(t *testing.T) {
	validUserID := uuid.New()
	validProjectID := uuid.New()
	backlogColumnID := uuid.New()
	doneColumnID := uuid.New()

	newBaseProject := func() *domain.Project {
		return &domain.Project{
			Id:          validProjectID,
			Name:        "Test Project",
			Description: "Test Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			UserId:      validUserID,
			Columns: []domain.ProjectColumn{
				{Id: backlogColumnID, ProjectId: validProjectID, Name: "Backlog", Description: "Waiting work", Color: "#64748B", Position: 0, IsDoneColumn: false},
				{Id: doneColumnID, ProjectId: validProjectID, Name: "Done", Description: "Completed", Color: "#059669", Position: 1, IsDoneColumn: true},
			},
		}
	}

	type testCase struct {
		name              string
		request           UpdateProjectColumnRequest
		mockSetup         func(*mockProjectRepository)
		expectedName      string
		expectedIsDone    bool
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful single column update",
			request: UpdateProjectColumnRequest{
				ProjectId:    validProjectID,
				ColumnId:     backlogColumnID,
				UserId:       validUserID,
				Name:         "In Progress",
				Description:  "Currently active",
				Color:        "#2563EB",
				IsDoneColumn: false,
			},
			mockSetup: func(repo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validProjectID).Return(newBaseProject(), nil)
				repo.On("GetColumnById", mock.Anything, backlogColumnID).Return(&domain.ProjectColumn{
					Id:           backlogColumnID,
					ProjectId:    validProjectID,
					Name:         "Backlog",
					Description:  "Waiting work",
					Color:        "#64748B",
					Position:     0,
					IsDoneColumn: false,
				}, nil)
				repo.On("UpdateColumn", mock.Anything, mock.MatchedBy(func(column *domain.ProjectColumn) bool {
					return column.Id == backlogColumnID && column.Name == "In Progress" && column.Description == "Currently active" && column.Color == "#2563EB" && !column.IsDoneColumn
				})).Return(nil)
				repo.On("MarkUpdatedAt", mock.Anything, validProjectID).Return(nil)
			},
			expectedName:   "In Progress",
			expectedIsDone: false,
			shouldSucceed:  true,
		},
		{
			name: "switches done column and clears previous one",
			request: UpdateProjectColumnRequest{
				ProjectId:    validProjectID,
				ColumnId:     backlogColumnID,
				UserId:       validUserID,
				Name:         "Done Later",
				Description:  "Now complete",
				Color:        "#2563EB",
				IsDoneColumn: true,
			},
			mockSetup: func(repo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validProjectID).Return(newBaseProject(), nil)
				repo.On("GetColumnById", mock.Anything, backlogColumnID).Return(&domain.ProjectColumn{
					Id:           backlogColumnID,
					ProjectId:    validProjectID,
					Name:         "Backlog",
					Description:  "Waiting work",
					Color:        "#64748B",
					Position:     0,
					IsDoneColumn: false,
				}, nil)
				repo.On("UpdateColumn", mock.Anything, mock.MatchedBy(func(column *domain.ProjectColumn) bool {
					return column.Id == doneColumnID && !column.IsDoneColumn
				})).Return(nil)
				repo.On("UpdateColumn", mock.Anything, mock.MatchedBy(func(column *domain.ProjectColumn) bool {
					return column.Id == backlogColumnID && column.IsDoneColumn
				})).Return(nil)
				repo.On("MarkUpdatedAt", mock.Anything, validProjectID).Return(nil)
			},
			expectedName:   "Done Later",
			expectedIsDone: true,
			shouldSucceed:  true,
		},
		{
			name: "rejects removing the only done column",
			request: UpdateProjectColumnRequest{
				ProjectId:    validProjectID,
				ColumnId:     doneColumnID,
				UserId:       validUserID,
				Name:         "Done",
				Description:  "Completed",
				Color:        "#059669",
				IsDoneColumn: false,
			},
			mockSetup: func(repo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validProjectID).Return(newBaseProject(), nil)
				repo.On("GetColumnById", mock.Anything, doneColumnID).Return(&domain.ProjectColumn{
					Id:           doneColumnID,
					ProjectId:    validProjectID,
					Name:         "Done",
					Description:  "Completed",
					Color:        "#059669",
					Position:     1,
					IsDoneColumn: true,
				}, nil)
			},
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo)

			svc := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)

			column, err := svc.UpdateColumn(context.Background(), tt.request)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, column)
				assert.Equal(t, tt.expectedName, column.Name)
				assert.Equal(t, tt.expectedIsDone, column.IsDoneColumn)
			} else {
				require.Error(t, err)
				require.Nil(t, column)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_CreateMember(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validMemberUserId := uuid.New()
	existingMemberUserId := uuid.New()

	validProject := &domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: validUserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
			{
				UserId: existingMemberUserId,
				Role:   domain.ProjectMemberRoleMember,
			},
		},
		UserId: validUserId,
	}

	validUser := &domain.User{
		Id:    validMemberUserId,
		Name:  "Test Member",
		Email: "member@example.com",
	}

	validMember := &domain.ProjectMember{
		ProjectId: validProjectId,
		UserId:    validMemberUserId,
		Role:      domain.ProjectMemberRoleMember,
	}

	type testCase struct {
		name              string
		request           CreateMemberRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedMember    *domain.ProjectMember
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful member creation",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
				repo.On("CreateMember", mock.Anything, mock.AnythingOfType("*domain.ProjectMember")).Return(nil)
			},
			shouldSucceed:  true,
			expectedError:  nil,
			expectedMember: validMember,
		},
		{
			name: "unauthorized error",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: uuid.Nil,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.UnauthorizedErrorCode),
			expectedError:     apperr.UnauthorizedError("unauthorized"),
		},
		{
			name: "user not found",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "nonexistent@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, apperr.NotFoundError("user not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
			expectedError:     apperr.NotFoundError("user not found"),
		},
		{
			name: "cannot add yourself as member",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "creator@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				selfUser := &domain.User{
					Id:    validUserId,
					Name:  "Creator",
					Email: "creator@example.com",
				}
				userRepo.On("GetByEmail", mock.Anything, "creator@example.com").Return(selfUser, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
			expectedError:     apperr.BusinessValidationError("you cannot add yourself as a member"),
		},
		{
			name: "project not found",
			request: CreateMemberRequest{
				ProjectId:     uuid.New(),
				Email:         "member@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
			expectedError:     apperr.NotFoundError("project not found"),
		},
		{
			name: "forbidden - not project owner",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
			expectedError:     apperr.ForbiddenError("forbidden"),
		},
		{
			name: "member already exists",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "existing@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				existingMemberUser := &domain.User{
					Id:    existingMemberUserId,
					Name:  "Existing Member",
					Email: "existing@example.com",
				}
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingMemberUser, nil)
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.DuplicateEntryErrorCode),
			expectedError:     apperr.DuplicateEntryError("member already exists"),
		},
		{
			name: "server error on create member",
			request: CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
				repo.On("CreateMember", mock.Anything, mock.AnythingOfType("*domain.ProjectMember")).Return(errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to create member", errors.New("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			member, err := service.CreateMember(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, member)
				assert.Equal(t, tt.request.ProjectId, member.ProjectId)
				assert.Equal(t, domain.ProjectMemberRoleMember, member.Role)
			} else {
				require.Error(t, err)
				require.Nil(t, member)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestProjectService_RemoveMember(t *testing.T) {
	creatorUserId := uuid.New()
	memberUserId := uuid.New()
	otherMemberUserId := uuid.New()
	validProjectId := uuid.New()

	validProject := &domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{
				UserId: creatorUserId,
				Role:   domain.ProjectMemberRoleCreator,
			},
			{
				UserId: memberUserId,
				Role:   domain.ProjectMemberRoleMember,
			},
			{
				UserId: otherMemberUserId,
				Role:   domain.ProjectMemberRoleMember,
			},
		},
		UserId: creatorUserId,
	}

	type testCase struct {
		name              string
		request           RemoveMemberRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful member removal",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: creatorUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
				repo.On("RemoveMember", mock.Anything, validProjectId, memberUserId).Return(nil)
			},
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "unauthorized error",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: uuid.Nil,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.UnauthorizedErrorCode),
			expectedError:     apperr.UnauthorizedError("unauthorized"),
		},
		{
			name: "server error on get project",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: creatorUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(nil, errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to get project", errors.New("database error")),
		},
		{
			name: "project not found",
			request: RemoveMemberRequest{
				ProjectId:     uuid.New(),
				UserId:        memberUserId,
				RequestUserId: creatorUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
			expectedError:     apperr.NotFoundError("project not found"),
		},
		{
			name: "cannot remove creator from project",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        creatorUserId,
				RequestUserId: creatorUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.BusinessValidationErrorCode),
			expectedError:     apperr.BusinessValidationError("cannot remove creator from project"),
		},
		{
			name: "member removes themselves from project",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: memberUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
				repo.On("RemoveMember", mock.Anything, validProjectId, memberUserId).Return(nil)
			},
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "forbidden - member trying to remove another member",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        otherMemberUserId,
				RequestUserId: memberUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
			expectedError:     apperr.ForbiddenError("forbidden"),
		},
		{
			name: "forbidden - not project member",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
			expectedError:     apperr.ForbiddenError("forbidden"),
		},
		{
			name: "server error on remove member",
			request: RemoveMemberRequest{
				ProjectId:     validProjectId,
				UserId:        memberUserId,
				RequestUserId: creatorUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
				repo.On("RemoveMember", mock.Anything, validProjectId, memberUserId).Return(errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
			expectedError:     apperr.ServerError("failed to remove member", errors.New("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockActivityRepo := &mockActivityRepository{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := NewProjectService(mockRepo, mockUserRepo, mockActivityRepo)
			ctx := context.Background()

			err := service.RemoveMember(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
