// Package config provides configuration file support for md-go-validator.
// Configuration is loaded from .md-go-validator.yaml (or .json) in the
// working directory. CLI flags override config file values.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	yaml "github.com/go-faster/yaml"
)

const (
	defaultConfigFileName = ".md-go-validator.yaml"
	defaultFormat         = "table"
	configFilePerms       = 0o600
)

// DefaultConfigFileName is the default config file name.
const DefaultConfigFileName = defaultConfigFileName

// ErrNotFound is returned when no config file is found.
var ErrNotFound = errors.New("no config file found")

// ErrUnsupportedFormat is returned for unknown config file extensions.
var ErrUnsupportedFormat = errors.New("unsupported config format")

// Config holds project-level configuration for md-go-validator.
type Config struct {
	// Languages specifies which languages to validate (e.g. ["go", "typescript"]).
	// Empty means "go" only (the default).
	Languages []string `json:"languages" yaml:"languages"`
	// Exclude specifies glob patterns to exclude from validation.
	// Example: ["vendor/*", "docs/generated/*"].
	Exclude []string `json:"exclude" yaml:"exclude"`
	// SkipDirectives specifies custom skip directive strings.
	// Blocks containing these directives are skipped.
	// Example: ["<!-- sketch -->", "// pseudo-code"].
	SkipDirectives []string `json:"skipDirectives" yaml:"skipDirectives"`
	// Format specifies the output format (table, json, yaml, csv, markdown, quiet).
	// Empty means "table" (the default).
	Format string `json:"format" yaml:"format"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		Languages:      []string{"go"},
		Exclude:        nil,
		SkipDirectives: nil,
		Format:         defaultFormat,
	}
}

// Load reads configuration from the given path. Supports YAML and JSON
// based on file extension.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is user-controlled config file
	if err != nil {
		return Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := Config{} //nolint:exhaustruct // zero-value fields are intentional for YAML unmarshaling

	switch ext := filepath.Ext(path); ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &cfg)
		if err != nil {
			return Config{}, fmt.Errorf("parse YAML config %s: %w", path, err)
		}
	case ".json":
		err = yaml.Unmarshal(data, &cfg)
		if err != nil {
			return Config{}, fmt.Errorf("parse JSON config %s: %w", path, err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported config format %s: %w", ext, ErrUnsupportedFormat)
	}

	return cfg, nil
}

// LoadFromDir searches for a config file in the given directory and loads it.
// Returns ErrNotFound if no config file exists.
func LoadFromDir(dir string) (Config, error) {
	candidates := []string{
		defaultConfigFileName,
		".md-go-validator.yml",
		".md-go-validator.json",
	}

	for _, name := range candidates {
		path := filepath.Join(dir, name)

		_, err := os.Stat(path)
		if err == nil {
			return Load(path)
		}
	}

	return Config{}, ErrNotFound
}

// Save writes the configuration to the given path in YAML format.
func Save(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(path, data, configFilePerms)
	if err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}

	return nil
}

// InitFile writes a default config file with commented examples to the given path.
func InitFile(path string) error {
	content := `# md-go-validator configuration
# See: https://github.com/LarsArtmann/md-go-validator

# Languages to validate (default: go)
languages:
  - go
  # - typescript
  # - rust
  # - nix
  # - hcl
  # - templ

# Glob patterns to exclude from validation
exclude:
  - "vendor/*"
  - "node_modules/*"
  - "docs/generated/*"
  # - "docs/research/*"

# Custom skip directives — blocks containing these are skipped
# skipDirectives:
#   - "<!-- sketch -->"
#   - "// pseudo-code"

# Output format: table, json, yaml, csv, markdown, quiet
format: table
`

	err := os.WriteFile(path, []byte(content), configFilePerms)
	if err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}

	return nil
}
