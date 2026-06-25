// md-go-validator validates code blocks in Markdown and MDX files.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
	"github.com/larsartmann/md-go-validator/pkg/baseline"
	cfgpkg "github.com/larsartmann/md-go-validator/pkg/config"
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

// Exit codes follow Unix linting-tool convention:
// 0 = success, 1 = validation errors found, 2 = tool/usage errors.
const (
	exitSuccess       = 0
	exitValidationErr = 1
	exitToolErr       = 2
)

// Flag name constants.
const (
	flagQuiet   = "--quiet"
	flagNoCode  = "--no-code"
	flagColor   = "--color"
	flagVerbose = "--verbose"
	flagFormat  = "--format"
	flagOutput  = "--output"
	flagTimeout = "--timeout"
	flagVersion = "--version"
	flagExclude = "--exclude"
	flagSkipDir = "--skip-directive"
	flagInit    = "--init"
)

// version is set at build time via ldflags.
//

var version = "dev"

// osExit allows mocking os.Exit in tests.
//
//nolint:gochecknoglobals // Required for testing os.Exit behavior
var osExit = os.Exit

type config struct {
	verbose        bool
	showCode       bool
	format         output.Format
	colorMode      output.ColorMode
	outputFile     string
	paths          []string
	timeout        time.Duration
	contextCfg     mdgovalidator.ContextConfig
	languages      []languages.Language
	exclude        []string
	skipDirectives []string
	initConfig     bool
	baselineFile   string
	listLangs      bool
	failOnSkipped  bool
}

func main() {
	cfg := parseArgs(os.Args[1:])

	os.Exit(runWithConfig(cfg))
}

// runWithConfig executes the validation pipeline and returns the exit code.
// Extracted from main for testability (no os.Exit calls here).
func runWithConfig(cfg config) int {
	if exitCode, handled := handleEarlyExit(cfg); handled {
		return exitCode
	}

	validator := mdgovalidator.New(cfg.verbose).
		WithLanguages(cfg.languages).
		WithExcludePatterns(types.NewExcludePatterns(cfg.exclude)).
		WithSkipDirectives(cfg.skipDirectives)

	ctx, cancel := cfg.contextCfg.Build()
	defer cancel()

	allResults, hadToolError := validatePaths(ctx, validator, cfg.paths)

	// Apply baseline filter if set.
	if cfg.baselineFile != "" {
		baseSet, err := baseline.Load(cfg.baselineFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading baseline: %v\n", err)

			return exitToolErr
		}

		filtered := baseSet.FilterNew(allResults)
		suppressed := len(allResults) - len(filtered)
		allResults = filtered

		if suppressed > 0 {
			fmt.Fprintf(os.Stderr, "Baseline: suppressed %d known errors from %s\n", suppressed, cfg.baselineFile)
		}
	}

	if cfg.outputFile != "" {
		err := writeOutputToFile(allResults, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)

			return exitToolErr
		}
	} else {
		output.PrintReport(allResults, cfg.format, cfg.colorMode, cfg.showCode)
	}

	if hadToolError {
		return exitToolErr
	}

	if mdgovalidator.HasErrors(allResults) {
		return exitValidationErr
	}

	if cfg.failOnSkipped && mdgovalidator.HasSkipped(allResults) {
		return exitValidationErr
	}

	return exitSuccess
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
	return mergeHandlers(
		coreHandlers(),
		appendableHandlers(),
		exitHandlers(),
	)
}

// coreHandlers returns handlers for the primary CLI flags.
func coreHandlers() map[string]argHandler {
	formatHandler := singleValueArgHandler(
		"format", output.ParseFormat,
		func(c *config, formatVal output.Format) { c.format = formatVal },
	)
	colorHandler := singleValueArgHandler(
		"color", output.ParseColorMode,
		func(c *config, cm output.ColorMode) { c.colorMode = cm },
	)
	outputHandler := stringArgHandler("output", func(c *config, s string) { c.outputFile = s })
	timeoutHandler := singleValueArgHandler(
		"timeout", time.ParseDuration,
		func(c *config, d time.Duration) {
			c.timeout = d
			c.contextCfg = c.contextCfg.WithTimeout(d)
		},
	)
	languagesHandler := languagesArgHandler()
	baselineHandler := stringArgHandler("baseline", func(c *config, s string) { c.baselineFile = s })
	configFileHandler := stringArgHandler("config", func(c *config, s string) {})

	return map[string]argHandler{
		"-v":         boolFlagHandler(func(c *config) { c.verbose = true }),
		flagVerbose:  boolFlagHandler(func(c *config) { c.verbose = true }),
		"-q":         boolFlagHandler(func(c *config) { c.format = output.FormatQuiet; c.showCode = false }),
		flagQuiet:    boolFlagHandler(func(c *config) { c.format = output.FormatQuiet; c.showCode = false }),
		flagNoCode:   boolFlagHandler(func(c *config) { c.showCode = false }),
		"-f":         formatHandler,
		flagFormat:   formatHandler,
		flagColor:    colorHandler,
		"-o":         outputHandler,
		flagOutput:   outputHandler,
		"-t":         timeoutHandler,
		flagTimeout:  timeoutHandler,
		"-l":         languagesHandler,
		"--language": languagesHandler,
		"--baseline": baselineHandler,
		"--config":   configFileHandler,
	}
}

