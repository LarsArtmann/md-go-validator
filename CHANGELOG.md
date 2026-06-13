# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
