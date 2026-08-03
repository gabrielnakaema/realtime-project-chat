package tasks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type taskRepositoryFixture struct {
	projectID       uuid.UUID
	projectColumnID uuid.UUID
	taskID          uuid.UUID
	dependencyID    uuid.UUID
	authorID        uuid.UUID
	responsibleID   uuid.UUID
	dueDate         time.Time
	doneAt          time.Time
	createdAt       time.Time
	updatedAt       time.Time
	userCreatedAt   time.Time
	updateID        uuid.UUID
	changeID        uuid.UUID
	updateCreatedAt time.Time
	changeCreatedAt time.Time
}

func TestTaskRepositoryListRowMappingIntegration(t *testing.T) {
	repository, pool := setupTaskRepositoryIntegrationTest(t)
	fixture := seedTaskRepositoryFixture(t, pool)
	ctx := context.Background()

	t.Run("ListByProjectId maps the complete task record", func(t *testing.T) {
		page, err := repository.ListByProjectId(ctx, fixture.projectID, nil, false, "", nil, 10)
		require.NoError(t, err)
		require.Len(t, page.Data, 2)
		assert.False(t, page.HasNext)

		task := findTask(t, page.Data, fixture.taskID)
		assertCommonTaskMapping(t, task, fixture)
		assert.Nil(t, task.Project)
		require.NotNil(t, task.Author)
		assert.Equal(t, fixture.authorID, task.Author.Id)
		assert.Equal(t, "Repository Author", task.Author.Name)
		assert.Equal(t, "author@example.com", task.Author.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Author.CreatedAt, time.Microsecond)
		require.NotNil(t, task.Responsible)
		assert.Equal(t, fixture.responsibleID, task.Responsible.Id)
		assert.Equal(t, "Repository Responsible", task.Responsible.Name)
		assert.Equal(t, "responsible@example.com", task.Responsible.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Responsible.CreatedAt, time.Microsecond)
	})

	t.Run("ListUserDueTasks maps the complete task record", func(t *testing.T) {
		page, err := repository.ListUserDueTasks(ctx, fixture.responsibleID, nil, nil, 10)
		require.NoError(t, err)
		require.Len(t, page.Data, 1)
		assert.False(t, page.HasNext)

		task := page.Data[0]
		assertCommonTaskMapping(t, task, fixture)
		assertProjectRelation(t, task.Project, fixture)
		require.NotNil(t, task.Author)
		assert.Equal(t, fixture.authorID, task.Author.Id)
		assert.Equal(t, "author@example.com", task.Author.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Author.CreatedAt, time.Microsecond)
		require.NotNil(t, task.Responsible)
		assert.Equal(t, fixture.responsibleID, task.Responsible.Id)
		assert.Equal(t, "Repository Responsible", task.Responsible.Name)
		assert.Equal(t, "responsible@example.com", task.Responsible.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Responsible.CreatedAt, time.Microsecond)
	})

	t.Run("SearchTasks maps the complete task record", func(t *testing.T) {
		page, err := repository.SearchTasks(ctx, SearchTasksRequest{
			UserId:      fixture.responsibleID,
			SearchQuery: "shared mapper target",
			Limit:       10,
		})
		require.NoError(t, err)
		require.Len(t, page.Data, 1)
		assert.False(t, page.HasNext)

		task := page.Data[0]
		assertCommonTaskMapping(t, task, fixture)
		assertProjectRelation(t, task.Project, fixture)
		require.NotNil(t, task.Author)
		assert.Equal(t, fixture.authorID, task.Author.Id)
		assert.Equal(t, "author@example.com", task.Author.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Author.CreatedAt, time.Microsecond)
		require.NotNil(t, task.Responsible)
		assert.Equal(t, fixture.responsibleID, task.Responsible.Id)
		assert.Equal(t, "responsible@example.com", task.Responsible.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Responsible.CreatedAt, time.Microsecond)
	})

	t.Run("GetById returns every task-owned detail relation", func(t *testing.T) {
		task, err := repository.GetById(ctx, fixture.taskID)
		require.NoError(t, err)
		assertCommonTaskMapping(t, *task, fixture)
		assert.Nil(t, task.Project, "project authorization is loaded separately by TaskService")

		require.NotNil(t, task.Author)
		assert.Equal(t, fixture.authorID, task.Author.Id)
		assert.Equal(t, "Repository Author", task.Author.Name)
		assert.Equal(t, "author@example.com", task.Author.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Author.CreatedAt, time.Microsecond)

		require.NotNil(t, task.Responsible)
		assert.Equal(t, fixture.responsibleID, task.Responsible.Id)
		assert.Equal(t, "Repository Responsible", task.Responsible.Name)
		assert.Equal(t, "responsible@example.com", task.Responsible.Email)
		assert.WithinDuration(t, fixture.userCreatedAt, task.Responsible.CreatedAt, time.Microsecond)

		require.Len(t, task.DependsOnTasks, 1)
		assert.Equal(t, fixture.dependencyID, task.DependsOnTasks[0].Id)
		assert.Equal(t, "Dependency task", task.DependsOnTasks[0].Title)
		assert.Equal(t, "MAP-1", task.DependsOnTasks[0].Code)

		require.Len(t, task.Updates, 1)
		update := task.Updates[0]
		assert.Equal(t, fixture.updateID, update.Id)
		assert.Equal(t, fixture.taskID, update.TaskId)
		assert.Equal(t, fixture.authorID, update.UserId)
		assert.Equal(t, domain.TaskUpdateTypeUpdated, update.UpdateType)
		assert.Equal(t, domain.ActionOriginMCPAgent, update.ActionOrigin)
		assert.WithinDuration(t, fixture.updateCreatedAt, update.CreatedAt, time.Microsecond)
		require.NotNil(t, update.User)
		assert.Equal(t, fixture.authorID, update.User.Id)
		assert.Equal(t, "Repository Author", update.User.Name)
		assert.Equal(t, "author@example.com", update.User.Email)

		require.Len(t, update.Changes, 1)
		change := update.Changes[0]
		assert.Equal(t, fixture.changeID, change.Id)
		assert.Equal(t, fixture.updateID, change.UpdateId)
		assert.Equal(t, "responsible_id", change.Field)
		assert.Equal(t, fixture.authorID.String(), change.OldValue)
		assert.Equal(t, fixture.responsibleID.String(), change.NewValue)
		require.NotNil(t, change.OldDisplayValue)
		assert.Equal(t, "Repository Author", *change.OldDisplayValue)
		require.NotNil(t, change.NewDisplayValue)
		assert.Equal(t, "Repository Responsible", *change.NewDisplayValue)
		assert.WithinDuration(t, fixture.changeCreatedAt, change.CreatedAt, time.Microsecond)
		require.NotNil(t, change.Subject)
		assert.Equal(t, fixture.responsibleID, change.Subject.Id)
		assert.Equal(t, "Repository Responsible", change.Subject.Name)
		assert.Equal(t, "responsible@example.com", change.Subject.Email)
	})
}

