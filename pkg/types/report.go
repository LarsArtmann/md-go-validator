// Package types provides data structures for code block validation results.
//
//nolint:revive // Package name "types" is descriptive and widely used in Go projects
package types

// ReportData contains aggregated validation results for reporting.
// This type is designed for serialization to various output formats.
type ReportData struct {
	// Summary contains the aggregated counts.
	Summary ReportSummary `json:"summary" yaml:"summary"`

	// Errors contains the detailed error entries.
	Errors []ErrorEntry `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// ReportSummary contains counts of validation results.
type ReportSummary struct {
	// Total is the total number of code blocks processed.
	Total uint `json:"total" yaml:"total"`

	// Valid is the number of successfully validated code blocks.
	Valid uint `json:"valid" yaml:"valid"`

	// Skipped is the number of skipped code blocks.
	Skipped uint `json:"skipped" yaml:"skipped"`

	// Errors is the number of code blocks with validation errors.
	Errors uint `json:"errors" yaml:"errors"`
}

// ErrorEntry represents a single validation error for reporting.
type ErrorEntry struct {
	// File is the path to the file containing the error.
	File FileID `json:"file" yaml:"file"`

	// Line is the 1-based line number where the error occurred.
	Line LineNumber `json:"line" yaml:"line"`

	// Block is the 1-based index of the code block.
	Block BlockIndex `json:"block" yaml:"block"`

	// Error is the error message.
	Error string `json:"error" yaml:"error"`

	// Code is the code snippet that caused the error (optional).
	Code string `json:"code,omitempty" yaml:"code,omitempty"`
}

// BuildReportData aggregates a slice of Results into ReportData.
func BuildReportData(results []Result, showCode bool) ReportData {
	var valid, skipped, errorCount uint
	var errorEntries []ErrorEntry

	for _, r := range results {
		switch r.Status {
		case StatusUnknown:
			// Unknown status - no action needed
		case StatusValid:
			valid++
		case StatusSkipped:
			skipped++
		case StatusError:
			errorCount++
			entry := ErrorEntry{
				File:  r.File,
				Line:  r.LineNumber,
				Block: r.Block,
				Error: r.Error.Error(),
				Code:  "",
			}
			if showCode {
				entry.Code = r.Code
			}
			errorEntries = append(errorEntries, entry)
		}
	}

	return ReportData{
		Summary: ReportSummary{
			Total:   uint(len(results)),
			Valid:   valid,
			Skipped: skipped,
			Errors:  errorCount,
		},
		Errors: errorEntries,
	}
}

// HasErrors returns true if the report contains any errors.
func (d ReportData) HasErrors() bool {
	return d.Summary.Errors > 0
}

// Success returns true if all blocks were valid or skipped.
func (d ReportData) Success() bool {
	return d.Summary.Errors == 0
}
