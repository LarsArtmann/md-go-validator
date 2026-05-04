package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
	"github.com/larsartmann/md-go-validator/pkg/languages"
	"github.com/larsartmann/md-go-validator/pkg/output"
	"github.com/larsartmann/md-go-validator/pkg/testutil"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

func runParseArgsFieldTest[T comparable](
	t *testing.T,
	name string,
	args []string,
	want T,
	get func(config) T,
) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		cfg := parseArgs(args)

		got := get(cfg)
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestParseArgsDefaults(t *testing.T) {
	t.Parallel()

	cfg := parseArgs([]string{})

	if cfg.verbose {
		t.Error("verbose should be false by default")
	}

	if !cfg.showCode {
		t.Error("showCode should be true by default")
	}

	if cfg.format != output.FormatTable {
		t.Errorf("format should be 'table' by default, got %q", cfg.format)
	}

	if cfg.colorMode != output.ColorModeAuto {
		t.Errorf("colorMode should be 'auto' by default, got %q", cfg.colorMode)
	}

	if len(cfg.paths) != 1 || cfg.paths[0] != "." {
		t.Errorf("paths should be ['.'], got %v", cfg.paths)
	}
}

func TestParseArgsVerboseFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"short form", []string{"-v", "."}},
		{"long form", []string{"--verbose", "."}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := parseArgs(tt.args)
			if !cfg.verbose {
				t.Error("verbose should be true")
			}
		})
	}
}

func TestParseArgsQuietFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"-q", []string{"-q", "."}},
		{"--quiet", []string{"--quiet", "."}},
		{"--no-code", []string{"--no-code", "."}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := parseArgs(tt.args)
			if cfg.showCode {
				t.Error("showCode should be false")
			}
		})
	}
}

func TestParseArgsPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      []string
		wantPaths []string
	}{
		{
			name:      "single file",
			args:      []string{"README.md"},
			wantPaths: []string{"README.md"},
		},
		{
			name:      "multiple paths",
			args:      []string{"docs/", "README.md", "CHANGELOG.md"},
			wantPaths: []string{"docs/", "README.md", "CHANGELOG.md"},
		},
		{
			name:      "flags with path",
			args:      []string{"-v", "-q", "test.md"},
			wantPaths: []string{"test.md"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := parseArgs(testCase.args)

			if len(cfg.paths) != len(testCase.wantPaths) {
				t.Errorf("len(paths) = %d, want %d", len(cfg.paths), len(testCase.wantPaths))

				return
			}

			for idx, pathEntry := range cfg.paths {
				if idx >= len(testCase.wantPaths) {
					t.Errorf(
						"paths[%d] = %q, out of bounds (wantPaths has %d elements)",
						idx,
						pathEntry,
						len(testCase.wantPaths),
					)

					continue
				}

				if pathEntry != testCase.wantPaths[idx] {
					t.Errorf("paths[%d] = %q, want %q", idx, pathEntry, testCase.wantPaths[idx])
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	t.Parallel()

	t.Run("non-existent path returns nil", func(t *testing.T) {
		t.Parallel()

		validator := mdgovalidator.New(false)
		results := validatePath(
			context.Background(),
			validator,
			"/nonexistent/path/that/does/not/exist",
		)

		if results != nil {
			t.Errorf("expected nil results for non-existent path, got %v", results)
		}
	})

	t.Run("valid markdown file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		content := []byte("```go\npackage main\n```\n")
		tmpFile := testutil.WriteTestFile(t, tmpDir, "test.md", content)

		validator := mdgovalidator.New(false)
		results := validatePath(context.Background(), validator, tmpFile)

		testutil.AssertResultCount(t, results, 1)
	})

	t.Run("directory with markdown files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		content := []byte("```go\npackage main\n```\n")
		testutil.WriteTestFile(t, tmpDir, "test.md", content)
		testutil.WriteTestFile(t, tmpDir, "test.txt", content)

		validator := mdgovalidator.New(false)
		results := validatePath(context.Background(), validator, tmpDir)

		testutil.AssertResultCount(t, results, 1)
	})

	t.Run("directory with MDX files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		content := []byte("```go\npackage main\n```\n")
		testutil.WriteTestFile(t, tmpDir, "test.md", content)
		testutil.WriteTestFile(t, tmpDir, "doc.mdx", content)
		testutil.WriteTestFile(t, tmpDir, "test.txt", content)

		validator := mdgovalidator.New(false)
		results := validatePath(context.Background(), validator, tmpDir)

		testutil.AssertResultCount(t, results, 2)
	})
}

func TestValidatePaths(t *testing.T) {
	t.Parallel()

	t.Run("multiple paths", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		content := []byte("```go\npackage main\n```\n")
		file1 := testutil.WriteTestFile(t, tmpDir, "file1.md", content)
		file2 := testutil.WriteTestFile(t, tmpDir, "file2.md", content)

		validator := mdgovalidator.New(false)
		results := validatePaths(context.Background(), validator, []string{file1, file2})

		testutil.AssertResultCount(t, results, 2)
	})

	t.Run("empty paths returns empty results", func(t *testing.T) {
		t.Parallel()

		validator := mdgovalidator.New(false)
		results := validatePaths(context.Background(), validator, []string{})

		testutil.AssertResultCount(t, results, 0)
	})
}

