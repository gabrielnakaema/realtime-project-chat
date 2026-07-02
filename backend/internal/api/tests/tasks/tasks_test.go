package tasks_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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

func TestTaskCodeFieldFlowsThroughTaskEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("task-code@example.com", "password123")
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

	createResp, err := client.POST("/tasks", map[string]any{
		"project_id":        projectID,
		"project_column_id": pendingColumnID,
		"title":             "Code task",
		"description":       "Task with code",
		"code":              "  TASK-301  ",
		"priority":          "medium",
		"tags":              []string{"backend"},
	})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createdTask map[string]any
	err = json.NewDecoder(createResp.Body).Decode(&createdTask)
	require.NoError(t, err)
	assert.Equal(t, "TASK-301", createdTask["code"])

	taskID := createdTask["id"].(string)

	getResp, err := client.GET(fmt.Sprintf("/tasks/%s", taskID))
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetchedTask map[string]any
	err = json.NewDecoder(getResp.Body).Decode(&fetchedTask)
	require.NoError(t, err)
	assert.Equal(t, "TASK-301", fetchedTask["code"])

	groupedResp, err := client.GET(fmt.Sprintf("/tasks/group-by-column?project_id=%s&project_column_ids=%s&archived=false&limit=15", projectID, pendingColumnID))
	require.NoError(t, err)
	defer groupedResp.Body.Close()
	require.Equal(t, http.StatusOK, groupedResp.StatusCode)

	var grouped map[string]cursorPage
	err = json.NewDecoder(groupedResp.Body).Decode(&grouped)
	require.NoError(t, err)
	require.Len(t, grouped[pendingColumnID].Data, 1)
	assert.Equal(t, "TASK-301", grouped[pendingColumnID].Data[0]["code"])

	updateResp, err := client.PUT(fmt.Sprintf("/tasks/%s", taskID), map[string]any{
		"title":             "Code task updated",
		"description":       "Task with cleared code",
		"code":              "   ",
		"project_column_id": pendingColumnID,
		"priority":          "high",
		"tags":              []string{"backend", "updated"},
	})
	require.NoError(t, err)
	defer updateResp.Body.Close()
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updatedTask map[string]any
	err = json.NewDecoder(updateResp.Body).Decode(&updatedTask)
	require.NoError(t, err)
	assert.Equal(t, "", updatedTask["code"])
}

