package projects_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validProjectColumnsPayload() []map[string]any {
	return []map[string]any{
		{
			"name":           "Pending",
			"description":    "Items waiting to be picked up.",
			"color":          "#64748B",
			"is_done_column": false,
		},
		{
			"name":           "Doing",
			"description":    "Items currently being worked on.",
			"color":          "#2563EB",
			"is_done_column": false,
		},
		{
			"name":           "Done",
			"description":    "Completed items.",
			"color":          "#059669",
			"is_done_column": true,
		},
	}
}

func validProjectPayload(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": description,
		"columns":     validProjectColumnsPayload(),
	}
}

func validProjectRepositoryPayload() map[string]any {
	return map[string]any{
		"repository_url":     "https://github.com/acme/project-chat-pubsub",
		"repository_owner":   "acme",
		"repository_name":    "project-chat-pubsub",
		"default_branch":     "main",
		"branch_name_prefix": "task/",
	}
}

func projectColumnsPayloadFromResponse(t *testing.T, rawColumns any) []map[string]any {
	t.Helper()

	columns, ok := rawColumns.([]interface{})
	require.True(t, ok, "expected columns array in project response")

	payload := make([]map[string]any, 0, len(columns))
	for _, rawColumn := range columns {
		column, ok := rawColumn.(map[string]interface{})
		require.True(t, ok, "expected column object in project response")

		payload = append(payload, map[string]any{
			"id":             column["id"],
			"name":           column["name"],
			"description":    column["description"],
			"color":          column["color"],
			"is_done_column": column["is_done_column"],
		})
	}

	return payload
}

func assertProjectColumnsContract(t *testing.T, rawColumns any, expectedProjectID string) {
	t.Helper()

	columns, ok := rawColumns.([]interface{})
	require.True(t, ok, "expected columns array in project response")
	require.Len(t, columns, 3)

	expectedColumns := []struct {
		name         string
		description  string
		color        string
		position     float64
		isDoneColumn bool
	}{
		{name: "Pending", description: "Items waiting to be picked up.", color: "#64748B", position: 0, isDoneColumn: false},
		{name: "Doing", description: "Items currently being worked on.", color: "#2563EB", position: 1, isDoneColumn: false},
		{name: "Done", description: "Completed items.", color: "#059669", position: 2, isDoneColumn: true},
	}

	for i, rawColumn := range columns {
		column, ok := rawColumn.(map[string]interface{})
		require.True(t, ok, "expected column object in project response")

		assert.Contains(t, column, "id")
		assert.Equal(t, expectedProjectID, column["project_id"])
		assert.Equal(t, expectedColumns[i].name, column["name"])
		assert.Equal(t, expectedColumns[i].description, column["description"])
		assert.Equal(t, expectedColumns[i].color, column["color"])
		assert.Equal(t, expectedColumns[i].position, column["position"])
		assert.Equal(t, expectedColumns[i].isDoneColumn, column["is_done_column"])
		assert.Contains(t, column, "created_at")
		assert.Contains(t, column, "updated_at")
	}
}

func assertProjectRepositoryFields(t *testing.T, response map[string]interface{}, expected map[string]any) {
	t.Helper()

	assert.Equal(t, expected["repository_url"], response["repository_url"])
	assert.Equal(t, expected["repository_owner"], response["repository_owner"])
	assert.Equal(t, expected["repository_name"], response["repository_name"])
	assert.Equal(t, expected["default_branch"], response["default_branch"])
	assert.Equal(t, expected["branch_name_prefix"], response["branch_name_prefix"])
}

func TestProjectsEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	t.Run("/projects - endpoints are protected by auth", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())

		resp, err := client.GET("/projects")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.POST("/projects", map[string]string{
			"name":        "Test Project",
			"description": "Test Description",
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.GET("/projects/123e4567-e89b-12d3-a456-426614174000")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.PUT("/projects/123e4567-e89b-12d3-a456-426614174000", map[string]string{
			"name":        "Test Project",
			"description": "Test Description",
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.PATCH("/projects/123e4567-e89b-12d3-a456-426614174000/columns/123e4567-e89b-12d3-a456-426614174001", map[string]string{
			"name":        "Doing",
			"description": "Updated",
			"color":       "#2563EB",
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resp, err = client.POST("/projects/123e4567-e89b-12d3-a456-426614174000/members", map[string]string{
			"email": "member@example.com",
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("POST /projects - create project", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())

		client.CreateUserAndLogin("test@example.com", "password123")

		payload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "Test Project", response["name"])
		assert.Equal(t, "Test Description", response["description"])
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "created_at")
		assert.Contains(t, response, "updated_at")
		assert.Contains(t, response, "members")
		assert.Contains(t, response, "columns")
		assertProjectRepositoryFields(t, response, map[string]any{
			"repository_url":     "",
			"repository_owner":   "",
			"repository_name":    "",
			"default_branch":     "",
			"branch_name_prefix": "",
		})
		assertProjectColumnsContract(t, response["columns"], response["id"].(string))
	})

	t.Run("POST /projects - create project with repository metadata", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		payload := validProjectPayload("Test Project", "Test Description")
		for key, value := range validProjectRepositoryPayload() {
			payload[key] = value
		}

		resp, err := client.POST("/projects", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assertProjectRepositoryFields(t, response, validProjectRepositoryPayload())
	})

	t.Run("POST /projects - create project with invalid fields", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		type testCase struct {
			name           string
			payload        map[string]any
			expectedText   []string
			expectedStatus int
		}
		tests := []testCase{
			{
				name: "missing description",
				payload: map[string]any{
					"name":    "Test Project",
					"columns": validProjectColumnsPayload(),
				},
				expectedText:   []string{"description is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "missing name",
				payload: map[string]any{
					"description": "Test Description",
					"columns":     validProjectColumnsPayload(),
				},
				expectedText:   []string{"name is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "missing name and description",
				payload: map[string]any{
					"columns": validProjectColumnsPayload(),
				},
				expectedText:   []string{"name is required", "description is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "blank name",
				payload: map[string]any{
					"name":        "",
					"description": "Test Description",
					"columns":     validProjectColumnsPayload(),
				},
				expectedText:   []string{"name is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "blank description",
				payload: map[string]any{
					"name":        "a",
					"description": "",
					"columns":     validProjectColumnsPayload(),
				},
				expectedText:   []string{"description is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "missing columns",
				payload: map[string]any{
					"name":        "Test Project",
					"description": "Test Description",
				},
				expectedText:   []string{"at least one column is required", "exactly one done column is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.POST("/projects", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				for _, expectedText := range tc.expectedText {
					assert.Contains(t, string(bodyBytes), expectedText)
				}
			})
		}
	})

	t.Run("GET /projects - list projects", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		payload := validProjectPayload("Test Project", "Test Description")
		resp, err := client.POST("/projects", payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		resp, err = client.GET("/projects")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 1, len(response))
		assert.Equal(t, "Test Project", response[0]["name"])
		assert.Equal(t, "Test Description", response[0]["description"])
		assert.Contains(t, response[0], "id")
		assert.Contains(t, response[0], "created_at")
		assert.Contains(t, response[0], "updated_at")
		assert.Contains(t, response[0], "members")
		assert.Contains(t, response[0], "columns")
		assertProjectRepositoryFields(t, response[0], map[string]any{
			"repository_url":     "",
			"repository_owner":   "",
			"repository_name":    "",
			"default_branch":     "",
			"branch_name_prefix": "",
		})
		assertProjectColumnsContract(t, response[0]["columns"], response[0]["id"].(string))
	})

	t.Run("GET /projects/{id} - get project", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		payload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		resp, err = client.GET("/projects/" + response["id"].(string))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&getResponse)
		require.NoError(t, err)

		assert.Equal(t, "Test Project", getResponse["name"])
		assert.Equal(t, "Test Description", getResponse["description"])
		assert.Contains(t, getResponse, "id")
		assert.Contains(t, getResponse, "created_at")
		assert.Contains(t, getResponse, "updated_at")
		assert.Contains(t, getResponse, "members")
		assert.Contains(t, getResponse, "columns")
		assertProjectRepositoryFields(t, getResponse, map[string]any{
			"repository_url":     "",
			"repository_owner":   "",
			"repository_name":    "",
			"default_branch":     "",
			"branch_name_prefix": "",
		})
		assertProjectColumnsContract(t, getResponse["columns"], getResponse["id"].(string))
	})

	t.Run("PUT /projects/{id} - update project", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		payload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		payload = map[string]any{
			"name":        "Updated Project",
			"description": "Updated Description",
			"columns":     projectColumnsPayloadFromResponse(t, response["columns"]),
		}
		for key, value := range validProjectRepositoryPayload() {
			payload[key] = value
		}

		resp, err = client.PUT("/projects/"+response["id"].(string), payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updateResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&updateResponse)
		require.NoError(t, err)

		assert.Equal(t, "Updated Project", updateResponse["name"])
		assert.Equal(t, "Updated Description", updateResponse["description"])
		assert.Contains(t, updateResponse, "id")
		assert.Contains(t, updateResponse, "created_at")
		assert.Contains(t, updateResponse, "updated_at")
		assert.Contains(t, updateResponse, "members")
		assert.Contains(t, updateResponse, "columns")
		assertProjectRepositoryFields(t, updateResponse, validProjectRepositoryPayload())
		assertProjectColumnsContract(t, updateResponse["columns"], updateResponse["id"].(string))
	})

	t.Run("PUT /projects/{id} - deletes a column and closes the resulting position gap", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		resp, err := client.POST("/projects", validProjectPayload("Test Project", "Test Description"))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&createResponse))

		columns := createResponse["columns"].([]any)
		pending := columns[0].(map[string]any)
		doing := columns[1].(map[string]any)
		done := columns[2].(map[string]any)
		payload := map[string]any{
			"name":        "Test Project",
			"description": "Test Description",
			"columns":     projectColumnsPayloadFromResponse(t, []any{pending, done}),
			"deleted_columns": []map[string]any{
				{"id": doing["id"], "move_tasks_to_column_id": done["id"]},
			},
		}

		resp, err = client.PUT("/projects/"+createResponse["id"].(string), payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var updateResponse map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&updateResponse))
		updatedColumns := updateResponse["columns"].([]any)
		require.Len(t, updatedColumns, 2)
		assert.Equal(t, "Pending", updatedColumns[0].(map[string]any)["name"])
		assert.Equal(t, float64(0), updatedColumns[0].(map[string]any)["position"])
		assert.Equal(t, "Done", updatedColumns[1].(map[string]any)["name"])
		assert.Equal(t, float64(1), updatedColumns[1].(map[string]any)["position"])
	})

	t.Run("PUT /projects/{id} - invalid column deletion target leaves the project unchanged", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		resp, err := client.POST("/projects", validProjectPayload("Test Project", "Test Description"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)

		columns := createResponse["columns"].([]interface{})
		remainingColumns := projectColumnsPayloadFromResponse(t, []interface{}{columns[0], columns[2]})
		removedColumnID := columns[1].(map[string]interface{})["id"]

		payload := map[string]any{
			"name":        "Should Not Persist",
			"description": "Should Not Persist",
			"columns":     remainingColumns,
			"deleted_columns": []map[string]any{
				{"id": removedColumnID, "move_tasks_to_column_id": uuid.New().String()},
			},
		}
		for key, value := range validProjectRepositoryPayload() {
			payload[key] = value
		}

		resp, err = client.PUT("/projects/"+createResponse["id"].(string), payload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		resp, err = client.GET("/projects/" + createResponse["id"].(string))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&getResponse)
		require.NoError(t, err)

		assert.Equal(t, "Test Project", getResponse["name"])
		assert.Equal(t, "Test Description", getResponse["description"])
		assertProjectRepositoryFields(t, getResponse, map[string]any{
			"repository_url":     "",
			"repository_owner":   "",
			"repository_name":    "",
			"default_branch":     "",
			"branch_name_prefix": "",
		})
		assertProjectColumnsContract(t, getResponse["columns"], getResponse["id"].(string))
	})

	t.Run("PATCH /projects/{id}/columns/{column_id} - update single column", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		resp, err := client.POST("/projects", validProjectPayload("Test Project", "Test Description"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)

		columns := createResponse["columns"].([]any)
		firstColumn := columns[0].(map[string]any)

		resp, err = client.PATCH("/projects/"+createResponse["id"].(string)+"/columns/"+firstColumn["id"].(string), map[string]any{
			"name":           "Backlog",
			"description":    "Freshly prioritized work.",
			"color":          "#1D4ED8",
			"is_done_column": false,
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var patchResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&patchResponse)
		require.NoError(t, err)

		assert.Equal(t, firstColumn["id"], patchResponse["id"])
		assert.Equal(t, "Backlog", patchResponse["name"])
		assert.Equal(t, "Freshly prioritized work.", patchResponse["description"])
		assert.Equal(t, "#1D4ED8", patchResponse["color"])
		assert.Equal(t, false, patchResponse["is_done_column"])

		resp, err = client.GET("/projects/" + createResponse["id"].(string))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&getResponse)
		require.NoError(t, err)

		getColumns := getResponse["columns"].([]any)
		updatedColumn := getColumns[0].(map[string]any)
		assert.Equal(t, "Backlog", updatedColumn["name"])
		assert.Equal(t, "Freshly prioritized work.", updatedColumn["description"])
		assert.Equal(t, "#1D4ED8", updatedColumn["color"])
	})

	t.Run("PATCH /projects/{id}/columns/{column_id} - switches done column", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("test@example.com", "password123")

		resp, err := client.POST("/projects", validProjectPayload("Test Project", "Test Description"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)

		columns := createResponse["columns"].([]any)
		firstColumn := columns[0].(map[string]any)
		doneColumn := columns[2].(map[string]any)

		resp, err = client.PATCH("/projects/"+createResponse["id"].(string)+"/columns/"+firstColumn["id"].(string), map[string]any{
			"name":           firstColumn["name"],
			"description":    firstColumn["description"],
			"color":          firstColumn["color"],
			"is_done_column": true,
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var patchResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&patchResponse)
		require.NoError(t, err)
		assert.Equal(t, true, patchResponse["is_done_column"])

		resp, err = client.GET("/projects/" + createResponse["id"].(string))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var getResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&getResponse)
		require.NoError(t, err)

		getColumns := getResponse["columns"].([]any)
		assert.Equal(t, true, getColumns[0].(map[string]any)["is_done_column"])
		assert.Equal(t, doneColumn["id"], getColumns[2].(map[string]any)["id"])
		assert.Equal(t, false, getColumns[2].(map[string]any)["is_done_column"])
	})

	t.Run("PATCH /projects/{id}/columns/{column_id} - forbidden for non-creator", func(t *testing.T) {
		testAPI.TruncateTables(t)

		ownerClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		ownerClient.CreateUserAndLogin("owner@example.com", "password123")

		resp, err := ownerClient.POST("/projects", validProjectPayload("Test Project", "Test Description"))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]any
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		require.NoError(t, err)

		memberClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		memberClient.CreateUserAndLogin("member@example.com", "password123")

		resp, err = ownerClient.POST("/projects/"+createResponse["id"].(string)+"/members", map[string]string{
			"email": "member@example.com",
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		firstColumn := createResponse["columns"].([]any)[0].(map[string]any)
		resp, err = memberClient.PATCH("/projects/"+createResponse["id"].(string)+"/columns/"+firstColumn["id"].(string), map[string]any{
			"name":           "Backlog",
			"description":    "Freshly prioritized work.",
			"color":          "#1D4ED8",
			"is_done_column": false,
		})
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("POST /projects/{id}/members - create member", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())

		client.CreateUserAndLogin("owner@example.com", "password123")

		projectPayload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", projectPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var projectResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&projectResponse)
		require.NoError(t, err)
		projectId := projectResponse["id"].(string)

		client2 := shared.NewHTTPClient(testAPI.GetBaseURL())
		client2.CreateUserAndLogin("member@example.com", "password123")

		memberPayload := map[string]string{
			"email": "member@example.com",
		}

		resp, err = client.POST("/projects/"+projectId+"/members", memberPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var memberResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&memberResponse)
		require.NoError(t, err)

		assert.Contains(t, memberResponse, "id")
		assert.Equal(t, projectId, memberResponse["project_id"])
		assert.Contains(t, memberResponse, "user_id")
		assert.Equal(t, "member", memberResponse["role"])
	})

	t.Run("POST /projects/{id}/members - create member with invalid fields", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("owner@example.com", "password123")

		projectPayload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", projectPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var projectResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&projectResponse)
		require.NoError(t, err)
		projectId := projectResponse["id"].(string)

		type testCase struct {
			name           string
			payload        map[string]string
			expectedText   []string
			expectedStatus int
		}

		tests := []testCase{
			{
				name:           "missing email",
				payload:        map[string]string{},
				expectedText:   []string{"email is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "blank email",
				payload: map[string]string{
					"email": "",
				},
				expectedText:   []string{"email is required"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
			{
				name: "invalid email format",
				payload: map[string]string{
					"email": "invalid-email",
				},
				expectedText:   []string{"email is invalid"},
				expectedStatus: http.StatusUnprocessableEntity,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.POST("/projects/"+projectId+"/members", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				for _, expectedText := range tc.expectedText {
					assert.Contains(t, string(bodyBytes), expectedText)
				}
			})
		}
	})

	t.Run("POST /projects/{id}/members - create member business validation errors", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("owner@example.com", "password123")

		projectPayload := validProjectPayload("Test Project", "Test Description")

		resp, err := client.POST("/projects", projectPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var projectResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&projectResponse)
		require.NoError(t, err)
		projectId := projectResponse["id"].(string)

		type testCase struct {
			name           string
			email          string
			expectedStatus int
			expectedText   string
		}

		tests := []testCase{
			{
				name:           "user not found",
				email:          "nonexistent@example.com",
				expectedStatus: http.StatusNotFound,
				expectedText:   "user not found",
			},
			{
				name:           "cannot add yourself",
				email:          "owner@example.com",
				expectedStatus: http.StatusUnprocessableEntity,
				expectedText:   "you cannot add yourself as a member",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				memberPayload := map[string]string{
					"email": tc.email,
				}

				resp, err := client.POST("/projects/"+projectId+"/members", memberPayload)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)

				bodyBytes, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Contains(t, string(bodyBytes), tc.expectedText)
			})
		}
	})

	t.Run("POST /projects/{id}/members - create member with invalid project id", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		client.CreateUserAndLogin("owner@example.com", "password123")

		memberPayload := map[string]string{
			"email": "member@example.com",
		}

		resp, err := client.POST("/projects/invalid-id/members", memberPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		bodyBytes, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(bodyBytes), "invalid project id")
	})

	t.Run("GET /projects/{id}/activities returns the actual project column for tasks", func(t *testing.T) {
		testAPI.TruncateTables(t)

		client := shared.NewHTTPClient(testAPI.GetBaseURL())
		_, err := client.CreateUserAndLogin("activities@example.com", "password123")
		require.NoError(t, err)

		actorID := createProjectTestUser(t, testAPI, "activity-actor@example.com")
		userID := getProjectTestCurrentUserID(t, client)
		projectID := uuid.New()
		projectColumnID := uuid.New()
		taskID := uuid.New()
		activityID := uuid.New()
		now := time.Now().UTC()

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO projects (id, user_id, name, description, created_at, updated_at)
			VALUES ($1, $2, 'Activity Project', 'Project Description', $3, $3)
		`, projectID, actorID, now)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO project_members (id, user_id, project_id, role)
			VALUES ($1, $2, $3, 'creator'), ($4, $5, $3, 'member')
		`, uuid.New(), actorID, projectID, uuid.New(), userID)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO project_columns (id, project_id, name, color, position, is_done_column, created_at, updated_at)
			VALUES ($1, $2, 'QA Review', '#F59E0B', 3, false, $3, $3)
		`, projectColumnID, projectID, now)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO tasks (id, project_id, title, description, project_column_id, author_id, priority, task_order, created_at, updated_at)
			VALUES ($1, $2, 'Task title', 'Task description', $3, $4, 'medium', '500000000000', $5, $5)
		`, taskID, projectID, projectColumnID, actorID, now)
		require.NoError(t, err)

		_, err = testAPI.DB.Exec(context.Background(), `
			INSERT INTO project_activity_logs (id, project_id, actor_id, activity_type, activity_data, entity_type, entity_id, created_at, updated_at)
			VALUES ($1, $2, $3, 'task.updated', '{}'::jsonb, 'task', $4, $5, $5)
		`, activityID, projectID, actorID, taskID, now)
		require.NoError(t, err)

		resp, err := client.GET("/projects/" + projectID.String() + "/activities")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))

		data, ok := response["data"].([]any)
		require.True(t, ok)
		require.Len(t, data, 1)

		activity, ok := data[0].(map[string]any)
		require.True(t, ok)
		task, ok := activity["task"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "qa review", task["status"])

		projectColumn, ok := task["project_column"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "QA Review", projectColumn["name"])
		assert.Equal(t, "#F59E0B", projectColumn["color"])
	})
}

func createProjectTestUser(t *testing.T, testAPI *shared.TestAPI, email string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	_, err := testAPI.DB.Exec(context.Background(), `
		INSERT INTO users (id, name, email, password)
		VALUES ($1, 'Actor', $2, 'password123')
	`, id, email)
	require.NoError(t, err)

	return id
}

func getProjectTestCurrentUserID(t *testing.T, client *shared.HTTPClient) uuid.UUID {
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
