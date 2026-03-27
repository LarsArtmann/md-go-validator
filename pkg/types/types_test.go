package types

import (
	"testing"
)

func TestFileID(t *testing.T) {
	t.Parallel()

	t.Run("NewFileID", func(t *testing.T) {
		t.Parallel()
		fid := NewFileID("path/to/file.md")
		if fid.String() != "path/to/file.md" {
			t.Errorf("expected 'path/to/file.md', got %q", fid.String())
		}
	})

	t.Run("Validate non-empty", func(t *testing.T) {
		t.Parallel()
		fid := NewFileID("path/to/file.md")
		if err := fid.Validate(); err != nil {
			t.Errorf("expected no error for non-empty FileID, got %v", err)
		}
	})

	t.Run("Validate empty", func(t *testing.T) {
		t.Parallel()
		fid := FileID("")
		if err := fid.Validate(); err == nil {
			t.Error("expected error for empty FileID")
		}
	})
}

//nolint:dupl // Test structure mirrors TestBlockIndex for consistency
func TestLineNumber(t *testing.T) {
	t.Parallel()

	t.Run("NewLineNumber", func(t *testing.T) {
		t.Parallel()
		ln := NewLineNumber(42)
		if ln.Int() != 42 {
			t.Errorf("expected 42, got %d", ln.Int())
		}
		if ln.String() != "42" {
			t.Errorf("expected '42', got %q", ln.String())
		}
	})

	t.Run("Validate valid", func(t *testing.T) {
		t.Parallel()
		ln := NewLineNumber(1)
		if err := ln.Validate(); err != nil {
			t.Errorf("expected no error for LineNumber >= 1, got %v", err)
		}
	})

	t.Run("Validate zero", func(t *testing.T) {
		t.Parallel()
		ln := NewLineNumber(0)
		if err := ln.Validate(); err == nil {
			t.Error("expected error for LineNumber == 0")
		}
	})

	t.Run("Validate large value", func(t *testing.T) {
		t.Parallel()
		ln := NewLineNumber(1000000)
		if err := ln.Validate(); err != nil {
			t.Errorf("expected no error for large LineNumber, got %v", err)
		}
	})
}

//nolint:dupl // Test structure mirrors TestLineNumber for consistency
func TestBlockIndex(t *testing.T) {
	t.Parallel()

	t.Run("NewBlockIndex", func(t *testing.T) {
		t.Parallel()
		bi := NewBlockIndex(7)
		if bi.Int() != 7 {
			t.Errorf("expected 7, got %d", bi.Int())
		}
		if bi.String() != "7" {
			t.Errorf("expected '7', got %q", bi.String())
		}
	})

	t.Run("Validate valid", func(t *testing.T) {
		t.Parallel()
		bi := NewBlockIndex(1)
		if err := bi.Validate(); err != nil {
			t.Errorf("expected no error for BlockIndex >= 1, got %v", err)
		}
	})

	t.Run("Validate zero", func(t *testing.T) {
		t.Parallel()
		bi := NewBlockIndex(0)
		if err := bi.Validate(); err == nil {
			t.Error("expected error for BlockIndex == 0")
		}
	})

	t.Run("Validate large value", func(t *testing.T) {
		t.Parallel()
		bi := NewBlockIndex(500000)
		if err := bi.Validate(); err != nil {
			t.Errorf("expected no error for large BlockIndex, got %v", err)
		}
	})
}

func TestValidationStatus(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			status   ValidationStatus
			expected string
		}{
			{StatusUnknown, "unknown"},
			{StatusValid, "valid"},
			{StatusSkipped, "skipped"},
			{StatusError, "error"},
			{ValidationStatus(99), "unknown"},
		}
		for _, tc := range tests {
			if got := tc.status.String(); got != tc.expected {
				t.Errorf("Status %d: expected %q, got %q", tc.status, tc.expected, got)
			}
		}
	})

	t.Run("IsTerminal", func(t *testing.T) {
		t.Parallel()
		if !StatusValid.IsTerminal() {
			t.Error("StatusValid should be terminal")
		}
		if !StatusSkipped.IsTerminal() {
			t.Error("StatusSkipped should be terminal")
		}
		if !StatusError.IsTerminal() {
			t.Error("StatusError should be terminal")
		}
		if StatusUnknown.IsTerminal() {
			t.Error("StatusUnknown should not be terminal")
		}
	})

	t.Run("ParseValidationStatus", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			input    string
			expected ValidationStatus
			ok       bool
		}{
			{"valid", StatusValid, true},
			{"skipped", StatusSkipped, true},
			{"error", StatusError, true},
			{"invalid", StatusUnknown, false},
			{"", StatusUnknown, false},
		}
		for _, tc := range tests {
			got, ok := ParseValidationStatus(tc.input)
			if ok != tc.ok {
				t.Errorf("ParseValidationStatus(%q): expected ok=%v, got %v", tc.input, tc.ok, ok)
			}
			if got != tc.expected {
				t.Errorf(
					"ParseValidationStatus(%q): expected %v, got %v",
					tc.input,
					tc.expected,
					got,
				)
			}
		}
	})
}

