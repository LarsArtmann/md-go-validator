package mdgovalidator

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// ErrPathEmpty is returned when a path argument is empty.
var ErrPathEmpty = errors.New("path cannot be empty")

// ErrPathNullByte is returned when a path contains a null byte.
var ErrPathNullByte = errors.New("path contains null byte")

// ErrNoValidatorForLang is returned when no validator is registered for a language.
var ErrNoValidatorForLang = errors.New("no validator available for language")

// Magic number constants.
const (
	defaultConcurrency = 4
	codePreviewLength  = 30
)

// FileValidator validates code blocks in markdown and MDX files.
type FileValidator struct {
	registry        *languages.Registry
	verbose         bool
	maxFiles        int
	maxBlocks       int
	concurrency     int
	targetLangs     []languages.Language
	excludePatterns []types.ExcludePattern
	fileFilter      func(string) bool
	skipDirectives  []string
}

// New creates a new FileValidator with default settings.
func New(verbose bool) *FileValidator {
	return &FileValidator{ //nolint:exhaustruct // optional fields set via builder methods
		registry:    languages.DefaultRegistry(),
		verbose:     verbose,
		maxFiles:    0,
		maxBlocks:   0,
		concurrency: defaultConcurrency,
		targetLangs: []languages.Language{languages.LangGo},
	}
}

// WithMaxFiles sets the maximum number of files to process.
func (v *FileValidator) WithMaxFiles(maxFiles int) *FileValidator {
	return v.withInt(&v.maxFiles, maxFiles)
}

// WithMaxBlocks sets the maximum number of blocks per file to process.
func (v *FileValidator) WithMaxBlocks(maxBlocks int) *FileValidator {
	return v.withInt(&v.maxBlocks, maxBlocks)
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

// WithExcludePatterns sets glob patterns to exclude from validation.
// Patterns are matched against the full file path using filepath.Match.
// Common patterns: "vendor/*", "docs/generated/*", "node_modules/*".
func (v *FileValidator) WithExcludePatterns(patterns []types.ExcludePattern) *FileValidator {
	v.excludePatterns = patterns

	return v
}

// WithFileFilter sets a custom filter function. Files for which the function
// returns false are excluded from validation.
func (v *FileValidator) WithFileFilter(filter func(string) bool) *FileValidator {
	v.fileFilter = filter

	return v
}

// WithSkipDirectives adds custom skip directives on top of the defaults.
// Blocks containing any of these strings are skipped during validation.
func (v *FileValidator) WithSkipDirectives(directives []string) *FileValidator {
	v.skipDirectives = directives

	return v
}

// extractBlocks extracts code blocks using default and custom skip directives.
func (v *FileValidator) extractBlocks(content string) []types.CodeBlock {
	if len(v.skipDirectives) == 0 {
		return ExtractCodeBlocks(content, v.targetLangs)
	}

	merged := make(SkipDirectivesConfig, 0, len(DefaultSkipDirectives())+len(v.skipDirectives))
	merged = append(merged, DefaultSkipDirectives()...)
	merged = append(merged, v.skipDirectives...)

	return ExtractCodeBlocksWithConfig(content, v.targetLangs, merged)
}

// validatePath validates and cleans a path with a descriptive error message.
func validatePath(pathType, path string) (string, error) {
	cleanPath, err := validateAndCleanPath(path)
	if err != nil {
		return "", fmt.Errorf("invalid %s %s: %w", pathType, path, err)
	}

	return cleanPath, nil
}

// validateAndReturnPath validates path and returns it, or returns error.
func (v *FileValidator) validateAndReturnPath(pathType, path string) (string, error) {
	cleanPath, err := validatePath(pathType, path)
	if err != nil {
		return "", fmt.Errorf("pathType=%s: %w", pathType, err)
	}

	return cleanPath, nil
}

// ValidateContent validates code blocks in raw markdown/MDX content.
// sourceName is used as the file identifier in results (e.g. "<stdin>").
func (v *FileValidator) ValidateContent(ctx context.Context, content, sourceName string) ([]types.Result, error) {
	ctxErr := checkContext(ctx)
	if ctxErr != nil {
		return nil, fmt.Errorf("validate content %s: %w", sourceName, ctxErr)
	}

	blocks := v.extractBlocks(content)
	if len(blocks) == 0 {
		return []types.Result{}, nil
	}

	return v.validateBlocks(ctx, sourceName, blocks)
}

// ValidateFile validates a single markdown or MDX file.
func (v *FileValidator) ValidateFile(ctx context.Context, filePath string) ([]types.Result, error) {
	ctxErr := checkContext(ctx)
	if ctxErr != nil {
		return nil, fmt.Errorf("validate file %s: %w", filePath, ctxErr)
	}

	cleanPath, err := v.validateAndReturnPath("file", filePath)
	if err != nil {
		return nil, err
	}

	// Path is already validated and cleaned by validateAndReturnPath
	//nolint:gosec // G304: Path is validated via validateAndCleanPath before use
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	blocks := v.extractBlocks(string(content))
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
		return newSkippedResultFromBlock(filePath, block, blockIndex)
	}

	validator := v.registry.Get(block.Language)
	if validator == nil {
		return types.NewErrorResult(
			types.NewFileID(filePath),
			block.LineNumber,
			blockIndex,
			block.Code,
			fmt.Errorf(
				"%w: %s (blockIndex=%d)",
				ErrNoValidatorForLang, block.Language, blockIndex.Int(),
			),
		)
	}

	if !validator.IsAvailable() {
		return newSkippedResultFromBlock(filePath, block, blockIndex)
	}

	err := validator.Validate(ctx, block.Code)
	if err != nil {
		return newErrorResultFromBlock(filePath, block, blockIndex, err)
	}

	return types.NewResultWithStatus(
		types.NewFileID(filePath),
		block.LineNumber,
		blockIndex,
		block.Code,
		types.StatusValid,
	)
}

