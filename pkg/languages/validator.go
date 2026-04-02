// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"fmt"
)

// ValidationError represents a syntax validation error.
type ValidationError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Line > 0 && e.Column > 0 {
		return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
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
		return fmt.Errorf("validator cannot be nil")
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
		return fmt.Errorf("no validator registered for language: %s", lang)
	}
	if !v.IsAvailable() {
		return fmt.Errorf("validator for %s is not available (required tools not installed)", lang)
	}
	return v.Validate(ctx, code)
}

// DefaultRegistry creates a registry with all built-in validators.
func DefaultRegistry() *Registry {
	r := NewRegistry()

	// Register Go validator (always available, built-in)
	if err := r.Register(&GoValidator{}); err != nil {
		panic(fmt.Sprintf("failed to register Go validator: %v", err))
	}

	// Register tree-sitter based validators (pure Go, always available)
	// These use embedded grammars and don't require external tools
	_ = r.Register(NewTreeSitterValidator(LangRust, "rust"))
	_ = r.Register(NewTreeSitterValidator(LangTypeScript, "typescript"))
	_ = r.Register(NewTreeSitterValidator(LangTSX, "tsx"))
	_ = r.Register(NewTreeSitterValidator(LangNix, "nix"))
	_ = r.Register(NewTreeSitterValidator(LangHCL, "hcl"))
	_ = r.Register(NewTreeSitterValidator(LangTerraform, "terraform"))
	_ = r.Register(NewTreeSitterValidator(LangTempl, "templ"))

	return r
}
