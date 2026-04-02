// Package languages provides language detection and validation support.
package languages

import (
	"fmt"
	"slices"
	"strings"
)

// Language represents a supported programming language.
type Language string

// Supported languages.
const (
	LangGo         Language = "go"
	LangTempl      Language = "templ"
	LangTypeScript Language = "typescript"
	LangTSX        Language = "tsx"
	LangNix        Language = "nix"
	LangRust       Language = "rust"
	LangHCL        Language = "hcl"
	LangTerraform  Language = "terraform"
)

// String returns the string representation of the language.
func (l Language) String() string {
	return string(l)
}

// AllLanguages returns all supported languages.
func AllLanguages() []Language {
	return []Language{
		LangGo,
		LangTempl,
		LangTypeScript,
		LangTSX,
		LangNix,
		LangRust,
		LangHCL,
		LangTerraform,
	}
}

// ParseLanguage parses a language identifier from markdown code block info string.
// Returns the language and true if recognized, zero value and false otherwise.
func ParseLanguage(lang string) (Language, bool) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch lang {
	case "go", "golang":
		return LangGo, true
	case "templ":
		return LangTempl, true
	case "ts", "typescript":
		return LangTypeScript, true
	case "tsx":
		return LangTSX, true
	case "nix":
		return LangNix, true
	case "rs", "rust":
		return LangRust, true
	case "hcl":
		return LangHCL, true
	case "tf", "terraform":
		return LangTerraform, true
	default:
		return "", false
	}
}

// IsSupported returns true if the language is supported for validation.
func IsSupported(lang string) bool {
	_, ok := ParseLanguage(lang)
	return ok
}

// Extensions returns common file extensions for the language.
func (l Language) Extensions() []string {
	switch l {
	case LangGo:
		return []string{".go"}
	case LangTempl:
		return []string{".templ"}
	case LangTypeScript:
		return []string{".ts"}
	case LangTSX:
		return []string{".tsx"}
	case LangNix:
		return []string{".nix"}
	case LangRust:
		return []string{".rs"}
	case LangHCL, LangTerraform:
		return []string{".hcl", ".tf", ".tfvars"}
	default:
		return nil
	}
}

// Validate checks if the language identifier is valid.
func (l Language) Validate() error {
	if slices.Contains(AllLanguages(), l) {
		return nil
	}
	return fmt.Errorf("unsupported language: %s", l)
}
