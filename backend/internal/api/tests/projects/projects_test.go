package projects_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validProjectColumnsPayload() []map[string]any {
	return []map[string]any{
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
	}
}

func validProjectPayload(name string, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"description": description,
		"columns":     validProjectColumnsPayload(),
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
		color        string
		position     float64
		isDoneColumn bool
	}{
		{name: "Pending", color: "#64748B", position: 0, isDoneColumn: false},
		{name: "Doing", color: "#2563EB", position: 1, isDoneColumn: false},
		{name: "Done", color: "#059669", position: 2, isDoneColumn: true},
	}

	for i, rawColumn := range columns {
		column, ok := rawColumn.(map[string]interface{})
		require.True(t, ok, "expected column object in project response")

		assert.Contains(t, column, "id")
		assert.Equal(t, expectedProjectID, column["project_id"])
		assert.Equal(t, expectedColumns[i].name, column["name"])
		assert.Equal(t, expectedColumns[i].color, column["color"])
		assert.Equal(t, expectedColumns[i].position, column["position"])
		assert.Equal(t, expectedColumns[i].isDoneColumn, column["is_done_column"])
		assert.Contains(t, column, "created_at")
		assert.Contains(t, column, "updated_at")
	}
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
		assertProjectColumnsContract(t, response["columns"], response["id"].(string))
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
		assertProjectColumnsContract(t, updateResponse["columns"], updateResponse["id"].(string))
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
}
