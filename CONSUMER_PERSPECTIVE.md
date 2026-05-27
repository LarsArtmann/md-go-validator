# Consumer Perspective — What's Missing

A brutally honest assessment of what a new user would find lacking, confusing, or blocking adoption.

---

## Critical Gaps (Adoption Blockers)

### 1. No `--version` Flag

Every CLI tool needs `--version`. The `.goreleaser.yml` injects `version`, `commit`, and `date` via ldflags, but the binary doesn't expose them. A user installing via `go install` or homebrew has no way to verify which version they're running or whether an upgrade worked.

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

### 9. No Error Codes or Machine-Actionable Output

`ValidationError` has error codes internally (`ErrCodeSyntax`, `ErrCodeNotAvailable`), but these are not exposed in the CLI output (JSON, table, etc.). A CI consumer cannot programmatically distinguish "syntax error" from "validator not available" from the output.

### 10. No Diff / Regression Mode

No `--baseline` or `--compare` flag. When adding this tool to an existing project with many broken code blocks, a consumer has no way to say "I know these 12 are broken, only fail on NEW errors." This makes incremental adoption painful.

---

## Moderate Gaps (Polish & Trust)

### 11. No `--dry-run` Flag

No way to see what *would* be validated without actually running validation. Useful for debugging config, excludes, and language filters.

### 12. No Progress Indicator

When validating a large directory, there is no progress bar, spinner, or file count indicator. The user stares at a blank terminal until results appear. `-v` prints per-block info, but there is no summary progress.

### 13. No Summary Exit Codes

The tool exits with 0 (all valid) or 1 (any error). No distinction between "errors found" vs. "tool crashed" vs. "no files found." Consumers in CI cannot tell if the step failed because of invalid code or because the path was wrong.

### 14. No Self-Validation

The tool's own `README.md`, `EXAMPLES.md`, and `CONTRIBUTING.md` contain code blocks, but there is no evidence the tool runs against its own docs in CI. Dogfooding builds trust.

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

| Priority | Item | Consumer Impact |
|----------|------|-----------------|
| P0 | `--version` flag | Trust & debugging |
| P0 | GitHub Action | CI adoption |
| P0 | Config file support | DX for real projects |
| P1 | Exclude patterns | Works on real repos |
| P1 | Fix CONTRIBUTING.md | Contributor trust |
| P1 | Pre-commit hook | Ecosystem discovery |
| P2 | Watch mode | Development DX |
| P2 | Diff/regression mode | Incremental adoption |
| P2 | Self-validation in CI | Trust through dogfooding |
| P3 | Shell completions | Power-user polish |
| P3 | Language discovery command | Discoverability |
| P3 | API stability docs | Library consumer confidence |
