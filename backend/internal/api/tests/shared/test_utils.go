package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPClient struct {
	BaseURL string
	Token   string
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
	}
}

func (c *HTTPClient) SetAuthToken(token string) {
	c.Token = token
}

func (c *HTTPClient) POST(endpoint string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return http.DefaultClient.Do(req)
}

func (c *HTTPClient) GET(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return http.DefaultClient.Do(req)
}

func (c *HTTPClient) PUT(endpoint string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", c.BaseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return http.DefaultClient.Do(req)
}

func (c *HTTPClient) DELETE(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return http.DefaultClient.Do(req)
}

func (c *HTTPClient) CreateUserAndLogin(email, password string) (string, error) {
	createPayload := map[string]string{
		"email":    email,
		"password": password,
		"name":     "Test User",
	}

	createResp, err := c.POST("/users", createPayload)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create user: status %d", createResp.StatusCode)
	}

	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}

	loginResp, err := c.POST("/auth/login", loginPayload)
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login: status %d", loginResp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(loginResp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	token, ok := response["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in login response")
	}

	c.SetAuthToken(token)
	return token, nil
}
