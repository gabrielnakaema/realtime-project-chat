package platformhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestParseURLUUID(t *testing.T) {
	want := uuid.New()
	r := withURLParam(httptest.NewRequest(http.MethodGet, "/", nil), "id", want.String())

	got, err := ParseURLUUID(r, "id")

	require.NoError(t, err)
	require.Equal(t, want, got)

	_, err = ParseURLUUID(httptest.NewRequest(http.MethodGet, "/", nil), "id")
	require.EqualError(t, err, "id is required")
}

func TestParseOptionalQueryUUID(t *testing.T) {
	want := uuid.New()

	tests := []struct {
		name      string
		query     string
		want      *uuid.UUID
		wantError bool
	}{
		{name: "missing", want: nil},
		{name: "valid", query: "?id=" + want.String(), want: &want},
		{name: "invalid", query: "?id=not-a-uuid", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			got, err := ParseOptionalQueryUUID(r, "id")

			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tt.want, *got)
		})
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		want      int32
		wantError bool
	}{
		{name: "default", want: 10},
		{name: "valid", query: "?limit=25", want: 25},
		{name: "zero", query: "?limit=0", wantError: true},
		{name: "above maximum", query: "?limit=51", wantError: true},
		{name: "not an integer", query: "?limit=invalid", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			got, err := ParseLimit(r, 10, 50)

			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseRFC3339Cursor(t *testing.T) {
	want := time.Date(2026, time.January, 1, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		query     string
		want      *time.Time
		wantError bool
	}{
		{name: "missing", want: nil},
		{name: "valid", query: "?before=" + want.Format(time.RFC3339), want: &want},
		{name: "invalid", query: "?before=not-a-date", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			got, err := ParseRFC3339Cursor(r, "before")

			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tt.want, *got)
		})
	}
}
