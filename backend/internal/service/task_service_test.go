package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTaskRepository struct {
	mock.Mock
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.Task, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Task), args.Error(1)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockTaskRepository) CreateChanges(ctx context.Context, task *domain.Task, changes []domain.TaskChange) error {
	args := m.Called(ctx, task, changes)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func TestTaskService_Create(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()

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

	validUser := domain.User{
		Id:    validUserId,
		Name:  "Test User",
		Email: "user@example.com",
	}

	validTask := domain.Task{
		Id:          validTaskId,
		ProjectId:   validProjectId,
		AuthorId:    validUserId,
		Author:      &validUser,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Changes:     []domain.TaskChange{},
	}

	type testCase struct {
		name              string
		request           service.CreateTaskRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedTask      *domain.Task
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful task creation",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateChanges", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskChange")).Return(nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
			},
			expectedTask:  &validTask,
			shouldSucceed: true,
			expectedError: nil,
		},
		{
			name: "unauthorized error",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "project not found",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("project not found"),
		},
		{
			name: "forbidden error",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
			expectedError:     domain.ForbiddenError("forbidden"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			tt.mockSetup(mockRepo, mockProjectRepo, mockUserRepo)
			service := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			task, err := service.Create(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, task)

				assert.Equal(t, tt.expectedTask.Title, task.Title)
				assert.Equal(t, tt.expectedTask.Description, task.Description)
				assert.Equal(t, tt.expectedTask.Status, domain.TaskStatusPending)
			} else {
				require.Error(t, err)
				require.Nil(t, task)

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_Update(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()

	validUser := domain.User{
		Id:    validUserId,
		Name:  "Test User",
		Email: "user@example.com",
	}

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

	validTask := domain.Task{
		Id:          validTaskId,
		ProjectId:   validProjectId,
		AuthorId:    validUserId,
		Author:      &validUser,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Changes:     []domain.TaskChange{},
	}

	type testCase struct {
		name                      string
		request                   service.UpdateTaskRequest
		mockSetup                 func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedTask              *domain.Task
		expectedError             error
		expectedErrorCode         string
		shouldSucceed             bool
		expectedTaskChangesLength int
	}

	tests := []testCase{
		{
			name: "successful task update",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateChanges", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskChange")).Return(nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
			},
			expectedTask:              &validTask,
			shouldSucceed:             true,
			expectedError:             nil,
			expectedTaskChangesLength: 3,
		},
		{
			name: "unauthorized error",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "project not found",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("project not found"),
		},
		{
			name: "forbidden error",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)

			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
			expectedError:     domain.ForbiddenError("forbidden"),
		},
		{
			name: "task not found",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(nil, domain.NotFoundError("task not found"))
			},
			expectedTask:      &validTask,
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
			expectedError:     domain.NotFoundError("task not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			tt.mockSetup(mockRepo, mockProjectRepo, mockUserRepo)
			service := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			task, err := service.Update(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, task)

				assert.Equal(t, tt.request.Title, task.Title)
				assert.Equal(t, tt.request.Description, task.Description)
				assert.Equal(t, tt.request.Status, task.Status)
				assert.Equal(t, tt.expectedTaskChangesLength, len(task.Changes))
			} else {
				require.Error(t, err)
				require.Nil(t, task)

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
