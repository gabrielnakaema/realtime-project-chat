package tasks_test

import (
	"context"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTaskService_SearchTasks(t *testing.T) {
	creatorID := uuid.New()
	memberID := uuid.New()
	projectID := uuid.New()
	pendingColumnID := uuid.New()
	doneColumnID := uuid.New()
	project := domain.Project{
		Id: projectID,
		Members: []domain.ProjectMember{
			{UserId: creatorID, Role: domain.ProjectMemberRoleCreator},
			{UserId: memberID, Role: domain.ProjectMemberRoleMember},
		},
		Columns: []domain.ProjectColumn{
			{Id: pendingColumnID, ProjectId: projectID, Name: "Pending"},
			{Id: doneColumnID, ProjectId: projectID, Name: "Done", IsDoneColumn: true},
		},
	}
	page := &utils.CursorPaginated[domain.Task]{Data: []domain.Task{{Id: uuid.New(), Title: "Search result"}}}

	t.Run("global search stays unscoped and trims query", func(t *testing.T) {
		repo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		repo.On("SearchTasks", mock.Anything, mock.MatchedBy(func(request tasks.SearchTasksRequest) bool {
			return request.UserId == memberID && request.ProjectId == nil && len(request.ProjectColumnIDs) == 0 &&
				request.SearchQuery == "needle" && !request.IncludeArchived && !request.IncludeDone && request.Limit == 15
		})).Return(page, nil)

		svc := tasks.NewTaskService(repo, projectRepo, &mockUserRepository{})
		result, err := svc.SearchTasks(context.Background(), tasks.SearchTasksRequest{
			UserId:      memberID,
			SearchQuery: "  needle  ",
			Limit:       15,
		})

		require.NoError(t, err)
		assert.Equal(t, page, result)
		repo.AssertExpectations(t)
	})

	t.Run("project search validates membership and selected columns", func(t *testing.T) {
		repo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		projectRepo.On("GetById", mock.Anything, projectID).Return(&project, nil)
		repo.On("SearchTasks", mock.Anything, mock.MatchedBy(func(request tasks.SearchTasksRequest) bool {
			return request.ProjectId != nil && *request.ProjectId == projectID &&
				assert.ObjectsAreEqual([]uuid.UUID{pendingColumnID, doneColumnID}, request.ProjectColumnIDs) && request.IncludeDone
		})).Return(page, nil)

		svc := tasks.NewTaskService(repo, projectRepo, &mockUserRepository{})
		result, err := svc.SearchTasks(context.Background(), tasks.SearchTasksRequest{
			UserId:           memberID,
			ProjectId:        &projectID,
			ProjectColumnIDs: []uuid.UUID{pendingColumnID, doneColumnID},
			SearchQuery:      "task",
			IncludeDone:      true,
			Limit:            25,
		})

		require.NoError(t, err)
		assert.Equal(t, page, result)
		repo.AssertExpectations(t)
		projectRepo.AssertExpectations(t)
	})

	t.Run("regular member cannot include archived tasks", func(t *testing.T) {
		repo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		projectRepo.On("GetById", mock.Anything, projectID).Return(&project, nil)
		svc := tasks.NewTaskService(repo, projectRepo, &mockUserRepository{})

		result, err := svc.SearchTasks(context.Background(), tasks.SearchTasksRequest{
			UserId:           memberID,
			ProjectId:        &projectID,
			ProjectColumnIDs: []uuid.UUID{pendingColumnID},
			SearchQuery:      "task",
			IncludeArchived:  true,
			Limit:            25,
		})

		require.Error(t, err)
		assert.Nil(t, result)
		var domainErr apperr.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, apperr.ForbiddenErrorCode, domainErr.Code)
	})

	t.Run("creator can include archived tasks", func(t *testing.T) {
		repo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		projectRepo.On("GetById", mock.Anything, projectID).Return(&project, nil)
		repo.On("SearchTasks", mock.Anything, mock.MatchedBy(func(request tasks.SearchTasksRequest) bool {
			return request.IncludeArchived
		})).Return(page, nil)
		svc := tasks.NewTaskService(repo, projectRepo, &mockUserRepository{})

		result, err := svc.SearchTasks(context.Background(), tasks.SearchTasksRequest{
			UserId:           creatorID,
			ProjectId:        &projectID,
			ProjectColumnIDs: []uuid.UUID{pendingColumnID},
			SearchQuery:      "task",
			IncludeArchived:  true,
			Limit:            25,
		})

		require.NoError(t, err)
		assert.Equal(t, page, result)
	})

	t.Run("rejects a column from another project", func(t *testing.T) {
		repo := &mockTaskRepository{}
		projectRepo := &mockProjectRepository{}
		projectRepo.On("GetById", mock.Anything, projectID).Return(&project, nil)
		svc := tasks.NewTaskService(repo, projectRepo, &mockUserRepository{})

		result, err := svc.SearchTasks(context.Background(), tasks.SearchTasksRequest{
			UserId:           memberID,
			ProjectId:        &projectID,
			ProjectColumnIDs: []uuid.UUID{uuid.New()},
			SearchQuery:      "task",
			Limit:            25,
		})

		require.Error(t, err)
		assert.Nil(t, result)
		var domainErr apperr.DomainError
		require.ErrorAs(t, err, &domainErr)
		assert.Equal(t, apperr.BusinessValidationErrorCode, domainErr.Code)
	})
}
