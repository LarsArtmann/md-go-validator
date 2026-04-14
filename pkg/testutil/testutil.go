package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

func WriteTestFile(t *testing.T, tmpDir, filename string, content []byte) string {
	t.Helper()
	path := filepath.Join(tmpDir, filename)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func AssertResultCount(t *testing.T, results []types.Result, expected int) {
	t.Helper()
	if len(results) != expected {
		t.Errorf("expected %d results, got %d", expected, len(results))
	}
}

func AssertBlockCount(t *testing.T, blocks []types.CodeBlock, expected int) {
	t.Helper()
	if len(blocks) != expected {
		t.Errorf("expected %d blocks, got %d", expected, len(blocks))
	}
}
