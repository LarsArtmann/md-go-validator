# ADR 001: Test Code Duplication Policy

**Status:** Accepted
**Date:** 2026-06-18
**Deciders:** Project maintainers

## Context

`art-dupl` is run at `-t 25` as the project's standard clone threshold. The
skill `deduplicate-code` distinguishes _harmful_ duplication from _idiomatic_
duplication and says: "Zero harmful duplication. Not zero report lines."

This project chooses to pursue **zero report lines at `-t 25`**: every clone
group the standard threshold surfaces is eliminated via extraction, refactor,
or by adding a shared test helper in `testutil` / `types/testing.go`. The
duplication policy is therefore "duplication is always a smell in this
codebase, even in tests".

The remaining clones at lower thresholds (`-t 20`, `-t 15`) are idiomatic Go
patterns — single-line function signatures, `t.Run` subtest scaffolding, the
generic `AssertZeroValue` helper (different packages, different type
parameters) — and are accepted on purpose.

## Decision

### Standard threshold

`art-dupl --semantic --sort total-tokens -t 25` must report **0 clone groups**.
This is the project's enforced bar and is checked in CI / pre-merge.

### Helper conventions

| Helper                      | Location                           | Purpose                                       |
| --------------------------- | ---------------------------------- | --------------------------------------------- |
| `assertPaths`               | `cmd/md-go-validator/main_test.go` | Assert a config has the expected paths        |
| `newGoValidator`            | `cmd/md-go-validator/main_test.go` | Build a Go-only `FileValidator` for tests     |
| `assertContains`            | `cmd/md-go-validator/main_test.go` | Assert a string contains a substring          |
| `assertSingleSkippedBlock`  | `pkg/extractor_test.go`            | Assert one extracted block is skipped         |
| `assertSingleBlockLanguage` | `pkg/extractor_test.go`            | Assert one extracted block has a language     |
| `validPlusSkipped`          | `pkg/output/output_test.go`        | Two-result sample for format tests            |
| `AssertErrorMessage`        | `pkg/types/testing.go`             | Assert an `ErrorEntry.Error` value            |
| `AssertStatus`              | `pkg/types/testing.go`             | Assert a `Result.Status` value                |
| `AssertIntEquals`           | `pkg/types/testing.go`             | Assert two ints are equal                     |
| `NewTestErrorResultAtZero`  | `pkg/types/testing.go`             | Build a `Result` at (test.md, 1, 0)           |
| `NewTestErrorResult`        | `pkg/testutil/testutil.go`         | Build a `Result` with a `TestError`           |
| `NewTestErrorResultWith`    | `pkg/testutil/testutil.go`         | Build a `Result` with a custom error          |
| `Result.ErrorMessage`       | `pkg/types/result.go`              | Domain method to extract an error string      |
| `reportStdinError`          | `cmd/md-go-validator/main.go`      | Stdin-pipeline error reporter                 |
| `exitHandler`               | `cmd/md-go-validator/main.go`      | Build the (int, bool) tuple for exit handlers |

## Consequences

- **CI gate:** A pre-merge check runs `art-dupl --semantic --sort total-tokens
-t 25` and fails on any non-zero group. The threshold is the contract.
- **Lower thresholds are noise:** At `-t 15` a handful of idiomatic patterns
  remain (subtest scaffolding, single-line function signatures, the generic
  assertion helper repeated across packages with different type parameters).
  These are not worth eliminating — the abstraction would be larger than the
  duplication.
- **Future audits:** A reviewer running the standard command should see
  `Found total 0 clone groups.` Anything else is a regression.

## Verification

```bash
art-dupl --semantic --sort total-tokens -t 25
# Expected: 0 clone groups.
```
