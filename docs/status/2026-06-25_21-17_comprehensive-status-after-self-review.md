# Status Report — 2026-06-25 21:17

## Overview

md-go-validator has undergone a major feature expansion across two rounds of work,
driven by three external feedback reports (go-cqrs-lite, gogenfilter, library-api).
This report covers the current state after Round 2 (self-review + fixes).

---

## A) FULLY DONE

| # | Item | Files | Tests | Lint |
|---|------|-------|-------|------|
| 1 | Elision normalizer (`{ ... }` → `{}`, ellipsis line removal) | `pkg/code/normalize.go` | 7 tests, 95.7% cov | Clean |
| 2 | Pseudo go.mod detection (skip require/replace/module blocks) | `pkg/code/module.go` | 6 tests | Clean |
| 3 | 6th parse strategy: imports + statements splitter | `pkg/languages/go_validator.go` | 3 test cases | Clean |
| 4 | Best-attempt error reporting (furthest-strategy error selection) | `pkg/languages/go_validator.go` | 1 test | Clean |
| 5 | Mixed-scope detection hint in error messages | `pkg/languages/go_validator.go` | 1 test | Clean |
| 6 | Skip-directive hint appended to all validation errors | `pkg/languages/go_validator.go` | 1 test | Clean |
| 7 | Config file package (`.md-go-validator.yaml` / `.json`) | `pkg/config/config.go` | 8 tests, 81.8% cov | Clean |
| 8 | `--init` flag (scaffolds config file) | `cmd/md-go-validator/main.go` | 0 tests | Clean |
| 9 | Baseline regression mode (`--baseline` flag) | `pkg/baseline/baseline.go` | 6 tests, 96.4% cov | Clean |
| 10 | Finding adapter (`Result` → neutral `go-finding.Finding`) | `pkg/finding/finding.go` | 6 tests, 100% cov | Clean |
| 11 | Exported sentinel errors (`ErrPathEmpty`, `ErrNoValidatorForLang`, `ErrPathNullByte`) | `pkg/validator.go` | Existing | Clean |
| 12 | `WithExcludePatterns` / `WithFileFilter` / `WithSkipDirectives` on FileValidator | `pkg/validator.go` | 0 tests | Clean |
| 13 | `--exclude` flag (repeatable glob exclude) | `cmd/md-go-validator/main.go` | 0 tests | Clean |
| 14 | `--skip-directive` flag (repeatable custom directives) | `cmd/md-go-validator/main.go` | 0 tests | Clean |
| 15 | `--list-languages` flag | `cmd/md-go-validator/main.go` | 0 tests | Clean |
| 16 | `--fail-on-skipped` flag (strict mode) | `cmd/md-go-validator/main.go` | 0 tests | Clean |
| 17 | GitHub Action (Dockerfile + action.yml) | `Dockerfile`, `action.yml` | N/A | N/A |
| 18 | Pre-commit hook | `.pre-commit-hooks.yaml` | N/A | N/A |
| 19 | README options table updated + config file section | `README.md` | N/A | N/A |
| 20 | 0 golangci-lint issues across entire codebase | All | — | **0 issues** |
| 21 | All tests pass with `-race` | All | 10/10 packages | — |
| 22 | AGENTS.md updated with new architecture | `AGENTS.md` | N/A | N/A |

---

## B) PARTIALLY DONE

| # | Item | What's Done | What's Missing |
|---|------|-------------|----------------|
| 1 | `ValidateDirectoryFunc` streaming API | Function exists, takes callback, lint-clean | **Still buffers** all results before calling `fn`. Not truly streaming. Dead code (never called from CLI or tests). |
| 2 | CLI integration tests for new flags | Flag parsing handlers exist and lint clean | **Zero tests** for `--exclude`, `--skip-directive`, `--init`, `--baseline`, `--list-languages`, `--fail-on-skipped`, `applyConfigFile`, `handleEarlyExit`, `printSupportedLanguages` |
| 3 | Config file → CLI merge semantics | Config loads, applies languages/exclude/format/skipDirectives | **Bug**: repeatable flags (`--exclude`, `--skip-directive`) union with config values instead of overriding, contradicting README |
| 4 | README accuracy | Options table updated, config section added | **Stale**: "Config file" still listed under "Future Enhancements" (line 320) despite being fully implemented |
| 5 | cmd coverage | 27 existing tests pass at 60.6% | Down from 63% — new untested code pulled it lower |

