package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/gabrielnakaema/project-chat/internal/utils"
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

func (m *mockTaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID, statuses []string, taskOrder int, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, projectId, statuses, taskOrder, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockTaskRepository) CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error {
	args := m.Called(ctx, task, updates)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockTaskRepository) GetSmallestOrderProjectTask(ctx context.Context, projectId uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, id, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, task, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) NormalizeProjectTaskOrders(ctx context.Context, projectId uuid.UUID) error {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return args.Error(0)
	}
	return args.Error(0)
}

func (m *mockTaskRepository) CountTasksByProjectIdAndStatus(ctx context.Context, projectId uuid.UUID, statuses []string) (map[string]int, error) {
	args := m.Called(ctx, projectId, statuses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockTaskRepository) ListUserDueTasks(ctx context.Context, userId uuid.UUID, statuses []string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, userId, statuses, cursorDueDate, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func (m *mockTaskRepository) SearchTasksForUser(ctx context.Context, userId uuid.UUID, statuses []string, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, userId, statuses, searchQuery, cursorDueDate, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
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
		Updates:     []domain.TaskUpdate{},
	}

	type testCase struct {
		name              string
		request           service.CreateTaskRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedTask      *domain.Task
		expectedError     error
		expectedErrorCode string
		shouldSucceed     bool
		checkFunc         func(t *testing.T, task *domain.Task) bool
	}

	tests := []testCase{
		{
			name: "successful task creation",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: validUserId,
				Priority:      string(domain.TaskPriorityLow),
				ResponsibleId: nil,
				DueDate:       nil,
				Tags:          []string{},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetSmallestOrderProjectTask", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
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
		{
			name: "remove empty tags",
			request: service.CreateTaskRequest{
				ProjectId:     validProjectId,
				Title:         "Test Task",
				Description:   "Test Description",
				RequestUserId: validUserId,
				Tags:          []string{"", "tag1", "tag2", "", "tag3"},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetSmallestOrderProjectTask", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
			},
			expectedTask:  &validTask,
			shouldSucceed: true,
			expectedError: nil,
			checkFunc: func(t *testing.T, task *domain.Task) bool {
				return assert.Equal(t, []string{"tag1", "tag2", "tag3"}, task.Tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}

			tt.mockSetup(mockRepo, mockProjectRepo, mockUserRepo)
			service := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
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

			if tt.checkFunc != nil {
				assert.True(t, tt.checkFunc(t, task))
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
		Updates:     []domain.TaskUpdate{},
		Tags:        []string{"tag1", "tag2", "tag3"},
	}

	type testCase struct {
		name                      string
		request                   service.UpdateTaskRequest
		mockSetup                 func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedTask              *domain.Task
		expectedError             error
		expectedErrorCode         string
		shouldSucceed             bool
		expectedTaskUpdatesLength int
		checkFunc                 func(t *testing.T, task *domain.Task) bool
	}

	tests := []testCase{
		{
			name: "successful task update",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Deascription",
				Status:        domain.TaskStatusDone,
				RequestUserId: validUserId,
				Priority:      domain.TaskPriorityHigh,
				ResponsibleId: nil,
				DueDate:       nil,
				Tags:          []string{},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateUpdates", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskUpdate")).Return(nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
			},
			expectedTask:              &validTask,
			shouldSucceed:             true,
			expectedError:             nil,
			expectedTaskUpdatesLength: 1,
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
		{
			name: "remove empty tags",
			request: service.UpdateTaskRequest{
				TaskId:        validTaskId,
				Title:         "Updated Task",
				Description:   "Updated Description",
				Status:        domain.TaskStatusDoing,
				RequestUserId: validUserId,
				Tags:          []string{"", "tag1", "tag2", "", "tag3"},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateUpdates", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskUpdate")).Return(nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
			},
			expectedTask:              &validTask,
			shouldSucceed:             true,
			expectedError:             nil,
			expectedTaskUpdatesLength: 1,
			checkFunc: func(t *testing.T, task *domain.Task) bool {
				return assert.Equal(t, []string{"tag1", "tag2", "tag3"}, task.Tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			tt.mockSetup(mockRepo, mockProjectRepo, mockUserRepo)
			service := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
			ctx := context.Background()

			task, err := service.Update(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, task)

				assert.Equal(t, tt.request.Title, task.Title)
				assert.Equal(t, tt.request.Description, task.Description)
				assert.Equal(t, tt.request.Status, task.Status)
				assert.Equal(t, tt.expectedTaskUpdatesLength, len(task.Updates))
			} else {
				require.Error(t, err)
				require.Nil(t, task)

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			if tt.checkFunc != nil {
				assert.True(t, tt.checkFunc(t, task))
			}

			mockRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_Archive(t *testing.T) {
	validUserId := uuid.New()
	memberUserId := uuid.New()
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
			{UserId: validUserId, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberUserId, Role: domain.ProjectMemberRoleMember},
		},
		UserId: validUserId,
	}

	validTask := domain.Task{
		Id:          validTaskId,
		ProjectId:   validProjectId,
		AuthorId:    validUserId,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.TaskStatusPending,
		Priority:    domain.TaskPriorityLow,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Updates:     []domain.TaskUpdate{},
	}

	type testCase struct {
		name              string
		request           service.ArchiveTaskRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful archive",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateUpdates", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskUpdate")).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "member can archive task",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: memberUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				memberUser := domain.User{Id: memberUserId, Name: "Member", Email: "member@example.com"}
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				userRepo.On("GetById", mock.Anything, memberUserId).Return(&memberUser, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
				repo.On("CreateUpdates", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskUpdate")).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "archived task has archived status",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Status == domain.TaskStatusArchived
				})).Return(nil)
				repo.On("CreateUpdates", mock.Anything, mock.AnythingOfType("*domain.Task"), mock.AnythingOfType("[]domain.TaskUpdate")).Return(nil)
			},
			shouldSucceed: true,
		},
		{
			name: "unauthorized",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "task not found",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(nil, domain.NotFoundError("task not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
		},
		{
			name: "project not found",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: validUserId,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.NotFoundErrorCode),
		},
		{
			name: "forbidden - not a project member",
			request: service.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: uuid.New(),
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}

			tt.mockSetup(mockRepo, mockProjectRepo, mockUserRepo)
			svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
			ctx := context.Background()

			task, err := svc.Archive(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, domain.TaskStatusArchived, task.Status)
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

func TestTaskService_List(t *testing.T) {
	validUserId := uuid.New()
	memberUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()

	validProject := domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{UserId: validUserId, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberUserId, Role: domain.ProjectMemberRoleMember},
		},
		UserId: validUserId,
	}

	validTask := domain.Task{
		Id:        validTaskId,
		ProjectId: validProjectId,
		Title:     "Test Task",
		Status:    domain.TaskStatusPending,
	}

	emptyPage := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{}, HasNext: false}
	onePage := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{validTask}, HasNext: false}

	type testCase struct {
		name              string
		request           service.ListTasksRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
		expectedLen       int
	}

	tests := []testCase{
		{
			name: "successful list",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				Statuses:      []string{string(domain.TaskStatusPending)},
				Limit:         15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(onePage, nil)
			},
			shouldSucceed: true,
			expectedLen:   1,
		},
		{
			name: "unauthorized",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: uuid.Nil,
				Statuses:      []string{string(domain.TaskStatusPending)},
				Limit:         15,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "forbidden - not a member",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: uuid.New(),
				Statuses:      []string{string(domain.TaskStatusPending)},
				Limit:         15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
		},
		{
			name: "creator can list archived tasks",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				Statuses:      []string{string(domain.TaskStatusArchived)},
				Limit:         15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
			},
			shouldSucceed: true,
			expectedLen:   0,
		},
		{
			name: "non-creator forbidden from listing archived tasks",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: memberUserId,
				Statuses:      []string{string(domain.TaskStatusArchived)},
				Limit:         15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
		},
		{
			name: "non-creator can list non-archived tasks",
			request: service.ListTasksRequest{
				ProjectId:     validProjectId,
				RequestUserId: memberUserId,
				Statuses:      []string{string(domain.TaskStatusPending), string(domain.TaskStatusDoing)},
				Limit:         15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
			},
			shouldSucceed: true,
			expectedLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}

			tt.mockSetup(mockRepo, mockProjectRepo)
			svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
			ctx := context.Background()

			result, err := svc.List(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Data, tt.expectedLen)
			} else {
				require.Error(t, err)
				require.Nil(t, result)

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_GroupByStatus(t *testing.T) {
	validUserId := uuid.New()
	memberUserId := uuid.New()
	validProjectId := uuid.New()

	validProject := domain.Project{
		Id:          validProjectId,
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Members: []domain.ProjectMember{
			{UserId: validUserId, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberUserId, Role: domain.ProjectMemberRoleMember},
		},
		UserId: validUserId,
	}

	emptyPage := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{}, HasNext: false}

	type testCase struct {
		name              string
		request           service.GroupByStatusRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful group by status",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    validUserId,
				Statuses:  []domain.TaskStatus{domain.TaskStatusPending, domain.TaskStatusDoing},
				Limit:     15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []string{string(domain.TaskStatusPending)}, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []string{string(domain.TaskStatusDoing)}, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
			},
			shouldSucceed: true,
		},
		{
			name: "unauthorized",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    uuid.Nil,
				Statuses:  []domain.TaskStatus{domain.TaskStatusPending},
				Limit:     15,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "creator can group by archived status",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    validUserId,
				Statuses:  []domain.TaskStatus{domain.TaskStatusArchived},
				Limit:     15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []string{string(domain.TaskStatusArchived)}, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
			},
			shouldSucceed: true,
		},
		{
			name: "non-creator forbidden from grouping by archived status",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    memberUserId,
				Statuses:  []domain.TaskStatus{domain.TaskStatusArchived},
				Limit:     15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
		},
		{
			name: "non-creator can group by non-archived statuses",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    memberUserId,
				Statuses:  []domain.TaskStatus{domain.TaskStatusPending, domain.TaskStatusDone},
				Limit:     15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []string{string(domain.TaskStatusPending)}, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []string{string(domain.TaskStatusDone)}, mock.Anything, mock.Anything, mock.Anything).Return(emptyPage, nil)
			},
			shouldSucceed: true,
		},
		{
			name: "forbidden - not a project member",
			request: service.GroupByStatusRequest{
				ProjectId: validProjectId,
				UserId:    uuid.New(),
				Statuses:  []domain.TaskStatus{domain.TaskStatusPending},
				Limit:     15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.ForbiddenErrorCode),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}

			tt.mockSetup(mockRepo, mockProjectRepo)
			svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
			ctx := context.Background()

			result, err := svc.GroupByStatus(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			} else {
				require.Error(t, err)
				require.Nil(t, result)

				var domainErr domain.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}
