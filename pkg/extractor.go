// Package mdgovalidator validates Go code blocks in Markdown files.
package mdgovalidator

import (
	"strings"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// SkipDirectives contains markdown directives to skip validation.
// This is a configuration list that can be customized before extraction.
type SkipDirectivesConfig []string

// DefaultSkipDirectives returns the standard set of skip directives.
func DefaultSkipDirectives() SkipDirectivesConfig {
	return SkipDirectivesConfig{
		"<!-- skip-validate -->",
		"<!-- skip-md-validate -->",
		"<!-- md-skip -->",
		"<!-- no-validate -->",
		"// skip-validate",
		"//nolint",
	}
}

// ExtractGoCodeBlocks extracts Go code blocks from markdown content.
func ExtractGoCodeBlocks(content string) []types.CodeBlock {
	var blocks []types.CodeBlock
	lines := strings.Split(content, "\n")

	state := newExtractorState()
	for i, line := range lines {
		state.processLine(i, line, &blocks)
	}

	return blocks
}

// ExtractGoCodeBlocksWithConfig extracts Go code blocks with custom skip directives.
func ExtractGoCodeBlocksWithConfig(content string, config SkipDirectivesConfig) []types.CodeBlock {
	var blocks []types.CodeBlock
	lines := strings.Split(content, "\n")

	state := newExtractorStateWithConfig(config)
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
	skipDirectives SkipDirectivesConfig
}

func newExtractorState() *extractorState {
	return &extractorState{
		inCodeBlock:    false,
		currentBlock:   strings.Builder{},
		blockStartLine: 0,
		skipNext:       false,
		skipDirectives: DefaultSkipDirectives(),
	}
}

func newExtractorStateWithConfig(config SkipDirectivesConfig) *extractorState {
	return &extractorState{
		inCodeBlock:    false,
		currentBlock:   strings.Builder{},
		blockStartLine: 0,
		skipNext:       false,
		skipDirectives: config,
	}
}

func (s *extractorState) processLine(lineNum int, line string, blocks *[]types.CodeBlock) {
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
	for _, directive := range s.skipDirectives {
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

func (s *extractorState) endCodeBlock(blocks *[]types.CodeBlock) {
	s.inCodeBlock = false
	code := s.currentBlock.String()

	if strings.TrimSpace(code) == "" {
		s.skipNext = false
		return
	}

	skipped := s.skipNext || s.hasSkipDirective(code)
	block := types.NewCodeBlock(types.NewLineNumber(s.blockStartLine), code)
	if skipped {
		block.MarkSkipped()
	} else {
		block.MarkValid()
	}
	*blocks = append(*blocks, block)
	s.skipNext = false
}
