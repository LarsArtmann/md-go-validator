// md-go-validator validates code blocks in Markdown files.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
	"github.com/larsartmann/md-go-validator/pkg/languages"
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
	format     output.Format
	colorMode  output.ColorMode
	outputFile string
	paths      []string
	timeout    time.Duration
	contextCfg mdgovalidator.ContextConfig
	languages  []languages.Language
}

func main() {
	cfg := parseArgs(os.Args[1:])
	validator := mdgovalidator.New(cfg.verbose).WithLanguages(cfg.languages)

	// Build context with timeout from config
	ctx, cancel := cfg.contextCfg.Build()
	defer cancel()

	allResults := validatePaths(ctx, validator, cfg.paths)

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

// argHandler defines a function type for handling an argument.
type argHandler func(args []string, i int, cfg *config) (int, bool)

// newArgHandlers creates the map of flag names to their handler functions.
// This is a function instead of a global to avoid gochecknoglobals lint violation.
func newArgHandlers() map[string]argHandler {
	return map[string]argHandler{
		"-v":         handleVerbose,
		"--verbose":  handleVerbose,
		"-q":         handleQuiet,
		"--quiet":    handleQuiet,
		"--no-code":  handleNoCode,
		"-f":         handleFormat,
		"--format":   handleFormat,
		"--color":    handleColor,
		"-o":         handleOutput,
		"--output":   handleOutput,
		"--timeout":  handleTimeout,
		"-t":         handleTimeout,
		"-l":         handleLanguages,
		"--language": handleLanguages,
		"-h":         handleHelp,
		"--help":     handleHelp,
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
		timeout:    0,
		contextCfg: mdgovalidator.DefaultContextConfig(),
		languages:  []languages.Language{languages.LangGo},
	}

	argHandlers := newArgHandlers()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if handler, ok := argHandlers[arg]; ok {
			advance, ok := handler(args, i, &cfg)
			if !ok {
				os.Exit(1)
			}
			i += advance
			continue
		}
		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
			printUsage()
			os.Exit(1)
		}
		cfg.paths = append(cfg.paths, arg)
	}

	if len(cfg.paths) == 0 {
		cfg.paths = []string{"."}
	}

	return cfg
}

func requireArg(args []string, i int, flagName string) bool {
	if i+1 >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: --%s requires an argument\n", flagName)
		printUsage()
		return false
	}
	return true
}

func hideCode(cfg *config) {
	cfg.showCode = false
}

func handleVerbose(_ []string, _ int, cfg *config) (int, bool) {
	cfg.verbose = true
	return 0, true
}

func handleQuiet(_ []string, _ int, cfg *config) (int, bool) {
	cfg.format = output.FormatQuiet
	hideCode(cfg)
	return 0, true
}

func handleNoCode(_ []string, _ int, cfg *config) (int, bool) {
	hideCode(cfg)
	return 0, true
}

func handleFormat(args []string, i int, cfg *config) (int, bool) {
	if !requireArg(args, i, "format") {
		return 0, false
	}
	format, err := output.ParseFormat(args[i+1])
	if err != nil {
		returnParseError("format", err)
		return 0, false
	}
	cfg.format = format
	return 1, true
}

func handleColor(args []string, i int, cfg *config) (int, bool) {
	if !requireArg(args, i, "color") {
		return 0, false
	}
	colorMode, err := output.ParseColorMode(args[i+1])
	if err != nil {
		returnParseError("color", err)
		return 0, false
	}
	cfg.colorMode = colorMode
	return 1, true
}

func returnParseError(name string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	printUsage()
}

func handleOutput(args []string, i int, cfg *config) (int, bool) {
	if !requireArg(args, i, "output") {
		return 0, false
	}
	cfg.outputFile = args[i+1]
	return 1, true
}

func handleHelp(_ []string, _ int, _ *config) (int, bool) {
	printUsage()
	os.Exit(0)
	return 0, false // exit called, but we need to return something
}

