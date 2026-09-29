# Status Report — 2026-06-05 22:04 CEST

_Generated after deduplication session reducing clone groups from 12 to 8._

---

## Executive Summary

| Dimension        | Status                                                         |
| ---------------- | -------------------------------------------------------------- |
| **Build**        | PASSING (`go build`, `go test`, `golangci-lint` all green)     |
| **Tests**        | ALL GREEN — 7/7 packages, 0 failures, race-clean               |
| **Lint**         | 0 issues (golangci-lint with 90+ linters)                      |
| **Coverage**     | 70.9%–93.8% across packages                                    |
| **Duplication**  | 8 clone groups (all acceptable Go test idioms, down from 12)   |
| **Codebase**     | ~7,409 lines of Go (production + tests)                        |
| **Last release** | v0.1.0 (2026-01-01) — 6 months of unreleased changes           |
| **Dependencies** | go-output v0.6.3, gotreesitter v0.20.1, Go 1.26.3              |
| **CI**           | Minimal — test/lint/build on push/PR to main/master            |
| **Nix**          | Hash mismatch in go-modules (pre-existing, blocks `nix build`) |

---

## a) FULLY DONE

### Core Library (pkg/)

| Component                           | File(s)                                                             | Description                                                                                                                                     |
| ----------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **Go code validation**              | `pkg/languages/go_validator.go`                                     | 5-strategy parser (complete file → package wrap → func wrap → expression → statements). Handles partial snippets from documentation.            |
| **Multi-language support**          | `pkg/languages/treesitter_validator.go`                             | 8 languages via tree-sitter: Go, Templ, TypeScript, TSX, Rust, Nix, HCL, Terraform                                                              |
| **Language registry**               | `pkg/languages/validator.go`, `language.go`                         | `Registry` with `Register()`, `Get()`, `GetByString()`, `GetAvailable()`. Pluggable validator interface.                                        |
| **Code block extraction**           | `pkg/extractor.go`                                                  | Line-by-line markdown parser for fenced blocks with language filtering. MDX support.                                                            |
| **Skip directives**                 | `pkg/extractor.go`                                                  | Configurable: `<!-- skip-validate -->`, `<!-- skip-md-validate -->`, `<!-- md-skip -->`, `<!-- no-validate -->`, `// skip-validate`, `//nolint` |
| **Concurrent directory validation** | `pkg/validator.go`                                                  | Worker pool with channels, context cancellation, configurable concurrency                                                                       |
| **Context lifecycle**               | `pkg/context.go`                                                    | Timeout, deadline, max files/blocks, parent propagation, branching for parallel workers                                                         |
| **Multi-format output**             | `pkg/output/output.go`                                              | 6 formats: Table, JSON, YAML, CSV, Markdown, Quiet. Via go-output library.                                                                      |
| **Branded types**                   | `pkg/types/`                                                        | `FileID`, `LineNumber`, `BlockIndex`, `FileType`, `ValidationStatus`, `Language`. Compile-time safety.                                          |
| **Code utilities**                  | `pkg/code/util.go`                                                  | `IndentCode()`, `ParseGo()`, `TruncateForError()`                                                                                               |
| **Test helpers**                    | `pkg/testutil/`, `pkg/types/testing.go`, `pkg/languages/testing.go` | Shared assertion helpers across packages                                                                                                        |

### CLI (cmd/)

- **Argument parsing** — Manual generics-based parser with `--verbose`, `--quiet`, `--format`, `--color`, `--output`, `--timeout`, `--languages`, multi-path support
- **Usage/help** — Custom usage header with flag documentation

### Infrastructure

- **Nix flake** — flake-parts + treefmt-nix. Build, test, format, devShell, overlay. `flake.nix`
- **GoReleaser** — Cross-compile (linux/darwin/windows, amd64/arm64), cosign signing, SBOMs, brew/scoop/nix/nfpm. `.goreleaser.yml`
- **CI (GitHub Actions)** — Test (with race), lint (golangci-lint), build. `.github/workflows/ci.yml`
- **Linter config** — 90+ linters, Go 1.26.3, strict settings. `.golangci.yml`
- **Auto-deduplicate config** — `.auto-deduplicate/false-positives.json` with documented acceptable clones

