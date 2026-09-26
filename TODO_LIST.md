# TODO List

> Short-term, actionable, bounded work items, verified against the actual code on 2026-09-26
> (docs-health pass: all 43 historical `2026-0*` reports read, annotated, and harvested).
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## High Impact

| Task                                                        | Status       | Impact | Effort | Evidence                                                                                                                                          |
| ----------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Fix skip-directive scoping bug                              | 🔴 `TODO`    | High   | 30min  | `pkg/extractor.go:104` checks directives on every line, including code content and regular markdown — README's own `- `//nolint`` bullet currently skips the Library Usage Go example (reproduced 2026-09-26 via `--verbose` dogfood). Latent bug flagged 2026-06-05, still present. |
| File-read errors as Results (`ValidationStatusFileError`)   | 🔴 `TODO`    | High   | 3h     | Unreadable file still escalates to a batch error; report shows nothing for that file. See status report 2026-08-05_12-34 §c.2 and §f.6-7.         |
| Dogfood website docs in CI                                  | 🔴 `TODO`    | High   | 30min  | `ci.yml:63` dogfood validates only root docs; `website/src/content/docs/` (where the 2026-08-05 mixed-scope bug lived) is not validated by any workflow. |
| Add drift guard: go.mod vs flake input                      | 🔴 `TODO`    | High   | 1h     | Currently aligned (go-finding v1.2.0 both sides); guard prevents future go-finding skew. Only `replace`'d flake input (`package.nix` postPatch).   |
| Decide on postPatch replace directive                       | 🔵 `BLOCKED` | High   | -      | `package.nix:26`; overlay confirmed non-functional: go-finding is private (SSH URL, no GOPRIVATE in derivation), so overlay build fails without replace. Decision hinges on whether go-finding is made public. |
| Publish Homebrew tap                                        | 🔴 `TODO`    | High   | 30min  | `.goreleaser.yml:102` has `skip_upload: true`                                                                                                     |
| Clean up stray git tags                                     | 🔴 `TODO`    | High   | 15min  | Tags `v1.0.0`/`v1.1.0`/`v1.2.0` point at commits older than `v0.2.0`/`v0.3.0` — confusing version history (found 2026-09-26 docs-health pass).     |

## Medium Impact

| Task                                              | Status    | Impact | Effort | Evidence                                                                                |
| ------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------------------------------- |
| `--continue-on-error` / `--fail-fast` flags       | 🔴 `TODO` | Med    | 2h     | No batch-behavior control; status report 2026-08-05_12-34 §f.13                          |
| `--max-errors N` bail-out flag                    | 🔴 `TODO` | Med    | 1h     | Huge trees need a stop condition; 2026-08-05 reports §f.14                               |
| Per-file stderr logging on read errors            | 🔴 `TODO` | Med    | 30min  | Failed paths only visible inside joined error; 2026-08-05 §f.8                           |
| `collectSupportedFiles` log-and-skip resilience   | 🔴 `TODO` | Med    | 1h     | One bad walk entry still fails the whole collection; 2026-08-05 §f.9                     |
| WARNING line when results are partial             | 🔴 `TODO` | Med    | 30min  | Report should say "N files skipped due to errors"; 2026-08-05 §f.15                      |
| SARIF output for file-read errors                 | 🔴 `TODO` | Med    | 1h     | CI cannot see unreadable files; 2026-08-05 §f.17                                         |
| Add `--dry-run` flag                              | 🔴 `TODO` | Med    | 1h     | Show what would be validated without running                                            |
| Add progress indicator for large dirs             | 🔴 `TODO` | Med    | 2h     | No spinner/progress bar; blank terminal during scan                                      |
| Generate shell completions (bash/zsh/fish)        | 🔴 `TODO` | Med    | 1h     | No completion code exists                                                                |
| Document API stability (stable vs experimental)   | 🔴 `TODO` | Med    | 1h     | Library consumers lack guidance; `pkg/` exports                                          |
| `govulncheck` + `gosec` in devShell               | 🔴 `TODO` | Med    | 30min  | Not in `flake.nix` devShell (flagged 2026-06-26, still absent)                           |
| Add SECURITY.md + GitHub issue templates          | 🔴 `TODO` | Med    | 30min  | Neither exists (verified 2026-09-26)                                                     |
| Deterministic cancellation test                   | 🔴 `TODO` | Med    | 1h     | Current test is timing-based (50ms); cancel via callback after N results instead         |
| Targeted unit tests for CLI/output gaps           | 🔴 `TODO` | Med    | 3h     | `applyConfigFormat`, `--config` via parseArgs, SARIF structure, doublestar `**`, `HasSkipped`, `--save-baseline` e2e — all untested (grep-verified 2026-09-26) |
| Add CSP as HTTP header in `firebase.json`         | 🔴 `TODO` | Med    | 30min  | CSP only via `<meta>` tag patched by `fix-csp.mjs`; header is defense in depth           |

## Low Impact

| Task                                         | Status    | Impact | Effort | Evidence                                                                 |
| -------------------------------------------- | --------- | ------ | ------ | ------------------------------------------------------------------------ |
| Run `nix flake check --all-systems`          | 🔴 `TODO` | Low    | 10min  | Only x86_64-linux verified (2026-09-26 run skipped darwin/aarch64)       |
| `.envrc.example` + direnv note in CONTRIBUTING | 🔴 `TODO` | Low  | 15min  | `.envrc` is gitignored; other devs lack the GOEXPERIMENT fix             |
| Benchmark hygiene: `b.ReportAllocs()`        | 🔴 `TODO` | Low    | 15min  | Missing in all benchmarks in `pkg/benchmark_test.go` (verified 2026-09-26) |
| Rune-safe `TruncateForError`                 | 🔴 `TODO` | Low    | 30min  | `pkg/code/util.go` truncates by bytes; can split multi-byte UTF-8        |
| Lint config hygiene                          | 🔴 `TODO` | Low    | 15min  | `exhaustruct` deprecated → `exhaustruct_v5`; stale `legacyerrors` nolint directive (golangci-lint v2.13 warnings, 2026-09-26) |
| Tree-sitter error line/column                | 🔴 `TODO` | Low    | 2h     | Non-Go errors still say "code contains parse errors" (`treesitter_validator.go:64`) |
| Brand `ValidationError.Line/Column`          | 🔴 `TODO` | Low    | 20min  | Still plain `int` (`pkg/languages/validator.go:104`); propagate real column in `finding.go` |
| Add `go mod tidy` check to CI                | 🔴 `TODO` | Low    | 15min  | Would have caught the 2026-06-05 broken build earlier                    |
| Baseline path normalization (relative)       | 🔴 `TODO` | Low    | 1h     | `main.go:572` uses `filepath.Abs`; baselines not portable across machines |
| Raise `pkg/baseline` coverage (73%)          | 🔴 `TODO` | Low    | 1h     | Lowest-coverage package (2026-09-26 `go test -cover`)                    |
| Verify vendorHash stability hypothesis       | 🔴 `TODO` | Low    | 30min  | proxyVendor behavior; see status report 2026-07-07 §e                    |

---

<!-- Source: docs-health pass 2026-09-26. Harvested from status reports 2026-06-26 → 2026-08-05
     (f-sections) + verified against code the same day. Every item above was grep/read-checked;
     already-shipped report items were annotated in place under docs/*/archived/ instead. -->
