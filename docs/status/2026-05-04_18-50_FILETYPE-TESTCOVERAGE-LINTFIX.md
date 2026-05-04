# Status Report — 2026-05-04 18:50

## Summary

Added `FileType` branded type, improved test coverage across 4 packages, fixed lint config,
bumped gotreesitter dependency. All tests pass, 0 lint issues.

## Coverage Report

| Package             | Before | After  |
| ------------------- | ------ | ------ |
| cmd/md-go-validator | 61.7%  | 71.1%  |
| pkg                 | 81.4%  | 81.7%  |
| pkg/code            | 100.0% | 100.0% |
| pkg/languages       | 66.7%  | 92.4%  |
| pkg/output          | 91.5%  | 91.5%  |
| pkg/testutil        | 86.8%  | 86.8%  |
| pkg/types           | 79.4%  | 92.8%  |

## What Changed

### FileType Branded Type

- Added `FileType` branded type in `pkg/types/identifiers.go` with constants:
  `FileTypeMarkdown`, `FileTypeMarkdownAlt`, `FileTypeMdx`
- Methods: `String()`, `IsSupported()`, `AllFileTypes()`
- Updated `supportedExtensions` map in `pkg/validator.go` from `map[string]bool`
  to `map[types.FileType]bool`
- `SupportedExtensions()` now returns `[]types.FileType` delegating to `types.AllFileTypes()`

### Test Coverage Improvements

- **pkg/languages**: New `go_validator_test.go` — GoValidator.Validate (valid + invalid),
  GoValidator.Language, GoValidator.IsAvailable, ValidationError.WithCode, ValidationError.Unwrap,
  TreeSitterValidator unavailable language, Language.Extensions all languages
- **pkg/types**: FileType tests, NewLineNumber negative, NewLineNumberFromUint,
  NewBlockIndex negative, NewBlockIndexFromUint, ValidationStatus.MarshalText,
  Result.String/Summary with errors, NewSkippedResultForTest
- **cmd/md-go-validator**: TestParseArgsCombinedFlags, TestUsageDetails, TestUsageHeader,
  TestParseLanguagesDirect, TestParseArgsTimeoutFlag, TestParseArgsLanguageFlag,
  TestWriteOutputToFile_AllFormats

### Lint Fixes

- Fixed `ireturn`/`nolintlint` catch-22: moved interface return exclusions to
  `.golangci.yml` `linters.exclusions.rules` instead of per-line nolint directives
- Removed stale `//nolint:ireturn` comments from `pkg/languages/validator.go`
- Fixed `wsl_v5`, `golines`, `varnamelen`, `goconst` issues in test files

### Dependency Updates

- `gotreesitter`: v0.13.4 → v0.15.3

## What's NOT Done Yet

| Task                                    | Status         | Impact | Work   |
| --------------------------------------- | -------------- | ------ | ------ |
| CI pipeline (.github/workflows/ci.yml)  | Not started    | High   | Low    |
| Integration tests (testdata/)           | Not started    | High   | Medium |
| Benchmark tests                         | Not started    | Medium | Low    |
| cmd/md-go-validator coverage 71% → 80%+ | Partially done | Medium | Medium |
| Pre-commit hook fix                     | Done (shebang) | Low    | Done   |

## Known Issues

- `cmd/md-go-validator` still at 71.1% — `main()`, `returnParseError()`, `handleHelp()`
  are uncovered (they call `os.Exit` which is hard to test without process isolation)
- Pre-commit hook shebang was fixed to `#!/usr/bin/env bash` but not tested on NixOS
