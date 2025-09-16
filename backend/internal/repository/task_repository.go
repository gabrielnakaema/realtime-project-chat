package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		pool: pool,
	}
}

func (tr *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	q := queries.New(tr.pool)
	params := queries.CreateTaskParams{
		ProjectID:   task.ProjectId,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AuthorID:    task.AuthorId,
	}

	id, err := q.CreateTask(ctx, params)
	if err != nil {
		return err
	}

	task.Id = id

	return nil
}

func (tr *TaskRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	q := queries.New(tr.pool)

	result, err := q.GetTaskById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("task not found")
		}
		return nil, err
	}

	task := domain.Task{
		Id:          result.TaskID,
		ProjectId:   result.TaskProjectID,
		AuthorId:    result.TaskAuthorID,
		Title:       result.TaskTitle,
		Description: result.TaskDescription,
		Status:      domain.TaskStatus(result.TaskStatus),
		CreatedAt:   result.TaskCreatedAt.Time,
		UpdatedAt:   result.TaskUpdatedAt.Time,
	}

	if result.TaskAuthorName.Valid {
		task.Author = &domain.User{
			Id:        result.TaskAuthorID,
			Name:      result.TaskAuthorName.String,
			Email:     result.TaskAuthorEmail.String,
			CreatedAt: result.TaskAuthorCreatedAt.Time,
		}
	}

	if result.TaskChanges != nil {
		bytes, err := json.Marshal(result.TaskChanges)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &task.Changes)
		if err != nil {
			return nil, err
		}
	}

	return &task, nil
}

func (tr *TaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.Task, error) {
	q := queries.New(tr.pool)

	results, err := q.ListTasksByProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {
		tasks = append(tasks, domain.Task{
			Id:          result.ID,
			ProjectId:   result.ProjectID,
			AuthorId:    result.AuthorID,
			Title:       result.Title,
			Description: result.Description,
			Status:      domain.TaskStatus(result.Status),
			CreatedAt:   result.CreatedAt.Time,
			UpdatedAt:   result.UpdatedAt.Time,
		})
	}

	return tasks, nil
}

func (tr *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	q := queries.New(tr.pool)

	params := queries.UpdateTaskParams{
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		ID:          task.Id,
	}

	return q.UpdateTask(ctx, params)
}

func (tr *TaskRepository) CreateChanges(ctx context.Context, task *domain.Task, changes []domain.TaskChange) error {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(tr.pool)
	qtx := q.WithTx(tx)

	for i, change := range changes {
		params := queries.CreateTaskChangeParams{
			TaskID:      task.Id,
			UserID:      pgtype.UUID{Bytes: change.AuthorId, Valid: true},
			Description: change.ChangeDescription,
		}

		id, err := qtx.CreateTaskChange(ctx, params)
		if err != nil {
			return err
		}

		changes[i].Id = id
	}

	return tx.Commit(ctx)
}
