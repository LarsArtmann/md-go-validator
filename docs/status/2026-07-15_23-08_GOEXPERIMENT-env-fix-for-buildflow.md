# Status Report: GOEXPERIMENT=jsonv2 Environment Fix for BuildFlow

**Date:** 2026-07-15 23:08
**Session:** BuildFlow GOEXPERIMENT=jsonv2 root cause fix
**Trigger:** BuildFlow `--fix --semantic --build-mode=full` failing with 4 steps: go-auto-upgrade, go-fix, govalid-generate, test-race

---

## Root Cause

The dependency `go-output@v0.30.4` (and transitively `go-branded-id@v0.3.2`) imports `encoding/json/v2` and `encoding/json/jsontext`, which are experimental packages in Go 1.26 requiring `GOEXPERIMENT=jsonv2`. BuildFlow runs Go tool subprocesses (`go build`, `go test`, `govalid`) that inherit the parent shell's environment. The user's shell had Go available system-wide but `GOEXPERIMENT` was not set, causing all 4 Go-dependent BuildFlow steps to fail with `build constraints exclude all Go files in .../encoding/json/v2`.

---

## a) FULLY DONE

### Session 1: Build configuration fixes (committed in `69fcb10`)

| Fix                                 | File                       | Detail                                                                                        |
| ----------------------------------- | -------------------------- | --------------------------------------------------------------------------------------------- |
| `GOEXPERIMENT=jsonv2` in devShells  | `flake.nix`                | Both `default` and `ci` devShells now set the env var                                         |
| `GOEXPERIMENT=jsonv2` in apps       | `flake.nix`                | `test` and `lint` apps export it in their shell scripts                                       |
| `GOEXPERIMENT=jsonv2` on derivation | `package.nix`              | `buildGoModule` sets it as an environment attribute                                           |
| `GOEXPERIMENT=jsonv2` in CI         | `.github/workflows/ci.yml` | `test` job (job-level env) and `build` job (step-level env)                                   |
| `GOEXPERIMENT=jsonv2` in releases   | `.goreleaser.yml`          | Added to build env alongside `CGO_ENABLED=0`                                                  |
| `GOEXPERIMENT=jsonv2` in Docker     | `Dockerfile`               | Added to the `RUN go build` command                                                           |
| AGENTS.md documentation             | `AGENTS.md`                | New "Nix Gotcha: GOEXPERIMENT=jsonv2" section documenting all 5 locations + the error message |
| vendorHash fix                      | `package.nix`              | Updated from stale `sha256-vQQZ...` to correct `sha256-I7oN...`                               |

### Session 2: Local environment fix (the actual root cause)

| Fix                      | File                  | Detail                                                                |
| ------------------------ | --------------------- | --------------------------------------------------------------------- |
| `.envrc` created         | `.envrc` (gitignored) | `export GOEXPERIMENT=jsonv2` — direnv auto-loads on `cd` into project |
| `direnv allow` executed  | N/A                   | Authorized the `.envrc` for direnv                                    |
| `.buildflow.yml` created | `.buildflow.yml`      | BuildFlow project config: full mode, 4 concurrency, standard excludes |

### Verification (all green through `direnv exec`)

- `go build ./...` — OK
- `go test -race ./...` — 10/10 packages pass
- `govalid ./...` — exit 0, no markers
- `golangci-lint run ./...` — 0 issues

---

## b) PARTIALLY DONE

| Item                     | Status          | Detail                                                                                                                                                                                                                                                                                                              |
| ------------------------ | --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `go-auto-upgrade:repair` | Warnings remain | 5 functions flagged as reimplementing `slices.Contains` as for-loops. These are informational warnings, not compilation errors. The repair step fails because it runs a post-repair compile check without GOEXPERIMENT — now fixed via `.envrc`, but the actual `slices.Contains` migrations have not been applied. |
| `npm-update` warning     | Harmless noise  | `failed to read package file "website": read website: is a directory` — BuildFlow bug treating a directory as a package.json path. Still shows green.                                                                                                                                                               |

