package types

import (
	"fmt"
	"strings"
)

// Result contains the result of validating a single code block.
// This is the primary output type of the validation process.
//
// Invariant: Status == StatusError if and only if Error != nil.
// The constructors below preserve this invariant; use Result.Validate() to
// check results constructed via struct literals (e.g. in tests).
type Result struct {
	// File is the path to the file being validated.
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

var (
	errErrorStatusWithoutError = fmt.Errorf("status is %s but Error is nil", StatusError)
	errErrorWithoutErrorStatus = fmt.Errorf("Error is non-nil but status is %s", StatusUnknown)
)

// newResultWithStatusAndError creates a Result with status and optional error.
// It enforces the StatusError ⟺ Error!=nil invariant.
func newResultWithStatusAndError(
	file FileID,
	line LineNumber,
	block BlockIndex,
	code string,
	status ValidationStatus,
	err error,
) Result {
	r := Result{
		File:       file,
		LineNumber: line,
		Block:      block,
		Code:       code,
		Status:     status,
		Error:      err,
	}

	vErr := r.Validate()
	if vErr != nil {
		panic(fmt.Sprintf("types: invalid Result constructed: %v", vErr))
	}

	return r
}

// NewResultWithStatus creates a Result with the given non-error status.
// Panics if status is StatusError — use NewErrorResult for error results.
func NewResultWithStatus(
	file FileID,
	line LineNumber,
	block BlockIndex,
	code string,
	status ValidationStatus,
) Result {
	return newResultWithStatusAndError(file, line, block, code, status, nil)
}

// NewErrorResult creates a new error result.
// Panics if err is nil — an error result requires an error.
func NewErrorResult(file FileID, line LineNumber, block BlockIndex, code string, err error) Result {
	return newResultWithStatusAndError(file, line, block, code, StatusError, err)
}

// Validate enforces the Result invariant: StatusError holds iff Error is non-nil.
// Returns an error describing the violation, or nil if the result is consistent.
func (r Result) Validate() error {
	if r.Status == StatusError && r.Error == nil {
		return errErrorStatusWithoutError
	}

	if r.Error != nil && r.Status != StatusError {
		return fmt.Errorf("Error is non-nil but status is %s: %w", r.Status, errErrorWithoutErrorStatus)
	}

	return nil
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