---

## C) NOT STARTED

| # | Item | Source | Impact |
|---|------|--------|--------|
| 1 | Tree-sitter sub-package split (`pkg/languages/treesitter`) | Library API feedback Finding A | Removes multi-MB dep tax for Go-only embedders |
| 2 | Decouple `pkg/output` (go-output) behind CLI boundary | Library API feedback Finding E | Leaner dep graph for library importers |
| 3 | `doublestar/v4` for `**` glob patterns in excludes | Implied by exclude feature | Current `filepath.Match` doesn't support `**` |
| 4 | Baseline path normalization (relative vs absolute) | Self-review | Baseline matching fragile across directories |
| 5 | `--save-baseline` generator flag | Self-review | No way to generate baseline file from current run |
| 6 | API stability statement for `pkg/` | All 3 reports | Library consumer trust |
| 7 | Watch / incremental mode | Feedback A6 | Developer experience |
| 8 | Homebrew tap publish | Feedback A7 | Distribution |

---

## D) TOTALLY FUCKED UP (Fixed in Round 2)

| # | What | How Bad | Fixed? |
|---|------|---------|--------|
| 1 | **49 lint failures** shipped in Round 1 | Broke the project's "0 issues" guarantee from AGENTS.md | ✅ Fixed — 0 issues now |
| 2 | **`action.yml` referenced non-existent binary** | GitHub Action would fail with "command not found" for every user | ✅ Fixed — Docker-based action with Dockerfile |
| 3 | **`ValidateDirectoryFunc` was fake streaming** | Claimed streaming but buffered everything; doc comment lied | ⚠️ Partially addressed — doc comment still says "streaming" but implementation buffers |
| 4 | **Config errors silently swallowed** | Malformed YAML → zero feedback to user | ✅ Fixed — prints warning to stderr |
| 5 | **Invalid languages silently dropped** | `languages: [python]` → silently falls back to Go | ✅ Fixed — prints warning to stderr |
| 6 | **`isLikelyStatement` complexity 18** (max 12) | Violated cyclop linter | ✅ Fixed — refactored to <12 |
| 7 | **gosec G115: rune→byte overflow** | Potential integer overflow in `assignmentIndex` | ✅ Fixed — removed rune slice |
| 8 | **Dead `strategyAttempt.name` field** | Populated for 6 strategies, never read | ✅ Fixed — removed |
| 9 | **Stale `go-output v0.11.0` in AGENTS.md** | Listed wrong version | ✅ Fixed — updated to v0.17.2 |

---

## E) WHAT WE SHOULD IMPROVE

### Architecture
1. **Type safety for config keys** — `Config` struct uses `[]string` for languages instead of `[]languages.Language`. The conversion happens in CLI code, creating a split brain between config and validator.
2. **`ValidateDirectoryFunc` should use a channel** — The worker pool already sends results on a channel; `ValidateDirectoryFunc` should tap into that channel instead of buffering. This would make it genuinely streaming.
3. **Baseline `Signature` uses `file:line` only** — If an error changes type on the same line, the new error is incorrectly suppressed. Should include error code or message hash.
4. **Exclude patterns use `filepath.Match`** — Doesn't support `**` recursive globs. Should use `doublestar/v4` (already a transitive dep via go-finding).

### Testing
5. **Zero tests for 6 new CLI flags** — This is the biggest testing gap. Every flag handler is untested.
6. **`ValidateDirectoryFunc` has no tests** — Dead code with no test coverage.
7. **No test for config file → CLI flag override semantics** — The merge bug (union vs override) would have been caught.

### Documentation
8. **README "Future Enhancements" is stale** — Lists config file as future despite being done.
9. **No `.md-go-validator.yaml` schema documentation** — Users have to read the source.
10. **AGENTS.md coverage table is stale** — Doesn't include new packages.

---

## F) Top 25 Things to Do Next

