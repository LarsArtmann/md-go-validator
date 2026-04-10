// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"fmt"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// TreeSitterValidator validates code using pure Go tree-sitter parsing.
type TreeSitterValidator struct {
	language Language
	langName string
}

// Language returns the language this validator handles.
func (v *TreeSitterValidator) Language() Language {
	return v.language
}

// IsAvailable always returns true for tree-sitter validators (embedded grammars).
func (v *TreeSitterValidator) IsAvailable() bool {
	entry := grammars.DetectLanguageByName(v.langName)
	return entry != nil && entry.Language != nil
}

// Validate validates code using tree-sitter parsing.
// Returns nil if the code parses without errors, or a ValidationError if invalid.
func (v *TreeSitterValidator) Validate(_ context.Context, code string) error {
	entry := grammars.DetectLanguageByName(v.langName)
	if entry == nil || entry.Language == nil {
		return newValidationError(
			fmt.Sprintf("language %q not available", v.langName),
			ErrCodeNotAvailable,
		)
	}

	lang := entry.Language()
	if lang == nil {
		return newValidationError(
			fmt.Sprintf("failed to load language %q", v.langName),
			ErrCodeNotAvailable,
		)
	}

	parser := gotreesitter.NewParser(lang)

	tree, err := parser.Parse([]byte(code))
	if err != nil {
		return newValidationError(
			fmt.Sprintf("failed to parse %s code: %v", v.langName, err),
			ErrCodeSyntax,
		)
	}
	defer tree.Release()

	root := tree.RootNode()
	if root == nil {
		return newValidationError(
			fmt.Sprintf("failed to get root node for %s code", v.langName),
			ErrCodeSyntax,
		)
	}

	if root.HasError() {
		return newValidationError(
			v.langName+" syntax error: code contains parse errors",
			ErrCodeSyntax,
		)
	}

	return nil
}

// NewTreeSitterValidator creates a new tree-sitter based validator.
func NewTreeSitterValidator(language Language, langName string) *TreeSitterValidator {
	return &TreeSitterValidator{
		language: language,
		langName: langName,
	}
}
