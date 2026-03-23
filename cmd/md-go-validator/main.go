// md-go-validator validates Go code blocks in Markdown files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
)

// osExit allows mocking os.Exit in tests.
//
//nolint:gochecknoglobals // Required for testing os.Exit behavior
var osExit = os.Exit

type config struct {
	verbose  bool
	showCode bool
	paths    []string
}

func main() {
	cfg := parseArgs(os.Args[1:])
	validator := mdgovalidator.New(cfg.verbose)
	allResults := validatePaths(validator, cfg.paths)
	mdgovalidator.PrintReport(allResults, cfg.showCode)

	if mdgovalidator.HasErrors(allResults) {
		osExit(1)
	}
}

func parseArgs(args []string) config {
	cfg := config{verbose: false, showCode: true, paths: []string{}}

	for _, arg := range args {
		switch arg {
		case "-v", "--verbose":
			cfg.verbose = true
		case "-q", "--quiet", "--no-code":
			cfg.showCode = false
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				//nolint:gosec // CLI tool - user controls input via command line
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n\n", arg)
				printUsage()
				os.Exit(1)
			}
			cfg.paths = append(cfg.paths, arg)
		}
	}

	if len(cfg.paths) == 0 {
		cfg.paths = []string{"."}
	}

	return cfg
}

func validatePaths(validator *mdgovalidator.Validator, paths []string) []mdgovalidator.Result {
	// Pre-allocate with estimated capacity (each path may produce multiple results)
	allResults := make([]mdgovalidator.Result, 0, len(paths)*10)

	for _, path := range paths {
		results := validatePath(validator, path)
		allResults = append(allResults, results...)
	}

	return allResults
}

func validatePath(validator *mdgovalidator.Validator, path string) []mdgovalidator.Result {
	absPath, err := filepath.Abs(path)
	if err != nil {
		//nolint:gosec // CLI tool - user controls input via command line
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", path, err)
		return nil
	}

	//nolint:gosec // CLI tool - user controls input via command line
	info, err := os.Stat(absPath)
	if err != nil {
		//nolint:gosec // CLI tool - user controls input via command line
		fmt.Fprintf(os.Stderr, "Path %s does not exist\n", absPath)
		return nil
	}

	var results []mdgovalidator.Result
	if info.IsDir() {
		results, err = validator.ValidateDirectory(absPath)
	} else {
		results, err = validator.ValidateFile(absPath)
	}
	if err != nil {
		//nolint:gosec // CLI tool - user controls input via command line
		fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", absPath, err)
		return nil
	}

	return results
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
