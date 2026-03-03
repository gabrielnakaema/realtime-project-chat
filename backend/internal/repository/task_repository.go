package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/queries"
	"github.com/gabrielnakaema/project-chat/internal/utils"
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

	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	params := queries.CreateTaskParams{
		ProjectID:   task.ProjectId,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		AuthorID:    task.AuthorId,
		Priority:    string(task.Priority),
		TaskOrder:   int32(task.Order),
	}

	if task.ResponsibleId != nil {
		params.ResponsibleID = pgtype.UUID{
			Bytes: *task.ResponsibleId,
			Valid: true,
		}
	}

	if task.DueDate != nil {
		params.DueDate = pgtype.Timestamptz{
			Time:  *task.DueDate,
			Valid: true,
		}
	}

	id, err := qtx.CreateTask(ctx, params)
	if err != nil {
		return err
	}

	task.Id = id

	for idx, update := range task.Updates {
		params := queries.CreateTaskUpdateParams{
			TaskID:     task.Id,
			UserID:     update.UserId,
			UpdateType: string(update.UpdateType),
		}

		id, err = qtx.CreateTaskUpdate(ctx, params)
		if err != nil {
			return err
		}

		task.Updates[idx].Id = id
	}

	if len(task.Tags) > 0 {
		for _, tag := range task.Tags {
			err = qtx.CreateTaskTag(ctx, queries.CreateTaskTagParams{
				TaskID: task.Id,
				Name:   tag,
			})
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
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
		Priority:    domain.TaskPriority(result.TaskPriority),
		Order:       int(result.TaskOrder),
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

	if result.TaskResponsibleID.Valid {
		task.ResponsibleId = (*uuid.UUID)(result.TaskResponsibleID.Bytes[:])

		task.Responsible = &domain.User{
			Id:        *task.ResponsibleId,
			Name:      result.TaskResponsibleName.String,
			Email:     result.TaskResponsibleEmail.String,
			CreatedAt: result.TaskResponsibleCreatedAt.Time,
		}
	}

	if result.TaskDueDate.Valid {
		task.DueDate = &result.TaskDueDate.Time
	}

	if result.TaskDoneAt.Valid {
		task.DoneAt = &result.TaskDoneAt.Time
	}

	if result.Tags != nil {
		bytes, err := json.Marshal(result.Tags)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &task.Tags)
		if err != nil {
			return nil, err
		}
	}

	if result.Updates != nil {
		bytes, err := json.Marshal(result.Updates)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bytes, &task.Updates)
		if err != nil {
			return nil, err
		}
	}

	return &task, nil
}

func (tr *TaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID, statuses []string, taskOrder int, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	q := queries.New(tr.pool)

	params := queries.ListTasksByProjectIdParams{
		ProjectID: projectId,
		Statuses:  nil,
		TaskOrder: pgtype.Int4{
			Int32: int32(taskOrder),
			Valid: true,
		},
		Limit: int32(limit + 1),
		CursorUpdatedAt: pgtype.Timestamptz{
			Valid: cursorUpdatedAt != nil,
		},
	}

	if cursorUpdatedAt != nil {
		params.CursorUpdatedAt = pgtype.Timestamptz{
			Time:  *cursorUpdatedAt,
			Valid: true,
		}
	} else {
		params.CursorUpdatedAt = pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().Add(1 * time.Hour),
		}
	}

	if len(statuses) > 0 {
		params.Statuses = statuses
	}

	results, err := q.ListTasksByProjectId(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {

		task := domain.Task{
			Id:          result.ID,
			ProjectId:   result.ProjectID,
			AuthorId:    result.AuthorID,
			Title:       result.Title,
			Description: result.Description,
			Status:      domain.TaskStatus(result.Status),
			Priority:    domain.TaskPriority(result.Priority),
			Order:       int(result.TaskOrder),
			CreatedAt:   result.CreatedAt.Time,
			UpdatedAt:   result.UpdatedAt.Time,
		}

		if result.AuthorAuthorID.Valid {
			user := domain.User{
				Id:   result.AuthorID,
				Name: result.AuthorName.String,
			}

			task.Author = &user
		}

		if result.ResponsibleResponsibleID.Valid {
			task.ResponsibleId = (*uuid.UUID)(result.ResponsibleResponsibleID.Bytes[:])
			task.Responsible = &domain.User{
				Id:   *task.ResponsibleId,
				Name: result.ResponsibleName.String,
			}
		}

		if result.Tags != nil {
			bytes, err := json.Marshal(result.Tags)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(bytes, &task.Tags)
			if err != nil {
				return nil, err
			}
		}

		if result.DueDate.Valid {
			task.DueDate = &result.DueDate.Time
		}

		if result.DoneAt.Valid {
			task.DoneAt = &result.DoneAt.Time
		}

		tasks = append(tasks, task)
	}

	minLen := min(len(tasks), int(limit))

	paginated := utils.CursorPaginated[domain.Task]{
		Data:    tasks[:minLen],
		HasNext: len(tasks) > int(limit),
	}

	return &paginated, nil
}

