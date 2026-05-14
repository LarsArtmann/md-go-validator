# Modularization Proposal — md-go-validator

**Date:** 2026-05-13 | **Status:** Draft

---

## 1. Executive Summary

md-go-validator is a single-module Go project (`github.com/larsartmann/md-go-validator`) with 7 packages. The project is small enough (~2.5k LOC) that a full multi-module split would be over-engineering. However, there is a critical architectural issue — a **bidirectional dependency** between `pkg/types` and `pkg/languages` — that should be resolved. The right move is a **targeted modularization** focused on:

1. Breaking the `pkg/types` ↔ `pkg/languages` cycle
2. Extracting `pkg/types` into a standalone, zero-dependency types module
3. Keeping the project as a **single go.mod** with clean internal boundaries

**Why NOT a multi-module split?**

- The project is a single CLI tool, not a library ecosystem
- All packages are tightly coupled to the same domain (markdown code validation)
- No external consumers would benefit from independent versioning
- The `goreleaser` build targets a single binary
- CI/CD runs as a single job — no parallelization benefit
- Adding go.work + multiple go.mods would add complexity without value

**What we DO instead:** Resolve the structural coupling, establish clean DAG within the single module, and define clear package boundaries that COULD be split later if the project grows.

---

## 2. Current State Analysis

### 2.1 Module Structure

Single `go.mod` — monolith with no `go.work`:

```
github.com/larsartmann/md-go-validator
├── cmd/md-go-validator/    (main — CLI entry point)
├── pkg/                    (core extraction & validation orchestration)
├── pkg/types/              (domain types: branded IDs, CodeBlock, Result, ReportData)
├── pkg/languages/          (language validators: Go, TreeSitter, Registry)
├── pkg/code/               (Go stdlib parser utilities)
├── pkg/output/             (multi-format report output)
├── pkg/testutil/           (test helpers)
└── pkg/testdata/           (test fixtures)
```

### 2.2 Dependency Graph (Current)

```
cmd/md-go-validator
├── pkg
│   ├── pkg/languages
│   │   └── pkg/code
│   └── pkg/types ──→ pkg/languages  ← CYCLE
├── pkg/output
│   └── pkg/types
└── pkg/types ──→ pkg/languages

pkg/testutil
└── pkg/types
```

### 2.3 The Cycle: `pkg/types` → `pkg/languages`

The critical problem: `pkg/types/code_block.go` imports `pkg/languages` because `CodeBlock` has a field of type `languages.Language`:

```go
type CodeBlock struct {
    LineNumber LineNumber
    Language   languages.Language  // ← this creates the upward dependency
    Code       string
    Status     ValidationStatus
}
```

This means:

- `pkg/languages` cannot import `pkg/types` (would create a compile-time cycle)
- `pkg/types` depends on the full `pkg/languages` package just for one type
- The dependency graph is NOT a DAG — it has a tangle at types ↔ languages

### 2.4 External Dependencies

| Package                      | External Dep                                   | Type       |
| ---------------------------- | ---------------------------------------------- | ---------- |
| `pkg/languages` (treesitter) | `github.com/odvcencio/gotreesitter` + grammars | Production |
| `pkg/output`                 | `github.com/larsartmann/go-output`             | Production |
| `pkg/code`                   | stdlib only (`go/parser`, `go/token`)          | Production |

### 2.5 God-Package Analysis

No god-packages detected. All packages are focused:

- `pkg/types` — 7 files, domain types (appropriate)
- `pkg/languages` — 6 files + 1 test helper, language validation (appropriate)
- `pkg/` — 4 files, orchestration (appropriate)

### 2.6 Test Dependency Analysis

| Test Package         | Internal Imports                                                  |
| -------------------- | ----------------------------------------------------------------- |
| `pkg/` tests         | `pkg/code`, `pkg/languages`, `pkg/testutil`, `pkg/types`          |
| `pkg/types` tests    | `pkg/languages`                                                   |
| `pkg/output` tests   | `pkg/types`                                                       |
| `pkg/testutil` tests | `pkg/languages`, `pkg/types`                                      |
| `cmd/` tests         | `pkg`, `pkg/languages`, `pkg/output`, `pkg/testutil`, `pkg/types` |

