package types

import "github.com/larsartmann/md-go-validator/pkg/languages"

// CodeBlock represents a code block extracted from a markdown or MDX file.
// This is the internal representation used during extraction.
type CodeBlock struct {
	// LineNumber is the 1-based line number where the code block starts.
	LineNumber LineNumber

	// Language is the programming language of the code block.
	Language languages.Language

	// Code is the actual source code content.
	Code string

	// Skipped indicates if this block should be skipped during validation.
	// Uses explicit Status for clarity over boolean.
	Status ValidationStatus
}

// NewCodeBlock creates a new CodeBlock with default values.
func NewCodeBlock(line LineNumber, lang languages.Language, code string) CodeBlock {
	return CodeBlock{
		LineNumber: line,
		Language:   lang,
		Code:       code,
		Status:     StatusUnknown,
	}
}

// MarkSkipped marks this code block as skipped.
func (b *CodeBlock) MarkSkipped() {
	b.Status = StatusSkipped
}

// MarkValid marks this code block as valid.
func (b *CodeBlock) MarkValid() {
	b.Status = StatusValid
}

// MarkError marks this code block as having an error.
func (b *CodeBlock) MarkError() {
	b.Status = StatusError
}

// IsSkipped returns true if this block should be skipped.
func (b *CodeBlock) IsSkipped() bool {
	return b.Status == StatusSkipped
}

// IsValid returns true if this block passed validation.
func (b *CodeBlock) IsValid() bool {
	return b.Status == StatusValid
}

// HasError returns true if this block has a validation error.
func (b *CodeBlock) HasError() bool {
	return b.Status == StatusError
}
