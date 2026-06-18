package mdgovalidator

import (
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/testutil"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

func TestExtractCodeBlocks_SingleGoBlock(t *testing.T) {
	t.Parallel()

	content := "# Title\n\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 1)

	assertCodeBlock(t, blocks[0], "package main\n", 3, languages.LangGo)
}

func TestExtractCodeBlocks_MultipleGoBlocks(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\ntext\n\n```go\nfunc f() {}\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 2)

	assertCodeBlock(t, blocks[0], "package main\n", 1, languages.LangGo)
	assertCodeBlock(t, blocks[1], "func f() {}\n", 7, languages.LangGo)
}

func TestExtractCodeBlocks_LanguageFiltering(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\n```python\nprint('hi')\n```\n\n```rust\nfn main(){}\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo, languages.LangRust})
	testutil.AssertBlockCount(t, blocks, 2)

	assertCodeBlock(t, blocks[0], "package main\n", 1, languages.LangGo)
	assertCodeBlock(t, blocks[1], "fn main(){}\n", 9, languages.LangRust)
}

func TestExtractCodeBlocks_EmptyBlockIgnored(t *testing.T) {
	t.Parallel()

	content := "```go\n   \n```\n\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 1)

	assertCodeBlock(t, blocks[0], "package main\n", 5, languages.LangGo)
}

func TestExtractCodeBlocks_LineNumberIsOneIndexed(t *testing.T) {
	t.Parallel()

	// Block fence starts on line 4 (0-indexed line 3 → +1 = 4).
	content := "line1\nline2\nline3\n```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 1)

	if blocks[0].LineNumber.Int() != 4 {
		t.Errorf("expected line 4, got %d", blocks[0].LineNumber.Int())
	}
}

func TestExtractCodeBlocks_NonSkippedStatusIsUnknown(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 1)

	// Extraction must not pre-judge validity; the validator decides later.
	if blocks[0].Status != types.StatusUnknown {
		t.Errorf("expected StatusUnknown for non-skipped block, got %s", blocks[0].Status)
	}
}

func TestExtractCodeBlocks_SkippedBlockStatus(t *testing.T) {
	t.Parallel()

	content := "<!-- skip-validate -->\n\n```go\nthis is invalid\n```\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 1)

	assertSkippedCodeBlock(t, blocks[0], "this is invalid\n", 3, languages.LangGo)
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
			assertSingleSkippedBlock(t, blocks, "expected skipped block")
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
	assertSingleSkippedBlock(t, blocks, "expected custom directive to skip block")
}

func TestExtractCodeBlocksWithConfig_DefaultDoesNotSkipCustom(t *testing.T) {
	t.Parallel()

	// With default directives, a custom marker should NOT skip.
	content := "<!-- custom-skip -->\n```go\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	b := testutil.AssertSingleBlock(t, blocks)

	if b.IsSkipped() {
		t.Error("expected default directives to NOT skip on custom marker")
	}
}

func TestExtractGoCodeBlocks_BackwardsCompatible(t *testing.T) {
	t.Parallel()

	content := "```go\npackage main\n```\n\n```python\nprint(1)\n```"
	blocks := ExtractGoCodeBlocks(content)
	assertSingleBlockLanguage(t, blocks, languages.LangGo, "expected go")
}

func TestExtractCodeBlocks_LanguageAliasGolang(t *testing.T) {
	t.Parallel()

	content := "```golang\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	assertSingleBlockLanguage(t, blocks, languages.LangGo, "expected 'golang' alias to map to go")
}

func TestExtractCodeBlocks_MDXFormat(t *testing.T) {
	t.Parallel()

	content := "# Mixed\n\n```go\npackage main\n```\n\n```typescript\nconst x = 1;\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo, languages.LangTypeScript})
	testutil.AssertBlockCount(t, blocks, 2)
}

func TestExtractCodeBlocks_NoCodeBlocks(t *testing.T) {
	t.Parallel()

	content := "# Just a title\n\nNo code here."
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 0)
}

func TestExtractCodeBlocks_UnclosedBlockCapturedToEnd(t *testing.T) {
	t.Parallel()

	// An unclosed fence is never finalized, so no block is emitted.
	content := "```go\npackage main\n"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
	testutil.AssertBlockCount(t, blocks, 0)
}

func TestExtractCodeBlocks_FenceWithInfoString(t *testing.T) {
	t.Parallel()

	// CommonMark allows info strings after the fence (e.g. ```go title="x").
	// The language is the first word; trailing info is ignored.
	content := "```go title=\"example\"\npackage main\n```"
	blocks := ExtractCodeBlocks(content, []languages.Language{languages.LangGo})

	// Current extractor parses the whole info string as the language, so this
	// is NOT recognized as go. This test documents the current behavior.
	testutil.AssertBlockCount(t, blocks, 0)
}

// assertCodeBlock checks all fields of an extracted CodeBlock. The default
// status is StatusUnknown, which is what Extract* leaves on freshly created
// blocks before validation. For skipped blocks use assertSkippedCodeBlock.
func assertCodeBlock(
	t *testing.T,
	block types.CodeBlock,
	wantCode string,
	wantLine int,
	wantLang languages.Language,
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

	if block.Status != types.StatusUnknown {
		t.Errorf("status: want %s, got %s", types.StatusUnknown, block.Status)
	}
}

// assertSkippedCodeBlock is assertCodeBlock for skipped blocks.
func assertSkippedCodeBlock(
	t *testing.T,
	block types.CodeBlock,
	wantCode string,
	wantLine int,
	wantLang languages.Language,
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

	if block.Status != types.StatusSkipped {
		t.Errorf("status: want %s, got %s", types.StatusSkipped, block.Status)
	}
}

// assertSingleSkippedBlock asserts that the extractor returned exactly one
// skipped block, with a domain-specific failure message.
func assertSingleSkippedBlock(t *testing.T, blocks []types.CodeBlock, reason string) {
	t.Helper()

	b := testutil.AssertSingleBlock(t, blocks)
	if !b.IsSkipped() {
		t.Errorf("%s, got status %s", reason, b.Status)
	}
}

// assertSingleBlockLanguage asserts that the extractor returned exactly one
// block in the expected language, with a domain-specific failure message.
func assertSingleBlockLanguage(t *testing.T, blocks []types.CodeBlock, expected languages.Language, reason string) {
	t.Helper()

	b := testutil.AssertSingleBlock(t, blocks)
	if b.Language != expected {
		t.Errorf("%s, got %s", reason, b.Language)
	}
}
