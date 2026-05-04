package types

import (
	"errors"
	"fmt"
)

var errUnsupportedStatus = errors.New("unsupported validation status")

// ValidationStatus represents the validation status of a code block.
// Uses explicit enum instead of boolean for clarity.
//
//nolint:recvcheck // UnmarshalText must use pointer receiver to mutate
type ValidationStatus uint

// Validation status constants.
const (
	StatusUnknown ValidationStatus = iota
	StatusValid
	StatusSkipped
	StatusError
)

// String returns the string representation of the status.
func (s ValidationStatus) String() string {
	switch s {
	case StatusUnknown:
		return "unknown"
	case StatusValid:
		return "valid"
	case StatusSkipped:
		return "skipped"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// IsTerminal returns true if this is a terminal (final) status.
func (s ValidationStatus) IsTerminal() bool {
	return s == StatusValid || s == StatusSkipped || s == StatusError
}

// MarshalText implements encoding.TextMarshaler for JSON/YAML serialization.
func (s ValidationStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON/YAML deserialization.
func (s *ValidationStatus) UnmarshalText(text []byte) error {
	parsed, ok := ParseValidationStatus(string(text))
	if !ok {
		return fmt.Errorf("%w: %s", errUnsupportedStatus, string(text))
	}

	*s = parsed

	return nil
}

// ParseValidationStatus parses a string into ValidationStatus.
func ParseValidationStatus(s string) (ValidationStatus, bool) {
	switch s {
	case "valid":
		return StatusValid, true
	case "skipped":
		return StatusSkipped, true
	case "error":
		return StatusError, true
	default:
		return StatusUnknown, false
	}
}
