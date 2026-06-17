# Full Code Review — 2026-06-17

> Comprehensive Senior-Architect review of `md-go-validator`.
> Baseline at review start: `go test ./...`, `golangci-lint run ./...`, `go vet` all green.
> **Critical regression found & fixed: `nix build .#` was broken.**

## TL;DR

The codebase is well-structured: strong branded types, functional options, clear package
boundaries, excellent linter coverage, ~85%+ test coverage. A handful of real defects were
fixed during this review; the remaining items are quality/polish improvements prioritized below.

---

## Fixed during this review (quick wins)

| #   | Severity     | File                       | Problem → Fix                                                                                                                                                                                                                                                                                       |
| --- | ------------ | -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Critical** | `flake.nix`, `package.nix` | `nix build .#` **broken**: go-output was upgraded to v0.11.0 in `go.mod`/`go.sum` but the pinned `vendorHash` was stale. Updated both copies to the authoritative hash `sha256-cq1…Scu0=`. Build restored.                                                                                          |
| 2   | **High**     | `flake.nix:134`            | Exported `overlays.default` was **completely broken** — `inherit (final) self;` threw `attribute 'self' missing` (nixpkgs has no `self`). The overlay never evaluated. Removed the broken binding; overlay now yields version `"dev"` (inherent overlay limitation, documented).                    |
| 3   | **High**     | `cmd/.../main.go`          | CLI `--help` **lied to users**: claimed tree-sitter languages require external CLIs (`tsc`, `rustc`, `nix-instantiate`, `terraform`, `hclfmt`). They use **embedded pure-Go grammars**. Help text corrected.                                                                                        |
| 4   | **High**     | `pkg/extractor.go`         | Extractor called `block.MarkValid()` on every non-skipped block **before validation ran** — a status lie. Blocks now stay `StatusUnknown` until the validator decides. (`MarkValid` retained as public API; only `IsSkipped()` is read downstream.)                                                 |
| 5   | **High**     | `pkg/types/report.go`      | `BuildReportData` would **panic** (`nil.Error()`) on the representable-but-invalid state `StatusError` + nil `error`. Added a nil-guard + regression test.                                                                                                                                          |
| 6   | **Med**      | `pkg/validator.go`         | **Split brain**: `supportedExtensions` global map vs `types.AllFileTypes()` were two sources of truth for supported file types (adding a type to one but not the other silently breaks `IsSupportedFile`). Removed the global; `types` is now the single source of truth via `types.ParseFileType`. |
| 7   | **Med**      | `pkg/validator.go`         | Modernized `filepath.Walk` → `filepath.WalkDir` (avoids a per-file `os.Lstat`, uses `fs.DirEntry`).                                                                                                                                                                                                 |
| 8   | **Low**      | `pkg/output/output.go`     | `--quiet` with no errors printed `"All N code blocks valid"` where N excluded skipped blocks — misleading. Now reports skipped count when > 0.                                                                                                                                                      |
| 9   | **Low**      | `AGENTS.md`                | Stale versions: Go `1.26.2` → `1.26.3`, go-output `v0.10.0` → `v0.11.0`.                                                                                                                                                                                                                            |

All fixes verified: `go test ./...` ✓, `go test -race ./...` ✓, `golangci-lint run ./...` (0 issues) ✓,
`nix build .#` ✓, `nix flake check` ✓.

### Note on `go.mod` / `go.sum`

A dependency upgrade (go-output v0.10.1→v0.11.0, go-branded-id v0.3.0→v0.3.1, go-toml v2.3.1→v2.4.0)
appeared in the working tree during the session but was **not authored by this review** (no `go get`
was run). It was treated as pre-existing/intentional WIP. Fix #1 simply makes the Nix build
consistent with it. Verify this upgrade is intended before committing.

---

## Remaining improvements (prioritized)

> **UPDATE: All items [A]–[K] below have been implemented and verified.
> `go test -race`, `golangci-lint` (0 issues), `nix build .#`, and `nix flake check` all pass.**
>
> Notable decisions during implementation:
>
> - **[G]** turned out to be **dead code**, not just a type mismatch:
>   `ContextConfig.MaxFiles`/`MaxBlocksPerFile` were never read by `Build()`/`Branch*`.
>   The real limits live on `FileValidator`. Removed the dead fields + their `With*` methods
>   (and their round-trip-only tests), sharpening `ContextConfig` to context lifecycle only.
> - **[A]** also uncovered a **dead `proxyVendor = true` binding** in the old `flake.nix`
>   (declared in `let`, never passed to `buildGoModule`). The known `vendorHash` was computed
>   for the default (vendor) layout, so `package.nix`'s actual `proxyVendor = true` broke the build.
>   Removed `proxyVendor` from `package.nix` to match — vendor mode is the nixpkgs default anyway.

### Pareto tiers

