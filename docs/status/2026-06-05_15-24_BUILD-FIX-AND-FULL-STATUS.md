# Status Report — 2026-06-05 15:24 CEST

_Generated after fixing the broken build caused by go-output v0.6.3 API migration._

---

## Executive Summary

| Dimension        | Status                                                          |
| ---------------- | --------------------------------------------------------------- |
| **Build**        | PASSING (was broken — fixed this session)                       |
| **Tests**        | ALL GREEN (7/7 packages, ~120 test functions)                   |
| **Lint**         | 0 issues (golangci-lint with 90+ linters)                       |
| **Coverage**     | 70.9%–93.8% across packages                                     |
| **Codebase**     | 7,429 lines of Go (production + tests)                          |
| **Last release** | v0.1.0 (2026-01-01) — unreleased changes accumulating since Jan |
| **Dependencies** | go-output v0.6.3, gotreesitter v0.20.1                          |
| **CI**           | Minimal — test/lint/build on push/PR, no self-validation        |

---

## a) FULLY DONE

### Core Functionality

- **Go code validation** — 5-strategy parser (complete file → package wrap → func wrap → expression → statements). Handles partial snippets from documentation. `pkg/languages/go_validator.go`
- **Multi-language support** — 8 languages via tree-sitter: Go, Templ, TypeScript, TSX, Rust, Nix, HCL, Terraform. `pkg/languages/treesitter_validator.go`
- **Code block extraction** — Line-by-line markdown parser for ```-fenced blocks with language filtering. `pkg/extractor.go`
- **Skip directives** — Configurable: `<!-- skip-validate -->`, `<!-- skip-md-validate -->`, `<!-- md-skip -->`, `<!-- no-validate -->`, `// skip-validate`, `//nolint`
- **Concurrent directory validation** — Worker pool with channels, context cancellation, configurable concurrency. `pkg/validator.go`
- **Context lifecycle management** — Timeout, deadline, max files/blocks, parent propagation, branching for parallel workers. `pkg/context.go`
- **Multi-format output** — Table, JSON, YAML, CSV, Markdown, Quiet. 6 formats via go-output library. `pkg/output/output.go`
- **Branded types** — `FileID`, `LineNumber`, `BlockIndex`, `FileType`, `ValidationStatus`, `Language`. Compile-time type safety across domain. `pkg/types/`

### Infrastructure

- **Nix flake** — flake-parts + treefmt-nix. Build, test, format, devShell, overlay. `flake.nix`
- **GoReleaser** — Cross-compile (linux/darwin/windows, amd64/arm64), cosign signing, SBOMs, brew/scoop/nix/nfpm. `.goreleaser.yml`
- **CI (GitHub Actions)** — Test (with race), lint (golangci-lint), build. `.github/workflows/ci.yml`
- **Linter config** — 90+ linters, Go 1.26.3, experimental build tags. `.golangci.yml`
- **CLI** — Manual arg parsing with generics, `--verbose`, `--quiet`, `--format`, `--color`, `--output`, `--timeout`, `--languages`, multi-path support. `cmd/md-go-validator/main.go`

### Quality

- **Zero code duplication** in production code (threshold 15). 12 acceptable clone groups in test code only.
- **Dedicated test helpers** — `pkg/testutil/` package with assertion helpers.
- **Integration tests** — Real testdata files (valid_go.md, invalid_go.md, skipped.md, mixed.mdx, edge_cases.md).
- **Benchmarks** — 6 benchmark functions covering hot paths.
- **Dead dependency cycle documented** — `pkg/types` ↔ `pkg/languages` cycle identified, proposal written at `docs/modularization/PROPOSAL.md`.

---

## b) PARTIALLY DONE

### Documentation

| Item                      | State | Details                                                                                               |
| ------------------------- | ----- | ----------------------------------------------------------------------------------------------------- |
| `README.md`               | 90%   | Polished, covers CLI + library, architecture diagram slightly misleading (references "External cmds") |
| `CHANGELOG.md`            | 80%   | Unreleased section is substantial — no version bump applied                                           |
| `CONSUMER_PERSPECTIVE.md` | Done  | 20 known gaps with priorities — excellent self-awareness                                              |
| `EXAMPLES.md`             | Done  | All 8 languages, serves as both docs and test fixture                                                 |
| `DOMAIN_LANGUAGE.md`      | 5%    | **Boilerplate template** — no actual domain terms defined                                             |
| `CONTRIBUTING.md`         | 60%   | References nonexistent files (CONTRIBUTING-setup.sh, justfile)                                        |

### GoReleaser

- Config is complete but **Homebrew and Scoop have `skip_upload: true`** — users can't install via these methods.
- Description says "Markdown validation tool for Go projects" — undersells multi-language capability.
- brew test block references `--version` flag which **doesn't exist**.

### Language Support

- 8 languages work, but **no JavaScript, Python, Java, C/C++** despite tree-sitter having grammars for all.
- Tree-sitter errors are **coarse** — "code contains parse errors" with no line/column info.

