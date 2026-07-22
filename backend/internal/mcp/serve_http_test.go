package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/gabrielnakaema/project-chat/internal/platform/apperr"
	"github.com/gabrielnakaema/project-chat/internal/platform/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeJSON(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}

type stubAuthService struct {
	result  *AuthenticatedAPIKey
	err     error
	called  bool
	secrets []string
}

func (s *stubAuthService) Authenticate(ctx context.Context, bearerSecret string) (*AuthenticatedAPIKey, error) {
	s.called = true
	s.secrets = append(s.secrets, bearerSecret)
	return s.result, s.err
}

type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, record.Clone())
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

func (h *captureHandler) attrsFor(event string) map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := len(h.records) - 1; i >= 0; i-- {
		record := h.records[i]
		if record.Message != event {
			continue
		}
		attrs := map[string]any{}
		record.Attrs(func(a slog.Attr) bool {
			attrs[a.Key] = a.Value.Any()
			return true
		})
		return attrs
	}
	return nil
}

func newTestHandler(auth authService) *Handler {
	return NewHandler(auth, &stubProjectService{}, &stubTaskService{}, &stubTaskCommentService{})
}

func newAuthedRequest(t *testing.T, capture *captureHandler, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")

	ctx := req.Context()
	if capture != nil {
		ctx = logger.WithLogger(ctx, slog.New(capture))
	}
	return req.WithContext(ctx)
}

func fullScopePrincipalKey() *AuthenticatedAPIKey {
	return &AuthenticatedAPIKey{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Scopes: domain.AllowedMCPAPIScopes,
	}
}

func TestServeHTTP_NonPOSTMethodRejected(t *testing.T) {
	auth := &stubAuthService{}
	handler := newTestHandler(auth)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, http.MethodPost, rec.Header().Get("Allow"))
	assert.Equal(t, 0, rec.Body.Len(), "non-POST requests must not process a body")
	assert.False(t, auth.called, "auth must not be attempted for a rejected method")
}

func TestServeHTTP_AuthorizationHeaderValidation(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header", header: ""},
		{name: "missing bearer prefix", header: "Token secret"},
		{name: "empty secret", header: "Bearer "},
		{name: "whitespace-only secret", header: "Bearer    "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := &stubAuthService{}
			handler := newTestHandler(auth)

			req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, auth.called, "authService should not be invoked without a valid bearer secret")

			var body map[string]any
			require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
			assert.Equal(t, float64(http.StatusUnauthorized), body["status"])
			assert.Equal(t, "unauthorized", body["message"])
		})
	}
}

