package mdgovalidator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

func TestExtractGoCodeBlocks(t *testing.T) {
	t.Parallel()

	t.Run("no code blocks", func(t *testing.T) {
		t.Parallel()
		blocks := ExtractGoCodeBlocks("Just text\nNo code here")
		if len(blocks) != 0 {
			t.Errorf("expected 0 blocks, got %d", len(blocks))
		}
	})

	t.Run("single go block", func(t *testing.T) {
		t.Parallel()
		content := "Some text\n```go\nfmt.Println(\"hello\")\n```\nMore text"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if blocks[0].LineNumber != types.NewLineNumber(2) {
			t.Errorf("expected line 2, got %d", blocks[0].LineNumber)
		}
	})

	t.Run("skip other languages", func(t *testing.T) {
		t.Parallel()
		content := "```python\nprint('hello')\n```\n```go\nfmt.Println(\"hello\")\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if blocks[0].LineNumber != types.NewLineNumber(4) {
			t.Errorf("expected line 4, got %d", blocks[0].LineNumber)
		}
	})

	t.Run("skip directive before block", func(t *testing.T) {
		t.Parallel()
		content := "<!-- skip-validate -->\n```go\npartial code\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if !blocks[0].IsSkipped() {
			t.Error("expected block to be skipped")
		}
	})

	t.Run("golang tag", func(t *testing.T) {
		t.Parallel()
		content := "```golang\nfmt.Println(\"hello\")\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
	})
}

func TestExtractGoCodeBlocks_SkipDirective(t *testing.T) {
	t.Parallel()
	content := `<!-- skip-validate -->
` + "```go" + `
type Partial struct {
    Name string
}
` + "```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !blocks[0].IsSkipped() {
		t.Error("expected block to be skipped")
	}
}

func TestExtractGoCodeBlocks_SkipInCode(t *testing.T) {
	t.Parallel()
	content := "```go\n//nolint\ntype Partial struct{}\n```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !blocks[0].IsSkipped() {
		t.Error("expected block to be skipped due to //nolint in code")
	}
}

func TestExtractGoCodeBlocks_EmptyBlock(t *testing.T) {
	t.Parallel()
	content := "```go\n\n```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty code, got %d", len(blocks))
	}
}

func TestValidateGoCode(t *testing.T) {
	t.Parallel()

	validCases := []struct {
		name string
		code string
	}{
		{"complete file", "package main\n\nfunc main() {}\n"},
		{"type declaration", "type User struct {\n\tName string\n}"},
		{"function signature", "func DoSomething() error"},
		{"import statement", "import \"fmt\""},
		{"variable declaration", "var x = 42"},
		{"expression", "x + y"},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateGoCode(tc.code); err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}

	invalidCases := []struct {
		name string
		code string
	}{
		{"invalid go.mod syntax", "require (\n\tgithub.com/pkg v1.0.0\n)"},
		{"invalid syntax", "func broken {"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateGoCode(tc.code); err == nil {
				t.Error("expected error for invalid syntax")
			}
		})
	}
}

func TestIndentCode(t *testing.T) {
	t.Parallel()
	input := "line1\nline2\n\nline4"
	expected := "\tline1\n\tline2\n\n\tline4\n"

	result := indentCode(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValidator_ValidateFile(t *testing.T) {
	t.Parallel()
	content := []byte(`# Test

` + "```go" + `
package main

func main() {
    fmt.Println("hello")
}
` + "```" + `
`)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	ctx := context.Background()
	results, err := v.ValidateFile(ctx, tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestValidator_ValidateFile_NonExistent(t *testing.T) {
	t.Parallel()
	v := New(false)
	ctx := context.Background()
	_, err := v.ValidateFile(ctx, "/nonexistent/path/file.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidator_ValidateDirectory(t *testing.T) {
	t.Parallel()
	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	ctx := context.Background()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result (only .md files), got %d", len(results))
	}
}

func TestHasErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()
		if HasErrors(nil) {
			t.Error("expected false for nil results")
		}
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(
				types.NewFileID("test.md"),
				types.NewLineNumber(1),
				types.NewBlockIndex(1),
				"package main",
			),
			types.NewValidResult(
				types.NewFileID("test.md"),
				types.NewLineNumber(5),
				types.NewBlockIndex(2),
				"package main",
			),
		}
		if HasErrors(results) {
			t.Error("expected false for valid results")
		}
	})

	t.Run("skipped doesn't count", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewSkippedResult(
				types.NewFileID("test.md"),
				types.NewLineNumber(1),
				types.NewBlockIndex(1),
				"skipped",
			),
		}
		if HasErrors(results) {
			t.Error("expected false for skipped error")
		}
	})

	t.Run("has error", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(
				types.NewFileID("test.md"),
				types.NewLineNumber(1),
				types.NewBlockIndex(1),
				"invalid",
				&testError{},
			),
		}
		if !HasErrors(results) {
			t.Error("expected true for error result")
		}
	})
}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestValidator_ValidateDirectory_Cancellation(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately before processing

	_, err := v.ValidateDirectory(ctx, tmpDir)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Verify it's a context cancellation error
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("expected context cancellation error, got: %v", err)
	}
}

