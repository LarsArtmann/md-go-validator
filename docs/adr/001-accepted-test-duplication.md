# ADR 001: Accepted Test Code Duplication

**Status:** Accepted
**Date:** 2026-06-18
**Deciders:** Project maintainers

## Context

`art-dupl` reports residual test-code clone groups that we have chosen to keep.
The skill `deduplicate-code` distinguishes _harmful_ duplication from
_idiomatic_ duplication and says: "Zero harmful duplication. Not zero report
lines."

This ADR records the deliberate acceptance of six clone groups so that future
maintainers do not refactor them and so the deduplication contract is explicit.

## Decision

Each remaining clone group below is preserved on purpose. None of them would
become more readable or maintainable if extracted.

### 1. `pkg/extractor_test.go:164-166` ↔ `pkg/extractor_test.go:186-188`

```go
if !blocks[0].IsSkipped() {
    t.Errorf("expected skipped block, got status %s", blocks[0].Status)
}
```

**Rationale:** Two distinct scenarios (default directive vs. custom directive)
asserting the same outcome on independently extracted blocks. A helper would
require a `t.Helper()` marker, a name, and a signature — making it longer than
the duplication it replaces. The duplication is a literal domain invariant
("this block must be skipped"), not a shared implementation.

### 2. `pkg/extractor_test.go:217-219` ↔ `pkg/extractor_test.go:232-234`

```go
if blocks[0].Language != languages.LangGo {
    t.Errorf("expected go, got %s", blocks[0].Language)
}
```

**Rationale:** Two distinct scenarios (literal `go` fence vs. `golang` alias)
asserting the same language tag after extraction. Same trade-off as #1: a
helper would not be shorter than the duplicated assertion. The two tests
exist precisely to verify that the alias resolves to the same language.

### 3. `pkg/testutil/testutil_test.go:290-294` ↔ `pkg/testutil/testutil_test.go:296-300`

```go
t.Run("fail when strings differ", func(t *testing.T) {
    t.Parallel()
    assertZeroValueFails(t, "string", "a", "b")
})
```

**Rationale:** Subtests with explicit, descriptive names that surface in
`-run` and IDE test runners. The shared assertion logic is already factored
into `assertZeroValueFails`. Collapsing into a table would lose the named
subtests — the idiomatic Go pattern for parameterised cases.

### 4. `pkg/validator_test.go:688-693` ↔ `pkg/validator_test.go:706-711`

```go
results := validate(t, "<content>")
testutil.AssertResultCount(t, results, N)
```

**Rationale:** Three sibling subtests validating different Markdown shapes
against `ValidateContent`. The shared boilerplate is already extracted into
the `validate` closure; the remaining two lines carry the per-case content
and expected count, which is the whole point of the test.

### 5. `pkg/output/output_test.go:146-148` ↔ `pkg/types/types_test.go:473-475`

```go
if report.Errors[0].Error != "<expected>" {
    t.Errorf("expected error message '<expected>', got %q", report.Errors[0].Error)
}
```

**Rationale:** Two assertions in two different packages, each verifying a
different error-message contract (the output formatter must surface the
validator's syntax message; the report builder must not invent a message
when the error is nil). The shared shape is the standard Go comparison
pattern; the divergent content is the actual test intent.

### 6. `cmd/md-go-validator/main_test.go:76-78` ↔ `cmd/md-go-validator/main_test.go:809-811`

```go
if len(cfg.paths) != 1 || cfg.paths[0] != "<expected>" {
    t.Errorf("expected paths=[<expected>], got %v", cfg.paths)
}
```

**Rationale:** Two assertions in two different test functions verifying that
`parseArgs` produces the right default path (`.`) and that an explicit
positional argument (`src/`) is preserved. The shared shape is a standard
slice-length-and-head equality check.

## Consequences

- **No further refactor attempts** for these six clone groups without
  revisiting this ADR.
- **Threshold:** We keep the deduplication threshold at `-t 20`. Lowering it
  would surface Go idioms (`if err != nil`, function signatures, error
  returns) that are not worth eliminating.
- **Future audits:** A reviewer running `art-dupl --semantic --sort
total-tokens -t 20` should expect exactly these six groups. New groups
  signal real duplication; missing groups signal accidental removal.

## Verification

```bash
art-dupl --semantic --sort total-tokens -t 20
# Expected: 6 clone groups, all in test code, all listed above.
```
