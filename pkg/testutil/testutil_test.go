package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// errTest is a sentinel error for testing.
var errTest = errors.New("test error")

func TestWriteTestFile(t *testing.T) {
	t.Parallel()

	t.Run("creates file with content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filename := "test.txt"
		content := []byte("hello world")

		path := WriteTestFile(t, tmpDir, filename, content)

		// Verify path
		expectedPath := filepath.Join(tmpDir, filename)
		if path != expectedPath {
			t.Errorf("WriteTestFile() = %v, want %v", path, expectedPath)
		}

		// Verify content
		got, err := os.ReadFile(path) //nolint:gosec // Path is validated by WriteTestFile helper
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if string(got) != string(content) {
			t.Errorf("file content = %v, want %v", string(got), string(content))
		}
	})

	t.Run("handles subdirectory path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		subdir := filepath.Join(tmpDir, "sub")

		mkdirErr := os.MkdirAll(subdir, 0o750)
		if mkdirErr != nil {
			t.Fatalf("failed to create subdir: %v", mkdirErr)
		}

		filename := "sub/nested.txt"
		content := []byte("nested")

		path := WriteTestFile(t, tmpDir, filename, content)

		expectedPath := filepath.Join(tmpDir, filename)
		if path != expectedPath {
			t.Errorf("WriteTestFile() = %v, want %v", path, expectedPath)
		}
	})
}

func newTestResult() types.Result {
	return types.NewResultWithStatus(
		types.NewFileID("test.md"),
		types.NewLineNumber(1),
		types.NewBlockIndex(1),
		"code",
		types.StatusValid,
	)
}

func newTestCodeBlock() types.CodeBlock {
	return types.NewCodeBlock(
		types.NewLineNumber(1),
		languages.LangGo,
		"code",
	)
}

func TestAssertResultCount(t *testing.T) {
	t.Parallel()

	t.Run("pass when count matches", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult(), newTestResult()}
		AssertResultCount(t, results, 2)
	})

	t.Run("fail when count differs", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult()}
		expect := 3

		tt := &testing.T{}
		AssertResultCount(tt, results, expect)

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})
}

func TestAssertMinResults(t *testing.T) {
	t.Parallel()

	t.Run("pass when count is greater", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult(), newTestResult(), newTestResult()}
		AssertMinResults(t, results, 2)
	})

	t.Run("pass when count equals min", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult()}
		AssertMinResults(t, results, 1)
	})

	t.Run("fail when count is less", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult()}
		expect := 3

		tt := &testing.T{}
		AssertMinResults(tt, results, expect)

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})
}

func TestAssertMaxResults(t *testing.T) {
	t.Parallel()

	t.Run("pass when count is less", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult()}
		AssertMaxResults(t, results, 3)
	})

	t.Run("pass when count equals max", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult(), newTestResult(), newTestResult()}
		AssertMaxResults(t, results, 3)
	})

	t.Run("fail when count is greater", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{newTestResult(), newTestResult(), newTestResult()}
		maxVal := 2

		tt := &testing.T{}
		AssertMaxResults(tt, results, maxVal)

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})
}

func TestAssertBlockCount(t *testing.T) {
	t.Parallel()

	t.Run("pass when count matches", func(t *testing.T) {
		t.Parallel()

		blocks := []types.CodeBlock{newTestCodeBlock(), newTestCodeBlock()}
		AssertBlockCount(t, blocks, 2)
	})

	t.Run("fail when count differs", func(t *testing.T) {
		t.Parallel()

		blocks := []types.CodeBlock{newTestCodeBlock()}
		expect := 5

		tt := &testing.T{}
		AssertBlockCount(tt, blocks, expect)

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})
}

func TestAssertContextNotNil(t *testing.T) {
	t.Parallel()

	t.Run("pass with non-nil context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		AssertContextNotNil(ctx, t)
	})

	// Note: Testing nil context causes panic because AssertContextNotNil
	// calls t.Fatal on the passed-in *testing.T, which doesn't work with
	// a separate testing.T instance
}

func TestAssertContextCondition(t *testing.T) {
	t.Parallel()

	t.Run("pass when not done matches expectDone=false", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		tt := &testing.T{}
		AssertContextCondition(ctx, tt, false, "context should not be done")

		if tt.Failed() {
			t.Error("expected test to pass")
		}
	})
}

func TestAssertContextErr(t *testing.T) {
	t.Parallel()

	t.Run("pass when error matches", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		expected := context.Canceled

		tt := &testing.T{}
		AssertContextErr(ctx, tt, expected, "context should have Canceled error")

		if tt.Failed() {
			t.Error("expected test to pass")
		}
	})

	t.Run("pass when deadline exceeded", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 1)
		cancel()

		expected := context.DeadlineExceeded

		tt := &testing.T{}
		AssertContextErr(ctx, tt, expected, "context should have DeadlineExceeded error")

		if tt.Failed() {
			t.Error("expected test to pass")
		}
	})
}

func TestAssertZeroValue(t *testing.T) {
	t.Parallel()

	t.Run("pass when values match", func(t *testing.T) {
		t.Parallel()

		AssertZeroValue(t, "string", "test", "test")
		AssertZeroValue(t, "int", 42, 42)
		AssertZeroValue(t, "bool", true, true)
	})

	t.Run("fail when strings differ", func(t *testing.T) {
		t.Parallel()

		tt := &testing.T{}
		AssertZeroValue(tt, "string", "a", "b")

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})

	t.Run("fail when ints differ", func(t *testing.T) {
		t.Parallel()

		tt := &testing.T{}
		AssertZeroValue(tt, "int", 1, 2)

		if !tt.Failed() {
			t.Error("expected test to fail")
		}
	})
}

func TestIsContextDone(t *testing.T) {
	t.Parallel()

	t.Run("returns true when done", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if !isContextDone(ctx) {
			t.Error("expected true for done context")
		}
	})

	t.Run("returns false when not done", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		if isContextDone(ctx) {
			t.Error("expected false for not-done context")
		}
	})

	t.Run("returns true when cancelled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if !isContextDone(ctx) {
			t.Error("expected true for cancelled context")
		}
	})
}

func TestAssertZeroValueGenerics(t *testing.T) {
	t.Parallel()

	t.Run("with error type", func(t *testing.T) {
		t.Parallel()

		AssertZeroValue(t, "error", errTest, errTest)
	})

	t.Run("with struct type", func(t *testing.T) {
		t.Parallel()

		type MyStruct struct {
			Name string
			Age  int
		}

		s1 := MyStruct{Name: "Alice", Age: 30}
		s2 := MyStruct{Name: "Alice", Age: 30}

		AssertZeroValue(t, "struct", s1, s2)
	})
}
