package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
)

func main() {
	verbose := false
	showCode := true
	paths := []string{}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-v", "--verbose":
			verbose = true
		case "-q", "--quiet":
			showCode = false
		case "--no-code":
			showCode = false
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n\n", arg)
				printUsage()
				os.Exit(1)
			}
			paths = append(paths, arg)
		}
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	validator := mdgovalidator.New(verbose)
	allResults := []mdgovalidator.Result{}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", path, err)
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Path %s does not exist\n", absPath)
			continue
		}

		var results []mdgovalidator.Result
		if info.IsDir() {
			results, err = validator.ValidateDirectory(absPath)
		} else {
			results, err = validator.ValidateFile(absPath)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", absPath, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	mdgovalidator.PrintReport(allResults, showCode)

	if mdgovalidator.HasErrors(allResults) {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`md-go-validator - Validate Go code blocks in Markdown files

USAGE:
    md-go-validator [OPTIONS] [PATH...]

OPTIONS:
    -v, --verbose    Show progress for each code block
    -q, --quiet      Only show summary (no code in errors)
    --no-code        Don't show code snippets in error output
    -h, --help       Show this help message

SKIP DIRECTIVES:
    Add these to skip validation of specific code blocks:
    <!-- skip-validate -->
    <!-- skip-md-validate -->
    <!-- md-skip -->
    <!-- no-validate -->
    // skip-validate
    //nolint

EXAMPLES:
    md-go-validator .                    # Validate all .md files
    md-go-validator README.md            # Validate a specific file
    md-go-validator docs/ README.md      # Validate multiple paths
    md-go-validator -v .                 # Verbose output`)
}
