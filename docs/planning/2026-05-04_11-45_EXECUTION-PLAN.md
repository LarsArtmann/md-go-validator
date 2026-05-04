# Comprehensive Execution Plan - md-go-validator

**Generated:** 2026-05-04  
**Priority:** Based on impact/work ratio

---

## PHASE 1: Critical Gaps (Must Fix)

### 1.1 Add tests for `pkg/code` (0% coverage) - IMPACT: HIGH, WORK: LOW

**Files:** `pkg/code/util.go`

| Function | What to Test |
|----------|--------------|
| `IndentCode` | Empty string, single line, multi-line, trailing newlines, only whitespace |
| `ParseGo` | Valid code, invalid code, empty string |

**Test file:** `pkg/code/util_test.go`

### 1.2 Add tests for `pkg/testutil` (0% coverage) - IMPACT: HIGH, WORK: LOW

**Files:** `pkg/testutil/testutil.go`

| Function | What to Test |
|----------|--------------|
| `WriteTestFile` | Creates file, correct content, handles errors |
| `AssertResultCount` | Pass/fail cases |
| `AssertMinResults` | Pass/fail cases |
| `AssertMaxResults` | Pass/fail cases |
| `AssertBlockCount` | Pass/fail cases |
| `AssertContextNotNil` | Nil vs non-nil |
| `AssertContextCondition` | Done/not-done states |
| `AssertContextErr` | Match/mismatch |
| `AssertZeroValue` | Generic helper |

**Test file:** `pkg/testutil/testutil_test.go`

---

## PHASE 2: Quick Wins (High Impact, Low Work)

### 2.1 Fix pre-commit hook - IMPACT: LOW, WORK: TRIVIAL

Either make executable or remove:
```bash
chmod +x .git/hooks/pre-commit
# OR remove if not needed
rm .git/hooks/pre-commit
```

### 2.2 Remove `justfile` (deprecated) - IMPACT: MEDIUM, WORK: LOW

The AGENTS.md says it's deprecated. If we have flake.nix, justfile adds confusion.

### 2.3 Remove stale documentation - IMPACT: LOW, WORK: LOW

Check if these are still relevant:
- `CLONE_ANALYSIS.md`
- `REFLECTION_AND_PLAN.md`

### 2.4 Export `SupportedExtensions()` - IMPACT: MEDIUM, WORK: LOW

Add public API for library users:
```go
// SupportedExtensions returns all supported file extensions.
func SupportedExtensions() []string { ... }

// IsSupportedFile checks if a file is supported.
func IsSupportedFile(path string) bool { ... }
```

---

## PHASE 3: Architecture Improvements (Medium Impact)

### 3.1 Add `FileType` branded type - IMPACT: MEDIUM, WORK: LOW

Currently extensions are raw strings. Create:
```go
// FileType represents a markdown file extension.
type FileType string

const (
    FileTypeMarkdown FileType = ".md"
    FileTypeMdx      FileType = ".mdx"
    // ...
)
```

### 3.2 Improve `cmd` test coverage (61.7% → 80%) - IMPACT: MEDIUM, WORK: MEDIUM

Missing coverage:
- Error paths (invalid paths, non-existent files)
- All output formats
- Language flag edge cases
- Timeout/cancellation

### 3.3 Add `pkg/languages` coverage (66.7% → 80%) - IMPACT: MEDIUM, WORK: MEDIUM

Tree-sitter validators are skipped when grammars unavailable - need to test this path.

---

## PHASE 4: Testing Infrastructure (Medium Impact, Medium Work)

### 4.1 Add integration tests - IMPACT: MEDIUM, WORK: MEDIUM

Create `testdata/` directory with real .md and .mdx files:
```
testdata/
├── valid/
│   ├── simple.md
│   └── with_go_code.md
├── invalid/
│   └── syntax_error.md
├── skipped/
│   └── with_skip_directive.md
└── mdx/
    └── component.mdx
```

### 4.2 Add benchmark tests - IMPACT: MEDIUM, WORK: LOW

For hot paths:
- `ExtractCodeBlocks` on large files
- `ValidateGoCode` parsing strategies
- Directory walking

### 4.3 Add fuzz tests - IMPACT: MEDIUM, WORK: MEDIUM

For:
- `ExtractCodeBlocks` edge cases
- `ParseGo` with malformed input

---

## PHASE 5: CI/CD (High Impact)

### 5.1 Add golangci-lint to CI - IMPACT: HIGH, WORK: LOW

```yaml
- name: Lint
  run: golangci-lint run ./...
```

### 5.2 Add `go test -race` to CI - IMPACT: MEDIUM, WORK: LOW

```yaml
- name: Race Tests
  run: go test -race ./...
```

### 5.3 Add test with coverage - IMPACT: MEDIUM, WORK: LOW

```yaml
- name: Test with Coverage
  run: go test -cover ./...
```

---

## PHASE 6: Future Enhancements (Lower Priority)

### 6.1 Consider `samber/lo` for functional utilities

Already in go-ecosystem reference. Could use for:
- `lo.Map`, `lo.Filter` instead of manual loops
- `lo.Try` for error handling

### 6.2 Consider `gookit/validate` for CLI arg validation

Currently manual validation. Could use library for:
- Built-in validators
- Better error messages

### 6.3 Add goreleaser cross-compilation CI

Already has `.goreleaser.yml` - just need CI to trigger it.

---

## EXECUTION ORDER

| Step | Task | Phase | Status |
|------|------|-------|--------|
| 1 | `pkg/code` tests | 1.1 | TODO |
| 2 | `pkg/testutil` tests | 1.2 | TODO |
| 3 | Fix pre-commit hook | 2.1 | TODO |
| 4 | Export `SupportedExtensions()` | 2.4 | TODO |
| 5 | Add `FileType` branded type | 3.1 | TODO |
| 6 | Remove stale docs | 2.3 | TODO |
| 7 | Remove `justfile` | 2.2 | TODO |
| 8 | Improve `cmd` coverage | 3.2 | TODO |
| 9 | Add integration tests | 4.1 | TODO |
| 10 | Add CI pipeline | 5 | TODO |

---

## GHOST SYSTEMS TO REVIEW

1. **`CLONE_ANALYSIS.md`** - Clone-specific analysis, likely stale
2. **`REFLECTION_AND_PLAN.md`** - Old planning doc, check if still relevant
3. **`justfile`** - Deprecated but still exists, creates confusion

---

## LIBRARIES TO CONSIDER

| Library | Use Case | Priority |
|---------|----------|----------|
| `samber/lo` | Functional utilities (Map, Filter, Try) | Low (code is simple enough) |
| `gookit/validate` | CLI arg validation | Low (current manual validation works) |
| `stretchr/testify` | Testing assertions | Medium (current testutil works) |

---

## NOTES

- Project has **strong type system** with branded types (FileID, LineNumber, BlockIndex, Language, ValidationStatus)
- **Good separation** between packages
- **0 golangci-lint issues** - well maintained
- **go-output** now properly integrated from GitHub
- **Next priority**: Fix 0% coverage packages (1.1, 1.2)
