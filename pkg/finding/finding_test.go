package finding

import (
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