func setupTaskRepositoryIntegrationTest(t *testing.T) (*TaskRepository, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	sqlDB := stdlib.OpenDB(*config.ConnConfig)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(sqlDB, filepath.Join(findBackendRoot(t), "migrations")))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return NewTaskRepository(pool), pool
}

func findBackendRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find backend root containing go.mod")
		}
		dir = parent
	}
}

func seedTaskRepositoryFixture(t *testing.T, pool *pgxpool.Pool) taskRepositoryFixture {
	t.Helper()
	ctx := context.Background()
	fixture := taskRepositoryFixture{
		projectID:       uuid.New(),
		projectColumnID: uuid.New(),
		taskID:          uuid.New(),
		dependencyID:    uuid.New(),
		authorID:        uuid.New(),
		responsibleID:   uuid.New(),
		dueDate:         time.Date(2030, time.March, 10, 12, 30, 0, 0, time.UTC),
		doneAt:          time.Date(2030, time.March, 9, 11, 15, 0, 0, time.UTC),
		createdAt:       time.Date(2029, time.December, 1, 9, 0, 0, 0, time.UTC),
		updatedAt:       time.Date(2030, time.January, 2, 10, 0, 0, 0, time.UTC),
		userCreatedAt:   time.Date(2029, time.November, 1, 8, 0, 0, 0, time.UTC),
		updateID:        uuid.New(),
		changeID:        uuid.New(),
		updateCreatedAt: time.Date(2030, time.January, 2, 10, 5, 0, 0, time.UTC),
		changeCreatedAt: time.Date(2030, time.January, 2, 10, 6, 0, 0, time.UTC),
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password, created_at)
		VALUES
			($1, 'Repository Author', 'author@example.com', 'password', $3),
			($2, 'Repository Responsible', 'responsible@example.com', 'password', $3)
	`, fixture.authorID, fixture.responsibleID, fixture.userCreatedAt)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO projects (
			id, user_id, name, description, repository_url, repository_owner,
			repository_name, default_branch, branch_name_prefix, created_at, updated_at
		)
		VALUES (
			$1, $2, 'Repository Mapping Project', 'Integration fixture',
			'https://example.com/acme/project-chat', 'acme', 'project-chat', 'main', 'task/', $3, $4
		)
	`, fixture.projectID, fixture.authorID, fixture.createdAt, fixture.updatedAt)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO project_members (user_id, project_id, role)
		VALUES ($1, $3, 'creator'), ($2, $3, 'member')
	`, fixture.authorID, fixture.responsibleID, fixture.projectID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO project_columns (id, project_id, name, description, color, position, is_done_column, created_at, updated_at)
		VALUES ($1, $2, 'Doing', 'Work is in progress', '#2563EB', 1, false, $3, $4)
	`, fixture.projectColumnID, fixture.projectID, fixture.createdAt, fixture.updatedAt)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO tasks (
			id, project_id, title, description, author_id, priority, due_date, done_at,
			responsible_id, task_order, project_column_id, code, version, created_at, updated_at
		)
		VALUES
			($1, $3, 'Shared mapper target', 'All common fields', $4, 'high', $6, $7, $5, 'a0', $8, 'MAP-42', 7, $9, $10),
			($2, $3, 'Dependency task', 'Dependency fixture', $4, 'low', NULL, NULL, NULL, 'b0', $8, 'MAP-1', 2, $9, $10)
	`,
		fixture.taskID,
		fixture.dependencyID,
		fixture.projectID,
		fixture.authorID,
		fixture.responsibleID,
		fixture.dueDate,
		fixture.doneAt,
		fixture.projectColumnID,
		fixture.createdAt,
		fixture.updatedAt,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO task_tags (task_id, name) VALUES ($1, 'backend'), ($1, 'urgent')
	`, fixture.taskID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO task_dependencies (task_id, depends_on_task_id) VALUES ($1, $2)
	`, fixture.taskID, fixture.dependencyID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO task_updates (id, task_id, user_id, update_type, action_origin, created_at)
		VALUES ($1, $2, $3, 'updated', 'mcp_agent', $4)
	`, fixture.updateID, fixture.taskID, fixture.authorID, fixture.updateCreatedAt)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO task_changes (
			id, update_id, subject_id, old_value_id, new_value_id, field,
			old_value, new_value, created_at
		)
		VALUES ($1, $2, $3, $4, $3, 'responsible_id', $5, $6, $7)
	`,
		fixture.changeID,
		fixture.updateID,
		fixture.responsibleID,
		fixture.authorID,
		fixture.authorID.String(),
		fixture.responsibleID.String(),
		fixture.changeCreatedAt,
	)
	require.NoError(t, err)

	return fixture
}

func findTask(t *testing.T, tasks []domain.Task, id uuid.UUID) domain.Task {
	t.Helper()
	for _, task := range tasks {
		if task.Id == id {
			return task
		}
	}
	t.Fatalf("task %s not found", id)
	return domain.Task{}
}

func assertCommonTaskMapping(t *testing.T, task domain.Task, fixture taskRepositoryFixture) {
	t.Helper()
	assert.Equal(t, fixture.taskID, task.Id)
	assert.Equal(t, fixture.projectID, task.ProjectId)
	assert.Equal(t, fixture.authorID, task.AuthorId)
	assert.Equal(t, "Shared mapper target", task.Title)
	assert.Equal(t, "All common fields", task.Description)
	assert.Equal(t, "MAP-42", task.Code)
	assert.Equal(t, domain.TaskStatusDoing, task.Status)
	assert.Equal(t, fixture.projectColumnID, task.ProjectColumnId)
	assert.Equal(t, domain.TaskPriorityHigh, task.Priority)
	assert.Equal(t, "a0", task.Order)
	assert.Equal(t, 7, task.Version)
	require.NotNil(t, task.ResponsibleId)
	assert.Equal(t, fixture.responsibleID, *task.ResponsibleId)
	require.NotNil(t, task.DueDate)
	assert.WithinDuration(t, fixture.dueDate, *task.DueDate, time.Microsecond)
	require.NotNil(t, task.DoneAt)
	assert.WithinDuration(t, fixture.doneAt, *task.DoneAt, time.Microsecond)
	assert.Nil(t, task.ArchivedAt)
	assert.ElementsMatch(t, []string{"backend", "urgent"}, task.Tags)
	assert.Equal(t, []uuid.UUID{fixture.dependencyID}, task.DependsOnTaskIds)
	assert.WithinDuration(t, fixture.createdAt, task.CreatedAt, time.Microsecond)
	assert.WithinDuration(t, fixture.updatedAt, task.UpdatedAt, time.Microsecond)
	require.NotNil(t, task.ProjectColumn)
	assert.Equal(t, fixture.projectColumnID, task.ProjectColumn.Id)
	assert.Equal(t, "Doing", task.ProjectColumn.Name)
	assert.Equal(t, "Work is in progress", task.ProjectColumn.Description)
	assert.Equal(t, "#2563EB", task.ProjectColumn.Color)
	assert.Equal(t, 1, task.ProjectColumn.Position)
	assert.False(t, task.ProjectColumn.IsDoneColumn)
	assert.WithinDuration(t, fixture.createdAt, task.ProjectColumn.CreatedAt, time.Microsecond)
	assert.WithinDuration(t, fixture.updatedAt, task.ProjectColumn.UpdatedAt, time.Microsecond)
}

func assertProjectRelation(t *testing.T, project *domain.Project, fixture taskRepositoryFixture) {
	t.Helper()
	require.NotNil(t, project)
	assert.Equal(t, fixture.projectID, project.Id)
	assert.Equal(t, fixture.authorID, project.UserId)
	assert.Equal(t, "Repository Mapping Project", project.Name)
	assert.Equal(t, "Integration fixture", project.Description)
	assert.Equal(t, "https://example.com/acme/project-chat", project.RepositoryURL)
	assert.Equal(t, "acme", project.RepositoryOwner)
	assert.Equal(t, "project-chat", project.RepositoryName)
	assert.Equal(t, "main", project.DefaultBranch)
	assert.Equal(t, "task/", project.BranchNamePrefix)
	assert.WithinDuration(t, fixture.createdAt, project.CreatedAt, time.Microsecond)
	assert.WithinDuration(t, fixture.updatedAt, project.UpdatedAt, time.Microsecond)
}
