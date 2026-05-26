package mdgovalidator

import (
	"context"
	"fmt"

	"github.com/larsartmann/md-go-validator/pkg/code"
	"github.com/larsartmann/md-go-validator/pkg/languages"
)

// ValidateGoCode validates Go code using multiple parsing strategies.
// It tries various approaches to handle partial code snippets commonly
// found in documentation.
func ValidateGoCode(goCode string) error {
	err := (&languages.GoValidator{}).Validate(context.Background(), goCode)
	if err != nil {
		return fmt.Errorf("validate go code (code=%q): %w", code.TruncateForError(goCode), err)
	}

	return nil
}