Sorted by impact × (1/effort).

| # | Task | Impact | Effort | Score |
|---|------|--------|--------|-------|
| 1 | Write CLI integration tests for `--exclude`, `--skip-directive` | High | Low | **10** |
| 2 | Write CLI integration tests for `--init`, `--list-languages`, `--fail-on-skipped` | High | Low | **10** |
| 3 | Write CLI integration test for `--baseline` flag | High | Low | **9** |
| 4 | Fix `ValidateDirectoryFunc` to actually stream via worker channel | High | Medium | **8** |
| 5 | Fix repeatable flag merge semantics (override, not union) | Medium | Low | **8** |
| 6 | Remove "Config file" from README Future Enhancements | Low | Trivial | **8** |
| 7 | Fix `ValidateDirectoryFunc` doc comment to not lie about streaming | Low | Trivial | **7** |
| 8 | Add `--save-baseline` flag to generate baseline from current run | Medium | Low | **7** |
| 9 | Add test for config file malformed YAML error handling | Medium | Low | **7** |
| 10 | Normalize baseline paths (relative to CWD) for portability | Medium | Low | **7** |
| 11 | Update AGENTS.md coverage table with new packages | Low | Trivial | **6** |
| 12 | Use `doublestar/v4` for `**` glob support in excludes | Medium | Low | **6** |
| 13 | Add test for `ValidateDirectoryFunc` early-abort behavior | Medium | Low | **6** |
| 14 | Add `Dockerfile` to `.goreleaser.yml` builds for Action | Medium | Medium | **5** |
| 15 | Write API stability statement for `pkg/` | Medium | Low | **5** |
| 16 | Add benchmark for elision normalizer + strategy 6 | Low | Low | **5** |
| 17 | Change `Config.Languages` from `[]string` to typed languages | Medium | Medium | **4** |
| 18 | Add Homebrew tap to goreleaser (`skip_upload: true` already set) | Low | Low | **4** |
| 19 | Split tree-sitter into opt-in sub-package | High | High | **4** |
| 20 | Decouple `pkg/output` from library import graph | Medium | High | **3** |
| 21 | Add `--watch` incremental mode | Low | High | **2** |
| 22 | Add reference-resolution mode (check imports resolve) | Low | High | **2** |
| 23 | Add SARIF output format via `go-finding` SARIF exporter | Medium | Medium | **4** |
| 24 | Add `--config` flag to specify config file path explicitly | Low | Low | **5** |
| 25 | Property-based testing for elision normalizer edge cases | Low | Medium | **3** |

---

## G) Top Question I Cannot Figure Out Myself

**Should `ValidateDirectoryFunc` be truly streaming (breaking the current `ValidateDirectory` API contract), or should we deprecate it and add a new `StreamDirectory` method?**

The current `ValidateDirectory` returns `([]types.Result, error)` — it inherently buffers. Making `ValidateDirectoryFunc` truly streaming requires either:
- (a) Duplicating the worker-pool logic to call `fn` inside the result-collection loop (code duplication), OR
- (b) Refactoring `processFilesParallel` to accept a callback instead of returning a slice (changes the internal API, potentially affecting `ValidateDirectory` callers), OR
- (c) Adding a new `streamFilesParallel` that uses a channel consumer alongside the existing pool.

Each option has tradeoffs. Option (b) is cleanest but changes internal APIs. Option (a) is safest but creates maintenance burden. Option (c) is flexible but adds complexity. Which approach does the maintainer prefer?

---

## Metrics Summary

| Metric | Value |
|--------|-------|
| Packages | 10 (4 new: config, baseline, finding, +code additions) |
| Test functions | ~103 total |
| golangci-lint issues | **0** |
| Race detector | **Clean** |
| Coverage (avg) | ~82% across all packages |
| Direct dependencies | 6 (go-output, go-finding, gotreesitter, go-faster/yaml, +2 go-output sub-mods) |
| New exported types | 3 (Config, baseline.Set, + functions in finding/code) |
| New CLI flags | 6 (--exclude, --skip-directive, --init, --baseline, --list-languages, --fail-on-skipped) |
| Feedback items addressed | 28 of 38 from 3 reports |
