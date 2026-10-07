# Status Report: GitHub Triage — Issue #8, Dependabot PRs #4/#7, CI + Website Repair

**Date:** 2026-10-07 07:19 CEST
**Scope:** This session only — GitHub issue/PR triage + all breakage discovered and fixed along the way.
**Session commits (master):**

| Commit    | Subject                                                                       |
| --------- | ----------------------------------------------------------------------------- |
| `99cea9a` | Fix flake build: pin Go 1.27 toolchain and public go-finding input (Fixes #8) |
| `199647d` | chore(deps): bump svgo (#7) — squash-merge of PR #7                           |
| `43acbdf` | Pin website TypeScript to 6 so astro check works again                        |
| `020cb24` | Patch 19 vulnerable transitive deps via pnpm workspace overrides              |
| `3f913bf` | docs: record pnpm 11 overrides and astro check TS 6 gotchas                   |

**End state:** 0 open issues, 0 open PRs. Master CI ✅ (Test/Lint/Build/Nix flake check), Website workflow ✅ (astro check + build + deploy), `pnpm audit` 0 vulnerabilities, `go test ./...` ✅, `golangci-lint` 0 issues, `nix build` / `nix run -- --help` / `nix flake check` ✅ locally.

---

## a) FULLY DONE

1. **Issue #8 fixed and closed** (auto-closed by `Fixes #8` in `99cea9a`, verification comment posted):
   - `package.nix`: builds with `go_1_27` via `(buildGoModule.override { go = go_1_27; })`; vendorHash recomputed (`buildflow -s nix-hash-fix --fix`, verified 5/5 targets, 0 findings)
   - `flake.nix`: `go_1_27` in devShell.default, devShell.ci, `apps.test`, `apps.lint`
   - `flake.lock` re-locked to the new input URL
   - All three issue-verification criteria pass locally: `nix build .#default`, `nix run .#default -- --help`, `nix flake check` (incl. `checks.test` with `doCheck = true` running go test in-sandbox on 1.27.1)
2. **CI root cause fixed** (same as #8): `.github/workflows/ci.yml` `setup-go` 1.26 → 1.27 in test/lint/build. Master CI went from 3-failing-jobs to fully green.
3. **go-finding flake input repaired**: `git+ssh://...v1.10.0` (unfetchable on CI, and a version split-brain vs go.mod) → public `github:LarsArtmann/go-finding?ref=v1.14.0` (matches go.mod; repo verified public, tag exists).
4. **PR #7 merged** (svgo 4.0.2 → 4.1.0, security hardening for `removeScripts`, GHSA-4vpr-x523-8j87 / GHSA-w27v-7q3p-w38r). Its rebased lockfile also silently resolved the website lockfile↔manifest drift.
5. **PR #4 closed** as superseded (astro target 7.2.8 behind master's ^7.3.5; sharp 0.35.4 already present; svgo landed via #7) with explanation comment.
6. **golangci-lint brought to 0 issues** (AGENTS.md quality bar): 5 stale `//nolint:exhaustruct` → `exhaustruct_v5`; intentional-default format switch annotated with `//nolint:exhaustive` + rationale; stale `run.go: 1.26.7` pin removed from `.golangci.yml` (auto-detect from go.mod prevents recurrence).
7. **Website deploy pipeline un-broken**: TS 7.0.2 (unsupported by `astro check`) → `typescript: ^6` (6.0.3); astro check 0 errors; build 16 pages; deploy green for the first time since 2026-10-01.
8. **19 vulnerabilities cleared (12 high)** via `overrides` in `website/pnpm-workspace.yaml` (devalue, fast-uri, js-yaml, http-cache-semantics, postcss-selector-parser, sharp, smol-toml, source-map-js). `pnpm audit`: "No known vulnerabilities found". Verified astro check + build still pass.
9. **AGENTS.md refreshed**: dep versions corrected (Go 1.27+, gotreesitter v0.55.1, go-output v0.38.3, go-finding v1.14.0 — were badly stale); new gotchas recorded (Go toolchain pin, exhaustruct_v5 rename, inert legacyerrors nolints, removed run.go pin, go-finding github URL, pnpm 11 overrides home, astro check TS 6 requirement).
10. **Full local verification pass**: `go test ./...` on go 1.27.0, lint 0 issues, nix build/run/check, `pnpm install --frozen-lockfile` (CI parity), `pnpm run build`, `pnpm exec astro check`.

## b) PARTIALLY DONE

1. **Release for issue #8 consumers** — fix is on `master` only. Issue #8's whole premise is tag-consuming flakes (`github:...#v1.3.0` is "unbuilt by construction"); those consumers are still broken until a tag is cut. Not a release session, but this is the gap between "issue closed" and "problem gone for consumers". → see f) #1.
2. **Website end-state** — workflow green, but I never verified the deployed site in a browser / via fetch (live URL is `md-go-validator.web.app`; the lars.software domain is a pre-existing DNS blocker). Deploy success ≠ correct content.
3. **Lint hygiene** — 0 issues, but knowingly left behind (documented in AGENTS.md, not fixed): inert `//nolint:legacyerrors` directives (golangci-lint warns "unknown linters"), dead `exhaustruct` entry in `.golangci.yml` test exclusions, CI golangci-lint still floats on `version: latest` (the very mechanism that caused the exhaustruct_v5 drift).
4. **Docs** — AGENTS.md updated, but TODO_LIST.md / FEATURES.md / CHANGELOG.md / ROADMAP.md were not touched; the section-(f) items below are not harvested into TODO_LIST.md yet (risk of entombing them in this timestamped file). Coverage table in AGENTS.md not re-measured.
5. **Nix verification breadth** — `nix flake check` ran on x86_64-linux only; aarch64-darwin / aarch64-linux (`--all-systems`) unverified. `nix run .#test` (-race app) not run locally.
6. **Dependabot security job** — its last run (at `43acbdf`) still shows failure; with `pnpm audit` now at zero it _should_ pass on the next scheduled run, but that is a prediction, not a verification.

## c) NOT STARTED

1. v1.3.1 (or v1.4.0) release so flake consumers at tags get the go-1.27 fix.
2. TODO_LIST.md / ROADMAP.md harvest of this report's section (f).
3. CHANGELOG.md entries for today's fix batch.
4. README check for stale "Go 1.26" / install instructions (never opened it this session).
5. `.goreleaser.yml` review against go 1.27 (goreleaser not run, not even `--dry-run`).
6. `nix flake check --all-systems` (darwin/arm64).
7. Pinning golangci-lint version in CI.
8. `.github/dependabot.yml` review (never read it).
9. Live-site verification after deploy.
10. Stray-tags cleanup (v1.0.0/v1.1.0/v1.2.0 — pre-existing, already tracked in TODO_LIST.md).

## d) TOTALLY FUCKED UP

**Nothing destructive or irreversible.** Three blemishes, all recovered or cosmetic:

1. **PR #4 close-comment shell quoting bug**: I wrapped the comment in double quotes with literal backticks → local command substitution executed `` `master` `` (error: "master: executable file not found in $PATH"). The comment WAS posted, but "already on `master`" almost certainly rendered with an empty inline-code span. Cosmetic, fixable via comment edit.
2. **Auto-commit daemon races** (3×): the daemon committed mid-work (`0ae6cf3`, `17595b5`, `d44d372`, `f1af077`) with heuristic messages; I briefly misread one as lost work, then soft-reset + squashed each into proper commits. History ended clean — but I should have checked `git log` before my first commit attempt instead of being surprised three times.
3. **First `package.nix` edit was syntactically wrong** (added both `(` and a trailing `)`); caught immediately by `nix-instantiate --parse`, fixed in seconds. Zero impact.

## e) WHAT WE SHOULD IMPROVE

1. **Follow my own project-discovery checklist**: I never opened TODO_LIST.md / FEATURES.md / ROADMAP.md / README this session (AGENTS.md Tier 2 requires it). Any of them could have flagged the release expectation for #8 earlier.
2. **Think "who consumes this fix" at close time**: closing #8 with only a master-branch fix left tag consumers broken. Consumer-facing fixes should trigger the "does this need a release?" question explicitly, every time.
3. **Commit before the daemon does**: when the daemon is active, stage + commit intentionally at each milestone (or at least `git log` first). Heuristic squash-dances cost time and attention.
4. **Drift-prone pins need either automation or removal**: this session removed one (`run.go`) and added one (`typescript: ^6` — unavoidable until Astro supports TS 7). Audit for more: `version: latest` in golangci-lint-action is the next one.
5. **Verify externally-visible claims**: "Website: success" was accepted without fetching the live URL; "Dependabot will pass next run" is a prediction. Green workflow ≠ verified outcome.
6. **Shell-quote discipline for `gh` bodies**: heredocs (`gh pr close 4 --comment "$(cat <<'EOF' ...")` would have prevented the backtick bug.

## f) Up to 50 things we should get done next

_Brainstorm sorted by impact — NOT a commitment list; harvest-worthy items belong in TODO_LIST.md, the rest in ROADMAP.md._

**Release & consumers (highest impact)**

1. Cut a release (v1.3.1/v1.4.0) so tag-consuming flakes get the go-1.27 fix.
2. `goreleaser release --dry-run` (or local `--snapshot`) to verify release tooling against go 1.27.
3. Check README + website docs for stale Go version / install-from-flake instructions; point them at the new tag.
4. Update CHANGELOG.md with today's batch.
5. Verify pkg.go.dev / module proxy pick up the new tag cleanly.

**Correctness & verification debt**
6. `nix flake check --all-systems` (aarch64-darwin, aarch64-linux).
7. Run `nix run .#test` (the -race app) and `go test -race ./...` locally post-changes.
8. Fetch/inspect the live site (md-go-validator.web.app) to confirm the deploy content, not just the green job.
9. Confirm the Dependabot Updates security job goes green on its next scheduled run.
10. Root-cause _why_ the Dependabot security job failed to resolve (the overrides treat the symptom; is there a resolver bug or a repo config factor?).
11. Review `.github/dependabot.yml` (never read this session).
12. Verify the `go-finding-src` replace-directive machinery is still needed now that input == go.mod version (v1.14.0 == v1.14.0) — candidate for simplification in `package.nix`.
13. Add a guard so go.mod's `go` line can never exceed the nix-pinned toolchain again (CI eval check or flake assertion) — prevents #8 recurrence.
14. CI's `nix flake check --no-build` only _evaluates_ — nothing in CI builds the nix package; consider a cached full `nix build` job (cachix/magic-nix-cache).
15. Fix the PR #4 close-comment rendering (empty inline code from the backtick bug) via comment edit.
16. Sweep docs for now-dead linter names (exhaustruct, legacyerrors) in EXAMPLES.md / CONTRIBUTING.md / website content.

**Lint & toolchain hygiene**
17. Pin golangci-lint version in ci.yml (reproducibility; kills the exhaustruct_v5-class drift).
18. Remove inert `//nolint:legacyerrors` directives (multiple files) — or actually run the errors.AsType modernization they were waiting for.
19. Evaluate `errors.As` → `errors.AsType[E]` migration (go-error-modernization pass; the legacyerrors nolints hint this was deferred).
20. Drop the dead `exhaustruct` entry from `.golangci.yml` test exclusions.
21. Verify AGENTS.md's `noinlineerr` gotcha — that linter is not in the enable list (doc drift either way).
22. Confirm the BuildFlow pre-commit hook still passes with the changed `.golangci.yml`.
23. Address `buildflow doctor` warnings seen this session: stale buildflow binary (built at 202b114, HEAD is 185bf82), legacy DB migration, go.mod go-line flipflop.
24. Pin `gopls` in the devshell to a build that matches go 1.27 (currently inherits nixpkgs default toolchain).
25. Decide whether go.mod should carry an explicit `toolchain` directive.

**Docs & process**
26. HARVEST this report's section (f) into TODO_LIST.md / ROADMAP.md (docs-health mode).
27. Slim AGENTS.md (buildflow doctor flagged 272 lines vs 220 budget).
28. Refresh the AGENTS.md coverage table (`go test -cover`).
29. Re-run a docs-health VERIFY pass: today's edits touched many doc claims (GOEXPERIMENT locations list, Nix section, Website section).
30. Record the session's "consumer-fix needs release" lesson in AGENTS.md or lessons.md (cross-project candidate).

**Dependency & supply-chain**
31. Consider Dependabot automerge for the npm_and_yarn group (with green-CI gates) so PRs like #4/#7 don't rot for weeks.
32. Decide whether the workspace `overrides` should be loosened when parent packages catch up (add a review date/comment).
33. Check whether Dependabot understands pnpm-workspace.yaml overrides for future update PRs (else pin-vs-update conflicts will recur).
34. The git "Stray Tags" cleanup (v1.0.0/v1.1.0/v1.2.0) — pre-existing TODO_LIST item, still open.
35. Audit other fleet repos for the same `git+ssh` flake-input pattern that broke CI here.
36. Audit other fleet repos for go.mod 1.27 + nixpkgs-default-go drift (issue #8's root cause is probably not unique to this repo).

**Nix quality (nix-review backlog)**
37. Consider `mkShellNoCC` for devShell.default (no C compiler needed for a pure-Go build).
38. Deduplicate the `go_1_27` pin: currently repeated in package.nix + 4 flake spots; consider a single `let goPkg = ...` binding or flake-level arg.
39. website/flake.nix got a whitespace-only touch from `nix fmt` this session — review that file separately someday (it was never actually read).
40. Consider a flake `checks` entry that fails evaluation if `pkgs.go.version` satisfies go.mod but the pin drifted.

**Website**
41. DNS: md-go-validator.lars.software still blocked on the Namecheap API key placeholder (pre-existing) — unblock or de-scope.
42. Consider adding `pnpm audit --prod` as a website CI gate (the supply-chain lockfile check exists; audit is stronger).
43. Decide TS 7 upgrade plan (wait for Astro stable `@astrojs/ts-content-mapper`, then drop the ^6 pin).
44. Sitemap/social-preview sanity check after today's dep bumps (pagefind index rebuilt fine locally; live behavior unverified).

**Repo meta**
45. Decide the merge policy for Dependabot PRs: squash (used for #7) vs merge-commit — document it so future handling is consistent.
46. Consider labeling/convention for superseded-PR closures (what was done manually for #4).
47. Investigate the `repo-private: true` flag in the Dependabot job definition (this repo is public — likely harmless Dependabot quirk, worth confirming).
48. Follow up on AGENTS.md Website section claim "URL: md-go-validator.lars.software (pending DNS propagation)" — stale since it's blocked, not propagating.
49. Consider adding a top-level flake `apps.mdv`-style alias or keep `apps.default` (cosmetic; only if consumers ask).
50. Post-release: re-run `nix build github:LarsArtmann/md-go-validator/<new-tag>` from a clean store to prove the issue-#8 consumer path end-to-end.

## g) Questions I can NOT figure out myself

1. **Release authority:** Should I cut v1.3.1 (or v1.4.0) _now_ so flake consumers at tags actually receive the issue-#8 fix — and if yes, version number preference? (Today's fix is only on master; the issue's consumers consume by tag.)
2. **Dependency-update policy:** Do you want Dependabot automerge enabled for the npm_and_yarn group (merge when CI is green), so group PRs stop rotting for weeks like #4 did — or do you prefer keeping the manual merge gate and just rely on the new workspace overrides?
3. **legacyerrors / errors.AsType:** The inert `//nolint:legacyerrors` directives suggest a deferred `errors.As` → `errors.AsType[E]` modernization. Should I run that migration in this repo now (removing the dead nolints properly), or do you want to keep them as-is for a coordinated fleet-wide pass?

---

_Report generated per session scope — no external research beyond this session's observations. Section (f) is brainstorm-grade; route through docs-health HARVEST before treating any item as committed work._