**1% → 51% (do first):**

- **[A] DRY the Nix build.** `flake.nix` fully duplicates `package.nix` (vendorHash, src, ldflags,
  meta). This duplication _caused_ build break #1 (hash updated in one place, missed the pattern).
  Have `flake.nix` do `pkgs.callPackage ./package.nix { inherit self; }` and delete the inline
  `buildGoModule`. Single source of truth for the build. ~15 min.

**4% → 64%:**

- **[B] Make `StatusError` + nil error unrepresentable.** Fix #5 added a defensive guard, but the
  type-level fix is better: `NewResultWithStatus` should reject `StatusError` (callers must use
  `NewErrorResult`), or split the constructors. Makes an impossible state a compile-time error. ~30 min.
- **[C] Add `pkg/extractor_test.go`.** The extractor — the most logic-heavy parser in the project
  (state machine, skip directives, multi-lang) — has **no dedicated unit tests**; it is only
  exercised indirectly. Direct table-driven tests would lock in behavior and raise `pkg` coverage. ~45 min.

**20% → 80%:**

- **[D] Simplify result-collection concurrency.** `processFilesParallel` → `collectResultsLoop` →
  `collectFromChan` + 2 goroutines + mutex is over-engineered for draining two channels closed
  back-to-back. A single goroutine ranging `results` then `errors` (channels are closed sequentially)
  removes ~40 lines and the `goroutineCount` constant. ~30 min.
- **[E] Fix off-by-one inconsistency in error messages.** `validator.go:199,253` compute
  `blockIndex.Int()-1` (1-based → 0-based) inside messages that elsewhere treat `blockIndex` as
  1-based. Pick one convention; prefer 1-based throughout. ~10 min.
- **[F] Consolidate ANSI color handling.** `printTableHeaderTo` uses raw `"\033[1;36m"` literals
  while `printErrorEntry` uses named constants (`ansiBold`, …). Unify on the constants (add
  cyan/green). ~20 min.

**Longer tail (optional):**

- **[G]** `ContextConfig.MaxFiles`/`MaxBlocksPerFile` use `int`; the rest of the domain uses branded
  `uint` types. Align for consistency.
- **[H]** `processJob` creates a per-file `context.WithCancel` it cancels immediately with no
  independent trigger — effectively the parent ctx. Simplify or document intent.
- **[I]** `ValidationError` uses pointer receivers only, so a value doesn't satisfy `error`. Make
  receivers consistent or document.
- **[J]** CI dogfood step (`ci.yml:59`) validates only 4 files; consider including `docs/` and
  `AGENTS.md`. Also consider running `nix flake check` in CI for parity with the documented workflow.
- **[K]** `TreeSitterValidator` stores both `language Language` and `langName string` (two sources
  of truth for "what grammar"). Derive one from the other where possible.

### Execution graph (D2)

```d2
direction: tb

fixed: "Fixed this review (9 items)" {
  shape: note
  style.fill: "#e6f4ea"
}

A: "A. DRY Nix build\nflake.nix -> package.nix"
B: "B. StatusError+nil\nunrepresentable (types)"
C: "C. pkg/extractor_test.go"
D: "D. Simplify result\ncollection concurrency"
E: "E. blockIndex off-by-one\nin error messages"
F: "F. Consolidate ANSI\nconstants (output)"
G: G
H_K: "G-K. Long tail\n(config consistency, CI, nits)"

A -> B -> C -> D -> E -> F -> G -> H_K

tier_1: "1% / 51%" {style.fill: "#fce8e6"}
tier_2: "4% / 64%" {style.fill: "#fef7e0"}
tier_3: "20% / 80%" {style.fill: "#e8f0fe"}

A -> tier_1: "do first" {style.stroke-dash: 4}
B -> tier_2
C -> tier_2
D -> tier_3
E -> tier_3
F -> tier_3
```

---

## What's genuinely good

- Branded types (`FileID`, `LineNumber`, `BlockIndex`, `Language`, `FileType`) with `Validate()` /
  `String()` — strong domain boundaries, hard to mix up.
- `ValidationStatus` enum over booleans — explicit, terminal-state aware.
- Functional-options builder on `FileValidator`; immutable `ContextConfig` with `With*` chaining.
- Single canonical 5-strategy Go validator; clean `Registry` abstraction for pluggable languages.
- Exceptionally thorough `golangci.yml` (60+ linters incl. `exhaustruct`, `wrapcheck`, `wsl_v5`).
- Pure-Go tree-sitter grammars → zero external runtime deps, trivial CI.
- Path-traversal/null-byte guards on all file inputs.

## Verdict

Solid, honest codebase. The biggest real bug was the broken Nix build (now fixed). After items A–C
the architecture would be notably tighter; D–F are polish. No split-brains remain in domain types;
the remaining "split brains" are config duplication (Nix) which item A eliminates.
