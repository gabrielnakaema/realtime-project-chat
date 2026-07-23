package mcp

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchTasksCursorRoundTrip(t *testing.T) {
	dueDate := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	want := searchTasksCursor{
		DueDate:   &dueDate,
		UpdatedAt: dueDate.Add(-time.Hour),
		TaskID:    uuid.New(),
	}

	encoded, err := encodeSearchTasksCursor(want)
	require.NoError(t, err)
	decoded, err := decodeSearchTasksCursor(encoded)
	require.NoError(t, err)
	assert.Equal(t, want, *decoded)
}

func TestSearchTasksCursorAllowsNullDueDate(t *testing.T) {
	want := searchTasksCursor{UpdatedAt: time.Now().UTC(), TaskID: uuid.New()}

	encoded, err := encodeSearchTasksCursor(want)
	require.NoError(t, err)
	decoded, err := decodeSearchTasksCursor(encoded)
	require.NoError(t, err)
	assert.Nil(t, decoded.DueDate)
	assert.Equal(t, want.TaskID, decoded.TaskID)
}

func TestSearchTasksCursorRejectsInvalidValue(t *testing.T) {
	_, err := decodeSearchTasksCursor("not-a-cursor")
	require.Error(t, err)
}
