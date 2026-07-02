package repository

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
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

func mapProjectColumn(columnID uuid.UUID, projectID uuid.UUID, name string, color string, position int32, isDone bool, createdAt pgtype.Timestamptz, updatedAt pgtype.Timestamptz) *domain.ProjectColumn {
	return &domain.ProjectColumn{
		Id:           columnID,
		ProjectId:    projectID,
		Name:         name,
		Color:        color,
		Position:     int(position),
		IsDoneColumn: isDone,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}
}

func compatibilityTaskStatus(projectColumnName string, archivedAt *time.Time) domain.TaskStatus {
	if archivedAt != nil {
		return domain.TaskStatusArchived
	}

	return domain.TaskStatus(strings.ToLower(projectColumnName))
}

func (tr *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	q := queries.New(tr.pool)
	actionOrigin := domain.ActionOriginFromContext(ctx)

	tx, err := tr.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	params := queries.CreateTaskParams{
		ProjectID:       task.ProjectId,
		Title:           task.Title,
		Description:     task.Description,
		Code:            pgtype.Text{},
		ProjectColumnID: task.ProjectColumnId,
		AuthorID:        task.AuthorId,
		Priority:        string(task.Priority),
		TaskOrder:       task.Order,
	}

	if task.Code != "" {
		params.Code = pgtype.Text{
			String: task.Code,
			Valid:  true,
		}
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

	id, err := qtx.CreateTask(ctx, params)
	if err != nil {
		return err
	}

	task.Id = id

	for idx, update := range task.Updates {
		params := queries.CreateTaskUpdateParams{
			TaskID:       task.Id,
			UserID:       update.UserId,
			UpdateType:   string(update.UpdateType),
			ActionOrigin: string(actionOrigin.OrUser()),
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

	if len(task.DependsOnTaskIds) > 0 {
		for _, dependsOnTaskID := range task.DependsOnTaskIds {
			err = qtx.CreateTaskDependency(ctx, queries.CreateTaskDependencyParams{
				TaskID:          task.Id,
				DependsOnTaskID: dependsOnTaskID,
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
		Id:              result.TaskID,
		ProjectId:       result.TaskProjectID,
		AuthorId:        result.TaskAuthorID,
		Title:           result.TaskTitle,
		Description:     result.TaskDescription,
		Code:            "",
		ProjectColumnId: result.TaskProjectColumnID,
		Priority:        domain.TaskPriority(result.TaskPriority),
		Order:           result.TaskOrder,
		CreatedAt:       result.TaskCreatedAt.Time,
		UpdatedAt:       result.TaskUpdatedAt.Time,
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

	if result.TaskCode.Valid {
		task.Code = result.TaskCode.String
	}

	if result.TaskDoneAt.Valid {
		task.DoneAt = &result.TaskDoneAt.Time
	}

	if result.TaskArchivedAt.Valid {
		task.ArchivedAt = &result.TaskArchivedAt.Time
	}

	if len(result.ProjectColumn) > 0 {
		task.ProjectColumn = &domain.ProjectColumn{}
		if err := json.Unmarshal(result.ProjectColumn, task.ProjectColumn); err != nil {
			return nil, err
		}
		task.Status = compatibilityTaskStatus(task.ProjectColumn.Name, task.ArchivedAt)
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

	if err := mapDependsOnTasksFromResult(result.DependsOnTasks, &task); err != nil {
		return nil, err
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

func (tr *TaskRepository) ListByProjectId(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID, archived bool, taskOrder string, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	q := queries.New(tr.pool)

	params := queries.ListTasksByProjectIdParams{
		ProjectID: projectId,
		Archived:  archived,
		TaskOrder: pgtype.Text{
			String: taskOrder,
			Valid:  taskOrder != "",
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
	}

	if len(projectColumnIDs) > 0 {
		params.ProjectColumnIds = projectColumnIDs
	}

	results, err := q.ListTasksByProjectId(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {

		task := domain.Task{
			Id:              result.ID,
			ProjectId:       result.ProjectID,
			AuthorId:        result.AuthorID,
			Title:           result.Title,
			Description:     result.Description,
			Code:            "",
			ProjectColumnId: result.ProjectColumnID,
			Priority:        domain.TaskPriority(result.Priority),
			Order:           result.TaskOrder,
			CreatedAt:       result.CreatedAt.Time,
			UpdatedAt:       result.UpdatedAt.Time,
			ProjectColumn: mapProjectColumn(
				result.ProjectColumnID2,
				result.ProjectColumnProjectID,
				result.ProjectColumnName,
				result.ProjectColumnColor,
				result.ProjectColumnPosition,
				result.ProjectColumnIsDoneColumn,
				result.ProjectColumnCreatedAt,
				result.ProjectColumnUpdatedAt,
			),
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

		if err := mapDependsOnTaskIdsFromResult(result.DependsOnTaskIds, &task); err != nil {
			return nil, err
		}

		if result.DueDate.Valid {
			task.DueDate = &result.DueDate.Time
		}

		if result.Code.Valid {
			task.Code = result.Code.String
		}

		if result.DoneAt.Valid {
			task.DoneAt = &result.DoneAt.Time
		}

		if result.ArchivedAt.Valid {
			task.ArchivedAt = &result.ArchivedAt.Time
		}

		task.Status = compatibilityTaskStatus(task.ProjectColumn.Name, task.ArchivedAt)

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
		Title:           task.Title,
		Description:     task.Description,
		Code:            pgtype.Text{},
		ProjectColumnID: task.ProjectColumnId,
		ID:              task.Id,
		TaskOrder:       task.Order,
		Priority:        string(task.Priority),
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

	if task.ArchivedAt != nil {
		params.ArchivedAt = pgtype.Timestamptz{
			Time:  *task.ArchivedAt,
			Valid: true,
		}
	}

	if task.Code != "" {
		params.Code = pgtype.Text{
			String: task.Code,
			Valid:  true,
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

	err = qtx.DeleteAllTaskDependencies(ctx, task.Id)
	if err != nil {
		return err
	}

	if len(task.DependsOnTaskIds) > 0 {
		for _, dependsOnTaskID := range task.DependsOnTaskIds {
			err = qtx.CreateTaskDependency(ctx, queries.CreateTaskDependencyParams{
				TaskID:          task.Id,
				DependsOnTaskID: dependsOnTaskID,
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
	actionOrigin := domain.ActionOriginFromContext(ctx)

	for idx, update := range updates {
		params := queries.CreateTaskUpdateParams{
			TaskID:       task.Id,
			UserID:       update.UserId,
			UpdateType:   string(update.UpdateType),
			ActionOrigin: string(actionOrigin.OrUser()),
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

				if change.OldValueId != nil {
					params.OldValueID = pgtype.UUID{
						Bytes: *change.OldValueId,
						Valid: true,
					}
				}

				if change.NewValueId != nil {
					params.NewValueID = pgtype.UUID{
						Bytes: *change.NewValueId,
						Valid: true,
					}
				}

				if change.OldDisplayValue != nil {
					params.OldDisplayValue = pgtype.Text{
						String: *change.OldDisplayValue,
						Valid:  true,
					}
				}

				if change.NewDisplayValue != nil {
					params.NewDisplayValue = pgtype.Text{
						String: *change.NewDisplayValue,
						Valid:  true,
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

func (tr *TaskRepository) GetFirstTaskInColumn(ctx context.Context, projectId uuid.UUID, projectColumnID uuid.UUID) (*domain.Task, error) {
	q := queries.New(tr.pool)

	result, err := q.GetFirstTaskInColumn(ctx, queries.GetFirstTaskInColumnParams{
		ProjectID:       projectId,
		ProjectColumnID: projectColumnID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("no tasks found for column")
		}
		return nil, err
	}

	task := domain.Task{
		Id:              result.ID,
		ProjectId:       result.ProjectID,
		AuthorId:        result.AuthorID,
		Title:           result.Title,
		Description:     result.Description,
		Code:            "",
		ProjectColumnId: result.ProjectColumnID,
		Priority:        domain.TaskPriority(result.Priority),
		Order:           result.TaskOrder,
		CreatedAt:       result.CreatedAt.Time,
		UpdatedAt:       result.UpdatedAt.Time,
	}
	task.Status = compatibilityTaskStatus("", nil)

	if result.Code.Valid {
		task.Code = result.Code.String
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
		Id:              result.ID,
		ProjectId:       result.ProjectID,
		AuthorId:        result.AuthorID,
		Title:           result.Title,
		Description:     result.Description,
		Code:            "",
		ProjectColumnId: result.ProjectColumnID,
		Priority:        domain.TaskPriority(result.Priority),
		Order:           result.TaskOrder,
		CreatedAt:       result.CreatedAt.Time,
		UpdatedAt:       result.UpdatedAt.Time,
	}
	task.Status = compatibilityTaskStatus("", nil)

	if result.Code.Valid {
		task.Code = result.Code.String
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
		TaskOrder:       task.Order,
		ProjectColumnID: task.ProjectColumnId,
		UserID:          userId,
		ID:              task.Id,
	}

	if task.DoneAt != nil {
		params.DoneAt = pgtype.Timestamptz{
			Time:  *task.DoneAt,
			Valid: true,
		}
	}

	result, err := qtx.MoveTask(ctx, params)
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

	task, err = tr.GetById(ctx, result.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NotFoundError("task not found")
		}
		return nil, err
	}

	return task, nil
}

func (tr *TaskRepository) CountTasksByProjectIdAndColumn(ctx context.Context, projectId uuid.UUID, projectColumnIDs []uuid.UUID) (map[string]int, error) {
	q := queries.New(tr.pool)

	params := queries.CountTasksByProjectIdAndColumnParams{
		ProjectID: projectId,
		Column2:   projectColumnIDs,
	}

	results, err := q.CountTasksByProjectIdAndColumn(ctx, params)
	if err != nil {
		return nil, err
	}

	result := map[string]int{}
	for _, r := range results {
		result[r.ProjectColumnID.String()] = int(r.Count)
	}

	return result, nil
}

func (tr *TaskRepository) ListUserDueTasks(ctx context.Context, userId uuid.UUID, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	q := queries.New(tr.pool)

	params := queries.ListUserDueTasksParams{
		ResponsibleID: pgtype.UUID{
			Bytes: userId,
			Valid: true,
		},
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

	results, err := q.ListUserDueTasks(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {
		task := domain.Task{
			Id:              result.ID,
			ProjectId:       result.ProjectID,
			ResponsibleId:   (*uuid.UUID)(result.ResponsibleResponsibleID.Bytes[:]),
			Title:           result.Title,
			Description:     result.Description,
			Code:            "",
			ProjectColumnId: result.ProjectColumnID,
			Priority:        domain.TaskPriority(result.Priority),
			Order:           result.TaskOrder,
			CreatedAt:       result.CreatedAt.Time,
			UpdatedAt:       result.UpdatedAt.Time,
			ProjectColumn: mapProjectColumn(
				result.ProjectColumnID2,
				result.ProjectColumnProjectID,
				result.ProjectColumnName,
				result.ProjectColumnColor,
				result.ProjectColumnPosition,
				result.ProjectColumnIsDoneColumn,
				result.ProjectColumnCreatedAt,
				result.ProjectColumnUpdatedAt,
			),
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

		if err := mapDependsOnTaskIdsFromResult(result.DependsOnTaskIds, &task); err != nil {
			return nil, err
		}

		if result.DueDate.Valid {
			task.DueDate = &result.DueDate.Time
		}

		if result.Code.Valid {
			task.Code = result.Code.String
		}

		if result.DoneAt.Valid {
			task.DoneAt = &result.DoneAt.Time
		}

		if result.ArchivedAt.Valid {
			task.ArchivedAt = &result.ArchivedAt.Time
		}

		task.Status = compatibilityTaskStatus(task.ProjectColumn.Name, task.ArchivedAt)

		tasks = append(tasks, task)
	}

	minLen := min(len(tasks), int(limit))

	paginated := utils.CursorPaginated[domain.Task]{
		Data:    tasks[:minLen],
		HasNext: len(tasks) > int(limit),
	}

	return &paginated, nil
}

func (tr *TaskRepository) SearchTasksForUser(ctx context.Context, userId uuid.UUID, searchQuery string, cursorDueDate *time.Time, cursorUpdatedAt *time.Time, limit int) (*utils.CursorPaginated[domain.Task], error) {
	q := queries.New(tr.pool)

	params := queries.SearchTasksForUserParams{
		UserID:          userId,
		Limit:           int32(limit + 1),
		Query:           pgtype.Text{String: searchQuery, Valid: true},
		CursorDueDate:   pgtype.Timestamptz{Valid: cursorDueDate != nil},
		CursorUpdatedAt: pgtype.Timestamptz{Valid: cursorUpdatedAt != nil},
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

	results, err := q.SearchTasksForUser(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := []domain.Task{}
	for _, result := range results {
		task := domain.Task{
			Id:              result.ID,
			ProjectId:       result.ProjectID,
			Title:           result.Title,
			Description:     result.Description,
			Code:            "",
			ProjectColumnId: result.ProjectColumnID,
			Priority:        domain.TaskPriority(result.Priority),
			Order:           result.TaskOrder,
			CreatedAt:       result.CreatedAt.Time,
			UpdatedAt:       result.UpdatedAt.Time,
			AuthorId:        result.AuthorID,
			ResponsibleId:   (*uuid.UUID)(result.ResponsibleID.Bytes[:]),
			ProjectColumn: mapProjectColumn(
				result.ProjectColumnID2,
				result.ProjectColumnProjectID,
				result.ProjectColumnName,
				result.ProjectColumnColor,
				result.ProjectColumnPosition,
				result.ProjectColumnIsDoneColumn,
				result.ProjectColumnCreatedAt,
				result.ProjectColumnUpdatedAt,
			),
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

		if err := mapDependsOnTaskIdsFromResult(result.DependsOnTaskIds, &task); err != nil {
			return nil, err
		}

		if result.DueDate.Valid {
			task.DueDate = &result.DueDate.Time
		}

		if result.Code.Valid {
			task.Code = result.Code.String
		}

		if result.DoneAt.Valid {
			task.DoneAt = &result.DoneAt.Time
		}

		if result.ArchivedAt.Valid {
			task.ArchivedAt = &result.ArchivedAt.Time
		}

		task.Status = compatibilityTaskStatus(task.ProjectColumn.Name, task.ArchivedAt)

		if result.ProjectProjectID != uuid.Nil {
			task.Project = &domain.Project{
				Id:          result.ProjectProjectID,
				Name:        result.ProjectName,
				Description: result.ProjectDescription,
				CreatedAt:   result.ProjectCreatedAt.Time,
				UpdatedAt:   result.ProjectUpdatedAt.Time,
				UserId:      result.ProjectUserID,
			}
		}

		if result.AuthorAuthorID.Valid {
			task.Author = &domain.User{
				Id:        *(*uuid.UUID)(result.AuthorAuthorID.Bytes[:]),
				Name:      result.AuthorName.String,
				Email:     result.AuthorEmail.String,
				CreatedAt: result.AuthorCreatedAt.Time,
			}
		}

		if result.ResponsibleResponsibleID.Valid {
			task.ResponsibleId = (*uuid.UUID)(result.ResponsibleResponsibleID.Bytes[:])
			task.Responsible = &domain.User{
				Id:    *task.ResponsibleId,
				Name:  result.ResponsibleName.String,
				Email: result.ResponsibleEmail.String,
			}
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

func (tr *TaskRepository) CountTasksInProjectByIds(ctx context.Context, projectId uuid.UUID, taskIds []uuid.UUID) (int, error) {
	if len(taskIds) == 0 {
		return 0, nil
	}

	q := queries.New(tr.pool)
	count, err := q.CountTasksInProjectByIds(ctx, queries.CountTasksInProjectByIdsParams{
		ProjectID: projectId,
		Column2:   taskIds,
	})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (tr *TaskRepository) ListTaskDependenciesByProjectId(ctx context.Context, projectId uuid.UUID) ([]domain.TaskDependencyEdge, error) {
	q := queries.New(tr.pool)
	rows, err := q.ListTaskDependenciesByProjectId(ctx, projectId)
	if err != nil {
		return nil, err
	}

	edges := make([]domain.TaskDependencyEdge, 0, len(rows))
	for _, row := range rows {
		edges = append(edges, domain.TaskDependencyEdge{
			TaskId:          row.TaskID,
			DependsOnTaskId: row.DependsOnTaskID,
		})
	}

	return edges, nil
}

func mapDependsOnTasksFromResult(raw interface{}, task *domain.Task) error {
	if raw == nil {
		task.DependsOnTasks = []domain.TaskDependencyRef{}
		task.DependsOnTaskIds = []uuid.UUID{}
		return nil
	}

	bytes, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	var refs []domain.TaskDependencyRef
	if err := json.Unmarshal(bytes, &refs); err != nil {
		return err
	}

	task.DependsOnTasks = refs
	task.DependsOnTaskIds = make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		task.DependsOnTaskIds = append(task.DependsOnTaskIds, ref.Id)
	}

	return nil
}

func mapDependsOnTaskIdsFromResult(raw interface{}, task *domain.Task) error {
	if raw == nil {
		task.DependsOnTaskIds = []uuid.UUID{}
		return nil
	}

	bytes, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	if err := json.Unmarshal(bytes, &ids); err != nil {
		return err
	}

	task.DependsOnTaskIds = ids
	return nil
}

func (tr *TaskRepository) GetTaskDependencyRefsByProjectAndIds(ctx context.Context, projectId uuid.UUID, ids []uuid.UUID) ([]domain.TaskDependencyRef, error) {
	const q = `
		SELECT id, title, coalesce(code, '') AS code
		FROM tasks
		WHERE project_id = $1 AND id = ANY($2::uuid[])
	`
	rows, err := tr.pool.Query(ctx, q, projectId, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	refs := make([]domain.TaskDependencyRef, 0, len(ids))
	for rows.Next() {
		var ref domain.TaskDependencyRef
		if err := rows.Scan(&ref.Id, &ref.Title, &ref.Code); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}

	return refs, rows.Err()
}

func (tr *TaskRepository) FindTaskRefsByProjectAndCode(ctx context.Context, projectId uuid.UUID, code string) ([]domain.TaskDependencyRef, error) {
	q := queries.New(tr.pool)

	results, err := q.FindTaskRefsByProjectAndCode(ctx, queries.FindTaskRefsByProjectAndCodeParams{
		ProjectID: projectId,
		Code:      code,
	})
	if err != nil {
		return nil, err
	}

	refs := make([]domain.TaskDependencyRef, 0, len(results))
	for _, result := range results {
		refs = append(refs, domain.TaskDependencyRef{
			Id:    result.ID,
			Title: result.Title,
			Code:  result.Code,
		})
	}

	return refs, nil
}

// trailingDigitsPattern matches a run of digits at the end of a task code
// prefix, e.g. the "9" in "BACKEND-9".
var trailingDigitsPattern = regexp.MustCompile(`\d+$`)

// taskCodeSequenceBase strips a trailing run of digits from prefix so that
// e.g. "BACKEND-9" and "BACKEND-" both group under the same numbering
// sequence. Falls back to the full prefix when it's entirely digits.
func taskCodeSequenceBase(prefix string) string {
	base := trailingDigitsPattern.ReplaceAllString(prefix, "")
	if base == "" {
		return prefix
	}
	return base
}

// taskCodeLikePattern escapes Postgres LIKE metacharacters so value is
// matched literally (e.g. a prefix of "AB_1" won't match "ABX1"), then
// appends a trailing wildcard for prefix matching.
func taskCodeLikePattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value) + "%"
}

func (tr *TaskRepository) SuggestTaskCodesByProjectPrefix(ctx context.Context, projectId uuid.UUID, prefix string, limit int) ([]domain.TaskCodeSuggestion, error) {
	q := queries.New(tr.pool)

	prefixLower := strings.ToLower(prefix)
	sequenceBase := taskCodeSequenceBase(prefix)

	results, err := q.SuggestTaskCodesByProjectPrefix(ctx, queries.SuggestTaskCodesByProjectPrefixParams{
		ProjectID:           projectId,
		SequenceBase:        sequenceBase,
		SequenceBasePattern: taskCodeLikePattern(strings.ToLower(sequenceBase)),
		PrefixPattern:       taskCodeLikePattern(prefixLower),
		Limit:               int32(limit),
	})
	if err != nil {
		return nil, err
	}

	suggestions := make([]domain.TaskCodeSuggestion, 0, len(results))
	for _, result := range results {
		suggestions = append(suggestions, domain.TaskCodeSuggestion{
			Code: result.Code,
			Kind: result.Kind,
		})
	}

	return suggestions, nil
}

func (tr *TaskRepository) SearchProjectTasksForDependencies(
	ctx context.Context,
	projectId uuid.UUID,
	query string,
	excludeTaskId *uuid.UUID,
	limit int,
) ([]domain.TaskDependencyRef, error) {
	q := queries.New(tr.pool)

	params := queries.SearchProjectTasksForDependenciesParams{
		ProjectID: projectId,
		Query:     query,
		Limit:     int32(limit),
	}

	if excludeTaskId != nil {
		params.ExcludeTaskID = pgtype.UUID{
			Bytes: *excludeTaskId,
			Valid: true,
		}
	}

	results, err := q.SearchProjectTasksForDependencies(ctx, params)
	if err != nil {
		return nil, err
	}

	tasks := make([]domain.TaskDependencyRef, 0, len(results))
	for _, result := range results {
		tasks = append(tasks, domain.TaskDependencyRef{
			Id:    result.ID,
			Title: result.Title,
			Code:  result.Code,
		})
	}

	return tasks, nil
}
