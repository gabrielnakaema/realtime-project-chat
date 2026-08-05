package tasks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchRequest(t *testing.T, query string) *http.Request {
	t.Helper()
	return httptest.NewRequest(http.MethodGet, "/tasks/search?"+query, nil)
}

func TestParseSearchTasksRequest(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	pendingColumnID := uuid.New()
	doneColumnID := uuid.New()
	cursorTaskID := uuid.New()

	t.Run("carries every filter the search service understands", func(t *testing.T) {
		query := "query=needle" +
			"&project_id=" + projectID.String() +
			"&project_column_ids=" + pendingColumnID.String() + "," + doneColumnID.String() +
			"&include_archived=true" +
			"&include_done=true" +
			"&limit=42" +
			"&task_id=" + cursorTaskID.String() +
			"&updated_at=2026-08-05T10:00:00Z" +
			"&due_date=2026-08-06T10:00:00Z"

		request, err := parseSearchTasksRequest(searchRequest(t, query), userID)

		require.NoError(t, err)
		assert.Equal(t, userID, request.UserId)
		assert.Equal(t, "needle", request.SearchQuery)
		assert.Equal(t, 42, request.Limit)

		require.NotNil(t, request.ProjectId)
		assert.Equal(t, projectID, *request.ProjectId)
		assert.Equal(t, []uuid.UUID{pendingColumnID, doneColumnID}, request.ProjectColumnIDs)

		assert.True(t, request.IncludeArchived)
		assert.True(t, request.IncludeDone, "a search must be able to reach completed tasks")

		require.NotNil(t, request.CursorTaskId)
		assert.Equal(t, cursorTaskID, *request.CursorTaskId)
		require.NotNil(t, request.CursorUpdatedAt)
		assert.Equal(t, time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC), request.CursorUpdatedAt.UTC())
		require.NotNil(t, request.CursorDueDate)
		assert.Equal(t, time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC), request.CursorDueDate.UTC())
	})

	t.Run("stays an unscoped search over active tasks by default", func(t *testing.T) {
		request, err := parseSearchTasksRequest(searchRequest(t, "query=needle"), userID)

		require.NoError(t, err)
		assert.Nil(t, request.ProjectId)
		assert.Empty(t, request.ProjectColumnIDs)
		assert.False(t, request.IncludeArchived)
		assert.False(t, request.IncludeDone)
		assert.Equal(t, 15, request.Limit)
	})

	t.Run("only true opts into the flags", func(t *testing.T) {
		request, err := parseSearchTasksRequest(searchRequest(t, "query=needle&include_done=1&include_archived=yes"), userID)

		require.NoError(t, err)
		assert.False(t, request.IncludeDone)
		assert.False(t, request.IncludeArchived)
	})

	t.Run("rejects a missing query", func(t *testing.T) {
		_, err := parseSearchTasksRequest(searchRequest(t, "project_id="+projectID.String()), userID)

		require.Error(t, err)
		assert.Equal(t, "query is required", err.Error())
	})

	t.Run("rejects filters it cannot read", func(t *testing.T) {
		_, err := parseSearchTasksRequest(searchRequest(t, "query=needle&project_id=not-a-uuid"), userID)
		require.Error(t, err)

		_, err = parseSearchTasksRequest(searchRequest(t, "query=needle&project_column_ids=not-a-uuid"), userID)
		require.Error(t, err)
	})
}