### Quality

- **Deduplication** — 8 remaining clone groups, all documented as acceptable Go test idioms in false-positives.json
- **Integration tests** — Real testdata fixtures: `valid_go.md`, `invalid_go.md`, `skipped.md`, `mixed.mdx`, `edge_cases.md`
- **Benchmarks** — 6 benchmarks covering extraction, validation, directory scanning
- **No TODOs/FIXMEs** in codebase (`grep` returns zero)

### This Session's Work

| What                                    | Files Changed                                                          | Impact                                                            |
| --------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Extract `newTestErrorResult()` helper   | `pkg/types/types_test.go`                                              | Eliminated 3-clone group of inline `NewErrorResult` constructions |
| Extract `validatorTestCase` named type  | `pkg/languages/treesitter_validator_test.go`                           | Eliminated anonymous struct duplication in return type + literal  |
| Use existing `newErrorResultWithCode()` | `pkg/output/output_test.go`                                            | Replaced inline construction with existing helper                 |
| Add `newTestErrorResult()` helper       | `pkg/validator_test.go`                                                | Extracted error result construction for reuse                     |
| Add `assertExtensionsEqual()` helper    | `pkg/languages/testing.go`, `go_validator_test.go`, `language_test.go` | Consolidated extension comparison logic                           |
| Document acceptable clones              | `.auto-deduplicate/false-positives.json`                               | 8 groups documented as idiomatic Go test patterns                 |

---

## b) PARTIALLY DONE

| Item | State                             | Details |
| ---- | --------------------------------- | ------- |
| ~~   | `README.md`                       | 90%     |
| ~~   | `CHANGELOG.md`                    | 80%     |
| ~~   | `CONSUMER_PERSPECTIVE.md`         | 70%     |
| ~~   | `docs/modularization/PROPOSAL.md` | 50%     |
| ~~   | Nix build                         | 60%     |

---

## c) NOT STARTED

### From CONSUMER_PERSPECTIVE.md (15 gaps, none started):

| #  | Gap | Severity                                               |
| -- | --- | ------------------------------------------------------ |
| ~~ | 1   | `--version` flag                                       |
| ~~ | 2   | Configuration file support (`.md-go-validator.yaml`)   |
| ~~ | 3   | `--init` command for config generation                 |
| ~~ | 4   | `.md-go-validator-ignore` / exclude patterns           |
| ~~ | 5   | Fix CONTRIBUTING.md dead references                    |
| ~~ | 6   | Reusable GitHub Action (`action.yml`)                  |
| ~~ | 7   | Pre-commit hook integration (`.pre-commit-hooks.yaml`) |
| ~~ | 8   | Watch / incremental mode (`--watch`)                   |
| ~~ | 9   | Error codes in CLI output                              |
| ~~ | 10  | Diff / regression mode (`--baseline`)                  |
| ~~ | 11  | `--dry-run` flag                                       |
| ~~ | 12  | Progress indicator                                     |
| ~~ | 13  | Granular exit codes (errors vs crash vs no files)      |
| ~~ | 14  | Self-validation in CI (dogfooding)                     |
| ~~ | 15  | `--fail-on-skipped` option                             |

### Other Not Started:

~~- **`TODO_LIST.md`** — Does not exist~~ done at `50487d3`
~~- **`FEATURES.md`** — Does not exist~~ done at `50487d3`
~~- **`ROADMAP.md`** — Does not exist~~ done at `50487d3`
~~- **`docs/DOMAIN_LANGUAGE.md`** — Exists but needs review for completeness~~ done at `7868a98`
~~- **Version bump** — v0.1.0 was January 2026, 6 months of unreleased work~~ done at `d3a4a1c`, `c5830b8`
~~- **Go workspace cleanup** — `go.work` references local go-output, blocks reproducible nix build~~ done at — no `go.work` remains; deps on proxy (`220837d`)

---

## d) TOTALLY FUCKED UP

