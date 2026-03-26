package mdgovalidator

import (
	"context"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// Validator defines the interface for validating Go code blocks in markdown files.
// This interface enables dependency injection and easier testing.
type Validator interface {
	// ValidateFile validates a single markdown file and returns validation results.
	// The context can be used for cancellation and timeout.
	ValidateFile(ctx context.Context, filePath string) ([]types.Result, error)

	// ValidateDirectory validates all markdown files in a directory recursively.
	// The context can be used for cancellation and timeout.
	ValidateDirectory(ctx context.Context, dirPath string) ([]types.Result, error)
}
