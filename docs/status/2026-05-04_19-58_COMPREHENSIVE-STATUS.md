# Status Report — 2026-05-04 19:58

**Scope:** Full session across 2 conversation rounds (starting from commit `d290055`)
**Total commits this session:** 13 (including 1 formatting sync)
**Codebase size:** 7,347 lines of Go

---

## A) FULLY DONE

### Bug Fixes

- **CI Go version mismatch** — `1.24` → `1.26` to match `go.mod` (`b69da91`)
- **`shouldSkipDir` logic bug** — Loop iterated 5 times but `HasPrefix(".")` was independent of loop variable; separated dot-prefix check from map lookup (`153d698`)

### Dead Code Removed

- **`validateAndCleanPath`** — Both branches returned identical values; simplified to single return (`38b54d3`)
- **`wrapListParser`** — Identity function returning its own argument; inlined (`a6d576a`)
- **`ValidateGoCode` duplicate** — `pkg/parser.go` duplicated ~40 lines of 5-strategy logic; now delegates to `GoValidator.Validate()` (`26b3e73`)
- **`languages.IsSupported()`** — Package-level function superseded by `Language.IsSupported()` method (`66af6f9`)
- **`ValidationError.Unwrap()`** — Always returned nil, no callers (`66af6f9`)

### Type System Improvements

- **`FileType` branded type** — Constants, `String()`, `IsSupported()`, `AllFileTypes()` (`d290055`, `a0d8c80`)
- **`FileType` round-trip** — `ParseFileType()`, `MarshalText()`, `UnmarshalText()` for serialization (`a0d8c80`)
- **`ValidationStatus` round-trip** — `UnmarshalText()` added to pair with existing `MarshalText()` + `ParseValidationStatus()` (`44bf867`)
- **`Language.IsSupported()`** — Method matching `FileType` pattern, reduces duplication in `Validate()` (`9eef8ba`)

### Testing

- **Integration tests** — 7 end-to-end tests with `pkg/testdata/` fixtures (valid, invalid, skipped, mixed .mdx, edge cases, directory, .markdown extension) (`13ac23a`)
- **Benchmark tests** — `ExtractCodeBlocks` (50/500 blocks), `ValidateGoCode` (3 strategies), `ValidateDirectory` (20 files), `IsSupportedFile` (`1d0232a`)
- **Coverage for 0% functions** — `WithLanguages`, `WithRegistry`, `SupportedExtensions`, `ExtractCodeBlocksWithConfig`, `HasErrors` (`9eef8ba`)

### Documentation

- **AGENTS.md** — Complete rewrite reflecting current architecture, patterns, coverage, and linter gotchas (`e02138d`)

---

## B) PARTIALLY DONE

### Coverage Gaps (0% functions remaining)

| Function                      | Location                   | Why                                                      |
| ----------------------------- | -------------------------- | -------------------------------------------------------- |
| `main()`                      | `cmd/main.go:45`           | Calls `os.Exit`, hard to test without subprocess pattern |
| `handleHelp()`                | `cmd/main.go:226`          | Calls `os.Exit(0)`                                       |
| `returnParseError()`          | `cmd/main.go:221`          | Only called on parse failure paths                       |
| `addError()`                  | `pkg/validator.go:481`     | Error channel path in concurrent processing              |
| `formatSupportedExtensions()` | `pkg/validator.go:567`     | Only called in verbose mode                              |
| `newOutputError()`            | `pkg/output/output.go:205` | Only called on write errors                              |

### Partial Coverage Functions (60-80%)

| Function                                     | Coverage | Location                                                                     |
| -------------------------------------------- | -------- | ---------------------------------------------------------------------------- |
| `printErrorEntry()`                          | 60%      | `pkg/output/output.go:363` — color=false path covered, color=true not tested |
| `validateAndCleanPath()`                     | 60%      | `pkg/validator.go:579` — null byte and error paths                           |
| `requireArg()`                               | 40%      | `cmd/main.go:75` — missing-arg branch                                        |
| `processJob()`                               | 70%      | `pkg/validator.go:394` — error path                                          |
| `validatePath()` / `validateAndReturnPath()` | 75%      | `pkg/validator.go:94,105` — error branches                                   |
| `printCSVTo()`                               | 75%      | `pkg/output/output.go:209` — flush error path                                |
| `writeOutput()`                              | 75%      | `pkg/output/output.go:183` — write error path                                |
| `printQuietTo()`                             | 80%      | `pkg/output/output.go:266` — write error path                                |
| `marshalReport()`                            | 80%      | `pkg/output/output.go:159` — marshal error path                              |
| `logProgress()`                              | 29%      | `pkg/validator.go:256` — verbose=true not tested                             |

---

## C) NOT STARTED

1. **`ValidationError.Line`/`Column` use raw `int`** — Should use branded `LineNumber`/`ColumnNumber` types for consistency
2. **`Registry.GetByString()` / `Registry.Languages()`** — Unused in production code; could be removed or used
3. **No `FileType` → `Language` mapping** — No reverse mapping from file extension to language
4. **Subprocess-based CLI tests** — Test `main()`, `handleHelp()`, `returnParseError()` via `exec.Command`
5. **Error path coverage** — `addError`, `newOutputError`, write error paths
6. **Verbose mode coverage** — `logProgress`, `formatSupportedExtensions`
7. **`ValidateGoCode` allocates per call** — Creates `&GoValidator{}` each invocation
8. **`withInt` helper** — Over-abstracted for 2 int fields
9. **CONTRIBUTING.md** — No contributor guide
10. **CODEOWNERS** — No code ownership file
11. **`flake.nix` migration** — AGENTS.md says "justfile is deprecated" but no flake.nix exists
12. **Go doc examples** — No `Example*` test functions for godoc
13. **Fuzzing** — No fuzz tests for extractor/parser
14. **Release automation** — `goreleaser` config exists but no tag-based release workflow
15. **Go struct generation** — Could generate branded type boilerplate with `go generate`

