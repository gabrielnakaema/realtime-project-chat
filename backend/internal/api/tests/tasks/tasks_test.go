package tasks_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cursorPage struct {
	Data    []map[string]any `json:"data"`
	HasNext bool             `json:"has_next"`
}

func taskProjectPayload() map[string]any {
	return map[string]any{
		"name":        "Tasks Project",
		"description": "Project for task endpoint tests",
		"columns": []map[string]any{
			{
				"name":           "Pending",
				"color":          "#64748B",
				"is_done_column": false,
			},
			{
				"name":           "Doing",
				"color":          "#2563EB",
				"is_done_column": false,
			},
			{
				"name":           "Done",
				"color":          "#059669",
				"is_done_column": true,
			},
		},
	}
}

func flattenGroupedTasks(grouped map[string]cursorPage) []map[string]any {
	tasks := []map[string]any{}
	for _, page := range grouped {
		tasks = append(tasks, page.Data...)
	}

	return tasks
}

func TestTaskArchiveAndRestoreEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("tasks@example.com", "password123")
	require.NoError(t, err)

	projectResp, err := client.POST("/projects", taskProjectPayload())
	require.NoError(t, err)
	defer projectResp.Body.Close()
	require.Equal(t, http.StatusCreated, projectResp.StatusCode)

	var project map[string]any
	err = json.NewDecoder(projectResp.Body).Decode(&project)
	require.NoError(t, err)

	projectID := project["id"].(string)
	columns := project["columns"].([]any)
	pendingColumnID := columns[0].(map[string]any)["id"].(string)
	doingColumnID := columns[1].(map[string]any)["id"].(string)

	createTask := func(title string) string {
		resp, postErr := client.POST("/tasks", map[string]any{
			"project_id":        projectID,
			"project_column_id": pendingColumnID,
			"title":             title,
			"description":       fmt.Sprintf("%s description", title),
			"priority":          "low",
			"tags":              []string{},
		})
		require.NoError(t, postErr)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var task map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&task)
		require.NoError(t, decodeErr)

		return task["id"].(string)
	}

	archivedTaskID := createTask("Archive me")
	activeTaskID := createTask("Keep me active")

	archiveResp, err := client.DELETE(fmt.Sprintf("/tasks/%s", archivedTaskID))
	require.NoError(t, err)
	defer archiveResp.Body.Close()
	require.Equal(t, http.StatusOK, archiveResp.StatusCode)

	groupedArchivedResp, err := client.GET(fmt.Sprintf("/tasks/group-by-column?project_id=%s&archived=true&limit=15", projectID))
	require.NoError(t, err)
	defer groupedArchivedResp.Body.Close()
	require.Equal(t, http.StatusOK, groupedArchivedResp.StatusCode)

	var archivedGrouped map[string]cursorPage
	err = json.NewDecoder(groupedArchivedResp.Body).Decode(&archivedGrouped)
	require.NoError(t, err)

	archivedTasks := flattenGroupedTasks(archivedGrouped)
	require.Len(t, archivedTasks, 1)
	assert.Equal(t, archivedTaskID, archivedTasks[0]["id"])

	restoreResp, err := client.POST(fmt.Sprintf("/tasks/%s/restore", archivedTaskID), map[string]any{
		"project_column_id": doingColumnID,
	})
	require.NoError(t, err)
	defer restoreResp.Body.Close()
	require.Equal(t, http.StatusOK, restoreResp.StatusCode)

	var restoredTask map[string]any
	err = json.NewDecoder(restoreResp.Body).Decode(&restoredTask)
	require.NoError(t, err)
	assert.Equal(t, doingColumnID, restoredTask["project_column_id"])
	assert.Nil(t, restoredTask["archived_at"])

	activeGroupedResp, err := client.GET(fmt.Sprintf("/tasks/group-by-column?project_id=%s&project_column_ids=%s&archived=false&limit=15", projectID, doingColumnID))
	require.NoError(t, err)
	defer activeGroupedResp.Body.Close()
	require.Equal(t, http.StatusOK, activeGroupedResp.StatusCode)

	var activeGrouped map[string]cursorPage
	err = json.NewDecoder(activeGroupedResp.Body).Decode(&activeGrouped)
	require.NoError(t, err)

	doingTasks := flattenGroupedTasks(activeGrouped)
	require.Len(t, doingTasks, 1)
	assert.Equal(t, archivedTaskID, doingTasks[0]["id"])

	groupedArchivedResp, err = client.GET(fmt.Sprintf("/tasks/group-by-column?project_id=%s&archived=true&limit=15", projectID))
	require.NoError(t, err)
	defer groupedArchivedResp.Body.Close()
	require.Equal(t, http.StatusOK, groupedArchivedResp.StatusCode)

	archivedGrouped = map[string]cursorPage{}
	err = json.NewDecoder(groupedArchivedResp.Body).Decode(&archivedGrouped)
	require.NoError(t, err)

	archivedTasks = flattenGroupedTasks(archivedGrouped)
	assert.Empty(t, archivedTasks)
	assert.NotEqual(t, archivedTaskID, activeTaskID)
}
