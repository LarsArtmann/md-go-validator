# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to TODO_LIST.md.

## Recently Shipped

| Item                       | Details                                                                                                  |
| -------------------------- | -------------------------------------------------------------------------------------------------------- |
| Documentation website      | [md-go-validator.lars.software](https://md-go-validator.lars.software) — Astro + Starlight + Tailwind v4 |
| CI/CD auto-deploy          | `.github/workflows/website.yml` — Firebase Hosting on push to master                                     |
| Partial-results-on-error   | Validation results survive file/dir failures (`c01632d`, `d40313d`)                                      |
| SARIF + baseline mode      | GitHub Code Scanning output (`1ea7d41`); regression mode (`9d11fa0`)                                     |

## Themes

### 1. Broader Language Coverage

Expand validation to more languages commonly found in documentation.

Raw ideas:

- Python support
- Java support
- C/C++ support
- SQL syntax validation
- Shell/Bash validation

### 2. Developer Experience

Reduce friction in the validation workflow for documentation authors.

Raw ideas:

- Watch mode (`--watch`) for instant feedback during doc editing
- Fix suggestions: auto-fix common syntax errors in code blocks
- LSP server mode for real-time inline diagnostics in editors
- Web playground for trying validation in the browser

### 3. Distribution and Discovery

Make the tool trivially easy to adopt and discover.

Raw ideas:

- AUR package for Arch Linux
- pnpm wrapper for JS-heavy teams
- Docker image publication to a registry (GHCR) for CI pipelines without Go/Nix

### 4. Architecture Evolution

Structural changes that need a design decision before they become tasks.

Raw ideas:

- Split tree-sitter validators into an opt-in sub-package (`pkg/languages/treesitter`) so Go-only embedders avoid the multi-MB dependency (asked by library consumers; breaks `DefaultRegistry()` API — needs a decision)
- Decouple `pkg/output` (go-output) behind the CLI boundary for a leaner library import graph
- Move implementation packages under `internal/` to shrink the public API surface
- Adopt `go-error-family` for classified errors (rejected/transient/input/internal), or write an ADR and silence the BuildFlow gate
- Cross-platform CI matrix (macOS/Windows) for path/ANSI verification
- HTML report output format

### 5. Testing Depth

Higher-assurance testing beyond table-driven unit tests.

Raw ideas:

- Property-based/fuzz tests for the extractor state machine and elision normalizer
- Snapshot tests for JSON/YAML/SARIF output regression
- BDD (Ginkgo/Gomega) specs for critical user journeys

## Open Questions

Decisions pending; nothing here is actionable until resolved.

1. **`postPatch` replace directive for `go-finding`** — keep (local iteration on a private dep) or drop (v1.2.0 is proxy-published; 3-place bump coupling is a bug magnet)? Blocked on whether `go-finding` goes public. Tracked as BLOCKED in `TODO_LIST.md`.
2. **`go-error-family` adoption** — the dependency appears only transitively today. Adopt for structured error classification, or ADR-and-silence the BuildFlow gate?
3. **Tree-sitter sub-package split** — see Architecture Evolution; `GoOnlyRegistry()` vs explicit `treesitter.RegisterAll()` vs status quo.
4. **Where should `ValidationError` live** — `pkg/languages/` or `pkg/types/`? (Raised 2026-06-26; unresolved.)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Runtime validation of executed code:** This is a syntax validator, not a test runner. Running code blocks is out of scope.
- **Linting/style enforcement:** Syntax validity is the contract. Opinionated style rules belong in linters (golangci-lint, eslint, etc.).
- **Markdown linting:** We validate embedded code blocks, not markdown structure. Tools like markdownlint own that space.
