package baseline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

var (
	errTestSyntax  = errors.New("syntax error")
	errTestGeneric = errors.New("err")
)

func TestSignature(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("README.md")
	lineNum := types.NewLineNumber(42)
	blockIdx := types.NewBlockIndex(0)

	r := types.NewErrorResult(fileID, lineNum, blockIdx, "bad code", errTestSyntax)

	sig := Signature(r)

	if sig != "README.md:42" {
		t.Errorf("expected 'README.md:42', got %q", sig)
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.txt")

	content := `# Known errors
README.md:10
docs/guide.md:25

# Another error
src/api.md:3
`

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	set, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if set.Count() != 3 {
		t.Errorf("expected 3 signatures, got %d", set.Count())
	}
}

func TestLoad_NotFound(t *testing.T) {
	t.Parallel()

	_, err := Load("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSet_Contains(t *testing.T) {
	t.Parallel()

	set := Set{signatures: map[string]bool{
		"README.md:10": true,
		"doc.md:5":     true,
	}}

	tests := []struct {
		name string
		file string
		line int
		want bool
	}{
		{"in baseline", "README.md", 10, true},
		{"not in baseline", "README.md", 11, false},
		{"different file", "other.md", 10, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := types.NewErrorResult(
				types.NewFileID(tc.file),
				types.NewLineNumber(tc.line),
				types.NewBlockIndex(0),
				"code",
				errTestGeneric,
			)

			if set.Contains(r) != tc.want {
				t.Errorf("Contains = %v, want %v", set.Contains(r), tc.want)
			}
		})
	}
}

func TestSet_FilterNew(t *testing.T) {
	t.Parallel()

	set := Set{signatures: map[string]bool{
		"doc.md:5": true,
	}}

	results := []types.Result{
		types.NewResultWithStatus(
			types.NewFileID("doc.md"),
			types.NewLineNumber(3),
			types.NewBlockIndex(0),
			"valid",
			types.StatusValid,
		),
		types.NewErrorResult(
			types.NewFileID("doc.md"),
			types.NewLineNumber(5),
			types.NewBlockIndex(1),
			"known error",
			errTestSyntax,
		),
		types.NewErrorResult(
			types.NewFileID("doc.md"),
			types.NewLineNumber(10),
			types.NewBlockIndex(2),
			"new error",
			errTestSyntax,
		),
	}

	filtered := set.FilterNew(results)

	// Valid result always passes, known error is filtered, new error passes.
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered results, got %d", len(filtered))
	}

	// First should be the valid result.
	if filtered[0].Status != types.StatusValid {
		t.Error("expected first result to be valid")
	}

	// Second should be the new error.
	if !filtered[1].HasError() {
		t.Error("expected second result to be an error")
	}

	if filtered[1].LineNumber.Int() != 10 {
		t.Errorf("expected line 10, got %d", filtered[1].LineNumber.Int())
	}
}

func TestSet_IsEmpty(t *testing.T) {
	t.Parallel()

	emptySet := Set{}
	if !emptySet.IsEmpty() {
		t.Error("expected empty set to be empty")
	}

	nonEmptySet := Set{signatures: map[string]bool{"a:1": true}}
	if nonEmptySet.IsEmpty() {
		t.Error("expected non-empty set to not be empty")
	}
}
