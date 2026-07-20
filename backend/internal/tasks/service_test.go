package tasks_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/events"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
	"github.com/gabrielnakaema/project-chat/internal/outbox"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTaskRepository struct {
	mock.Mock
	builtEvents     []outbox.Message
	moveLockColumns []uuid.UUID
}

func mustGenerateTestOrder(t *testing.T, left, right string) string {
	t.Helper()

	key, err := fracindex.GenerateKeyBetween(left, right)
	require.NoError(t, err)

	return key
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error {
	args := m.Called(ctx, task)
	err := args.Error(0)
	if err == nil && buildEvents != nil {
		m.builtEvents = append(m.builtEvents, buildEvents(task)...)
	}
	return err
}

func (m *mockTaskRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) FindTaskRefsByProjectAndCode(ctx context.Context, projectId uuid.UUID, code string) ([]domain.TaskDependencyRef, error) {
	args := m.Called(ctx, projectId, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskDependencyRef), args.Error(1)
}

func (m *mockTaskRepository) SuggestTaskCodesByProjectPrefix(ctx context.Context, projectId uuid.UUID, prefix string, limit int) ([]domain.TaskCodeSuggestion, error) {
	args := m.Called(ctx, projectId, prefix, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskCodeSuggestion), args.Error(1)
}

func (m *mockTaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, projectId, projectColumnIDs, archived, taskOrder, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task, buildEvents func(*domain.Task) []outbox.Message) error {
	args := m.Called(ctx, task)
	err := args.Error(0)
	if err == nil && buildEvents != nil {
		m.builtEvents = append(m.builtEvents, buildEvents(task)...)
	}
	return err
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

func (m *mockTaskRepository) WithProjectColumnMoveLock(ctx context.Context, projectColumnID uuid.UUID, fn func(context.Context) error) error {
	m.moveLockColumns = append(m.moveLockColumns, projectColumnID)
	return fn(ctx)
}

func (m *mockTaskRepository) MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID, buildEvents func(*domain.Task) []outbox.Message) (*domain.Task, error) {
	args := m.Called(ctx, task, userId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	result := args.Get(0).(*domain.Task)
	if args.Error(1) == nil && buildEvents != nil {
		m.builtEvents = append(m.builtEvents, buildEvents(result)...)
	}
	return result, args.Error(1)
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

func (m *mockTaskRepository) SearchProjectTasksForDependencies(ctx context.Context, projectId uuid.UUID, query string, excludeTaskId *uuid.UUID, limit int) ([]domain.TaskDependencyRef, error) {
	args := m.Called(ctx, projectId, query, excludeTaskId, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskDependencyRef), args.Error(1)
}

func (m *mockTaskRepository) SearchTasksForUser(ctx context.Context, userId uuid.UUID, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	args := m.Called(ctx, userId, searchQuery, cursorDueDate, cursorUpdatedAt, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.Task]), args.Error(1)
}

func (m *mockTaskRepository) GetTaskDependencyRefsByProjectAndIds(ctx context.Context, projectId uuid.UUID, taskIds []uuid.UUID) ([]domain.TaskDependencyRef, error) {
	args := m.Called(ctx, projectId, taskIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskDependencyRef), args.Error(1)
}

func (m *mockTaskRepository) ListTaskDependenciesByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.TaskDependencyEdge, error) {
	args := m.Called(ctx, projectId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TaskDependencyEdge), args.Error(1)
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
		request           tasks.CreateTaskRequest
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
			request: tasks.CreateTaskRequest{
				ProjectId:       validProjectId,
				ProjectColumnId: pendingStatusID,
				Title:           "Test Task",
				Description:     "Test Description",
				Code:            "  TASK-101  ",
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
			checkFunc: func(t *testing.T, task *domain.Task) bool {
				return assert.Equal(t, "TASK-101", task.Code)
			},
		},
		{
			name: "unauthorized error",
			request: tasks.CreateTaskRequest{
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
			request: tasks.CreateTaskRequest{
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
			request: tasks.CreateTaskRequest{
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
			request: tasks.CreateTaskRequest{
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
			request: tasks.CreateTaskRequest{
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
			svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			task, err := svc.Create(ctx, tt.request)

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

func TestTaskService_FindTaskByCode(t *testing.T) {
	validUserID := uuid.New()
	validProjectID := uuid.New()
	validTaskID := uuid.New()
	validColumnID := uuid.New()

	validProject := &domain.Project{
		Id: validProjectID,
		Members: []domain.ProjectMember{
			{UserId: validUserID, Role: domain.ProjectMemberRoleCreator},
		},
		Columns: []domain.ProjectColumn{
			{Id: validColumnID, ProjectId: validProjectID, Name: "Pending"},
		},
	}
	validTask := &domain.Task{
		Id:              validTaskID,
		ProjectId:       validProjectID,
		ProjectColumnId: validColumnID,
		Title:           "Resolve MCP lookup",
		Code:            "BACKEND-5",
	}

	t.Run("returns matching task", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}

		projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
		taskRepo.On("FindTaskRefsByProjectAndCode", mock.Anything, validProjectID, "BACKEND-5").Return([]domain.TaskDependencyRef{
			{Id: validTaskID, Title: validTask.Title, Code: validTask.Code},
		}, nil)
		taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		task, err := svc.FindTaskByCode(context.Background(), tasks.FindTaskByCodeRequest{
			ProjectId: validProjectID,
			UserId:    validUserID,
			Code:      "  BACKEND-5  ",
		})

		require.NoError(t, err)
		require.NotNil(t, task)
		assert.Equal(t, validTaskID, task.Id)
	})

	t.Run("returns not found when no active task matches", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}

		projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
		taskRepo.On("FindTaskRefsByProjectAndCode", mock.Anything, validProjectID, "BACKEND-5").Return([]domain.TaskDependencyRef{}, nil)

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		task, err := svc.FindTaskByCode(context.Background(), tasks.FindTaskByCodeRequest{
			ProjectId: validProjectID,
			UserId:    validUserID,
			Code:      "BACKEND-5",
		})

		require.Error(t, err)
		assert.Nil(t, task)
		assert.Equal(t, domain.NotFoundErrorCode, err.(domain.DomainError).Code)
	})

	t.Run("returns ambiguity error when duplicates exist", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}

		projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
		taskRepo.On("FindTaskRefsByProjectAndCode", mock.Anything, validProjectID, "BACKEND-5").Return([]domain.TaskDependencyRef{
			{Id: validTaskID, Title: "Resolve MCP lookup", Code: "BACKEND-5"},
			{Id: uuid.New(), Title: "Duplicate task", Code: "BACKEND-5"},
		}, nil)

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		task, err := svc.FindTaskByCode(context.Background(), tasks.FindTaskByCodeRequest{
			ProjectId: validProjectID,
			UserId:    validUserID,
			Code:      "BACKEND-5",
		})

		require.Error(t, err)
		assert.Nil(t, task)
		domainErr := err.(domain.DomainError)
		assert.Equal(t, domain.BusinessValidationErrorCode, domainErr.Code)
		assert.Equal(t, "task code matches multiple tasks in this project", domainErr.Message)
	})
}

func TestTaskService_SuggestTaskCodes(t *testing.T) {
	validUserID := uuid.New()
	validProjectID := uuid.New()
	validProject := &domain.Project{
		Id:      validProjectID,
		Members: []domain.ProjectMember{{UserId: validUserID, Role: domain.ProjectMemberRoleCreator}},
	}

	t.Run("returns suggestions for project member", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}
		expected := []domain.TaskCodeSuggestion{
			{Code: "BACKEND-3", Kind: "next"},
			{Code: "BACKEND-2", Kind: "existing"},
		}

		projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
		taskRepo.On("SuggestTaskCodesByProjectPrefix", mock.Anything, validProjectID, "BACKEND-", 8).Return(expected, nil)

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		result, err := svc.SuggestTaskCodes(context.Background(), tasks.SuggestTaskCodesRequest{
			ProjectId: validProjectID,
			UserId:    validUserID,
			Prefix:    "  BACKEND-  ",
			Limit:     8,
		})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		taskRepo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
	})

	t.Run("rejects short prefix", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		result, err := svc.SuggestTaskCodes(context.Background(), tasks.SuggestTaskCodesRequest{
			ProjectId: validProjectID,
			UserId:    validUserID,
			Prefix:    "B",
			Limit:     8,
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.BusinessValidationErrorCode, err.(domain.DomainError).Code)
	})

	t.Run("rejects non-member", func(t *testing.T) {
		taskRepo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		userRepo := &mockUserRepository{}

		projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)

		svc := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
		result, err := svc.SuggestTaskCodes(context.Background(), tasks.SuggestTaskCodesRequest{
			ProjectId: validProjectID,
			UserId:    uuid.New(),
			Prefix:    "BACKEND-",
			Limit:     8,
		})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ForbiddenErrorCode, err.(domain.DomainError).Code)
		projectRepo.AssertExpectations(t)
	})
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

	svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)

	result, err := svc.GroupByColumn(context.Background(), tasks.GroupByColumnRequest{
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
		request                   tasks.UpdateTaskRequest
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
			request: tasks.UpdateTaskRequest{
				TaskId:          validTaskId,
				Title:           "Updated Task",
				Description:     "Updated Deascription",
				Code:            func() *string { s := "  TASK-202  "; return &s }(),
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
			checkFunc: func(t *testing.T, task *domain.Task) bool {
				return assert.Equal(t, "TASK-202", task.Code)
			},
		},
		{
			name: "unauthorized error",
			request: tasks.UpdateTaskRequest{
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
			request: tasks.UpdateTaskRequest{
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
			request: tasks.UpdateTaskRequest{
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
			request: tasks.UpdateTaskRequest{
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
			request: tasks.UpdateTaskRequest{
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
			svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			task, err := svc.Update(ctx, tt.request)

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

	taskWithCode := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Author:          &validUser,
		Title:           "Test Task",
		Description:     "Test Description",
		Code:            "EXISTING-42",
		Status:          domain.TaskStatusPending,
		ProjectColumnId: pendingStatusID,
		ProjectColumn:   &domain.ProjectColumn{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Updates:         []domain.TaskUpdate{},
		Tags:            []string{},
	}

	baseUpdateRequest := func() tasks.UpdateTaskRequest {
		return tasks.UpdateTaskRequest{
			TaskId:          validTaskId,
			Title:           "Test Task",
			Description:     "Test Description",
			ProjectColumnId: pendingStatusID,
			RequestUserId:   validUserId,
			Priority:        domain.TaskPriorityLow,
			Tags:            []string{},
		}
	}

	t.Run("nil Code preserves existing task code", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&taskWithCode, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		req := baseUpdateRequest()
		req.Code = nil // not provided — keep existing

		task, err := svc.Update(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "EXISTING-42", task.Code, "existing code must be preserved when Code is nil")
	})

	t.Run("pointer to empty string clears existing task code", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&taskWithCode, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		req := baseUpdateRequest()
		empty := ""
		req.Code = &empty

		task, err := svc.Update(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "", task.Code, "code must be cleared when Code points to an empty string")
	})

	t.Run("pointer to new value replaces existing task code", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&taskWithCode, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		req := baseUpdateRequest()
		newCode := "  NEW-99  "
		req.Code = &newCode

		task, err := svc.Update(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "NEW-99", task.Code, "new code must be trimmed and applied")
	})
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
		request           tasks.ArchiveTaskRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful archive",
			request: tasks.ArchiveTaskRequest{
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
			request: tasks.ArchiveTaskRequest{
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
			request: tasks.ArchiveTaskRequest{
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
			request: tasks.ArchiveTaskRequest{
				TaskId:        validTaskId,
				RequestUserId: uuid.Nil,
			},
			mockSetup:         func(repo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {},
			shouldSucceed:     false,
			expectedErrorCode: string(domain.UnauthorizedErrorCode),
		},
		{
			name: "task not found",
			request: tasks.ArchiveTaskRequest{
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
			request: tasks.ArchiveTaskRequest{
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
			request: tasks.ArchiveTaskRequest{
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
			svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
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

	mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

	svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)

	task, err := svc.Archive(context.Background(), tasks.ArchiveTaskRequest{
		TaskId:        validTaskId,
		RequestUserId: validUserId,
	})

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, domain.TaskStatusArchived, task.Status)

	require.Len(t, mockRepo.builtEvents, 1)
	assert.Equal(t, events.TaskUpdated, mockRepo.builtEvents[0].Topic)
	taskUpdatedPayload, ok := mockRepo.builtEvents[0].Payload.(*events.TaskUpdatedPayload)
	require.True(t, ok)
	require.NotNil(t, taskUpdatedPayload.PreviousProjectColumnID)
	assert.Equal(t, pendingStatusID, *taskUpdatedPayload.PreviousProjectColumnID)
	assert.NotNil(t, taskUpdatedPayload.Task.ArchivedAt)

	mockRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
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

	mockRepo.On("GetById", mock.Anything, validTaskId).Return(&archivedTask, nil)
	mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(task *domain.Task) bool {
		return task.ProjectColumnId == doingStatusID &&
			task.ArchivedAt == nil &&
			task.Status == domain.TaskStatusDoing &&
			task.DoneAt == nil
	})).Return(nil)

	svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)

	task, err := svc.Restore(context.Background(), tasks.RestoreTaskRequest{
		TaskId:          validTaskId,
		ProjectColumnId: doingStatusID,
		RequestUserId:   validUserId,
	})

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, doingStatusID, task.ProjectColumnId)
	assert.Nil(t, task.ArchivedAt)
	assert.Equal(t, domain.TaskStatusDoing, task.Status)

	require.Len(t, mockRepo.builtEvents, 1)
	assert.Equal(t, events.TaskUpdated, mockRepo.builtEvents[0].Topic)
	taskUpdatedPayload, ok := mockRepo.builtEvents[0].Payload.(*events.TaskUpdatedPayload)
	require.True(t, ok)
	assert.Equal(t, doingStatusID, taskUpdatedPayload.Task.ProjectColumnId)
	assert.Nil(t, taskUpdatedPayload.Task.ArchivedAt)
	assert.Nil(t, taskUpdatedPayload.PreviousProjectColumnID)

	mockRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
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

	svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
	task, err := svc.Update(context.Background(), tasks.UpdateTaskRequest{
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
		request           tasks.ListTasksRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
		expectedLen       int
	}

	tests := []testCase{
		{
			name: "successful list",
			request: tasks.ListTasksRequest{
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
			request: tasks.ListTasksRequest{
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
			request: tasks.ListTasksRequest{
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
			request: tasks.ListTasksRequest{
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
			request: tasks.ListTasksRequest{
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
			svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
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
		request           tasks.GroupByColumnRequest
		mockSetup         func(*mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
	}

	tests := []testCase{
		{
			name: "successful group by status",
			request: tasks.GroupByColumnRequest{
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
			request: tasks.GroupByColumnRequest{
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
			request: tasks.GroupByColumnRequest{
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
			request: tasks.GroupByColumnRequest{
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
			svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
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

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		result, err := svc.MarkTaskDone(context.Background(), tasks.MarkTaskDoneRequest{
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

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		result, err := svc.MarkTaskDone(context.Background(), tasks.MarkTaskDoneRequest{
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

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		result, err := svc.AssignTaskToSelf(context.Background(), tasks.AssignTaskToSelfRequest{
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

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		result, err := svc.AssignTaskToSelf(context.Background(), tasks.AssignTaskToSelfRequest{
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

func TestTaskService_Dependencies(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	dependencyTaskId := uuid.New()
	otherDependencyTaskId := uuid.New()
	pendingStatusID := uuid.New()
	doingStatusID := uuid.New()
	doneStatusID := uuid.New()

	validProject := domain.Project{
		Id:      validProjectId,
		UserId:  validUserId,
		Members: []domain.ProjectMember{{UserId: validUserId, Role: domain.ProjectMemberRoleCreator}},
		Columns: []domain.ProjectColumn{
			{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
			{Id: doingStatusID, ProjectId: validProjectId, Name: "Doing", Color: "#2563EB", Position: 1},
			{Id: doneStatusID, ProjectId: validProjectId, Name: "Done", Color: "#059669", Position: 2, IsDoneColumn: true},
		},
	}

	validUser := domain.User{Id: validUserId, Name: "Test User", Email: "user@example.com"}
	validTask := domain.Task{
		Id:              validTaskId,
		ProjectId:       validProjectId,
		AuthorId:        validUserId,
		Title:           "Task",
		Description:     "Description",
		ProjectColumnId: pendingStatusID,
		ProjectColumn:   &domain.ProjectColumn{Id: pendingStatusID, ProjectId: validProjectId, Name: "Pending", Color: "#64748B", Position: 0},
		Priority:        domain.TaskPriorityMedium,
	}

	t.Run("create clears empty dependencies to empty slice", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockUserRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
		mockRepo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, pendingStatusID).Return(nil, domain.NotFoundError("not found"))
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		task, err := svc.Create(context.Background(), tasks.CreateTaskRequest{
			ProjectId:       validProjectId,
			ProjectColumnId: pendingStatusID,
			Title:           "Independent task",
			Description:     "No dependencies",
			RequestUserId:   validUserId,
			Priority:        string(domain.TaskPriorityMedium),
		})

		require.NoError(t, err)
		require.NotNil(t, task)
		assert.Equal(t, []uuid.UUID{}, task.DependsOnTaskIds)
	})

	t.Run("create with dependencies", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockUserRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
		mockRepo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, pendingStatusID).Return(nil, domain.NotFoundError("not found"))
		mockRepo.On("GetTaskDependencyRefsByProjectAndIds", mock.Anything, validProjectId, []uuid.UUID{dependencyTaskId}).Return([]domain.TaskDependencyRef{{Id: dependencyTaskId, Title: "Dependency Task", Code: "DEP-1"}}, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		task, err := svc.Create(context.Background(), tasks.CreateTaskRequest{
			ProjectId:        validProjectId,
			ProjectColumnId:  pendingStatusID,
			Title:            "Blocked task",
			Description:      "Depends on another task",
			RequestUserId:    validUserId,
			Priority:         string(domain.TaskPriorityMedium),
			DependsOnTaskIds: []uuid.UUID{dependencyTaskId, dependencyTaskId},
		})

		require.NoError(t, err)
		require.NotNil(t, task)
		assert.Equal(t, []uuid.UUID{dependencyTaskId}, task.DependsOnTaskIds)
	})

	t.Run("create rejects unknown dependencies", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockUserRepo.On("GetById", mock.Anything, validUserId).Return(&validUser, nil)
		mockRepo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, pendingStatusID).Return(nil, domain.NotFoundError("not found"))
		mockRepo.On("GetTaskDependencyRefsByProjectAndIds", mock.Anything, validProjectId, []uuid.UUID{dependencyTaskId}).Return([]domain.TaskDependencyRef{}, nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		_, err := svc.Create(context.Background(), tasks.CreateTaskRequest{
			ProjectId:        validProjectId,
			ProjectColumnId:  pendingStatusID,
			Title:            "Blocked task",
			Description:      "Depends on another task",
			RequestUserId:    validUserId,
			Priority:         string(domain.TaskPriorityMedium),
			DependsOnTaskIds: []uuid.UUID{dependencyTaskId},
		})

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.BusinessValidationErrorCode, domainErr.Code)
	})

	t.Run("update rejects self dependency", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		_, err := svc.Update(context.Background(), tasks.UpdateTaskRequest{
			TaskId:           validTaskId,
			Title:            validTask.Title,
			Description:      validTask.Description,
			ProjectColumnId:  pendingStatusID,
			RequestUserId:    validUserId,
			Priority:         validTask.Priority,
			DependsOnTaskIds: []uuid.UUID{validTaskId},
		})

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.BusinessValidationErrorCode, domainErr.Code)
	})

	t.Run("update rejects dependency cycle", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockRepo.On("GetTaskDependencyRefsByProjectAndIds", mock.Anything, validProjectId, []uuid.UUID{dependencyTaskId}).Return([]domain.TaskDependencyRef{{Id: dependencyTaskId, Title: "Dependency Task", Code: "DEP-1"}}, nil)
		mockRepo.On("ListTaskDependenciesByProjectId", mock.Anything, validProjectId).Return([]domain.TaskDependencyEdge{
			{TaskId: dependencyTaskId, DependsOnTaskId: validTaskId},
		}, nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		_, err := svc.Update(context.Background(), tasks.UpdateTaskRequest{
			TaskId:           validTaskId,
			Title:            validTask.Title,
			Description:      validTask.Description,
			ProjectColumnId:  pendingStatusID,
			RequestUserId:    validUserId,
			Priority:         validTask.Priority,
			DependsOnTaskIds: []uuid.UUID{dependencyTaskId},
		})

		require.Error(t, err)
		var domainErr domain.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, domain.BusinessValidationErrorCode, domainErr.Code)
	})

	t.Run("update accepts valid dependencies", func(t *testing.T) {
		mockRepo := &mockTaskRepository{}
		mockProjectRepo := &mockProjectRepository{}
		mockUserRepo := &mockUserRepository{}

		mockRepo.On("GetById", mock.Anything, validTaskId).Return(&validTask, nil)
		mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&validProject, nil)
		mockRepo.On("GetTaskDependencyRefsByProjectAndIds", mock.Anything, validProjectId, []uuid.UUID{dependencyTaskId, otherDependencyTaskId}).Return([]domain.TaskDependencyRef{{Id: dependencyTaskId, Title: "Dependency Task", Code: "DEP-1"}, {Id: otherDependencyTaskId, Title: "Other Dependency Task", Code: "DEP-2"}}, nil)
		mockRepo.On("ListTaskDependenciesByProjectId", mock.Anything, validProjectId).Return([]domain.TaskDependencyEdge{}, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

		svc := tasks.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo)
		task, err := svc.Update(context.Background(), tasks.UpdateTaskRequest{
			TaskId:           validTaskId,
			Title:            validTask.Title,
			Description:      validTask.Description,
			ProjectColumnId:  pendingStatusID,
			RequestUserId:    validUserId,
			Priority:         validTask.Priority,
			DependsOnTaskIds: []uuid.UUID{dependencyTaskId, otherDependencyTaskId},
		})

		require.NoError(t, err)
		require.NotNil(t, task)
		assert.Equal(t, []uuid.UUID{dependencyTaskId, otherDependencyTaskId}, task.DependsOnTaskIds)
	})
}
