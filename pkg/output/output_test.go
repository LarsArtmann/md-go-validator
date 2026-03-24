package output

import (
	"testing"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
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
		results := []mdgovalidator.Result{}
		report := buildReportData(results, false)

		if report.Summary.Total != 0 {
			t.Errorf("expected Total 0, got %d", report.Summary.Total)
		}
		if report.Summary.Valid != 0 {
			t.Errorf("expected Valid 0, got %d", report.Summary.Valid)
		}
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []mdgovalidator.Result{
			{File: "a.md", LineNumber: 1, CodeBlock: 1, Skipped: false, Error: nil},
			{File: "b.md", LineNumber: 2, CodeBlock: 1, Skipped: false, Error: nil},
		}
		report := buildReportData(results, false)

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
		results := []mdgovalidator.Result{
			{File: "a.md", LineNumber: 1, CodeBlock: 1, Skipped: true, Error: nil},
			{File: "b.md", LineNumber: 2, CodeBlock: 1, Skipped: false, Error: nil},
		}
		report := buildReportData(results, false)

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
		results := []mdgovalidator.Result{
			{File: "a.md", LineNumber: 1, CodeBlock: 1, Skipped: false, Error: &testError{msg: "syntax error"}},
			{File: "b.md", LineNumber: 2, CodeBlock: 1, Skipped: false, Error: nil},
		}
		report := buildReportData(results, false)

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
		results := []mdgovalidator.Result{
			{File: "a.md", LineNumber: 1, CodeBlock: 1, Code: "package main", Skipped: false, Error: &testError{msg: "syntax error"}},
		}
		report := buildReportData(results, true)

		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Code != "package main" {
			t.Errorf("expected code 'package main', got %q", report.Errors[0].Code)
		}
	})

	t.Run("hide code", func(t *testing.T) {
		t.Parallel()
		results := []mdgovalidator.Result{
			{File: "a.md", LineNumber: 1, CodeBlock: 1, Code: "package main", Skipped: false, Error: &testError{msg: "syntax error"}},
		}
		report := buildReportData(results, false)

		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error entry, got %d", len(report.Errors))
		}
		if report.Errors[0].Code != "" {
			t.Errorf("expected empty code, got %q", report.Errors[0].Code)
		}
	})
}

func TestEscapeCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello", "hello"},
		{"with comma", "hello,world", "\"hello,world\""},
		{"with quote", `say "hello"`, `"say ""hello"""`},
		{"with newline", "line1\nline2", "\"line1\nline2\""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := escapeCSV(tt.input)
			if got != tt.want {
				t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
