// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"time"
)

// NixValidator validates Nix code using nix-instantiate.
type NixValidator struct {
	external *ExternalValidator
}

// Language returns the language this validator handles.
func (v *NixValidator) Language() Language {
	return LangNix
}

// IsAvailable checks if nix-instantiate is installed.
func (v *NixValidator) IsAvailable() bool {
	return v.external.IsAvailable()
}

// Validate validates Nix code.
func (v *NixValidator) Validate(ctx context.Context, code string) error {
	return v.external.Validate(ctx, code)
}

// NewNixValidator creates a new Nix validator.
func NewNixValidator() *NixValidator {
	return &NixValidator{
		external: NewExternalValidator(
			LangNix,
			"nix-instantiate",
			[]string{"--parse", "--"},
			30*time.Second,
			false, // uses stdin with -- -
			".nix",
			[]string{"nix-instantiate", "--version"},
		),
	}
}
