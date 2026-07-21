package mdgovalidator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	codeutil "github.com/larsartmann/md-go-validator/pkg/code"
	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/testutil"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

const (
	testPackageMain = "package main"
	testFileID      = "test.md"
)

type validResultSpec struct {
	fileID string
	line   int
	block  int
	code   string
}

func newValidResult(spec validResultSpec) types.Result {
	return types.NewResultWithStatus(
		types.NewFileID(spec.fileID),
		types.NewLineNumber(spec.line),
		types.NewBlockIndex(spec.block),
		spec.code,
		types.StatusValid,
	)
}

func newValidResults(specs ...validResultSpec) []types.Result {
	results := make([]types.Result, 0, len(specs))
	for _, s := range specs {
		results = append(results, newValidResult(s))
	}

	return results
}

func newTestErrorResult(fileID string, line, block int, code, errMsg string) types.Result {
	return testutil.NewTestErrorResult(fileID, line, block, code, errMsg)
}

func TestExtractGoCodeBlocks(t *testing.T) {
	t.Parallel()

	t.Run("no code blocks", func(t *testing.T) {
		t.Parallel()

		blocks := ExtractGoCodeBlocks("Just text\nNo code here")
		testutil.AssertBlockCount(t, blocks, 0)
	})

	runExtractGoBlockAtLineTest(t, "single go block",
		"Some text\n```go\nfmt.Println(\"hello\")\n```\nMore text", 2)
	runExtractGoBlockAtLineTest(t, "skip other languages",
		"```python\nprint('hello')\n```\n```go\nfmt.Println(\"hello\")\n```", 4)

	t.Run("skip directive before block", func(t *testing.T) {
		t.Parallel()

		content := "<!-- skip-validate -->\n```go\npartial code\n```"

		blocks := ExtractGoCodeBlocks(content)
		b := testutil.AssertSingleBlock(t, blocks)

		if !b.IsSkipped() {
			t.Error("expected block to be skipped")
		}
	})

	t.Run("golang tag", func(t *testing.T) {
		t.Parallel()

		content := "```golang\nfmt.Println(\"hello\")\n```"
		_ = extractAndAssertBlockCount(t, content, 1)
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

	blocks := extractAndAssertBlockCount(t, content, 1)
	assertBlockSkipped(t, blocks[0])
}

func TestExtractGoCodeBlocks_SkipInCode(t *testing.T) {
	t.Parallel()

	content := "```go\n//nolint\ntype Partial struct{}\n```"

	blocks := extractAndAssertBlockCount(t, content, 1)
	assertBlockSkipped(t, blocks[0])
}

func TestExtractGoCodeBlocks_EmptyBlock(t *testing.T) {
	t.Parallel()

	content := "```go\n\n```"

	blocks := ExtractGoCodeBlocks(content)
	testutil.AssertBlockCount(t, blocks, 0)
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

			err := ValidateGoCode(tc.code)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}

	invalidCases := []struct {
		name string
		code string
	}{
		{"invalid syntax", "func broken {"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateGoCode(tc.code)
			if err == nil {
				t.Error("expected error for invalid syntax")
			}
		})
	}
}

func TestIndentCode(t *testing.T) {
	t.Parallel()

	input := "line1\nline2\n\nline4"
	expected := "\tline1\n\tline2\n\n\tline4\n"

	result := codeutil.IndentCode(input)
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

	for _, ext := range []string{".md", ".mdx"} {
		t.Run(ext, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tmpFile := testutil.WriteTestFile(t, tmpDir, "test"+ext, content)

			v := New(false)
			ctx := context.Background()

			results, err := v.ValidateFile(ctx, tmpFile)
			if err != nil {
				t.Fatalf("ValidateFile error: %v", err)
			}

			testutil.AssertResultCount(t, results, 1)
		})
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
	testutil.WriteTestFile(t, tmpDir, "test.md", content)
	testutil.WriteTestFile(t, tmpDir, "test.txt", content)

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	testutil.AssertResultCount(t, results, 1)
}

func TestValidator_ValidateDirectory_MDX(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	testutil.WriteTestFile(t, tmpDir, "test.md", content)
	testutil.WriteTestFile(t, tmpDir, "doc.mdx", content)
	testutil.WriteTestFile(t, tmpDir, "test.txt", content)

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	testutil.AssertResultCount(t, results, 2)
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

		results := newValidResults(
			validResultSpec{fileID: testFileID, line: 1, block: 1, code: testPackageMain},
			validResultSpec{fileID: testFileID, line: 5, block: 2, code: testPackageMain},
		)
		if HasErrors(results) {
			t.Error("expected false for valid results")
		}
	})

	t.Run("skipped doesn't count", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{
			types.NewSkippedResultForTest("test.md", 1, 1, "skipped"),
		}
		if HasErrors(results) {
			t.Error("expected false for skipped error")
		}
	})

	t.Run("has error", func(t *testing.T) {
		t.Parallel()

		results := []types.Result{
			newTestErrorResult("test.md", 1, 1, "invalid", "test error"),
		}
		if !HasErrors(results) {
			t.Error("expected true for error result")
		}
	})
}

