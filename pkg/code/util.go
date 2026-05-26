// Package code provides utility functions for code processing.
package code

import (
	"fmt"
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
	if err != nil {
		return fmt.Errorf("parse Go code (code=%q): %w", TruncateForError(code), err)
	}

	return nil
}

// TruncateForError truncates code string for use in error messages.
func TruncateForError(code string) string {
	const maxLen = 50
	if len(code) > maxLen {
		return code[:maxLen] + "..."
	}

	return code
}
