package mdgovalidator

import (
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

func TestExtractCodeBlocks_SingleGoBlock(t *testing.T) {
	t.Parallel()

	content := "# Title\n\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	assertCodeBlock(t, blocks[0], "package main\n", 3, languages.LangGo, types.StatusUnknown)
}

func TestExtractCodeBlocks_MultipleGoBlocks(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\ntext\n\n```go\nfunc f() {}\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}

	assertCodeBlock(t, blocks[0], "package main\n", 1, languages.LangGo, types.StatusUnknown)
	assertCodeBlock(t, blocks[1], "func f() {}\n", 7, languages.LangGo, types.StatusUnknown)
}

func TestExtractCodeBlocks_LanguageFiltering(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\n```python\nprint('hi')\n```\n\n```rust\nfn main(){}\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo, languages.LangRust})

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks (go+rust, python filtered), got %d", len(blocks))
	}

	assertCodeBlock(t, blocks[0], "package main\n", 1, languages.LangGo, types.StatusUnknown)
	assertCodeBlock(t, blocks[1], "fn main(){}\n", 9, languages.LangRust, types.StatusUnknown)
}

func TestExtractCodeBlocks_EmptyBlockIgnored(t *testing.T) {
	t.Parallel()

	content := "```go\n   \n```\n\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block (empty ignored), got %d", len(blocks))
	}

	assertCodeBlock(t, blocks[0], "package main\n", 5, languages.LangGo, types.StatusUnknown)
}

func TestExtractCodeBlocks_LineNumberIsOneIndexed(t *testing.T) {
	t.Parallel()

	// Block fence starts on line 4 (0-indexed line 3 → +1 = 4).
	content := "line1\nline2\nline3\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].LineNumber.Int() != 4 {
		t.Errorf("expected line 4, got %d", blocks[0].LineNumber.Int())
	}
}

func TestExtractCodeBlocks_NonSkippedStatusIsUnknown(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	// Extraction must not pre-judge validity; the validator decides later.
	if blocks[0].Status != types.StatusUnknown {
		t.Errorf("expected StatusUnknown for non-skipped block, got %s", blocks[0].Status)
	}
}

func TestExtractCodeBlocks_SkippedBlockStatus(t *testing.T) {
	t.Parallel()

	content := "<!-- skip-validate -->\n\n```go\nthis is invalid\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	assertCodeBlock(t, blocks[0], "this is invalid\n", 3, languages.LangGo, types.StatusSkipped)
}

func TestExtractCodeBlocks_DefaultSkipDirectives(t *testing.T) {
	t.Parallel()

	directives := DefaultSkipDirectives()

	expected := map[string]bool{
		"<!-- skip-validate -->":    true,
		"<!-- skip-md-validate -->": true,
		"<!-- md-skip -->":          true,
		"<!-- no-validate -->":      true,
		"// skip-validate":          true,
		"//nolint":                  true,
	}

	if len(directives) != len(expected) {
		t.Fatalf("expected %d directives, got %d", len(expected), len(directives))
	}

	for _, d := range directives {
		if !expected[d] {
			t.Errorf("unexpected directive %q", d)
		}
	}
}

func TestExtractCodeBlocks_AllDefaultDirectivesSkipBlock(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "html skip-validate before block",
			content: "<!-- skip-validate -->\n```go\nx\n```",
		},
		{
			name:    "nolint inside block",
			content: "```go\n//nolint\nx\n```",
		},
		{
			name:    "md-skip before block",
			content: "<!-- md-skip -->\n```go\nx\n```",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			blocks := ExtractCodeBlocks(tc.content, []languages.Language{languages.LangGo})
			if len(blocks) != 1 {
				t.Fatalf("expected 1 block, got %d", len(blocks))
			}

			if !blocks[0].IsSkipped() {
				t.Errorf("expected skipped block, got status %s", blocks[0].Status)
			}
		})
	}
}

func TestExtractCodeBlocksWithConfig_CustomDirectives(t *testing.T) {
	t.Parallel()

	// Custom directive that is NOT in the default set.
	content := "<!-- custom-skip -->\n```go\nx\n```"
	blocks := ExtractCodeBlocksWithConfig(
		content,
		[]languages.Language{languages.LangGo},
		SkipDirectivesConfig{"<!-- custom-skip -->"},
	)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if !blocks[0].IsSkipped() {
		t.Errorf("expected custom directive to skip block, got status %s", blocks[0].Status)
	}
}

func TestExtractCodeBlocksWithConfig_DefaultDoesNotSkipCustom(t *testing.T) {
	t.Parallel()

	// With default directives, a custom marker should NOT skip.
	content := "<!-- custom-skip -->\n```go\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}

	if blocks[0].IsSkipped() {
		t.Error("expected default directives to NOT skip on custom marker")
	}
}

func TestExtractGoCodeBlocks_BackwardsCompatible(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\n```python\nprint(1)\n```"
	blocks := ExtractGoCodeBlocks(content)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 Go block, got %d", len(blocks))
	}

	if blocks[0].Language != languages.LangGo {
		t.Errorf("expected go, got %s", blocks[0].Language)
	}
}

func TestExtractCodeBlocks_LanguageAliasGolang(t *testing.T) {
	t.Parallel()

	content := "```golang\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 1 {
		t.Fatalf("expected 'golang' alias to map to go, got %d blocks", len(blocks))
	}

	if blocks[0].Language != languages.LangGo {
		t.Errorf("expected go, got %s", blocks[0].Language)
	}
}

func TestExtractCodeBlocks_MDXFormat(t *testing.T) {
	t.Parallel()

	content := "# Mixed\n\n```go\npackage main\n```\n\n```typescript\nconst x = 1;\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo, languages.LangTypeScript})

	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks from MDX, got %d", len(blocks))
	}
}

func TestExtractCodeBlocks_NoCodeBlocks(t *testing.T) {
	t.Parallel()

	content := "# Just a title\n\nNo code here."
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks, got %d", len(blocks))
	}
}

func TestExtractCodeBlocks_UnclosedBlockCapturedToEnd(t *testing.T) {
	t.Parallel()

	// An unclosed fence is never finalized, so no block is emitted.
	content := "```go\npackage main\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks for unclosed fence, got %d", len(blocks))
	}
}

func TestExtractCodeBlocks_FenceWithInfoString(t *testing.T) {
	t.Parallel()

	// CommonMark allows info strings after the fence (e.g. ```go title="x").
	// The language is the first word; trailing info is ignored.
	content := "```go title=\"example\"\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	// Current extractor parses the whole info string as the language, so this
	// is NOT recognized as go. This test documents the current behavior.
	if len(blocks) != 0 {
		t.Fatalf("expected 0 blocks (info string not stripped), got %d", len(blocks))
	}
}

// assertCodeBlock checks all fields of an extracted CodeBlock.
func assertCodeBlock(
	t *testing.T,
	block types.CodeBlock,
	wantCode string,
	wantLine int,
	wantLang languages.Language,
	wantStatus types.ValidationStatus,
) {
	t.Helper()

	if block.Code != wantCode {
		t.Errorf("code: want %q, got %q", wantCode, block.Code)
	}

	if block.LineNumber.Int() != wantLine {
		t.Errorf("line: want %d, got %d", wantLine, block.LineNumber.Int())
	}

	if block.Language != wantLang {
		t.Errorf("language: want %s, got %s", wantLang, block.Language)
	}

	if block.Status != wantStatus {
		t.Errorf("status: want %s, got %s", wantStatus, block.Status)
	}
}
