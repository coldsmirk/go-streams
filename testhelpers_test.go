package streams

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// createTempFile creates a temp file with optional initial content and returns its path.
// The file is registered for removal with t.Cleanup.
func createTempFile(t *testing.T, pattern string, content string) string {
	t.Helper()
	dir := t.TempDir() // automatically cleaned up
	path := filepath.Join(dir, pattern)
	f, err := os.Create(path)
	require.NoError(t, err, "create temp file should succeed")
	if content != "" {
		_, err := f.WriteString(content)
		require.NoError(t, err, "write temp file should succeed")
	}
	err = f.Close()
	require.NoError(t, err, "close temp file should succeed")
	return path
}
