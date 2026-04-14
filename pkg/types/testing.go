package types

import "testing"

// TestError is a simple error implementation for testing.
type TestError struct {
	msg string
}

// NewTestError creates a new TestError with the given message.
func NewTestError(msg string) *TestError {
	return &TestError{msg: msg}
}

// Error implements the error interface.
func (e *TestError) Error() string {
	return e.msg
}

func AssertSingleError(t *testing.T, report *ReportData) {
	t.Helper()

	if len(report.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(report.Errors))
	}
}

func AssertSingleErrorWithCode(t *testing.T, report *ReportData, expectedCode string) {
	t.Helper()
	AssertSingleError(t, report)

	if report.Errors[0].Code != expectedCode {
		t.Errorf("expected code %q, got %q", expectedCode, report.Errors[0].Code)
	}
}

func AssertReportTotalAndValid(t *testing.T, report *ReportData, total, valid uint) {
	t.Helper()

	if report.Summary.Total != total {
		t.Errorf("expected Total %d, got %d", total, report.Summary.Total)
	}

	if report.Summary.Valid != valid {
		t.Errorf("expected Valid %d, got %d", valid, report.Summary.Valid)
	}
}

func AssertReportSummary(t *testing.T, report *ReportData, total, valid, errors, skipped uint) {
	t.Helper()
	AssertReportTotalAndValid(t, report, total, valid)

	if report.Summary.Errors != errors {
		t.Errorf("expected Errors %d, got %d", errors, report.Summary.Errors)
	}

	if report.Summary.Skipped != skipped {
		t.Errorf("expected Skipped %d, got %d", skipped, report.Summary.Skipped)
	}
}

func NewSkippedResultForTest(
	fileID string,
	lineNumber, blockIndex int,
	reason string,
) Result {
	return NewResultWithStatus(
		NewFileID(fileID),
		NewLineNumber(lineNumber),
		NewBlockIndex(blockIndex),
		reason,
		StatusSkipped,
	)
}
