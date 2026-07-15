# md-go-validator - Project Instructions

## Overview

A Go tool that validates code blocks embedded in Markdown and MDX documentation files.
Uses multiple parsing strategies to handle partial code snippets commonly found in technical documentation.

## Tech Stack

- Go 1.26.4+
- [gotreesitter](https://github.com/odvcencio/gotreesitter) v0.37.0 — pure Go tree-sitter for multi-language parsing
- [go-output](https://github.com/larsartmann/go-output) v0.30.4 — multi-format output (JSON, YAML, CSV, table, markdown)
- [go-finding](https://github.com/larsartmann/go-finding) v1.2.0 — neutral Finding type for SARIF/LSP/JSON interchange
- [go-faster/yaml](https://github.com/go-faster/yaml) v0.4.6 — YAML parsing for config files
- Library code in `pkg/`
- CLI entry point in `cmd/md-go-validator/`

## Build Commands

**All Go commands require `GOEXPERIMENT=jsonv2`** because `go-output` imports `encoding/json/v2`.
The nix devShell sets this automatically; outside it, prefix commands or `export GOEXPERIMENT=jsonv2`.

```bash
nix build .#                   # Build the package
nix flake check                # Run all checks (format, build, test)
nix fmt                        # Format .nix and .go files
nix develop                    # Enter dev shell (sets GOEXPERIMENT=jsonv2)
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
- Enables SARIF/LSP/JSON interchange with a one-liner
- Uses branded `finding.FilePath` type (v1.2.0 API)

### pkg/output/

- `PrintReport()` / `PrintReportTo()` — Multi-format output (table/json/yaml/csv/markdown/quiet/sarif)
- `Format`, `ColorMode` — type aliases from go-output

### pkg/code/

- `IndentCode()` — Indent for function wrapping
- `ParseGo()` — Go stdlib parser wrapper
- `NormalizeDocIdioms()` — Normalizes documentation elision idioms (`{ ... }` → `{}`, ellipsis-only lines dropped)
- `IsPseudoModuleFile()` — Detects go.mod directives in Go code blocks (require/replace/module)

## Key Patterns

### Multi-Strategy Parsing (single implementation in GoValidator)

0. Pre-processing: NormalizeDocIdioms ({ ... } → {}, ellipsis line removal)
1. Pseudo go.mod detection (skip module directives)
2. Complete file parsing
3. Package wrapper (`package main`)
4. Function wrapper (`func main()`)
5. Expression (`_ = <code>`)
6. Statements (function body)
7. Imports + statements (split import block from statements — the dominant docs pattern)

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

Run `go test -cover ./...` for current numbers.

| Package       | Coverage |
| ------------- | -------- |
| pkg           | 85.2%    |
| pkg/baseline  | 73.0%    |
| pkg/code      | 95.7%    |
| pkg/config    | 84.8%    |
| pkg/finding   | 100.0%   |
| pkg/languages | 88.0%    |
| pkg/output    | 85.7%    |
| pkg/testutil  | 75.0%    |
| pkg/types     | 81.0%    |
| cmd           | 73.9%    |

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

### Nix Gotcha: go-finding-src replace directive

`package.nix` injects a `replace github.com/larsartmann/go-finding => <flake-input-src>`
via `postPatch`. This means the nix build compiles against whatever version the
`go-finding-src` flake input points to, **not** the version in `go.mod`.

**Every `go-finding` bump requires a coordinated 3-place update:**

1. `go.mod` / `go.sum`
2. `flake.nix` `go-finding-src` input ref
3. `flake.lock` re-lock

Forgetting any one produces a split-brain where `go build` (uses go.mod) passes
but `nix build` (uses flake input via replace) fails, or vice versa.

The overlay path (`flake.overlays.default`) calls `package.nix` **without**
`go-finding-src`, so the replace is skipped there.

### Nix Gotcha: GOEXPERIMENT=jsonv2

The dependency `go-output` (and transitively `go-branded-id`) imports
`encoding/json/v2` and `encoding/json/jsontext`, which are experimental
packages in Go 1.26 requiring `GOEXPERIMENT=jsonv2`.

This is set in:

1. `flake.nix` devShells (`GOEXPERIMENT = "jsonv2"`)
2. `flake.nix` apps (`export GOEXPERIMENT=jsonv2` in test/lint scripts)
3. `package.nix` (`GOEXPERIMENT = "jsonv2"` on the derivation)
4. `.github/workflows/ci.yml` (`GOEXPERIMENT: jsonv2` env on test/build jobs)
5. `.golangci.yml` (`goexperiment.jsonv2` build tag for linting)

Without it, `go build` and `go test` fail with:
`build constraints exclude all Go files in .../encoding/json/v2`

## Release

```bash
goreleaser release
```

## Website

- **URL:** [md-go-validator.lars.software](https://md-go-validator.lars.software) (pending DNS propagation)
- **Live:** [md-go-validator.web.app](https://md-go-validator.web.app)
- **Stack:** Astro 7 + Starlight + Tailwind v4 (teal accent `#14b8a6`)
- **Firebase:** Shared `lars-software` project, hosting target `md-go-validator`
- **CI/CD:** `.github/workflows/website.yml` (two-job: build + deploy)
- **Secret:** `FIREBASE_SERVICE_ACCOUNT` (firebase-adminsdk key for lars-software)
- **DNS:** Staged in `domains/lars.software.tf` (CNAME + ACME TXT, BLOCKED on placeholder Namecheap API key)
- **Build:** `cd website && nix shell nixpkgs#nodejs -c npm run build`
- **Deploy:** `cd website && nix shell nixpkgs#nodejs nixpkgs#firebase-tools -c firebase deploy --only hosting:md-go-validator --project lars-software`
