# Consumer Perspective — What's Missing

A brutally honest assessment of what a new user would find lacking, confusing, or blocking adoption.

> **Note:** Feature status now lives in [FEATURES.md](FEATURES.md). Actionable
> improvements live in [TODO_LIST.md](TODO_LIST.md). This file tracks the
> consumer-perspective lens on remaining gaps.

---

## Resolved

| Item                          | Resolution                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------------- |
| ✅ `--version` flag           | Added `--version` / `-V`                                                              |
| ✅ Error codes in output      | `errorCode` field in JSON/YAML output (`syntax`, `not_available`, `not_registered`)   |
| ✅ Summary exit codes         | 0=success, 1=validation errors, 2=tool/usage errors                                   |
| ✅ Self-validation in CI      | CI dogfoods — runs validator on its own docs                                          |
| ✅ STDIN support              | `cat file.md \| md-go-validator -`                                                    |
| ✅ JSON schema                | Output contract documented at `docs/json-schema.json`                                 |
| ✅ Config file support (#2)   | `.md-go-validator.yaml` with `--config`; `pkg/config/config.go`                       |
| ✅ `--init` command (#3)      | `md-go-validator --init` scaffolds default config                                     |
| ✅ Exclude patterns (#4)      | `--exclude` flag with glob `**` support; `pkg/types/identifiers.go`                   |
| ✅ CONTRIBUTING.md fixed (#5) | Current `CONTRIBUTING.md` references no dead files                                    |
| ✅ GitHub Action (#6)         | `action.yml` published; `uses: LarsArtmann/md-go-validator@v1`                        |
| ✅ Pre-commit hook (#7)       | `.pre-commit-hooks.yaml` published                                                    |
| ✅ Baseline/regression (#10)  | `--baseline` / `--save-baseline`; `pkg/baseline/`                                     |
| ✅ `--fail-on-skipped` (#15)  | Exit 1 if any blocks are skipped                                                      |
| ✅ `--list-languages` (#17)   | `md-go-validator --list-languages`                                                    |
| ✅ DESCRIPTION mismatch (#20) | `.goreleaser.yml` now says "Multi-language code block validator for Markdown and MDX" |

---

## Still Open

### High Priority

| Item               | Consumer Impact   | Tracking                 |
| ------------------ | ----------------- | ------------------------ |
| Homebrew tap       | macOS adoption    | `TODO_LIST.md`           |
| postPatch decision | Build reliability | `TODO_LIST.md` (blocked) |

### Medium Priority

| Item               | Consumer Impact             | Tracking       |
| ------------------ | --------------------------- | -------------- |
| `--dry-run` flag   | Config debugging DX         | `TODO_LIST.md` |
| Progress indicator | Large-repo UX               | `TODO_LIST.md` |
| Shell completions  | Power-user polish           | `TODO_LIST.md` |
| API stability docs | Library consumer confidence | `TODO_LIST.md` |

### Long-term (Roadmap)

| Item            | Consumer Impact   | Tracking     |
| --------------- | ----------------- | ------------ |
| Watch mode      | Development DX    | `ROADMAP.md` |
| More languages  | Broader adoption  | `ROADMAP.md` |
| Fix suggestions | Error recovery DX | `ROADMAP.md` |