---

## c) NOT STARTED

- **`--version` flag** — ldflags inject version/commit/date, but no CLI flag exposes them. (#1 consumer complaint in CONSUMER_PERSPECTIVE.md)
- **Config file support** — No `.md-go-validator.yaml` or similar. All configuration is CLI flags only.
- **`--init` command** — No way to bootstrap a config file.
- **Exclude/ignore patterns** — No `.gitignore`-style pattern support for excluding files/dirs.
- **GitHub Action** — No reusable GitHub Action for CI pipelines.
- **Pre-commit hook** — No pre-commit integration.
- **Watch mode** — No file watcher for live re-validation.
- **Error codes in output** — `ErrorCode` type exists but not exposed in CLI output (only in library API).
- **Diff/regression mode** — No way to compare results against a baseline.
- **Self-validation in CI** — The tool doesn't validate its own markdown docs in CI.
- **Tilde (`~~~`) fence support** — Only ``` backtick fences parsed.
- **Indented code block support** — 4-space indented blocks not extracted.
- **Plugin/extension system** — No mechanism for custom validators.

---

## d) TOTALLY FUCKED UP

### Build Was Broken (FIXED THIS SESSION)

The `go-output` dependency was upgraded to v0.6.3 in commit `80b78b4` but the import paths weren't updated. `MarshalYAML`, `NewCSVWriter`, and `CSVWriter` were moved to sub-packages:

| Symbol         | Old Import  | New Import                |
| -------------- | ----------- | ------------------------- |
| `MarshalYAML`  | `go-output` | `go-output/serialization` |
| `NewCSVWriter` | `go-output` | `go-output/delimited`     |
| `CSVWriter`    | `go-output` | `go-output/delimited`     |

**Fixed in this session** by updating imports in `pkg/output/output.go` and running `go mod tidy`.

### Skip Directive False Positives

Skip directives (`//nolint`, `// skip-validate`) are checked on **every line including inside code blocks**. A code block containing `//nolint` as actual code content will be incorrectly skipped. This is a **latent bug** in `pkg/extractor.go:104`.

### Dead Configuration Fields

`ContextConfig.MaxFiles` and `ContextConfig.MaxBlocksPerFile` are stored but **never used** by `Build()`. They duplicate fields on `FileValidator`. Dead code in `pkg/context.go`.

### `docs/DOMAIN_LANGUAGE.md` Is Empty

Entirely a boilerplate template with placeholder entries. No actual domain terms defined. This file has been empty since creation.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Break the `pkg/types` ↔ `pkg/languages` dependency cycle** — Proposal exists at `docs/modularization/PROPOSAL.md`: move `Language` branded type to `pkg/types/`. Not yet executed.
2. **Add line/column info to tree-sitter errors** — Currently just "code contains parse errors". Tree-sitter provides position data that's being ignored.
3. **Fix skip directive scope** — Only check directives outside code blocks, not inside.

### Code Quality

4. **Inconsistent `os.Exit` usage in main.go** — `osExit` variable exists for testability but isn't used consistently (some calls use raw `os.Exit`).
5. **`validatePath` swallows errors** — Returns nil even on failure, so overall exit code can be 0 when paths fail.
6. **`Registry` not thread-safe** — Safe by current usage pattern, but fragile. Add `sync.RWMutex`.
7. **`TruncateForError` truncates by bytes, not runes** — Could split multi-byte UTF-8.
8. **Tree-sitter validator registration errors silently ignored** — `_ = r.Register(...)` with no feedback.
9. **Hardcoded ANSI escape codes** in `printTableHeaderTo`/`printErrorEntry` instead of using a color library.

### Testing

10. **No output content assertions** — `PrintReport` tests are smoke tests (verify no panic, but don't check actual output).
11. **No `main()` end-to-end test** — CLI binary is never executed in tests.
12. **No `ExtractCodeBlocks` (multi-language) test** — Only `ExtractGoCodeBlocks` is tested.
13. **Missing `b.ReportAllocs()`** in all 6 benchmarks.
14. **No concurrent access tests** for Registry or Validator.

### Documentation & Release

15. **Fill in `DOMAIN_LANGUAGE.md`** — Define actual terms: Code Block, Validation Result, Skip Directive, Language Registry, etc.
16. **Bump version** — Unreleased section in CHANGELOG has months of changes. Should be v0.2.0 or v1.0.0.
17. **Fix `CONTRIBUTING.md`** — Remove dead references to nonexistent files.
18. **Enable Homebrew/Scoop uploads** in GoReleaser — or remove the configs if not ready.
19. **Add `--version` flag** — ldflags already inject the data, just needs CLI exposure.
20. **Add self-validation step to CI** — Run the tool on its own markdown docs.

### CI

21. **Add `go mod tidy` check** to CI — would have caught the broken build earlier.
22. **Add Nix flake check** to CI — `nix flake check` runs format + build + test.
23. **No OS matrix** — CI only runs on ubuntu-latest.

---

## f) Top 25 Things We Should Get Done Next

| #   | Priority | Item                                                                | Impact | Effort |
| --- | -------- | ------------------------------------------------------------------- | ------ | ------ |
| 1   | P0       | Add `--version` flag (#1 consumer complaint)                        | High   | Low    |
| 2   | P0       | Fix skip directive false positives (check only outside code blocks) | High   | Low    |
| 3   | P0       | Add `go mod tidy` check to CI                                       | High   | Low    |
| 4   | P0       | Bump version to v0.2.0 and release                                  | High   | Low    |
| 5   | P1       | Break `pkg/types` ↔ `pkg/languages` dependency cycle                | High   | Medium |
| 6   | P1       | Add tree-sitter error line/column info                              | High   | Medium |
| 7   | P1       | Add `--exclude` / `--ignore` patterns                               | High   | Medium |
| 8   | P1       | Add output content assertions in tests                              | Medium | Medium |
| 9   | P1       | Add self-validation step to CI                                      | Medium | Low    |
| 10  | P1       | Fill in `DOMAIN_LANGUAGE.md` with actual domain terms               | Medium | Low    |
| 11  | P1       | Fix `validatePath` error swallowing in main.go                      | Medium | Low    |
| 12  | P1       | Add `main()` end-to-end CLI test                                    | Medium | Medium |
| 13  | P1       | Add config file support (`.md-go-validator.yaml`)                   | High   | Medium |
| 14  | P2       | Fix inconsistent `os.Exit` vs `osExit` in main.go                   | Low    | Low    |
| 15  | P2       | Make `Registry` thread-safe with `sync.RWMutex`                     | Low    | Low    |
| 16  | P2       | Fix `TruncateForError` to truncate by runes                         | Low    | Low    |
| 17  | P2       | Log tree-sitter validator registration errors instead of ignoring   | Low    | Low    |
| 18  | P2       | Replace hardcoded ANSI codes with go-output color library           | Low    | Low    |
| 19  | P2       | Fix `CONTRIBUTING.md` dead references                               | Low    | Low    |
| 20  | P2       | Add `b.ReportAllocs()` to all benchmarks                            | Low    | Low    |
| 21  | P2       | Add JavaScript support via tree-sitter                              | Medium | Low    |
| 22  | P2       | Add `~~~` tilde fence support in extractor                          | Medium | Low    |
| 23  | P3       | Add OS matrix to CI (linux/darwin/windows)                          | Low    | Low    |
| 24  | P3       | Enable Homebrew/Scoop uploads or remove configs                     | Low    | Low    |
| 25  | P3       | Add Nix flake check to CI                                           | Low    | Low    |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Why was the go-output dependency upgraded to v0.6.3 without updating the import paths?**

Commit `80b78b4` ("chore(deps): update all dependencies to latest versions") bumped go-output to v0.6.3 and gotreesitter to v0.20.0, but the import paths in `pkg/output/output.go` still referenced the old API (`output.MarshalYAML`, `output.NewCSVWriter`, `output.CSVWriter`) which had been moved to sub-packages (`serialization`, `delimited`) in that version.

This means the build has been broken since that commit. Was this commit never tested? Was there a `go.work` with a local replace directive that masked the issue? The AGENTS.md mentions "Known issue: nix build fails due to go-output API mismatch (go.work uses local go-output with newer API)" — suggesting a local go.work file existed at some point that made the build work locally but not in CI. **That go.work file is now gone**, so the broken imports surfaced.

---

## Coverage Breakdown

| Package         | Coverage | Change (since last report) |
| --------------- | -------- | -------------------------- |
| `pkg`           | 87.5%    | +2.9% (was 84.6%)          |
| `pkg/code`      | 93.8%    | -6.2% (was 100%)           |
| `pkg/languages` | 90.6%    | -1.9% (was 92.5%)          |
| `pkg/output`    | 91.5%    | 0% (stable)                |
| `pkg/types`     | 93.6%    | +0.8% (was 92.8%)          |
| `pkg/testutil`  | 86.8%    | N/A (not tracked before)   |
| `cmd`           | 70.9%    | 0% (stable)                |

---

## Changes This Session

1. **Fixed broken build** — Updated imports in `pkg/output/output.go`:
   - `output.MarshalYAML` → `serialization.MarshalYAML`
   - `output.NewCSVWriter` → `delimited.NewCSVWriter`
   - `output.CSVWriter` → `delimited.CSVWriter`
2. **Ran `go mod tidy`** — Added `go-output/delimited` and `go-output/serialization` as direct dependencies, resolved all transitive deps.
3. **Bumped gotreesitter** — v0.20.0 → v0.20.1 (from tidy).
4. **Verified** — Build passes, all 7 packages pass tests, 0 linter issues.

---

_Waiting for further instructions._
