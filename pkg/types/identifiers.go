package types

import (
	"fmt"
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
		return fmt.Errorf("FileID cannot be empty")
	}
	return nil
}

// LineNumber is a branded type representing a line number in a file.
// Uses uint for natural alignment (lines start at 1, not 0).
type LineNumber uint

// NewLineNumber creates a LineNumber from an int.
func NewLineNumber(n int) LineNumber {
	return LineNumber(n)
}

// NewLineNumberFromUint creates a LineNumber from a uint.
func NewLineNumberFromUint(n uint) LineNumber {
	return LineNumber(n)
}

// Int returns the LineNumber as int.
func (l LineNumber) Int() int {
	return int(l)
}

// String returns the LineNumber as string.
func (l LineNumber) String() string {
	return strconv.FormatUint(uint64(l), 10)
}

// Validate checks if the LineNumber is valid (>= 1).
func (l LineNumber) Validate() error {
	if l == 0 {
		return fmt.Errorf("LineNumber must be >= 1, got 0")
	}
	return nil
}

// BlockIndex is a branded type representing a code block index within a file.
// Uses uint for natural indexing (blocks start at 1 for user display).
type BlockIndex uint

// NewBlockIndex creates a BlockIndex from an int.
func NewBlockIndex(n int) BlockIndex {
	return BlockIndex(n)
}

// NewBlockIndexFromUint creates a BlockIndex from a uint.
func NewBlockIndexFromUint(n uint) BlockIndex {
	return BlockIndex(n)
}

// Int returns the BlockIndex as int.
func (b BlockIndex) Int() int {
	return int(b)
}

// String returns the BlockIndex as string.
func (b BlockIndex) String() string {
	return strconv.FormatUint(uint64(b), 10)
}

// Validate checks if the BlockIndex is valid (>= 1).
func (b BlockIndex) Validate() error {
	if b == 0 {
		return fmt.Errorf("BlockIndex must be >= 1, got 0")
	}
	return nil
}
