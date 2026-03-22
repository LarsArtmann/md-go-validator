// Package mdgovalidator validates Go code blocks in Markdown files.
package mdgovalidator

import "strings"

// Skip directives that can be placed in markdown to skip validation
var SkipDirectives = []string{
	"<!-- skip-validate -->",
	"<!-- skip-md-validate -->",
	"<!-- md-skip -->",
	"<!-- no-validate -->",
	"// skip-validate",
	"//nolint",
}

// CodeBlock represents a Go code block extracted from markdown
type CodeBlock struct {
	LineNumber int
	Code       string
	Skipped    bool
}

// ExtractGoCodeBlocks extracts Go code blocks from markdown content
func ExtractGoCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock

	lines := splitLines(content)
	inCodeBlock := false
	var currentBlock strings.Builder
	blockStartLine := 0
	skipNext := false

	for i := range len(lines) {
		line := lines[i]
		trimmed := stringsTrimSpace(line)

		// Check for skip directives before code blocks
		for _, directive := range SkipDirectives {
			if stringsContains(line, directive) {
				skipNext = true
				break
			}
		}

		// Check for code block start
		if stringsHasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Check if it's a Go code block
				lang, _ := stringsCutPrefix(trimmed, "```")
				lang = stringsTrimSpace(lang)

				// Support various Go language tags: go, Go, golang
				if lang == "go" || lang == "Go" || lang == "golang" {
					inCodeBlock = true
					blockStartLine = i + 1 // 1-indexed
					currentBlock.Reset()
				}
			} else {
				// End of code block
				inCodeBlock = false
				code := currentBlock.String()
				if stringsTrimSpace(code) != "" {
					// Check if code itself contains skip directives
					skipped := skipNext
					if !skipped {
						for _, directive := range SkipDirectives {
							if stringsContains(code, directive) {
								skipped = true
								break
							}
						}
					}
					blocks = append(blocks, CodeBlock{
						LineNumber: blockStartLine,
						Code:       code,
						Skipped:    skipped,
					})
				}
				skipNext = false
			}
			continue
		}

		if inCodeBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
		}
	}

	return blocks
}

// Import stdlib functions for testability without exposing imports in docs
var (
	splitLines       = stringsSplitLines
	stringsTrimSpace = strings.TrimSpace
	stringsContains  = strings.Contains
	stringsHasPrefix = strings.HasPrefix
	stringsCutPrefix = strings.CutPrefix
)

func stringsSplitLines(s string) []string {
	return strings.Split(s, "\n")
}
