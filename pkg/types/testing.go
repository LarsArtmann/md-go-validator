package types

import "testing"

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
	if report.Summary.Total != total {
		t.Errorf("expected Total %d, got %d", total, report.Summary.Total)
	}
	if report.Summary.Valid != valid {
		t.Errorf("expected Valid %d, got %d", valid, report.Summary.Valid)
	}
	if report.Summary.Errors != errors {
		t.Errorf("expected Errors %d, got %d", errors, report.Summary.Errors)
	}
	if report.Summary.Skipped != skipped {
		t.Errorf("expected Skipped %d, got %d", skipped, report.Summary.Skipped)
	}
}
