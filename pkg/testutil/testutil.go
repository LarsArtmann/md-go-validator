// Package testutil provides test utility functions for the md-go-validator package.
package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// Magic number constants.
const defaultFilePermissions = 0o600

// WriteTestFile writes test content to a file and returns the path.
func WriteTestFile(t *testing.T, tmpDir, filename string, content []byte) string {
	t.Helper()

	path := filepath.Join(tmpDir, filename)

	err := os.WriteFile(path, content, defaultFilePermissions)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

// AssertResultCount fails if the number of results doesn't match the expected count.
func AssertResultCount(t *testing.T, results []types.Result, expected int) {
	t.Helper()

	if len(results) != expected {
		t.Errorf("expected %d results, got %d", expected, len(results))
	}
}

// AssertMinResults fails if the number of results is less than minVal.
func AssertMinResults(t *testing.T, results []types.Result, minVal int) {
	t.Helper()

	if len(results) < minVal {
		t.Errorf("expected at least %d results, got %d", minVal, len(results))
	}
}

// AssertMaxResults fails if the number of results is greater than maxVal.
func AssertMaxResults(t *testing.T, results []types.Result, maxVal int) {
	t.Helper()

	if len(results) > maxVal {
		t.Errorf("expected at most %d results, got %d", maxVal, len(results))
	}
}

// AssertBlockCount fails if the number of code blocks doesn't match the expected count.
func AssertBlockCount(t *testing.T, blocks []types.CodeBlock, expected int) {
	t.Helper()

	if len(blocks) != expected {
		t.Errorf("expected %d blocks, got %d", expected, len(blocks))
	}
}

// AssertSingleBlock fails if the slice is not exactly one block. Returns the
// single block so callers can chain further assertions.
func AssertSingleBlock(t *testing.T, blocks []types.CodeBlock) types.CodeBlock {
	t.Helper()

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	return blocks[0]
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// AssertContextNotNil fails if ctx is nil.
func AssertContextNotNil(ctx context.Context, t *testing.T) {
	t.Helper()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}

// AssertContextCondition fails if ctx done state doesn't match expectDone.
func AssertContextCondition(ctx context.Context, t *testing.T, expectDone bool, msg string) {
	t.Helper()

	done := isContextDone(ctx)
	if done == expectDone {
		return
	}

	t.Fatal(msg)
}

// AssertContextErr fails if ctx.Err() != expected.
func AssertContextErr(ctx context.Context, t *testing.T, expected error, msg string) {
	t.Helper()

	done := isContextDone(ctx)
	if !done {
		t.Fatal(msg)
	}

	if !errors.Is(ctx.Err(), expected) { //nolint:legacyerrors // context error sentinel
		t.Errorf("expected %v, got %v", expected, ctx.Err())
	}
}

// AssertZeroValue fails if got != expected.
func AssertZeroValue[T comparable](t *testing.T, name string, got, expected T) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %s %v, got %v", name, expected, got)
	}
}

// NewTestErrorResult builds a types.Result with the given fields and the
// supplied message wrapped in a *types.TestError. Mirrors types.NewErrorResult
// for tests that need the test error type and live in packages that shouldn't
// import each other.
func NewTestErrorResult(fileID string, line, block int, code, errMsg string) types.Result {
	return NewTestErrorResultWith(fileID, line, block, code, types.NewTestError(errMsg))
}

// NewTestErrorResultWith builds a types.Result with the given fields and the
// supplied error.
func NewTestErrorResultWith(fileID string, line, block int, code string, err error) types.Result {
	return types.NewErrorResult(
		types.NewFileID(fileID),
		types.NewLineNumber(line),
		types.NewBlockIndex(block),
		code,
		err,
	)
}
