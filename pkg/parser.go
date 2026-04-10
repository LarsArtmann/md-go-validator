package mdgovalidator

import (
	"fmt"
	"go/parser"
	"go/token"

	codeutil "github.com/larsartmann/md-go-validator/pkg/code"
)

// ValidateGoCode validates Go code using multiple parsing strategies.
// It tries various approaches to handle partial code snippets commonly
// found in documentation.
func ValidateGoCode(code string) error {
	fset := token.NewFileSet()

	// Strategy 1: Try parsing as a complete file
	if tryParseGo(fset, code) == nil {
		return nil
	}

	// Strategy 2: Try wrapping in a package main declaration
	wrapped := "package main\n\n" + code
	if tryParseGo(fset, wrapped) == nil {
		return nil
	}

	// Strategy 3: Try wrapping in package main with func main
	// For code that looks like statements
	indented := codeutil.IndentCode(code)
	wrappedFunc := "package main\n\nfunc main() {\n" + indented + "\n}"
	if tryParseGo(fset, wrappedFunc) == nil {
		return nil
	}

	// Strategy 4: Try as expression in a function
	exprCode := "package main\n\nfunc _() {\n_ = " + code + "\n}"
	if tryParseGo(fset, exprCode) == nil {
		return nil
	}

	// Strategy 5: Try as multiple statements
	stmtCode := "package main\n\nfunc _() {\n" + indented + "\n}"
	if tryParseGo(fset, stmtCode) == nil {
		return nil
	}

	// All strategies failed - return the original error for reporting
	_, originalErr := parser.ParseFile(fset, "snippet.go", code, parser.AllErrors)
	if originalErr != nil {
		return fmt.Errorf("operation on %q failed: %w", code, originalErr)
	}
	return nil
}

// tryParseGo attempts to parse Go code and returns nil on success, or the error on failure.
func tryParseGo(fset *token.FileSet, code string) error {
	_, err := parser.ParseFile(fset, "snippet.go", code, parser.AllErrors)
	return err
}