func (tr *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	q := queries.New(tr.pool)

	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	params := queries.UpdateTaskParams{
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		ID:          task.Id,
		TaskOrder:   int32(task.Order),
		Priority:    string(task.Priority),
	}

	if task.ResponsibleId != nil {
		params.ResponsibleID = pgtype.UUID{
			Bytes: *task.ResponsibleId,
			Valid: true,
		}
	}

	if task.DueDate != nil {
		params.DueDate = pgtype.Timestamptz{
			Time:  *task.DueDate,
			Valid: true,
		}
	}

	if task.DoneAt != nil {
		params.DoneAt = pgtype.Timestamptz{
			Time:  *task.DoneAt,
			Valid: true,
		}
	}

	err = qtx.DeleteAllTaskTags(ctx, task.Id)
	if err != nil {
		return err
	}

	if len(task.Tags) > 0 {
		for _, tag := range task.Tags {
			err = qtx.CreateTaskTag(ctx, queries.CreateTaskTagParams{
				TaskID: task.Id,
				Name:   tag,
			})
			if err != nil {
				return err
			}
		}
	}

	err = qtx.UpdateTask(ctx, params)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (tr *TaskRepository) CreateUpdates(ctx context.Context, task *domain.Task, updates []domain.TaskUpdate) error {
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := queries.New(tr.pool)
	qtx := q.WithTx(tx)

	for idx, update := range updates {
		params := queries.CreateTaskUpdateParams{
			TaskID:     task.Id,
			UserID:     update.UserId,
			UpdateType: string(update.UpdateType),
		}

		id, err := qtx.CreateTaskUpdate(ctx, params)
		if err != nil {
			return err
		}

		updates[idx].Id = id

		if len(update.Changes) > 0 {
			for changeIdx, change := range update.Changes {
				params := queries.CreateTaskChangeParams{
					UpdateID: pgtype.UUID{
						Bytes: id,
						Valid: true,
					},
					Field:    change.Field,
					OldValue: change.OldValue,
					NewValue: change.NewValue,
				}

				if change.SubjectId != nil {
					params.SubjectID = pgtype.UUID{
						Bytes: *change.SubjectId,
						Valid: true,
					}
				}

				id, err := qtx.CreateTaskChange(ctx, params)
				if err != nil {
					return err
				}

				updates[idx].Changes[changeIdx].Id = id
			}
		}
	}

	return tx.Commit(ctx)
}

func (tr *TaskRepository) GetSmallestOrderProjectTask(ctx context.Context, projectId uuid.UUID) (*domain.Task, error) {
	q := queries.New(tr.pool)

	result, err := q.GetSmallestOrderProjectTask(ctx, projectId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("no tasks found for project")
		}
		return nil, err
	}

	task := domain.Task{
		Id:          result.ID,
		ProjectId:   result.ProjectID,
		AuthorId:    result.AuthorID,
		Title:       result.Title,
		Description: result.Description,
		Status:      domain.TaskStatus(result.Status),
		Priority:    domain.TaskPriority(result.Priority),
		Order:       int(result.TaskOrder),
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}

	return &task, nil
}

func (tr *TaskRepository) GetProjectTaskAfterId(ctx context.Context, id uuid.UUID, projectId uuid.UUID) (*domain.Task, error) {
	q := queries.New(tr.pool)

	params := queries.GetProjectTaskAfterIdParams{
		ID:        id,
		ProjectID: projectId,
	}

	result, err := q.GetProjectTaskAfterId(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("task not found")
		}
		return nil, err
	}

	task := domain.Task{
		Id:          result.ID,
		ProjectId:   result.ProjectID,
		AuthorId:    result.AuthorID,
		Title:       result.Title,
		Description: result.Description,
		Status:      domain.TaskStatus(result.Status),
		Priority:    domain.TaskPriority(result.Priority),
		Order:       int(result.TaskOrder),
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
	}

	return &task, nil
}

