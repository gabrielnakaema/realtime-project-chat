package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalTimeStringPreservesCursorPrecision(t *testing.T) {
	want := time.Date(2026, time.July, 23, 10, 40, 57, 825935000, time.FixedZone("test", -3*60*60))

	encoded := optionalTimeString(&want)
	require.NotNil(t, encoded)
	decoded, err := time.Parse(time.RFC3339Nano, *encoded)
	require.NoError(t, err)
	assert.True(t, want.Equal(decoded))
}
