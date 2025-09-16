package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProjectRepository struct {
	mock.Mock
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, topic events.Topic, payload interface{}) error {
	return nil
}

func (m *mockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectRepository) ListByUserId(ctx context.Context, userId uuid.UUID, memberRole string) ([]domain.Project, error) {
	args := m.Called(ctx, userId, memberRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *mockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) CreateMember(ctx context.Context, member *domain.ProjectMember) error {
	args := m.Called(ctx, member)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockProjectRepository) RemoveMember(ctx context.Context, projectId uuid.UUID, userId uuid.UUID) error {
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

func TestProjectService_Create(t *testing.T) {
	validUserId := uuid.New()

	type testCase struct {
		name              string
		request           service.CreateProjectRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedProject   *domain.Project
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project creation",
			request: service.CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				UserId:      validUserId,
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
			request: service.CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				UserId:      uuid.Nil,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "server error",
			request: service.CreateProjectRequest{
				Name:        "Test Project",
				Description: "Test Description",
				UserId:      validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Project")).Return(errors.New("error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ServerErrorCode),
			expectedError:     domain.ServerError("failed to create project", errors.New("error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockPublisher := &mockPublisher{}
			tt.mockSetup(mockRepo, mockUserRepo)
			service := service.NewProjectService(mockRepo, mockUserRepo, mockPublisher)
			ctx := context.Background()

			project, err := service.Create(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr domain.DomainError
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
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("project not found"),
		},
		{
			name:   "forbidden",
			id:     validProjectId,
			userId: uuid.New(),
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
			expectedError:     domain.ForbiddenError("forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockPublisher := &mockPublisher{}
			tt.mockSetup(mockRepo, mockUserRepo)
			service := service.NewProjectService(mockRepo, mockUserRepo, mockPublisher)
			ctx := context.Background()

			project, err := service.GetById(ctx, tt.id, tt.userId)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.id, project.Id)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr domain.DomainError
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
		request           service.ListProjectsByUserIdRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project list by user id",
			request: service.ListProjectsByUserIdRequest{
				UserId:             validUserId,
				MemberRole:         domain.ProjectMemberRoleCreator,
				ShouldFilterByRole: true,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("ListByUserId", mock.Anything, validUserId, "creator").Return([]domain.Project{validProject}, nil)
			},
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "throws server error",
			request: service.ListProjectsByUserIdRequest{
				UserId:             validUserId,
				MemberRole:         domain.ProjectMemberRoleCreator,
				ShouldFilterByRole: true,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("ListByUserId", mock.Anything, validUserId, "creator").Return(nil, errors.New("server error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ServerErrorCode),
			expectedError:     domain.ServerError("failed to list projects", errors.New("server error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockPublisher := &mockPublisher{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := service.NewProjectService(mockRepo, mockUserRepo, mockPublisher)
			ctx := context.Background()

			projects, err := service.ListByUserId(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, projects)
			} else {
				require.Error(t, err)
				require.Nil(t, projects)

				var domainErr domain.DomainError
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
	}

	type testCase struct {
		name              string
		request           service.UpdateProjectRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedProject   *domain.Project
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful project update",
			request: service.UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Update", mock.Anything, &validProject).Return(nil)
			},
			expectedProject: &validProject,
			shouldSucceed:   true,
			expectedError:   nil,
		},
		{
			name: "throws server error",
			request: service.UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Update", mock.Anything, &validProject).Return(errors.New("server error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ServerErrorCode),
			expectedError:     domain.ServerError("failed to update project", errors.New("server error")),
		},
		{
			name: "throws forbidden error",
			request: service.UpdateProjectRequest{
				Id:          validProjectId,
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      uuid.New(),
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
			expectedError:     domain.ForbiddenError("forbidden"),
		},
		{
			name: "throws not found error",
			request: service.UpdateProjectRequest{
				Id:          uuid.New(),
				Name:        "Updated Project",
				Description: "Updated Description",
				UserId:      validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("project not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockPublisher := &mockPublisher{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := service.NewProjectService(mockRepo, mockUserRepo, mockPublisher)
			ctx := context.Background()

			project, err := service.Update(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, project)

				assert.Equal(t, tt.request.Id, project.Id)
				assert.Equal(t, tt.request.Name, project.Name)
				assert.Equal(t, tt.request.Description, project.Description)
			} else {
				require.Error(t, err)
				require.Nil(t, project)

				var domainErr domain.DomainError
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
		request           service.CreateMemberRequest
		mockSetup         func(*mockProjectRepository, *mockUserRepository)
		expectedMember    *domain.ProjectMember
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful member creation",
			request: service.CreateMemberRequest{
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
			request: service.CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: uuid.Nil,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "user not found",
			request: service.CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "nonexistent@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, domain.NotFoundError("user not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("user not found"),
		},
		{
			name: "cannot add yourself as member",
			request: service.CreateMemberRequest{
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
			expectedErrorCode: string(domain.BusinessValidationErrorCode),
			expectedError:     domain.BusinessValidationError("you cannot add yourself as a member"),
		},
		{
			name: "project not found",
			request: service.CreateMemberRequest{
				ProjectId:     uuid.New(),
				Email:         "member@example.com",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, mock.Anything).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("project not found"),
		},
		{
			name: "forbidden - not project owner",
			request: service.CreateMemberRequest{
				ProjectId:     validProjectId,
				Email:         "member@example.com",
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockProjectRepository, userRepo *mockUserRepository) {
				userRepo.On("GetByEmail", mock.Anything, "member@example.com").Return(validUser, nil)
				repo.On("GetById", mock.Anything, validProjectId).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
			expectedError:     domain.ForbiddenError("forbidden"),
		},
		{
			name: "member already exists",
			request: service.CreateMemberRequest{
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
			expectedErrorCode: string(domain.DuplicateEntryErrorCode),
			expectedError:     domain.DuplicateEntryError("member already exists"),
		},
		{
			name: "server error on create member",
			request: service.CreateMemberRequest{
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
			expectedErrorCode: string(domain.ServerErrorCode),
			expectedError:     domain.ServerError("failed to create member", errors.New("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			mockPublisher := &mockPublisher{}
			tt.mockSetup(mockRepo, mockUserRepo)

			service := service.NewProjectService(mockRepo, mockUserRepo, mockPublisher)
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

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
