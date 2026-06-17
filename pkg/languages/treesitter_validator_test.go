package languages

import (
	"context"
	"testing"
)

func TestTreeSitterValidator(t *testing.T) {
	t.Parallel()

	for _, tt := range treeSitterValidatorTests() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewTreeSitterValidator(tt.language)

			if !validator.IsAvailable() {
				t.Skipf("language %q not available", tt.language)
			}

			err := validator.Validate(context.Background(), tt.validCode)
			if err != nil {
				t.Errorf("expected valid code to pass, got error: %v", err)
			}

			err = validator.Validate(context.Background(), tt.invalidCode)
			if err == nil {
				t.Error("expected invalid code to fail, got no error")
			}
		})
	}
}

type validatorTestCase struct {
	name        string
	language    Language
	validCode   string
	invalidCode string
}

func treeSitterValidatorTests() []validatorTestCase {
	return []validatorTestCase{
		{
			name:        "rust",
			language:    LangRust,
			validCode:   `fn main() { println!("Hello"); }`,
			invalidCode: `fn main() { println!("Hello`,
		},
		{
			name:        "typescript",
			language:    LangTypeScript,
			validCode:   `const x: number = 42;`,
			invalidCode: `const x: number = {`,
		},
		{
			name:        "nix",
			language:    LangNix,
			validCode:   `{ pkgs ? import <nixpkgs> { } }: pkgs.hello`,
			invalidCode: `{ pkgs ? import <nixpkgs> { }`,
		},
		{
			name:        "hcl",
			language:    LangHCL,
			validCode:   `resource "test" "example" { name = "test" }`,
			invalidCode: `resource "test" "example" { name = `,
		},
		{
			name:        "templ",
			language:    LangTempl,
			validCode:   "package main\ntempl Hello() { <p>Hello</p> }",
			invalidCode: "package main\ntempl Hello() { <p>Hello",
		},
	}
}

func TestTreeSitterValidator_Language(t *testing.T) {
	t.Parallel()

	validator := NewTreeSitterValidator(LangRust)

	if got := validator.Language(); got != LangRust {
		t.Errorf("expected language %q, got %q", LangRust, got)
	}
}

func TestTreeSitterValidator_IsAvailable(t *testing.T) {
	t.Parallel()

	validator := NewTreeSitterValidator(LangRust)

	if !validator.IsAvailable() {
		t.Skip("rust grammar not available")
	}
}
