// md-go-validator validates Go code blocks in Markdown files.
package main

import (
	"os"
	"path/filepath"
	"testing"

	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
)

func TestParseArgsDefaults(t *testing.T) {
	t.Parallel()

	cfg := parseArgs([]string{})

	if cfg.verbose {
		t.Error("verbose should be false by default")
	}
	if !cfg.showCode {
		t.Error("showCode should be true by default")
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
		results := validatePath(validator, "/nonexistent/path/that/does/not/exist")

		if results != nil {
			t.Errorf("expected nil results for non-existent path, got %v", results)
		}
	})

	t.Run("valid markdown file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.md")

		content := []byte("```go\npackage main\n```\n")
		if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
			t.Fatal(err)
		}

		validator := mdgovalidator.New(false)
		results := validatePath(validator, tmpFile)

		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("directory with markdown files", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.md")

		content := []byte("```go\npackage main\n```\n")
		if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
			t.Fatal(err)
		}

		// Non-markdown file should be ignored
		txtFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(txtFile, content, 0o600); err != nil {
			t.Fatal(err)
		}

		validator := mdgovalidator.New(false)
		results := validatePath(validator, tmpDir)

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

		// Create two markdown files
		file1 := filepath.Join(tmpDir, "file1.md")
		file2 := filepath.Join(tmpDir, "file2.md")

		content := []byte("```go\npackage main\n```\n")
		if err := os.WriteFile(file1, content, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file2, content, 0o600); err != nil {
			t.Fatal(err)
		}

		validator := mdgovalidator.New(false)
		results := validatePaths(validator, []string{file1, file2})

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
	})

	t.Run("empty paths returns empty results", func(t *testing.T) {
		t.Parallel()

		validator := mdgovalidator.New(false)
		results := validatePaths(validator, []string{})

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
