package tasks_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/outbox"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTaskCommentRepository struct {
	mock.Mock
}

func (m *mockTaskCommentRepository) Create(ctx context.Context, comment *domain.TaskComment, parentCommentID *uuid.UUID, buildEvents func(*domain.TaskComment) []outbox.Message) error {
	args := m.Called(ctx, comment, parentCommentID)
	return args.Error(0)
}

func (m *mockTaskCommentRepository) ListByTaskID(
	ctx context.Context,
	taskID uuid.UUID,
	before *time.Time,
	beforeID *uuid.UUID,
	after *time.Time,
	afterID *uuid.UUID,
	limit int,
) (*utils.CursorPaginated[domain.TaskComment], error) {
	args := m.Called(ctx, taskID, before, beforeID, after, afterID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.CursorPaginated[domain.TaskComment]), args.Error(1)
}

func TestTaskCommentService_Create(t *testing.T) {
	validUserID := uuid.New()
	validTaskID := uuid.New()
	validProjectID := uuid.New()
	validParentCommentID := uuid.New()

	validUser := &domain.User{
		Id:        validUserID,
		Name:      "Test User",
		Email:     "user@example.com",
		CreatedAt: time.Now(),
	}

	validTask := &domain.Task{
		Id:        validTaskID,
		ProjectId: validProjectID,
		Title:     "Test Task",
	}

	validProject := &domain.Project{
		Id: validProjectID,
		Members: []domain.ProjectMember{
			{
				UserId: validUserID,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
	}

	type testCase struct {
		name              string
		request           tasks.CreateTaskCommentRequest
		mockSetup         func(*mockTaskCommentRepository, *mockTaskRepository, *mockProjectRepository, *mockUserRepository)
		expectedErrorCode string
		shouldSucceed     bool
		checkFunc         func(t *testing.T, comment *domain.TaskComment)
	}

	tests := []testCase{
		{
			name: "successful comment creation",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserID).Return(validUser, nil)
				commentRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskComment"), (*uuid.UUID)(nil)).Return(nil).Run(func(args mock.Arguments) {
					comment := args.Get(1).(*domain.TaskComment)
					comment.ID = uuid.New().String()
				})
			},
			shouldSucceed: true,
			checkFunc: func(t *testing.T, comment *domain.TaskComment) {
				assert.Equal(t, validTaskID, comment.Task.Id)
				assert.Equal(t, validUserID, comment.User.Id)
				assert.Equal(t, "Test comment", comment.Content)
				assert.NotEmpty(t, comment.ID)
				assert.NotZero(t, comment.CreatedAt)
				assert.NotZero(t, comment.UpdatedAt)
				assert.Empty(t, comment.Replies)
			},
		},
		{
			name: "successful reply creation",
			request: tasks.CreateTaskCommentRequest{
				TaskID:          validTaskID,
				RequestUserID:   validUserID,
				Content:         "Reply comment",
				ParentCommentID: &validParentCommentID,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserID).Return(validUser, nil)
				commentRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskComment"), &validParentCommentID).Return(nil).Run(func(args mock.Arguments) {
					comment := args.Get(1).(*domain.TaskComment)
					comment.ID = uuid.New().String()
				})
			},
			shouldSucceed: true,
			checkFunc: func(t *testing.T, comment *domain.TaskComment) {
				assert.Equal(t, "Reply comment", comment.Content)
				assert.Equal(t, validUserID, comment.User.Id)
			},
		},
		{
			name: "unauthorized error",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: uuid.Nil,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.UnauthorizedErrorCode),
		},
		{
			name: "task not found",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(nil, apperr.NotFoundError("task not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
		},
		{
			name: "project not found",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
		},
		{
			name: "forbidden when user is not a project member",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: uuid.New(),
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
		},
		{
			name: "server error when getting user",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserID).Return(&domain.User{}, errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
		},
		{
			name: "server error when creating comment",
			request: tasks.CreateTaskCommentRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Content:       "Test comment",
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository, userRepo *mockUserRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				userRepo.On("GetById", mock.Anything, validUserID).Return(validUser, nil)
				commentRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskComment"), (*uuid.UUID)(nil)).Return(errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommentRepo := &mockTaskCommentRepository{}
			mockTaskRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			tt.mockSetup(mockCommentRepo, mockTaskRepo, mockProjectRepo, mockUserRepo)

			svc := tasks.NewTaskCommentService(mockCommentRepo, mockTaskRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			comment, err := svc.Create(ctx, tt.request)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, comment)
				if tt.checkFunc != nil {
					tt.checkFunc(t, comment)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, comment)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockCommentRepo.AssertExpectations(t)
			mockTaskRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestTaskCommentService_ListByTaskID(t *testing.T) {
	validUserID := uuid.New()
	validTaskID := uuid.New()
	validProjectID := uuid.New()
	beforeCommentID := uuid.New()
	beforeTime := time.Now()

	validTask := &domain.Task{
		Id:        validTaskID,
		ProjectId: validProjectID,
		Title:     "Test Task",
	}

	validProject := &domain.Project{
		Id: validProjectID,
		Members: []domain.ProjectMember{
			{
				UserId: validUserID,
				Role:   domain.ProjectMemberRoleCreator,
			},
		},
	}

	expectedComments := &utils.CursorPaginated[domain.TaskComment]{
		Data: []domain.TaskComment{
			{
				ID:      uuid.New().String(),
				Content: "Root comment",
				User: &domain.User{
					Id: validUserID,
				},
				Replies: []domain.TaskComment{
					{
						ID:      uuid.New().String(),
						Content: "Reply comment",
						User: &domain.User{
							Id: validUserID,
						},
					},
				},
			},
		},
		HasNext:     true,
		HasPrevious: true,
	}

	type testCase struct {
		name              string
		request           tasks.ListTaskCommentsRequest
		mockSetup         func(*mockTaskCommentRepository, *mockTaskRepository, *mockProjectRepository)
		expectedErrorCode string
		shouldSucceed     bool
		checkFunc         func(t *testing.T, comments *utils.CursorPaginated[domain.TaskComment])
	}

	tests := []testCase{
		{
			name: "successful comment listing",
			request: tasks.ListTaskCommentsRequest{
				TaskID:          validTaskID,
				RequestUserID:   validUserID,
				Limit:           10,
				Before:          &beforeTime,
				BeforeCommentID: &beforeCommentID,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				commentRepo.On("ListByTaskID", mock.Anything, validTaskID, &beforeTime, &beforeCommentID, (*time.Time)(nil), (*uuid.UUID)(nil), 10).Return(expectedComments, nil)
			},
			shouldSucceed: true,
			checkFunc: func(t *testing.T, comments *utils.CursorPaginated[domain.TaskComment]) {
				assert.Len(t, comments.Data, 1)
				assert.True(t, comments.HasNext)
				assert.True(t, comments.HasPrevious)
				assert.Equal(t, "Root comment", comments.Data[0].Content)
				assert.Len(t, comments.Data[0].Replies, 1)
				assert.Equal(t, "Reply comment", comments.Data[0].Replies[0].Content)
			},
		},
		{
			name: "unauthorized error",
			request: tasks.ListTaskCommentsRequest{
				TaskID:        validTaskID,
				RequestUserID: uuid.Nil,
				Limit:         10,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.UnauthorizedErrorCode),
		},
		{
			name: "task not found",
			request: tasks.ListTaskCommentsRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Limit:         10,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(nil, apperr.NotFoundError("task not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
		},
		{
			name: "project not found",
			request: tasks.ListTaskCommentsRequest{
				TaskID:        validTaskID,
				RequestUserID: validUserID,
				Limit:         10,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(nil, apperr.NotFoundError("project not found"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.NotFoundErrorCode),
		},
		{
			name: "forbidden when user is not a project member",
			request: tasks.ListTaskCommentsRequest{
				TaskID:        validTaskID,
				RequestUserID: uuid.New(),
				Limit:         10,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ForbiddenErrorCode),
		},
		{
			name: "server error when listing comments",
			request: tasks.ListTaskCommentsRequest{
				TaskID:          validTaskID,
				RequestUserID:   validUserID,
				Limit:           10,
				Before:          &beforeTime,
				BeforeCommentID: &beforeCommentID,
			},
			mockSetup: func(commentRepo *mockTaskCommentRepository, taskRepo *mockTaskRepository, projectRepo *mockProjectRepository) {
				taskRepo.On("GetById", mock.Anything, validTaskID).Return(validTask, nil)
				projectRepo.On("GetById", mock.Anything, validProjectID).Return(validProject, nil)
				commentRepo.On("ListByTaskID", mock.Anything, validTaskID, &beforeTime, &beforeCommentID, (*time.Time)(nil), (*uuid.UUID)(nil), 10).Return(nil, errors.New("database error"))
			},
			shouldSucceed:     false,
			expectedErrorCode: string(apperr.ServerErrorCode),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommentRepo := &mockTaskCommentRepository{}
			mockTaskRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}
			tt.mockSetup(mockCommentRepo, mockTaskRepo, mockProjectRepo)

			svc := tasks.NewTaskCommentService(mockCommentRepo, mockTaskRepo, mockProjectRepo, mockUserRepo)
			ctx := context.Background()

			comments, err := svc.ListByTaskID(ctx, tt.request)

			if tt.shouldSucceed {
				require.NoError(t, err)
				require.NotNil(t, comments)
				if tt.checkFunc != nil {
					tt.checkFunc(t, comments)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, comments)

				var domainErr apperr.DomainError
				if assert.ErrorAs(t, err, &domainErr) {
					assert.Equal(t, tt.expectedErrorCode, string(domainErr.Code))
				}
			}

			mockCommentRepo.AssertExpectations(t)
			mockTaskRepo.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}