func TestPrintUsage(t *testing.T) {
	t.Parallel()

	// Just verify it doesn't panic
	printUsage()
}

func TestParseArgsFormatFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantFormat output.Format
	}{
		{"json short", []string{"-f", "json", "."}, output.FormatJSON},
		{"json long", []string{"--format", "json", "."}, output.FormatJSON},
		{"table", []string{"-f", "table", "."}, output.FormatTable},
		{"markdown", []string{"-f", "markdown", "."}, output.FormatMarkdown},
		{"markdown alias md", []string{"-f", "md", "."}, output.FormatMarkdown},
		{"yaml", []string{"-f", "yaml", "."}, output.FormatYAML},
		{"yaml alias yml", []string{"-f", "yml", "."}, output.FormatYAML},
		{"csv", []string{"-f", "csv", "."}, output.FormatCSV},
		{"quiet", []string{"-f", "quiet", "."}, output.FormatQuiet},
		{"quiet alias q", []string{"-f", "q", "."}, output.FormatQuiet},
	}

	for _, tt := range tests {
		runParseArgsFieldTest(
			t,
			tt.name,
			tt.args,
			tt.wantFormat,
			func(cfg config) output.Format { return cfg.format },
		)
	}
}

func TestParseArgsColorFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		wantColorMode output.ColorMode
	}{
		{"always", []string{"--color", "always", "."}, output.ColorModeAlways},
		{"never", []string{"--color", "never", "."}, output.ColorModeNever},
		{"auto", []string{"--color", "auto", "."}, output.ColorModeAuto},
	}

	for _, tt := range tests {
		runParseArgsFieldTest(
			t,
			tt.name,
			tt.args,
			tt.wantColorMode,
			func(cfg config) output.ColorMode { return cfg.colorMode },
		)
	}
}

func TestParseArgsOutputFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           []string
		wantOutputFile string
	}{
		{"short form", []string{"-o", "report.json", "."}, "report.json"},
		{"long form", []string{"--output", "output/report.yaml", "."}, "output/report.yaml"},
		{"with path", []string{"-o", "/tmp/report.md", "README.md"}, "/tmp/report.md"},
	}

	for _, tt := range tests {
		runParseArgsFieldTest(
			t,
			tt.name,
			tt.args,
			tt.wantOutputFile,
			func(cfg config) string { return cfg.outputFile },
		)
	}
}

