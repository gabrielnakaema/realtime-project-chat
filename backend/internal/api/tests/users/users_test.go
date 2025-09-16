package users_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	client := shared.NewHTTPClient(testAPI.GetBaseURL())

	t.Run("POST /users - create user", func(t *testing.T) {
		payload := map[string]string{
			"email":    "newuser@example.com",
			"password": "securepassword123",
			"name":     "New User",
		}

		resp, err := client.POST("/users", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "newuser@example.com", response["email"])
		assert.Equal(t, "New User", response["name"])
		assert.Contains(t, response, "id")
		assert.Contains(t, response, "created_at")

		assert.NotContains(t, response, "password")
	})

	t.Run("POST /users - duplicate email", func(t *testing.T) {
		payload := map[string]string{
			"email":    "duplicate@example.com",
			"password": "password123",
			"name":     "First User",
		}

		resp, err := client.POST("/users", payload)
		require.NoError(t, err)
		resp.Body.Close()

		// Try to create user with same email
		payload["name"] = "Second User"
		resp, err = client.POST("/users", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("POST /users - invalid email", func(t *testing.T) {
		payload := map[string]string{
			"email":    "not-an-email",
			"password": "password123",
			"name":     "Test User",
		}

		resp, err := client.POST("/users", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("POST /users - missing required fields", func(t *testing.T) {
		testCases := []struct {
			name    string
			payload map[string]string
		}{
			{
				name: "missing email",
				payload: map[string]string{
					"password": "password123",
					"name":     "Test User",
				},
			},
			{
				name: "missing password",
				payload: map[string]string{
					"email": "test@example.com",
					"name":  "Test User",
				},
			},
			{
				name: "missing name",
				payload: map[string]string{
					"email":    "test@example.com",
					"password": "password123",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.POST("/users", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			})
		}
	})

	t.Run("POST /users - weak password", func(t *testing.T) {
		payload := map[string]string{
			"email":    "weak@example.com",
			"password": "123",
			"name":     "Weak Password User",
		}

		resp, err := client.POST("/users", payload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})
}
