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
	// For code that looks like statements
	indented := codeutil.IndentCode(code)
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

	// All strategies failed - return the original error for reporting
	_, originalErr := parser.ParseFile(token.NewFileSet(), "snippet.go", code, parser.AllErrors)
	if originalErr != nil {
		return fmt.Errorf("operation on %q failed: %w", code, originalErr)
	}
	return nil
}
