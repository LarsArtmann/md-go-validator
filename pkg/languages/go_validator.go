// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"errors"
	"go/parser"
	"go/scanner"
	"go/token"

	codeutil "github.com/larsartmann/md-go-validator/pkg/code"
)

// GoValidator validates Go code using the standard library parser.
type GoValidator struct{}

// Language returns the language this validator handles.
func (v *GoValidator) Language() Language {
	return LangGo
}

// IsAvailable always returns true for Go (stdlib is always available).
func (v *GoValidator) IsAvailable() bool {
	return true
}

// Validate validates Go code using multiple parsing strategies.
func (v *GoValidator) Validate(_ context.Context, code string) error {
	fset := token.NewFileSet()

	// Strategy 1: Try parsing as a complete file
	if v.tryParse(fset, code) == nil {
		return nil
	}

	// Strategy 2: Try wrapping in a package main declaration
	wrapped := "package main\n\n" + code
	if v.tryParse(fset, wrapped) == nil {
		return nil
	}

	// Strategy 3: Try wrapping in package main with func main
	indented := codeutil.IndentCode(code)
	wrappedFunc := "package main\n\nfunc main() {\n" + indented + "\n}"
	if v.tryParse(fset, wrappedFunc) == nil {
		return nil
	}

	// Strategy 4: Try as expression in a function
	exprCode := "package main\n\nfunc _() {\n_ = " + code + "\n}"
	if v.tryParse(fset, exprCode) == nil {
		return nil
	}

	// Strategy 5: Try as multiple statements
	stmtCode := "package main\n\nfunc _() {\n" + indented + "\n}"
	if v.tryParse(fset, stmtCode) == nil {
		return nil
	}

	// All strategies failed - extract error with position information
	return v.createValidationError(fset, code)
}

// tryParse attempts to parse code and returns nil on success, or the error on failure.
func (v *GoValidator) tryParse(fset *token.FileSet, code string) error {
	_, err := parser.ParseFile(fset, "snippet.go", code, parser.AllErrors)
	return err
}

// createValidationError extracts line/column from Go parser errors.
func (v *GoValidator) createValidationError(fset *token.FileSet, code string) error {
	_, err := parser.ParseFile(fset, "snippet.go", code, parser.AllErrors)
	if err == nil {
		return nil
	}

	var line, column int
	var message string

	// Try to extract position information from scanner.ErrorList
	var errList scanner.ErrorList
	if errors.As(err, &errList) && len(errList) > 0 {
		firstErr := errList[0]
		message = firstErr.Msg
		if firstErr.Pos.IsValid() {
			line = firstErr.Pos.Line
			column = firstErr.Pos.Column
		}
	} else {
		// Fallback: use error string without position
		message = err.Error()
	}

	return (&ValidationError{
		Message: "Go syntax error: " + message,
		Line:    line,
		Column:  column,
		Code:    ErrCodeUnknown,
	}).WithCode(ErrCodeSyntax)
}
