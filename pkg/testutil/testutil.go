package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

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

func AssertMinResults(t *testing.T, results []types.Result, min int) {
	t.Helper()

	if len(results) < min {
		t.Errorf("expected at least %d results, got %d", min, len(results))
	}
}

func AssertMaxResults(t *testing.T, results []types.Result, max int) {
	t.Helper()

	if len(results) > max {
		t.Errorf("expected at most %d results, got %d", max, len(results))
	}
}

func AssertBlockCount(t *testing.T, blocks []types.CodeBlock, expected int) {
	t.Helper()

	if len(blocks) != expected {
		t.Errorf("expected %d blocks, got %d", expected, len(blocks))
	}
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func AssertContextNotNil(t *testing.T, ctx context.Context) {
	t.Helper()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}

func AssertContextCondition(t *testing.T, ctx context.Context, expectDone bool, msg string) {
	t.Helper()

	done := isContextDone(ctx)
	if done == expectDone {
		return
	}

	t.Fatal(msg)
}

func AssertContextErr(t *testing.T, ctx context.Context, expected error, msg string) {
	t.Helper()

	done := isContextDone(ctx)
	if !done {
		t.Fatal(msg)
	}

	if !errors.Is(ctx.Err(), expected) {
		t.Errorf("expected %v, got %v", expected, ctx.Err())
	}
}

func AssertZeroValue[T comparable](t *testing.T, name string, got, expected T) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %s %v, got %v", name, expected, got)
	}
}
