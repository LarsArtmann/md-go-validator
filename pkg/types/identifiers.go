package types

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	errFileIDEmpty         = errors.New("FileID cannot be empty")
	errUintMinOne          = errors.New("value must be >= 1, got 0")
	errUnsupportedFileType = errors.New("unsupported file type")
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
		return errFileIDEmpty
	}

	return nil
}

// Validatable is an interface for types that can be validated.
//
//nolint:iface // Mirrors positiveUintValidator constraint in test helpers for type validation
type Validatable interface {
	Validate() error
}

func formatUintValue[T ~uint](v T) string {
	return strconv.FormatUint(uint64(v), 10)
}

func validateUintMinOne[T ~uint](v T, typeName string) error {
	if v == 0 {
		return fmt.Errorf("%w: %s", errUintMinOne, typeName)
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

// FileType is a branded type representing a supported file type.
// Prevents accidentally mixing file extensions with other strings.
//
//nolint:recvcheck // UnmarshalText must use pointer receiver to mutate
type FileType string

// Supported file type constants.
const (
	// FileTypeMarkdown represents standard Markdown files (.md).
	FileTypeMarkdown FileType = ".md"
	// FileTypeMarkdownAlt represents alternative Markdown extension (.markdown).
	FileTypeMarkdownAlt FileType = ".markdown"
	// FileTypeMdx represents MDX files (.mdx).
	FileTypeMdx FileType = ".mdx"
)

// String returns the underlying string value.
func (f FileType) String() string {
	return string(f)
}

// IsSupported returns true if this FileType is a recognized file type.
func (f FileType) IsSupported() bool {
	switch f {
	case FileTypeMarkdown, FileTypeMarkdownAlt, FileTypeMdx:
		return true
	default:
		return false
	}
}

// AllFileTypes returns all supported file types.
func AllFileTypes() []FileType {
	return []FileType{FileTypeMarkdown, FileTypeMarkdownAlt, FileTypeMdx}
}

// ParseFileType parses a string into a FileType.
// Returns the FileType and true if recognized, zero value and false otherwise.
func ParseFileType(s string) (FileType, bool) {
	ft := FileType(s)

	return ft, ft.IsSupported()
}

// UnmarshalText implements encoding.TextUnmarshaler for deserialization.
func (f *FileType) UnmarshalText(text []byte) error {
	parsed, ok := ParseFileType(string(text))
	if !ok {
		return fmt.Errorf("%w: %s", errUnsupportedFileType, string(text))
	}

	*f = parsed

	return nil
}

// MarshalText implements encoding.TextMarshaler for serialization.
func (f FileType) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
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