// appendableHandlers returns handlers for repeatable flags.
func appendableHandlers() map[string]argHandler {
	return map[string]argHandler{
		flagExclude: appendArgHandler("exclude",
			func(c *config, s string) { c.exclude = append(c.exclude, s) }),
		flagSkipDir: appendArgHandler("skip-directive",
			func(c *config, s string) { c.skipDirectives = append(c.skipDirectives, s) }),
	}
}

// exitHandlers returns handlers for flags that exit immediately.
func exitHandlers() map[string]argHandler {
	return map[string]argHandler{
		flagInit:            boolFlagHandler(func(c *config) { c.initConfig = true }),
		"--list-languages":  boolFlagHandler(func(c *config) { c.listLangs = true }),
		"--fail-on-skipped": boolFlagHandler(func(c *config) { c.failOnSkipped = true }),
		"-h":                handleHelp,
		"--help":            handleHelp,
		"-V":                handleVersion,
		flagVersion:         handleVersion,
	}
}

// mergeHandlers combines multiple handler maps into one.
func mergeHandlers(handlerMaps ...map[string]argHandler) map[string]argHandler {
	result := make(map[string]argHandler)

	for _, handlerMap := range handlerMaps {
		maps.Copy(result, handlerMap)
	}

	return result
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

// appendArgHandler creates a handler for repeatable string flags that append
// to a slice in the config.
func appendArgHandler(flagName string, setter func(*config, string)) argHandler {
	return func(args []string, idx int, cfg *config) (int, bool) {
		if !requireArg(args, idx, flagName) {
			return 0, false
		}

		setter(cfg, args[idx+1])

		return 1, true
	}
}

// applyConfigFile merges config file values into the CLI config struct.
// CLI flags are applied later and override these values.
func applyConfigFile(cfg *config, fileCfg cfgpkg.Config, cfgErr error) {
	if cfgErr != nil {
		if !errors.Is(cfgErr, cfgpkg.ErrNotFound) {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file: %v\n", cfgErr)
		}

		return
	}

	if len(fileCfg.Languages) > 0 {
		cfg.languages = fileCfg.Languages
	}

	// exclude and skipDirectives are applied AFTER CLI parsing (in parseArgs)
	// so CLI flags override — not union with — config file values.
	applyConfigFormat(cfg, fileCfg.Format)
}

// applyConfigFormat sets the output format from the config file if valid.
func applyConfigFormat(cfg *config, format string) {
	if format == "" {
		return
	}

	parsedFormat, err := output.ParseFormat(format)
	if err == nil {
		cfg.format = parsedFormat
	}
}

// handleEarlyExit processes flags that exit before validation (--init, --list-languages).
// Returns (exitCode, true) if the flag was handled.
func handleEarlyExit(cfg config) (int, bool) {
	if cfg.initConfig {
		err := cfgpkg.InitFile(cfgpkg.DefaultConfigFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config file: %v\n", err)

			return exitToolErr, true
		}

		//nolint:forbidigo // CLI success output requires direct stdout writing
		fmt.Printf("Created %s with default configuration\n", cfgpkg.DefaultConfigFileName)

		return exitSuccess, true
	}

	if cfg.listLangs {
		printSupportedLanguages()

		return exitSuccess, true
	}

	return 0, false
}

// printSupportedLanguages prints all supported languages to stdout.
func printSupportedLanguages() {
	//nolint:forbidigo // CLI output requires direct stdout writing
	for _, lang := range languages.AllLanguages() {
		fmt.Printf("%s (%s)\n", lang, strings.Join(lang.Extensions(), ", "))
	}
}

func parseArgs(args []string) config {
	// Pre-scan for --config flag so we know which file to load before
	// the main parsing loop applies config-file defaults.
	configPath := findFlagValue(args, "--config")

	var fileCfg cfgpkg.Config

	var cfgErr error
	if configPath != "" {
		fileCfg, cfgErr = cfgpkg.Load(configPath)
	} else {
		fileCfg, cfgErr = cfgpkg.LoadFromDir(".")
	}

	cfg := config{ //nolint:exhaustruct // fields set by CLI flags later
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

	// Apply config file values as defaults (CLI flags override).
	applyConfigFile(&cfg, fileCfg, cfgErr)

	argHandlers := newArgHandlers()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if handler, ok := argHandlers[arg]; ok {
			advance, ok := handler(args, i, &cfg)
			if !ok {
				os.Exit(exitToolErr)
			}

			i += advance

			continue
		}

		if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
			printUsage()
			os.Exit(exitToolErr)
		}

		cfg.paths = append(cfg.paths, arg)
	}

	if len(cfg.paths) == 0 {
		cfg.paths = []string{"."}
	}

	applyConfigRepeatable(&cfg, fileCfg, cfgErr)

	return cfg
}