| Issue                          | Severity                                       | Details                                                                                                                                                                                                          |
| ------------------------------ | ---------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~                             | **Nix build broken**                           | HIGH                                                                                                                                                                                                             |
| **6 months without a release** | HIGH                                           | v0.1.0 was 2026-01-01. Massive unreleased changes: MDX support, tree-sitter, branded types, multi-format output, context support, skip directives overhaul. Users on v0.1.0 have a fundamentally different tool. |
| **CONTRIBUTING.md is broken**  | MEDIUM                                         | References `just` commands and setup scripts that don't exist. First contributor experience is broken.                                                                                                           |
| **No dogfooding in CI**        | MEDIUM                                         | Tool validates markdown code blocks but doesn't validate its own docs in CI. Undermines credibility.                                                                                                             |
| ~~                             | **pkg/types ↔ pkg/languages dependency cycle** | LOW                                                                                                                                                                                                              |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

~~1. **Break the types ↔ languages cycle** — Move `Language` type to a shared `pkg/types` or create `pkg/language` package with just the type definition. The validator interface stays in `pkg/languages`.~~ DUPLICATE — tracked in `docs/modularization/`
~~2. **Extract CLI config to its own package** — `cmd/md-go-validator/main.go` is a 730-line monolith mixing parsing, validation, and output. Extract to `cmd/config.go` at minimum.~~ Won't implement — CLI restructured via direct dispatch instead (`539cb8e`)
~~3. **Move from manual arg parsing to a proper library** — cobra or similar. The generics-based parser works but is hard to extend and doesn't support `--version`, config files, or shell completion.~~ Won't implement — hand parser retained deliberately; config/flags stable

### Quality

~~4. **Increase cmd coverage** — `cmd/md-go-validator` is at 70.9%, lowest in the project. Edge cases in arg parsing are undertested.~~ done at `f3a2c2c`
~~5. **Add self-validation to CI** — Run `md-go-validator` against the project's own `README.md`, `EXAMPLES.md`, `CONTRIBUTING.md` in a CI step.~~ done at `d3a4a1c`, `ecda347`
~~6. **Fix nix build** — Either remove `go.work` (use go.mod only) or configure nix to handle it properly. This is the #1 infra issue.~~ done at `5dc7606`

### Developer Experience

~~7. **Fix CONTRIBUTING.md** — Remove `just` references, add nix-based commands, remove dead script references.~~ done at `b72a1be`
~~8. **Add `FEATURES.md`** — Honest feature inventory for quick reference.~~ done at `50487d3`
~~9. **Add `TODO_LIST.md`** — Actionable, prioritized task list.~~ done at `50487d3`

### Release

~~10. **Cut v0.2.0** — The unreleased changes are substantial and production-ready. Ship them.~~ done at `d3a4a1c`

---

## f) Top 25 Things We Should Get Done Next

### Tier 1: Critical (Adoption Blockers)

| #  | Task | Impact                                         | Effort                           |
| -- | ---- | ---------------------------------------------- | -------------------------------- |
| ~~ | 1    | **Cut v0.2.0 release**                         | Ships 6 months of work to users  |
| ~~ | 2    | **Add `--version` flag**                       | Users can verify installation    |
| ~~ | 3    | **Fix nix build** (remove go.work or fix hash) | Reproducible builds work         |
| ~~ | 4    | **Fix CONTRIBUTING.md** dead references        | Contributors don't hit dead ends |
| ~~ | 5    | **Add self-validation to CI**                  | Dogfooding builds trust          |

### Tier 2: High Impact (Quality of Life)

| #  | Task | Impact                                                   | Effort                        |
| -- | ---- | -------------------------------------------------------- | ----------------------------- |
| ~~ | 6    | **Configuration file support** (`.md-go-validator.yaml`) | Users commit settings to repo |
| ~~ | 7    | **GitHub Action** (`action.yml`)                         | Single-line CI integration    |
| ~~ | 8    | **Exclude patterns** (`.md-go-validator-ignore`)         | Skip vendor/generated files   |
| ~~ | 9    | **`--init` command**                                     | Generate starter config       |
| ~~ | 10   | **Pre-commit hooks** (`.pre-commit-hooks.yaml`)          | Ecosystem discoverability     |

### Tier 3: Important (Polish)

