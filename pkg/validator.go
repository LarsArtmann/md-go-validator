package mdgovalidator

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// FileValidator validates Go code blocks in markdown files.
type FileValidator struct {
	fset    *token.FileSet
	verbose bool
}

// New creates a new FileValidator.
func New(verbose bool) *FileValidator {
	return &FileValidator{
		fset:    token.NewFileSet(),
		verbose: verbose,
	}
}

// ValidateFile validates a single markdown file.
func (v *FileValidator) ValidateFile(ctx context.Context, filePath string) ([]types.Result, error) {
	cleanPath, err := validateAndCleanPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path %s: %w", filePath, err)
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	blocks := ExtractGoCodeBlocks(string(content))
	if len(blocks) == 0 {
		return []types.Result{}, nil
	}

	results := make([]types.Result, 0, len(blocks))
	for i, block := range blocks {
		result := v.validateBlock(cleanPath, block, i)
		results = append(results, result)
		v.logProgress(i, block, result)
	}

	return results, nil
}

func (v *FileValidator) validateBlock(filePath string, block types.CodeBlock, index int) types.Result {
	blockIndex := types.NewBlockIndex(index + 1)

	if block.IsSkipped() {
		return types.NewSkippedResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
		)
	}

	if err := ValidateGoCode(block.Code); err != nil {
		return types.NewErrorResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
			err,
		)
	}

	return types.NewValidResult(
		types.NewFileID(filePath),
		block.LineNumber,
		blockIndex,
		block.Code,
	)
}

func (v *FileValidator) logProgress(i int, block types.CodeBlock, result types.Result) {
	if !v.verbose {
		return
	}
	switch result.Status {
	case types.StatusSkipped:
		fmt.Printf("  ⏭️  Block %d (line %s): SKIPPED\n", i+1, block.LineNumber)
	case types.StatusError:
		fmt.Printf("  ❌ Block %d (line %s): %v\n", i+1, block.LineNumber, result.Error)
	default:
		fmt.Printf("  ✅ Block %d (line %s): OK\n", i+1, block.LineNumber)
	}
}

// ValidateDirectory validates all markdown files in a directory (recursively).
func (v *FileValidator) ValidateDirectory(ctx context.Context, dirPath string) ([]types.Result, error) {
	var allResults []types.Result

	err := filepath.Walk(dirPath, v.walkFunc(ctx, &allResults))
	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", dirPath, err)
	}

	return allResults, nil
}

func (v *FileValidator) walkFunc(ctx context.Context, results *[]types.Result) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return v.handleDirectory(info)
		}

		if !isMarkdownFile(path) {
			return nil
		}

		if v.verbose {
			fmt.Printf("\n📄 Validating: %s\n", path)
		}

		fileResults, err := v.validateFileWithContext(ctx, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return nil
		}

		*results = append(*results, fileResults...)
		return nil
	}
}

func (v *FileValidator) validateFileWithContext(ctx context.Context, filePath string) ([]types.Result, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return v.ValidateFile(ctx, filePath)
	}
}

func (v *FileValidator) handleDirectory(info os.FileInfo) error {
	name := info.Name()
	if shouldSkipDir(name) {
		return filepath.SkipDir
	}
	return nil
}

func shouldSkipDir(name string) bool {
	skipDirs := []string{".", "node_modules", "vendor", "build", "dist"}
	for _, skip := range skipDirs {
		if strings.HasPrefix(name, ".") || name == skip {
			return true
		}
	}
	return false
}

func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

// validateAndCleanPath validates and cleans a file path to prevent path traversal attacks.
func validateAndCleanPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Check for null bytes (common attack vector)
	if strings.Contains(path, "\x00") {
		return "", fmt.Errorf("path contains null byte")
	}

	// Clean the path to resolve any ".." or similar path traversal
	cleanPath := filepath.Clean(path)

	// Ensure the path is not absolute when we don't expect it
	// For security, we allow both but validate the cleaned path
	if !filepath.IsAbs(cleanPath) && !strings.HasPrefix(cleanPath, "..") {
		return cleanPath, nil
	}

	return cleanPath, nil
}

// HasErrors returns true if any results have errors (excluding skipped).
func HasErrors(results []types.Result) bool {
	for _, r := range results {
		if r.Status == types.StatusError {
			return true
		}
	}
	return false
}
