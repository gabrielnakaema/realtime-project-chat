package mcp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mcpProjectPayload() map[string]any {
	return map[string]any{
		"name":        "MCP Project",
		"description": "Project for MCP tests",
		"columns": []map[string]any{
			{"name": "Pending", "color": "#64748B", "is_done_column": false},
			{"name": "Done", "color": "#059669", "is_done_column": true},
		},
	}
}

func createMCPRequest(t *testing.T, baseURL string, apiKey string, payload map[string]any) map[string]any {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/mcp", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var decoded map[string]any
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	require.NoError(t, err)

	return decoded
}

func postMCP(t *testing.T, baseURL string, apiKey string, payload map[string]any) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/mcp", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func TestMCPAPIKeyLifecycleAndToolAuth(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	testAPI.TruncateTables(t)

	client := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err := client.CreateUserAndLogin("mcp@example.com", "password123")
	require.NoError(t, err)

	projectResp, err := client.POST("/projects", mcpProjectPayload())
	require.NoError(t, err)
	defer projectResp.Body.Close()
	require.Equal(t, http.StatusCreated, projectResp.StatusCode)

	var project map[string]any
	err = json.NewDecoder(projectResp.Body).Decode(&project)
	require.NoError(t, err)

	projectID := project["id"].(string)
	columns := project["columns"].([]any)
	pendingColumnID := columns[0].(map[string]any)["id"].(string)

	taskResp, err := client.POST("/tasks", map[string]any{
		"project_id":        projectID,
		"project_column_id": pendingColumnID,
		"title":             "MCP Task",
		"description":       "Created for MCP",
		"priority":          "medium",
		"tags":              []string{"mcp"},
	})
	require.NoError(t, err)
	defer taskResp.Body.Close()
	require.Equal(t, http.StatusCreated, taskResp.StatusCode)

	var task map[string]any
	err = json.NewDecoder(taskResp.Body).Decode(&task)
	require.NoError(t, err)

	keyResp, err := client.POST("/users/me/mcp-api-keys", map[string]any{
		"name":   "Read Only Agent",
		"scopes": []string{"projects:read", "tasks:read"},
	})
	require.NoError(t, err)
	defer keyResp.Body.Close()
	require.Equal(t, http.StatusCreated, keyResp.StatusCode)

	var created map[string]any
	err = json.NewDecoder(keyResp.Body).Decode(&created)
	require.NoError(t, err)

	rawSecret, ok := created["raw_secret"].(string)
	require.True(t, ok)
	require.NotEmpty(t, rawSecret)

	keyMetadata := created["key"].(map[string]any)
	keyID := keyMetadata["id"].(string)
	assert.Equal(t, "Read Only Agent", keyMetadata["name"])

	listResp, err := client.GET("/users/me/mcp-api-keys")
	require.NoError(t, err)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed []map[string]any
	err = json.NewDecoder(listResp.Body).Decode(&listed)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, keyID, listed[0]["id"])
	_, hasRawSecret := listed[0]["raw_secret"]
	assert.False(t, hasRawSecret)

	initializeResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "initialize",
		"params":  map[string]any{},
	})
	assert.Equal(t, "2025-06-18", initializeResp["result"].(map[string]any)["protocolVersion"])

	initializedResp := postMCP(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	})
	defer initializedResp.Body.Close()
	assert.Equal(t, http.StatusAccepted, initializedResp.StatusCode)

	resourcesResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1.5",
		"method":  "resources/list",
		"params":  map[string]any{},
	})
	assert.Empty(t, resourcesResp["result"].(map[string]any)["resources"])

	listProjectsResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "2",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_projects",
			"arguments": map[string]any{},
		},
	})
	assert.False(t, listProjectsResp["result"].(map[string]any)["isError"] == true)

	missingScopeResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "move_task",
			"arguments": map[string]any{
				"task_id":                  task["id"].(string),
				"project_id":               projectID,
				"target_project_column_id": pendingColumnID,
			},
		},
	})
	result := missingScopeResp["result"].(map[string]any)
	assert.Equal(t, true, result["isError"])
	structured := result["structuredContent"].(map[string]any)
	assert.Equal(t, "missing_scope", structured["error"].(map[string]any)["type"])

	revokeReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/users/me/mcp-api-keys/%s", testAPI.GetBaseURL(), keyID), nil)
	require.NoError(t, err)
	revokeReq.Header.Set("Authorization", "Bearer "+client.Token)
	revokeResp, err := http.DefaultClient.Do(revokeReq)
	require.NoError(t, err)
	defer revokeResp.Body.Close()
	require.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

	reuseBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "4",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_projects",
			"arguments": map[string]any{},
		},
	})
	require.NoError(t, err)

	reuseReq, err := http.NewRequest(http.MethodPost, testAPI.GetBaseURL()+"/mcp", bytes.NewReader(reuseBody))
	require.NoError(t, err)
	reuseReq.Header.Set("Content-Type", "application/json")
	reuseReq.Header.Set("Authorization", "Bearer "+rawSecret)

	reuseResp, err := http.DefaultClient.Do(reuseReq)
	require.NoError(t, err)
	defer reuseResp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, reuseResp.StatusCode)
}
