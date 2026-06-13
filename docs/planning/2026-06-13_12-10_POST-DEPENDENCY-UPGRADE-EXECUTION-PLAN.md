# Post-Dependency Upgrade Execution Plan

_Generated 2026-06-13 12:10 CEST after upgrading go-output to v0.10.0 and restoring the nix build._

---

## Pareto Breakdown: What Actually Moves the Needle?

### 1% Effort → 51% Impact

| Task | Why It Dominates |
| ---- | ---------------- |
| **Cut v0.2.0 release** | Six months of unreleased work (MDX, tree-sitter, branded types, multi-format output, context, skip directives) instantly becomes available to users. Everything else is polish on a still-unreleased artifact. |

### 4% Effort → 64% Impact (add the basics)

| Task | Why It Matters |
| ---- | -------------- |
| **Add `--version` flag** | Users cannot even verify they installed the right binary. Required for any release. |
| **Fix `CONTRIBUTING.md` dead references** | First-time contributors currently hit non-existent `just` commands. Low effort, high trust impact. |
| **Fix `flake.overlays.default` (`package.nix`)** | The overlay is advertised but broken. Anyone consuming this flake as an overlay gets a hard error. |
| **Add self-validation to CI** | The tool validates markdown code blocks but doesn't validate its own docs. Dogfooding is free credibility. |

### 20% Effort → 80% Impact (complete the developer/CI story)

Add the remaining high-leverage items that turn a working library into an adoptable product:

- Configuration file support (`.md-go-validator.yaml`)
- Exclude patterns (`.md-go-validator-ignore`)
- `--init` command
- Reusable GitHub Action (`action.yml`)
- Pre-commit hooks (`.pre-commit-hooks.yaml`)
- Granular exit codes
- Error codes in JSON output
- `TODO_LIST.md` and `FEATURES.md`
- Structured errors with `go-error-family`

---

## Comprehensive Plan (Sorted by Impact vs Effort)

Each task is designed to be **30–100 minutes** of focused work. Total: ~25 tasks.

| #  | Task                                                                 | Impact | Effort | Tier | Notes |
| -- | -------------------------------------------------------------------- | ------ | ------ | ---- | ----- |
| 1  | **Cut v0.2.0 release**                                               | 🔥🔥🔥 | S      | P0   | Tag + goreleaser. Unlocks all subsequent value. |
| 2  | **Add `--version` flag**                                             | 🔥🔥🔥 | S      | P0   | Read `main.version` ldflag; print and exit 0. Tests in `cmd/`. |
| 3  | **Fix `flake.overlays.default`**                                     | 🔥🔥   | S      | P0   | Either create `package.nix` or remove overlay export. |
| 4  | **Fix `CONTRIBUTING.md` dead references**                            | 🔥🔥   | S      | P0   | Replace `just` with `nix` commands; remove dead script refs. |
| 5  | **Add self-validation to CI**                                        | 🔥🔥   | S      | P0   | Run built binary against `README.md`, `EXAMPLES.md`, `CONTRIBUTING.md`. |
| 6  | **Create `FEATURES.md`**                                             | 🔥🔥   | S      | P1   | Honest inventory by status (DONE / PARTIAL / PLANNED). |
| 7  | **Create `TODO_LIST.md`**                                            | 🔥🔥   | S      | P1   | Short/mid-term actionable tasks, not vague ideas. |
| 8  | **Add granular exit codes**                                          | 🔥🔥   | S      | P1   | 0=valid, 1=errors, 2=crash, 3=no files found. |
| 9  | **Add error codes to JSON output**                                   | 🔥     | S      | P1   | Extend `types.ErrorEntry` with stable error code. |
| 10 | **Add `--init` command**                                             | 🔥🔥   | S      | P1   | Generate `.md-go-validator.yaml` with defaults. |
| 11 | **Configuration file support (`.md-go-validator.yaml`)**             | 🔥🔥🔥 | M      | P1   | Use `viper` or `koanf`. Merge file < env < CLI flags. |
| 12 | **Exclude patterns (`.md-go-validator-ignore`)**                     | 🔥🔥   | M      | P1   | Use `doublestar` for `.gitignore`-style matching. |
| 13 | **GitHub Action (`action.yml`)**                                     | 🔥🔥🔥 | M      | P1   | Composite action using released binary. Single-line CI integration. |
| 14 | **Pre-commit hooks (`.pre-commit-hooks.yaml`)**                      | 🔥🔥   | S      | P1   | Hook definition for ecosystem discoverability. |
| 15 | **Introduce `go-error-family` for structured errors**                | 🔥🔥   | M      | P1   | Classify errors: input/validation/internal/rejected. |
| 16 | **Refactor `Result` into a sum type**                                | 🔥🔥   | M      | P2   | `ValidResult` / `SkippedResult` / `ErrorResult` + interface. |
| 17 | **Create `pkg/config` domain type**                                  | 🔥🔥   | M      | P2   | Centralize defaults, validation, flag mapping. |
| 18 | **Break `pkg/types` ↔ `pkg/languages` cycle**                        | 🔥🔥   | M      | P2   | Move `Language` type to `pkg/language` or `pkg/types`. |
| 19 | **Increase `cmd/` test coverage to 85%+**                            | 🔥     | M      | P2   | Currently 70.9%; focus on flag parsing edge cases. |
| 20 | **Add `internal/` package boundary**                                 | 🔥     | M      | P2   | Move non-public implementation under `internal/`. |
| 21 | **Remove stray `md-go-validator` binary + update `.gitignore`**      | 🔥     | S      | P2   | Prevents accidental commit. |
| 22 | **Add progress indicator for large directories**                     | 🔥     | S      | P3   | `cheggaaa/pb` or bubbletea; respect `--quiet`. |
| 23 | **Add `--dry-run` flag**                                             | 🔥     | S      | P3   | List files/blocks that would be validated. |
| 24 | **Implement watch mode (`--watch`)**                                 | 🔥     | M      | P3   | Use `fsnotify`; debounce and re-run. |
| 25 | **Migrate CLI to `cobra` + `viper`**                                 | 🔥🔥   | L      | P3   | Enables subcommands, completion, config, `--version`. Replaces hand parser. |

