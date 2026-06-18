# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