Key observation: `pkg/types/types_test.go` imports `pkg/languages` — the cycle infects the test layer too.

---

## 3. Proposed Structure

### 3.1 Decision: Single Module, Clean Boundaries

Keep one `go.mod` but fix the cycle by extracting `Language` type to a neutral location.

### 3.2 The Fix: Extract Language Identifier to `pkg/types`

The `languages.Language` type is a **branded string** — it has no dependencies on `pkg/languages` infrastructure (no registry, no validators, no tree-sitter). Only the type definition and its methods live in `pkg/languages/language.go`.

**Move `Language` type to `pkg/types/language.go`:**

```
pkg/types/language.go     ← NEW (Language branded type + methods)
pkg/types/code_block.go   ← UPDATED (uses types.Language instead of languages.Language)
pkg/languages/language.go ← UPDATED (type alias or removed, constructors remain)
```

### 3.3 Resulting Dependency Graph (Clean DAG)

```
cmd/md-go-validator
├── pkg
│   ├── pkg/languages
│   │   ├── pkg/code
│   │   └── pkg/types         ← now points DOWN
│   └── pkg/types             ← leaf node (zero internal deps)
├── pkg/output
│   └── pkg/types
└── pkg/types                 ← leaf node (zero internal deps)

pkg/testutil
└── pkg/types
```

All arrows point downward. `pkg/types` is a true leaf with zero internal dependencies.

### 3.4 Module Boundary Definition

| Package         | Purpose                                                                              | Depends On (prod)                                 | Depends On (test)                                                 | Public API                                                              |
| --------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------------- |
| `pkg/types`     | Domain types: branded IDs, Language, CodeBlock, Result, ReportData, ValidationStatus | _(none — pure types)_                             | `testing` (stdlib only)                                           | All branded types, constructors, assertions                             |
| `pkg/code`      | Go stdlib parser utilities                                                           | _(none — stdlib only)_                            | _(none)_                                                          | `IndentCode`, `ParseGo`                                                 |
| `pkg/languages` | Language validation registry and validators                                          | `pkg/types`, `pkg/code`                           | `testing` (stdlib only)                                           | `Validator`, `Registry`, `GoValidator`, `TreeSitterValidator`           |
| `pkg/output`    | Multi-format report output                                                           | `pkg/types`, `go-output`                          | `pkg/types`                                                       | `PrintReport`, `PrintReportTo`, `Format`, `ColorMode`                   |
| `pkg/`          | Extraction, validation orchestration, context management                             | `pkg/types`, `pkg/languages`                      | `pkg/code`, `pkg/testutil`, `pkg/types`                           | `FileValidator`, `ExtractCodeBlocks`, `ValidateGoCode`, `ContextConfig` |
| `pkg/testutil`  | Test helpers (file writing, assertions)                                              | `pkg/types`                                       | `pkg/types`                                                       | `WriteTestFile`, `Assert*` functions                                    |
| `cmd/`          | CLI entry point                                                                      | `pkg`, `pkg/languages`, `pkg/output`, `pkg/types` | `pkg`, `pkg/languages`, `pkg/output`, `pkg/testutil`, `pkg/types` | _(none — package main)_                                                 |

### 3.5 What Moves Where

| Change                                  | Details                                                                                                                                     |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `Language` type → `pkg/types`           | Move `Language` branded type, its methods, constants (`LangGo`, etc.), `AllLanguages()`, `ParseLanguage()` to `pkg/types/language.go`       |
| `pkg/languages/language.go` → re-export | Keep `type Language = types.Language` alias + `AllLanguages`/`ParseLanguage` wrappers for backward compat, OR remove and update all callers |
| `pkg/types/code_block.go`               | Change `languages.Language` → `Language` (now in same package)                                                                              |
| `pkg/languages/validator.go`            | Import `pkg/types` for `Language` in interface signatures                                                                                   |
| All other files                         | Update imports as needed                                                                                                                    |

