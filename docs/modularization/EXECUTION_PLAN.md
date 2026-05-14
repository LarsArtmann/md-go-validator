# Execution Plan — md-go-validator Modularization

**Date:** 2026-05-13 | **Status:** Ready for Execution

---

## Overview

This plan resolves the `pkg/types` ↔ `pkg/languages` dependency cycle by moving the `Language` type to `pkg/types`. The project remains a single Go module.

**Total estimated effort:** 2–3 hours

---

## Task Dependency Graph

```
Task 1 (create pkg/types/language.go)
    │
    ├──→ Task 2 (update pkg/types internal references)
    │
    ├──→ Task 3 (update pkg/ references)
    │
    ├──→ Task 4 (update pkg/languages references)
    │
    ├──→ Task 5 (update cmd/ references)
    │
    ├──→ Task 6 (update test references)
    │
    └──→ Task 7 (delete old file + final verification)
```

---

## Tasks

### Task 1: Create `pkg/types/language.go` [1% → 51% impact]

**What:** Copy the entire contents of `pkg/languages/language.go` to `pkg/types/language.go`, changing only the package declaration.

**Steps:**

1. Create `pkg/types/language.go` with package `types`
2. Contents: `Language` type, all constants (`LangGo`, `LangTempl`, etc.), `AllLanguages()`, `ParseLanguage()`, `Extensions()`, `Validate()`, `IsSupported()`, and `errUnsupportedLang`
3. Remove the import of `pkg/languages` from `pkg/types/code_block.go` (change `languages.Language` → `Language`)
4. Update `NewCodeBlock` signature to use `Language` (same package)

**Verification:**

```bash
go build ./pkg/types
go test ./pkg/types/...
```

**Rollback:** `git checkout -- pkg/types/`

**Effort:** 15 min

---

### Task 2: Update `pkg/types` internal references

**What:** Fix `pkg/types/types_test.go` to use `LangGo` directly instead of `languages.LangGo`.

**Steps:**

1. Remove `"github.com/larsartmann/md-go-validator/pkg/languages"` import from `types_test.go`
2. Change `languages.LangGo` → `LangGo` (line 225, 230)

**Verification:**

```bash
go test ./pkg/types/...
```

**Rollback:** `git checkout -- pkg/types/types_test.go`

**Effort:** 5 min

---

### Task 3: Update `pkg/` references [4% → 64% impact]

**What:** Update all `languages.Language` and `languages.Lang*` references in `pkg/` to use `types.Language` and `types.Lang*`.

**Files to update:**

- `pkg/extractor.go` — imports `languages.Language`, `languages.LangGo`, `languages.ParseLanguage`
- `pkg/validator.go` — imports `languages.Language`, `languages.LangGo`
- `pkg/parser.go` — imports `pkg/languages` (may need `pkg/types` instead)

**Steps:**

1. Add `"github.com/larsartmann/md-go-validator/pkg/types"` import where needed
2. Change `languages.Language` → `types.Language`
3. Change `languages.LangGo` → `types.LangGo`
4. Change `languages.ParseLanguage` → `types.ParseLanguage`
5. Remove `pkg/languages` import if no longer needed
6. Keep `pkg/languages` import where `Registry` or `Validator` types are used

**Verification:**

```bash
go build ./pkg
go test ./pkg/...
```

**Rollback:** `git checkout -- pkg/extractor.go pkg/validator.go pkg/parser.go`

**Effort:** 20 min

---

### Task 4: Update `pkg/languages` references [4% → 64% impact]

**What:** Update `pkg/languages` to import `Language` from `pkg/types` instead of defining it locally.

**Files to update:**

- Delete `pkg/languages/language.go` (or gut it)
- `pkg/languages/validator.go` — uses `Language` type in `Registry`, `Validator` interface, `GetByString`
- `pkg/languages/go_validator.go` — `GoValidator.Language()` returns `Language`
- `pkg/languages/treesitter_validator.go` — `TreeSitterValidator.Language()` returns `Language`

**Steps:**

