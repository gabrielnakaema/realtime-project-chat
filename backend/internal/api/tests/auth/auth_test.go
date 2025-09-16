package auth_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/api/tests/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthEndpoints(t *testing.T) {
	testAPI, cleanup := shared.SetupTestAPI(t)
	defer cleanup()

	client := shared.NewHTTPClient(testAPI.GetBaseURL())

	t.Run("POST /auth/login", func(t *testing.T) {
		createUserPayload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Test User",
		}

		resp, err := client.POST("/users", createUserPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		loginPayload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}

		loginResp, err := client.POST("/auth/login", loginPayload)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		assert.Equal(t, http.StatusOK, loginResp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(loginResp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "user")
		assert.NotEmpty(t, response["access_token"])
		assert.NotEmpty(t, response["refresh_token"])
	})

	t.Run("POST /auth/login invalid credentials", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "wrongpassword",
		}

		resp, err := client.POST("/auth/login", loginPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("POST /auth/login incorrect password", func(t *testing.T) {
		createUserPayload := map[string]string{
			"email":    "test2@example.com",
			"password": "password123",
			"name":     "Test User",
		}

		resp, err := client.POST("/users", createUserPayload)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		loginPayload := map[string]string{
			"email":    createUserPayload["email"],
			"password": "wrongpassword",
		}

		resp, err = client.POST("/auth/login", loginPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("POST /auth/login validation errors", func(t *testing.T) {
		type testCase struct {
			name           string
			payload        map[string]string
			expectedStatus int
		}

		validationErrorStatus := http.StatusUnprocessableEntity

		testCases := []testCase{
			{
				name: "no email",
				payload: map[string]string{
					"password": "password123",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "no password",
				payload: map[string]string{
					"email": "test@example.com",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "empty email",
				payload: map[string]string{
					"email":    "",
					"password": "password123",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "empty password",
				payload: map[string]string{
					"email":    "test@example.com",
					"password": "",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "invalid email",
				payload: map[string]string{
					"email":    "invalid",
					"password": "password123",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "invalid email missing text after @",
				payload: map[string]string{
					"email":    "test@",
					"password": "password123",
				},
				expectedStatus: validationErrorStatus,
			},
			{
				name:           "no email and no password",
				payload:        map[string]string{},
				expectedStatus: validationErrorStatus,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.POST("/auth/login", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			})
		}
	})

	t.Run("POST /auth/refresh-token", func(t *testing.T) {
		loginPayload := map[string]string{
			"email":    "refresh@example.com",
			"password": "password123",
		}

		createUserPayload := map[string]string{
			"email":    "refresh@example.com",
			"password": "password123",
			"name":     "Refresh Test User",
		}

		resp, err := client.POST("/users", createUserPayload)
		require.NoError(t, err)
		resp.Body.Close()

		loginResp, err := client.POST("/auth/login", loginPayload)
		require.NoError(t, err)
		defer loginResp.Body.Close()

		var loginResponse map[string]interface{}
		err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		require.NoError(t, err)

		refreshPayload := map[string]string{
			"refresh_token": loginResponse["refresh_token"].(string),
		}

		refreshResp, err := client.POST("/auth/refresh-token", refreshPayload)
		require.NoError(t, err)
		defer refreshResp.Body.Close()

		assert.Equal(t, http.StatusOK, refreshResp.StatusCode)

		var refreshResponse map[string]interface{}
		err = json.NewDecoder(refreshResp.Body).Decode(&refreshResponse)
		require.NoError(t, err)

		assert.Contains(t, refreshResponse, "access_token")
		assert.NotEmpty(t, refreshResponse["access_token"])

		// If same token is used again, it should return 401
		secondRefreshResp, err := client.POST("/auth/refresh-token", refreshPayload)
		require.NoError(t, err)
		defer secondRefreshResp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, secondRefreshResp.StatusCode)
	})

	t.Run("POST /auth/refresh-token validation errors", func(t *testing.T) {
		type testCase struct {
			name           string
			payload        map[string]string
			expectedStatus int
		}

		validationErrorStatus := http.StatusUnprocessableEntity

		testCases := []testCase{
			{
				name:           "no refresh token",
				payload:        map[string]string{},
				expectedStatus: validationErrorStatus,
			},
			{
				name: "empty refresh token",
				payload: map[string]string{
					"refresh_token": "",
				},
				expectedStatus: validationErrorStatus,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := client.POST("/auth/refresh-token", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()
				assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			})
		}
	})

	t.Run("POST /auth/refresh-token invalid refresh token", func(t *testing.T) {
		refreshPayload := map[string]string{
			"refresh_token": "invalid-refresh-token",
		}

		resp, err := client.POST("/auth/refresh-token", refreshPayload)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
