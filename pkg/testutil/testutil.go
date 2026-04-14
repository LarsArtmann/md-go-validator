package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// TestError is a simple error implementation for testing.
type TestError struct {
	msg string
}

// NewTestError creates a new TestError with the given message.
func NewTestError(msg string) *TestError {
	return &TestError{msg: msg}
}

// Error implements the error interface.
func (e *TestError) Error() string {
	return e.msg
}

// ValidResultSpec defines a result for testing.
type ValidResultSpec struct {
	FileID string
	Line   int
	Block  int
	Code   string
}

// NewValidResults creates a slice of valid results from specs.
func NewValidResults(specs ...ValidResultSpec) []types.Result {
	results := make([]types.Result, len(specs))
	for i, s := range specs {
		results[i] = types.NewValidResult(
			types.NewFileID(s.FileID),
			types.NewLineNumber(s.Line),
			types.NewBlockIndex(s.Block),
			s.Code,
		)
	}

	return results
}

// validResultSpec is an internal version with unexported fields.
type validResultSpec struct {
	fileID string
	line   int
	block  int
	code   string
}

// NewValidResultsFromSpecs creates results from unexported specs.
func NewValidResultsFromSpecs(specs []validResultSpec) []types.Result {
	results := make([]types.Result, len(specs))
	for i, s := range specs {
		results[i] = types.NewValidResult(
			types.NewFileID(s.fileID),
			types.NewLineNumber(s.line),
			types.NewBlockIndex(s.block),
			s.code,
		)
	}

	return results
}

func WriteTestFile(t *testing.T, tmpDir, filename string, content []byte) string {
	t.Helper()

	path := filepath.Join(tmpDir, filename)
	err := os.WriteFile(path, content, 0o600)
	if err != nil {
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
