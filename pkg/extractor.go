// Package mdgovalidator validates code blocks in Markdown files.
package mdgovalidator

import (
	"slices"
	"strings"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// SkipDirectivesConfig contains markdown directives to skip validation.
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

// ExtractCodeBlocks extracts code blocks for specified languages from markdown content.
func ExtractCodeBlocks(content string, langs []languages.Language) []types.CodeBlock {
	var blocks []types.CodeBlock
	lines := strings.Split(content, "\n")

	state := newExtractorState(langs)
	for i, line := range lines {
		state.processLine(i, line, &blocks)
	}

	return blocks
}

// ExtractGoCodeBlocks extracts Go code blocks from markdown content (backwards compatible).
func ExtractGoCodeBlocks(content string) []types.CodeBlock {
	return ExtractCodeBlocks(content, []languages.Language{languages.LangGo})
}

// ExtractCodeBlocksWithConfig extracts code blocks with custom skip directives.
func ExtractCodeBlocksWithConfig(
	content string,
	langs []languages.Language,
	config SkipDirectivesConfig,
) []types.CodeBlock {
	var blocks []types.CodeBlock
	lines := strings.Split(content, "\n")

	state := newExtractorStateWithConfig(langs, config)
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
	targetLangs    []languages.Language
	currentLang    languages.Language
}

func newExtractorState(langs []languages.Language) *extractorState {
	return &extractorState{
		inCodeBlock:    false,
		currentBlock:   strings.Builder{},
		blockStartLine: 0,
		skipNext:       false,
		skipDirectives: DefaultSkipDirectives(),
		targetLangs:    langs,
		currentLang:    "",
	}
}

func newExtractorStateWithConfig(
	langs []languages.Language,
	config SkipDirectivesConfig,
) *extractorState {
	return &extractorState{
		inCodeBlock:    false,
		currentBlock:   strings.Builder{},
		blockStartLine: 0,
		skipNext:       false,
		skipDirectives: config,
		targetLangs:    langs,
		currentLang:    "",
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
	parsedLang, ok := languages.ParseLanguage(lang)
	if !ok {
		return // Not a supported language
	}

	// Check if this language is in our target list
	if !s.isTargetLanguage(parsedLang) {
		return
	}

	s.inCodeBlock = true
	s.blockStartLine = lineNum + 1 // 1-indexed
	s.currentLang = parsedLang
	s.currentBlock.Reset()
}

func (s *extractorState) isTargetLanguage(lang languages.Language) bool {
	return slices.Contains(s.targetLangs, lang)
}

func (s *extractorState) endCodeBlock(blocks *[]types.CodeBlock) {
	s.inCodeBlock = false
	code := s.currentBlock.String()

	if strings.TrimSpace(code) == "" {
		s.skipNext = false
		s.currentLang = ""
		return
	}

	skipped := s.skipNext || s.hasSkipDirective(code)
	block := types.NewCodeBlock(types.NewLineNumber(s.blockStartLine), s.currentLang, code)
	if skipped {
		block.MarkSkipped()
	} else {
		block.MarkValid()
	}
	if blocks != nil {
		*blocks = append(*blocks, block)
	}
	s.skipNext = false
	s.currentLang = ""
}
