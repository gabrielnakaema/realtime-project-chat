package service_test

import (
	"context"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/fracindex"
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

	taskA := domain.Task{Id: taskIdA, ProjectId: validProjectId, Order: "500000000000"}
	taskB := domain.Task{Id: taskIdB, ProjectId: validProjectId, Order: "750000000000"}

	topOrder := mustMoveOrder(t, "", "")
	beforeFirstOrder := mustMoveOrder(t, "", taskA.Order)
	betweenOrder := mustMoveOrder(t, taskA.Order, taskB.Order)
	afterLastOrder := mustMoveOrder(t, taskB.Order, "")

	type testCase struct {
		name          string
		request       service.MoveTaskRequest
		mockSetup     func(*mockTaskRepository, *mockProjectRepository)
		expectedOrder string
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
				repo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, domain.TaskStatusPending).Return(nil, domain.NotFoundError("not found"))
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == topOrder
				}), validUserId).Return(&domain.Task{Order: topOrder}, nil)
			},
			expectedOrder: topOrder,
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
				repo.On("GetFirstTaskInColumn", mock.Anything, validProjectId, domain.TaskStatusPending).Return(&taskA, nil)
				repo.On("MoveTask", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
					return t.Order == beforeFirstOrder
				}), validUserId).Return(&domain.Task{Order: beforeFirstOrder}, nil)
			},
			expectedOrder: beforeFirstOrder,
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
					return t.Order == betweenOrder
				}), validUserId).Return(&domain.Task{Order: betweenOrder}, nil)
			},
			expectedOrder: betweenOrder,
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
					return t.Order == afterLastOrder
				}), validUserId).Return(&domain.Task{Order: afterLastOrder}, nil)
			},
			expectedOrder: afterLastOrder,
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

func mustMoveOrder(t *testing.T, left, right string) string {
	t.Helper()

	key, err := fracindex.GenerateKeyBetween(left, right)
	assert.NoError(t, err)

	return key
}
