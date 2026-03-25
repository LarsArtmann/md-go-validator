// Command md-go-validator validates Go code blocks in Markdown files.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
	"github.com/larsartmann/md-go-validator/pkg/output"
)

// osExit allows mocking os.Exit in tests.
//
//nolint:gochecknoglobals // Required for testing os.Exit behavior
var osExit = os.Exit

type config struct {
	verbose   bool
	showCode  bool
	format    output.OutputFormat
	colorMode output.ColorMode
	paths     []string
}

func main() {
	cfg := parseArgs(os.Args[1:])
	validator := mdgovalidator.New(cfg.verbose)
	allResults := validatePaths(validator, cfg.paths)
	output.PrintReport(allResults, cfg.format, cfg.colorMode, cfg.showCode)

	if mdgovalidator.HasErrors(allResults) {
		osExit(1)
	}
}

func parseArgs(args []string) config {
	cfg := config{
		verbose:   false,
		showCode:  true,
		format:    output.FormatTable,
		colorMode: output.ColorModeAuto,
		paths:     []string{},
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-v", "--verbose":
			cfg.verbose = true
		case "-q", "--quiet":
			cfg.showCode = false
			cfg.format = output.FormatQuiet
		case "--no-code":
			cfg.showCode = false
		case "-f", "--format":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --format requires an argument\n\n")
				printUsage()
				os.Exit(1)
			}
			i++
			format, err := output.ParseFormat(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				printUsage()
				os.Exit(1)
			}
			cfg.format = format
		case "--color":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --color requires an argument\n\n")
				printUsage()
				os.Exit(1)
			}
			i++
			colorMode, err := output.ParseColorMode(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
				printUsage()
				os.Exit(1)
			}
			cfg.colorMode = colorMode
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
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
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", path, err)
		return nil
	}

	info, err := os.Stat(absPath)
	if err != nil {
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
    -q, --quiet      Quiet mode (summary only, no code in errors)
    --no-code        Don't show code snippets in error output
    -f, --format     Output format (table, json, markdown, yaml, csv, quiet)
    --color          Color mode (auto, always, never)
    -h, --help       Show this help message

OUTPUT FORMATS:
    table    Terminal table (default)
    json     JSON output (machine-readable)
    markdown Markdown table
    yaml     YAML output
    csv      CSV output
    quiet    Summary only (no details)

COLOR MODES:
    auto     Respect NO_COLOR and CI detection (default)
    always   Force ANSI colors
    never    Disable colors

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
    md-go-validator -v .                 # Verbose output
    md-go-validator -f json .            # JSON output for CI
    md-go-validator -f markdown .         # Markdown table output
    md-go-validator --color never .      # Disable colors`)
}
