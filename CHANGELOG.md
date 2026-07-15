# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- Documentation website at [md-go-validator.lars.software](https://md-go-validator.lars.software) (Astro + Starlight + Tailwind v4)
- CI/CD workflow for automatic website deployment to Firebase Hosting (`.github/workflows/website.yml`)
- SARIF output format for CI integration (GitHub Code Scanning)
- `--config` flag for explicit config file path
- `--save-baseline` flag and improved baseline signature precision (includes error code)
- `--list-languages` flag to print all supported languages
- `--fail-on-skipped` flag for strict mode (exit 1 if any blocks skipped)
- `**` recursive glob support for exclude patterns via doublestar
- FEATURES.md, TODO_LIST.md, ROADMAP.md documentation

### Changed

- Rewrote README.md for public presence (badges, comparison table, install/usage, library API, CI integration)
- Replaced proprietary LICENSE template with MIT License
- Updated `.goreleaser.yml` homepage URLs to `https://md-go-validator.lars.software`
- Upgraded go-output to v0.30.1 (now uses `go-output/delimited` and `go-output/serialization` sub-packages)
- Upgraded go-finding to v1.2.0 (breaking: `Position.File` is now branded `FilePath` type)
- Upgraded gotreesitter to v0.37.0
- Upgraded Go version to 1.26.4
- CLI flags now override (not union with) config file repeatable values
- `ValidateDirectoryFunc` now truly streams results via worker pool
- Type `Config.Languages` as `[]languages.Language` to eliminate stringly-typed split brain
- `ExcludePattern` is now a branded type with encapsulated `Match` logic
- Added `GOPRIVATE` for `github.com/larsartmann/*` modules in CI devShell
- CI uses golangci-lint-action v7 with `GOEXPERIMENT=jsonv2`

### Removed

- Ghost `CodeBlock` methods that split validation status
- Ghost `ContextConfig.Branch` methods (simplified to `Build()` / `BranchWithTimeout()`)
- CLI indirection wrappers (direct handler dispatch)

### Fixed

- Panic in `formatSupportedExtensions` during verbose mode
- `ValidateDirectoryFunc` streaming correctness via worker pool

## [0.3.0] - 2026-06-17

### Added

- STDIN support: pipe markdown via `-` (e.g., `cat README.md | md-go-validator -`)
- Structured exit codes: 0=success, 1=validation errors, 2=tool/usage errors
- `ErrorCode` branded type with `String()`, `Validate()`, and JSON text marshaling
- `ErrorCode` threaded through `Result` and `ErrorEntry` (visible in JSON/YAML output)
- `FileValidator.ValidateContent` for validating raw markdown content
- Nix package `meta.platforms` attribute (all platforms) and maintainer record
- Dedicated extractor test suite (16 table-driven tests)
- CLI integration tests for exit codes, `--output`, `--timeout`, `--language`
- CI `nix flake check` job and expanded dogfooding (validator runs on its own docs)

### Changed

- Single source of truth for supported file types (consolidated into `types` package)
- Simplified directory-processing concurrency (~70 lines removed)
- Consolidated ANSI color constants into the `output` package
- Consistent `ValidationError` receivers and deduplicated `TreeSitterValidator`
- Updated Go dependencies (go-output v0.11.0, gotreesitter v0.20.2, go-toml v2.4.0)

### Removed

- Dead `ContextConfig` limit fields (split-brain with `FileValidator`)

### Fixed

- `nix build .#` was completely broken (restored via `callPackage ./package.nix`)
- Nix `overlays.default` threw `attribute 'self' missing` (now functional)
- `BuildReportData` could panic on `StatusError` with nil `Error` (made unrepresentable)
- Extractor pre-marked code blocks valid before validation ran
- CLI help falsely claimed tree-sitter languages require external tools (`tsc`, `rustc`, etc.)
- Misleading quiet-mode summary output
- Stale `vendorHash` after `go.sum` cleanup (removed unused test-only dependencies)

### Security

- `Result` `StatusError ⟺ Error != nil` invariant now enforced at construction time

## [0.2.0] - 2026-06-13

### Added

- MDX file support (`.mdx` files are now scanned alongside `.md` and `.markdown`)
- Multi-format output support (JSON, YAML, Markdown, CSV, Table, Quiet)
- `--format` / `-f` flag for output format selection
- `--color` flag for color mode control (auto, always, never)
- `--version` / `-V` flag for printing the binary version
- `--timeout` / `-t` flag for validation timeouts
- Context support for cancellation in `ValidateFile` and `ValidateDirectory`
- `Validator` interface for dependency injection and testing
- Branded types: `FileID`, `LineNumber`, `BlockIndex`
- `ValidationStatus` enum replacing boolean `Skipped` field
- `SkipDirectivesConfig` type with `DefaultSkipDirectives()` function
- GitHub Actions CI workflow (`.github/workflows/ci.yml`)
- `pkg/output` package for output formatting using `go-output`
- `pkg/types` package for domain types
- Nix flake overlay (`flake.overlays.default`) via `package.nix`

### Changed

- Migrated `PrintReport` to `output.PrintReport` in `pkg/output`
- Replaced global mutable `SkipDirectives` with immutable `SkipDirectivesConfig`
- Improved test coverage: types 91%, output 92%, pkg 85.2%, cmd 70.9%
- CSV output now uses `go-output`'s `CSVWriter` for proper escaping
- Upgraded `go-output` to v0.10.0
- Rewrote `CONTRIBUTING.md` for the current Nix-based workflow

### Deprecated

- `mdgovalidator.PrintReport()` - use `output.PrintReport()` instead

### Removed

- Global mutable `SkipDirectives` variable (replaced by `SkipDirectivesConfig`)
- Deprecated `PrintReport` from `pkg/validator.go`

### Fixed

- Path traversal security (gosec G304) with `validateAndCleanPath`
- Context cancellation support in directory walking
- Nix build: updated `vendorHash` after `go-output` v0.10.0 upgrade
- Nix overlay: created missing `package.nix` referenced by `flake.overlays.default`
- `CONTRIBUTING.md` dead references to `just` and non-existent scripts

### Security

- Added path validation to prevent path traversal attacks

## [0.1.0] - 2026-01-01

### Added

- Initial release
