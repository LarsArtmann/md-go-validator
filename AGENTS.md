# md-go-validator - Project Instructions

## Overview

A Go tool that validates code blocks embedded in Markdown and MDX documentation files.
Uses multiple parsing strategies to handle partial code snippets commonly found in technical documentation.

## Tech Stack

- Go 1.26.3+
- [gotreesitter](https://github.com/odvcencio/gotreesitter) v0.20.2 — pure Go tree-sitter for multi-language parsing
- [go-output](https://github.com/larsartmann/go-output) v0.11.0 — multi-format output (JSON, YAML, CSV, table)
- [go-finding](https://github.com/larsartmann/go-finding) v1.0.0 — neutral Finding type for SARIF/LSP/JSON interchange
- [go-faster/yaml](https://github.com/go-faster/yaml) v0.4.6 — YAML parsing for config files
- Library code in `pkg/`
- CLI entry point in `cmd/md-go-validator/`

## Build Commands

```bash
nix build .#                   # Build the package
nix flake check                # Run all checks (format, build, test)
nix fmt                        # Format .nix and .go files
nix develop                    # Enter dev shell
go build ./cmd/md-go-validator
go test ./...
go test -cover ./...
go test -bench=. -benchmem ./pkg/
go test -race ./...
golangci-lint run ./...
go run ./cmd/md-go-validator [options] [path...]
goreleaser release
```

## Architecture

### pkg/parser.go

- `ValidateGoCode(code string) error` — Delegates to `GoValidator.Validate()` (single source of truth)

### pkg/extractor.go

- `ExtractCodeBlocks(content, langs)` — Extracts code blocks for specified languages
- `ExtractGoCodeBlocks(content)` — Backwards-compatible Go-only extraction
- `ExtractCodeBlocksWithConfig(content, langs, config)` — Custom skip directives
- `SkipDirectivesConfig` — Configurable skip directives

### pkg/validator.go

- `FileValidator` — Main validator with file/directory validation
- `New(verbose) → WithLanguages().WithMaxFiles().WithConcurrency().WithExcludePatterns().WithSkipDirectives().WithFileFilter()` — Functional options pattern
- `SupportedExtensions()` → `[]types.FileType`
- `IsSupportedFile(path)` → bool
- `ValidateDirectoryFunc(ctx, dir, callback)` — Streaming validation with per-result callback
- Concurrent directory processing via worker pool with channels
- Exported sentinel errors: `ErrPathEmpty`, `ErrNoValidatorForLang`, `ErrPathNullByte`

### pkg/context.go

- `ContextConfig` — Timeout, deadline, parent context propagation (file/block limits live on `FileValidator`, not here)
- `Build()` / `Branch()` / `BranchWithTimeout()` — Context lifecycle management

### pkg/types/

Branded types for type safety:

- `FileID(string)`, `LineNumber(uint)`, `BlockIndex(uint)`, `FileType(string)`
- `ExcludePattern(string)` — with `Match(path) bool` method encapsulating glob matching
- `ValidationStatus` — enum: unknown/valid/skipped/error
- `CodeBlock`, `Result`, `ReportData`, `ErrorEntry`
- All have `Validate()`, `String()` methods

### pkg/languages/

- `Language(string)` — branded type with `IsSupported()`, `Validate()`, `Extensions()`
- `Validator` interface — `Language()`, `Validate(ctx, code)`, `IsAvailable()`
- `Registry` — validator registry with `Register()`, `Get()`, `GetByString()`, `GetAvailable()`
- `GoValidator` — stdlib parser with 6-strategy approach + elision normalization + pseudo go.mod detection
- `TreeSitterValidator` — tree-sitter based validator for rust/typescript/tsx/nix/hcl/terraform/templ

### pkg/config/

- `Config` struct with `Languages []languages.Language` (typed, not string)
- `Load()`, `LoadFromDir()`, `Save()`, `InitFile()` for `.md-go-validator.yaml`
- Language validation happens at YAML parse time via `UnmarshalText`

### pkg/baseline/

- `Set`, `Signature()`, `Load()`, `Save()`, `FilterNew()` — baseline regression mode
- Signature format: `file:line:errorcode` (includes error code for precision)

### pkg/finding/

- `FromResult()` / `FromResults()`: converts `types.Result` → neutral `go-finding.Finding`

### pkg/output/

- `PrintReport()` / `PrintReportTo()` — Multi-format output (table/json/yaml/csv/markdown/quiet)
- `Format`, `ColorMode` — type aliases from go-output

### pkg/code/

- `IndentCode()` — Indent for function wrapping
- `ParseGo()` — Go stdlib parser wrapper
- `NormalizeDocIdioms()` — Normalizes documentation elision idioms (`{ ... }` → `{}`, ellipsis-only lines dropped)
- `IsPseudoModuleFile()` — Detects go.mod directives in Go code blocks (require/replace/module)

### pkg/config/

- `Config` struct — Project-level config (languages, exclude, skip-directives, format)
- `Load(path)` / `LoadFromDir(dir)` — YAML/JSON config loading
- `Save(path, cfg)` — Write config
- `InitFile(path)` — Scaffold default `.md-go-validator.yaml`

### pkg/finding/

- `FromResult(r)` / `FromResults([]r)` — Convert validation Results to neutral go-finding Findings
- Enables SARIF/LSP/JSON interchange with a one-liner

### pkg/baseline/

- `Set` — Collection of known error signatures (file:line)
- `Load(path)` — Read baseline file
- `FilterNew(results)` — Filter out known errors, return only new ones

## Key Patterns

### Multi-Strategy Parsing (single implementation in GoValidator)

0. Pre-processing: NormalizeDocIdioms ({ ... } → {}, ellipsis line removal)
0. Pseudo go.mod detection (skip module directives)
1. Complete file parsing
2. Package wrapper (`package main`)
3. Function wrapper (`func main()`)
4. Expression (`_ = <code>`)
5. Statements (function body)
6. Imports + statements (split import block from statements — the dominant docs pattern)

Error reporting uses best-attempt selection (highest error line from the strategy that parsed furthest), plus mixed-scope detection and skip-directive hints.

### Skip Directives

- `<!-- skip-validate -->`, `<!-- skip-md-validate -->`, `<!-- md-skip -->`, `<!-- no-validate -->`
- `// skip-validate`, `//nolint`

### Branded Types

All domain values use branded types to prevent mixing (FileID vs BlockIndex vs LineNumber).
Pattern: type + `New*()` constructor + `String()` + `Validate()` methods.

## Code Style

- Functional programming: immutability, pure functions, composition
- Early returns over nested conditionals
- Small, focused functions (max 10 complexity)
- All tests must use `t.Parallel()`
- All struct fields must be explicit in tests (exhaustruct)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- `golangci-lint run ./...` must show 0 issues

## Testing

- Unit tests in `*_test.go` files alongside source
- Integration tests in `pkg/integration_test.go` using `pkg/testdata/` fixtures
- Benchmarks in `pkg/benchmark_test.go`
- Test helpers in `pkg/testutil/` and `pkg/types/testing.go`

## Linter Gotchas

- `ireturn` vs `nolintlint` catch-22: resolved via `.golangci.yml` `linters.exclusions.rules`
- `wsl_v5` requires blank lines before certain constructs
- `noinlineerr` forbids `if err := ...; err != nil`
- `wrapcheck` requires wrapping errors from external packages

## Coverage

| Package       | Coverage |
| ------------- | -------- |
| pkg           | 80.1%    |
| pkg/baseline  | 73.0%    |
| pkg/code      | 95.7%    |
| pkg/config    | 84.8%    |
| pkg/finding   | 100.0%   |
| pkg/languages | 88.0%    |
| pkg/output    | 91.0%    |
| pkg/types     | 83.7%    |
| cmd           | 74.0%    |

## Nix

- `flake.nix` uses flake-parts + treefmt-nix
- Inputs: nixpkgs (nixos-unstable), systems, flake-parts, treefmt-nix (all with proper `follows`)
- `nix build .#` — build the package
- `nix flake check` — format check + build check + test check
- `nix fmt` — formats .nix (nixfmt) and .go (gofmt) via treefmt
- `nix develop` — dev shell with go, gopls, golangci-lint, goreleaser
- Source filtering via `lib.fileset` (only includes go.mod, go.sum, cmd/, pkg/)
- Version derived from git: `self.rev or self.dirtyRev or "dev"`
- Overlay exported at `overlays.default` via `package.nix`
- Previous nix build issue resolved: go.work removed, go-output v0.10.0 published with stable API

## Release

```bash
goreleaser release
```
