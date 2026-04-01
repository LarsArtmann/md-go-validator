// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"time"
)

// TemplValidator validates templ code using the templ CLI.
type TemplValidator struct {
	external *ExternalValidator
}

// Language returns the language this validator handles.
func (v *TemplValidator) Language() Language {
	return LangTempl
}

// IsAvailable checks if the templ CLI is installed.
func (v *TemplValidator) IsAvailable() bool {
	return v.external.IsAvailable()
}

// Validate validates templ code.
func (v *TemplValidator) Validate(ctx context.Context, code string) error {
	return v.external.Validate(ctx, code)
}

// NewTemplValidator creates a new templ validator.
func NewTemplValidator() *TemplValidator {
	return &TemplValidator{
		external: NewExternalValidator(
			LangTempl,
			"templ",
			[]string{"fmt", "-stdin"},
			30*time.Second,
			false, // uses stdin
			".templ",
			[]string{"templ", "version"},
		),
	}
}
