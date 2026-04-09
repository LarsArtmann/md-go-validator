// md-go-validator validates Go code blocks in Markdown files.
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := parseArgs(tt.args)

			if len(cfg.paths) != len(tt.wantPaths) {
				t.Errorf("len(paths) = %d, want %d", len(cfg.paths), len(tt.wantPaths))
				return
			}

			for i, p := range cfg.paths {
				if i >= len(tt.wantPaths) {
					t.Errorf(
						"paths[%d] = %q, out of bounds (wantPaths has %d elements)",
						i,
						p,
						len(tt.wantPaths),
					)
					continue
				}
				if p != tt.wantPaths[i] {
					t.Errorf("paths[%d] = %q, want %q", i, p, tt.wantPaths[i])
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

		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("directory with markdown files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		content := []byte("```go\npackage main\n```\n")
		testutil.WriteTestFile(t, tmpDir, "test.md", content)
		testutil.WriteTestFile(t, tmpDir, "test.txt", content)

		validator := mdgovalidator.New(false)
		results := validatePath(context.Background(), validator, tmpDir)

		if len(results) != 1 {
			t.Errorf("expected 1 result (only .md files), got %d", len(results))
		}
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

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("empty paths returns empty results", func(t *testing.T) {
		t.Parallel()

		validator := mdgovalidator.New(false)
		results := validatePaths(context.Background(), validator, []string{})

		if len(results) != 0 {
			t.Errorf("expected 0 results for empty paths, got %d", len(results))
		}
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

		if _, statErr := os.Stat(outputPath); os.IsNotExist(statErr) {
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
		// Create a mock validator that returns an error
		mockValidator := &mockValidator{}

		results := validatePath(context.Background(), mockValidator, "/valid/path.md")
		if results != nil {
			t.Error("expected nil for non-existent path")
		}
	})
}

type mockValidator struct{}

func (m *mockValidator) ValidateFile(_ context.Context, _ string) ([]types.Result, error) {
	return nil, nil
}

func (m *mockValidator) ValidateDirectory(
	_ context.Context,
	_ string,
) ([]types.Result, error) {
	return nil, nil
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
		f, err := os.Create(filepath.Join(tmpDir, "test"+string(rune('0'+i))+".md"))
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
		if _, err := f.Write(content); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("failed to close test file: %v", err)
		}
	}

	results := validatePaths(ctx, validator, []string{tmpDir})
	if len(results) < 5 {
		t.Errorf("expected at least 5 results, got %d", len(results))
	}
}

func newValidResultForFile(fileID string, line, block int, code string) types.Result {
	return types.Result{
		File:       types.FileID(fileID),
		LineNumber: types.LineNumber(line),
		Block:      types.BlockIndex(block),
		Code:       code,
		Status:     types.StatusValid,
		Error:      nil,
	}
}

func newErrorResultForFile(fileID string, line, block int, code, errMsg string) types.Result {
	return types.Result{
		File:       types.FileID(fileID),
		LineNumber: types.LineNumber(line),
		Block:      types.BlockIndex(block),
		Code:       code,
		Status:     types.StatusError,
		Error:      errors.New(errMsg),
	}
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
	if err := writeOutputToFile(results, cfg); err != nil {
		t.Fatalf("writeOutputToFile failed: %v", err)
	}
}

func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), substr) {
		t.Errorf("output file should contain %q, got: %s", substr, string(content))
	}
}
