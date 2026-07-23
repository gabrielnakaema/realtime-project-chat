package tasks_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceTaskSearchRemainsInclusiveAndActiveOnly(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)
	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("task-search@example.com", "password123")
	require.NoError(t, err)

	projectResp, err := client.POST("/projects", taskProjectPayload())
	require.NoError(t, err)
	defer projectResp.Body.Close()
	require.Equal(t, http.StatusCreated, projectResp.StatusCode)

	var project map[string]any
	require.NoError(t, json.NewDecoder(projectResp.Body).Decode(&project))
	projectID := project["id"].(string)
	columns := project["columns"].([]any)
	pendingColumnID := columns[0].(map[string]any)["id"].(string)
	doneColumnID := columns[2].(map[string]any)["id"].(string)

	createTask := func(columnID, title, description, code string) string {
		t.Helper()
		resp, createErr := client.POST("/tasks", map[string]any{
			"project_id":        projectID,
			"project_column_id": columnID,
			"title":             title,
			"description":       description,
			"code":              code,
			"priority":          "medium",
			"tags":              []string{},
		})
		require.NoError(t, createErr)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var task map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&task))
		return task["id"].(string)
	}

	titleTaskID := createTask(pendingColumnID, "Needle in title", "Other description", "TITLE-1")
	descriptionTaskID := createTask(pendingColumnID, "Description match", "Contains NEEDLE here", "DESC-1")
	codeTaskID := createTask(pendingColumnID, "Code match", "Other description", "CODE-NEEDLE")
	createTask(doneColumnID, "Needle completed", "Done task", "DONE-1")
	archivedTaskID := createTask(pendingColumnID, "Needle archived", "Archived task", "ARCHIVE-1")

	archiveResp, err := client.DELETE("/tasks/" + archivedTaskID)
	require.NoError(t, err)
	defer archiveResp.Body.Close()
	require.Equal(t, http.StatusOK, archiveResp.StatusCode)

	searchResp, err := client.GET("/tasks/search?query=nEeDlE&limit=15")
	require.NoError(t, err)
	defer searchResp.Body.Close()
	require.Equal(t, http.StatusOK, searchResp.StatusCode)

	var page cursorPage
	require.NoError(t, json.NewDecoder(searchResp.Body).Decode(&page))
	require.Len(t, page.Data, 3)
	resultIDs := make([]string, 0, len(page.Data))
	for _, task := range page.Data {
		resultIDs = append(resultIDs, task["id"].(string))
	}
	assert.ElementsMatch(t, []string{titleTaskID, descriptionTaskID, codeTaskID}, resultIDs)
}
