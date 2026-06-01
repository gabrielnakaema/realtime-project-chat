package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
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

func mustGenerateTestOrder(t *testing.T, left, right string) string {
	t.Helper()

	key, err := fracindex.GenerateKeyBetween(left, right)
	require.NoError(t, err)

	return key
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

func (m *mockTaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, projectId, projectColumnIDs, archived, taskOrder, cursorUpdatedAt, limit)
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

func (m *mockTaskRepository) GetFirstTaskInColumn(ctx context.Context, projectId uuid.UUID, projectColumnID uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, projectId, projectColumnID)
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

func (m *mockTaskRepository) CountTasksByProjectIdAndColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID) (map[string]int, error) {
	args := m.Called(ctx, projectId, projectColumnIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *mockTaskRepository) ListUserDueTasks(ctx context.Context, userId uuid.UUID, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, userId, cursorDueDate, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func (m *mockTaskRepository) SearchTasksForUser(ctx context.Context, userId uuid.UUID, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, userId, searchQuery, cursorDueDate, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func TestTaskService_Create(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
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

	validUser := domain.User{
		Id:    validUserId,
		Name:  "Test User",
		Email: "user@example.com",
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Author:          &validUser,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		ProjectColumn:   &domain.ProjectColumn{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Order:           mustGenerateTestOrder(t, "", ""),
		Updates:         []domain.TaskUpdate{},
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
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   validUserId,
				Priority:        string(domain.TaskPriorityLow),
				ResponsibleId:   nil,
				DueDate:         nil,
				Tags:            []string{},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, pendingStatusID).Return(nil, domain.NotFoundError("not found"))
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
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "project not found",
			request: service.CreateTaskRequest{
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   validUserId,
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
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   uuid.New(),
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
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   validUserId,
				Tags:            []string{"", "tag1", "tag2", "", "tag3"},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, pendingStatusID).Return(nil, domain.NotFoundError("not found"))
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
		{
			name: "responsible must be a project member",
			request: service.CreateTaskRequest{
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				RequestUserId:   validUserId,
				ResponsibleId: func() *uuid.UUID {
					id := uuid.New()
					return &id
				}(),
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.BusinessValidationErrorCode),
			expectedError:     domain.BusinessValidationError("responsible is not a member of the project"),
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

func TestTaskService_GroupByColumn_DefaultStatuses(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()
	doneStatusID := uuid.New()

	mockRepo := &mockTaskRepository{}
	mockProjectRepo := &mockProjectRepository{}
	mockUserRepo := &mockUserRepository{}

	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&domain.Project{
		Id: validProjectId,
		Members: []domain.ProjectMember{
			{UserId: validUserId},
		},
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}, nil)

	emptyPage := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{}, HasNext: false}
	mockRepo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil).Once()
	mockRepo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{doingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil).Once()
	mockRepo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{doneStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil).Once()

	svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})

	result, err := svc.GroupByColumn(context.Background(), service.GroupByColumnRequest{
		ProjectId: validProjectId,
		UserId:    validUserId,
		Limit:     15,
	})

	require.NoError(t, err)
	require.Len(t, result, 3)
	_, ok := result[pendingStatusID.String()]
	require.True(t, ok)
	_, ok = result[doingStatusID.String()]
	require.True(t, ok)
	_, ok = result[doneStatusID.String()]
	require.True(t, ok)
}

func TestTaskService_Update(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()
	doneStatusID := uuid.New()

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
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Author:          &validUser,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		ProjectColumn:   &domain.ProjectColumn{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Updates:         []domain.TaskUpdate{},
		Tags:            []string{"tag1", "tag2", "tag3"},
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
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Deascription",
				ProjectColumnId: doneStatusID,
				RequestUserId:   validUserId,
				Priority:        domain.TaskPriorityHigh,
				ResponsibleId:   nil,
				DueDate:         nil,
				Tags:            []string{},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
			},
			expectedTask:              &validTask,
			shouldSucceed:             true,
			expectedError:             nil,
			expectedTaskUpdatesLength: 0,
		},
		{
			name: "unauthorized error",
			request: service.UpdateTaskRequest{
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Description",
				ProjectColumnId: doingStatusID,
				RequestUserId:   uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
			expectedError:     domain.UnauthorizedError("unauthorized"),
		},
		{
			name: "project not found",
			request: service.UpdateTaskRequest{
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Description",
				ProjectColumnId: doingStatusID,
				RequestUserId:   validUserId,
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
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Description",
				ProjectColumnId: doingStatusID,
				RequestUserId:   uuid.New(),
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
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Description",
				ProjectColumnId: doingStatusID,
				RequestUserId:   validUserId,
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
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Description",
				ProjectColumnId: doingStatusID,
				RequestUserId:   validUserId,
				Tags:            []string{"", "tag1", "tag2", "", "tag3"},
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
			},
			expectedTask:              &validTask,
			shouldSucceed:             true,
			expectedError:             nil,
			expectedTaskUpdatesLength: 0,
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
				assert.Equal(t, tt.request.ProjectColumnId, task.ProjectColumnId)
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
	pendingStatusID := uuid.New()

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
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		},
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		Priority:        domain.TaskPriorityLow,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Updates:         []domain.TaskUpdate{},
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
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
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
				repo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
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
				repo.On("Update", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Status == domain.TaskStatusArchived
				})).Return(nil)
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

