// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"time"
)

// TypeScriptValidator validates TypeScript code using tsc.
type TypeScriptValidator struct {
	external *ExternalValidator
}

// Language returns the language this validator handles.
func (v *TypeScriptValidator) Language() Language {
	return LangTypeScript
}

// IsAvailable checks if the TypeScript compiler is installed.
func (v *TypeScriptValidator) IsAvailable() bool {
	return v.external.IsAvailable()
}

// Validate validates TypeScript code.
func (v *TypeScriptValidator) Validate(ctx context.Context, code string) error {
	return v.external.Validate(ctx, code)
}

// NewTypeScriptValidator creates a new TypeScript validator.
func NewTypeScriptValidator() *TypeScriptValidator {
	return &TypeScriptValidator{
		external: NewExternalValidator(
			LangTypeScript,
			"tsc",
			[]string{"--noEmit", "--allowJs", "--skipLibCheck", "--strict", "{{FILE}}"},
			60*time.Second,
			true, // needs file
			".ts",
			[]string{"tsc", "--version"},
		),
	}
}