---

## 4. DAG Verification

### Before (broken — cycle):

```
pkg/types → pkg/languages → pkg/code
     ↑            │
     └────────────┘  (CYCLE via CodeBlock.Language field)
```

### After (clean — strict DAG):

```
cmd/md-go-validator
    → pkg
        → pkg/languages
            → pkg/code
            → pkg/types   ✓
        → pkg/types       ✓
    → pkg/output
        → pkg/types       ✓
    → pkg/types           ✓ (leaf)
```

**Proof of acyclicity:** `pkg/types` has zero internal imports. All other packages depend on it but nothing depends upward. `pkg/code` has zero internal imports. `pkg/languages` depends only on `pkg/types` and `pkg/code`. Strict DAG confirmed.

---

## 5. Replace / Workspace Strategy

**Not applicable.** Single module — no replace directives or go.work needed.

If the project later grows to publish `pkg/types` or `pkg/languages` as independent libraries, the migration path is straightforward because `pkg/types` will already be a self-contained, zero-dependency package.

---

## 6. Test Dependency Isolation

### After refactoring:

| Package         | Test-Only Internal Deps                              | Concern                                                                                                                                  |
| --------------- | ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/types`     | _(none)_                                             | Test helpers in `testing.go` only use stdlib `testing`                                                                                   |
| `pkg/languages` | _(none)_                                             | `testing.go` only uses stdlib `testing`                                                                                                  |
| `pkg/output`    | `pkg/types`                                          | Uses `types.AssertReportTotalAndValid` etc. — these are type-assertion helpers on `types.ReportData`, which lives in `pkg/types`. Clean. |
| `pkg/testutil`  | _(none — tests use `pkg/types` and `pkg/languages`)_ | testutil's own tests verify the helpers work — acceptable cross-package test deps                                                        |

No test-only dependencies leak into production code. Clean.

---

## 7. Interface Extraction Plan

Not needed. The `languages.Validator` interface is already cleanly defined in `pkg/languages/validator.go`. The `Registry` pattern provides proper interface-based decoupling. No further extraction required.

---

## 8. Versioning Strategy

**Root-only versioning** — single git tag `v1.x.x` via `goreleaser`. This is the correct strategy for a single-binary CLI tool. No module-level versioning needed.

---

## 9. Migration Strategy

See `EXECUTION_PLAN.md` for the step-by-step plan.

---

## 10. Risk Assessment

| Risk                                                             | Likelihood | Impact | Mitigation                                                         |
| ---------------------------------------------------------------- | ---------- | ------ | ------------------------------------------------------------------ |
| Breaking `Language` type move causes cascade of import changes   | Medium     | Low    | Use type alias initially, then remove                              |
| `pkg/languages` tests depend on `Language` being in same package | Medium     | Low    | Language test helpers stay in `pkg/languages`; only the type moves |
| `goreleaser` config breaks                                       | Low        | Medium | Verify `goreleaser release --snapshot` after changes               |
| CI pipeline fails                                                | Low        | Medium | Run full test suite + lint after each step                         |
| Linter complaints about new import patterns                      | Medium     | Low    | Run `golangci-lint` after each step, fix incrementally             |

---

## 11. Build System Impact

| System                     | Change Required                                               |
| -------------------------- | ------------------------------------------------------------- |
| `go.mod`                   | No changes (same module)                                      |
| `.goreleaser.yml`          | No changes                                                    |
| `.github/workflows/ci.yml` | No changes                                                    |
| `.golangci.yml`            | Possibly update `ireturn` exclusions if new interfaces appear |

---

## 12. Alternative Considered: Full Multi-Module Split

**Rejected** for these reasons:

1. **No external consumers** — `pkg/types`, `pkg/languages`, `pkg/output` are not imported by any external project
2. **Single binary deliverable** — goreleaser produces one binary; splitting modules adds build complexity for zero benefit
3. **Small project** — ~2.5k LOC across 7 packages is not enough to justify the overhead of go.work, multiple go.mods, replace directives
4. **Test coupling** — integration tests span multiple packages; splitting would require complex test dependency management
5. **When to reconsider:** If the project adds a plugin system, gRPC API, or becomes a library that others import, then extract `pkg/types` and `pkg/languages` as separate modules

---

## Summary of Key Decisions

1. **Single module** — keep one go.mod, no go.work
2. **Break the cycle** — move `Language` type to `pkg/types`
3. **`pkg/types` becomes a true leaf** — zero internal dependencies
4. **Clean DAG** — all arrows point downward
5. **Root-only versioning** — single semver tag via goreleaser

---

## 13. Self-Review Findings (Phase 4)

### Critical Review Questions

**1. What did we forget?**

- The `Language` type's `Extensions()` method is only used within `pkg/languages` tests — no external caller needs it. However, `Extensions()` is a natural method on a Language type and should move with it.
- `ParseLanguage` is called from 3 external sites: `cmd/md-go-validator/main.go`, `pkg/extractor.go`, and `pkg/languages/validator.go`. Moving it to `pkg/types` is correct since it's pure string parsing with zero dependencies.
- `AllLanguages()` is only called within `pkg/languages` (by `IsSupported()` and `DefaultRegistry`). After the move, `pkg/types` would own it, and `pkg/languages` would call `types.AllLanguages()`.

**2. What could be done better?**

- Consider a two-phase approach: (a) move the type, (b) clean up backward compat aliases. This avoids a big-bang refactor.
- The `languages.Language` usage is spread across 35 call sites. A bulk find-replace is risky. Better to move the type definition, then update imports one package at a time.

**3. Split brains check?**

- No split brains will be introduced. `Language` will exist in exactly one place (`pkg/types`). `pkg/languages` will import `pkg/types` and reference it as `types.Language` or via a type alias.
- **Recommendation:** Do NOT use a type alias in `pkg/languages`. Just update all call sites to use `types.Language` directly. Type aliases create confusion about where the "real" type lives.

**4. Module boundaries at the right granularity?**

- Yes. Moving only `Language` (a branded string) to `pkg/types` is the minimal change that breaks the cycle. No need to move `Validator`, `Registry`, `ErrorCode`, or `ValidationError` — those belong in `pkg/languages`.

**5. Can existing code be reused?**

- Yes. The `Language` type is already self-contained in `pkg/languages/language.go`. It only imports `errors`, `fmt`, `slices`, `strings` — all stdlib. Moving it is a pure file move + import update.

**6. Type model improvements?**

- `Language` is already a well-designed branded type. No changes needed.

**7. Banned dependencies check?**

- `gotreesitter` — not in the banned list. Pure Go tree-sitter. Acceptable.
- `go-output` — Lars's own library. Acceptable.
- No banned dependencies found.

**8. Will CI be faster?**

- N/A — no CI speed improvement expected. The goal is structural cleanliness, not performance.

**9. Test-only dependency isolation?**

- After the move, `pkg/types/types_test.go` will no longer import `pkg/languages`. The only usage is `languages.LangGo` in `TestCodeBlock`, which will become `LangGo` (same package). Clean.

**10. Versioning strategy realistic?**

- Root-only versioning is correct for a single-binary CLI. No changes needed.

### Updated Decisions Based on Self-Review

1. **No type alias** — update all call sites to use `types.Language` directly
2. **Move everything from `language.go`** — `Language` type, constants, `AllLanguages()`, `ParseLanguage()`, `Extensions()`, `Validate()`, `IsSupported()` — all move to `pkg/types/language.go`
3. **Delete `pkg/languages/language.go`** — after the move, this file becomes empty; delete it entirely
4. **Update imports in 6 packages** — `pkg/`, `pkg/extractor.go`, `pkg/validator.go`, `cmd/md-go-validator/main.go`, `pkg/types/types_test.go`, and all other files using `languages.Language` or `languages.Lang*`
5. **One commit per package** — each import update is a separate, revertable commit
