package mdgovalidator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractGoCodeBlocks(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedBlocks int
		expectedLines  []int
	}{
		{
			name:           "no code blocks",
			content:        "Just text\nNo code here",
			expectedBlocks: 0,
		},
		{
			name:           "single go block",
			content:        "Some text\n```go\nfmt.Println(\"hello\")\n```\nMore text",
			expectedBlocks: 1,
			expectedLines:  []int{2},
		},
		{
			name: "multiple go blocks",
			content: `# Title

` + "```go" + `
code block 1
` + "```" + `

Text between

` + "```go" + `
code block 2
` + "```",
			expectedBlocks: 2,
			expectedLines:  []int{3, 9},
		},
		{
			name:           "skip other languages",
			content:        "```python\nprint('hello')\n```\n```go\nfmt.Println(\"hello\")\n```",
			expectedBlocks: 1,
			expectedLines:  []int{4},
		},
		{
			name:           "skip directive before block",
			content:        "<!-- skip-validate -->\n```go\npartial code\n```",
			expectedBlocks: 1,
			expectedLines:  []int{2},
		},
		{
			name:           "skip directive in code",
			content:        "```go\n// skip-validate\npartial code\n```",
			expectedBlocks: 1,
			expectedLines:  []int{1},
		},
		{
			name:           "golang tag",
			content:        "```golang\nfmt.Println(\"hello\")\n```",
			expectedBlocks: 1,
			expectedLines:  []int{1},
		},
		{
			name:           "Go tag (capitalized)",
			content:        "```Go\nfmt.Println(\"hello\")\n```",
			expectedBlocks: 1,
			expectedLines:  []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := ExtractGoCodeBlocks(tt.content)
			if len(blocks) != tt.expectedBlocks {
				t.Errorf("expected %d blocks, got %d", tt.expectedBlocks, len(blocks))
			}
			for i, line := range tt.expectedLines {
				if i < len(blocks) && blocks[i].LineNumber != line {
					t.Errorf("block %d: expected line %d, got %d", i, line, blocks[i].LineNumber)
				}
			}
		})
	}
}

func TestExtractGoCodeBlocks_SkipDirective(t *testing.T) {
	content := `<!-- skip-validate -->
` + "```go" + `
type Partial struct {
    Name string
}
` + "```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !blocks[0].Skipped {
		t.Error("expected block to be skipped")
	}
}

func TestExtractGoCodeBlocks_SkipInCode(t *testing.T) {
	content := "```go\n//nolint\ntype Partial struct{}\n```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if !blocks[0].Skipped {
		t.Error("expected block to be skipped due to //nolint in code")
	}
}

func TestExtractGoCodeBlocks_EmptyBlock(t *testing.T) {
	content := "```go\n\n```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty code, got %d", len(blocks))
	}
}

func TestValidateGoCode(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		expectErr bool
	}{
		{
			name:      "complete file",
			code:      "package main\n\nfunc main() {}\n",
			expectErr: false,
		},
		{
			name:      "type declaration",
			code:      "type User struct {\n\tName string\n}",
			expectErr: false,
		},
		{
			name:      "function signature",
			code:      "func DoSomething() error",
			expectErr: false,
		},
		{
			name:      "import statement",
			code:      "import \"fmt\"",
			expectErr: false,
		},
		{
			name:      "variable declaration",
			code:      "var x = 42",
			expectErr: false,
		},
		{
			name:      "expression",
			code:      "x + y",
			expectErr: false,
		},
		{
			name:      "statements",
			code:      "result, err := doSomething()\nif err != nil {\n\treturn err\n}",
			expectErr: false,
		},
		{
			name:      "invalid go.mod syntax",
			code:      "require (\n\tgithub.com/pkg v1.0.0\n)",
			expectErr: true,
		},
		{
			name:      "invalid syntax",
			code:      "func broken {",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoCode(tt.code)
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestIndentCode(t *testing.T) {
	input := "line1\nline2\n\nline4"
	expected := "\tline1\n\tline2\n\n\tline4\n"

	result := indentCode(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValidator_ValidateFile(t *testing.T) {
	content := []byte(`# Test

` + "```go" + `
package main

func main() {
    fmt.Println("hello")
}
` + "```" + `
`)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	results, err := v.ValidateFile(tmpFile)
	if err != nil {
		t.Fatalf("ValidateFile error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestValidator_ValidateFile_NonExistent(t *testing.T) {
	v := New(false)
	_, err := v.ValidateFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidator_ValidateDirectory(t *testing.T) {
	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(false)
	results, err := v.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ValidateDirectory error: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result (only .md files), got %d", len(results))
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		results  []Result
		expected bool
	}{
		{
			name:     "empty results",
			results:  []Result{},
			expected: false,
		},
		{
			name: "all valid",
			results: []Result{
				{Error: nil},
				{Error: nil},
			},
			expected: false,
		},
		{
			name: "skipped doesn't count",
			results: []Result{
				{Skipped: true, Error: &testError{}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasErrors(tt.results); got != tt.expected {
				t.Errorf("HasErrors() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
