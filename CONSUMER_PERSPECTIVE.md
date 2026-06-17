# Consumer Perspective — What's Missing

A brutally honest assessment of what a new user would find lacking, confusing, or blocking adoption.

---

## Recently Resolved

| Item                           | Resolution                                                                        |
| ------------------------------ | --------------------------------------------------------------------------------- |
| ✅ `--version` flag (#1)       | Added `--version` / `-V`                                                          |
| ✅ Error codes in output (#9)  | `errorCode` field now in JSON/YAML output (syntax, not_available, not_registered) |
| ✅ Summary exit codes (#13)    | 0=success, 1=validation errors, 2=tool/usage errors                               |
| ✅ Self-validation in CI (#14) | CI dogfoods — runs validator on its own docs                                      |
| ✅ STDIN support               | `cat file.md \| md-go-validator -`                                                |
| ✅ JSON schema                 | Output contract documented at `docs/json-schema.json`                             |

---

## Critical Gaps (Adoption Blockers)

### 1. ~~No `--version` Flag~~ ✅ RESOLVED

`--version` / `-V` now prints the version.

### 2. No Configuration File Support

There is no `.md-go-validator.yaml`, `.md-go-validator.toml`, or similar config file. For repos with many docs, typing `-l go,typescript,rust --timeout 30s -f json -o results.json` every time is painful. Consumers expect to commit a config file to their repo so `md-go-validator .` "just works" with their settings.

### 3. No `--init` Command

No way to generate a starter config file. Consumers discover settings by reading the README, not by running `md-go-validator --init`.

### 4. No `.md-go-validator-ignore` / Exclude Patterns

There is no way to exclude directories or files (e.g., `vendor/`, `node_modules/`, generated docs, third-party markdown). A consumer with a large repo will get false positives from vendored or auto-generated markdown they don't control.

### 5. CONTRIBUTING.md References Nonexistent Files

`CONTRIBUTING.md` references `./CONTRIBUTING-setup.sh` and `just install` / `just lint` / `just test`, but there is no `justfile` (marked deprecated in AGENTS.md) and no setup script. A new contributor following the guide will hit dead ends immediately.

---

## Major Gaps (Quality of Life)

### 6. No GitHub Action

The README shows a manual CI integration snippet, but there is no reusable GitHub Action (`action.yml`). Consumers want `uses: LarsArtmann/md-go-validator@v1` — not a 4-step YAML block with `go install`. This is the single highest-leverage missing feature for adoption.

### 7. No Pre-commit Hook Integration

No `.pre-commit-hooks.yaml` file. Many teams use pre-commit.com to validate before commits. Adding this file makes the tool discoverable to a huge ecosystem.

### 8. No Watch / Incremental Mode

No `--watch` flag for development workflows. When writing docs, consumers want instant feedback as they save files, not a manual re-run.

### 9. ~~No Error Codes or Machine-Actionable Output~~ ✅ RESOLVED

`errorCode` is now exposed in JSON/YAML output (values: `syntax`, `not_available`, `not_registered`, `unknown`). See `docs/json-schema.json`.

### 10. No Diff / Regression Mode

No `--baseline` or `--compare` flag. When adding this tool to an existing project with many broken code blocks, a consumer has no way to say "I know these 12 are broken, only fail on NEW errors." This makes incremental adoption painful.

---

## Moderate Gaps (Polish & Trust)

### 11. No `--dry-run` Flag

No way to see what _would_ be validated without actually running validation. Useful for debugging config, excludes, and language filters.

### 12. No Progress Indicator

When validating a large directory, there is no progress bar, spinner, or file count indicator. The user stares at a blank terminal until results appear. `-v` prints per-block info, but there is no summary progress.

### 13. ~~No Summary Exit Codes~~ ✅ RESOLVED

Exit codes now distinguish: 0=success, 1=validation errors, 2=tool/usage errors (file not found, bad flags, etc.).

### 14. ~~No Self-Validation~~ ✅ RESOLVED

CI now dogfoods — the validator runs against its own docs in the test/build pipeline.

### 15. No `--fail-on-skipped` Option

No way to enforce zero skip directives. Teams that want strict validation (no `<!-- skip-validate -->` allowed) cannot enforce it.

---

## Minor Gaps (Nice to Have)

### 16. No Shell Completions

No generation of bash/zsh/fish completion scripts. The CLI has enough flags that tab completion would help.

### 17. No `--languages` Discovery Command

No way to list supported languages from the CLI. A consumer must read the README to discover what `-l` accepts. A `md-go-validator --languages` or `md-go-validator list-languages` command would help.

### 18. Homebrew Tap Not Published

`.goreleaser.yml` has `skip_upload: true` on the brew and scoop sections. Consumers cannot `brew install md-go-validator`.

### 19. No API Stability Documentation

No statement about which parts of the Go API are stable vs. experimental. Library consumers (`pkg/` importers) don't know what might break between versions.

### 20. DESCRIPTION Mismatch

`.goreleaser.yml` describes the tool as "Markdown validation tool for Go projects" — but it validates multi-language code blocks, not Go-specific markdown. The description undersells the tool and confuses potential users who think it only handles Go.

---

## Prioritized Impact

| Priority | Item                       | Consumer Impact              | Status  |
| -------- | -------------------------- | ---------------------------- | ------- |
| ~~P0~~   | ~~`--version` flag~~       | ~~Trust & debugging~~        | ✅ Done |
| P0       | GitHub Action              | CI adoption                  | Open    |
| P0       | Config file support        | DX for real projects         | Open    |
| P1       | Exclude patterns           | Works on real repos          | Open    |
| P1       | Fix CONTRIBUTING.md        | Contributor trust            | Open    |
| P1       | Pre-commit hook            | Ecosystem discovery          | Open    |
| P2       | Watch mode                 | Development DX               | Open    |
| P2       | Diff/regression mode       | Incremental adoption         | Open    |
| ~~P2~~   | ~~Self-validation in CI~~  | ~~Trust through dogfooding~~ | ✅ Done |
| P3       | Shell completions          | Power-user polish            | Open    |
| P3       | Language discovery command | Discoverability              | Open    |
| P3       | API stability docs         | Library consumer confidence  | Open    |
