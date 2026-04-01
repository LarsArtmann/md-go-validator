// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"time"
)

// RustValidator validates Rust code using rustc.
type RustValidator struct {
	external *ExternalValidator
}

// Language returns the language this validator handles.
func (v *RustValidator) Language() Language {
	return LangRust
}

// IsAvailable checks if rustc is installed.
func (v *RustValidator) IsAvailable() bool {
	return v.external.IsAvailable()
}

// Validate validates Rust code.
func (v *RustValidator) Validate(ctx context.Context, code string) error {
	return v.external.Validate(ctx, code)
}

// NewRustValidator creates a new Rust validator.
func NewRustValidator() *RustValidator {
	return &RustValidator{
		external: NewExternalValidator(
			LangRust,
			"rustc",
			[]string{"--emit=metadata", "--crate-type", "lib", "-", "-o", "/dev/null"},
			60*time.Second,
			false, // uses stdin with -
			".rs",
			[]string{"rustc", "--version"},
		),
	}
}
