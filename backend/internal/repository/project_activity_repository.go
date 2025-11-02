package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/gabrielnakaema/project-chat/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectActivityRepository struct {
	pool *pgxpool.Pool
}

func NewProjectActivityRepository(pool *pgxpool.Pool) *ProjectActivityRepository {
	return &ProjectActivityRepository{
		pool: pool,
	}
}

func (par *ProjectActivityRepository) Create(ctx context.Context, activity *domain.ProjectActivity) error {
	bytes, err := json.Marshal(activity.ActivityData)
	if err != nil {
		return err
	}

	q := queries.New(par.pool)
	params := queries.CreateProjectActivityLogParams{
		ProjectID:    activity.Project.Id,
		ActorID:      activity.Actor.Id,
		ActivityType: string(activity.ActivityType),
		ActivityData: bytes,
		EntityType:   pgtype.Text{String: string(activity.EntityType), Valid: true},
		EntityID:     pgtype.UUID{Bytes: activity.EntityId, Valid: true},
	}

	result, err := q.CreateProjectActivityLog(ctx, params)
	if err != nil {
		return err
	}

	activity.Id = result

	return nil
}

type ListProjectActivityLogsParams struct {
	ProjectId       uuid.UUID
	UserId          uuid.UUID
	BeforeCreatedAt time.Time
	BeforeId        uuid.UUID
	Limit           int32
}

func (par *ProjectActivityRepository) List(ctx context.Context, params ListProjectActivityLogsParams) (*utils.CursorPaginated[domain.ProjectActivity], error) {
	q := queries.New(par.pool)

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	filterByUser := params.UserId != uuid.Nil

	results, err := q.ListProjectActivityLogs(ctx, queries.ListProjectActivityLogsParams{
		ProjectID: pgtype.UUID{
			Bytes: params.ProjectId,
			Valid: params.ProjectId != uuid.Nil,
		},
		UserID: pgtype.UUID{
			Bytes: params.UserId,
			Valid: params.UserId != uuid.Nil,
		},
		BeforeCreatedAt: pgtype.Timestamptz{
			Time:  params.BeforeCreatedAt,
			Valid: params.BeforeCreatedAt != time.Time{},
		},
		BeforeID: pgtype.UUID{
			Bytes: params.BeforeId,
			Valid: params.BeforeId != uuid.Nil,
		},
		Limit: pgtype.Int4{
			Int32: params.Limit,
			Valid: params.Limit > 0,
		},
		FilterByUser: filterByUser,
	})
	if err != nil {
		return nil, err
	}

	activities := make([]domain.ProjectActivity, len(results))
	for i, result := range results {
		activity := domain.ProjectActivity{
			Id: result.ID,
			Project: domain.Project{
				Id:          result.ProjectID,
				Name:        result.ProjectName,
				Description: result.ProjectDescription,
				CreatedAt:   result.ProjectCreatedAt.Time,
				UpdatedAt:   result.ProjectUpdatedAt.Time,
				UserId:      result.ProjectUserID,
			},
			Actor: domain.User{
				Id:    result.ActorID,
				Name:  result.ActorName,
				Email: result.ActorEmail,
			},
			ActivityType: domain.ActivityType(result.ActivityType),
			EntityType:   domain.EntityType(result.EntityType.String),
			EntityId:     result.EntityID.Bytes,
			CreatedAt:    result.CreatedAt.Time,
			UpdatedAt:    result.UpdatedAt.Time,
			ActivityData: result.ActivityData,
		}

		if activity.EntityType == domain.TaskEntityType {
			var task domain.Task
			err := json.Unmarshal(result.Entity, &task)
			if err != nil {
				return nil, domain.ServerError("failed to unmarshal task", err)
			}
			activity.Task = &task
		}

		if activity.EntityType == domain.ProjectMemberEntityType {
			var projectMember domain.ProjectMember
			err := json.Unmarshal(result.Entity, &projectMember)
			if err != nil {
				return nil, domain.ServerError("failed to unmarshal project member", err)
			}
			activity.ProjectMember = &projectMember
		}

		activities[i] = activity
	}

	hasNext := len(activities) > int(limit)
	data := activities
	if hasNext {
		data = activities[:limit]
	}

	paginated := utils.CursorPaginated[domain.ProjectActivity]{
		Data:    data,
		HasNext: hasNext,
	}

	return &paginated, nil
}