---

## Detailed Breakdown (Fine Granularity, ≤15 min tasks)

Only the top 10 tasks are decomposed here. Each sub-task should fit in ~15 minutes.

### Task 1: Cut v0.2.0 release

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 1.1 | Update `CHANGELOG.md` unreleased section into v0.2.0 | Legal/clean release notes |
| 1.2 | Bump any hardcoded version strings | Ensure consistency |
| 1.3 | Run full verification: `go test -race`, `golangci-lint`, `nix flake check` | Release gate |
| 1.4 | Tag `v0.2.0` and push | Git release trigger |
| 1.5 | Run `goreleaser release` | Artifacts, brew, scoop, nix |

### Task 2: Add `--version` flag

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 2.1 | Add `--version` to flag constants and help text | Discoverability |
| 2.2 | Parse `--version` before other flags | Early exit path |
| 2.3 | Print `main.version` (fallback to `dev`) | Works with ldflags |
| 2.4 | Add tests for `--version` and missing version | Coverage |
| 2.5 | Update `README.md` examples | Docs |

### Task 3: Fix `flake.overlays.default`

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 3.1 | Decide: create `package.nix` or remove overlay | Product decision |
| 3.2 | If creating: extract package expression from `flake.nix` | DRY |
| 3.3 | If removing: delete `flake.overlays.default` | Simplicity |
| 3.4 | Verify `nix flake check` and `nix build .#` still pass | No regression |
| 3.5 | Update `AGENTS.md` and README if overlay changes | Docs |

### Task 4: Fix `CONTRIBUTING.md` dead references

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 4.1 | Replace all `just` commands with `nix` equivalents | Accuracy |
| 4.2 | Remove references to non-existent setup scripts | Accuracy |
| 4.3 | Add nix dev shell instructions | Onboarding |
| 4.4 | Add golangci-lint and test commands | Completeness |
| 4.5 | Verify no dead internal links | Quality |

### Task 5: Add self-validation to CI

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 5.1 | Add CI step to build binary | Need artifact |
| 5.2 | Run binary against `README.md`, `EXAMPLES.md`, `CONTRIBUTING.md` | Dogfooding |
| 5.3 | Use `--format=json` and fail on errors | CI-friendly |
| 5.4 | Verify in a test PR or local act | Confidence |

### Task 6: Create `FEATURES.md`

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 6.1 | Inventory all features from code | Completeness |
| 6.2 | Categorize: DONE / PARTIAL / PLANNED / WORTH CONSIDERING | Honesty |
| 6.3 | Cross-check against `CONSUMER_PERSPECTIVE.md` gaps | Consistency |
| 6.4 | Add to `AGENTS.md` if appropriate | Discoverability |

### Task 7: Create `TODO_LIST.md`

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 7.1 | Pull items from this plan and consumer gaps | Source of truth |
| 7.2 | Mark status and owner (if any) | Accountability |
| 7.3 | Keep scoped to short/mid-term (next 1–3 months) | Actionable |
| 7.4 | Add links to relevant files/issues | Navigation |

### Task 8: Add granular exit codes

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 8.1 | Define exit-code enum/consts | Type safety |
| 8.2 | Map validation outcomes to codes | Logic |
| 8.3 | Update main exit logic | Wiring |
| 8.4 | Add tests for each exit code | Coverage |
| 8.5 | Document in README | User contract |

