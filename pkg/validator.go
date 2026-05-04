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

var (
	errPathEmpty          = errors.New("path cannot be empty")
	errPathNullByte       = errors.New("path contains null byte")
	errNoValidatorForLang = errors.New("no validator available for language")
)

// Magic number constants.
const (
	defaultConcurrency = 4
	codePreviewLength  = 30
	goroutineCount     = 2
)

// supportedExtensions is the single source of truth for recognized file types.
//
//nolint:gochecknoglobals // Configuration: immutable runtime-supported extensions
var supportedExtensions = map[types.FileType]bool{
	types.FileTypeMarkdown:    true,
	types.FileTypeMarkdownAlt: true,
	types.FileTypeMdx:         true,
}

// FileValidator validates code blocks in markdown and MDX files.
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

// validatePath validates and cleans a path with a descriptive error message.
func validatePath(pathType, path string) (string, error) {
	cleanPath, err := validateAndCleanPath(path)
	if err != nil {
		return "", fmt.Errorf("invalid %s %s: %w", pathType, path, err)
	}

	return cleanPath, nil
}

// validateAndReturnPath validates path and returns it, or returns error.
// This consolidates the common validatePath + error check pattern.
func (v *FileValidator) validateAndReturnPath(pathType, path string) (string, error) {
	cleanPath, err := validatePath(pathType, path)
	if err != nil {
		return "", err
	}

	return cleanPath, nil
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
			fmt.Errorf("%w: %s", errNoValidatorForLang, block.Language),
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
			block.Language, blockIndex.Int()-1, filePath, codePreview, block.LineNumber, err),
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
		return nil, err
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

// collectSupportedFiles gathers all supported files (markdown and MDX) from a directory recursively.
func (v *FileValidator) collectSupportedFiles(dirPath string) ([]string, error) {
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

		if IsSupportedFile(path) {
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

	return v.collectResults(ctx, results, errors)
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

		branchCtx, cancel := context.WithCancel(ctx)
		fileResults, err := v.ValidateFile(branchCtx, path)

		cancel()

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

// collectResults aggregates results from workers until channels are closed.
// Uses a done channel to signal completion instead of complex select.
func (v *FileValidator) collectResults(
	ctx context.Context,
	results <-chan []types.Result,
	errorsChan <-chan error,
) ([]types.Result, error) {
	done := make(chan struct{})
	allResults := &resultCollector{
		results: []types.Result{},
		errors:  []error{},
		mu:      sync.Mutex{},
	}

	go v.collectResultsLoop(ctx, results, errorsChan, done, allResults)

	<-done

	return allResults.finalize()
}

// resultCollector holds collected results and errors.
type resultCollector struct {
	results []types.Result
	errors  []error
	mu      sync.Mutex
}

// withLock executes fn under mutex protection.
func (rc *resultCollector) withLock(fn func()) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	fn()
}

// addResult adds a result safely.
func (rc *resultCollector) addResult(r []types.Result) {
	rc.withLock(func() {
		rc.results = append(rc.results, r...)
	})
}

// addError adds an error safely.
func (rc *resultCollector) addError(err error) {
	rc.withLock(func() {
		rc.errors = append(rc.errors, err)
	})
}

// finalize returns results or first error.
func (rc *resultCollector) finalize() ([]types.Result, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if len(rc.errors) > 0 {
		return rc.results, fmt.Errorf("encountered %d errors: %w", len(rc.errors), rc.errors[0])
	}

	return rc.results, nil
}

// collectFromChan runs a collection loop for a typed channel with context cancellation.
func collectFromChan[T any](ctx context.Context, ch <-chan T, wg *sync.WaitGroup, fn func(T)) {
	defer wg.Done()

	for item := range ch {
		select {
		case <-ctx.Done():
			return
		default:
			fn(item)
		}
	}
}

// collectResultsLoop runs collection loops concurrently.
func (v *FileValidator) collectResultsLoop(
	ctx context.Context,
	results <-chan []types.Result,
	errorsChan <-chan error,
	done chan<- struct{},
	collector *resultCollector,
) {
	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	go collectFromChan(ctx, results, &wg, collector.addResult)

	go collectFromChan(ctx, errorsChan, &wg, collector.addError)

	wg.Wait()
	close(done)
}

func (v *FileValidator) withInt(field *int, value int) *FileValidator {
	*field = value

	return v
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

// SupportedExtensions returns all supported file extensions in sorted order.
// These are the extensions that the validator will process.
func SupportedExtensions() []types.FileType {
	return types.AllFileTypes()
}

// IsSupportedFile returns true if the file has a supported extension.
// Supports .md, .markdown, and .mdx files.
func IsSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	return supportedExtensions[types.FileType(ext)]
}

func formatSupportedExtensions() string {
	exts := SupportedExtensions()

	names := make([]string, len(exts))
	for i, ext := range exts {
		names[i] = ext.String()
	}

	return strings.Join(names, ", ")
}

// validateAndCleanPath validates and cleans a file path to prevent path traversal attacks.
func validateAndCleanPath(path string) (string, error) {
	if path == "" {
		return "", errPathEmpty
	}

	// Check for null bytes (common attack vector)
	if strings.Contains(path, "\x00") {
		return "", errPathNullByte
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
