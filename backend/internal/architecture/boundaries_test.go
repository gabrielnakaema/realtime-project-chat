package architecture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckImportBoundaries(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, root, "internal/platform/config/config.go", `package config
import _ "github.com/gabrielnakaema/project-chat/internal/tasks"
`)
	writeGoFile(t, root, "internal/chat/service.go", `package chat
import _ "github.com/gabrielnakaema/project-chat/internal/project"
`)
	writeGoFile(t, root, "internal/realtime/app/app.go", `package app
import (
  _ "github.com/gabrielnakaema/project-chat/internal/chat"
  _ "github.com/gabrielnakaema/project-chat/internal/project"
)
`)

	violations, err := CheckImportBoundaries(root)
	require.NoError(t, err)
	require.Len(t, violations, 2)
	require.Contains(t, violations[0].Error()+violations[1].Error(), "platform must not import capability")
	require.Contains(t, violations[0].Error()+violations[1].Error(), `capability "chat" must not import capability "project"`)
}

func writeGoFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}