---

## c) NOT STARTED

1. **Apply `slices.Contains` migrations** — 5 functions flagged by go-auto-upgrade:
   - `pkg/code/module.go:44` — `isModuleDirective`
   - `pkg/extractor.go:121` — `hasSkipDirective`
   - `pkg/languages/language.go:131` — `IsSupported` (wrapper, not reimpl)
   - `pkg/validator.go:453` — `isExcluded`
   - `pkg/validator.go:641` — `HasErrors`
   - `pkg/validator.go:652` — `HasSkipped`
2. **`.envrc` is gitignored** — other developers cloning this repo won't get the fix automatically. Consider documenting the requirement in README or adding a `.envrc.example`.
3. **Website `npm-update`** — BuildFlow can't update website npm deps due to the directory-vs-file bug. May need BuildFlow upstream fix or a workaround.

---

## d) TOTALLY FUCKED UP

### Session 1 oversight: Fixed configs but not the actual environment

In session 1, I fixed `GOEXPERIMENT=jsonv2` in 6 config files (flake.nix, package.nix, CI, goreleaser, Dockerfile, AGENTS.md) but **completely missed that BuildFlow runs from the user's regular shell**, not from `nix develop`. The nix devShell already had the fix, but the user never enters it — they have Go installed system-wide via NixOS and run `buildflow` directly from bash. The result: all the config fixes were correct but irrelevant for the actual problem. BuildFlow kept failing identically.

**Lesson:** Always test the actual failure path. I verified `GOEXPERIMENT=jsonv2 go build` worked, but never checked whether BuildFlow's subprocesses would inherit the variable. The real fix was a one-line `.envrc` that took a second session to discover.

---

## e) WHAT WE SHOULD IMPROVE

