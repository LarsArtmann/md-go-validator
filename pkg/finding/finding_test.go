package finding

import (
	"context"
	"errors"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

var errTestFinding = errors.New("some error")

func TestFromResult_Error(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("README.md")
	lineNum := types.NewLineNumber(42)
	blockIdx := types.NewBlockIndex(0)

	valErr := (&languages.ValidationError{
		Message: "Go syntax error: expected '}', found 'EOF'",
		Line:    3,
		Column:  1,
	}).WithCode(languages.ErrCodeSyntax)

	r := types.NewErrorResult(fileID, lineNum, blockIdx, "bad code", valErr)

	result, ok := FromResult(r)
	if !ok {
		t.Fatal("expected ok=true for error result")
	}

	if result.ToolName != ToolName {
		t.Errorf("expected tool %q, got %q", ToolName, result.ToolName)
	}

	if result.Rule != RuleName {
		t.Errorf("expected rule %q, got %q", RuleName, result.Rule)
	}

	if result.Severity != "error" {
		t.Errorf("expected severity 'error', got %q", result.Severity)
	}

	if result.Position.File != "README.md" {
		t.Errorf("expected file 'README.md', got %q", result.Position.File)
	}

	if result.Position.Line != 42 {
		t.Errorf("expected line 42, got %d", result.Position.Line)
	}
}

func TestFromResult_Valid(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("README.md")
	lineNum := types.NewLineNumber(42)
	blockIdx := types.NewBlockIndex(0)

	r := types.NewResultWithStatus(fileID, lineNum, blockIdx, "valid code", types.StatusValid)

	_, ok := FromResult(r)
	if ok {
		t.Error("expected ok=false for valid result")
	}
}

func TestFromResult_Skipped(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("README.md")
	lineNum := types.NewLineNumber(42)
	blockIdx := types.NewBlockIndex(0)

	r := types.NewResultWithStatus(fileID, lineNum, blockIdx, "skipped code", types.StatusSkipped)

	_, ok := FromResult(r)
	if ok {
		t.Error("expected ok=false for skipped result")
	}
}

func TestFromResults(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("test.md")
	lineNum := types.NewLineNumber(10)
	blockIdx := types.NewBlockIndex(0)

	valErr := (&languages.ValidationError{
		Message: "syntax error",
		Line:    1,
	}).WithCode(languages.ErrCodeSyntax)

	results := []types.Result{
		types.NewResultWithStatus(fileID, lineNum, blockIdx, "valid", types.StatusValid),
		types.NewErrorResult(fileID, lineNum, blockIdx, "bad", valErr),
		types.NewResultWithStatus(fileID, lineNum, types.NewBlockIndex(1), "valid2", types.StatusValid),
	}

	findings := FromResults(results)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	if findings[0].Position.File != "test.md" {
		t.Errorf("expected file 'test.md', got %q", findings[0].Position.File)
	}
}

func TestFromResults_Empty(t *testing.T) {
	t.Parallel()

	findings := FromResults(nil)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestFromResult_GenericError(t *testing.T) {
	t.Parallel()

	fileID := types.NewFileID("doc.md")
	lineNum := types.NewLineNumber(5)
	blockIdx := types.NewBlockIndex(0)

	r := types.NewErrorResult(fileID, lineNum, blockIdx, "code", errTestFinding)

	result, ok := FromResult(r)
	if !ok {
		t.Fatal("expected ok=true for error result")
	}

	if result.Message != "some error" {
		t.Errorf("expected message 'some error', got %q", result.Message)
	}
}

// TestFromResult_RoundTrip_RealValidationError exercises the full chain:
// broken Go code -> GoValidator.Validate -> real *ValidationError ->
// types.NewErrorResult (extracts ErrorCode via errors.AsType) ->
// finding.FromResult. This proves the conversion preserves the branded
// FilePath, source line, message, and severity against a real parser error,
// not a hand-constructed ValidationError.
func TestFromResult_RoundTrip_RealValidationError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := &languages.GoValidator{}

	// Unclosed function body: every strategy fails, yielding a real parse error.
	brokenCode := "func broken() {\n\tfmt.Println(\"no closing brace\"\n"

	err := validator.Validate(ctx, brokenCode)
	if err == nil {
		t.Fatal("expected validation error for broken Go code, got nil")
	}

	// NewErrorResult walks the error chain to extract ErrorCode.
	fileID := types.NewFileID("docs/examples.md")
	lineNum := types.NewLineNumber(15)
	blockIdx := types.NewBlockIndex(0)
	result := types.NewErrorResult(fileID, lineNum, blockIdx, brokenCode, err)

	// Round-trip: Result -> Finding.
	found, ok := FromResult(result)
	if !ok {
		t.Fatal("expected ok=true for error result")
	}

	if found.Position.File != "docs/examples.md" {
		t.Errorf("expected file 'docs/examples.md', got %q", found.Position.File)
	}

	if found.Position.Line != 15 {
		t.Errorf("expected line 15, got %d", found.Position.Line)
	}

	if found.ToolName != ToolName {
		t.Errorf("expected tool %q, got %q", ToolName, found.ToolName)
	}

	if found.Rule != RuleName {
		t.Errorf("expected rule %q, got %q", RuleName, found.Rule)
	}

	if found.Severity != "error" {
		t.Errorf("expected severity 'error', got %q", found.Severity)
	}

	if found.Message == "" {
		t.Error("expected non-empty message from real validation error")
	}

	if result.ErrorCode != languages.ErrCodeSyntax {
		t.Errorf("expected ErrorCode ErrCodeSyntax, got %v", result.ErrorCode)
	}
}
