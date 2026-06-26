package mdgovalidator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/testutil"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

func testdataPath(t *testing.T, filename string) string {
	t.Helper()

	p, err := filepath.Abs(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestIntegration_ValidGoFile(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, testdataPath(t, "valid_go.md"))
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if HasErrors(results) {
		for _, r := range results {
			if r.HasError() {
				t.Errorf("unexpected error at %s:%s: %v", r.File, r.LineNumber, r.Error)
			}
		}
	}

	testutil.AssertResultCount(t, results, 2)

	for _, r := range results {
		if r.Status != types.StatusValid {
			t.Errorf("expected StatusValid, got %s", r.Status)
		}
	}
}

func TestIntegration_InvalidGoFile(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, testdataPath(t, "invalid_go.md"))
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 2)

	if !HasErrors(results) {
		t.Error("expected errors in invalid Go file")
	}
}

func TestIntegration_SkippedBlocks(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, testdataPath(t, "skipped.md"))
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 3)

	skippedCount := 0

	for _, r := range results {
		if r.Status == types.StatusSkipped {
			skippedCount++
		}
	}

	if skippedCount != 2 {
		t.Errorf("expected 2 skipped blocks, got %d", skippedCount)
	}
}

func TestIntegration_MixedLanguagesMDX(t *testing.T) {
	t.Parallel()

	v := New(false).WithLanguages([]languages.Language{
		languages.LangGo,
		languages.LangTypeScript,
		languages.LangRust,
	})
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, testdataPath(t, "mixed.mdx"))
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 3)

	if HasErrors(results) {
		for _, r := range results {
			if r.HasError() {
				t.Errorf("unexpected error: %v", r.Error)
			}
		}
	}
}

func TestIntegration_EdgeCases(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, testdataPath(t, "edge_cases.md"))
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	testutil.AssertResultCount(t, results, 0)
}

func TestIntegration_ValidateDirectory(t *testing.T) {
	t.Parallel()

	v := New(false)
	ctx := context.Background()

	tdPath := testdataPath(t, ".")

	results, err := v.ValidateDirectory(ctx, tdPath)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results from testdata directory")
	}
}

func TestIntegration_MarkdownAltExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	content := []byte("# Test\n\n```go\npackage main\n```\n")
	altFile := filepath.Join(tmpDir, "test.markdown")

	err := os.WriteFile(altFile, content, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	v := New(false)
	ctx := context.Background()

	results, err := v.ValidateFile(ctx, altFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result for .markdown file, got %d", len(results))
	}

	if results[0].Status != types.StatusValid {
		t.Errorf("expected StatusValid, got %s", results[0].Status)
	}
}

func TestIntegration_VerboseDirectoryValidation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	content := []byte("# Test\n\n```go\npackage main\n```\n")

	err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	v := New(true)
	ctx := context.Background()

	results, err := v.ValidateDirectory(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory verbose error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results from verbose directory validation")
	}
}