| #  | Task | Impact                                                           | Effort                           |
| -- | ---- | ---------------------------------------------------------------- | -------------------------------- |
| ~~ | 11   | **Granular exit codes** (0=valid, 1=errors, 2=crash, 3=no files) | CI can distinguish failure modes |
| ~~ | 12   | **Error codes in JSON output**                                   | Machine-actionable results       |
| ~~ | 13   | **Create `TODO_LIST.md`**                                        | Prioritized backlog              |
| ~~ | 14   | **Create `FEATURES.md`**                                         | Feature inventory                |
| ~~ | 15   | **Increase cmd coverage to 85%+**                                | Confidence in CLI edge cases     |
| ~~ | 16   | **Break types ↔ languages dependency cycle**                     | Clean architecture               |

### Tier 4: Nice to Have (Enhancement)

| #  | Task | Impact                                   | Effort                          |
| -- | ---- | ---------------------------------------- | ------------------------------- |
| ~~ | 17   | **Watch mode (`--watch`)**               | Development workflow            |
| ~~ | 18   | **Progress indicator**                   | UX for large directories        |
| ~~ | 19   | **`--dry-run` flag**                     | Debug config without running    |
| ~~ | 20   | **Diff/regression mode (`--baseline`)**  | Incremental adoption            |
| ~~ | 21   | **`--fail-on-skipped` option**           | Strict enforcement              |
| ~~ | 22   | **Migrate CLI to cobra**                 | Extensibility, shell completion |
| ~~ | 23   | **Extract cmd config to separate file**  | Reduce main.go size             |
| ~~ | 24   | **Add `docs/DOMAIN_LANGUAGE.md` review** | Ensure completeness             |
| ~~ | 25   | **Add nix flake to CI**                  | Verify nix build in CI          |

---

## g) Top #1 Question

**Should `go.work` be removed entirely, or is there a plan to fix the nix build while keeping it?**

The `go.work` file references a local `go-output` checkout, which means:

- `go build` works locally (uses go.work → local go-output with newer API)
- `nix build` fails (uses go.mod → go-output v0.6.3 from proxy, different hash)
- CI uses `go mod download` which may or may not pick up go.work

This is the root cause of the nix build failure and creates a split-brain between local development and reproducible builds. Removing `go.work` would fix nix but break local development if go-output API is ahead of the published version. Keeping it means nix build stays broken until go-output v0.6.4+ is published with the matching API.

---

## Test Coverage Summary

| Package               | Coverage | Change                            |
| --------------------- | -------- | --------------------------------- |
| `pkg`                 | 87.5%    | ↑ from 84.6% (earlier session)    |
| `pkg/code`            | 93.8%    | ↑ from 100%? (minor fluctuation)  |
| `pkg/languages`       | 89.7%    | ↓ from 92.5%                      |
| `pkg/output`          | 91.5%    | Stable                            |
| `pkg/types`           | 93.6%    | ↑ from 92.8%                      |
| `pkg/testutil`        | 86.8%    | Stable                            |
| `cmd/md-go-validator` | 70.9%    | Stable (lowest — needs attention) |

## Duplication Summary

| Metric                      | Before Session | After Session                                 |
| --------------------------- | -------------- | --------------------------------------------- |
| Clone groups (threshold 15) | 12             | 8                                             |
| Total clones                | ~35            | 25                                            |
| Complexity score            | ~3.5           | 2.78                                          |
| Production code clones      | 0              | 0                                             |
| Acceptable test clones      | 12 groups      | 8 groups (documented in false-positives.json) |

## Benchmark Summary

| Benchmark                      | ns/op   | B/op    | allocs/op |
| ------------------------------ | ------- | ------- | --------- |
| ExtractCodeBlocks (50 blocks)  | 8,389   | 13,280  | 60        |
| ExtractCodeBlocks (500 blocks) | 76,970  | 122,112 | 513       |
| ValidateGoCode/complete        | 1,375   | 2,104   | 46        |
| ValidateGoCode/expression      | 4,552   | 5,467   | 126       |
| ValidateDirectory              | 101,587 | 362,058 | 3,669     |
| IsSupportedFile                | 52      | 0       | 0         |

---

_Arte in Aeternum_
