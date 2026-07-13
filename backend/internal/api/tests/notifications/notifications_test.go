package notifications_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestSplitAPI(t)
	defer cleanup()

	t.Run("notifications endpoints are protected by auth", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.NotificationBaseURL)

		resp, err := client.GET("/notifications")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.GET("/notifications/unread-count")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("list, count and mark notifications as read", func(t *testing.T) {
		testAPI.TruncateTables(t)

		monolithClient := shared.NewHTTPClient(testAPI.MonolithBaseURL)
		token, err := monolithClient.CreateUserAndLogin("notifications@example.com", "password123")
		require.NoError(t, err)

		client := shared.NewHTTPClient(testAPI.NotificationBaseURL)
		client.SetAuthToken(token)

		actorID := createUser(t, testAPI, "actor@example.com")
		userID := getCurrentUserID(t, monolithClient)
		projectID := uuid.New()
		projectColumnID := uuid.New()
		taskID := uuid.New()
		notificationID := uuid.New()

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO projects (id, user_id, name, description)
			VALUES ($1, $2, 'Test Project', 'Project Description')
		`, projectID, actorID)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO project_members (id, user_id, project_id, role)
			VALUES ($1, $2, $3, 'creator'), ($4, $5, $3, 'member')
		`, uuid.New(), actorID, projectID, uuid.New(), userID)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
				INSERT INTO project_columns (id, project_id, name, color, position, is_done_column)
				VALUES ($1, $2, 'QA Review', '#F59E0B', 3, false)
			`, projectColumnID, projectID)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO tasks (id, project_id, title, description, project_column_id, author_id, priority, task_order)
			VALUES ($1, $2, 'Task title', 'Task description', $3, $4, 'medium', '500000000000')
		`, taskID, projectID, projectColumnID, actorID)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO notifications (id, user_id, actor_id, project_id, task_id, type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'task.assigned', $6, $6)
		`, notificationID, userID, actorID, projectID, taskID, time.Now().UTC())
		require.NoError(t, err)

		resp, err := client.GET("/notifications")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResponse map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResponse))
		assert.Len(t, listResponse["data"], 1)

		data, ok := listResponse["data"].([]any)
		require.True(t, ok)
		notification, ok := data[0].(map[string]any)
		require.True(t, ok)
		task, ok := notification["task"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "qa review", task["status"])

		projectColumn, ok := task["project_column"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "QA Review", projectColumn["name"])
		assert.Equal(t, "#F59E0B", projectColumn["color"])

		resp, err = client.GET("/notifications/unread-count")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var countResponse map[string]int
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&countResponse))
		assert.Equal(t, 1, countResponse["count"])

		resp, err = client.POST("/notifications/"+notificationID.String()+"/read", map[string]string{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = client.GET("/notifications/unread-count")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&countResponse))
		assert.Equal(t, 0, countResponse["count"])

		secondNotificationID := uuid.New()
		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO notifications (id, user_id, actor_id, project_id, type, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'project.member.created', $5, $5)
		`, secondNotificationID, userID, actorID, projectID, time.Now().UTC())
		require.NoError(t, err)

		resp, err = client.POST("/notifications/read-all", map[string]string{})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		resp, err = client.GET("/notifications/unread-count")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&countResponse))
		assert.Equal(t, 0, countResponse["count"])
	})
}

func createUser(t *testing.T, testAPI *shared.TestSplitAPI, email string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := testAPI.DB.Exec(context.Background(), `
		INSERT INTO users (id, name, email, password)
		VALUES ($1, 'Actor', $2, 'password123')
	`, id, email)
	require.NoError(t, err)

	return id
}

func getCurrentUserID(t *testing.T, client *shared.HTTPClient) uuid.UUID {
	t.Helper()

	resp, err := client.GET("/users/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))

	parsed, err := uuid.Parse(user.ID)
	require.NoError(t, err)
	return parsed
}