func TestValidator_ValidateDirectory_CancellationDuringProcessing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTestMarkdownFiles(t, tmpDir, "test%d.md", 5)

	v := New(false).WithConcurrency(1) // Single worker for predictable cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Context will cancel during processing
	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil && !strings.Contains(err.Error(), "cancelled") {
		t.Fatalf("expected cancellation error, got: %v", err)
	}

	// Some results may have been collected before cancellation
	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
}

func TestValidator_ValidateFile_Empty(t *testing.T) {
	t.Parallel()
	content := []byte("No code blocks here")

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	ctx := context.Background()
	results, err := v.ValidateFile(ctx, tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results for empty file, got %d", len(results))
	}
}

func TestValidator_ValidateDirectory_SkipDirs(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()

	// Create files in various directories
	subdirs := []string{".hidden", "vendor", "node_modules", "build", "dist", "normal"}
	for _, dir := range subdirs {
		dirPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dirPath, "test.md"), content, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	v := New(false)
	ctx := context.Background()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	// Only normal directory should be processed
	if len(results) != 1 {
		t.Errorf("expected 1 result (only normal dir), got %d", len(results))
	}
}

func TestValidator_WithMaxFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTestMarkdownFiles(t, tmpDir, "file%d.md", 10)

	v := New(false).WithMaxFiles(3)
	ctx := context.Background()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results (max files limit), got %d", len(results))
	}
}

func TestValidator_WithMaxBlocks(t *testing.T) {
	t.Parallel()

	// Create file with 5 blocks
	block := "```go\npackage main\n```\n"
	content := []byte(block + block + block + block + block)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "multi.md")
	if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
		t.Fatal(err)
	}

	v := New(false).WithMaxBlocks(3)
	ctx := context.Background()
	results, err := v.ValidateFile(ctx, tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results (max blocks limit), got %d", len(results))
	}
}

func TestValidator_WithConcurrency(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTestMarkdownFiles(t, tmpDir, "file%d.md", 4)

	v := New(false).WithConcurrency(2)
	ctx := context.Background()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}

func TestValidator_ParallelValidation(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()

	// Create 8 files
	for i := 0; i < 8; i++ {
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("file%d.md", i))
		if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	v := New(false).WithConcurrency(4)
	ctx := context.Background()

	start := time.Now()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != 8 {
		t.Errorf("expected 8 results, got %d", len(results))
	}

	// With parallel processing, this should complete reasonably fast
	// Sequential would take longer with file I/O overhead
	if elapsed > 5*time.Second {
		t.Logf("Validation took %v - may be running sequentially", elapsed)
	}
}

func TestValidator_ChainMethods(t *testing.T) {
	t.Parallel()

	v := New(false).WithMaxFiles(10).WithMaxBlocks(5).WithConcurrency(3)

	if v.maxFiles != 10 {
		t.Errorf("expected maxFiles 10, got %d", v.maxFiles)
	}
	if v.maxBlocks != 5 {
		t.Errorf("expected maxBlocks 5, got %d", v.maxBlocks)
	}
	if v.concurrency != 3 {
		t.Errorf("expected concurrency 3, got %d", v.concurrency)
	}
}

func createTestMarkdownFiles(t *testing.T, tmpDir, pattern string, count int) {
	t.Helper()
	content := []byte("```go\npackage main\n```\n")
	for i := 0; i < count; i++ {
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf(pattern, i))
		if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
