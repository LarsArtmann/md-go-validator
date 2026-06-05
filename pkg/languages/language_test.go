package languages

import (
	"testing"
)

const (
	testLangGo          = "Go"
	testLangTempl       = "Templ"
	testLangTypeScript  = "TypeScript"
	testLangNix         = "Nix"
	testLangRust        = "Rust"
	testLangHCL         = "HCL"
	testLangUnknown     = "Unknown"
	testInputTempl      = "templ"
	testInputTypeScript = "typescript"
	testInputNix        = "nix"
	testInputRust       = "rust"
	testInputHCL        = "hcl"
)

func TestLanguage_String(t *testing.T) {
	t.Parallel()

	testStringMethod(t, "Language.String", []struct {
		name     string
		value    Language
		expected string
	}{
		{testLangGo, LangGo, "go"},
		{testLangTempl, LangTempl, string(LangTempl)},
		{testLangTypeScript, LangTypeScript, string(LangTypeScript)},
		{testLangNix, LangNix, string(LangNix)},
		{testLangRust, LangRust, string(LangRust)},
		{testLangHCL, LangHCL, string(LangHCL)},
	}, func(l Language) string { return l.String() })
}

func TestParseLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantLang Language
		wantOk   bool
	}{
		{"go lowercase", "go", LangGo, true},
		{"Go uppercase", "Go", LangGo, true},
		{"golang", "golang", LangGo, true},
		{testInputTempl, string(LangTempl), LangTempl, true},
		{testInputTypeScript, string(LangTypeScript), LangTypeScript, true},
		{"ts", string(LangTSX), LangTSX, true},
		{"tsx", string(LangTSX), LangTSX, true},
		{testInputNix, string(LangNix), LangNix, true},
		{testInputRust, string(LangRust), LangRust, true},
		{"rs", string(LangRust), LangRust, true},
		{testInputHCL, string(LangHCL), LangHCL, true},
		{"terraform", string(LangTerraform), LangTerraform, true},
		{"tf", string(LangTerraform), LangTerraform, true},
		{"unknown", "python", "", false},
		{"empty", "", "", false},
		{"whitespace", "  go  ", LangGo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotLang, gotOk := ParseLanguage(tt.input)
			if gotOk != tt.wantOk {
				t.Errorf("ParseLanguage() ok = %v, want %v", gotOk, tt.wantOk)

				return
			}

			if gotOk && gotLang != tt.wantLang {
				t.Errorf("ParseLanguage() lang = %v, want %v", gotLang, tt.wantLang)
			}
		})
	}
}

func TestLanguage_Extensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lang     Language
		expected []string
	}{
		{testLangGo, LangGo, []string{".go"}},
		{testLangTypeScript, LangTypeScript, []string{".ts"}},
		{testLangRust, LangRust, []string{".rs"}},
		{testLangUnknown, Language("unknown"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.lang.Extensions()
			assertExtensionsEqual(t, got, tt.expected)
		})
	}
}

func TestLanguage_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		lang    Language
		wantErr bool
	}{
		{"Go", LangGo, false},
		{"TypeScript", LangTypeScript, false},
		{"Unknown", Language("unknown"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.lang.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAllLanguages(t *testing.T) {
	t.Parallel()

	langs := AllLanguages()
	if len(langs) == 0 {
		t.Error("AllLanguages() returned empty slice")
	}

	for _, lang := range langs {
		err := lang.Validate()
		if err != nil {
			t.Errorf("Language %q failed validation: %v", lang, err)
		}
	}
}

func TestLanguage_IsSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lang     Language
		expected bool
	}{
		{"Go", LangGo, true},
		{"Rust", LangRust, true},
		{"Unknown", Language("python"), false},
		{"Empty", Language(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.lang.IsSupported(); got != tt.expected {
				t.Errorf("IsSupported() = %v, want %v", got, tt.expected)
			}
		})
	}
}
