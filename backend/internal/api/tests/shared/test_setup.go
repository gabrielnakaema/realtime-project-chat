package shared

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/api"
	"github.com/gabrielnakaema/project-chat/internal/mcpapi"
	"github.com/gabrielnakaema/project-chat/internal/notificationapi"
	"github.com/gabrielnakaema/project-chat/internal/repository"
	"github.com/gabrielnakaema/project-chat/internal/tasks"
	tasksv1 "github.com/gabrielnakaema/project-chat/internal/tasks/v1"
	"github.com/gabrielnakaema/project-chat/internal/tasksapi"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
)

type TestAPI struct {
	Server            *httptest.Server
	BaseURL           string
	DB                *pgxpool.Pool
	PostgresContainer *postgres.PostgresContainer
}

func SetupTestAPI(t *testing.T) (*TestAPI, func()) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
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
	require.NoError(t, err, "Failed to start postgres container")

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	os.Setenv("DB_DSN", dsn)
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing")
	os.Setenv("ENV", "test")
	os.Setenv("PUBSUB_BROKERS", "localhost:9092")

	projectRoot := findProjectRoot(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(projectRoot)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "Failed to create database connection pool")

	err = goose.SetDialect("postgres")
	require.NoError(t, err)

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	sqlDB := stdlib.OpenDB(*config.ConnConfig)
	defer sqlDB.Close()

	migrationsPath := filepath.Join(projectRoot, "migrations")
	err = goose.Up(sqlDB, migrationsPath)
	require.NoError(t, err, "Failed to run migrations")

	os.Setenv("TASKS_SERVICE_PORT", "0")
	taskRepo := tasks.NewTaskRepository(pool)
	taskCommentRepo := tasks.NewTaskCommentRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	grpcTaskService := tasks.NewTaskService(taskRepo, projectRepo, userRepo)
	grpcTaskCommentService := tasks.NewTaskCommentService(taskCommentRepo, taskRepo, projectRepo, userRepo)

	grpcListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to listen for tasks gRPC server")
	grpcServer := grpc.NewServer()
	tasksv1.RegisterTaskServiceServer(grpcServer, tasks.NewGRPCServer(grpcTaskService, grpcTaskCommentService))
	go grpcServer.Serve(grpcListener)
	os.Setenv("TASKS_AUTHORIZATION_GRPC_TARGET", grpcListener.Addr().String())

	apiInstance, err := api.NewApi()
	require.NoError(t, err, "Failed to create API instance")

	tasksInstance, err := tasksapi.New()
	require.NoError(t, err, "Failed to create tasks API instance")

	mcpInstance, err := mcpapi.New()
	require.NoError(t, err, "Failed to create mcp API instance")

	apiRouter := apiInstance.Router()
	tasksRouter := tasksInstance.Router()
	mcpRouter := mcpInstance.Router()
	combined := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/mcp":
			mcpRouter.ServeHTTP(w, r)
		case r.URL.Path == "/tasks" || strings.HasPrefix(r.URL.Path, "/tasks/"):
			tasksRouter.ServeHTTP(w, r)
		default:
			apiRouter.ServeHTTP(w, r)
		}
	})
	server := httptest.NewServer(combined)

	testAPI := &TestAPI{
		Server:            server,
		BaseURL:           server.URL,
		DB:                pool,
		PostgresContainer: postgresContainer,
	}

	cleanup := func() {
		server.Close()
		grpcServer.Stop()
		mcpInstance.Close()
		tasksInstance.Close()
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate postgres container: %v", err)
		}
	}

	return testAPI, cleanup
}

type TestSplitAPI struct {
	MonolithBaseURL     string
	NotificationBaseURL string
	DB                  *pgxpool.Pool
	PostgresContainer   *postgres.PostgresContainer
}

func SetupTestSplitAPI(t *testing.T) (*TestSplitAPI, func()) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
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
	require.NoError(t, err, "Failed to start postgres container")

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	os.Setenv("DB_DSN", dsn)
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing")
	os.Setenv("ENV", "test")
	os.Setenv("PUBSUB_BROKERS", "localhost:9092")
	os.Setenv("NOTIFICATION_SERVICE_PORT", "0")

	projectRoot := findProjectRoot(t)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(projectRoot)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "Failed to create database connection pool")

	err = goose.SetDialect("postgres")
	require.NoError(t, err)

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)

	sqlDB := stdlib.OpenDB(*config.ConnConfig)
	defer sqlDB.Close()

	migrationsPath := filepath.Join(projectRoot, "migrations")
	err = goose.Up(sqlDB, migrationsPath)
	require.NoError(t, err, "Failed to run migrations")

	apiInstance, err := api.NewApi()
	require.NoError(t, err, "Failed to create API instance")
	monolithServer := httptest.NewServer(apiInstance.Router())

	notificationAPIInstance, err := notificationapi.New()
	require.NoError(t, err, "Failed to create notification API instance")
	notificationServer := httptest.NewServer(notificationAPIInstance.Router())

	testAPI := &TestSplitAPI{
		MonolithBaseURL:     monolithServer.URL,
		NotificationBaseURL: notificationServer.URL,
		DB:                  pool,
		PostgresContainer:   postgresContainer,
	}

	cleanup := func() {
		monolithServer.Close()
		notificationServer.Close()
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate postgres container: %v", err)
		}
	}

	return testAPI, cleanup
}

func (ta *TestSplitAPI) TruncateTables(t *testing.T) {
	ta.DB.Exec(context.Background(), `DO
$$
DECLARE
    tables text;
BEGIN
    SELECT string_agg(format('%I.%I', schemaname, tablename), ', ')
    INTO tables
    FROM pg_tables
    WHERE schemaname = 'public';

    EXECUTE 'TRUNCATE TABLE ' || tables || ' RESTART IDENTITY CASCADE';
END;
$$;`)
}

func (ta *TestAPI) GetBaseURL() string {
	return ta.BaseURL
}

func (ta *TestAPI) GetContext() context.Context {
	return context.Background()
}

func findProjectRoot(t *testing.T) string {
	dir, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			require.Fail(t, "Could not find project root (go.mod file)")
		}
		dir = parent
	}
}

func (ta *TestAPI) TruncateTables(t *testing.T) {
	ta.DB.Exec(context.Background(), `DO
$$
DECLARE
    tables text;
BEGIN
    SELECT string_agg(format('%I.%I', schemaname, tablename), ', ')
    INTO tables
    FROM pg_tables
    WHERE schemaname = 'public';

    EXECUTE 'TRUNCATE TABLE ' || tables || ' RESTART IDENTITY CASCADE';
END;
$$;`)
}
