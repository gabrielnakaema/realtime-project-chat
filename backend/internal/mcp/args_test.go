package mcp

import (
	"testing"
	"time"

	"github.com/gabrielnakaema/project-chat/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequiredUUIDArg(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name    string
		args    map[string]any
		wantErr bool
		want    uuid.UUID
	}{
		{name: "key missing", args: map[string]any{}, wantErr: true},
		{name: "non-string value", args: map[string]any{"id": 123}, wantErr: true},
		{name: "malformed uuid string", args: map[string]any{"id": "not-a-uuid"}, wantErr: true},
		{name: "valid uuid string", args: map[string]any{"id": validID.String()}, want: validID},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := requiredUUIDArg(tc.args, "id")
			if tc.wantErr {
				require.Error(t, err)
				assertBusinessValidationError(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOptionalUUIDArg(t *testing.T) {
	validID := uuid.New()

	t.Run("absent returns nil, nil", func(t *testing.T) {
		got, err := optionalUUIDArg(map[string]any{}, "id")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("explicit nil returns nil, nil", func(t *testing.T) {
		got, err := optionalUUIDArg(map[string]any{"id": nil}, "id")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("non-string errors", func(t *testing.T) {
		_, err := optionalUUIDArg(map[string]any{"id": 42}, "id")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("malformed uuid errors", func(t *testing.T) {
		_, err := optionalUUIDArg(map[string]any{"id": "nope"}, "id")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid uuid returns pointer to parsed value", func(t *testing.T) {
		got, err := optionalUUIDArg(map[string]any{"id": validID.String()}, "id")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, validID, *got)
	})
}

func TestOptionalUUIDSliceArg(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	t.Run("absent returns nil", func(t *testing.T) {
		got, err := optionalUUIDSliceArg(map[string]any{}, "ids")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("non-array value errors", func(t *testing.T) {
		_, err := optionalUUIDSliceArg(map[string]any{"ids": "not-an-array"}, "ids")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("array with non-string item errors", func(t *testing.T) {
		_, err := optionalUUIDSliceArg(map[string]any{"ids": []any{id1.String(), 5}}, "ids")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("array with malformed uuid errors", func(t *testing.T) {
		_, err := optionalUUIDSliceArg(map[string]any{"ids": []any{id1.String(), "bad"}}, "ids")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid array is parsed", func(t *testing.T) {
		got, err := optionalUUIDSliceArg(map[string]any{"ids": []any{id1.String(), id2.String()}}, "ids")
		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{id1, id2}, got)
	})
}

func TestRequiredStringArg(t *testing.T) {
	t.Run("missing key errors", func(t *testing.T) {
		_, err := requiredStringArg(map[string]any{}, "title")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("empty string errors", func(t *testing.T) {
		_, err := requiredStringArg(map[string]any{"title": ""}, "title")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("whitespace-only string errors", func(t *testing.T) {
		_, err := requiredStringArg(map[string]any{"title": "   "}, "title")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("value is trimmed on success", func(t *testing.T) {
		got, err := requiredStringArg(map[string]any{"title": "  hello  "}, "title")
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})
}

func TestOptionalLimitArg(t *testing.T) {
	t.Run("absent returns default", func(t *testing.T) {
		got, err := optionalLimitArg(map[string]any{}, "limit", 15)
		require.NoError(t, err)
		assert.Equal(t, 15, got)
	})

	t.Run("non-numeric type errors", func(t *testing.T) {
		_, err := optionalLimitArg(map[string]any{"limit": "10"}, "limit", 15)
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("out-of-range value errors", func(t *testing.T) {
		for _, value := range []float64{0, -3, maxToolResultLimit + 1} {
			_, err := optionalLimitArg(map[string]any{"limit": value}, "limit", 15)
			require.Error(t, err)
			assertBusinessValidationError(t, err)
		}
	})

	t.Run("fractional value errors", func(t *testing.T) {
		_, err := optionalLimitArg(map[string]any{"limit": float64(7.9)}, "limit", 15)
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid integer is accepted", func(t *testing.T) {
		got, err := optionalLimitArg(map[string]any{"limit": float64(7)}, "limit", 15)
		require.NoError(t, err)
		assert.Equal(t, 7, got)
	})
}

func TestOptionalBoolArg(t *testing.T) {
	t.Run("absent returns default", func(t *testing.T) {
		got, err := optionalBoolArg(map[string]any{}, "flag", true)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("wrong type errors", func(t *testing.T) {
		_, err := optionalBoolArg(map[string]any{"flag": "true"}, "flag", false)
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid bool passes through", func(t *testing.T) {
		got, err := optionalBoolArg(map[string]any{"flag": false}, "flag", true)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestRequiredTaskPriorityArg(t *testing.T) {
	t.Run("invalid value errors naming allowed values", func(t *testing.T) {
		_, err := requiredTaskPriorityArg(map[string]any{"priority": "urgent"}, "priority")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
		assert.Contains(t, err.Error(), "low")
		assert.Contains(t, err.Error(), "medium")
		assert.Contains(t, err.Error(), "high")
	})

	t.Run("mixed-case input is accepted and lowercased", func(t *testing.T) {
		got, err := requiredTaskPriorityArg(map[string]any{"priority": "HIGH"}, "priority")
		require.NoError(t, err)
		assert.Equal(t, domain.TaskPriorityHigh, got)
	})
}

func TestOptionalTimeArg(t *testing.T) {
	t.Run("absent returns nil", func(t *testing.T) {
		got, err := optionalTimeArg(map[string]any{}, "due_date")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("non-RFC3339 string errors", func(t *testing.T) {
		_, err := optionalTimeArg(map[string]any{"due_date": "2026-07-22"}, "due_date")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid RFC3339 is parsed", func(t *testing.T) {
		got, err := optionalTimeArg(map[string]any{"due_date": "2026-07-22T10:00:00Z"}, "due_date")
		require.NoError(t, err)
		require.NotNil(t, got)
		want, _ := time.Parse(time.RFC3339, "2026-07-22T10:00:00Z")
		assert.True(t, want.Equal(*got))
	})
}

func TestOptionalStringSliceArg(t *testing.T) {
	t.Run("non-array errors", func(t *testing.T) {
		_, err := optionalStringSliceArg(map[string]any{"tags": "not-array"}, "tags")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("array with non-string item errors", func(t *testing.T) {
		_, err := optionalStringSliceArg(map[string]any{"tags": []any{"ok", 5}}, "tags")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("valid array is trimmed", func(t *testing.T) {
		got, err := optionalStringSliceArg(map[string]any{"tags": []any{"  a  ", "b"}}, "tags")
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, got)
	})
}

func TestOptionalTrimmedStringArg(t *testing.T) {
	t.Run("absent returns zero value", func(t *testing.T) {
		got, err := optionalTrimmedStringArg(map[string]any{}, "code")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("non-string errors", func(t *testing.T) {
		_, err := optionalTrimmedStringArg(map[string]any{"code": 5}, "code")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("value is trimmed on success", func(t *testing.T) {
		got, err := optionalTrimmedStringArg(map[string]any{"code": "  TASK-1  "}, "code")
		require.NoError(t, err)
		assert.Equal(t, "TASK-1", got)
	})
}

func TestOptionalTrimmedStringArgPointer(t *testing.T) {
	t.Run("absent returns nil", func(t *testing.T) {
		got, err := optionalTrimmedStringArgPointer(map[string]any{}, "code")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("non-string errors", func(t *testing.T) {
		_, err := optionalTrimmedStringArgPointer(map[string]any{"code": 5}, "code")
		require.Error(t, err)
		assertBusinessValidationError(t, err)
	})

	t.Run("whitespace is trimmed on success", func(t *testing.T) {
		got, err := optionalTrimmedStringArgPointer(map[string]any{"code": "  TASK-1  "}, "code")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "TASK-1", *got)
	})
}

func TestParamsFromRaw(t *testing.T) {
	t.Run("empty raw message returns empty map", func(t *testing.T) {
		got := paramsFromRaw(nil)
		assert.Equal(t, map[string]any{}, got)
	})

	t.Run("malformed JSON returns empty map", func(t *testing.T) {
		got := paramsFromRaw([]byte(`{not-json`))
		assert.Equal(t, map[string]any{}, got)
	})

	t.Run("valid JSON object is decoded", func(t *testing.T) {
		got := paramsFromRaw([]byte(`{"uri":"project-chat://server/guide"}`))
		assert.Equal(t, map[string]any{"uri": "project-chat://server/guide"}, got)
	})
}

func assertBusinessValidationError(t *testing.T, err error) {
	t.Helper()
	result := toolErrorResult(err)
	structured := result["structuredContent"].(map[string]any)
	assert.Equal(t, "business_validation", structured["error"].(map[string]any)["type"])
}
