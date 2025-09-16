package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		data           interface{}
		headers        http.Header
		expectedStatus int
		expectedBody   string
		expectedHeader string
		shouldErr      bool
	}{
		{
			name:           "successful write with struct",
			status:         http.StatusOK,
			data:           testStruct{Name: "John", Age: 30, Email: "john@example.com"},
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"John","age":30,"email":"john@example.com"}`,
			shouldErr:      false,
		},
		{
			name:           "successful write with map",
			status:         http.StatusCreated,
			data:           map[string]string{"message": "success"},
			headers:        nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"success"}`,
			shouldErr:      false,
		},
		{
			name:           "write with custom headers",
			status:         http.StatusOK,
			data:           map[string]string{"test": "data"},
			headers:        http.Header{"X-Custom-Header": []string{"custom-value"}},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"test":"data"}`,
			expectedHeader: "custom-value",
			shouldErr:      false,
		},
		{
			name:           "write with nil data",
			status:         http.StatusOK,
			data:           nil,
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "null",
			shouldErr:      false,
		},
		{
			name:           "write with array",
			status:         http.StatusOK,
			data:           []string{"item1", "item2", "item3"},
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `["item1","item2","item3"]`,
			shouldErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := WriteJSON(w, tt.status, tt.data, tt.headers)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				body := strings.TrimSpace(w.Body.String())
				assert.Equal(t, tt.expectedBody, body)

				if tt.expectedHeader != "" {
					assert.Equal(t, tt.expectedHeader, w.Header().Get("X-Custom-Header"))
				}
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		dest        interface{}
		expectedErr string
		setup       func() (*http.Request, *httptest.ResponseRecorder)
	}{
		{
			name: "successful read",
			body: `{"name":"John","age":30,"email":"john@example.com"}`,
			dest: &testStruct{},
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":30,"email":"john@example.com"}`)), httptest.NewRecorder()
			},
		},
		{
			name:        "empty body",
			body:        "",
			dest:        &testStruct{},
			expectedErr: "body must not be empty",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader("")), httptest.NewRecorder()
			},
		},
		{
			name:        "malformed JSON",
			body:        `{"name":"John","age":}`,
			dest:        &testStruct{},
			expectedErr: "body contains badly-formed JSON",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":}`)), httptest.NewRecorder()
			},
		},
		{
			name:        "unknown field",
			body:        `{"name":"John","age":30,"unknown_field":"value"}`,
			dest:        &testStruct{},
			expectedErr: "body contains unknown key",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":30,"unknown_field":"value"}`)), httptest.NewRecorder()
			},
		},
		{
			name:        "wrong type",
			body:        `{"name":"John","age":"thirty","email":"john@example.com"}`,
			dest:        &testStruct{},
			expectedErr: "body contains incorrect JSON type for field",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":"thirty","email":"john@example.com"}`)), httptest.NewRecorder()
			},
		},
		{
			name:        "multiple JSON objects",
			body:        `{"name":"John"}{"name":"Jane"}`,
			dest:        &testStruct{},
			expectedErr: "body must contain a single JSON",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				return httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John"}{"name":"Jane"}`)), httptest.NewRecorder()
			},
		},
		{
			name:        "body too large",
			body:        `{"data":"` + strings.Repeat("a", 2_000_000) + `"}`, // 2MB
			dest:        &testStruct{},
			expectedErr: "body must not be larger than",
			setup: func() (*http.Request, *httptest.ResponseRecorder) {
				largeJSON := `{"data":"` + strings.Repeat("a", 2_000_000) + `"}`
				return httptest.NewRequest("POST", "/test", strings.NewReader(largeJSON)), httptest.NewRecorder()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w := tt.setup()

			err := ReadJSON(w, r, tt.dest)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)

				if testStruct, ok := tt.dest.(*testStruct); ok {
					assert.Equal(t, "John", testStruct.Name)
					assert.Equal(t, 30, testStruct.Age)
					assert.Equal(t, "john@example.com", testStruct.Email)
				}
			}
		})
	}
}

// Tests panic to avoid developer error when passing a non-pointer to ReadJSON
func TestReadJSON_InvalidUnmarshalError(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", strings.NewReader(`{"test":"value"}`))
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r.(error).Error(), "json: Unmarshal")
		}
	}()

	_ = ReadJSON(w, r, testStruct{})
}

func TestReadJSON_MaxBytesReader(t *testing.T) {
	largeValue := strings.Repeat("a", 1_500_000) // 1.5MB
	largeJSON := fmt.Sprintf(`{"large_field":"%s"}`, largeValue)

	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeJSON))
	w := httptest.NewRecorder()

	var dest map[string]string
	err := ReadJSON(w, r, &dest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body must not be larger than")
}

func TestWriteJSON_UnmarshalableData(t *testing.T) {
	w := httptest.NewRecorder()

	err := WriteJSON(w, http.StatusOK, make(chan int), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type")
}

func TestReadJSON_SyntaxErrorOffset(t *testing.T) {
	r := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John",}`))
	w := httptest.NewRecorder()

	var dest testStruct
	err := ReadJSON(w, r, &dest)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body contains badly-formed JSON (at character")
}

func TestJSON_RoundTrip(t *testing.T) {
	original := testStruct{Name: "Test", Age: 25, Email: "test@example.com"}

	jsonData, err := json.Marshal(original)
	assert.NoError(t, err)

	r := httptest.NewRequest("POST", "/test", bytes.NewReader(jsonData))
	w := httptest.NewRecorder()

	var dest testStruct
	err = ReadJSON(w, r, &dest)
	assert.NoError(t, err)
	assert.Equal(t, original, dest)
}
