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

// Result contains the result of validating a single code block
type Result struct {
	File       string
	LineNumber int
	CodeBlock  int
	Code       string
	Skipped    bool
	Error      error
}

// Validator validates Go code blocks in markdown files
type Validator struct {
	fset    *token.FileSet
	verbose bool
}

// New creates a new validator
func New(verbose bool) *Validator {
	return &Validator{
		fset:    token.NewFileSet(),
		verbose: verbose,
	}
}

// ValidateFile validates a single markdown file
func (v *Validator) ValidateFile(filePath string) ([]Result, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	blocks := ExtractGoCodeBlocks(string(content))
	if len(blocks) == 0 {
		return []Result{}, nil
	}

	var results []Result
	for i, block := range blocks {
		result := Result{
			File:       filePath,
			LineNumber: block.LineNumber,
			CodeBlock:  i + 1,
			Code:       block.Code,
			Skipped:    block.Skipped,
		}

		if !block.Skipped {
			result.Error = ValidateGoCode(block.Code)
		}

		results = append(results, result)

		if v.verbose {
			if block.Skipped {
				fmt.Printf("  ⏭️  Block %d (line %d): SKIPPED\n", i+1, block.LineNumber)
			} else if result.Error != nil {
				fmt.Printf("  ❌ Block %d (line %d): %v\n", i+1, block.LineNumber, result.Error)
			} else {
				fmt.Printf("  ✅ Block %d (line %d): OK\n", i+1, block.LineNumber)
			}
		}
	}

	return results, nil
}

// ValidateDirectory validates all markdown files in a directory (recursively)
func (v *Validator) ValidateDirectory(dirPath string) ([]Result, error) {
	var allResults []Result

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "build" ||
				name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		if v.verbose {
			fmt.Printf("\n📄 Validating: %s\n", path)
		}

		results, err := v.ValidateFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			return nil
		}

		allResults = append(allResults, results...)
		return nil
	})

	return allResults, err
}

// PrintReport prints a summary report of validation results
func PrintReport(results []Result, showCode bool) {
	var errors []Result
	var valid, skipped int

	for _, r := range results {
		if r.Skipped {
			skipped++
		} else if r.Error != nil {
			errors = append(errors, r)
		} else {
			valid++
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 VALIDATION REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total code blocks: %d\n", len(results))
	fmt.Printf("✅ Valid: %d\n", valid)
	fmt.Printf("⏭️  Skipped: %d\n", skipped)
	fmt.Printf("❌ Invalid: %d\n", len(errors))

	if len(errors) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("❌ ERRORS FOUND:")
		fmt.Println(strings.Repeat("-", 60))

		for _, e := range errors {
			fmt.Printf("\n📍 %s:%d (block #%d)\n", e.File, e.LineNumber, e.CodeBlock)
			fmt.Printf("   Error: %v\n", e.Error)

			if showCode {
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
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}

// HasErrors returns true if any results have errors (excluding skipped)
func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Error != nil && !r.Skipped {
			return true
		}
	}
	return false
}
