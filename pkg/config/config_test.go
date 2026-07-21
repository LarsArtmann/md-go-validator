package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/md-go-validator/pkg/languages"
)

var errTestWrite = errors.New("test write error")

func TestDefault(t *testing.T) {
	t.Parallel()

	cfg := Default()

	if cfg.Format != defaultFormat {
		t.Errorf("expected format %q, got %q", defaultFormat, cfg.Format)
	}

	if len(cfg.Languages) != 1 || cfg.Languages[0] != "go" {
		t.Errorf("expected languages ['go'], got %v", cfg.Languages)
	}
}

func TestLoad_YAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, defaultConfigFileName)

	content := `languages:
  - go
  - typescript
exclude:
  - "vendor/*"
  - "docs/generated/*"
skipDirectives:
  - "<!-- sketch -->"
format: json
`

	err := os.WriteFile(path, []byte(content), configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(cfg.Languages))
	}

	if len(cfg.Exclude) != 2 {
		t.Errorf("expected 2 excludes, got %d", len(cfg.Exclude))
	}

	if len(cfg.SkipDirectives) != 1 {
		t.Errorf("expected 1 skip directive, got %d", len(cfg.SkipDirectives))
	}

	if cfg.Format != "json" {
		t.Errorf("expected format 'json', got %q", cfg.Format)
	}
}

func TestLoad_JSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, ".md-go-validator.json")

	content := `{"languages":["go","rust"],"format":"yaml"}`

	err := os.WriteFile(path, []byte(content), configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(cfg.Languages))
	}

	if cfg.Format != "yaml" {
		t.Errorf("expected format 'yaml', got %q", cfg.Format)
	}
}

func TestLoadFromDir_Found(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, defaultConfigFileName)

	err := os.WriteFile(path, []byte("format: quiet\n"), configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Format != "quiet" {
		t.Errorf("expected format 'quiet', got %q", cfg.Format)
	}
}

func TestLoadFromDir_NotFound(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	_, err := LoadFromDir(dir)
	if !errors.Is(err, ErrNotFound) { //nolint:legacyerrors // value sentinel
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	cfg := Config{
		Languages:      []languages.Language{languages.LangGo, languages.LangRust},
		Exclude:        nil,
		SkipDirectives: nil,
		Format:         "json",
	}

	err := Save(path, cfg)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(loaded.Languages))
	}

	if loaded.Format != "json" {
		t.Errorf("expected format 'json', got %q", loaded.Format)
	}
}

func TestInitFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, defaultConfigFileName)

	err := InitFile(path)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // test file path
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty config file")
	}

	// Verify the file is loadable.
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Format != defaultFormat {
		t.Errorf("expected format %q, got %q", defaultFormat, cfg.Format)
	}
}

func TestLoad_UnsupportedFormat(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.txt")

	err := os.WriteFile(path, []byte("invalid"), configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if !errors.Is(err, ErrUnsupportedFormat) { //nolint:legacyerrors // value sentinel
		t.Errorf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	malformed := []byte("languages: [go\n  broken: {{{")

	err := os.WriteFile(path, malformed, configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Error("expected error for malformed YAML, got nil")
	}
}

func TestLoad_InvalidLanguage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "badlang.yaml")
	content := []byte("languages:\n  - python\n")

	err := os.WriteFile(path, content, configFilePerms)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(path)
	if err == nil {
		t.Error("expected error for unsupported language in config, got nil")
	}
}