func TestWriteOutputToFile(t *testing.T) {
	t.Parallel()

	t.Run("creates parent directories", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "subdir", "nested", "report.json")

		results := []types.Result{
			newValidResultForFile("test.md", 1, 1, "package main"),
		}

		cfg := newTestConfig(outputPath, output.FormatJSON)
		assertWriteOutputToFile(t, results, cfg)

		_, statErr := os.Stat(outputPath)
		if os.IsNotExist(statErr) {
			t.Error("output file was not created")
		}
	})

	t.Run("writes JSON content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "report.json")

		results := []types.Result{
			newErrorResultForFile("test.md", 10, 1, "bad code", "syntax error"),
		}

		cfg := newTestConfig(outputPath, output.FormatJSON)
		assertWriteOutputToFile(t, results, cfg)
		assertFileContains(t, outputPath, "test.md")
	})

	t.Run("writes CSV content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "report.csv")

		results := []types.Result{
			newValidResultForFile("test.md", 5, 1, "package main"),
		}

		cfg := newTestConfig(outputPath, output.FormatCSV)
		assertWriteOutputToFile(t, results, cfg)
		assertFileContains(t, outputPath, "test.md")
	})
}

func TestValidatePathWithErrors(t *testing.T) {
	t.Parallel()

	t.Run("path resolution error", func(t *testing.T) {
		t.Parallel()

		validator := mdgovalidator.New(false)

		results := validatePath(context.Background(), validator, "/valid/path.md")
		if results != nil {
			t.Error("expected nil for non-existent path")
		}
	})
}

func TestValidatePathsCapacity(t *testing.T) {
	t.Parallel()

	// Test that we pre-allocate correctly
	validator := mdgovalidator.New(false)
	ctx := context.Background()

	// Create files with multiple code blocks each
	tmpDir := t.TempDir()

	for i := range 5 {
		content := []byte("```go\npackage main\n```\n```go\npackage main\n```\n")

		//nolint:gosec // G304: Controlled test data in temp directory
		testFile, err := os.Create(filepath.Join(tmpDir, "test"+string(rune('0'+i))+".md"))
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		_, writeErr := testFile.Write(content)
		if writeErr != nil {
			t.Fatalf("failed to write test file: %v", writeErr)
		}

		closeErr := testFile.Close()
		if closeErr != nil {
			t.Fatalf("failed to close test file: %v", closeErr)
		}
	}

	results := validatePaths(ctx, validator, []string{tmpDir})
	testutil.AssertMinResults(t, results, 5)
}

func newValidResultForFile(fileID string, line, block int, code string) types.Result {
	return types.NewResultWithStatus(
		types.NewFileID(fileID),
		types.NewLineNumber(line),
		types.NewBlockIndex(block),
		code,
		types.StatusValid,
	)
}

func newErrorResultForFile(fileID string, line, block int, code, errMsg string) types.Result {
	return types.NewErrorResult(
		types.NewFileID(fileID),
		types.NewLineNumber(line),
		types.NewBlockIndex(block),
		code,
		errors.New(errMsg), //nolint:err113 // Test helper - errMsg is controlled test data
	)
}

func newTestConfig(outputFile string, format output.Format) config {
	return config{
		verbose:    false,
		showCode:   true,
		format:     format,
		colorMode:  output.ColorModeNever,
		outputFile: outputFile,
		paths:      nil,
		timeout:    0,
		contextCfg: mdgovalidator.DefaultContextConfig(),
		languages:  []languages.Language{languages.LangGo},
	}
}

func assertWriteOutputToFile(t *testing.T, results []types.Result, cfg config) {
	t.Helper()

	err := writeOutputToFile(results, cfg)
	if err != nil {
		t.Fatalf("writeOutputToFile failed: %v", err)
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()

	//nolint:gosec // G304: Controlled test data path
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), substr) {
		t.Errorf("output file should contain %q, got: %s", substr, string(content))
	}
}

func TestParseArgsTimeoutFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantTimeout time.Duration
	}{
		{"short form", []string{"-t", "30s", "."}, 30 * time.Second},
		{"long form", []string{"--timeout", "5m", "."}, 5 * time.Minute},
	}

	for _, tt := range tests {
		runParseArgsFieldTest(
			t,
			tt.name,
			tt.args,
			tt.wantTimeout,
			func(cfg config) time.Duration { return cfg.timeout },
		)
	}
}

func TestParseArgsLanguageFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		wantLanguage languages.Language
	}{
		{"short go", []string{"-l", "go", "."}, languages.LangGo},
		{"long typescript", []string{"--language", "typescript", "."}, languages.LangTypeScript},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := parseArgs(tt.args)
			if len(cfg.languages) == 0 {
				t.Fatal("expected at least one language")
			}

			if cfg.languages[0] != tt.wantLanguage {
				t.Errorf("got %s, want %s", cfg.languages[0], tt.wantLanguage)
			}
		})
	}
}

func TestWriteOutputToFile_AllFormats(t *testing.T) {
	t.Parallel()

	formats := []struct {
		name   string
		format output.Format
	}{
		{"table", output.FormatTable},
		{"json", output.FormatJSON},
		{"markdown", output.FormatMarkdown},
		{"yaml", output.FormatYAML},
		{"csv", output.FormatCSV},
		{"quiet", output.FormatQuiet},
	}

	results := []types.Result{
		newValidResultForFile("test.md", 1, 1, "package main"),
	}

	for _, tt := range formats {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "report."+tt.name)
			cfg := newTestConfig(outputPath, tt.format)

			err := writeOutputToFile(results, cfg)
			if err != nil {
				t.Errorf("format %s: writeOutputToFile error: %v", tt.name, err)
			}

			_, statErr := os.Stat(outputPath)
			if statErr != nil {
				t.Errorf("format %s: output file not created: %v", tt.name, statErr)
			}
		})
	}
}

func TestUsageHeader(t *testing.T) {
	t.Parallel()

	h := usageHeader()
	if !strings.Contains(h, "md-go-validator") {
		t.Error("expected usage header to contain program name")
	}

	if !strings.Contains(h, "OPTIONS") {
		t.Error("expected usage header to contain OPTIONS")
	}
}

func TestUsageDetails(t *testing.T) {
	t.Parallel()

	details := usageDetails()

	sections := []string{
		"OUTPUT FORMATS",
		"COLOR MODES",
		"SUPPORTED LANGUAGES",
		"SUPPORTED FILE TYPES",
		"SKIP DIRECTIVES",
		"EXAMPLES",
	}
	for _, section := range sections {
		if !strings.Contains(details, section) {
			t.Errorf("expected usage details to contain %q", section)
		}
	}
}

func TestParseLanguagesDirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{"single go", "go", 1, false},
		{"multiple", "go,typescript,nix", 3, false},
		{"with spaces", "go , typescript", 2, false},
		{"unsupported", "python", 0, true},
		{"mixed valid invalid", "go,python", 0, true},
		{"rust alias", "rs", 1, false},
		{"golang", "golang", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			langs, err := parseLanguages(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(langs) != tt.wantLen {
				t.Errorf("expected %d languages, got %d", tt.wantLen, len(langs))
			}
		})
	}
}

func TestParseArgsCombinedFlags(t *testing.T) {
	t.Parallel()

	args := []string{
		"-v", "-f", "json",
		"--color", "never",
		"-o", "out.json",
		"-l", "go",
		"src/",
	}

	cfg := parseArgs(args)

	if !cfg.verbose {
		t.Error("expected verbose=true")
	}

	if cfg.format != output.FormatJSON {
		t.Errorf("expected format=json, got %s", cfg.format)
	}

	if cfg.colorMode != output.ColorModeNever {
		t.Errorf("expected colorMode=never, got %s", cfg.colorMode)
	}

	if cfg.outputFile != "out.json" {
		t.Errorf("expected outputFile=out.json, got %s", cfg.outputFile)
	}

	if len(cfg.languages) != 1 || cfg.languages[0] != languages.LangGo {
		t.Errorf("expected languages=[go], got %v", cfg.languages)
	}

	if len(cfg.paths) != 1 || cfg.paths[0] != "src/" {
		t.Errorf("expected paths=[src/], got %v", cfg.paths)
	}
}
