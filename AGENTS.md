# md-go-validator - Project Instructions

## Overview

A Go tool that validates code blocks embedded in Markdown and MDX documentation files.
Uses multiple parsing strategies to handle partial code snippets commonly found in technical documentation.

## Tech Stack

- Go 1.26.2+
- [gotreesitter](https://github.com/odvcencio/gotreesitter) v0.20.2 — pure Go tree-sitter for multi-language parsing
- [go-output](https://github.com/larsartmann/go-output) v0.10.0 — multi-format output (JSON, YAML, CSV, table)
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
- `New(verbose) → WithLanguages().WithMaxFiles().WithConcurrency()` — Functional options pattern
- `SupportedExtensions()` → `[]types.FileType`
- `IsSupportedFile(path)` → bool
- Concurrent directory processing via worker pool with channels

### pkg/context.go

- `ContextConfig` — Timeout, deadline, max files/blocks, parent context propagation
- `Build()` / `Branch()` / `BranchWithTimeout()` — Context lifecycle management

### pkg/types/

Branded types for type safety:

- `FileID(string)`, `LineNumber(uint)`, `BlockIndex(uint)`, `FileType(string)`
- `ValidationStatus` — enum: unknown/valid/skipped/error
- `CodeBlock`, `Result`, `ReportData`, `ErrorEntry`
- All have `Validate()`, `String()` methods

### pkg/languages/

- `Language(string)` — branded type with `IsSupported()`, `Validate()`, `Extensions()`
- `Validator` interface — `Language()`, `Validate(ctx, code)`, `IsAvailable()`
- `Registry` — validator registry with `Register()`, `Get()`, `GetByString()`, `GetAvailable()`
- `GoValidator` — stdlib parser with 5-strategy approach (the canonical implementation)
- `TreeSitterValidator` — tree-sitter based validator for rust/typescript/tsx/nix/hcl/terraform/templ

### pkg/output/

- `PrintReport()` / `PrintReportTo()` — Multi-format output (table/json/yaml/csv/markdown/quiet)
- `Format`, `ColorMode` — type aliases from go-output

### pkg/code/

- `IndentCode()` — Indent for function wrapping
- `ParseGo()` — Go stdlib parser wrapper

## Key Patterns

### Multi-Strategy Parsing (single implementation in GoValidator)

1. Complete file parsing
2. Package wrapper (`package main`)
3. Function wrapper (`func main()`)
4. Expression (`_ = <code>`)
5. Statements (function body)

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
| pkg           | 84.6%    |
| pkg/code      | 100%     |
| pkg/languages | 92.5%    |
| pkg/output    | 91.5%    |
| pkg/types     | 92.8%    |
| cmd           | 70.9%    |

## Nix

- `flake.nix` uses flake-parts + treefmt-nix
- Inputs: nixpkgs (nixos-unstable), systems, flake-parts, treefmt-nix (all with proper `follows`)
- `nix build .#` — build the package
- `nix flake check` — format check + build check + test check
- `nix fmt` — formats .nix (nixfmt) and .go (gofmt) via treefmt
- `nix develop` — dev shell with go, gopls, golangci-lint, goreleaser
- Source filtering via `lib.fileset` (only includes go.mod, go.sum, cmd/, pkg/)
- Version derived from git: `self.rev or self.dirtyRev or "dev"`
- Overlay exported at `overlays.default` for consumption by other flakes
- Previous nix build issue resolved: go.work removed, go-output v0.10.0 published with stable API

## Release

```bash
goreleaser release
```