### Task 9: Add error codes to JSON output

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 9.1 | Add `Code` field to `types.ErrorEntry` | Data model |
| 9.2 | Classify errors in validators (syntax/not-found/etc.) | Mapping |
| 9.3 | Update `BuildReportData` | Report wiring |
| 9.4 | Update JSON tests | Coverage |
| 9.5 | Document error-code table | API contract |

### Task 10: Add `--init` command

| #  | Sub-task | Rationale |
| -- | -------- | --------- |
| 10.1 | Define default config struct/defaults | Reuse config type |
| 10.2 | Write YAML to `.md-go-validator.yaml` | Output |
| 10.3 | Handle existing file (error or overwrite flag) | Safety |
| 10.4 | Add tests | Coverage |
| 10.5 | Document in README | Usage |

---

## Type Model & Architecture Improvements

These cut across multiple tasks above and should be kept in mind while implementing:

| Current Pain | Proposed Model | Affected Tasks |
| ------------ | -------------- | -------------- |
| `config` struct in `main.go` built by hand | `pkg/config.Config` with constructor + validation | 10, 11, 17, 25 |
| `types.Result` is one struct with optional `Error` | Sum type: `ValidResult | SkippedResult | ErrorResult` | 9, 16 |
| `pkg/types` imports `pkg/languages` for `Language` | Move `Language` to `pkg/language` (type-only package) | 18 |
| Direct `os` calls in validator and CLI | Use `io/fs.FS` or `afero.Fs` | 12, 24 |
| Plain `fmt.Errorf` | `go-error-family` classified errors | 15, 9 |
| Hardcoded ANSI codes in table output | `chroma` or `lipgloss` for rendering | 25 (CLI refactor) |

---

## Existing Code to Reuse

Before implementing from scratch, leverage:

| Existing Asset | How to Reuse |
| -------------- | ------------ |
| `pkg/context.go` | Context branching is already in place for `--watch` debounce/re-run |
| `pkg/validator.go` | `FileValidator` options pattern can be driven by `pkg/config.Config` |
| `pkg/output/output.go` | Adding JSON error codes is a localized change to `types.ErrorEntry` + `marshalReport` |
| `pkg/types/testing.go` | Test helpers already exist for results; extend for new result subtypes |
| `flake.nix` package expression | Extract into `package.nix` for the overlay fix |
| `.goreleaser.yml` | Release pipeline already configured for v0.2.0 |

---

## D2 Execution Graph

```d2
direction: down

P0: {
  label: |md
    **P0 — Ship Existing Value**
    (1% → 51% impact)
  |
  release: Cut v0.2.0 release
  version: Add --version flag
  overlay: Fix flake overlay
  contrib: Fix CONTRIBUTING.md
  dogfood: Add CI self-validation
}

P1: {
  label: |md
    **P1 — Developer/CI Adoption**
    (4% → 64% impact)
  |
  features: Create FEATURES.md
  todo: Create TODO_LIST.md
  exit: Granular exit codes
  errcodes: JSON error codes
  init: --init command
  config: Config file support
  ignore: Exclude patterns
  action: GitHub Action
  hooks: Pre-commit hooks
  errors: go-error-family errors
}

P2: {
  label: |md
    **P2 — Architecture & Quality**
    (20% → 80% impact)
  |
  sumtype: Result sum type
  configpkg: pkg/config domain type
  cycle: Break types/languages cycle
  coverage: Increase cmd coverage
  internal: internal/ boundary
  cleanup: Remove stray binary
}

P3: {
  label: |md
    **P3 — Polish & Extensibility**
    (remaining 20%)
  |
  progress: Progress indicator
  dryrun: --dry-run
  watch: --watch mode
  cobra: Migrate to cobra/viper
}

P0.release -> P0.version -> P1.features
P0.version -> P1.init
P0.version -> P1.config
P1.config -> P1.ignore
P1.config -> P2.configpkg
P2.configpkg -> P2.sumtype
P2.sumtype -> P2.cycle
P1.errors -> P2.sumtype
P1.exit -> P2.sumtype
P1.errcodes -> P2.sumtype
P2.cycle -> P3.cobra
P2.internal -> P3.cobra
P1.hooks -> P3.cobra
```

---

## Decision Log

| Decision | Rationale |
| -------- | --------- |
| Do **not** migrate to `cobra`/`viper` in P0/P1 | High effort; current parser works. Defer until subcommands (`init`, `watch`) make the complexity worth it. |
| Do **not** introduce `afero` immediately | `io/fs.FS` is built-in and sufficient for read-only validation; adopt only if virtual filesystem becomes critical. |
| Move `Language` to `pkg/language` (recommendation) | Smallest, cleanest cycle break. Keeps `pkg/types` focused on report/result types. |
| Use `viper` for config (recommendation) | Mature, integrates cleanly with `cobra` later, supports env + file + flag precedence. |
| Use `doublestar` for ignore patterns (recommendation) | De-facto standard for `.gitignore`-style glob matching in Go. |
