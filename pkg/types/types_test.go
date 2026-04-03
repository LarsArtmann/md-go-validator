package types

import (
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
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

	t.Run("Validate", func(t *testing.T) {
		t.Parallel()
		testPositiveIntValidator(t, "LineNumber", NewLineNumber, "LineNumber must be >= 1")
	})
}

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

	t.Run("Validate", func(t *testing.T) {
		t.Parallel()
		testPositiveIntValidator(t, "BlockIndex", NewBlockIndex, "BlockIndex must be >= 1")
	})
}

// positiveIntValidator is a constraint for types with a Validate method.
type positiveIntValidator interface {
	Validate() error
}

func testPositiveIntValidator[TP positiveIntValidator](t *testing.T, name string, newFunc func(int) TP, _ string) {
	tests := []struct {
		value int
		valid bool
		desc  string
	}{
		{1, true, name + " >= 1"},
		{0, false, name + " == 0"},
		{1000000, true, "large " + name},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			v := newFunc(tc.value)
			err := v.Validate()
			if tc.valid && err != nil {
				t.Errorf("expected no error for %s, got %v", tc.desc, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("expected error for %s", tc.desc)
			}
		})
	}
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
		block := NewCodeBlock(NewLineNumber(10), languages.LangGo, "package main")
		if block.LineNumber.Int() != 10 {
			t.Errorf("expected LineNumber 10, got %d", block.LineNumber.Int())
		}
		if block.Language != languages.LangGo {
			t.Errorf("expected Language Go, got %v", block.Language)
		}
		if block.Code != "package main" {
			t.Errorf("expected 'package main', got %q", block.Code)
		}
		if block.Status != StatusUnknown {
			t.Errorf("expected StatusUnknown, got %v", block.Status)
		}
	})
}

func TestCodeBlockMarkMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		markFunc       func(*CodeBlock)
		expectedStatus Status
		checkFunc      func(*CodeBlock) bool
		expectedLabel  string
	}{
		{
			name:           "MarkValid",
			markFunc:       func(b *CodeBlock) { b.MarkValid() },
			expectedStatus: StatusValid,
			checkFunc:      func(b *CodeBlock) bool { return b.IsValid() },
			expectedLabel:  "IsValid()",
		},
		{
			name:           "MarkError",
			markFunc:       func(b *CodeBlock) { b.MarkError() },
			expectedStatus: StatusError,
			checkFunc:      func(b *CodeBlock) bool { return b.HasError() },
			expectedLabel:  "HasError()",
		},
		{
			name:           "MarkSkipped",
			markFunc:       func(b *CodeBlock) { b.MarkSkipped() },
			expectedStatus: StatusSkipped,
			checkFunc:      func(b *CodeBlock) bool { return b.IsSkipped() },
			expectedLabel:  "IsSkipped()",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var block CodeBlock
			tc.markFunc(&block)
			if block.Status != tc.expectedStatus {
				t.Errorf("expected %v, got %v", tc.expectedStatus, block.Status)
			}
			if !tc.checkFunc(&block) {
				t.Errorf("expected %s to return true", tc.expectedLabel)
			}
		})
	}
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
		AssertReportTotalAndValid(t, &report, 0, 0)
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			NewValidResult(NewFileID("a.md"), NewLineNumber(1), NewBlockIndex(1), "pkg"),
			NewValidResult(NewFileID("b.md"), NewLineNumber(1), NewBlockIndex(1), "pkg"),
		}
		report := BuildReportData(results, false)
		AssertReportSummary(t, &report, 2, 2, 0, 0)
	})
}

func errorResultsForTesting() []Result {
	return []Result{
		NewErrorResult(
			NewFileID("a.md"),
			NewLineNumber(1),
			NewBlockIndex(1),
			"pkg",
			&testError{msg: "err"},
		),
	}
}

func TestReportData_BuildReportData(t *testing.T) {
	t.Parallel()

	t.Run("with errors and show code", func(t *testing.T) {
		t.Parallel()
		results := errorResultsForTesting()
		report := BuildReportData(results, true)
		AssertSingleErrorWithCode(t, &report, "pkg")
	})

	t.Run("with errors and hide code", func(t *testing.T) {
		t.Parallel()
		results := errorResultsForTesting()
		report := BuildReportData(results, false)
		AssertSingleErrorWithCode(t, &report, "")
	})
}

func TestReportData_HasErrors(t *testing.T) {
	t.Parallel()

	testReportDataBoolCase(t, "has errors", ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 1}, true, func(r ReportData) bool {
		return r.HasErrors()
	})

	testReportDataBoolCase(t, "no errors", ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 0}, false, func(r ReportData) bool {
		return r.HasErrors()
	})
}

func TestReportData_Success(t *testing.T) {
	t.Parallel()

	testReportDataBoolCase(t, "success", ReportSummary{Total: 7, Valid: 5, Skipped: 2, Errors: 0}, true, func(r ReportData) bool {
		return r.Success()
	})

	testReportDataBoolCase(t, "failure", ReportSummary{Total: 0, Valid: 0, Skipped: 0, Errors: 1}, false, func(r ReportData) bool {
		return r.Success()
	})
}

func testReportDataBoolCase(t *testing.T, name string, summary ReportSummary, expected bool, method func(ReportData) bool) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		report := ReportData{
			Summary: summary,
			Errors:  nil,
		}
		if method(report) != expected {
			t.Errorf("expected %s to return %v, got %v", name, expected, method(report))
		}
	})
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
