// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"errors"
	"go/parser"
	"go/scanner"
	"go/token"
	"strings"
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
	// Strategy 1: Try parsing as a complete file
	_, err := parser.ParseFile(token.NewFileSet(), "snippet.go", code, parser.AllErrors)
	if err == nil {
		return nil
	}

	// Strategy 2: Try wrapping in a package main declaration
	wrapped := "package main\n\n" + code
	_, err = parser.ParseFile(token.NewFileSet(), "snippet.go", wrapped, parser.AllErrors)
	if err == nil {
		return nil
	}

	// Strategy 3: Try wrapping in package main with func main
	indented := indentCode(code)
	wrappedFunc := "package main\n\nfunc main() {\n" + indented + "\n}"
	_, err = parser.ParseFile(token.NewFileSet(), "snippet.go", wrappedFunc, parser.AllErrors)
	if err == nil {
		return nil
	}

	// Strategy 4: Try as expression in a function
	exprCode := "package main\n\nfunc _() {\n_ = " + code + "\n}"
	_, err = parser.ParseFile(token.NewFileSet(), "snippet.go", exprCode, parser.AllErrors)
	if err == nil {
		return nil
	}

	// Strategy 5: Try as multiple statements
	stmtCode := "package main\n\nfunc _() {\n" + indented + "\n}"
	_, err = parser.ParseFile(token.NewFileSet(), "snippet.go", stmtCode, parser.AllErrors)
	if err == nil {
		return nil
	}

	// All strategies failed - extract error with position information
	return v.createValidationError(code)
}

// createValidationError extracts line/column from Go parser errors.
func (v *GoValidator) createValidationError(code string) error {
	fset := token.NewFileSet()
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
			pos := fset.Position(firstErr.Pos)
			line = pos.Line
			column = pos.Column
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

// indentCode indents each non-empty line of code with a tab.
func indentCode(code string) string {
	lines := strings.Split(code, "\n")
	var result strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result.WriteString("\t")
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}