func TestCodeBlock(t *testing.T) {
	t.Parallel()

	t.Run("NewCodeBlock", func(t *testing.T) {
		t.Parallel()
		block := NewCodeBlock(NewLineNumber(10), "package main")
		if block.LineNumber.Int() != 10 {
			t.Errorf("expected LineNumber 10, got %d", block.LineNumber.Int())
		}
		if block.Code != "package main" {
			t.Errorf("expected 'package main', got %q", block.Code)
		}
		if block.Status != StatusUnknown {
			t.Errorf("expected StatusUnknown, got %v", block.Status)
		}
	})

	t.Run("MarkSkipped", func(t *testing.T) {
		t.Parallel()
		var block CodeBlock
		block.MarkSkipped()
		if block.Status != StatusSkipped {
			t.Errorf("expected StatusSkipped, got %v", block.Status)
		}
		if !block.IsSkipped() {
			t.Error("expected IsSkipped() to return true")
		}
	})

	t.Run("MarkValid", func(t *testing.T) {
		t.Parallel()
		var block CodeBlock
		block.MarkValid()
		if block.Status != StatusValid {
			t.Errorf("expected StatusValid, got %v", block.Status)
		}
		if !block.IsValid() {
			t.Error("expected IsValid() to return true")
		}
	})

	t.Run("MarkError", func(t *testing.T) {
		t.Parallel()
		var block CodeBlock
		block.MarkError()
		if block.Status != StatusError {
			t.Errorf("expected StatusError, got %v", block.Status)
		}
		if !block.HasError() {
			t.Error("expected HasError() to return true")
		}
	})
}

func TestResult(t *testing.T) {
	t.Parallel()

	t.Run("NewValidResult", func(t *testing.T) {
		t.Parallel()
		r := NewValidResult(
			NewFileID("test.md"),
			NewLineNumber(5),
			NewBlockIndex(1),
			"package main",
		)
		if r.File != NewFileID("test.md") {
			t.Errorf("expected FileID test.md, got %v", r.File)
		}
		if r.Status != StatusValid {
			t.Errorf("expected StatusValid, got %v", r.Status)
		}
	})

	t.Run("NewSkippedResult", func(t *testing.T) {
		t.Parallel()
		r := NewSkippedResult(NewFileID("test.md"), NewLineNumber(5), NewBlockIndex(1), "skip me")
		if r.Status != StatusSkipped {
			t.Errorf("expected StatusSkipped, got %v", r.Status)
		}
	})

	t.Run("NewErrorResult", func(t *testing.T) {
		t.Parallel()
		err := &testError{msg: "syntax error"}
		r := NewErrorResult(
			NewFileID("test.md"),
			NewLineNumber(5),
			NewBlockIndex(1),
			"invalid",
			err,
		)
		if r.Status != StatusError {
			t.Errorf("expected StatusError, got %v", r.Status)
		}
		if !r.HasError() {
			t.Error("expected HasError() to return true")
		}
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		r := NewValidResult(
			NewFileID("test.md"),
			NewLineNumber(5),
			NewBlockIndex(1),
			"package main",
		)
		s := r.String()
		if s != "test.md:5 (block #1): valid" {
			t.Errorf("unexpected string: %q", s)
		}
	})

	t.Run("Summary", func(t *testing.T) {
		t.Parallel()
		r := NewValidResult(
			NewFileID("test.md"),
			NewLineNumber(5),
			NewBlockIndex(1),
			"package main",
		)
		summary := r.Summary()
		if summary == "" {
			t.Error("expected non-empty summary")
		}
	})
}

func TestBuildReportData(t *testing.T) {
	t.Parallel()

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()
		report := BuildReportData([]Result{}, false)
		if report.Summary.Total != 0 {
			t.Errorf("expected Total 0, got %d", report.Summary.Total)
		}
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			NewValidResult(NewFileID("a.md"), NewLineNumber(1), NewBlockIndex(1), "pkg"),
			NewValidResult(NewFileID("b.md"), NewLineNumber(1), NewBlockIndex(1), "pkg"),
		}
		report := BuildReportData(results, false)
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

	t.Run("with errors and show code", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			NewErrorResult(
				NewFileID("a.md"),
				NewLineNumber(1),
				NewBlockIndex(1),
				"pkg",
				&testError{msg: "err"},
			),
		}
		report := BuildReportData(results, true)
		if len(report.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(report.Errors))
		}
		if report.Errors[0].Code != "pkg" {
			t.Errorf("expected code 'pkg', got %q", report.Errors[0].Code)
		}
	})

	t.Run("with errors and hide code", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			NewErrorResult(
				NewFileID("a.md"),
				NewLineNumber(1),
				NewBlockIndex(1),
				"pkg",
				&testError{msg: "err"},
			),
		}
		report := BuildReportData(results, false)
		if report.Errors[0].Code != "" {
			t.Errorf("expected empty code, got %q", report.Errors[0].Code)
		}
	})
}

func TestReportData_HasErrors(t *testing.T) {
	t.Parallel()

	t.Run("has errors", func(t *testing.T) {
		t.Parallel()
		report := ReportData{
			Summary: ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 1},
			Errors:  nil,
		}
		if !report.HasErrors() {
			t.Error("expected HasErrors() to return true")
		}
	})

	t.Run("no errors", func(t *testing.T) {
		t.Parallel()
		report := ReportData{
			Summary: ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 0},
			Errors:  nil,
		}
		if report.HasErrors() {
			t.Error("expected HasErrors() to return false")
		}
	})
}

func TestReportData_Success(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		report := ReportData{
			Summary: ReportSummary{Total: 7, Valid: 5, Skipped: 2, Errors: 0},
			Errors:  nil,
		}
		if !report.Success() {
			t.Error("expected Success() to return true")
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		report := ReportData{
			Summary: ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 1},
			Errors:  nil,
		}
		if report.Success() {
			t.Error("expected Success() to return false")
		}
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
