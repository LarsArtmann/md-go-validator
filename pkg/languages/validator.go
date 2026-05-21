package languages

import (
	"context"
	"errors"
	"fmt"
)

var errValidatorNil = errors.New("validator cannot be nil")

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
		return errValidatorNil
	}

	lang := v.Language()

	err := lang.Validate()
	if err != nil {
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
		return newValidationError(
			fmt.Sprintf(
				"no validator registered for language: %s (code=%q)",
				lang, truncateCode(code),
			),
			ErrCodeNotRegistered,
		)
	}

	if !v.IsAvailable() {
		return newValidationError(
			fmt.Sprintf(
				"validator for %s is not available (code=%q)",
				lang, truncateCode(code),
			),
			ErrCodeNotAvailable,
		)
	}

	err := v.Validate(ctx, code)
	if err != nil {
		return fmt.Errorf(
			"validation failed for %s (code=%q): %w",
			lang, truncateCode(code), err,
		)
	}

	return nil
}

func truncateCode(code string) string {
	const maxLen = 50
	if len(code) > maxLen {
		return code[:maxLen] + "..."
	}

	return code
}

func newValidationError(message string, code ErrorCode) *ValidationError {
	return &ValidationError{
		Message: message,
		Code:    code,
		Line:    0,
		Column:  0,
	}
}

// DefaultRegistry creates a registry with all built-in validators.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	// Register Go validator (always available, built-in)
	err := r.Register(&GoValidator{})
	if err != nil {
		panic(fmt.Sprintf("failed to register Go validator: %v", err))
	}

	// Register tree-sitter based validators using a loop for maintainability.
	// Errors are silently ignored since these are optional validators
	// and may not have grammar support compiled in.
	treeSitterValidators := []struct {
		lang Language
		name string
	}{
		{LangRust, string(LangRust)},
		{LangTypeScript, string(LangTypeScript)},
		{LangTSX, string(LangTSX)},
		{LangNix, string(LangNix)},
		{LangHCL, string(LangHCL)},
		{LangTerraform, string(LangTerraform)},
		{LangTempl, string(LangTempl)},
	}

	for _, v := range treeSitterValidators {
		_ = r.Register(NewTreeSitterValidator(v.lang, v.name))
	}

	return r
}
