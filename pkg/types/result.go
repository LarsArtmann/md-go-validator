package types

import (
	"fmt"
	"strings"
)

// Result contains the result of validating a single code block.
// This is the primary output type of the validation process.
type Result struct {
	// File is the path to the markdown file being validated.
	File FileID

	// LineNumber is the 1-based line number where the code block starts.
	LineNumber LineNumber

	// Block is the 1-based index of the code block within the file.
	Block BlockIndex

	// Code is the actual Go source code content.
	Code string

	// Status indicates whether the block is valid, skipped, or has an error.
	Status ValidationStatus

	// Error is the validation error if Status is StatusError.
	// nil otherwise.
	Error error
}

// newResult creates a Result with the given status and no error.
func newResult(
	file FileID,
	line LineNumber,
	block BlockIndex,
	code string,
	status ValidationStatus,
) Result {
	return Result{
		File:       file,
		LineNumber: line,
		Block:      block,
		Code:       code,
		Status:     status,
		Error:      nil,
	}
}

// NewValidResult creates a new valid result.
func NewValidResult(file FileID, line LineNumber, block BlockIndex, code string) Result {
	return newResult(file, line, block, code, StatusValid)
}

// NewSkippedResult creates a new skipped result.
func NewSkippedResult(file FileID, line LineNumber, block BlockIndex, code string) Result {
	return newResult(file, line, block, code, StatusSkipped)
}

// NewErrorResult creates a new error result.
func NewErrorResult(file FileID, line LineNumber, block BlockIndex, code string, err error) Result {
	result := newResult(file, line, block, code, StatusError)
	result.Error = err
	return result
}

// String returns a human-readable summary of the result.
func (r Result) String() string {
	if r.Error != nil {
		return fmt.Sprintf("%s:%s (block #%s): %v", r.File, r.LineNumber, r.Block, r.Error)
	}
	return fmt.Sprintf("%s:%s (block #%s): %s", r.File, r.LineNumber, r.Block, r.Status)
}

// HasError returns true if this result has an error.
func (r Result) HasError() bool {
	return r.Status == StatusError && r.Error != nil
}

// Summary returns a one-line summary suitable for logging.
func (r Result) Summary() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("file=%s", r.File))
	parts = append(parts, fmt.Sprintf("line=%s", r.LineNumber))
	parts = append(parts, fmt.Sprintf("block=%s", r.Block))
	parts = append(parts, fmt.Sprintf("status=%s", r.Status))
	if r.Error != nil {
		parts = append(parts, fmt.Sprintf("error=%v", r.Error))
	}
	return strings.Join(parts, " ")
}
