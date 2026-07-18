<h1 align="center">md-go-validator</h1>

<p align="center"><strong>Validate code blocks in your Markdown and MDX documentation.</strong></p>

<p align="center">
<a href="https://pkg.go.dev/github.com/larsartmann/md-go-validator"><img src="https://pkg.go.dev/badge/github.com/larsartmann/md-go-validator.svg" alt="Go Reference"></a>
<a href="https://github.com/LarsArtmann/md-go-validator/actions/workflows/ci.yml"><img src="https://github.com/LarsArtmann/md-go-validator/actions/workflows/ci.yml/badge.svg" alt="CI"></a>

<a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
</p>

<p align="center">
<a href="https://md-go-validator.lars.software">Documentation</a> · <a href="https://pkg.go.dev/github.com/larsartmann/md-go-validator">API Reference</a> · <a href="CHANGELOG.md">Changelog</a>
</p>

---

A CLI tool and Go library that validates code blocks embedded in Markdown and MDX documentation files. It catches syntax errors in your documentation before users do, using multiple parsing strategies to handle the partial code snippets commonly found in technical docs.

Supports Go (built-in parser), TypeScript, TSX, Rust, Nix, HCL/Terraform, and Templ (all via pure-Go tree-sitter, no CGO required).

## Why?

Documentation rots. Code examples that compiled when written break silently after refactoring. Most projects have no way to catch this until a user files a bug report.

**md-go-validator** runs in CI and validates every code block in your `.md` and `.mdx` files. It understands that documentation snippets are rarely complete programs, so it tries multiple parsing strategies, from complete-file parsing down to wrapping fragments as expressions or statements.

## Comparison

| Feature                     | markdownlint | prettier | md-go-validator |
| --------------------------- | :----------: | :------: | :-------------: |
| Validates code block syntax |              |          |        ✓        |
| Go parser (multi-strategy)  |              |          |        ✓        |
| Tree-sitter (7 languages)   |              |          |        ✓        |
| Pure Go (no CGO)            |              |          |        ✓        |
| CI exit codes               |      ✓       |    ✓     |        ✓        |
| JSON/SARIF output           |              |          |        ✓        |
| Baseline regression mode    |              |          |        ✓        |
| Skip directives             |      ✓       |          |        ✓        |

## How It Works

1. **Extract** code blocks from Markdown/MDX files (recursive directory scanning)
2. **Route** each block to the correct language validator via the registry
3. **Validate** Go blocks through 6 progressive parsing strategies (complete file, package wrapper, function wrapper, expression, statements, imports+statements)
4. **Validate** other languages via embedded pure-Go tree-sitter parsers
5. **Report** results in table, JSON, YAML, CSV, Markdown, SARIF, or quiet format

## Install

```bash
go install github.com/larsartmann/md-go-validator@latest
```

Or use the GitHub Action in your workflow (no install needed):

```yaml
- uses: LarsArtmann/md-go-validator@v1
  with:
    path: .
```

## Usage

```bash
# Validate all Go code blocks in the current directory
md-go-validator .

# Validate multiple languages
md-go-validator -l go,typescript,rust .

# JSON output for CI integration
md-go-validator -f json -o results.json .

# Pipe markdown via stdin
cat README.md | md-go-validator -
```

## CLI Options

| Option              | Description                                                    |
| ------------------- | -------------------------------------------------------------- |
| `-v, --verbose`     | Show progress for each code block                              |
| `-q, --quiet`       | Only show summary (no code in errors)                          |
| `-l, --language`    | Comma-separated list of languages to validate                  |
| `-f, --format`      | Output format (table, json, markdown, yaml, csv, quiet, sarif) |
| `--color`           | Color mode (auto, always, never)                               |
| `-o, --output`      | Write output to file                                           |
| `-t, --timeout`     | Timeout for validation (e.g. 30s, 5m)                          |
| `--exclude`         | Glob pattern to exclude (repeatable, supports `**`)            |
| `--skip-directive`  | Custom skip directive (repeatable)                             |
| `--init`            | Create a default `.md-go-validator.yaml` config file           |
| `--baseline`        | Baseline file of known errors; only new errors fail            |
| `--save-baseline`   | Generate baseline file from current run's errors               |
| `--config`          | Path to config file (default: auto-discover in CWD)            |
| `--list-languages`  | List all supported languages                                   |
| `--fail-on-skipped` | Exit 1 if any blocks are skipped                               |

### Configuration File

