package mdgovalidator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractGoCodeBlocks(t *testing.T) {
	t.Parallel()

	t.Run("no code blocks", func(t *testing.T) {
		t.Parallel()
		blocks := ExtractGoCodeBlocks("Just text\nNo code here")
		if len(blocks) != 0 {
			t.Errorf("expected 0 blocks, got %d", len(blocks))
		}
	})

	t.Run("single go block", func(t *testing.T) {
		t.Parallel()
		content := "Some text\n```go\nfmt.Println(\"hello\")\n```\nMore text"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if blocks[0].LineNumber != 2 {
			t.Errorf("expected line 2, got %d", blocks[0].LineNumber)
		}
	})

	t.Run("skip other languages", func(t *testing.T) {
		t.Parallel()
		content := "```python\nprint('hello')\n```\n```go\nfmt.Println(\"hello\")\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if blocks[0].LineNumber != 4 {
			t.Errorf("expected line 4, got %d", blocks[0].LineNumber)
		}
	})

	t.Run("skip directive before block", func(t *testing.T) {
		t.Parallel()
		content := "<!-- skip-validate -->\n```go\npartial code\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if !blocks[0].Skipped {
			t.Error("expected block to be skipped")
		}
	})

	t.Run("golang tag", func(t *testing.T) {
		t.Parallel()
		content := "```golang\nfmt.Println(\"hello\")\n```"
		blocks := ExtractGoCodeBlocks(content)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
	})
}

func TestExtractGoCodeBlocks_SkipDirective(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	content := "```go\n\n```"

	blocks := ExtractGoCodeBlocks(content)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty code, got %d", len(blocks))
	}
}

func TestValidateGoCode(t *testing.T) {
	t.Parallel()

	t.Run("complete file", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("package main\n\nfunc main() {}\n"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("type declaration", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("type User struct {\n\tName string\n}"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("function signature", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("func DoSomething() error"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("import statement", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("import \"fmt\""); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("variable declaration", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("var x = 42"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("expression", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("x + y"); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid go.mod syntax", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("require (\n\tgithub.com/pkg v1.0.0\n)"); err == nil {
			t.Error("expected error for go.mod syntax")
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		t.Parallel()
		if err := ValidateGoCode("func broken {"); err == nil {
			t.Error("expected error for invalid syntax")
		}
	})
}

func TestIndentCode(t *testing.T) {
	t.Parallel()
	input := "line1\nline2\n\nline4"
	expected := "\tline1\n\tline2\n\n\tline4\n"

	result := indentCode(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValidator_ValidateFile(t *testing.T) {
	t.Parallel()
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
	if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
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
	t.Parallel()
	v := New(false)
	_, err := v.ValidateFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestValidator_ValidateDirectory(t *testing.T) {
	t.Parallel()
	content := []byte("```go\npackage main\n```\n")

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.md"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), content, 0o600); err != nil {
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
	t.Parallel()

	t.Run("empty results", func(t *testing.T) {
		t.Parallel()
		if HasErrors(nil) {
			t.Error("expected false for nil results")
		}
	})

	t.Run("all valid", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			{File: "", LineNumber: 0, CodeBlock: 0, Code: "", Skipped: false, Error: nil},
			{File: "", LineNumber: 0, CodeBlock: 0, Code: "", Skipped: false, Error: nil},
		}
		if HasErrors(results) {
			t.Error("expected false for valid results")
		}
	})

	t.Run("skipped doesn't count", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			{File: "", LineNumber: 0, CodeBlock: 0, Code: "", Skipped: true, Error: &testError{}},
		}
		if HasErrors(results) {
			t.Error("expected false for skipped error")
		}
	})

	t.Run("has error", func(t *testing.T) {
		t.Parallel()
		results := []Result{
			{File: "", LineNumber: 0, CodeBlock: 0, Code: "", Skipped: false, Error: &testError{}},
		}
		if !HasErrors(results) {
			t.Error("expected true for error result")
		}
	})
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
