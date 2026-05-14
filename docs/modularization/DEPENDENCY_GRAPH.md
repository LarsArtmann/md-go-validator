# Dependency Analysis — md-go-validator

**Date:** 2026-05-13

---

## Current Dependency Graph

### Internal Package Dependencies (Production Code)

```
┌─────────────────────────────────────────────────────┐
│                  cmd/md-go-validator                 │
│    imports: pkg, pkg/languages, pkg/output,          │
│             pkg/types                                │
└──────────┬──────────┬──────────┬────────────────────┘
           │          │          │
           ▼          ▼          ▼
┌──────────────┐ ┌──────────┐ ┌──────────────┐
│     pkg      │ │ pkg/output│ │  pkg/types   │
│              │ │          │ │              │
│ imports:     │ │ imports: │ │ imports:     │
│  pkg/languages│ │ pkg/types│ │  pkg/languages│ ← CYCLE
│  pkg/types   │ │ go-output│ │              │
└──────┬───────┘ └────┬─────┘ └──────┬───────┘
       │              │               │
       ▼              │               │
┌──────────────┐      │               │
│ pkg/languages│◄─────┼───────────────┘
│              │      │
│ imports:     │      │
│  pkg/code    │      │
│  gotreesitter│      │
└──────┬───────┘      │
       │              │
       ▼              ▼
┌──────────────┐ ┌──────────────┐
│   pkg/code   │ │  pkg/testutil│
│              │ │              │
│ imports:     │ │ imports:     │
│  stdlib only │ │  pkg/types   │
│              │ │  stdlib      │
└──────────────┘ └──────────────┘
```

### The Cycle in Detail

```
pkg/types/code_block.go:
    import "github.com/larsartmann/md-go-validator/pkg/languages"

    type CodeBlock struct {
        Language languages.Language  // ← upward dependency
    }

pkg/languages/language.go:
    type Language string  // ← definition lives here
```

This means:

- `pkg/types` cannot be used without pulling in `pkg/languages` (and transitively `pkg/code`, `gotreesitter`)
- `pkg/languages` cannot import `pkg/types` (compile-time cycle)
- Any future consumer of just the types gets the entire validation infrastructure

---

## Proposed Dependency Graph

### After Moving `Language` to `pkg/types`

```
┌─────────────────────────────────────────────────────┐
│                  cmd/md-go-validator                 │
│    imports: pkg, pkg/languages, pkg/output,          │
│             pkg/types                                │
└──────────┬──────────┬──────────┬────────────────────┘
           │          │          │
           ▼          ▼          ▼
┌──────────────┐ ┌──────────┐ ┌──────────────┐
│     pkg      │ │ pkg/output│ │  pkg/types   │
│              │ │          │ │              │
│ imports:     │ │ imports: │ │ imports:     │
│  pkg/languages│ │ pkg/types│ │  (none)      │ ← LEAF
│  pkg/types   │ │ go-output│ │              │
└──────┬───────┘ └────┬─────┘ └──────────────┘
       │              │
       ▼              │
┌──────────────┐      │
│ pkg/languages│      │
│              │      │
│ imports:     │      │
│  pkg/types   │◄─────┘
│  pkg/code    │
│  gotreesitter│
└──────┬───────┘
       │
       ▼
┌──────────────┐ ┌──────────────┐
│   pkg/code   │ │  pkg/testutil│
│              │ │              │
│ imports:     │ │ imports:     │
│  (none)      │ │  pkg/types   │
│              │ │  stdlib      │
└──────────────┘ └──────────────┘
```

### Strict DAG — All arrows point downward

| Package         | Level    | Internal Dependencies                             |
| --------------- | -------- | ------------------------------------------------- |
| `pkg/types`     | 0 (leaf) | _(none)_                                          |
| `pkg/code`      | 0 (leaf) | _(none)_                                          |
| `pkg/testutil`  | 0 (leaf) | `pkg/types`                                       |
| `pkg/languages` | 1        | `pkg/types`, `pkg/code`                           |
| `pkg/output`    | 1        | `pkg/types`                                       |
| `pkg/`          | 2        | `pkg/types`, `pkg/languages`                      |
| `cmd/`          | 3        | `pkg`, `pkg/languages`, `pkg/output`, `pkg/types` |

---

## External Dependency Map

| Package         | Direct External Dependencies                                     |
| --------------- | ---------------------------------------------------------------- |
| `pkg/types`     | _(none)_                                                         |
| `pkg/code`      | _(none — stdlib `go/parser`, `go/token`)_                        |
| `pkg/languages` | `gotreesitter v0.15.3` (+ grammars)                              |
| `pkg/output`    | `go-output v0.2.0`                                               |
| `pkg/`          | _(none — all external deps are via pkg/languages and pkg/types)_ |
| `pkg/testutil`  | _(none)_                                                         |
| `cmd/`          | _(none — all external deps are via transitive imports)_          |

### External Dependency Burden

If a consumer wants ONLY the domain types (e.g., to build a custom validator):

- **Current:** Must import `gotreesitter`, `go-output`, and all their transitive deps
- **After fix:** `pkg/types` has ZERO external deps — clean import

---

## Coupling Metrics

| Metric                           | Before                    | After                           |
| -------------------------------- | ------------------------- | ------------------------------- |
| Cycles in dependency graph       | 1 (`types` ↔ `languages`) | 0                               |
| Packages with zero internal deps | 2 (`code`, `testutil`)    | 3 (`types`, `code`, `testutil`) |
| Max dependency depth             | 3                         | 3                               |
| Packages importing `pkg/types`   | 6                         | 6                               |
| Packages `pkg/types` imports     | 1 (`pkg/languages`)       | 0                               |

---

## Test Dependency Graph

```
cmd/ tests → pkg, pkg/languages, pkg/output, pkg/testutil, pkg/types
pkg/ tests → pkg/code, pkg/languages, pkg/testutil, pkg/types
pkg/types tests → pkg/types (no cross-pkg after fix)
pkg/languages tests → (none — same package)
pkg/output tests → pkg/types
pkg/code tests → (none — same package)
pkg/testutil tests → pkg/languages, pkg/types
```

Note: After the fix, `pkg/types` tests will NOT need `pkg/languages` — the cycle is broken at the test level too.