Create a `.md-go-validator.yaml` in your project root:

```yaml
languages:
  - go
  - typescript
exclude:
  - "vendor/*"
skipDirectives:
  - "<!-- sketch -->"
format: table
```

Run `md-go-validator --init` to scaffold a default config file. CLI flags override config file values.

## Supported Languages

| Language      | Identifier(s)            | Parser                |
| ------------- | ------------------------ | --------------------- |
| Go            | `go`, `golang`           | Built-in (6-strategy) |
| TypeScript    | `typescript`, `ts`       | Tree-sitter           |
| TSX           | `tsx`                    | Tree-sitter           |
| Rust          | `rust`, `rs`             | Tree-sitter           |
| Nix           | `nix`                    | Tree-sitter           |
| HCL/Terraform | `hcl`, `terraform`, `tf` | Tree-sitter           |
| Templ         | `templ`                  | Tree-sitter           |

All parsers are embedded in the binary. No external tools required.

## Go Parsing Strategies

The Go validator tries multiple approaches to handle partial code snippets:

0. **Pre-processing** — Normalize documentation idioms (`{ ... }` to `{}`, drop ellipsis-only lines)
1. **Pseudo go.mod detection** — Skip module directives (`require`, `replace`, `module`)
2. **Complete File** — Parse as-is
3. **Package Wrapper** — Wrap in `package main`
4. **Function Wrapper** — Wrap in `func main() { ... }`
5. **Expression** — Wrap as `_ = <code>`
6. **Statements** — Wrap as function body
7. **Imports + Statements** — Split import block from statements (the dominant docs pattern)

Error reporting uses best-attempt selection (the error from the strategy that parsed furthest).

## Skip Directives

For intentionally incomplete code snippets, place these **before** the code block or **inside** it:

````markdown
<!-- skip-validate -->

```go
type MyStruct struct {
    Name string
}
```
````

````

### Available Directives

- `<!-- skip-validate -->`
- `<!-- skip-md-validate -->`
- `<!-- md-skip -->`
- `<!-- no-validate -->`
- `// skip-validate`
- `//nolint`

## Output Formats

```bash
md-go-validator -f json .       # JSON for CI integration
md-go-validator -f sarif .      # SARIF for GitHub Code Scanning
md-go-validator -f yaml .       # YAML
md-go-validator -f csv .        # CSV for spreadsheets
md-go-validator -f markdown .   # Markdown table
````

The JSON output format is [documented with a JSON Schema](docs/json-schema.json). Each error entry includes an `errorCode` field (`syntax`, `not_available`, `not_registered`, or `unknown`) for programmatic error classification.

## Exit Codes

| Code | Meaning           | When                                       |
| ---- | ----------------- | ------------------------------------------ |
| `0`  | Success           | All code blocks valid or skipped           |
| `1`  | Validation errors | One or more code blocks have syntax errors |
| `2`  | Tool error        | File not found, bad flags, I/O errors      |

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
    validator := mdgovalidator.New(true).
        WithLanguages([]languages.Language{
            languages.LangGo,
            languages.LangTypeScript,
        })

    ctx := context.Background()
    results, err := validator.ValidateFile(ctx, "README.md")
    if err != nil {
        panic(err)
    }

    for _, r := range results {
        if r.HasError() {
            fmt.Printf("Error at line %s: %v\n", r.LineNumber, r.Error)
        }
    }
}
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

Or use the built-in GitHub Action:

```yaml
- uses: LarsArtmann/md-go-validator@v1
  with:
    path: .
    args: "-f json"
```

## Dependencies

| Dependency                                                | Purpose                                             |
| --------------------------------------------------------- | --------------------------------------------------- |
| [gotreesitter](https://github.com/odvcencio/gotreesitter) | Pure-Go tree-sitter for multi-language parsing      |
| [go-output](https://github.com/larsartmann/go-output)     | Multi-format output (JSON, YAML, CSV, SARIF, table) |
| [go-finding](https://github.com/larsartmann/go-finding)   | Neutral Finding type for SARIF/LSP/JSON interchange |
| [go-faster/yaml](https://github.com/go-faster/yaml)       | YAML config file parsing                            |

## Development

```bash
nix build .#          # Build the package
nix flake check       # Run all checks (format, build, test)
nix develop           # Enter dev shell
go test ./...         # Run tests
golangci-lint run ./...  # Lint
```

## License

MIT License. See [LICENSE](LICENSE) for details.
