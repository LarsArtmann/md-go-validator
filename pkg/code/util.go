// Package code provides utility functions for code processing.
package code

import "strings"

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