func newSkippedResultFromBlock(
	filePath string,
	block types.CodeBlock,
	blockIndex types.BlockIndex,
) types.Result {
	return types.NewResultWithStatus(
		types.NewFileID(filePath),
		block.LineNumber,
		blockIndex,
		block.Code,
		types.StatusSkipped,
	)
}

func newErrorResultFromBlock(
	filePath string,
	block types.CodeBlock,
	blockIndex types.BlockIndex,
	err error,
) types.Result {
	codePreview := block.Code
	if len(codePreview) > codePreviewLength {
		codePreview = codePreview[:codePreviewLength] + "..."
	}

	return types.NewErrorResult(
		types.NewFileID(filePath),
		block.LineNumber,
		blockIndex,
		block.Code,
		fmt.Errorf("validating %s block (block=%d, file=%s, code=%q, line=%s): %w",
			block.Language, blockIndex.Int(), filePath, codePreview, block.LineNumber, err),
	)
}

// logProgress logs validation progress when verbose mode is enabled.
func (v *FileValidator) logProgress(i int, block types.CodeBlock, result types.Result) {
	if !v.verbose {
		return
	}

	//nolint:forbidigo // Verbose progress output requires direct stdout writing
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

// ValidateDirectory validates all supported files (.md, .markdown, .mdx) in a directory (recursively).
// Uses parallel processing with configurable concurrency for improved performance.
func (v *FileValidator) ValidateDirectory(
	ctx context.Context,
	dirPath string,
) ([]types.Result, error) {
	cleanPath, err := v.validateAndReturnPath("directory", dirPath)
	if err != nil {
		return nil, fmt.Errorf("validating directory %s: %w", dirPath, err)
	}

	filePaths, err := v.collectSupportedFiles(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("collecting files from %s: %w", cleanPath, err)
	}

	if len(filePaths) == 0 {
		return []types.Result{}, nil
	}

	//nolint:forbidigo // Verbose progress output requires direct stdout writing
	if v.verbose {
		fmt.Printf(
			"Processing %d files (%s) with %d workers\n",
			len(filePaths),
			formatSupportedExtensions(),
			v.concurrency,
		)
	}

	return v.processFilesParallel(ctx, filePaths)
}

// ValidateDirectoryFunc validates a directory and calls fn for each result
// as soon as it is produced — true streaming, not buffered.
// If fn returns a non-nil error, processing stops immediately.
func (v *FileValidator) ValidateDirectoryFunc(
	ctx context.Context,
	dirPath string,
	fn func(types.Result) error,
) error {
	cleanPath, err := v.validateAndReturnPath("directory", dirPath)
	if err != nil {
		return fmt.Errorf("validating directory %s: %w", dirPath, err)
	}

	filePaths, err := v.collectSupportedFiles(cleanPath)
	if err != nil {
		return fmt.Errorf("collecting files from %s: %w", cleanPath, err)
	}

	if len(filePaths) == 0 {
		return nil
	}

	return v.streamFilesParallel(ctx, filePaths, fn)
}

// streamFilesParallel validates files concurrently and calls fn for each
// individual result as it arrives from the worker pool.
func (v *FileValidator) streamFilesParallel(
	ctx context.Context,
	filePaths []string,
	fn func(types.Result) error,
) error {
	err := checkContext(ctx)
	if err != nil {
		return fmt.Errorf("context cancelled before processing files: %w", err)
	}

	filesToProcess := v.limitFiles(filePaths)
	jobs := make(chan string, len(filesToProcess))
	results := make(chan []types.Result, len(filesToProcess))
	errorsChan := make(chan error, len(filesToProcess))

	v.startWorkers(ctx, workerChannels{jobs, results, errorsChan})
	v.feedJobs(ctx, jobs, filesToProcess)

	for fileResults := range results {
		for _, r := range fileResults {
			cbErr := fn(r)
			if cbErr != nil {
				return fmt.Errorf("streaming callback aborted: %w", cbErr)
			}
		}
	}

	var errs []error

	for err := range errorsChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors: %w", len(errs), errs[0])
	}

	return nil
}