// applyConfigRepeatable applies config-file repeatable values (exclude,
// skipDirectives) only if the CLI flags didn't set them. This ensures
// CLI flags override — not union with — config file values.
func applyConfigRepeatable(cfg *config, fileCfg cfgpkg.Config, cfgErr error) {
	if cfgErr != nil {
		return
	}

	if len(cfg.exclude) == 0 {
		cfg.exclude = fileCfg.Exclude
	}

	if len(cfg.skipDirectives) == 0 {
		cfg.skipDirectives = fileCfg.SkipDirectives
	}
}

func returnParseError(_ string, err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	printUsage()
}

// findFlagValue scans args for a --flag value pair and returns the value.
// Returns "" if the flag is not found or has no following argument.
func findFlagValue(args []string, flagName string) string {
	for i, arg := range args {
		if arg == flagName && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

// exitingHandler builds an argHandler that runs action() and then terminates
// the process via osExit. handled reports whether the handler performed work
// beyond exiting (true for handleVersion which prints output; false for
// handleHelp which just exits).
func exitingHandler(action func(), handled bool) argHandler {
	return func(_ []string, _ int, _ *config) (int, bool) {
		action()
		osExit(0)

		return 0, handled
	}
}

// handleHelp and handleVersion are package-level handlers. They are wired
// into the arg-dispatch table, not invoked from a single function scope.
//
//nolint:gochecknoglobals // wired into arg-dispatch table
var (
	handleHelp    = exitingHandler(printUsage, false)
	handleVersion = exitingHandler(func() {
		//nolint:forbidigo // CLI version output requires direct stdout writing
		fmt.Println("md-go-validator", version)
	}, true)
)

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
	return singleValueArgHandler(flagName, parser, setter)
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
) ([]types.Result, bool) {
	allResults := make([]types.Result, 0, len(paths)*resultsCapacityMultiplier)

	var hadToolError bool

	for _, path := range paths {
		results, ok := validatePath(ctx, validator, path)
		if !ok {
			hadToolError = true
		}

		allResults = append(allResults, results...)
	}

	return allResults, hadToolError
}

func validatePath(
	ctx context.Context,
	validator *mdgovalidator.FileValidator,
	path string,
) ([]types.Result, bool) {
	if path == "-" {
		return validateStdin(ctx, validator)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path %s: %v\n", path, err)

		return nil, false
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Path %s does not exist\n", absPath)

		return nil, false
	}

	var results []types.Result
	if info.IsDir() {
		results, err = validator.ValidateDirectory(ctx, absPath)
	} else {
		results, err = validator.ValidateFile(ctx, absPath)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", absPath, err)

		return nil, false
	}

	return results, true
}

// stdinSourceName is the file identifier used for stdin-validated results.
const stdinSourceName = "<stdin>"

// reportStdinError prints a stdin-pipeline error to stderr and signals failure
// to the caller via a (nil, false) tuple. Centralised so both steps share
// identical behaviour.
func reportStdinError(stage string, err error) ([]types.Result, bool) {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", stage, err)

	return nil, false
}

func validateStdin(
	ctx context.Context,
	validator *mdgovalidator.FileValidator,
) ([]types.Result, bool) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return reportStdinError("reading stdin", err)
	}

	results, err := validator.ValidateContent(ctx, string(content), stdinSourceName)
	if err != nil {
		return reportStdinError("validating stdin", err)
	}

	return results, true
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
    md-go-validator [OPTIONS] -        # Read markdown from stdin

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
    --exclude         Glob pattern to exclude (repeatable, e.g. --exclude vendor/*)
    --skip-directive  Custom skip directive (repeatable, e.g. --skip-directive "// example")
    --init            Create a default .md-go-validator.yaml config file
    --baseline        Baseline file of known errors (file:line per line); only new errors fail
    --list-languages  Print all supported languages and exit
    --fail-on-skipped  Exit non-zero if any blocks were skipped (strict mode)
    -h, --help        Show this help message
    -V, --version     Show version information

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
    go          Go (built-in stdlib parser, always available)
    templ       Templ (embedded tree-sitter grammar)
    typescript  TypeScript (embedded tree-sitter grammar)
    tsx         TypeScript JSX (embedded tree-sitter grammar)
    nix         Nix (embedded tree-sitter grammar)
    rust        Rust (embedded tree-sitter grammar)
    hcl         HCL/Terraform (embedded tree-sitter grammar)
    terraform   Alias for HCL (embedded tree-sitter grammar)

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
    cat README.md | md-go-validator -          # Validate markdown from stdin
`
}