func (tr *TaskRepository) MoveTask(ctx context.Context, task *domain.Task, userId uuid.UUID) (*domain.Task, error) {
	q := queries.New(tr.pool)
	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	params := queries.MoveTaskParams{
		TaskOrder: int32(task.Order),
		Status:    string(task.Status),
		UserID:    userId,
		ID:        task.Id,
	}

	result, err := qtx.MoveTask(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("task not found")
		}
		return nil, err
	}

	task.Id = result.ID
	task.Order = int(result.TaskOrder)
	task.Status = domain.TaskStatus(result.Status)

	task, err = tr.GetById(ctx, result.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("task not found")
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (tr *TaskRepository) NormalizeProjectTaskOrders(ctx context.Context, projectId uuid.UUID) error {
	const multiplier = 1000

	query := `
		UPDATE tasks t
			set task_order = (t2.seqnum + 1) * $2
			from (select t2.id, row_number() over(ORDER BY t2.task_order ASC, t2.updated_at DESC) as seqnum from tasks t2 where t2.project_id = $1) t2
		where t.id = t2.id
	`

	_, err := tr.pool.Exec(ctx, query, projectId, multiplier)
	if err != nil {
		return err
	}

	return nil
}

func (tr *TaskRepository) CountTasksByProjectIdAndStatus(ctx context.Context, projectId uuid.UUID, statuses []string) (map[string]int, error) {
	q := queries.New(tr.pool)

	params := queries.CountTasksByProjectIdAndStatusParams{
		ProjectID: projectId,
		Column2:   statuses,
	}

	results, err := q.CountTasksByProjectIdAndStatus(ctx, params)
	if err != nil {
		return nil, err
	}

	result := map[string]int{}
	for _, r := range results {
		result[r.Status] = int(r.Count)
	}

	return result, nil
}

func (tr *TaskRepository) ListUserDueTasks(ctx context.Context, userId uuid.UUID, statuses []string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	q := queries.New(tr.pool)

	params := queries.ListUserDueTasksParams{
		ResponsibleID: pgtype.UUID{
			Bytes: userId,
			Valid: true,
		},
		Statuses: statuses,
		CursorDueDate: pgtype.Timestamptz{
			Valid: cursorDueDate != nil,
		},
		CursorUpdatedAt: pgtype.Timestamptz{
			Valid: cursorUpdatedAt != nil,
		},
		Limit: int32(limit + 1),
	}

	if cursorUpdatedAt != nil {
		params.CursorUpdatedAt = pgtype.Timestamptz{
			Time:  *cursorUpdatedAt,
			Valid: true,
		}
	} else {
		params.CursorUpdatedAt = pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().Add(1 * time.Hour),
		}
	}

	if cursorDueDate != nil {
		params.CursorDueDate = pgtype.Timestamptz{
			Time:  *cursorDueDate,
			Valid: true,
		}
	}

	if len(statuses) > 0 {
		params.Statuses = statuses
	}

	results, err := q.ListUserDueTasks(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {
		task := domain.Task{
			Id:            result.ID,
			ProjectId:     result.ProjectID,
			ResponsibleId: (*uuid.UUID)(result.ResponsibleResponsibleID.Bytes[:]),
			Title:         result.Title,
			Description:   result.Description,
			Status:        domain.TaskStatus(result.Status),
			Priority:      domain.TaskPriority(result.Priority),
			Order:         int(result.TaskOrder),
			CreatedAt:     result.CreatedAt.Time,
			UpdatedAt:     result.UpdatedAt.Time,
			Project: &domain.Project{
				Id:          result.ProjectProjectID,
				Name:        result.ProjectName,
				Description: result.ProjectDescription,
				CreatedAt:   result.ProjectCreatedAt.Time,
				UpdatedAt:   result.ProjectUpdatedAt.Time,
				UserId:      result.ProjectUserID,
			},
		}

		if result.Tags != nil {
			bytes, err := json.Marshal(result.Tags)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(bytes, &task.Tags)
			if err != nil {
				return nil, err
			}
		}

		if result.DueDate.Valid {
			task.DueDate = &result.DueDate.Time
		}

		if result.DoneAt.Valid {
			task.DoneAt = &result.DoneAt.Time
		}

		tasks = append(tasks, task)
	}

	minLen := min(len(tasks), int(limit))

	paginated := utils.CursorPaginated[domain.Task]{
		Data:    tasks[:minLen],
		HasNext: len(tasks) > int(limit),
	}

	return &paginated, nil
}
