package types

import (
	"errors"
	"strconv"
)

// FileID is a branded type representing a file path.
// Prevents accidentally mixing file paths with other strings.
type FileID string

// NewFileID creates a FileID from a string path.
func NewFileID(path string) FileID {
	return FileID(path)
}

// String returns the underlying string value.
func (f FileID) String() string {
	return string(f)
}

// Validate checks if the FileID is non-empty.
func (f FileID) Validate() error {
	if f == "" {
		return errors.New("FileID cannot be empty")
	}
	return nil
}

// Validatable is an interface for types that can be validated.
type Validatable interface {
	Validate() error
}

func formatUintValue[T ~uint](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}

func validateUintMinOne[T ~uint](v T, typeName string) error {
	if v == 0 {
		return errors.New(typeName + " must be >= 1, got 0")
	}
	return nil
}

// LineNumber is a branded type representing a line number in a file.
// Uses uint for natural alignment (lines start at 1, not 0).
type LineNumber uint

// NewLineNumber creates a LineNumber from an int.
// Negative values are converted to 0.
func NewLineNumber(n int) LineNumber {
	if n < 0 {
		return 0
	}
	return LineNumber(n)
}

// NewLineNumberFromUint creates a LineNumber from a uint.
func NewLineNumberFromUint(n uint) LineNumber {
	return LineNumber(n)
}

// Int returns the LineNumber as int.
//
//nolint:gosec // G115: Line numbers are always small values that fit in int
func (l LineNumber) Int() int {
	return int(l)
}

// String returns the LineNumber as string.
func (l LineNumber) String() string {
	return formatUintValue(l)
}

// Validate checks if the LineNumber is valid (>= 1).
func (l LineNumber) Validate() error {
	return validateUintMinOne(l, "LineNumber")
}

// BlockIndex is a branded type representing a code block index within a file.
// Uses uint for natural indexing (blocks start at 1 for user display).
type BlockIndex uint

// NewBlockIndex creates a BlockIndex from an int.
// Negative values are converted to 0.
func NewBlockIndex(n int) BlockIndex {
	if n < 0 {
		return 0
	}
	return BlockIndex(n)
}

// NewBlockIndexFromUint creates a BlockIndex from a uint.
func NewBlockIndexFromUint(n uint) BlockIndex {
	return BlockIndex(n)
}

// Int returns the BlockIndex as int.
//
//nolint:gosec // G115: Block indices are always small values that fit in int
func (b BlockIndex) Int() int {
	return int(b)
}

// String returns the BlockIndex as string.
func (b BlockIndex) String() string {
	return formatUintValue(b)
}

// Validate checks if the BlockIndex is valid (>= 1).
func (b BlockIndex) Validate() error {
	return validateUintMinOne(b, "BlockIndex")
}
