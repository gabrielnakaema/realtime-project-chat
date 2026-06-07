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

	t.Run("PUT /users/me/password - unauthorized", func(t *testing.T) {
		resp, err := client.PUT("/users/me/password", map[string]string{
			"old_password":              "password123",
			"new_password":              "newpassword123",
			"new_password_confirmation": "newpassword123",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("PUT /users/me/password - success", func(t *testing.T) {
		authenticatedClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		_, err := authenticatedClient.CreateUserAndLogin("changepassword@example.com", "password123")
		require.NoError(t, err)

		resp, err := authenticatedClient.PUT("/users/me/password", map[string]string{
			"old_password":              "password123",
			"new_password":              "newpassword123",
			"new_password_confirmation": "newpassword123",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		oldPasswordClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		oldPasswordResp, err := oldPasswordClient.POST("/auth/login", map[string]string{
			"email":    "changepassword@example.com",
			"password": "password123",
		})
		require.NoError(t, err)
		defer oldPasswordResp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, oldPasswordResp.StatusCode)

		newPasswordClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		newPasswordResp, err := newPasswordClient.POST("/auth/login", map[string]string{
			"email":    "changepassword@example.com",
			"password": "newpassword123",
		})
		require.NoError(t, err)
		defer newPasswordResp.Body.Close()
		assert.Equal(t, http.StatusOK, newPasswordResp.StatusCode)
	})

	t.Run("PUT /users/me/password - incorrect old password", func(t *testing.T) {
		authenticatedClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		_, err := authenticatedClient.CreateUserAndLogin("wrongoldpassword@example.com", "password123")
		require.NoError(t, err)

		resp, err := authenticatedClient.PUT("/users/me/password", map[string]string{
			"old_password":              "wrongpassword",
			"new_password":              "newpassword123",
			"new_password_confirmation": "newpassword123",
		})
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

		stillOldPasswordClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		loginResp, err := stillOldPasswordClient.POST("/auth/login", map[string]string{
			"email":    "wrongoldpassword@example.com",
			"password": "password123",
		})
		require.NoError(t, err)
		defer loginResp.Body.Close()
		assert.Equal(t, http.StatusOK, loginResp.StatusCode)
	})

	t.Run("PUT /users/me/password - validation errors", func(t *testing.T) {
		authenticatedClient := shared.NewHTTPClient(testAPI.GetBaseURL())
		_, err := authenticatedClient.CreateUserAndLogin("passwordvalidation@example.com", "password123")
		require.NoError(t, err)

		testCases := []struct {
			name    string
			payload map[string]string
		}{
			{
				name: "short new password",
				payload: map[string]string{
					"old_password":              "password123",
					"new_password":              "123",
					"new_password_confirmation": "123",
				},
			},
			{
				name: "confirmation mismatch",
				payload: map[string]string{
					"old_password":              "password123",
					"new_password":              "newpassword123",
					"new_password_confirmation": "differentpassword123",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := authenticatedClient.PUT("/users/me/password", tc.payload)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			})
		}
	})
}
