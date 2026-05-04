// md-go-validator validates code blocks in Markdown and MDX files.
package main

import (
	"context"
	"errors"
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

var errUnsupportedLanguage = errors.New("unsupported language")

// Magic number constants.
const (
	resultsCapacityMultiplier = 10
	defaultDirPermissions     = 0o750
	defaultFilePermissions    = 0o600
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
		err := writeOutputToFile(allResults, cfg)
		if err != nil {
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

func requireArg(args []string, i int, flagName string) bool {
	if i+1 >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: --%s requires an argument\n", flagName)
		printUsage()

		return false
	}

	return true
}

// newArgHandlers creates the map of flag names to their handler functions.
// This is a function instead of a global to avoid gochecknoglobals lint violation.
func newArgHandlers() map[string]argHandler {
	verboseHandler := boolFlagHandler(func(c *config) { c.verbose = true })
	quietHandler := boolFlagHandler(
		func(c *config) { c.format = output.FormatQuiet; c.showCode = false },
	)
	noCodeHandler := boolFlagHandler(func(c *config) { c.showCode = false })
	formatHandler := singleValueArgHandler(
		"format",
		output.ParseFormat,
		func(c *config, f output.Format) { c.format = f },
	)
	colorHandler := singleValueArgHandler(
		"color",
		output.ParseColorMode,
		func(c *config, cm output.ColorMode) { c.colorMode = cm },
	)
	outputHandler := stringArgHandler("output", func(c *config, s string) { c.outputFile = s })
	timeoutHandler := singleValueArgHandler(
		"timeout",
		time.ParseDuration,
		func(c *config, d time.Duration) {
			c.timeout = d
			c.contextCfg = c.contextCfg.WithTimeout(d)
		},
	)
	languagesHandler := languagesArgHandler()

	return map[string]argHandler{
		"-v":         verboseHandler,
		"--verbose":  verboseHandler,
		"-q":         quietHandler,
		"--quiet":    quietHandler,
		"--no-code":  noCodeHandler,
		"-f":         formatHandler,
		"--format":   formatHandler,
		"--color":    colorHandler,
		"-o":         outputHandler,
		"--output":   outputHandler,
		"-t":         timeoutHandler,
		"--timeout":  timeoutHandler,
		"-l":         languagesHandler,
		"--language": languagesHandler,
		"-h":         handleHelp,
		"--help":     handleHelp,
	}
}

// boolFlagHandler creates a handler for boolean flags.
func boolFlagHandler(setter func(*config)) argHandler {
	return func(_ []string, _ int, cfg *config) (int, bool) {
		setter(cfg)

		return 0, true
	}
}

// singleValueArgHandler creates a handler for single-value parsed arguments.
func singleValueArgHandler[T any](
	flagName string,
	parse func(string) (T, error),
	setter func(*config, T),
) argHandler {
	return func(args []string, idx int, cfg *config) (int, bool) {
		if !requireArg(args, idx, flagName) {
			return 0, false
		}

		val, err := parse(args[idx+1])
		if err != nil {
			returnParseError(flagName, err)

			return 0, false
		}

		setter(cfg, val)

		return 1, true
	}
}

// parseStringValue is a parse function that returns the input string unchanged.
func parseStringValue(s string) (string, error) {
	return s, nil
}

func stringArgHandler(flagName string, setter func(*config, string)) argHandler {
	return singleValueArgHandler(flagName, parseStringValue, setter)
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

func returnParseError(_ string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	printUsage()
}

func handleHelp(_ []string, _ int, _ *config) (int, bool) {
	printUsage()
	os.Exit(0)

	return 0, false // exit called, but we need to return something
}

// languagesArgHandler creates a handler for language flags that accepts comma-separated language names.
func languagesArgHandler() argHandler {
	return listArgHandler(
		"language",
		parseLanguages,
		func(cfg *config, langs []languages.Language) {
			cfg.languages = langs
		},
	)
}

// listArgHandler creates a handler for flags that accept comma-separated values.
func listArgHandler[T any](
	flagName string,
	parser func(string) ([]T, error),
	setter func(*config, []T),
) argHandler {
	return singleValueArgHandler(flagName, wrapListParser(parser), setter)
}

// wrapListParser wraps a list parser to return a single-element slice.
func wrapListParser[T any](parser func(string) ([]T, error)) func(string) ([]T, error) {
	return parser
}

// parseLanguages parses a comma-separated string of language names.
func parseLanguages(s string) ([]languages.Language, error) {
	langStrs := strings.Split(s, ",")
	result := make([]languages.Language, 0, len(langStrs))

	for _, lang := range langStrs {
		lang = strings.TrimSpace(strings.ToLower(lang))

		parsed, ok := languages.ParseLanguage(lang)
		if !ok {
			return nil, fmt.Errorf("%w: %s", errUnsupportedLanguage, lang)
		}

		result = append(result, parsed)
	}

	return result, nil
}

func validatePaths(
	ctx context.Context,
	validator *mdgovalidator.FileValidator,
	paths []string,
) []types.Result {
	allResults := make([]types.Result, 0, len(paths)*resultsCapacityMultiplier)

	for _, path := range paths {
		results := validatePath(ctx, validator, path)
		allResults = append(allResults, results...)
	}

	return allResults
}

func validatePath(
	ctx context.Context,
	validator *mdgovalidator.FileValidator,
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
		err := os.MkdirAll(dir, defaultDirPermissions)
		if err != nil {
			return fmt.Errorf("create parent directories (%d results): %w", len(results), err)
		}
	}

	file, err := os.OpenFile(
		cfg.outputFile,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		defaultFilePermissions,
	)
	if err != nil {
		return fmt.Errorf("open output file for writing (%d results, path=%s): %w",
			len(results), cfg.outputFile, err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			err = fmt.Errorf("close output file (%d results, path=%s): %w",
				len(results), cfg.outputFile, closeErr)
		}
	}()

	printErr := output.PrintReportTo(
		file,
		results,
		cfg.format,
		cfg.colorMode,
		cfg.showCode,
	)
	if printErr != nil {
		return fmt.Errorf(
			"write report (%d results, format=%s): %w",
			len(results),
			cfg.format,
			printErr,
		)
	}

	return nil
}

func printUsage() {
	//nolint:forbidigo // CLI help output requires direct stdout writing
	fmt.Print(usageHeader())
	//nolint:forbidigo // CLI help output requires direct stdout writing
	fmt.Print(usageDetails())
}

func usageHeader() string {
	return `md-go-validator - Validate code blocks in Markdown and MDX files

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

`
}

func usageDetails() string {
	return `OUTPUT FORMATS:
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

SUPPORTED FILE TYPES:
    .md        Markdown
    .markdown  Markdown (alternative extension)
    .mdx       MDX (Markdown + JSX)

SKIP DIRECTIVES:
    Add these to skip validation of specific code blocks:
    <!-- skip-validate -->
    <!-- skip-md-validate -->
    <!-- md-skip -->
    <!-- no-validate -->
    // skip-validate
    //nolint

EXAMPLES:
    md-go-validator .                           # Validate all .md and .mdx files
    md-go-validator README.md                   # Validate a specific file
    md-go-validator -v .                        # Verbose output
    md-go-validator -f json .                 # JSON output for CI
    md-go-validator -l go,typescript .        # Validate Go and TypeScript
    md-go-validator -l templ,nix .            # Validate Templ and Nix
    md-go-validator --color never .             # Disable colors
    md-go-validator -o report.json -f json .  # Write JSON to file
    md-go-validator --timeout 30s .           # 30 second timeout
`
}