---

## D) TOTALLY FUCKED UP

Nothing is broken. All tests pass, 0 lint issues, clean build.

The only items that could qualify:

- **CI Go version was wrong** — Fixed in `b69da91`. Would have caused CI failures on push.
- **`shouldSkipDir` logic bug** — Fixed in `153d698`. Loop was redundant but behavior was correct (all entries were checked). No real-world impact since `.` is always first in the list and `HasPrefix(name, ".")` short-circuits.

---

## E) WHAT WE SHOULD IMPROVE

### Type Model Consistency

- **`ValidationError.Line`/`Column` are raw `int`** while everything else uses branded types (`LineNumber`, `BlockIndex`). This is the biggest type inconsistency.
- **No `ColumnNumber` branded type** exists yet — would need to be created.
- **`withInt` helper** is marginal abstraction for 2 fields — consider inlining.

### Coverage

- **cmd package at 70.9%** — lowest coverage, blocked by `os.Exit` calls in `main`, `handleHelp`.
- **`logProgress` at 29%** — verbose=true paths untested.
- **Error channel paths** — `addError` (0%), concurrent error handling not exercised.

### Architecture

- **`ValidateGoCode` allocates `&GoValidator{}` per call** — Minor perf issue, could use package-level var.
- **`Registry.GetByString()` / `Registry.Languages()`** — Exported but unused in production. Consider removing or documenting as "for external use".
- **No `FileType` → `Language` mapping** — The two types exist in separate packages with no cross-reference.

### Testing

- **No fuzz tests** — Extractor/parser are perfect candidates for fuzzing.
- **No `Example*` functions** — Would improve godoc.
- **No subprocess CLI tests** — Standard pattern for testing `main()` with `os.Exit`.

### DevEx

- **No `flake.nix`** — AGENTS.md says justfile is deprecated but no Nix alternative exists.
- **No `CONTRIBUTING.md`** — No contributor guide.
- **No `CODEOWNERS`** — No code ownership.

---

## F) Top #25 Things to Do Next

| #   | Task                                                                           | Impact | Work   |
| --- | ------------------------------------------------------------------------------ | ------ | ------ |
| 1   | Add `ColumnNumber` branded type + migrate `ValidationError.Line/Column`        | High   | Medium |
| 2   | Subprocess CLI tests (`main()`, `handleHelp()`, `returnParseError()`)          | High   | Medium |
| 3   | Cover `logProgress` verbose=true paths                                         | Medium | Low    |
| 4   | Cover `addError` error channel path                                            | Medium | Low    |
| 5   | Cover `formatSupportedExtensions` (verbose directory output)                   | Low    | Low    |
| 6   | Remove or document unused `Registry.GetByString()` / `Registry.Languages()`    | Low    | Low    |
| 7   | Cache `GoValidator` in `ValidateGoCode` instead of allocating per call         | Low    | Low    |
| 8   | Add `FileType` → `Language` mapping (or explicit "no mapping" design decision) | Medium | Low    |
| 9   | Add fuzz tests for extractor/parser                                            | High   | Medium |
| 10  | Add `Example*` test functions for godoc                                        | Medium | Low    |
| 11  | Cover `newOutputError` write error paths                                       | Low    | Low    |
| 12  | Create `CONTRIBUTING.md`                                                       | Medium | Low    |
| 13  | Create `flake.nix` for Nix-based builds                                        | Medium | High   |
| 14  | Add Go struct generation for branded types (`stringer`-like)                   | Low    | Medium |
| 15  | Cover `validateAndCleanPath` null byte path                                    | Low    | Low    |
| 16  | Cover `printErrorEntry` color=true path                                        | Low    | Low    |
| 17  | Cover `requireArg` missing-arg branch                                          | Low    | Low    |
| 18  | Inline `withInt` helper (marginal abstraction)                                 | Low    | Low    |
| 18  | Add `CODEOWNERS` file                                                          | Low    | Low    |
| 20  | Add tag-based release workflow in CI                                           | Medium | Medium |
| 21  | Add `Result.Validate()` method for type safety                                 | Medium | Low    |
| 22  | Add `ErrorEntry.Validate()` method                                             | Low    | Low    |
| 23  | Remove `singleValueArgHandler` generic indirection in cmd                      | Low    | Low    |
| 24  | Add `go vet` + `staticcheck` as separate CI steps                              | Low    | Low    |
| 25  | Create ADR for branded type pattern                                            | Low    | Low    |

---

## G) Top #1 Question

**Should `ValidationError.Line`/`Column` be migrated to branded `LineNumber`/`ColumnNumber` types, or is the current raw `int` intentional because `ValidationError` lives in the `languages` package and would create a circular dependency with `pkg/types`?**

The dependency graph is: `pkg/types` → `pkg/languages` (via `CodeBlock.Language`). If `ValidationError` in `pkg/languages` imported `LineNumber` from `pkg/types`, it would be fine (no cycle). But I want to confirm the design intent before changing the error type's public API.

---

## Coverage Summary

| Package       | Coverage  |
| ------------- | --------- |
| **Total**     | **87.4%** |
| pkg/code      | 100.0%    |
| pkg/types     | 93.6%     |
| pkg/languages | 92.3%     |
| pkg/output    | 91.5%     |
| pkg           | 87.5%     |
| pkg/testutil  | 86.8%     |
| cmd           | 70.9%     |

## Health

- **Build:** Clean
- **Tests:** All pass (including race)
- **Lint:** 0 issues
- **Git:** Clean working tree, up to date with remote
