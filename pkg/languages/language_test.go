package languages

import (
	"testing"
)

func TestLanguage_String(t *testing.T) {
	t.Parallel()

	testStringMethod(t, "Language.String", []struct {
		name     string
		value    Language
		expected string
	}{
		{"Go", LangGo, "go"},
		{"Templ", LangTempl, "templ"},
		{"TypeScript", LangTypeScript, "typescript"},
		{"Nix", LangNix, "nix"},
		{"Rust", LangRust, "rust"},
		{"HCL", LangHCL, "hcl"},
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
		{"templ", "templ", LangTempl, true},
		{"typescript", "typescript", LangTypeScript, true},
		{"ts", "ts", LangTypeScript, true},
		{"tsx", "tsx", LangTSX, true},
		{"nix", "nix", LangNix, true},
		{"rust", "rust", LangRust, true},
		{"rs", "rs", LangRust, true},
		{"hcl", "hcl", LangHCL, true},
		{"terraform", "terraform", LangTerraform, true},
		{"tf", "tf", LangTerraform, true},
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

func TestIsSupported(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lang     string
		expected bool
	}{
		{"go", "go", true},
		{"typescript", "typescript", true},
		{"python", "python", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsSupported(tt.lang); got != tt.expected {
				t.Errorf("IsSupported() = %v, want %v", got, tt.expected)
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
		{"Go", LangGo, []string{".go"}},
		{"TypeScript", LangTypeScript, []string{".ts"}},
		{"Rust", LangRust, []string{".rs"}},
		{"Unknown", Language("unknown"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.lang.Extensions()
			if len(got) != len(tt.expected) {
				t.Errorf("Extensions() length = %v, want %v", len(got), len(tt.expected))

				return
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("Extensions()[%d] = %v, want %v", i, got[i], tt.expected[i])
				}
			}
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

	// Check that all languages are valid
	for _, lang := range langs {
		err := lang.Validate()
		if err != nil {
			t.Errorf("Language %q failed validation: %v", lang, err)
		}
	}
}
