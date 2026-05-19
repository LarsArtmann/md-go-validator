package languages

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var errUnsupportedLang = errors.New("unsupported language")

// Language represents a supported programming language.
type Language string

// Supported language constants.
const (
	// LangGo represents the Go programming language.
	LangGo Language = "go"
	// LangTempl represents the Templ templating language.
	LangTempl Language = "templ"
	// LangTypeScript represents the TypeScript programming language.
	LangTypeScript Language = "typescript"
	// LangTSX represents TypeScript with JSX support.
	LangTSX Language = "tsx"
	// LangNix represents the Nix expression language.
	LangNix Language = "nix"
	// LangRust represents the Rust programming language.
	LangRust Language = "rust"
	// LangHCL represents HashiCorp Configuration Language.
	LangHCL Language = "hcl"
	// LangTerraform represents Terraform configuration (alias for HCL).
	LangTerraform Language = "terraform"
)

// Extension constants for file types.
const (
	extGo         = ".go"
	extTempl      = ".templ"
	extTypeScript = ".ts"
	extTSX        = ".tsx"
	extNix        = ".nix"
	extRust       = ".rs"
	extHCL        = ".hcl"
	extTF         = ".tf"
	extTFVars     = ".tfvars"
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

// golangAlias is an alias for the Go language identifier.
const golangAlias = "golang"

// ParseLanguage parses a language identifier from a markdown/MDX code block info string.
// Returns the language and true if recognized, zero value and false otherwise.
func ParseLanguage(lang string) (Language, bool) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch lang {
	case string(LangGo), golangAlias:
		return LangGo, true
	case string(LangTempl):
		return LangTempl, true
	case string(LangTypeScript), "ts":
		return LangTypeScript, true
	case string(LangTSX):
		return LangTSX, true
	case string(LangNix):
		return LangNix, true
	case string(LangRust), "rs":
		return LangRust, true
	case string(LangHCL):
		return LangHCL, true
	case string(LangTerraform), "tf":
		return LangTerraform, true
	default:
		return "", false
	}
}

// Extensions returns common file extensions for the language.
func (l Language) Extensions() []string {
	switch l {
	case LangGo:
		return []string{extGo}
	case LangTempl:
		return []string{extTempl}
	case LangTypeScript:
		return []string{extTypeScript}
	case LangTSX:
		return []string{extTSX}
	case LangNix:
		return []string{extNix}
	case LangRust:
		return []string{extRust}
	case LangHCL, LangTerraform:
		return []string{extHCL, extTF, extTFVars}
	default:
		return nil
	}
}

// Validate checks if the language identifier is valid.
func (l Language) Validate() error {
	if l.IsSupported() {
		return nil
	}

	return fmt.Errorf("%w: %s", errUnsupportedLang, l)
}

// IsSupported returns true if this language is a recognized, supported language.
func (l Language) IsSupported() bool {
	return slices.Contains(AllLanguages(), l)
}
