package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func WriteTestFile(t *testing.T, tmpDir, filename string, content []byte) string {
	t.Helper()
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
