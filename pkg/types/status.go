package types

// ValidationStatus represents the validation status of a code block.
// Uses explicit enum instead of boolean for clarity.
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