func TestTaskService_Archive_PublishesPreviousProjectColumnID(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	pendingStatusID := uuid.New()

	validProject := domain.Project{
		Id:      validProjectId,
		UserId:  validUserId,
		Members: []domain.ProjectMember{{UserId: validUserId, Role: domain.ProjectMemberRoleCreator}},
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		},
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		Priority:        domain.TaskPriorityLow,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockRepo := &mockTaskRepository{}
	mockProjectRepo := &mockProjectRepository{}
	mockUserRepo := &mockUserRepository{}
	mockPublisher := &mockPublisher{}

	mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockPublisher.On("Publish", mock.Anything, events.TaskUpdated, mock.MatchedBy(func(payload events.Payload) bool {
		taskUpdatedPayload, ok := payload.(*events.TaskUpdatedPayload)
		if !ok {
			return false
		}

		return taskUpdatedPayload.PreviousProjectColumnID != nil &&
			*taskUpdatedPayload.PreviousProjectColumnID == pendingStatusID &&
			taskUpdatedPayload.Task.ArchivedAt != nil
	})).Return(nil)

	svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, mockPublisher)

	task, err := svc.Archive(context.Background(), service.ArchiveTaskRequest{
		TaskId:        validTaskId,
		RequestUserId: validUserId,
	})

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, domain.TaskStatusArchived, task.Status)

	mockRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestTaskService_Restore(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()

	archivedAt := time.Now().Add(-time.Hour)

	validProject := domain.Project{
		Id:      validProjectId,
		UserId:  validUserId,
		Members: []domain.ProjectMember{{UserId: validUserId, Role: domain.ProjectMemberRoleCreator}},
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
		},
	}

	archivedTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Title:           "Archived Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusArchived,
		ProjectColumnId: pendingStatusID,
		Priority:        domain.TaskPriorityLow,
		ArchivedAt:      &archivedAt,
		CreatedAt:       time.Now().Add(-2 * time.Hour),
		UpdatedAt:       time.Now().Add(-time.Hour),
	}

	mockRepo := &mockTaskRepository{}
	mockProjectRepo := &mockProjectRepository{}
	mockUserRepo := &mockUserRepository{}
	mockPublisher := &mockPublisher{}

	mockRepo.On("GetById", mock.Anything, validTaskId).Return(&archivedTask, nil)
	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(task *domain.Task) bool {
		return task.ProjectColumnId == doingStatusID &&
			task.ArchivedAt == nil &&
			task.Status == domain.TaskStatusDoing &&
			task.DoneAt == nil
	})).Return(nil)
	mockPublisher.On("Publish", mock.Anything, events.TaskUpdated, mock.MatchedBy(func(payload events.Payload) bool {
		taskUpdatedPayload, ok := payload.(*events.TaskUpdatedPayload)
		if !ok {
			return false
		}

		return taskUpdatedPayload.Task.ProjectColumnId == doingStatusID &&
			taskUpdatedPayload.Task.ArchivedAt == nil &&
			taskUpdatedPayload.PreviousProjectColumnID == nil
	})).Return(nil)

	svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, mockPublisher)

	task, err := svc.Restore(context.Background(), service.RestoreTaskRequest{
		TaskId:          validTaskId,
		ProjectColumnId: doingStatusID,
		RequestUserId:   validUserId,
	})

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, doingStatusID, task.ProjectColumnId)
	assert.Nil(t, task.ArchivedAt)
	assert.Equal(t, domain.TaskStatusDoing, task.Status)

	mockRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestTaskService_Update_LoadsResponsibleDetailsForChangedAssignee(t *testing.T) {
	validUserId := uuid.New()
	newResponsibleID := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	pendingStatusID := uuid.New()

	validUser := domain.User{
		Id:    validUserId,
		Name:  "Test User",
		Email: "user@example.com",
	}

	newResponsible := domain.User{
		Id:    newResponsibleID,
		Name:  "Maria",
		Email: "maria@example.com",
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
			{
				UserId: newResponsibleID,
				Role:   domain.ProjectMemberRoleMember,
			},
		},
		UserId: validUserId,
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		},
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Author:          &validUser,
		Title:           "Test Task",
		Description:     "Test Description",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockRepo := &mockTaskRepository{}
	mockProjectRepo := &mockProjectRepository{}
	mockUserRepo := &mockUserRepository{}

	mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
	mockUserRepo.On("GetById", mock.Anything, newResponsibleID).Return(&newResponsible, nil)

	svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
	task, err := svc.Update(context.Background(), service.UpdateTaskRequest{
		TaskId:          validTaskId,
		Title:           "Test Task",
		Description:     "Test Description",
		ProjectColumnId: pendingStatusID,
		RequestUserId:   validUserId,
		Priority:        domain.TaskPriorityLow,
		ResponsibleId:   &newResponsibleID,
		Tags:            []string{},
	})

	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotNil(t, task.Responsible)
	assert.Equal(t, newResponsibleID, task.Responsible.Id)
	assert.Equal(t, "Maria", task.Responsible.Name)

	mockRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTaskService_List(t *testing.T) {
	validUserId := uuid.New()
	memberUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
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
			{UserId: validUserId, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberUserId, Role: domain.ProjectMemberRoleMember},
		},
		UserId: validUserId,
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}

	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		Title:           "Test Task",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
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
				ProjectId:        validProjectId,
				RequestUserId:    validUserId,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Limit:            15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID}, false, "", (*time.Time)(nil), 15).Return(onePage, nil)
			},
			shouldSucceed: true,
			expectedLen:   1,
		},
		{
			name: "unauthorized",
			request: service.ListTasksRequest{
				ProjectId:        validProjectId,
				RequestUserId:    uuid.Nil,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Limit:            15,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "forbidden - not a member",
			request: service.ListTasksRequest{
				ProjectId:        validProjectId,
				RequestUserId:    uuid.New(),
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Limit:            15,
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
				ProjectId:        validProjectId,
				RequestUserId:    validUserId,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Archived:         true,
				Limit:            15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID}, true, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
			},
			shouldSucceed: true,
			expectedLen:   0,
		},
		{
			name: "non-creator can list non-archived tasks",
			request: service.ListTasksRequest{
				ProjectId:        validProjectId,
				RequestUserId:    memberUserId,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID, doingStatusID},
				Limit:            15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID, doingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
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

func TestTaskService_GroupByColumn(t *testing.T) {
	validUserId := uuid.New()
	memberUserId := uuid.New()
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
			{UserId: validUserId, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberUserId, Role: domain.ProjectMemberRoleMember},
		},
		UserId: validUserId,
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}

	emptyPage := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{}, HasNext: false}

	type testCase struct {
		name              string
		request           service.GroupByColumnRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful group by status",
			request: service.GroupByColumnRequest{
				ProjectId:        validProjectId,
				UserId:           validUserId,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID, doingStatusID},
				Limit:            15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{doingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
			},
			shouldSucceed: true,
		},
		{
			name: "unauthorized",
			request: service.GroupByColumnRequest{
				ProjectId:        validProjectId,
				UserId:           uuid.Nil,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Limit:            15,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "non-creator can group by non-archived statuses",
			request: service.GroupByColumnRequest{
				ProjectId:        validProjectId,
				UserId:           memberUserId,
				ProjectColumnIDs: []uuid.UUID{pendingStatusID, doneStatusID},
				Limit:            15,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				projectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{pendingStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
				repo.On("ListByProjectId", mock.Anything, validProjectId, []uuid.UUID{doneStatusID}, false, "", (*time.Time)(nil), 15).Return(emptyPage, nil)
			},
			shouldSucceed: true,
		},
		{
			name: "forbidden - not a project member",
			request: service.GroupByColumnRequest{
				ProjectId:        validProjectId,
				UserId:           uuid.New(),
				ProjectColumnIDs: []uuid.UUID{pendingStatusID},
				Limit:            15,
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

			result, err := svc.GroupByColumn(ctx, tt.request)

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

func TestTaskService_MarkTaskDone(t *testing.T) {
	requestUserID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	pendingColumnID := uuid.New()
	doneColumnID := uuid.New()

	project := &domain.Project{
		Id: projectID,
		Members: []domain.ProjectMember{
			{UserId: requestUserID},
		},
		Columns: []domain.ProjectColumn{
			{Id: pendingColumnID, ProjectId: projectID, Name: "Pending"},
			{Id: doneColumnID, ProjectId: projectID, Name: "Done", IsDoneColumn: true},
		},
	}
	task := &domain.Task{
		Id:              taskID,
		ProjectId:       projectID,
		ProjectColumnId: pendingColumnID,
		Status:          domain.TaskStatusPending,
	}

	t.Run("moves task into the single done column", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockRepo.On("GetById", mock.Anything, taskID).Return(task, nil)
		mockProjectRepo.On("GetById", mock.Anything, projectID).Return(project, nil)
		mockRepo.On("GetFirstTaskInColumn", mock.Anything, projectID, doneColumnID).Return(nil, domain.NotFoundError("not found"))
		mockRepo.On("MoveTask", mock.Anything, mock.MatchedBy(func(updated *domain.Task) bool {
			return updated.ProjectColumnId == doneColumnID && updated.DoneAt != nil
		}), requestUserID).Return(&domain.Task{
			Id:              taskID,
			ProjectId:       projectID,
			ProjectColumnId: doneColumnID,
			Status:          domain.TaskStatusDone,
			DoneAt:          func() *time.Time { now := time.Now(); return &now }(),
		}, nil)

		svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
		result, err := svc.MarkTaskDone(context.Background(), service.MarkTaskDoneRequest{
			TaskId:        taskID,
			RequestUserId: requestUserID,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, doneColumnID, result.ProjectColumnId)
		assert.NotNil(t, result.DoneAt)

		mockRepo.AssertExpectations(t)
		mockProjectRepo.AssertExpectations(t)
	})

	t.Run("fails when the project does not have exactly one done column", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		projectWithMultipleDoneColumns := &domain.Project{
			Id: projectID,
			Members: []domain.ProjectMember{
				{UserId: requestUserID},
			},
			Columns: []domain.ProjectColumn{
				{Id: pendingColumnID, ProjectId: projectID, Name: "Pending"},
				{Id: doneColumnID, ProjectId: projectID, Name: "Done", IsDoneColumn: true},
				{Id: uuid.New(), ProjectId: projectID, Name: "Verified", IsDoneColumn: true},
			},
		}

		mockRepo.On("GetById", mock.Anything, taskID).Return(task, nil)
		mockProjectRepo.On("GetById", mock.Anything, projectID).Return(projectWithMultipleDoneColumns, nil)

		svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
		result, err := svc.MarkTaskDone(context.Background(), service.MarkTaskDoneRequest{
			TaskId:        taskID,
			RequestUserId: requestUserID,
		})

		require.Error(t, err)
		require.Nil(t, result)

		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, string(domain.BusinessValidationErrorCode), string(domainErr.Code))
		assert.Equal(t, "project must have exactly one done column", domainErr.Message)

		mockRepo.AssertExpectations(t)
		mockProjectRepo.AssertExpectations(t)
	})
}

func TestTaskService_AssignTaskToSelf(t *testing.T) {
	requestUserID := uuid.New()
	projectID := uuid.New()
	taskID := uuid.New()
	columnID := uuid.New()

	project := &domain.Project{
		Id: projectID,
		Members: []domain.ProjectMember{
			{UserId: requestUserID},
		},
		Columns: []domain.ProjectColumn{
			{Id: columnID, ProjectId: projectID, Name: "Doing"},
		},
	}
	user := &domain.User{
		Id:    requestUserID,
		Name:  "Agent Owner",
		Email: "owner@example.com",
	}

	t.Run("assigns the task to the authenticated user", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		task := &domain.Task{
			Id:              taskID,
			ProjectId:       projectID,
			ProjectColumnId: columnID,
			ProjectColumn:   &domain.ProjectColumn{Id: columnID, ProjectId: projectID, Name: "Doing"},
			Title:           "Task",
			Description:     "Desc",
			Priority:        domain.TaskPriorityMedium,
			Tags:            []string{"backend"},
		}

		mockRepo.On("GetById", mock.Anything, taskID).Return(task, nil)
		mockProjectRepo.On("GetById", mock.Anything, projectID).Return(project, nil)
		mockUserRepo.On("GetById", mock.Anything, requestUserID).Return(user, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(updated *domain.Task) bool {
			return updated.ResponsibleId != nil && *updated.ResponsibleId == requestUserID
		})).Return(nil)

		svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
		result, err := svc.AssignTaskToSelf(context.Background(), service.AssignTaskToSelfRequest{
			TaskId:        taskID,
			RequestUserId: requestUserID,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.ResponsibleId)
		assert.Equal(t, requestUserID, *result.ResponsibleId)

		mockRepo.AssertExpectations(t)
		mockProjectRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("returns the existing task when it is already assigned to the user", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		task := &domain.Task{
			Id:              taskID,
			ProjectId:       projectID,
			ProjectColumnId: columnID,
			ResponsibleId:   &requestUserID,
			Title:           "Task",
			Description:     "Desc",
			Priority:        domain.TaskPriorityMedium,
		}

		mockRepo.On("GetById", mock.Anything, taskID).Return(task, nil)
		mockProjectRepo.On("GetById", mock.Anything, projectID).Return(project, nil)

		svc := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
		result, err := svc.AssignTaskToSelf(context.Background(), service.AssignTaskToSelfRequest{
			TaskId:        taskID,
			RequestUserId: requestUserID,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.ResponsibleId)
		assert.Equal(t, requestUserID, *result.ResponsibleId)

		mockRepo.AssertExpectations(t)
		mockProjectRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}