func TestValidator_ValidateDirectory_Cancellation(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()
	testutil.WriteTestFile(t, tmpDir, "test.md", content)

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
	testutil.AssertMaxResults(t, results, 5)
}

func TestValidator_ValidateFile_Empty(t *testing.T) {
	t.Parallel()

	content := []byte("No code blocks here")

	tmpDir := t.TempDir()
	tmpFile := testutil.WriteTestFile(t, tmpDir, "empty.md", content)

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 0)
}

func TestValidator_ValidateDirectory_SkipDirs(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()

	// Create files in various directories
	subdirs := []string{".hidden", "vendor", "node_modules", "build", "dist", "normal"}
	for _, dir := range subdirs {
		dirPath := filepath.Join(tmpDir, dir)

		err := os.MkdirAll(dirPath, 0o750)
		if err != nil {
			t.Fatal(err)
		}

		testutil.WriteTestFile(t, dirPath, "test.md", content)
	}

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	// Only normal directory should be processed
	testutil.AssertResultCount(t, results, 1)
}

func TestValidator_ValidateDirectoryFunc_Streaming(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	testutil.WriteTestFile(t, tmpDir, "a.md", content)
	testutil.WriteTestFile(t, tmpDir, "b.md", content)
	testutil.WriteTestFile(t, tmpDir, "c.md", content)

	v := New(false)
	ctx := context.Background()

	var count int

	err := v.ValidateDirectoryFunc(ctx, tmpDir, func(_ types.Result) error {
		count++

		return nil
	})
	if err != nil {
		t.Fatalf("ValidateDirectoryFunc error: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 callback calls, got %d", count)
	}
}

func TestValidator_ValidateDirectoryFunc_EarlyAbort(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	testutil.WriteTestFile(t, tmpDir, "a.md", content)
	testutil.WriteTestFile(t, tmpDir, "b.md", content)

	v := New(false)
	ctx := context.Background()

	errAbort := errors.New("abort") //nolint:err113 // test sentinel error

	var count int

	err := v.ValidateDirectoryFunc(ctx, tmpDir, func(_ types.Result) error {
		count++

		return errAbort
	})
	if err == nil {
		t.Fatal("expected error from early abort")
	}

	if !errors.Is(err, errAbort) { //nolint:legacyerrors // value sentinel
		t.Errorf("expected errAbort in chain, got %v", err)
	}
}

func TestValidator_ValidateDirectoryFunc_EmptyDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	v := New(false)
	ctx := context.Background()

	var called bool

	err := v.ValidateDirectoryFunc(ctx, tmpDir, func(_ types.Result) error {
		called = true

		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("callback should not be called for empty directory")
	}
}

func TestValidator_WithMaxFiles(t *testing.T) {
	t.Parallel()

	validateDirectoryWithFiles(t, 10, func(v *FileValidator) *FileValidator {
		return v.WithMaxFiles(3)
	}, 3, "expected 3 results (max files limit)")
}

func TestValidator_WithConcurrency(t *testing.T) {
	t.Parallel()

	validateDirectoryWithFiles(t, 4, func(v *FileValidator) *FileValidator {
		return v.WithConcurrency(2)
	}, 4, "expected 4 results")
}

func TestValidator_WithMaxBlocks(t *testing.T) {
	t.Parallel()

	// Create file with 5 blocks
	block := "```go\npackage main\n```\n"
	content := []byte(block + block + block + block + block)
	tmpDir := t.TempDir()
	tmpFile := testutil.WriteTestFile(t, tmpDir, "multi.md", content)

	v := New(false).WithMaxBlocks(3)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 3)
}