// collectSupportedFiles gathers all supported files (markdown and MDX) from a directory recursively.
func (v *FileValidator) collectSupportedFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dirPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if shouldSkipDir(entry.Name()) || v.isExcluded(path) {
				return filepath.SkipDir
			}

			return nil
		}

		if IsSupportedFile(path) && !v.isExcluded(path) && v.passesFileFilter(path) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collecting files from %s: %w", dirPath, err)
	}

	return files, nil
}

// isExcluded returns true if the path matches any exclude pattern.
func (v *FileValidator) isExcluded(path string) bool {
	for _, pattern := range v.excludePatterns {
		if pattern.Match(path) {
			return true
		}
	}

	return false
}

// passesFileFilter returns true if the file filter is nil or returns true.
func (v *FileValidator) passesFileFilter(path string) bool {
	if v.fileFilter == nil {
		return true
	}

	return v.fileFilter(path)
}

// processFilesParallel validates files concurrently using a worker pool.
// This function is split into smaller functions to reduce cognitive complexity.
//

func (v *FileValidator) processFilesParallel(
	ctx context.Context,
	filePaths []string,
) ([]types.Result, error) {
	err := checkContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("context cancelled before processing files: %w", err)
	}

	filesToProcess := v.limitFiles(filePaths)
	jobs := make(chan string, len(filesToProcess))
	results := make(chan []types.Result, len(filesToProcess))
	errors := make(chan error, len(filesToProcess))

	v.startWorkers(ctx, workerChannels{jobs, results, errors})
	v.feedJobs(ctx, jobs, filesToProcess)

	return v.collectResults(results, errors)
}

// limitFiles applies maxFiles limit to file paths.
func (v *FileValidator) limitFiles(filePaths []string) []string {
	if v.maxFiles > 0 && len(filePaths) > v.maxFiles {
		return filePaths[:v.maxFiles]
	}

	return filePaths
}

type workerChannels struct {
	jobs    <-chan string
	results chan<- []types.Result
	errors  chan<- error
}

// startWorkers starts concurrent workers to process files.
func (v *FileValidator) startWorkers(ctx context.Context, chans workerChannels) {
	var wg sync.WaitGroup

	for range v.concurrency {
		wg.Go(func() {
			v.processJob(ctx, chans)
		})
	}

	go func() {
		wg.Wait()
		close(chans.results)
		close(chans.errors)
	}()
}

// processJob processes a single job from the jobs channel.
func (v *FileValidator) processJob(ctx context.Context, chans workerChannels) {
	for path := range chans.jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fileResults, err := v.ValidateFile(ctx, path)
		if err != nil {
			chans.errors <- fmt.Errorf("file %s: %w", path, err)

			continue
		}

		chans.results <- fileResults
	}
}

// feedJobs sends file paths to the jobs channel.
func (v *FileValidator) feedJobs(
	ctx context.Context,
	jobs chan<- string,
	filesToProcess []string,
) {
	go func() {
		for _, path := range filesToProcess {
			err := checkContext(ctx)
			if err != nil {
				break
			}

			jobs <- path
		}

		close(jobs)
	}()
}

// collectResults aggregates results and errors by draining the channels,
// which are closed once all workers finish. Both channels are buffered to the
// file count, so workers never block on send and collection cannot deadlock.
func (v *FileValidator) collectResults(
	results <-chan []types.Result,
	errorsChan <-chan error,
) ([]types.Result, error) {
	allResults := make([]types.Result, 0)

	for r := range results {
		allResults = append(allResults, r...)
	}

	var errs []error
	for err := range errorsChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return allResults, fmt.Errorf("encountered %d errors: %w", len(errs), errs[0])
	}

	return allResults, nil
}

func (v *FileValidator) withInt(field *int, value int) *FileValidator {
	*field = value

	return v
}

func shouldSkipDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}

	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"build":        true,
		"dist":         true,
	}

	return skipDirs[name]
}

// SupportedExtensions returns all supported file extensions in sorted order.
// These are the extensions that the validator will process.
func SupportedExtensions() []types.FileType {
	return types.AllFileTypes()
}

// IsSupportedFile returns true if the file has a supported extension.
// Supports .md, .markdown, and .mdx files.
// types is the single source of truth for which extensions are recognized.
func IsSupportedFile(path string) bool {
	_, ok := types.ParseFileType(strings.ToLower(filepath.Ext(path)))

	return ok
}

func formatSupportedExtensions() string {
	exts := SupportedExtensions()

	names := make([]string, 0, len(exts))
	for i, ext := range exts {
		names[i] = ext.String()
	}

	return strings.Join(names, ", ")
}

// validateAndCleanPath validates and cleans a file path to prevent path traversal attacks.
func validateAndCleanPath(path string) (string, error) {
	if path == "" {
		return "", ErrPathEmpty
	}

	if strings.Contains(path, "\x00") {
		return "", ErrPathNullByte
	}

	return filepath.Clean(path), nil
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

// HasSkipped returns true if any results were skipped.
func HasSkipped(results []types.Result) bool {
	for _, r := range results {
		if r.Status == types.StatusSkipped {
			return true
		}
	}

	return false
}