func TestTaskCodeSuggestionsEndpoint(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("task-code-suggestions@example.com", "password123")
	require.NoError(t, err)

	projectResp, err := client.POST("/projects", taskProjectPayload())
	require.NoError(t, err)
	defer projectResp.Body.Close()
	require.Equal(t, http.StatusCreated, projectResp.StatusCode)

	var project map[string]any
	err = json.NewDecoder(projectResp.Body).Decode(&project)
	require.NoError(t, err)

	projectID := project["id"].(string)
	pendingColumnID := project["columns"].([]any)[0].(map[string]any)["id"].(string)

	createTask := func(title string, code string) string {
		resp, postErr := client.POST("/tasks", map[string]any{
			"project_id":        projectID,
			"project_column_id": pendingColumnID,
			"title":             title,
			"description":       fmt.Sprintf("%s description", title),
			"code":              code,
			"priority":          "medium",
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

	createTask("Backend one", "BACKEND-001")
	archivedTaskID := createTask("Backend nine", "BACKEND-009")
	createTask("Frontend one", "FRONTEND-001")
	createTask("API one", "API-1")

	archiveResp, err := client.DELETE(fmt.Sprintf("/tasks/%s", archivedTaskID))
	require.NoError(t, err)
	defer archiveResp.Body.Close()
	require.Equal(t, http.StatusOK, archiveResp.StatusCode)

	resp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=BACKEND-&limit=4", projectID))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	data := result["data"].([]any)
	require.NotEmpty(t, data)

	next := data[0].(map[string]any)
	assert.Equal(t, "BACKEND-010", next["code"])
	assert.Equal(t, "next", next["kind"])

	codes := []string{}
	for _, item := range data {
		codes = append(codes, item.(map[string]any)["code"].(string))
	}
	assert.Contains(t, codes, "BACKEND-009")
	assert.Contains(t, codes, "BACKEND-001")
	assert.NotContains(t, codes, "FRONTEND-001")

	typedNumberResp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=BACKEND-009&limit=4", projectID))
	require.NoError(t, err)
	defer typedNumberResp.Body.Close()
	require.Equal(t, http.StatusOK, typedNumberResp.StatusCode)

	var typedNumberResult map[string]any
	err = json.NewDecoder(typedNumberResp.Body).Decode(&typedNumberResult)
	require.NoError(t, err)

	typedNumberData := typedNumberResult["data"].([]any)
	require.NotEmpty(t, typedNumberData)
	typedNumberNext := typedNumberData[0].(map[string]any)
	assert.Equal(t, "BACKEND-010", typedNumberNext["code"])
	assert.Equal(t, "next", typedNumberNext["kind"])

	singleDigitResp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=API-1&limit=4", projectID))
	require.NoError(t, err)
	defer singleDigitResp.Body.Close()
	require.Equal(t, http.StatusOK, singleDigitResp.StatusCode)

	var singleDigitResult map[string]any
	err = json.NewDecoder(singleDigitResp.Body).Decode(&singleDigitResult)
	require.NoError(t, err)

	singleDigitData := singleDigitResult["data"].([]any)
	require.NotEmpty(t, singleDigitData)
	singleDigitNext := singleDigitData[0].(map[string]any)
	assert.Equal(t, "API-2", singleDigitNext["code"])
	assert.Equal(t, "next", singleDigitNext["kind"])

	lowerCaseResp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=backend-&limit=4", projectID))
	require.NoError(t, err)
	defer lowerCaseResp.Body.Close()
	require.Equal(t, http.StatusOK, lowerCaseResp.StatusCode)

	var lowerCaseResult map[string]any
	err = json.NewDecoder(lowerCaseResp.Body).Decode(&lowerCaseResult)
	require.NoError(t, err)

	lowerCaseData := lowerCaseResult["data"].([]any)
	require.NotEmpty(t, lowerCaseData)
	lowerCaseNext := lowerCaseData[0].(map[string]any)
	assert.Equal(t, "BACKEND-010", lowerCaseNext["code"], "next suggestion should follow the casing of existing codes, not the typed prefix")
	assert.Equal(t, "next", lowerCaseNext["kind"])

	createTask("Underscore literal", "AB_1")
	createTask("Underscore lookalike", "ABX1")

	wildcardResp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=AB_&limit=10", projectID))
	require.NoError(t, err)
	defer wildcardResp.Body.Close()
	require.Equal(t, http.StatusOK, wildcardResp.StatusCode)

	var wildcardResult map[string]any
	err = json.NewDecoder(wildcardResp.Body).Decode(&wildcardResult)
	require.NoError(t, err)

	wildcardCodes := []string{}
	for _, item := range wildcardResult["data"].([]any) {
		wildcardCodes = append(wildcardCodes, item.(map[string]any)["code"].(string))
	}
	assert.Contains(t, wildcardCodes, "AB_1")
	assert.NotContains(t, wildcardCodes, "ABX1", "'_' in the prefix must be matched literally, not as a LIKE wildcard")

	createTask("Server two", "SRV-2")
	createTask("Server ten", "SRV-10")

	orderingResp, err := client.GET(fmt.Sprintf("/tasks/code-suggestions?project_id=%s&prefix=SRV-&limit=10", projectID))
	require.NoError(t, err)
	defer orderingResp.Body.Close()
	require.Equal(t, http.StatusOK, orderingResp.StatusCode)

	var orderingResult map[string]any
	err = json.NewDecoder(orderingResp.Body).Decode(&orderingResult)
	require.NoError(t, err)

	orderingCodes := []string{}
	for _, item := range orderingResult["data"].([]any) {
		orderingCodes = append(orderingCodes, item.(map[string]any)["code"].(string))
	}
	srv2Index := slices.Index(orderingCodes, "SRV-2")
	srv10Index := slices.Index(orderingCodes, "SRV-10")
	require.NotEqual(t, -1, srv2Index)
	require.NotEqual(t, -1, srv10Index)
	assert.Less(t, srv10Index, srv2Index, "existing codes should be ordered numerically so SRV-10 ranks above SRV-2")
}

func TestTaskDependenciesEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("task-deps@example.com", "password123")
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

	createTask := func(title string, dependsOn []string) map[string]any {
		payload := map[string]any{
			"project_id":        projectID,
			"project_column_id": pendingColumnID,
			"title":             title,
			"description":       fmt.Sprintf("%s description", title),
			"priority":          "medium",
			"tags":              []string{},
		}
		if dependsOn != nil {
			payload["depends_on_task_ids"] = dependsOn
		}

		resp, postErr := client.POST("/tasks", payload)
		require.NoError(t, postErr)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var created map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&created)
		require.NoError(t, decodeErr)

		return created
	}

	blocker := createTask("Blocker", nil)
	blockerID := blocker["id"].(string)

	blocked := createTask("Blocked", []string{blockerID})
	blockedID := blocked["id"].(string)
	assert.Equal(t, []any{blockerID}, blocked["depends_on_task_ids"])

	getResp, err := client.GET(fmt.Sprintf("/tasks/%s", blockedID))
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	var fetchedTask map[string]any
	err = json.NewDecoder(getResp.Body).Decode(&fetchedTask)
	require.NoError(t, err)
	assert.Equal(t, []any{blockerID}, fetchedTask["depends_on_task_ids"])

	dependsOnTasks := fetchedTask["depends_on_tasks"].([]any)
	require.Len(t, dependsOnTasks, 1)
	dependsOnTask := dependsOnTasks[0].(map[string]any)
	assert.Equal(t, blockerID, dependsOnTask["id"])
	assert.Equal(t, "Blocker", dependsOnTask["title"])

	cycleResp, err := client.PUT(fmt.Sprintf("/tasks/%s", blockerID), map[string]any{
		"title":               "Blocker",
		"description":         "Cycle attempt",
		"project_column_id":   pendingColumnID,
		"priority":            "medium",
		"tags":                []string{},
		"depends_on_task_ids": []string{blockedID},
	})
	require.NoError(t, err)
	defer cycleResp.Body.Close()
	assert.Equal(t, http.StatusUnprocessableEntity, cycleResp.StatusCode)

	clearResp, err := client.PUT(fmt.Sprintf("/tasks/%s", blockedID), map[string]any{
		"title":               "Blocked",
		"description":         "No dependencies",
		"project_column_id":   pendingColumnID,
		"priority":            "medium",
		"tags":                []string{},
		"depends_on_task_ids": []string{},
	})
	require.NoError(t, err)
	defer clearResp.Body.Close()
	require.Equal(t, http.StatusOK, clearResp.StatusCode)

	var clearedTask map[string]any
	err = json.NewDecoder(clearResp.Body).Decode(&clearedTask)
	require.NoError(t, err)
	assert.Equal(t, []any{}, clearedTask["depends_on_task_ids"])
}

func TestCreateTaskResponsibleIdValidation(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("responsible-id@example.com", "password123")
	require.NoError(t, err)

	projectResp, err := client.POST("/projects", taskProjectPayload())
	require.NoError(t, err)
	defer projectResp.Body.Close()
	require.Equal(t, http.StatusCreated, projectResp.StatusCode)

	var project map[string]any
	err = json.NewDecoder(projectResp.Body).Decode(&project)
	require.NoError(t, err)

	projectID := project["id"].(string)
	pendingColumnID := project["columns"].([]any)[0].(map[string]any)["id"].(string)

	basePayload := map[string]any{
		"project_id":        projectID,
		"project_column_id": pendingColumnID,
		"title":             "Task",
		"description":       "Description",
		"priority":          "medium",
	}

	t.Run("invalid uuid string is rejected", func(t *testing.T) {
		payload := map[string]any{}
		for k, v := range basePayload {
			payload[k] = v
		}
		payload["responsible_id"] = "not-a-uuid"

		resp, postErr := client.POST("/tasks", payload)
		require.NoError(t, postErr)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("null responsible_id is accepted", func(t *testing.T) {
		payload := map[string]any{}
		for k, v := range basePayload {
			payload[k] = v
		}
		payload["responsible_id"] = nil

		resp, postErr := client.POST("/tasks", payload)
		require.NoError(t, postErr)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		assert.Nil(t, created["responsible_id"])
	})

	t.Run("omitted responsible_id is accepted", func(t *testing.T) {
		resp, postErr := client.POST("/tasks", basePayload)
		require.NoError(t, postErr)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		assert.Nil(t, created["responsible_id"])
	})
}

func TestTaskProjectSearchForDependenciesEndpoint(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("project-search@example.com", "password123")
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

	createResp, err := client.POST("/tasks", map[string]any{
		"project_id":        projectID,
		"project_column_id": pendingColumnID,
		"title":             "Auth middleware",
		"description":       "<p>Implement JWT validation middleware</p>",
		"code":              "BACKEND-9",
		"priority":          "medium",
		"tags":              []string{},
	})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createdTask map[string]any
	err = json.NewDecoder(createResp.Body).Decode(&createdTask)
	require.NoError(t, err)
	taskID := createdTask["id"].(string)

	searchByCodeResp, err := client.GET(fmt.Sprintf("/tasks/project-search?project_id=%s&query=backend-9", projectID))
	require.NoError(t, err)
	defer searchByCodeResp.Body.Close()
	require.Equal(t, http.StatusOK, searchByCodeResp.StatusCode)

	var searchByCodeResult map[string]any
	err = json.NewDecoder(searchByCodeResp.Body).Decode(&searchByCodeResult)
	require.NoError(t, err)
	searchByCodeData := searchByCodeResult["data"].([]any)
	require.Len(t, searchByCodeData, 1)
	searchByCodeTask := searchByCodeData[0].(map[string]any)
	assert.Equal(t, taskID, searchByCodeTask["id"])
	assert.Equal(t, "Auth middleware", searchByCodeTask["title"])
	assert.Equal(t, "BACKEND-9", searchByCodeTask["code"])

	searchByDescriptionResp, err := client.GET("/tasks/project-search?project_id=" + projectID + "&query=jwt+validation")
	require.NoError(t, err)
	defer searchByDescriptionResp.Body.Close()
	require.Equal(t, http.StatusOK, searchByDescriptionResp.StatusCode)

	var searchByDescriptionResult map[string]any
	err = json.NewDecoder(searchByDescriptionResp.Body).Decode(&searchByDescriptionResult)
	require.NoError(t, err)
	searchByDescriptionData := searchByDescriptionResult["data"].([]any)
	require.Len(t, searchByDescriptionData, 1)

	excludeResp, err := client.GET(fmt.Sprintf("/tasks/project-search?project_id=%s&query=auth&exclude_task_id=%s", projectID, taskID))
	require.NoError(t, err)
	defer excludeResp.Body.Close()
	require.Equal(t, http.StatusOK, excludeResp.StatusCode)

	var excludeResult map[string]any
	err = json.NewDecoder(excludeResp.Body).Decode(&excludeResult)
	require.NoError(t, err)
	assert.Empty(t, excludeResult["data"].([]any))
}
