package output

import (
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

func TestParseFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    OutputFormat
		wantErr bool
	}{
		{"table", "table", FormatTable, false},
		{"json", "json", FormatJSON, false},
		{"markdown", "markdown", FormatMarkdown, false},
		{"md alias", "md", FormatMarkdown, false},
		{"yaml", "yaml", FormatYAML, false},
		{"yml alias", "yml", FormatYAML, false},
		{"csv", "csv", FormatCSV, false},
		{"quiet", "quiet", FormatQuiet, false},
		{"q alias", "q", FormatQuiet, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildReportData(t *testing.T) {
	t.Parallel()

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{}
		report := types.BuildReportData(results, false)

		if report.Summary.Total != 0 {
			t.Errorf("expected Total 0, got %d", report.Summary.Total)
		}
		if report.Summary.Valid != 0 {
			t.Errorf("expected Valid 0, got %d", report.Summary.Valid)
		}
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), ""),
			types.NewValidResult(types.NewFileID("b.md"), types.NewLineNumber(2), types.NewBlockIndex(1), ""),
		}
		report := types.BuildReportData(results, false)

		if report.Summary.Total != 2 {
			t.Errorf("expected Total 2, got %d", report.Summary.Total)
		}
		if report.Summary.Valid != 2 {
			t.Errorf("expected Valid 2, got %d", report.Summary.Valid)
		}
		if report.Summary.Errors != 0 {
			t.Errorf("expected Errors 0, got %d", report.Summary.Errors)
		}
	})

	t.Run("with skipped", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewSkippedResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), ""),
			types.NewValidResult(types.NewFileID("b.md"), types.NewLineNumber(2), types.NewBlockIndex(1), ""),
		}
		report := types.BuildReportData(results, false)

		if report.Summary.Total != 2 {
			t.Errorf("expected Total 2, got %d", report.Summary.Total)
		}
		if report.Summary.Valid != 1 {
			t.Errorf("expected Valid 1, got %d", report.Summary.Valid)
		}
		if report.Summary.Skipped != 1 {
			t.Errorf("expected Skipped 1, got %d", report.Summary.Skipped)
		}
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "", &testError{msg: "syntax error"}),
			types.NewValidResult(types.NewFileID("b.md"), types.NewLineNumber(2), types.NewBlockIndex(1), ""),
		}
		report := types.BuildReportData(results, false)

		if report.Summary.Total != 2 {
			t.Errorf("expected Total 2, got %d", report.Summary.Total)
		}
		if report.Summary.Valid != 1 {
			t.Errorf("expected Valid 1, got %d", report.Summary.Valid)
		}
		if report.Summary.Errors != 1 {
			t.Errorf("expected Errors 1, got %d", report.Summary.Errors)
		}
		if len(report.Errors) != 1 {
			t.Errorf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Error != "syntax error" {
			t.Errorf("expected error message 'syntax error', got %q", report.Errors[0].Error)
		}
	})

	t.Run("show code", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main", &testError{msg: "syntax error"}),
		}
		report := types.BuildReportData(results, true)

		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Code != "package main" {
			t.Errorf("expected code 'package main', got %q", report.Errors[0].Code)
		}
	})

	t.Run("hide code", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main", &testError{msg: "syntax error"}),
		}
		report := types.BuildReportData(results, false)

		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Code != "" {
			t.Errorf("expected empty code, got %q", report.Errors[0].Code)
		}
	})
}

func TestTruncateCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		maxLen  uint
		want    string
	}{
		{"empty", "", 50, ""},
		{"short", "hello", 50, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncate", "this is a very long code snippet that should be truncated", 20, "this is a very lo..."},
		{"truncate to short", "hello world", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncateCode(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateCode(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func TestPrintReport(t *testing.T) {
	t.Parallel()

	t.Run("JSON format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatJSON, ColorModeNever, false)
	})

	t.Run("Markdown format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatMarkdown, ColorModeNever, false)
	})

	t.Run("Markdown format with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "bad code", &testError{msg: "syntax error"}),
		}
		PrintReport(results, FormatMarkdown, ColorModeNever, false)
		PrintReport(results, FormatMarkdown, ColorModeNever, true)
	})

	t.Run("YAML format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatYAML, ColorModeNever, false)
	})

	t.Run("CSV format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatCSV, ColorModeNever, false)
		PrintReport(results, FormatCSV, ColorModeNever, true)
	})

	t.Run("CSV format with error", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "bad code", &testError{msg: "syntax error"}),
		}
		PrintReport(results, FormatCSV, ColorModeNever, true)
	})

	t.Run("Quiet format with no errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatQuiet, ColorModeNever, false)
	})

	t.Run("Quiet format with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "bad code", &testError{msg: "syntax error"}),
		}
		PrintReport(results, FormatQuiet, ColorModeNever, false)
	})

	t.Run("Table format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
			types.NewSkippedResult(types.NewFileID("b.md"), types.NewLineNumber(2), types.NewBlockIndex(1), "// skip"),
		}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("Table format with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "bad\ncode", &testError{msg: "syntax error"}),
		}
		PrintReport(results, FormatTable, ColorModeNever, false)
		PrintReport(results, FormatTable, ColorModeNever, true)
	})

	t.Run("Table format with color", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatTable, ColorModeAlways, false)
	})

	t.Run("Table format with no errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("Table format empty results", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("default format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewValidResult(types.NewFileID("a.md"), types.NewLineNumber(1), types.NewBlockIndex(1), "package main"),
		}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})
}
