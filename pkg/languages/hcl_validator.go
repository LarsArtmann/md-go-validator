// Package languages provides language detection and validation support.
package languages

import (
	"context"
	"time"
)

// HCLValidator validates HCL/Terraform code using terraform or hclfmt.
type HCLValidator struct {
	external *ExternalValidator
}

// Language returns the language this validator handles.
func (v *HCLValidator) Language() Language {
	return LangHCL
}

// IsAvailable checks if terraform or hclfmt is installed.
func (v *HCLValidator) IsAvailable() bool {
	return v.external.IsAvailable()
}

// Validate validates HCL/Terraform code.
func (v *HCLValidator) Validate(ctx context.Context, code string) error {
	return v.external.Validate(ctx, code)
}

// NewHCLValidator creates a new HCL validator.
func NewHCLValidator() *HCLValidator {
	return &HCLValidator{
		external: NewExternalValidator(
			LangHCL,
			"terraform",
			[]string{"fmt", "-check=true", "-"},
			30*time.Second,
			false, // uses stdin
			".hcl",
			[]string{"terraform", "version"},
		),
	}
}
