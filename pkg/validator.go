package mdgovalidator

import (
	"bufio"
	"bytes"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Result contains the result of validating a single code block.
type Result struct {
	File       string
	LineNumber int
	CodeBlock  int
	Code       string
	Skipped    bool
	Error      error
}

// Validator validates Go code blocks in markdown files.
type Validator struct {
	fset    *token.FileSet
	verbose bool
}

// New creates a new validator.
func New(verbose bool) *Validator {
	return &Validator{
		fset:    token.NewFileSet(),
		verbose: verbose,
	}
}

// ValidateFile validates a single markdown file.
func (v *Validator) ValidateFile(filePath string) ([]Result, error) {
	//nolint:gosec // CLI tool - user controls input via command line
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", filePath, err)
	}

	blocks := ExtractGoCodeBlocks(string(content))
	if len(blocks) == 0 {
		return []Result{}, nil
	}

	results := make([]Result, 0, len(blocks))
	for i, block := range blocks {
		result := Result{
			File:       filePath,
			LineNumber: block.LineNumber,
			CodeBlock:  i + 1,
			Code:       block.Code,
			Skipped:    block.Skipped,
			Error:      nil,
		}

		if !block.Skipped {
			result.Error = ValidateGoCode(block.Code)
		}

		results = append(results, result)
		v.logProgress(i, block, result)
	}

	return results, nil
}

func (v *Validator) logProgress(i int, block CodeBlock, result Result) {
	if !v.verbose {
		return
	}
	switch {
	case block.Skipped:
		fmt.Printf("  ⏭️  Block %d (line %d): SKIPPED\n", i+1, block.LineNumber)
	case result.Error != nil:
		fmt.Printf("  ❌ Block %d (line %d): %v\n", i+1, block.LineNumber, result.Error)
	default:
		fmt.Printf("  ✅ Block %d (line %d): OK\n", i+1, block.LineNumber)
	}
}

// ValidateDirectory validates all markdown files in a directory (recursively).
func (v *Validator) ValidateDirectory(dirPath string) ([]Result, error) {
	var allResults []Result

	err := filepath.Walk(dirPath, v.walkFunc(&allResults))
	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", dirPath, err)
	}

	return allResults, nil
}

func (v *Validator) walkFunc(results *[]Result) filepath.WalkFunc {
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

		fileResults, err := v.ValidateFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return nil
		}

		*results = append(*results, fileResults...)
		return nil
	}
}

func (v *Validator) handleDirectory(info os.FileInfo) error {
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

// PrintReport prints a summary report of validation results.
func PrintReport(results []Result, showCode bool) {
	var errors []Result
	var valid, skipped int

	for _, r := range results {
		switch {
		case r.Skipped:
			skipped++
		case r.Error != nil:
			errors = append(errors, r)
		default:
			valid++
		}
	}

	printHeader(len(results), valid, skipped, len(errors))
	printErrors(errors, showCode)
	fmt.Println("\n" + strings.Repeat("=", 60))
}

func printHeader(total, valid, skipped, errorCount int) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 VALIDATION REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total code blocks: %d\n", total)
	fmt.Printf("✅ Valid: %d\n", valid)
	fmt.Printf("⏭️  Skipped: %d\n", skipped)
	fmt.Printf("❌ Invalid: %d\n", errorCount)
}

func printErrors(errors []Result, showCode bool) {
	if len(errors) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("❌ ERRORS FOUND:")
	fmt.Println(strings.Repeat("-", 60))

	for _, e := range errors {
		printError(e, showCode)
	}
}

func printError(e Result, showCode bool) {
	fmt.Printf("\n📍 %s:%d (block #%d)\n", e.File, e.LineNumber, e.CodeBlock)
	fmt.Printf("   Error: %v\n", e.Error)

	if !showCode {
		return
	}

	fmt.Println("\n   Code:")
	fmt.Println("   " + strings.Repeat("-", 50))
	scanner := bufio.NewScanner(bytes.NewBufferString(e.Code))
	lineNum := 1
	for scanner.Scan() {
		fmt.Printf("   %3d | %s\n", lineNum, scanner.Text())
		lineNum++
	}
	fmt.Println("   " + strings.Repeat("-", 50))
}

// HasErrors returns true if any results have errors (excluding skipped).
func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Error != nil && !r.Skipped {
			return true
		}
	}
	return false
}