1. Add `"github.com/larsartmann/md-go-validator/pkg/types"` import to `validator.go`
2. Change all `Language` references to `types.Language` (it's no longer defined in this package)
3. Update `Validator` interface: `Language() types.Language`
4. Update `Registry` struct: `map[types.Language]Validator`
5. Update `GetByString` to call `types.ParseLanguage` instead of local `ParseLanguage`
6. Update `GoValidator.Language()` return type
7. Update `TreeSitterValidator.Language()` return type and constructor
8. Delete `pkg/languages/language.go`
9. Update `pkg/languages/doc.go` if needed

**Verification:**

```bash
go build ./pkg/languages
go test ./pkg/languages/...
```

**Rollback:** `git checkout -- pkg/languages/`

**Effort:** 25 min

---

### Task 5: Update `cmd/md-go-validator` references

**What:** Update CLI entry point to use `types.Language`, `types.LangGo`, `types.ParseLanguage`.

**Files to update:**

- `cmd/md-go-validator/main.go`
- `cmd/md-go-validator/main_test.go`

**Steps:**

1. Add `"github.com/larsartmann/md-go-validator/pkg/types"` import
2. Change `languages.Language` → `types.Language`
3. Change `languages.LangGo` → `types.LangGo`
4. Change `languages.ParseLanguage` → `types.ParseLanguage`
5. Keep `pkg/languages` import for `Registry`, `DefaultRegistry`, `Validator`

**Verification:**

```bash
go build ./cmd/md-go-validator
go test ./cmd/md-go-validator/...
```

**Rollback:** `git checkout -- cmd/`

**Effort:** 15 min

---

### Task 6: Update remaining test references

**What:** Fix all test files that reference `languages.Language` or `languages.Lang*`.

**Files to update:**

- `pkg/validator_test.go` — `languages.Language`, `languages.LangGo`
- `pkg/integration_test.go` — `languages.LangGo`, `languages.LangTypeScript`, `languages.LangRust`
- `pkg/benchmark_test.go` — `languages.Language`, `languages.LangGo`
- `pkg/testutil/testutil_test.go` — `languages.LangGo`
- `pkg/languages/language_test.go` — update or delete tests for moved functions
- `pkg/languages/validator_test.go` — may use `Language` directly (same package after move)
- `pkg/languages/go_validator_test.go` — uses `LangGo` (now needs `types.LangGo`)
- `pkg/languages/treesitter_validator_test.go` — uses `LangRust`, `LangTypeScript`, etc.

**Steps:**

1. Update imports in each test file
2. Change `languages.LangGo` → `types.LangGo` etc.
3. For `pkg/languages/*_test.go` files: since they're in `package languages`, they need `import types` and `types.LangGo`
4. Move or update `TestAllLanguages`, `TestParseLanguage`, `TestExtensions`, `TestValidate`, `TestIsSupported` to `pkg/types/` test files
5. Remove duplicate tests from `pkg/languages/language_test.go`

**Verification:**

```bash
go test ./...
```

**Rollback:** `git checkout -- *_test.go pkg/languages/*_test.go`

**Effort:** 30 min

---

### Task 7: Final verification and cleanup [20% → 80% impact]

**What:** Ensure everything builds, passes tests, passes lint, and has no remaining `pkg/languages/language.go`.

**Steps:**

1. Verify `pkg/languages/language.go` is deleted
2. Run full build: `go build ./...`
3. Run full tests: `go test -race -cover ./...`
4. Run lint: `golangci-lint run ./...`
5. Run `go mod tidy`
6. Verify no remaining references to `languages.Language` in `pkg/types/` (should only be `types.Language` now)
7. Verify `pkg/types/` has zero internal imports: `grep -r 'md-go-validator/pkg/' pkg/types/*.go | grep -v types_test`

**Verification:**

```bash
go build ./...
go test -race -cover ./...
golangci-lint run ./...
go mod tidy
go vet ./...
```

**Rollback:** N/A — this is the final state

**Effort:** 15 min

---

## Impact Summary

| Tier      | Tasks                                      | Impact                                   |
| --------- | ------------------------------------------ | ---------------------------------------- |
| 1% → 51%  | Task 1 (create language.go in types)       | Breaks the cycle at its root             |
| 4% → 64%  | Tasks 3, 4 (update pkg/ and pkg/languages) | Makes the cycle break visible everywhere |
| 20% → 80% | Tasks 5, 6, 7 (cmd, tests, cleanup)        | Complete the migration, all green        |

---

## Risk Mitigation

| Risk                               | Mitigation                                                                                                         |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Import cycle re-introduced         | CI will catch this — Go compiler refuses cycles                                                                    |
| Linter failures after move         | Run `golangci-lint` after each task, fix incrementally                                                             |
| Missed `languages.Lang*` reference | `grep -r 'languages\.Lang\|languages\.Language\|languages\.ParseLanguage\|languages\.AllLanguages'` will catch all |
| Test regression                    | Full test suite after each task                                                                                    |
| `ireturn` linter issue             | The `Validator.Language()` returns `types.Language` — may need `.golangci.yml` update                              |

---

## Post-Migration Verification Checklist

- [ ] `go build ./...` passes
- [ ] `go test -race -cover ./...` passes
- [ ] `golangci-lint run ./...` passes with 0 issues
- [ ] `go mod tidy` changes nothing
- [ ] `go vet ./...` reports no issues
- [ ] `pkg/types/` has zero internal package imports (verify with grep)
- [ ] No `languages.Language` or `languages.Lang*` references remain outside `pkg/languages/`
- [ ] `pkg/languages/language.go` is deleted
- [ ] Coverage has not decreased
- [ ] `.goreleaser.yml` still works: `goreleaser release --snapshot --clean`
