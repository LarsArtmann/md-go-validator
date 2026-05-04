# md-go-validator

A CLI tool and Go library that validates code blocks in Markdown and MDX files.

Ensures code examples in your documentation are syntactically valid. Uses multiple parsing strategies to handle partial code snippets (imports, type declarations, function bodies) commonly found in technical documentation.

Supports `.md`, `.markdown`, and `.mdx` files.

- **CLI tool** - Validate entire documentation repositories
- **Go library** - Integrate validation into your own tools
- **CI-friendly** - Exit code 1 on validation errors
- **Multi-language** - Validate Go, TypeScript, Rust, Nix, HCL/Terraform, and Templ
- **Pure Go** - Uses gotreesitter (pure Go tree-sitter), no CGO required
- **Cross-platform** - Compiles to any GOOS/GOARCH
- **Pluggable** - Easy to add new language validators

## Features

- **Multi-Language Support** (all built-in, no external tools needed):
  - **Go** - built-in Go parser
  - **TypeScript/TSX** - tree-sitter parser
  - **Rust** - tree-sitter parser
  - **Nix** - tree-sitter parser
  - **HCL/Terraform** - tree-sitter parser
  - **Templ** - tree-sitter parser
- **Pure Go** - Uses [gotreesitter](https://github.com/odvcencio/gotreesitter) for syntax parsing
- **Multiple Parsing Strategies** - Handles partial code:
  - Complete files
  - Package declarations only
  - Function bodies
  - Expressions
  - Statements
- **Skip Directives** - Mark intentionally incomplete code
- **Recursive Scanning** - Validates entire documentation trees
- **CI-Friendly** - Exit code 1 on errors, structured output

## Installation

```bash
go install github.com/larsartmann/md-go-validator@latest
```

## Usage

```bash
# Validate all Go code blocks in current directory (default)
md-go-validator .

# Validate specific file
md-go-validator README.md

# Validate multiple paths
md-go-validator docs/ README.md

# Validate multiple languages
md-go-validator -l go,typescript,rust .

# Verbose output (show each block)
md-go-validator -v .

# Quiet mode (summary only)
md-go-validator -q .
```

## Options

| Option           | Description                                   |
| ---------------- | --------------------------------------------- |
| `-v, --verbose`  | Show progress for each code block             |
| `-q, --quiet`    | Only show summary (no code in errors)         |
| `--no-code`      | Don't show code snippets in error output      |
| `-l, --language` | Comma-separated list of languages to validate |
| `-f, --format`   | Output format (table, json, yaml, csv, quiet) |
| `--color`        | Color mode (auto, always, never)              |
| `-o, --output`   | Write output to file                          |
| `-t, --timeout`  | Timeout for validation (e.g., 30s, 5m)        |
| `-h, --help`     | Show help message                             |

## Supported Languages

| Language   | Identifier(s)            | Parser      |
| ---------- | ------------------------ | ----------- |
| Go         | `go`, `golang`           | Built-in    |
| TypeScript | `typescript`, `ts`       | Tree-sitter |
| TSX        | `tsx`                    | Tree-sitter |
| Rust       | `rust`, `rs`             | Tree-sitter |
| Nix        | `nix`                    | Tree-sitter |
| HCL        | `hcl`, `terraform`, `tf` | Tree-sitter |
| Templ      | `templ`                  | Tree-sitter |

All parsers are embedded in the binary - no external tools required.

## Language Selection Examples

```bash
# Validate Go and TypeScript only
md-go-validator -l go,typescript .

# Validate all supported languages found in markdown
md-go-validator -l go,templ,typescript,nix,rust,hcl .

# Short form
md-go-validator -l go,ts,rs .
```

## Skip Directives

For intentionally incomplete code snippets (common in documentation), place these **before** the code block or **inside** it:

````markdown
<!-- skip-validate -->

```go
// This is intentionally partial
type MyStruct struct {
    Name string
}
```
````

### Available Directives

- `<!-- skip-validate -->`
- `<!-- skip-md-validate -->`
- `<!-- md-skip -->`
- `<!-- no-validate -->`
- `// skip-validate`
- `//nolint`

## How It Works

### Go Parsing Strategies

The Go validator tries multiple approaches to parse code:

1. **Complete File** - Parse as-is
2. **Package Wrapper** - Wrap in `package main`
3. **Function Wrapper** - Wrap in `func main() { ... }`
4. **Expression** - Wrap as `_ = <code>`
5. **Statements** - Wrap as function body

This handles:

- Complete programs
- Type declarations
- Function signatures
- Import statements
- Variable declarations
- Individual expressions

### Tree-Sitter Language Validators

For other languages, the tool uses [gotreesitter](https://github.com/odvcencio/gotreesitter) — a pure Go tree-sitter binding embedded in the binary. No external CLI tools required:

- **TypeScript** — tree-sitter parser
- **TSX** — tree-sitter parser
- **Rust** — tree-sitter parser
- **Nix** — tree-sitter parser
- **HCL/Terraform** — tree-sitter parser
- **Templ** — tree-sitter parser

All parsers are compiled into the binary — no external dependencies needed.

## Output Formats

```bash
# JSON output for CI integration
md-go-validator -f json .

# YAML output
md-go-validator -f yaml .

# CSV for spreadsheets
md-go-validator -f csv .

# Markdown table
md-go-validator -f markdown .
```

## CI Integration

```yaml
# .github/workflows/docs.yml
name: Validate Documentation
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go install github.com/larsartmann/md-go-validator@latest
      - run: md-go-validator -f json -o results.json .
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"

    mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
    "github.com/larsartmann/md-go-validator/pkg/languages"
)

func main() {
    // Create validator for specific languages
    validator := mdgovalidator.New(true).
        WithLanguages([]languages.Language{
            languages.LangGo,
            languages.LangTypeScript,
        })

    // Validate a file
    ctx := context.Background()
    results, err := validator.ValidateFile(ctx, "README.md")
    if err != nil {
        panic(err)
    }

    // Process results
    for _, r := range results {
        if r.HasError() {
            fmt.Printf("Error at line %s: %v\n", r.LineNumber, r.Error)
        }
    }
}
```

## Architecture

The validator uses a pluggable architecture:

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Extractor     │────▶│  Language Registry│────▶│   Validators    │
│                 │     │                  │     │                 │
│ - Parse markdown│     │ - Map lang to    │     │ - Go (built-in) │
│ - Find code     │     │   validator      │     │ - External cmds │
│   blocks        │     │ - Check avail.   │     │ - Custom        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

### Adding a New Language Validator

<!-- skip-validate -->

```go
// Create a validator
type MyLanguageValidator struct{}

func (v *MyLanguageValidator) Language() languages.Language {
    return languages.Language("mylang")
}

func (v *MyLanguageValidator) Validate(ctx context.Context, code string) error {
    // Validation logic here
    return nil
}

func (v *MyLanguageValidator) IsAvailable() bool {
    // Check if required tools are installed
    return true
}

// Register it
registry := languages.NewRegistry()
registry.Register(&MyLanguageValidator{})
```

## Future Enhancements

- **More languages** - Python, Java, C/C++, etc.
- **Custom validators** - User-defined validation rules
- **Fix suggestions** - Auto-fix common syntax errors

## License

MIT License - See [LICENSE](LICENSE) for details.
