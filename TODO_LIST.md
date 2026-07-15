# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
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

| Task                                   | Status       | Impact | Effort | Evidence                                                            |
| -------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------- |
| Decide on postPatch replace directive  | 🔵 `BLOCKED` | High   | -      | `package.nix:25`; user design decision: keep for local iter or drop |
| Add drift guard: go.mod vs flake input | 🔴 `TODO`    | High   | 1h     | Prevents go-finding version skew; see status report 2026-07-07 §e   |
| Publish Homebrew tap                   | 🔴 `TODO`    | High   | 30min  | `.goreleaser.yml:101` has `skip_upload: true`                       |

## Medium Impact

| Task                                            | Status    | Impact | Effort | Evidence                                                 |
| ----------------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------- |
| Add `--dry-run` flag                            | 🔴 `TODO` | Med    | 1h     | Show what would be validated without running             |
| Add progress indicator for large dirs           | 🔴 `TODO` | Med    | 2h     | No spinner/progress bar; blank terminal during scan      |
| Generate shell completions (bash/zsh/fish)      | 🔴 `TODO` | Med    | 1h     | No completion code exists                                |
| Document API stability (stable vs experimental) | 🔴 `TODO` | Med    | 1h     | Library consumers lack guidance; `pkg/` exports          |
| Add comment in package.nix explaining postPatch | 🔴 `TODO` | Med    | 5min   | `package.nix:25`; future readers may unknowingly delete  |
| Add finding round-trip integration test         | 🔴 `TODO` | Med    | 30min  | Exercise `finding.FromResult` against branded `FilePath` |
| Run `nix flake check --all-systems`             | 🔴 `TODO` | Med    | 10min  | Only x86_64-linux verified so far                        |

## Low Impact

| Task                                   | Status    | Impact | Effort | Evidence                                                 |
| -------------------------------------- | --------- | ------ | ------ | -------------------------------------------------------- |
| Verify vendorHash stability hypothesis | 🔴 `TODO` | Low    | 30min  | proxyVendor behavior; see status report 2026-07-07 §e    |
| Run golangci-lint standalone           | 🔴 `TODO` | Low    | 5min   | Confirm 0 issues independent of flake check              |
| Audit other flake inputs for skew      | 🔴 `TODO` | Low    | 15min  | Only go-finding-src has replace today                    |
| Confirm overlay builds without replace | 🔴 `TODO` | Low    | 10min  | `flake.nix:118` calls package.nix without go-finding-src |

---

<!-- Source: Compiled from status report 2026-07-07 Top 25 and CONSUMER_PERSPECTIVE.md
     remaining open items. Each item verified against code before inclusion. -->
