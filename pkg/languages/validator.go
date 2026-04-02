// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"errors"
	"fmt"
)

// ErrorCode represents the type of validation error for programmatic handling.
type ErrorCode uint

// Error codes for different validation failure types.
const (
	// ErrCodeUnknown indicates an unspecified error type.
	ErrCodeUnknown ErrorCode = iota
	// ErrCodeSyntax indicates a syntax parsing error.
	ErrCodeSyntax
	// ErrCodeNotAvailable indicates the validator is not available.
	ErrCodeNotAvailable
	// ErrCodeNotRegistered indicates no validator is registered for the language.
	ErrCodeNotRegistered
)

// ValidationError represents a syntax validation error.
type ValidationError struct {
	// Message is the human-readable error description.
	Message string
	// Line is the 1-based line number where the error occurred (0 if unknown).
	Line int
	// Column is the 1-based column number where the error occurred (0 if unknown).
	Column int
	// Code is the error code for programmatic handling.
	Code ErrorCode
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
}

// WithCode returns a new ValidationError with the specified error code.
func (e *ValidationError) WithCode(code ErrorCode) *ValidationError {
	return &ValidationError{
		Message: e.Message,
		Line:    e.Line,
		Column:  e.Column,
		Code:    code,
	}
}

// Unwrap returns the wrapped error if any. Implements errors.Unwrap.
func (e *ValidationError) Unwrap() error {
	return nil
}

// Validator validates code for a specific language.
type Validator interface {
	// Language returns the language this validator handles.
	Language() Language

	// Validate validates the given code.
	// Returns nil if the code is valid, or a ValidationError if invalid.
	Validate(ctx context.Context, code string) error

	// IsAvailable returns true if the validator can run (e.g., external tools are installed).
	IsAvailable() bool
}

// Registry manages validators for different languages.
type Registry struct {
	validators map[Language]Validator
}

// NewRegistry creates a new validator registry.
func NewRegistry() *Registry {
	return &Registry{
		validators: make(map[Language]Validator),
	}
}

// Register adds a validator to the registry.
func (r *Registry) Register(v Validator) error {
	if v == nil {
		return errors.New("validator cannot be nil")
	}

	lang := v.Language()
	if err := lang.Validate(); err != nil {
		return fmt.Errorf("cannot register validator: %w", err)
	}

	r.validators[lang] = v
	return nil
}

// Get returns the validator for a language, or nil if not registered.
func (r *Registry) Get(lang Language) Validator {
	return r.validators[lang]
}

// GetByString looks up a validator by language string (e.g., "go", "typescript").
func (r *Registry) GetByString(lang string) Validator {
	l, ok := ParseLanguage(lang)
	if !ok {
		return nil
	}
	return r.Get(l)
}

// GetAvailable returns all validators that are available (tools installed).
func (r *Registry) GetAvailable() []Validator {
	var available []Validator
	for _, v := range r.validators {
		if v.IsAvailable() {
			available = append(available, v)
		}
	}
	return available
}

// Languages returns all registered languages.
func (r *Registry) Languages() []Language {
	langs := make([]Language, 0, len(r.validators))
	for lang := range r.validators {
		langs = append(langs, lang)
	}
	return langs
}

// Validate validates code for a specific language.
func (r *Registry) Validate(ctx context.Context, lang Language, code string) error {
	v := r.Get(lang)
	if v == nil {
		return &ValidationError{
			Message: fmt.Sprintf("no validator registered for language: %s", lang),
			Code:    ErrCodeNotRegistered,
		}
	}
	if !v.IsAvailable() {
		return &ValidationError{
			Message: fmt.Sprintf("validator for %s is not available (required tools not installed)", lang),
			Code:    ErrCodeNotAvailable,
		}
	}
	if err := v.Validate(ctx, code); err != nil {
		return fmt.Errorf("validation failed for %s: %w", lang, err)
	}
	return nil
}

// DefaultRegistry creates a registry with all built-in validators.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	// Register Go validator (always available, built-in)
	if err := r.Register(&GoValidator{}); err != nil {
		panic(fmt.Sprintf("failed to register Go validator: %v", err))
	}

	// Register tree-sitter based validators using a loop for maintainability.
	// Errors are silently ignored since these are optional validators
	// and may not have grammar support compiled in.
	treeSitterValidators := []struct {
		lang Language
		name string
	}{
		{LangRust, "rust"},
		{LangTypeScript, "typescript"},
		{LangTSX, "tsx"},
		{LangNix, "nix"},
		{LangHCL, "hcl"},
		{LangTerraform, "terraform"},
		{LangTempl, "templ"},
	}

	//nolint:errcheck // Optional validators, failure is acceptable
	for _, v := range treeSitterValidators {
		_ = r.Register(NewTreeSitterValidator(v.lang, v.name))
	}

	return r
}