func handleTimeout(args []string, i int, cfg *config) (int, bool) {
	if !requireArg(args, i, "timeout") {
		return 0, false
	}
	duration, err := time.ParseDuration(args[i+1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid timeout value %q: %v\n", args[i+1], err)
		printUsage()
		return 0, false
	}
	cfg.timeout = duration
	cfg.contextCfg = cfg.contextCfg.WithTimeout(duration)
	return 1, true
}

func handleLanguages(args []string, i int, cfg *config) (int, bool) {
	if !requireArg(args, i, "language") {
		return 0, false
	}

	// Parse comma-separated list of languages
	langs := strings.Split(args[i+1], ",")
	cfg.languages = nil // Reset to empty slice

	for _, lang := range langs {
		lang = strings.TrimSpace(strings.ToLower(lang))
		parsed, ok := languages.ParseLanguage(lang)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: unsupported language %q\n", lang)
			fmt.Fprintf(
				os.Stderr,
				"Supported languages: %s\n",
				strings.Join(getLanguageNames(), ", "),
			)
			return 0, false
		}
		cfg.languages = append(cfg.languages, parsed)
	}

	return 1, true
}

func getLanguageNames() []string {
	names := make([]string, 0, len(languages.AllLanguages()))
	for _, lang := range languages.AllLanguages() {
		names = append(names, string(lang))
	}
	return names
}

func validatePaths(
	ctx context.Context,
	validator mdgovalidator.Validator,
	paths []string,
) []types.Result {
	allResults := make([]types.Result, 0, len(paths)*10)

	for _, path := range paths {
		results := validatePath(ctx, validator, path)
		allResults = append(allResults, results...)
	}

	return allResults
}

func validatePath(
	ctx context.Context,
	validator mdgovalidator.Validator,
	path string,
) []types.Result {
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
	dir := filepath.Dir(cfg.outputFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("create parent directories (%d results): %w", len(results), err)
		}
	}

	file, err := os.Create(cfg.outputFile)
	if err != nil {
		return fmt.Errorf(
			"create output file (%d results, path=%s): %w",
			len(results),
			cfg.outputFile,
			err,
		)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf(
			"close output file (%d results, path=%s): %w",
			len(results),
			cfg.outputFile,
			err,
		)
	}

	file, err = os.OpenFile(cfg.outputFile, os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open output file for writing (%d results, path=%s): %w",
			len(results), cfg.outputFile, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fmt.Errorf("close output file (%d results, path=%s): %w",
				len(results), cfg.outputFile, closeErr)
		}
	}()

	if err := output.PrintReportTo(
		file,
		results,
		cfg.format,
		cfg.colorMode,
		cfg.showCode,
	); err != nil {
		return fmt.Errorf("write report (%d results, format=%s): %w", len(results), cfg.format, err)
	}
	return nil
}

func printUsage() {
	fmt.Println(`md-go-validator - Validate code blocks in Markdown files

USAGE:
    md-go-validator [OPTIONS] [PATH...]

OPTIONS:
    -v, --verbose     Show progress for each code block
    -q, --quiet       Quiet mode (summary only, no code in errors)
    --no-code         Don't show code snippets in error output
    -f, --format      Output format (table, json, markdown, yaml, csv, quiet)
    --color           Color mode (auto, always, never)
    -o, --output      Write output to file (creates parent dirs if needed)
    -t, --timeout     Timeout for validation (e.g., 30s, 5m, 1h)
    -l, --language    Comma-separated list of languages to validate
                      (go, templ, typescript, tsx, nix, rust, hcl, terraform)
    -h, --help        Show this help message

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

SUPPORTED LANGUAGES:
    go          Go (built-in, always available)
    templ       Templ (requires 'templ' CLI)
    typescript  TypeScript (requires 'tsc')
    tsx         TypeScript JSX (requires 'tsc')
    nix         Nix (requires 'nix-instantiate')
    rust        Rust (requires 'rustc')
    hcl         HCL/Terraform (requires 'terraform' or 'hclfmt')
    terraform   Alias for HCL

SKIP DIRECTIVES:
    Add these to skip validation of specific code blocks:
    <!-- skip-validate -->
    <!-- skip-md-validate -->
    <!-- md-skip -->
    <!-- no-validate -->
    // skip-validate
    //nolint

EXAMPLES:
    md-go-validator .                           # Validate all .md files
    md-go-validator README.md                   # Validate a specific file
    md-go-validator -v .                        # Verbose output
    md-go-validator -f json .                 # JSON output for CI
    md-go-validator -l go,typescript .        # Validate Go and TypeScript
    md-go-validator -l templ,nix .            # Validate Templ and Nix
    md-go-validator --color never .             # Disable colors
    md-go-validator -o report.json -f json .  # Write JSON to file
    md-go-validator --timeout 30s .           # 30 second timeout`)
}
