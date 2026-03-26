// Command md-go-validator validates Go code blocks in Markdown files.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
	"github.com/larsartmann/md-go-validator/pkg/output"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// osExit allows mocking os.Exit in tests.
//
//nolint:gochecknoglobals // Required for testing os.Exit behavior
var osExit = os.Exit

type config struct {
	verbose    bool
	showCode   bool
	format     output.OutputFormat
	colorMode  output.ColorMode
	outputFile string
	paths      []string
}

func main() {
	cfg := parseArgs(os.Args[1:])
	validator := mdgovalidator.New(cfg.verbose)
	ctx := context.Background()
	allResults := validatePaths(validator, ctx, cfg.paths)

	if cfg.outputFile != "" {
		if err := writeOutputToFile(allResults, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			osExit(1)
			return
		}
	} else {
		output.PrintReport(allResults, cfg.format, cfg.colorMode, cfg.showCode)
	}

	if mdgovalidator.HasErrors(allResults) {
		osExit(1)
	}
}

func parseArgs(args []string) config {
	cfg := config{
		verbose:    false,
		showCode:   true,
		format:     output.FormatTable,
		colorMode:  output.ColorModeAuto,
		outputFile: "",
		paths:      []string{},
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
		case "-o", "--output":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --output requires an argument\n\n")
				printUsage()
				os.Exit(1)
			}
			i++
			cfg.outputFile = args[i]
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

func validatePaths(validator mdgovalidator.Validator, ctx context.Context, paths []string) []types.Result {
	// Pre-allocate with estimated capacity (each path may produce multiple results)
	allResults := make([]types.Result, 0, len(paths)*10)

	for _, path := range paths {
		results := validatePath(validator, ctx, path)
		allResults = append(allResults, results...)
	}

	return allResults
}

func validatePath(validator mdgovalidator.Validator, ctx context.Context, path string) []types.Result {
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

	var results []types.Result
	if info.IsDir() {
		results, err = validator.ValidateDirectory(ctx, absPath)
	} else {
		results, err = validator.ValidateFile(ctx, absPath)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", absPath, err)
		return nil
	}

	return results
}

func writeOutputToFile(results []types.Result, cfg config) error {
	if err := os.MkdirAll(filepath.Dir(cfg.outputFile), 0o755); err != nil {
		return fmt.Errorf("create parent directories: %w", err)
	}

	file, err := os.Create(cfg.outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer file.Close()

	return output.PrintReportTo(file, results, cfg.format, cfg.colorMode, cfg.showCode)
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
    -o, --output     Write output to file (creates parent dirs if needed)
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
    md-go-validator .                       # Validate all .md files
    md-go-validator README.md               # Validate a specific file
    md-go-validator -v .                    # Verbose output
    md-go-validator -f json .               # JSON output for CI
    md-go-validator -f markdown .           # Markdown table output
    md-go-validator --color never .         # Disable colors
    md-go-validator -o report.json -f json . # Write JSON to file`)
}