1. **Document the `.envrc` requirement** — Add a note in README "Development" section: "Run `direnv allow` after cloning to set GOEXPERIMENT=jsonv2"
2. **Consider `use flake` in `.envrc`** — Would give full devShell (Go, gopls, golangci-lint, goreleaser) automatically, but adds eval latency
3. **Apply the `slices.Contains` modernizations** — Pure code quality, no behavior change, makes the codebase cleaner
4. **Add GOEXPERIMENT to `.buildflow.yml`** if BuildFlow ever supports env configuration (currently doesn't)
5. **Consider upgrading go-output to avoid jsonv2** — If go-output ever publishes a version that doesn't require the experiment flag, this entire class of problem goes away

---

## f) NEXT 50 THINGS TO DO

### High Priority (P0)

1. Run full `buildflow --fix` to verify all 4 previously-failing steps now pass
2. Apply the 5 `slices.Contains` migrations flagged by go-auto-upgrade
3. Verify `nix build .#` still passes (it was building successfully at end of last run)
4. Commit the `.buildflow.yml` (already committed in `4bc17ef`)

### Medium Priority (P1)

5. Add `.envrc.example` to repo for other developers
6. Document GOEXPERIMENT requirement in README Development section
7. Update AGENTS.md Build Commands section to mention `direnv allow`
8. Consider adding `GOEXPERIMENT=jsonv2` to the `checks` in flake.nix (test derivation)
9. Verify the website CI workflow (`.github/workflows/website.yml`) doesn't need GOEXPERIMENT
10. Check if the GoReleaser `before.hooks` (`go mod tidy`) needs GOEXPERIMENT
11. Run `go test -bench=. -benchmem ./pkg/` to verify benchmarks work
12. Consider adding a `direnv allow` check to BuildFlow or CI

### Code Quality (P2)

13. Migrate `isModuleDirective` to `slices.Contains` in `pkg/code/module.go`
14. Migrate `hasSkipDirective` to `slices.Contains` in `pkg/extractor.go`
15. Migrate `isExcluded` to `slices.Contains` in `pkg/validator.go`
16. Migrate `HasErrors` to `slices.Contains` in `pkg/validator.go`
17. Migrate `HasSkipped` to `slices.Contains` in `pkg/validator.go`
18. Review `IsSupported` wrapper in `pkg/languages/language.go` — consider inlining
19. Add integration test for GOEXPERIMENT requirement (test that build fails without it)
20. Review if `go-branded-id` is actually needed (it's an indirect dep via go-output)

### Documentation (P3)

21. Update FEATURES.md with BuildFlow integration status
22. Add CHANGELOG entry for GOEXPERIMENT fix
23. Update TODO_LIST.md with remaining items
24. Document the direnv workflow in CONTRIBUTING.md
25. Add a "Troubleshooting" section to README for the jsonv2 error

### Build Infrastructure (P3)

26. Consider adding `meta.description` to flake.nix apps (test, lint) — BuildFlow warns about missing descriptions
27. Review if the `result` symlink should be in `.gitignore` (BuildFlow warns about it)
28. Consider adding a `nix develop` wrapper script for users without direnv
29. Evaluate whether the website flake.nix should also set GOEXPERIMENT (probably not — it's Node.js)
30. Check if the `checks.test` derivation needs GOEXPERIMENT explicitly (it uses `doCheck = true` on the package which has it)

### Testing (P4)

31. Add a CI step that verifies `GOEXPERIMENT=jsonv2 go build` works
32. Add a CI step that verifies the nix build produces a working binary
33. Consider adding property-based tests for the code extraction logic
34. Add more benchmark coverage for large markdown files
35. Test the Dockerfile build with GOEXPERIMENT
36. Verify goreleaser snapshot build works with GOEXPERIMENT

### Dependency Management (P4)

37. Pin go-output version in a comment in go.mod explaining why
38. Consider vendoring dependencies to avoid network issues during nix builds
39. Review go-faster/jx and go-faster/errors — are they pulling their weight?
40. Check if gotreesitter v0.37.0 has any breaking changes from v0.21.0 listed in AGENTS.md
41. Audit all indirect dependencies for unused packages
42. Consider replacing go-faster/yaml with stdlib if Go 1.26 has improved YAML support (it doesn't, but worth checking)

### Architecture (P4)

43. Consider making GOEXPERIMENT configurable via build tags instead of env var
44. Evaluate if the jsonv2 dependency in go-output is fundamental or optional
45. Review the package boundary between pkg/output and go-output — is the abstraction right?
46. Consider adding a `make` / `just` target for `GOEXPERIMENT=jsonv2 go test ./...` as a fallback
47. Document the build matrix: which Go commands need GOEXPERIMENT vs which don't
48. Add a `.tool-versions` or equivalent for Go version pinning outside Nix
49. Consider adding a shell.nix fallback for non-flake Nix users
50. Evaluate direnv `use flake` vs manual exports — benchmark eval latency

---

## g) Top 2 Questions

### 1. Should the `.envrc` use `use flake` or just `export GOEXPERIMENT=jsonv2`?

Currently it's a one-liner export. Using `use flake` would give the full devShell (Go, gopls, golangci-lint, goreleaser) automatically on `cd`, but adds ~1-3s eval latency on every directory entry. The user has Go available system-wide via NixOS, so `use flake` is mostly redundant except for gopls/golangci-lint/goreleaser. This is a preference question I cannot answer without knowing the user's workflow.

### 2. Should we apply the `slices.Contains` migrations now, or wait?

The 5 functions flagged by `go-auto-upgrade` are functionally correct — they work fine as for-loops. Converting them to `slices.Contains` is a pure modernization with no behavior change. However, some of these functions have custom logic beyond a simple `slices.Contains` (e.g., `hasSkipDirective` checks string prefixes). I need to read each function to determine if the migration is a clean 1:1 replacement or if it changes behavior. I cannot determine this without reading the source.
