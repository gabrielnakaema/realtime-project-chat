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
	doneColumnID := columns[1].(map[string]any)["id"].(string)

	taskResp, err := client.POST("/tasks", map[string]any{
		"project_id":        projectID,
		"project_column_id": pendingColumnID,
		"title":             "MCP Task",
		"description":       "Created for MCP",
		"code":              "MCP-1",
		"priority":          "medium",
		"tags":              []string{"mcp"},
	})
	require.NoError(t, err)
	defer taskResp.Body.Close()
	require.Equal(t, http.StatusCreated, taskResp.StatusCode)

	var task map[string]any
	err = json.NewDecoder(taskResp.Body).Decode(&task)
	require.NoError(t, err)

	scopeResp, err := client.GET("/users/me/mcp-api-keys/scopes")
	require.NoError(t, err)
	defer scopeResp.Body.Close()
	require.Equal(t, http.StatusOK, scopeResp.StatusCode)

	var availableScopes []map[string]any
	err = json.NewDecoder(scopeResp.Body).Decode(&availableScopes)
	require.NoError(t, err)
	require.NotEmpty(t, availableScopes)
	assert.Contains(t, availableScopes, map[string]any{
		"scope": "projects:board:read",
		"label": "Read project boards",
		"title": "Inspect board columns and the tasks grouped inside them.",
	})
	assert.Contains(t, availableScopes, map[string]any{
		"scope": "tasks:create",
		"label": "Create tasks",
		"title": "Create new tasks in project columns you can access.",
	})
	assert.Contains(t, availableScopes, map[string]any{
		"scope": "tasks:update",
		"label": "Update tasks",
		"title": "Edit task details such as title, description, assignee, due date, and tags.",
	})
	assert.Contains(t, availableScopes, map[string]any{
		"scope": "tasks:comments:read",
		"label": "Read task comments",
		"title": "Load recent task comments when an MCP client requests extra context.",
	})

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

	editorKeyResp, err := client.POST("/users/me/mcp-api-keys", map[string]any{
		"name": "Editor Agent",
		"scopes": []string{
			"projects:board:read",
			"tasks:create",
			"tasks:read",
			"tasks:update",
			"tasks:comments:read",
			"tasks:comment",
		},
	})
	require.NoError(t, err)
	defer editorKeyResp.Body.Close()
	require.Equal(t, http.StatusCreated, editorKeyResp.StatusCode)

	var editorCreated map[string]any
	err = json.NewDecoder(editorKeyResp.Body).Decode(&editorCreated)
	require.NoError(t, err)

	editorRawSecret, ok := editorCreated["raw_secret"].(string)
	require.True(t, ok)
	require.NotEmpty(t, editorRawSecret)

	keyMetadata := created["key"].(map[string]any)
	keyID := keyMetadata["id"].(string)
	assert.Equal(t, "Read Only Agent", keyMetadata["name"])

	updateResp, err := client.PUT(fmt.Sprintf("/users/me/mcp-api-keys/%s", keyID), map[string]any{
		"name": "  Board mover  ",
		"scopes": []string{
			"projects:read",
			"tasks:read",
			"tasks:move",
			"tasks:move",
		},
	})
	require.NoError(t, err)
	defer updateResp.Body.Close()
	require.Equal(t, http.StatusOK, updateResp.StatusCode)

	var updatedKey map[string]any
	err = json.NewDecoder(updateResp.Body).Decode(&updatedKey)
	require.NoError(t, err)
	assert.Equal(t, "Board mover", updatedKey["name"])
	assert.Equal(t, []any{"projects:read", "tasks:move", "tasks:read"}, updatedKey["scopes"])

	invalidScopeResp, err := client.PUT(fmt.Sprintf("/users/me/mcp-api-keys/%s", keyID), map[string]any{
		"name":   "Invalid key",
		"scopes": []string{"tasks:read", "tasks:teleport"},
	})
	require.NoError(t, err)
	defer invalidScopeResp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, invalidScopeResp.StatusCode)

	var invalidScopeBody map[string]any
	err = json.NewDecoder(invalidScopeResp.Body).Decode(&invalidScopeBody)
	require.NoError(t, err)
	assert.Equal(t, "invalid mcp api key scope", invalidScopeBody["message"])

	emptyScopeResp, err := client.PUT(fmt.Sprintf("/users/me/mcp-api-keys/%s", keyID), map[string]any{
		"name":   "Empty scopes",
		"scopes": []string{},
	})
	require.NoError(t, err)
	defer emptyScopeResp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, emptyScopeResp.StatusCode)

	var emptyScopeBody map[string]any
	err = json.NewDecoder(emptyScopeResp.Body).Decode(&emptyScopeBody)
	require.NoError(t, err)
	assert.Equal(t, "Validation Failed", emptyScopeBody["message"])

	otherClient := shared.NewHTTPClient(testAPI.GetBaseURL())
	_, err = otherClient.CreateUserAndLogin("mcp-other@example.com", "password123")
	require.NoError(t, err)

	notFoundUpdateResp, err := otherClient.PUT(fmt.Sprintf("/users/me/mcp-api-keys/%s", keyID), map[string]any{
		"name":   "Other user edit",
		"scopes": []string{"tasks:read"},
	})
	require.NoError(t, err)
	defer notFoundUpdateResp.Body.Close()
	require.Equal(t, http.StatusNotFound, notFoundUpdateResp.StatusCode)

	listResp, err := client.GET("/users/me/mcp-api-keys")
	require.NoError(t, err)
	defer listResp.Body.Close()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	var listed []map[string]any
	err = json.NewDecoder(listResp.Body).Decode(&listed)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	foundReadOnlyKey := false
	for _, listedKey := range listed {
		if listedKey["id"] == keyID {
			foundReadOnlyKey = true
		}
		_, hasRawSecret := listedKey["raw_secret"]
		assert.False(t, hasRawSecret)
	}
	assert.True(t, foundReadOnlyKey)

	initializeResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "initialize",
		"params":  map[string]any{},
	})
	initializeResult := initializeResp["result"].(map[string]any)
	assert.Equal(t, "2025-06-18", initializeResult["protocolVersion"])
	assert.Contains(t, initializeResult["instructions"], "project-chat://server/guide")

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
	resources := resourcesResp["result"].(map[string]any)["resources"].([]any)
	require.Len(t, resources, 2)
	assert.Equal(t, "project-chat://server/guide", resources[0].(map[string]any)["uri"])

	readGuideResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1.6",
		"method":  "resources/read",
		"params": map[string]any{
			"uri": "project-chat://server/guide",
		},
	})
	guideContents := readGuideResp["result"].(map[string]any)["contents"].([]any)
	require.Len(t, guideContents, 1)
	guideText := guideContents[0].(map[string]any)["text"].(string)
	assert.Contains(t, guideText, "Recommended workflow")
	assert.Contains(t, guideText, "list_projects")
	assert.NotContains(t, guideText, "create_task")

	toolsResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1.75",
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	tools := toolsResp["result"].(map[string]any)["tools"].([]any)
	toolNames := make([]string, 0, len(tools))
	for _, rawTool := range tools {
		tool := rawTool.(map[string]any)
		toolNames = append(toolNames, tool["name"].(string))
		assert.NotEmpty(t, tool["title"])
		assert.NotNil(t, tool["outputSchema"])
	}
	assert.Contains(t, toolNames, "list_projects")
	assert.Contains(t, toolNames, "get_task")
	assert.NotContains(t, toolNames, "create_task")
	assert.NotContains(t, toolNames, "list_task_comments")

	editorToolsResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "1.8",
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	editorTools := editorToolsResp["result"].(map[string]any)["tools"].([]any)
	editorToolNames := make([]string, 0, len(editorTools))
	for _, rawTool := range editorTools {
		editorToolNames = append(editorToolNames, rawTool.(map[string]any)["name"].(string))
	}
	assert.Contains(t, editorToolNames, "create_task")
	assert.Contains(t, editorToolNames, "update_task")
	assert.Contains(t, editorToolNames, "list_task_comments")

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
	listProjectsResult := listProjectsResp["result"].(map[string]any)
	listProjectsContent := listProjectsResult["content"].([]any)
	assert.Contains(t, listProjectsContent[0].(map[string]any)["text"], "Listed 1 visible project")
	assert.Contains(t, listProjectsContent[1].(map[string]any)["text"], `"projects"`)
	assert.NotNil(t, listProjectsResult["structuredContent"])

	missingScopeResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "add_task_comment",
			"arguments": map[string]any{
				"task_id": task["id"].(string),
				"content": "This should fail without comment scope",
			},
		},
	})
	result := missingScopeResp["result"].(map[string]any)
	assert.Equal(t, true, result["isError"])
	structured := result["structuredContent"].(map[string]any)
	assert.Equal(t, "missing_scope", structured["error"].(map[string]any)["type"])

	moveTaskResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3.25",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "move_task",
			"arguments": map[string]any{
				"task_id":                  task["id"].(string),
				"project_id":               projectID,
				"target_project_column_id": doneColumnID,
			},
		},
	})
	moveTaskStructured := moveTaskResp["result"].(map[string]any)["structuredContent"].(map[string]any)
	assert.Equal(t, doneColumnID, moveTaskStructured["task"].(map[string]any)["project_column_id"])

	missingCommentsScopeResp := createMCPRequest(t, testAPI.GetBaseURL(), rawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3.5",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_task",
			"arguments": map[string]any{
				"task_id":          task["id"].(string),
				"include_comments": true,
			},
		},
	})
	missingCommentsResult := missingCommentsScopeResp["result"].(map[string]any)
	assert.Equal(t, true, missingCommentsResult["isError"])
	missingCommentsStructured := missingCommentsResult["structuredContent"].(map[string]any)
	assert.Equal(t, "missing_scope", missingCommentsStructured["error"].(map[string]any)["type"])

	boardResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3.75",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_project_board",
			"arguments": map[string]any{
				"project_id": projectID,
			},
		},
	})
	boardStructured := boardResp["result"].(map[string]any)["structuredContent"].(map[string]any)
	assert.Equal(t, projectID, boardStructured["project"].(map[string]any)["id"])

	createTaskResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3.8",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_task",
			"arguments": map[string]any{
				"project_id":        projectID,
				"project_column_id": pendingColumnID,
				"title":             "MCP Created Task",
				"description":       "Created through MCP",
				"code":              "MCP-2",
				"priority":          "high",
				"tags":              []string{"mcp", "created"},
			},
		},
	})
	createdTask := createTaskResp["result"].(map[string]any)["structuredContent"].(map[string]any)["task"].(map[string]any)
	createdTaskID := createdTask["id"].(string)
	assert.Equal(t, "MCP Created Task", createdTask["title"])
	assert.Equal(t, "MCP-2", createdTask["code"])
	assert.Contains(t, createTaskResp["result"].(map[string]any)["content"].([]any)[0].(map[string]any)["text"], "MCP Created Task")

	updateTaskResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "3.9",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "update_task",
			"arguments": map[string]any{
				"task_id":           createdTaskID,
				"project_column_id": pendingColumnID,
				"title":             "MCP Updated Task",
				"description":       "Updated through MCP",
				"code":              "MCP-3",
				"priority":          "medium",
				"tags":              []string{"mcp", "updated"},
			},
		},
	})
	updatedTask := updateTaskResp["result"].(map[string]any)["structuredContent"].(map[string]any)["task"].(map[string]any)
	assert.Equal(t, "MCP Updated Task", updatedTask["title"])
	assert.Equal(t, "MCP-3", updatedTask["code"])
	assert.Contains(t, updateTaskResp["result"].(map[string]any)["content"].([]any)[0].(map[string]any)["text"], "MCP Updated Task")

	addCommentResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "4.0",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "add_task_comment",
			"arguments": map[string]any{
				"task_id": createdTaskID,
				"content": "Comment added through MCP",
			},
		},
	})
	addedComment := addCommentResp["result"].(map[string]any)["structuredContent"].(map[string]any)["comment"].(map[string]any)
	assert.Equal(t, "Comment added through MCP", addedComment["content"])

	listCommentsResp := createMCPRequest(t, testAPI.GetBaseURL(), editorRawSecret, map[string]any{
		"jsonrpc": "2.0",
		"id":      "4.1",
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_task_comments",
			"arguments": map[string]any{
				"task_id": createdTaskID,
				"limit":   10,
			},
		},
	})
	comments := listCommentsResp["result"].(map[string]any)["structuredContent"].(map[string]any)["comments"].([]any)
	require.NotEmpty(t, comments)
	assert.Equal(t, "Comment added through MCP", comments[0].(map[string]any)["content"])

	revokeReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/users/me/mcp-api-keys/%s", testAPI.GetBaseURL(), keyID), nil)
	require.NoError(t, err)
	revokeReq.Header.Set("Authorization", "Bearer "+client.Token)
	revokeResp, err := http.DefaultClient.Do(revokeReq)
	require.NoError(t, err)
	defer revokeResp.Body.Close()
	require.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

	revokedUpdateResp, err := client.PUT(fmt.Sprintf("/users/me/mcp-api-keys/%s", keyID), map[string]any{
		"name":   "Revoked edit",
		"scopes": []string{"projects:read", "tasks:read"},
	})
	require.NoError(t, err)
	defer revokedUpdateResp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, revokedUpdateResp.StatusCode)

	var revokedUpdateBody map[string]any
	err = json.NewDecoder(revokedUpdateResp.Body).Decode(&revokedUpdateBody)
	require.NoError(t, err)
	assert.Equal(t, "mcp api key is already revoked", revokedUpdateBody["message"])

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
