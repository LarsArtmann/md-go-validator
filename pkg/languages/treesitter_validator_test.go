package languages

import (
	"context"
	"testing"
)

func TestTreeSitterValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		language    Language
		langName    string
		validCode   string
		invalidCode string
	}{
		{
			name:        "rust",
			language:    LangRust,
			langName:    "rust",
			validCode:   `fn main() { println!("Hello"); }`,
			invalidCode: `fn main() { println!("Hello`,
		},
		{
			name:        "typescript",
			language:    LangTypeScript,
			langName:    "typescript",
			validCode:   `const x: number = 42;`,
			invalidCode: `const x: number = {`,
		},
		{
			name:        "nix",
			language:    LangNix,
			langName:    "nix",
			validCode:   `{ pkgs ? import <nixpkgs> { } }: pkgs.hello`,
			invalidCode: `{ pkgs ? import <nixpkgs> { }`,
		},
		{
			name:        "hcl",
			language:    LangHCL,
			langName:    "hcl",
			validCode:   `resource "test" "example" { name = "test" }`,
			invalidCode: `resource "test" "example" { name = `,
		},
		{
			name:        "templ",
			language:    LangTempl,
			langName:    "templ",
			validCode:   "package main\ntempl Hello() { <p>Hello</p> }",
			invalidCode: "package main\ntempl Hello() { <p>Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewTreeSitterValidator(tt.language, tt.langName)

			if !validator.IsAvailable() {
				t.Skipf("language %q not available", tt.langName)
			}

			if err := validator.Validate(context.Background(), tt.validCode); err != nil {
				t.Errorf("expected valid code to pass, got error: %v", err)
			}

			if err := validator.Validate(context.Background(), tt.invalidCode); err == nil {
				t.Error("expected invalid code to fail, got no error")
			}
		})
	}
}

func TestTreeSitterValidator_Language(t *testing.T) {
	t.Parallel()

	validator := NewTreeSitterValidator(LangRust, "rust")

	if got := validator.Language(); got != LangRust {
		t.Errorf("expected language %q, got %q", LangRust, got)
	}
}

func TestTreeSitterValidator_IsAvailable(t *testing.T) {
	t.Parallel()

	validator := NewTreeSitterValidator(LangRust, "rust")

	if !validator.IsAvailable() {
		t.Skip("rust grammar not available")
	}
}
