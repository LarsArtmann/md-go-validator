package output

import (
	"fmt"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

func testParseFunc[T any](t *testing.T, funcName string, tests []struct {
	name    string
	input   string
	want    T
	wantErr bool
}, parseFunc func(string) (T, error),
) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseFunc(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%q) error = %v, wantErr %v", funcName, tt.input, err, tt.wantErr)
				return
			}
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("%s(%q) = %v, want %v", funcName, tt.input, got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    Format
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

	testParseFunc(t, "ParseFormat", tests, ParseFormat)
}

func TestParseColorMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    ColorMode
		wantErr bool
	}{
		{"auto", "auto", ColorModeAuto, false},
		{"always", "always", ColorModeAlways, false},
		{"never", "never", ColorModeNever, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	testParseFunc(t, "ParseColorMode", tests, ParseColorMode)
}

func TestBuildReportData(t *testing.T) {
	t.Parallel()

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{}
		report := types.BuildReportData(results, false)

		types.AssertReportTotalAndValid(t, &report, 0, 0)
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			newValidResultWithCode("a.md", 1, 1, ""),
			newValidResultWithCode("b.md", 2, 1, ""),
		}
		report := types.BuildReportData(results, false)

		types.AssertReportSummary(t, &report, 2, 2, 0, 0)
	})

	t.Run("with skipped", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			newSkippedResultWithReason("a.md", 1, 1, ""),
			newValidResultWithCode("b.md", 2, 1, ""),
		}
		report := types.BuildReportData(results, false)

		types.AssertReportSummary(t, &report, 2, 1, 0, 1)
	})

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			types.NewErrorResult(
				types.NewFileID("a.md"),
				types.NewLineNumber(1),
				types.NewBlockIndex(1),
				"",
				&testError{msg: "syntax error"},
			),
			newValidResultWithCode("b.md", 2, 1, ""),
		}
		report := types.BuildReportData(results, false)

		types.AssertReportSummary(t, &report, 2, 1, 1, 0)
		if len(report.Errors) != 1 {
			t.Errorf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Error != "syntax error" {
			t.Errorf("expected error message 'syntax error', got %q", report.Errors[0].Error)
		}
	})

	t.Run("show code", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{newErrorResultWithCode("package main")}
		report := types.BuildReportData(results, true)

		types.AssertSingleErrorWithCode(t, &report, "package main")
	})

	t.Run("hide code", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{newErrorResultWithCode("package main")}
		report := types.BuildReportData(results, false)

		types.AssertSingleErrorWithCode(t, &report, "")
	})
}

func TestTruncateCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen uint
		want   string
	}{
		{"empty", "", 50, ""},
		{"short", "hello", 50, "hello"},
		{"exact length", "hello", 5, "hello"},
		{
			"truncate",
			"this is a very long code snippet that should be truncated",
			20,
			"this is a very lo...",
		},
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
		assertPrintReportValid(t, FormatJSON, ColorModeNever, false)
	})

	t.Run("Markdown format", func(t *testing.T) {
		t.Parallel()
		assertPrintReportValid(t, FormatMarkdown, ColorModeNever, false)
	})

	t.Run("Markdown format with errors", func(t *testing.T) {
		t.Parallel()
		assertPrintReportWithErrors(t, FormatMarkdown, "bad code")
	})

	t.Run("YAML format", func(t *testing.T) {
		t.Parallel()
		assertPrintReportValid(t, FormatYAML, ColorModeNever, false)
	})

	t.Run("CSV format", func(t *testing.T) {
		t.Parallel()
		assertPrintReportValid(t, FormatCSV, ColorModeNever, false)
		assertPrintReportValid(t, FormatCSV, ColorModeNever, true)
	})

	t.Run("CSV format with error", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{newErrorResultWithCode("bad code")}
		PrintReport(results, FormatCSV, ColorModeNever, true)
	})

	t.Run("Quiet format with no errors", func(t *testing.T) {
		t.Parallel()
		assertPrintReportValid(t, FormatQuiet, ColorModeNever, false)
	})

	t.Run("Quiet format with errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{newErrorResultWithCode("bad code")}
		PrintReport(results, FormatQuiet, ColorModeNever, false)
	})

	t.Run("Table format", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{
			newValidResultWithCode("a.md", 1, 1, "package main"),
			newSkippedResultWithReason("b.md", 2, 1, "// skip"),
		}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("Table format with errors", func(t *testing.T) {
		t.Parallel()
		assertPrintReportWithErrors(t, FormatTable, "bad\ncode")
	})

	t.Run("Table format with color", func(t *testing.T) {
		t.Parallel()
		assertPrintReportValid(t, FormatTable, ColorModeAlways, false)
	})

	t.Run("Table format with no errors", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{newValidResultWithCode("a.md", 1, 1, "package main")}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("Table format empty results", func(t *testing.T) {
		t.Parallel()
		results := []types.Result{}
		PrintReport(results, FormatTable, ColorModeNever, false)
	})

	t.Run("default format", func(t *testing.T) {
		t.Parallel()
		assertPrintReportTable(t, "a.md", 1, 1, "package main")
	})
}

func newValidResultWithCode(fileID string, lineNumber, blockIndex int, code string) types.Result {
	return types.NewValidResult(
		types.NewFileID(fileID),
		types.NewLineNumber(lineNumber),
		types.NewBlockIndex(blockIndex),
		code,
	)
}

type validResultSpec struct {
	fileID string
	line   int
	block  int
	code   string
}

func newValidResults(specs ...validResultSpec) []types.Result {
	results := make([]types.Result, len(specs))
	for i, s := range specs {
		results[i] = newValidResultWithCode(s.fileID, s.line, s.block, s.code)
	}
	return results
}

func newErrorResultWithCode(code string) types.Result {
	return types.NewErrorResult(
		types.NewFileID("a.md"),
		types.NewLineNumber(1),
		types.NewBlockIndex(1),
		code,
		&testError{msg: "syntax error"},
	)
}

func newSkippedResultWithReason(
	fileID string,
	lineNumber, blockIndex int,
	reason string,
) types.Result {
	return types.NewSkippedResultForTest(fileID, lineNumber, blockIndex, reason)
}

func assertPrintReportValid(t *testing.T, format Format, colorMode ColorMode, quiet bool) {
	t.Helper()
	results := []types.Result{newValidResultWithCode("a.md", 1, 1, "package main")}
	PrintReport(results, format, colorMode, quiet)
}

func assertPrintReportTable(t *testing.T, fileID string, line, block int, code string) {
	t.Helper()
	results := []types.Result{newValidResultWithCode(fileID, line, block, code)}
	PrintReport(results, FormatTable, ColorModeNever, false)
}

func assertPrintReportWithErrors(t *testing.T, format Format, code string) {
	t.Helper()
	results := []types.Result{newErrorResultWithCode(code)}
	PrintReport(results, format, ColorModeNever, false)
	PrintReport(results, format, ColorModeNever, true)
}
