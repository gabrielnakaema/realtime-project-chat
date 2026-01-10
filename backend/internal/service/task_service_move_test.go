package service_test

import (
	"context"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskService_Move(t *testing.T) {
	validUserId := uuid.New()
	validProjectId := uuid.New()
	validTaskId := uuid.New()
	taskIdA := uuid.New()
	taskIdB := uuid.New()

	taskA := domain.Task{Id: taskIdA, ProjectId: validProjectId, Order: 1000}
	taskB := domain.Task{Id: taskIdB, ProjectId: validProjectId, Order: 2000}

	type testCase struct {
		name          string
		request       service.MoveTaskRequest
		mockSetup     func(*mockTaskRepository, *mockProjectRepository)
		expectedOrder int
		shouldSucceed bool
	}

	tests := []testCase{
		{
			name: "move to top (no existing tasks)",
			request: service.MoveTaskRequest{
				TaskId:        validTaskId,
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				AfterTaskId:   nil,
				Status:        domain.TaskStatusPending,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&domain.Task{Id: validTaskId, ProjectId: validProjectId, Status: domain.TaskStatusPending}, nil)
				repo.On("GetSmallestOrderProjectTask", mock.Anything, validProjectId).Return(nil, domain.NotFoundError("not found"))
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == 1000
				}), validUserId).Return(&domain.Task{Order: 1000}, nil)
			},
			expectedOrder: 1000,
			shouldSucceed: true,
		},
		{
			name: "move to top (existing tasks)",
			request: service.MoveTaskRequest{
				TaskId:        validTaskId,
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				AfterTaskId:   nil,
				Status:        domain.TaskStatusPending,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&domain.Task{Id: validTaskId, ProjectId: validProjectId, Status: domain.TaskStatusPending}, nil)
				repo.On("GetSmallestOrderProjectTask", mock.Anything, validProjectId).Return(&taskA, nil)
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == 500
				}), validUserId).Return(&domain.Task{Order: 500}, nil)
			},
			expectedOrder: 500,
			shouldSucceed: true,
		},
		{
			name: "move after task A (between A and B)",
			request: service.MoveTaskRequest{
				TaskId:        validTaskId,
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				AfterTaskId:   &taskIdA,
				Status:        domain.TaskStatusPending,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&domain.Task{Id: validTaskId, ProjectId: validProjectId, Status: domain.TaskStatusPending}, nil)
				repo.On("GetById", mock.Anything, taskIdA).Return(&taskA, nil)
				repo.On("GetProjectTaskAfterId", mock.Anything, taskIdA, validProjectId).Return(&taskB, nil)
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == 1500
				}), validUserId).Return(&domain.Task{Order: 1500}, nil)
			},
			expectedOrder: 1500,
			shouldSucceed: true,
		},
		{
			name: "move after task B (end of list)",
			request: service.MoveTaskRequest{
				TaskId:        validTaskId,
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				AfterTaskId:   &taskIdB,
				Status:        domain.TaskStatusPending,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&domain.Task{Id: validTaskId, ProjectId: validProjectId, Status: domain.TaskStatusPending}, nil)
				repo.On("GetById", mock.Anything, taskIdB).Return(&taskB, nil)
				repo.On("GetProjectTaskAfterId", mock.Anything, taskIdB, validProjectId).Return(nil, domain.NotFoundError("not found"))
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == 3000
				}), validUserId).Return(&domain.Task{Order: 3000}, nil)
			},
			expectedOrder: 3000,
			shouldSucceed: true,
		},
		{
			name: "normalization triggered",
			request: service.MoveTaskRequest{
				TaskId:        validTaskId,
				ProjectId:     validProjectId,
				RequestUserId: validUserId,
				AfterTaskId:   &taskIdA,
				Status:        domain.TaskStatusPending,
			},
			mockSetup: func(repo *mockTaskRepository, projectRepo *mockProjectRepository) {
				repo.On("GetById", mock.Anything, validTaskId).Return(&domain.Task{Id: validTaskId, ProjectId: validProjectId, Status: domain.TaskStatusPending}, nil)
				repo.On("GetById", mock.Anything, taskIdA).Return(&domain.Task{Id: taskIdA, Order: 1000}, nil).Once()
				repo.On("GetProjectTaskAfterId", mock.Anything, taskIdA, validProjectId).Return(&domain.Task{Id: taskIdB, Order: 1001}, nil).Once()

				repo.On("NormalizeProjectTaskOrders", mock.Anything, validProjectId).Return(nil)

				repo.On("GetById", mock.Anything, taskIdA).Return(&domain.Task{Id: taskIdA, Order: 1000}, nil).Once()
				repo.On("GetProjectTaskAfterId", mock.Anything, taskIdA, validProjectId).Return(&domain.Task{Id: taskIdB, Order: 2000}, nil).Once()

				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == 1500
				}), validUserId).Return(&domain.Task{Order: 1500}, nil)
			},
			expectedOrder: 1500,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockTaskRepository{}
			mockProjectRepo := &mockProjectRepository{}
			mockUserRepo := &mockUserRepository{}

			mockProjectRepo.On("GetById", mock.Anything, validProjectId).Return(&domain.Project{
				Id:      validProjectId,
				Members: []domain.ProjectMember{{UserId: validUserId}},
			}, nil)

			tt.mockSetup(mockRepo, mockProjectRepo)
			service := service.NewTaskService(mockRepo, mockProjectRepo, mockUserRepo, &mockPublisher{})
			ctx := context.Background()

			task, err := service.Move(ctx, tt.request)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, tt.expectedOrder, task.Order)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
