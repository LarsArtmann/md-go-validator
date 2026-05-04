package mdgovalidator

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
)

func generateMarkdownWithBlocks(blockCount int) string {
	var builder strings.Builder

	for i := range blockCount {
		fmt.Fprintf(&builder, "# Section %d\n\n", i)
		builder.WriteString("```go\n")
		builder.WriteString("package main\n")
		builder.WriteString("```\n\n")
	}

	return builder.String()
}

func BenchmarkExtractCodeBlocks(b *testing.B) {
	content := generateMarkdownWithBlocks(50)

	b.ResetTimer()

	for b.Loop() {
		ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	}
}

func BenchmarkExtractCodeBlocks_LargeFile(b *testing.B) {
	content := generateMarkdownWithBlocks(500)

	b.ResetTimer()

	for b.Loop() {
		ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	}
}

func BenchmarkExtractGoCodeBlocks(b *testing.B) {
	content := generateMarkdownWithBlocks(50)

	b.ResetTimer()

	for b.Loop() {
		ExtractGoCodeBlocks(content)
	}
}

func BenchmarkValidateGoCode(b *testing.B) {
	cases := []struct {
		name string
		code string
	}{
		{"complete_file", "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"},
		{"expression", "fmt.Println(\"hello\")"},
		{"statement", "x := 42"},
	}

	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			for b.Loop() {
				_ = ValidateGoCode(bc.code)
			}
		})
	}
}

func BenchmarkValidateDirectory(b *testing.B) {
	content := generateMarkdownWithBlocks(10)
	tmpDir := b.TempDir()

	for i := range 20 {
		err := os.WriteFile(
			fmt.Sprintf("%s/file%d.md", tmpDir, i),
			[]byte(content),
			0o600,
		)
		if err != nil {
			b.Fatal(err)
		}
	}

	v := New(false)
	ctx := context.Background()

	b.ResetTimer()

	for b.Loop() {
		results, err := v.ValidateDirectory(ctx, tmpDir)
		if err != nil {
			b.Fatal(err)
		}

		_ = results
	}
}

func BenchmarkIsSupportedFile(b *testing.B) {
	paths := []string{"README.md", "guide.markdown", "blog.mdx", "script.go", "style.css"}

	b.ResetTimer()

	for b.Loop() {
		for _, p := range paths {
			IsSupportedFile(p)
		}
	}
}