func TestValidator_ParallelValidation(t *testing.T) {
	t.Parallel()

	content := []byte("```go\npackage main\n```\n")
	tmpDir := t.TempDir()

	// Create 8 files
	for i := range 8 {
		testutil.WriteTestFile(t, tmpDir, fmt.Sprintf("file%d.md", i), content)
	}

	v := New(false).WithConcurrency(4)
	ctx := context.Background()

	start := time.Now()
	results, err := v.ValidateDirectory(ctx, tmpDir)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	testutil.AssertResultCount(t, results, 8)

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

func TestIsSupportedFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		expected bool
	}{
		{"README.md", true},
		{"docs/guide.markdown", true},
		{"blog/post.mdx", true},
		{"README.MD", true},
		{"readme.Mdx", true},
		{"file.txt", false},
		{"script.go", false},
		{"style.css", false},
		{"Makefile", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			got := IsSupportedFile(tt.path)
			if got != tt.expected {
				t.Errorf("IsSupportedFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func createTestMarkdownFiles(t *testing.T, tmpDir, pattern string, count int) {
	t.Helper()

	content := []byte("```go\npackage main\n```\n")
	for i := range count {
		testutil.WriteTestFile(t, tmpDir, fmt.Sprintf(pattern, i), content)
	}
}

func validateDirectoryWithFiles(
	t *testing.T,
	fileCount int,
	configure func(*FileValidator) *FileValidator,
	expectedResults int,
	msg string,
) {
	t.Helper()
	tmpDir := t.TempDir()
	createTestMarkdownFiles(t, tmpDir, "file%d.md", fileCount)

	v := configure(New(false))
	ctx := context.Background()

	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != expectedResults {
		t.Errorf("%s, got %d", msg, len(results))
	}
}

func assertBlockAtLine(t *testing.T, block types.CodeBlock, expectedLine int) {
	t.Helper()

	if block.LineNumber != types.NewLineNumber(expectedLine) {
		t.Errorf("expected line %d, got %d", expectedLine, block.LineNumber)
	}
}

func assertBlockSkipped(t *testing.T, block types.CodeBlock) {
	t.Helper()

	if !block.IsSkipped() {
		t.Error("expected block to be skipped")
	}
}

func extractAndAssertBlockCount(t *testing.T, content string, _ int) []types.CodeBlock {
	t.Helper()

	blocks := ExtractGoCodeBlocks(content)
	testutil.AssertBlockCount(t, blocks, 1)

	return blocks
}

func extractAndAssertBlockAtLine(t *testing.T, content string, expectedLine int) []types.CodeBlock {
	t.Helper()
	blocks := extractAndAssertBlockCount(t, content, 1)
	assertBlockAtLine(t, blocks[0], expectedLine)

	return blocks
}

func runExtractGoBlockAtLineTest(t *testing.T, name, content string, expectedLine int) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()
		extractAndAssertBlockAtLine(t, content, expectedLine)
	})
}

func TestValidator_WithLanguages(t *testing.T) {
	t.Parallel()

	v := New(false).WithLanguages([]languages.Language{languages.LangGo})

	if len(v.targetLangs) != 1 || v.targetLangs[0] != languages.LangGo {
		t.Errorf("expected [go], got %v", v.targetLangs)
	}
}

func TestValidator_WithRegistry(t *testing.T) {
	t.Parallel()

	customReg := languages.NewRegistry()

	v := New(false).WithRegistry(customReg)

	if v.registry != customReg {
		t.Error("expected custom registry to be set")
	}
}

func TestSupportedExtensions(t *testing.T) {
	t.Parallel()

	exts := SupportedExtensions()

	if len(exts) != 3 {
		t.Fatalf("expected 3 extensions, got %d", len(exts))
	}

	expected := map[string]bool{".md": true, ".markdown": true, ".mdx": true}
	for _, ext := range exts {
		if !expected[ext.String()] {
			t.Errorf("unexpected extension: %s", ext)
		}
	}
}

func TestExtractCodeBlocksWithConfig(t *testing.T) {
	t.Parallel()

	content := "```go\nfmt.Println(\"hello\")\n```\n"
	customConfig := SkipDirectivesConfig{"// custom-skip"}

	blocks := ExtractCodeBlocksWithConfig(
		content,
		[]languages.Language{languages.LangGo},
		customConfig,
	)
	testutil.AssertBlockCount(t, blocks, 1)
}

func TestHasErrors_NoErrors(t *testing.T) {
	t.Parallel()

	results := []types.Result{
		newValidResult(validResultSpec{fileID: "test.md", line: 1, block: 1, code: "package main"}),
	}

	if HasErrors(results) {
		t.Error("expected no errors")
	}
}

func TestValidator_ValidateContent(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	validate := func(t *testing.T, content string) []types.Result {
		t.Helper()

		results, err := v.ValidateContent(ctx, content, "<test>")
		if err != nil {
			t.Fatalf("ValidateContent error: %v", err)
		}

		return results
	}

	cases := []struct {
		name    string
		content string
		count   int
		wantErr bool
	}{
		{name: "valid go content", content: "```go\npackage main\nfunc main() {}\n```\n", count: 1},
		{name: "invalid go content", content: "```go\nbroken syntax\n```\n", count: 1, wantErr: true},
		{name: "no code blocks", content: "# Just markdown\n\nNo code here.\n", count: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			results := validate(t, tc.content)
			testutil.AssertResultCount(t, results, tc.count)

			if tc.wantErr && !HasErrors(results) {
				t.Error("expected validation errors")
			}
		})
	}

	t.Run("cancelled context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := v.ValidateContent(ctx, "```go\npackage main\n```\n", "<test>")
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})
}

func TestFormatSupportedExtensions(t *testing.T) {
	t.Parallel()

	result := formatSupportedExtensions()
	if result == "" {
		t.Error("formatSupportedExtensions() returned empty string")
	}

	for _, ext := range SupportedExtensions() {
		if !strings.Contains(result, ext.String()) {
			t.Errorf("formatSupportedExtensions() = %q, missing extension %s", result, ext)
		}
	}
}
