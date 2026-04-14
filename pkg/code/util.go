// Package code provides utility functions for code processing.
package code

import (
	"go/parser"
	"go/token"
	"strings"
)

// IndentCode indents each non-empty line of code with a tab.
func IndentCode(code string) string {
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

// ParseGo attempts to parse Go code using the standard library parser.
// Returns nil on success, or the parse error on failure.
func ParseGo(fset *token.FileSet, code string) error {
	_, err := parser.ParseFile(fset, "snippet.go", code, parser.AllErrors)
	return err
}
