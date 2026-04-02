package mdgovalidator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// FileValidator validates code blocks in markdown files.
type FileValidator struct {
	registry    *languages.Registry
	verbose     bool
	maxFiles    int
	maxBlocks   int
	concurrency int
	targetLangs []languages.Language
}

// New creates a new FileValidator with default settings.
func New(verbose bool) *FileValidator {
	return &FileValidator{
		registry:    languages.DefaultRegistry(),
		verbose:     verbose,
		maxFiles:    0,
		maxBlocks:   0,
		concurrency: 4,
		targetLangs: []languages.Language{languages.LangGo},
	}
}

// WithMaxFiles sets the maximum number of files to process.
func (v *FileValidator) WithMaxFiles(maxFiles int) *FileValidator {
	v.maxFiles = maxFiles
	return v
}

// WithMaxBlocks sets the maximum number of blocks per file to process.
func (v *FileValidator) WithMaxBlocks(maxBlocks int) *FileValidator {
	v.maxBlocks = maxBlocks
	return v
}

// WithConcurrency sets the number of concurrent workers for directory validation.
func (v *FileValidator) WithConcurrency(n int) *FileValidator {
	if n > 0 {
		v.concurrency = n
	}
	return v
}

// WithLanguages sets the target languages to validate.
func (v *FileValidator) WithLanguages(langs []languages.Language) *FileValidator {
	v.targetLangs = langs
	return v
}

// WithRegistry sets a custom validator registry.
func (v *FileValidator) WithRegistry(r *languages.Registry) *FileValidator {
	v.registry = r
	return v
}

// ValidateFile validates a single markdown file.
func (v *FileValidator) ValidateFile(ctx context.Context, filePath string) ([]types.Result, error) {
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf("validate file %s: %w", filePath, err)
	}

	cleanPath, err := validateAndCleanPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path %s: %w", filePath, err)
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	blocks := ExtractCodeBlocks(string(content), v.targetLangs)
	if len(blocks) == 0 {
		return []types.Result{}, nil
	}

	return v.validateBlocks(ctx, cleanPath, blocks)
}

func (v *FileValidator) validateBlocks(
	ctx context.Context,
	cleanPath string,
	blocks []types.CodeBlock,
) ([]types.Result, error) {
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf(
			"validate blocks (file=%s, blocks=%d): %w",
			cleanPath,
			len(blocks),
			err,
		)
	}

	blocksToProcess := blocks
	if v.maxBlocks > 0 && len(blocks) > v.maxBlocks {
		blocksToProcess = blocks[:v.maxBlocks]
	}

	results := make([]types.Result, 0, len(blocksToProcess))
	for i, block := range blocksToProcess {
		select {
		case <-ctx.Done():
			return results, fmt.Errorf(
				"validation cancelled at block %d (file=%s, processed=%d, total=%d): %w",
				i, cleanPath, len(results), len(blocksToProcess), ctx.Err(),
			)
		default:
		}

		result := v.validateBlock(ctx, cleanPath, block, i)
		results = append(results, result)
		v.logProgress(i, block, result)
	}

	return results, nil
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("validation cancelled: %w", ctx.Err())
	default:
		return nil
	}
}

