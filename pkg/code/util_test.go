package code

import (
	"go/token"
	"testing"
)

func TestIndentCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "\n",
		},
		{
			name:     "single line",
			input:    "package main",
			expected: "\tpackage main\n",
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: "\tline1\n\tline2\n\tline3\n",
		},
		{
			name:     "trailing newline",
			input:    "line1\n",
			expected: "\tline1\n\n",
		},
		{
			name:     "only whitespace lines",
			input:    "code\n  \nmore",
			expected: "\tcode\n  \n\tmore\n",
		},
		{
			name:     "code block with empty last line",
			input:    "line1\nline2\n",
			expected: "\tline1\n\tline2\n\n",
		},
		{
			name:     "indented code",
			input:    "    indented",
			expected: "\t    indented\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IndentCode(tt.input)
			if result != tt.expected {
				t.Errorf("IndentCode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func parseGoValidCases() []struct {
	name string
	code string
} {
	return []struct {
		name string
		code string
	}{
		{name: "valid package", code: "package main"},
		{name: "valid with imports", code: "package main\n\nimport \"fmt\""},
		{name: "valid function", code: "package main\n\nfunc main() {}"},
		{name: "valid full program", code: `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`},
		{name: "valid type declaration", code: "package main\n\ntype Foo struct {}"},
	}
}

func parseGoInvalidCases() []struct {
	name string
	code string
} {
	return []struct {
		name string
		code string
	}{
		{name: "invalid syntax", code: "package main\n\nfunc {"},
		{name: "invalid - missing paren", code: "package main\n\nfunc main() {"},
		{name: "invalid - bad import", code: "import \"fmt\""},
		{name: "empty string", code: ""},
	}
}

func TestParseGo(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()

	for _, tt := range parseGoValidCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ParseGo(fset, tt.code)
			if err != nil {
				t.Errorf("ParseGo() unexpected error = %v", err)
			}
		})
	}

	for _, tt := range parseGoInvalidCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ParseGo(fset, tt.code)
			if err == nil {
				t.Error("ParseGo() expected error, got nil")
			}
		})
	}
}
