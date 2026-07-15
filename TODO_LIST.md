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

| Task                                   | Status       | Impact | Effort | Evidence                                                                                                                                                                                                       |
| -------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Decide on postPatch replace directive  | 🔵 `BLOCKED` | High   | -      | `package.nix:26`; overlay confirmed non-functional: go-finding is private (SSH URL, no GOPRIVATE in derivation), so overlay build fails without replace. Decision hinges on whether go-finding is made public. |
| Add drift guard: go.mod vs flake input | 🔴 `TODO`    | High   | 1h     | Currently aligned (v1.2.0 both sides); guard prevents future go-finding skew. Only replace'd flake input.                                                                                                      |
| Publish Homebrew tap                   | 🔴 `TODO`    | High   | 30min  | `.goreleaser.yml:101` has `skip_upload: true`                                                                                                                                                                  |

## Medium Impact

| Task                                            | Status    | Impact | Effort | Evidence                                            |
| ----------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------- |
| Add `--dry-run` flag                            | 🔴 `TODO` | Med    | 1h     | Show what would be validated without running        |
| Add progress indicator for large dirs           | 🔴 `TODO` | Med    | 2h     | No spinner/progress bar; blank terminal during scan |
| Generate shell completions (bash/zsh/fish)      | 🔴 `TODO` | Med    | 1h     | No completion code exists                           |
| Document API stability (stable vs experimental) | 🔴 `TODO` | Med    | 1h     | Library consumers lack guidance; `pkg/` exports     |
| Run `nix flake check --all-systems`             | 🔴 `TODO` | Med    | 10min  | Only x86_64-linux verified so far                   |

## Low Impact

| Task                                   | Status    | Impact | Effort | Evidence                                              |
| -------------------------------------- | --------- | ------ | ------ | ----------------------------------------------------- |
| Verify vendorHash stability hypothesis | 🔴 `TODO` | Low    | 30min  | proxyVendor behavior; see status report 2026-07-07 §e |

---

<!-- Source: Compiled from status report 2026-07-07 Top 25 and CONSUMER_PERSPECTIVE.md
     remaining open items. Each item verified against code before inclusion.
     Completed this session (removed per DONE policy): golangci-lint standalone
     (0 issues), package.nix postPatch comment, finding round-trip integration
     test, flake input skew audit (no skew — only go-finding-src has replace),
     slices.ContainsFunc + errors.AsType modernizations.
     Overlay-without-replace confirmed BLOCKED: go-finding is a private repo. -->