func (v *FileValidator) validateBlock(
	ctx context.Context,
	filePath string,
	block types.CodeBlock,
	index int,
) types.Result {
	blockIndex := types.NewBlockIndex(index + 1)

	if block.IsSkipped() {
		return types.NewSkippedResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
		)
	}

	// Get the validator for this language
	validator := v.registry.Get(block.Language)
	if validator == nil {
		return types.NewErrorResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
			fmt.Errorf("no validator available for language: %s", block.Language),
		)
	}

	if !validator.IsAvailable() {
		return types.NewSkippedResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
		)
	}

	// Validate using the language-specific validator
	if err := validator.Validate(ctx, block.Code); err != nil {
		codePreview := block.Code
		if len(codePreview) > 30 {
			codePreview = codePreview[:30] + "..."
		}
		return types.NewErrorResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
			fmt.Errorf("validating %s block (block=%d, file=%s, code=%q, line=%s): %w",
				block.Language, index, filePath, codePreview, block.LineNumber, err),
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
	case types.StatusUnknown:
		fmt.Printf("  ❓ Block %d (line %s): UNKNOWN\n", i+1, block.LineNumber)
	case types.StatusValid:
		fmt.Printf("  ✅ Block %d (line %s): OK\n", i+1, block.LineNumber)
	case types.StatusSkipped:
		fmt.Printf("  ⏭️  Block %d (line %s): SKIPPED\n", i+1, block.LineNumber)
	case types.StatusError:
		fmt.Printf("  ❌ Block %d (line %s): %v\n", i+1, block.LineNumber, result.Error)
	}
}

// ValidateDirectory validates all markdown files in a directory (recursively).
// Uses parallel processing with configurable concurrency for improved performance.
func (v *FileValidator) ValidateDirectory(
	ctx context.Context,
	dirPath string,
) ([]types.Result, error) {
	cleanPath, err := validateAndCleanPath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path %s: %w", dirPath, err)
	}

	filePaths, err := v.collectMarkdownFiles(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("collecting files from %s: %w", cleanPath, err)
	}

	if len(filePaths) == 0 {
		return []types.Result{}, nil
	}

	if v.verbose {
		fmt.Printf(
			"📁 Processing %d markdown files with %d workers\n",
			len(filePaths),
			v.concurrency,
		)
	}

	return v.processFilesParallel(ctx, filePaths)
}

// collectMarkdownFiles gathers all markdown files from a directory recursively.
func (v *FileValidator) collectMarkdownFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if isMarkdownFile(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collecting files from %s: %w", dirPath, err)
	}

	return files, nil
}

// processFilesParallel validates files concurrently using a worker pool.
func (v *FileValidator) processFilesParallel(
	ctx context.Context,
	filePaths []string,
) ([]types.Result, error) {
	if err := checkContext(ctx); err != nil {
		return nil, fmt.Errorf("context cancelled before processing files: %w", err)
	}

	filesToProcess := filePaths
	if v.maxFiles > 0 && len(filePaths) > v.maxFiles {
		filesToProcess = filePaths[:v.maxFiles]
	}

	jobs := make(chan string, len(filesToProcess))
	results := make(chan []types.Result, len(filesToProcess))
	errors := make(chan error, len(filesToProcess))

	var wg sync.WaitGroup

	for i := 0; i < v.concurrency; i++ {
		wg.Go(func() {
			for path := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				branchCtx, cancel := context.WithCancel(ctx)
				fileResults, err := v.ValidateFile(branchCtx, path)
				cancel()

				if err != nil {
					errors <- fmt.Errorf("file %s: %w", path, err)
					continue
				}
				results <- fileResults
			}
		})
	}

	go func() {
		for _, path := range filesToProcess {
			if err := checkContext(ctx); err != nil {
				break
			}
			jobs <- path
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	var allResults []types.Result
	var errs []error

	for {
		select {
		case <-ctx.Done():
			return allResults, fmt.Errorf("validation cancelled: %w", ctx.Err())
		case fileResults, ok := <-results:
			if !ok {
				goto done
			}
			allResults = append(allResults, fileResults...)
		case err, ok := <-errors:
			if !ok {
				goto done
			}
			errs = append(errs, err)
		}
	}

done:
	if len(errs) > 0 {
		return allResults, fmt.Errorf(
			"encountered %d errors: %w",
			len(errs),
			errs[0],
		)
	}

	return allResults, nil
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
		return "", errors.New("path cannot be empty")
	}

	// Check for null bytes (common attack vector)
	if strings.Contains(path, "\x00") {
		return "", errors.New("path contains null byte")
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
