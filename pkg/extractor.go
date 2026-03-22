// Package mdgovalidator validates Go code blocks in Markdown files.
package mdgovalidator

import "strings"

// SkipDirectives contains markdown directives to skip validation.
//
//nolint:gochecknoglobals // Configuration list, not mutable state
var SkipDirectives = []string{
	"<!-- skip-validate -->",
	"<!-- skip-md-validate -->",
	"<!-- md-skip -->",
	"<!-- no-validate -->",
	"// skip-validate",
	"//nolint",
}

// CodeBlock represents a Go code block extracted from markdown.
type CodeBlock struct {
	LineNumber int
	Code       string
	Skipped    bool
}

// ExtractGoCodeBlocks extracts Go code blocks from markdown content.
func ExtractGoCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	state := newExtractorState()
	for i, line := range lines {
		state.processLine(i, line, &blocks)
	}

	return blocks
}

type extractorState struct {
	inCodeBlock    bool
	currentBlock   strings.Builder
	blockStartLine int
	skipNext       bool
}

func newExtractorState() *extractorState {
	return &extractorState{
		inCodeBlock:    false,
		currentBlock:   strings.Builder{},
		blockStartLine: 0,
		skipNext:       false,
	}
}

func (s *extractorState) processLine(lineNum int, line string, blocks *[]CodeBlock) {
	trimmed := strings.TrimSpace(line)

	if s.hasSkipDirective(line) {
		s.skipNext = true
	}

	if !strings.HasPrefix(trimmed, "```") {
		s.handleCodeContent(line)
		return
	}

	if s.inCodeBlock {
		s.endCodeBlock(blocks)
	} else {
		s.startCodeBlock(trimmed, lineNum)
	}
}

func (s *extractorState) hasSkipDirective(line string) bool {
	for _, directive := range SkipDirectives {
		if strings.Contains(line, directive) {
			return true
		}
	}
	return false
}

func (s *extractorState) handleCodeContent(line string) {
	if s.inCodeBlock {
		s.currentBlock.WriteString(line)
		s.currentBlock.WriteString("\n")
	}
}

func (s *extractorState) startCodeBlock(trimmed string, lineNum int) {
	lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
	if isGoLanguage(lang) {
		s.inCodeBlock = true
		s.blockStartLine = lineNum + 1 // 1-indexed
		s.currentBlock.Reset()
	}
}

func isGoLanguage(lang string) bool {
	return lang == "go" || lang == "Go" || lang == "golang"
}

func (s *extractorState) endCodeBlock(blocks *[]CodeBlock) {
	s.inCodeBlock = false
	code := s.currentBlock.String()

	if strings.TrimSpace(code) == "" {
		s.skipNext = false
		return
	}

	skipped := s.skipNext || s.hasSkipDirective(code)
	*blocks = append(*blocks, CodeBlock{
		LineNumber: s.blockStartLine,
		Code:       code,
		Skipped:    skipped,
	})
	s.skipNext = false
}