func TestServeHTTP_AuthenticateDomainErrorMapsToHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "unauthorized", err: apperr.UnauthorizedError("bad secret"), wantStatus: http.StatusUnauthorized},
		{name: "forbidden", err: apperr.ForbiddenError("blocked"), wantStatus: http.StatusForbidden},
		{name: "not found", err: apperr.NotFoundError("key not found"), wantStatus: http.StatusNotFound},
		{name: "business validation", err: apperr.BusinessValidationError("bad state"), wantStatus: http.StatusUnprocessableEntity},
		{name: "unmapped domain error defaults to 500", err: apperr.ServerError("boom", nil), wantStatus: http.StatusInternalServerError},
		{name: "non-domain error defaults to 500", err: assertErr{}, wantStatus: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth := &stubAuthService{err: tc.err}
			handler := newTestHandler(auth)

			req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

type assertErr struct{}

func (assertErr) Error() string { return "generic failure" }

func TestServeHTTP_MalformedJSONBodyReturns400(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)

	req := newAuthedRequest(t, nil, `{not-json`)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestServeHTTP_Notifications(t *testing.T) {
	t.Run("notifications/initialized is accepted", func(t *testing.T) {
		auth := &stubAuthService{result: fullScopePrincipalKey()}
		handler := newTestHandler(auth)
		capture := &captureHandler{}

		req := newAuthedRequest(t, capture, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Equal(t, 0, rec.Body.Len())

		attrs := capture.attrsFor("mcp_notification")
		require.NotNil(t, attrs)
		assert.Equal(t, "notifications/initialized", attrs["tool_name"])
		assert.Equal(t, "accepted", attrs["status"])
	})

	t.Run("unknown notification method is ignored but still accepted", func(t *testing.T) {
		auth := &stubAuthService{result: fullScopePrincipalKey()}
		handler := newTestHandler(auth)
		capture := &captureHandler{}

		req := newAuthedRequest(t, capture, `{"jsonrpc":"2.0","method":"notifications/mystery"}`)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Equal(t, 0, rec.Body.Len())

		attrs := capture.attrsFor("mcp_notification")
		require.NotNil(t, attrs)
		assert.Equal(t, "notifications/mystery", attrs["tool_name"])
		assert.Equal(t, "ignored", attrs["status"])
	})
}

func TestServeHTTP_Initialize(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)
	capture := &captureHandler{}

	req := newAuthedRequest(t, capture, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
	result := body["result"].(map[string]any)

	assert.Equal(t, protocolVersion, result["protocolVersion"])
	capabilities := result["capabilities"].(map[string]any)
	assert.Contains(t, capabilities, "tools")
	assert.Contains(t, capabilities, "resources")
	serverInfo := result["serverInfo"].(map[string]any)
	assert.Equal(t, "project-chat-mcp", serverInfo["name"])
	assert.Contains(t, result["instructions"], serverGuideURI)

	attrs := capture.attrsFor("mcp_request")
	require.NotNil(t, attrs)
	assert.Equal(t, "initialize", attrs["tool_name"])
	assert.Equal(t, "success", attrs["status"])
	assert.Contains(t, attrs, "latency_ms")
	assert.Contains(t, attrs, "key_id")
	assert.Contains(t, attrs, "owner_user_id")
}

func TestServeHTTP_ToolsListAndResourcesListDelegateToPrincipalScopes(t *testing.T) {
	key := &AuthenticatedAPIKey{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Scopes: []domain.MCPAPIScope{domain.MCPAPIScopeProjectsRead},
	}
	auth := &stubAuthService{result: key}
	handler := newTestHandler(auth)

	req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]any
	require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
	tools := body["result"].(map[string]any)["tools"].([]any)
	require.Len(t, tools, 1)
	assert.Equal(t, "list_projects", tools[0].(map[string]any)["name"])

	req = newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
	resources := body["result"].(map[string]any)["resources"].([]any)
	require.Len(t, resources, 2)
}

func TestServeHTTP_ResourcesRead(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)

	t.Run("missing uri is an invalid params RPC error", func(t *testing.T) {
		req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{}}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		rpcErr := body["error"].(map[string]any)
		assert.Equal(t, float64(-32602), rpcErr["code"])
	})

	t.Run("valid uri returns guide contents", func(t *testing.T) {
		req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"project-chat://server/guide"}}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		result := body["result"].(map[string]any)
		contents := result["contents"].([]any)
		require.Len(t, contents, 1)
	})

	t.Run("unknown uri surfaces readResource error as a tool error result", func(t *testing.T) {
		req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"project-chat://unknown"}}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		assert.Nil(t, body["error"])
		result := body["result"].(map[string]any)
		assert.Equal(t, true, result["isError"])
		structured := result["structuredContent"].(map[string]any)
		assert.Equal(t, "not_found", structured["error"].(map[string]any)["type"])
	})
}

func TestServeHTTP_PromptsList(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)

	req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"prompts/list"}`)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]any
	require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
	prompts := body["result"].(map[string]any)["prompts"].([]any)
	assert.Empty(t, prompts)
}

func TestServeHTTP_ToolsCall(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)

	t.Run("malformed params JSON is an invalid params RPC error", func(t *testing.T) {
		req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":"not-an-object"}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		rpcErr := body["error"].(map[string]any)
		assert.Equal(t, float64(-32602), rpcErr["code"])
	})

	t.Run("unknown tool name surfaces as a tool error result, not an RPC error", func(t *testing.T) {
		capture := &captureHandler{}
		req := newAuthedRequest(t, capture, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"does_not_exist","arguments":{}}}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		assert.Nil(t, body["error"])
		result := body["result"].(map[string]any)
		assert.Equal(t, true, result["isError"])
		structured := result["structuredContent"].(map[string]any)
		assert.Equal(t, "not_found", structured["error"].(map[string]any)["type"])

		attrs := capture.attrsFor("mcp_request")
		require.NotNil(t, attrs)
		assert.Equal(t, "does_not_exist", attrs["tool_name"])
		assert.Equal(t, "failure", attrs["status"])
	})

	t.Run("successful call returns a tool success result", func(t *testing.T) {
		req := newAuthedRequest(t, nil, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_projects","arguments":{}}}`)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		var body map[string]any
		require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
		result := body["result"].(map[string]any)
		assert.NotContains(t, result, "isError")
		assert.Contains(t, result, "structuredContent")
		assert.Contains(t, result, "content")
	})
}

func TestServeHTTP_UnknownMethodReturnsMethodNotFound(t *testing.T) {
	auth := &stubAuthService{result: fullScopePrincipalKey()}
	handler := newTestHandler(auth)
	capture := &captureHandler{}

	req := newAuthedRequest(t, capture, `{"jsonrpc":"2.0","id":1,"method":"totally/unknown"}`)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body map[string]any
	require.NoError(t, decodeJSON(rec.Body.Bytes(), &body))
	rpcErr := body["error"].(map[string]any)
	assert.Equal(t, float64(-32601), rpcErr["code"])
	assert.Equal(t, "method not found", rpcErr["message"])

	attrs := capture.attrsFor("mcp_request")
	require.NotNil(t, attrs)
	assert.Equal(t, "totally/unknown", attrs["tool_name"])
	assert.Equal(t, "failure", attrs["status"])
}
