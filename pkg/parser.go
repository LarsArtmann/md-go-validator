package mdgovalidator

import (
	"context"
	"fmt"

	"github.com/larsartmann/md-go-validator/pkg/languages"
)

// ValidateGoCode validates Go code using multiple parsing strategies.
// It tries various approaches to handle partial code snippets commonly
// found in documentation.
func ValidateGoCode(code string) error {
	err := (&languages.GoValidator{}).Validate(context.Background(), code)
	if err != nil {
		return fmt.Errorf("validate go code (code=%q): %w", truncateForError(code), err)
	}

	return nil
}

func truncateForError(code string) string {
	const maxLen = 50
	if len(code) > maxLen {
		return code[:maxLen] + "..."
	}

	return code
}
