# Comprehensive Status Report — MDX Support & Architecture Improvements

**Date:** 2026-05-04 10:08  
**Session Focus:** Add `.mdx` file support, then critically review and improve the implementation

---

## A) FULLY DONE

| #   | Item                                                    | Commit    | Detail                                                               |
| --- | ------------------------------------------------------- | --------- | -------------------------------------------------------------------- |
| 1   | Add `.mdx` to extension check                           | `1624cea` | `supportedExtensions` map is the single source of truth              |
| 2   | Extract extensions into `supportedExtensions` map       | `1624cea` | O(1) lookup, easy to add future extensions                           |
| 3   | Rename `isMarkdownFile` → `isSupportedFile`             | `edf8003` | Name now correctly reflects all supported types                      |
| 4   | Rename `collectMarkdownFiles` → `collectSupportedFiles` | `49295a8` | Method name accurate for .md/.markdown/.mdx                          |
| 5   | Add 10 unit tests for `isSupportedFile`                 | `e408ff2` | All 3 extensions, case-insensitive variants, 5 negative cases        |
| 6   | Derive verbose message dynamically                      | `7bede34` | `formatSupportedExtensions()` eliminates hardcoded extension strings |
| 7   | Update stale doc comments across `pkg/`                 | `c9cf503` | 6 comments updated: extractor, validator, language, types            |
| 8   | Fix CLI help header + package comment                   | `33a4109` | "Markdown and MDX files" in both places                              |
| 9   | Update README.md                                        | `80a582d` | Description + supported file types mention                           |
| 10  | Update CHANGELOG.md                                     | `80a582d` | MDX entry under `[Unreleased] / Added`                               |
| 11  | Update AGENTS.md                                        | `80a582d` | Overview mentions MDX                                                |
| 12  | Add MDX tests in `pkg/validator_test.go`                | `80a582d` | `ValidateFile_MDX`, `ValidateDirectory_MDX`                          |
| 13  | Add MDX test in `cmd/md-go-validator/main_test.go`      | `80a582d` | `directory_with_MDX_files` subtest                                   |
| 14  | Push all commits to remote                              | Done      | 8 commits on master                                                  |

---

## B) PARTIALLY DONE

| #   | Item                  | Status | What's Left                                                                                                                                            |
| --- | --------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | Test name consistency | Stale  | Two cmd test names still say "valid markdown file" / "directory with markdown files" — they should say "supported file" or similar to reflect .mdx too |

---

## C) NOT STARTED

| #   | Item                                                           | Priority | Detail                                                  |
| --- | -------------------------------------------------------------- | -------- | ------------------------------------------------------- |
| 1   | `pkg/code/util.go` tests                                       | High     | `IndentCode` and `ParseGo` have 0% coverage             |
| 2   | `pkg/testutil/testutil.go` tests                               | Medium   | 7+ exported helpers, 0% coverage                        |
| 3   | `cmd/md-go-validator` coverage improvement                     | Medium   | Currently 60.8% — many paths untested                   |
| 4   | `pkg/languages` coverage improvement                           | Medium   | Currently 66.7%                                         |
| 5   | Export `SupportedExtensions` / `IsSupportedFile` as public API | Low      | Library users may want to check extensions              |
| 6   | `--extension` CLI flag for custom extensions                   | Low      | Allow users to add custom file types                    |
| 7   | MDX-specific integration test with JSX syntax                  | Medium   | Test that MDX files with JSX components parse correctly |
| 8   | Example directory (`examples/`) with sample .md and .mdx files | Low      | Referenced in old status reports as a TODO              |

---

## D) TOTALLY FUCKED UP

| #   | Item                              | Detail                                                                                                                                                                                                                                                                                         |
| --- | --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `pkg/output/output.go` LSP errors | 16 compile errors from broken `go-output` import — the LSP/gopls cannot resolve the local `replace` directive (`=> ../go-output`) properly. Tests still pass because Go compiler handles `replace` correctly, but IDE experience is broken. **Pre-existing issue, not caused by our changes.** |
| 2   | Pre-commit hook not executable    | `.git/hooks/pre-commit` exists but isn't set as executable — git warns on every commit. **Pre-existing.**                                                                                                                                                                                      |
| 3   | `justfile` still exists           | AGENTS.md says "justfile is deprecated" and should be migrated to `flake.nix`, but it's still present. **Pre-existing.**                                                                                                                                                                       |

---

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **Single source of truth is now good** — `supportedExtensions` map is the canonical list. Adding a new extension requires exactly one change.
2. **Naming is now accurate** — `isSupportedFile` and `collectSupportedFiles` don't mislead.
3. **Dynamic formatting** — `formatSupportedExtensions()` means the verbose message can't go stale.

### Still Needs Improvement

1. **Extension config is private** — `supportedExtensions` is unexported. Library users can't query or extend it. Consider a public API.
2. **No file-type detection on `ValidateFile`** — Currently accepts any file path. Could validate that the extension is supported before processing (better error message for `.txt` files passed directly).
3. **`pkg/code` has zero tests** — Critical utility package, untested.
4. **`pkg/testutil` has zero tests** — Test infrastructure itself is untested.
5. **cmd coverage at 60.8%** — Missing coverage for error paths, edge cases in arg parsing.
6. **Test names inconsistent** — "valid markdown file" vs "directory with MDX files" vs "directory with markdown files" — should use consistent terminology.

### Type Model Improvements

7. **`FileType` branded type** — Currently extensions are raw strings. A `FileType` type (like `FileID`, `LineNumber`) would make the type system stronger and prevent mixing extensions with other strings.
8. **`SupportedExtensions()` public func** — Returns the canonical list for library consumers.
9. **`FileExtension(path) FileType`** — Extract extension with type safety.

### Library Ecosystem

10. **`go-output` is a local replace** — Consider publishing it or vendoring to improve portability.
11. **Pre-commit hook** — Make executable or remove.
12. **justfile → flake.nix** — Per AGENTS.md directive.

---

## F) TOP 25 THINGS TO DO NEXT

Sorted by **impact / work ratio** (highest first):

| Rank | Item                                                                             | Impact | Work    | Category    |
| ---- | -------------------------------------------------------------------------------- | ------ | ------- | ----------- |
| 1    | Fix stale cmd test names ("markdown file" → "supported file")                    | Medium | Trivial | Consistency |
| 2    | Add tests for `pkg/code/util.go` (`IndentCode`, `ParseGo`)                       | High   | Low     | Testing     |
| 3    | Add file-type validation in `ValidateFile` (reject unsupported extensions early) | Medium | Low     | UX          |
| 4    | Add `FileType` branded type for extensions                                       | Medium | Low     | Types       |
| 5    | Export `SupportedExtensions()` and `IsSupportedFile()` as public API             | Medium | Low     | API         |
| 6    | Add MDX integration test with JSX content                                        | Medium | Low     | Testing     |
| 7    | Make pre-commit hook executable                                                  | Low    | Trivial | DevEx       |
| 8    | Increase `cmd` test coverage (error paths, edge cases)                           | Medium | Medium  | Testing     |
| 9    | Increase `pkg/languages` test coverage (currently 66.7%)                         | Medium | Medium  | Testing     |
| 10   | Add `--extension` CLI flag for custom file types                                 | Medium | Medium  | Feature     |
| 11   | Add tests for `pkg/testutil/testutil.go`                                         | Low    | Medium  | Testing     |
| 12   | Publish or vendor `go-output` to fix LSP resolution                              | High   | Medium  | Infra       |
| 13   | Add `examples/` directory with sample .md and .mdx files                         | Low    | Medium  | Docs        |
| 14   | Add `.mdx` mention to README supported file types table                          | Low    | Trivial | Docs        |
| 15   | Migrate justfile → flake.nix                                                     | Medium | High    | Build       |
| 16   | Add context-aware error messages (include file type in errors)                   | Low    | Low     | UX          |
| 17   | Add `WithExtensions()` option to `FileValidator`                                 | Medium | Low     | API         |
| 18   | Benchmark tests for large .mdx files with many JSX components                    | Low    | Medium  | Perf        |
| 19   | Add `FileExtension(path) FileType` helper                                        | Low    | Low     | Types       |
| 20   | Stream-based file processing for very large files                                | Low    | High    | Perf        |
| 21   | Fuzz testing for extractor/parser                                                | Low    | Medium  | Testing     |
| 22   | GitHub Actions: add .mdx file to test fixtures                                   | Low    | Trivial | CI          |
| 23   | Remove `CLONE_ANALYSIS.md` and `REFLECTION_AND_PLAN.md` if stale                 | Low    | Trivial | Cleanup     |
| 24   | Add goreleaser config for cross-compiled binaries                                | Low    | Medium  | Release     |
| 25   | Consider `embed` for default config instead of hardcoded values                  | Low    | Medium  | Arch        |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should `ValidateFile` reject unsupported file extensions (e.g., `.txt`, `.go`) or silently process them?**

Currently, `ValidateFile` accepts _any_ file path — the extension check is only in `collectSupportedFiles` (directory walking). If a user passes `md-go-validator styles.css`, it will try to parse it as markdown and extract code blocks. This is arguably correct (the file might contain code blocks regardless of extension), but it could also produce confusing results. The alternative is to validate the extension first and return an error like `"unsupported file type: .css (supported: .md, .markdown, .mdx)"`. This is a **product decision** I cannot make alone — it depends on whether the tool should be strict or permissive about file types.

---

## Test Coverage Summary

| Package               | Coverage | Status            |
| --------------------- | -------- | ----------------- |
| `pkg`                 | 81.9%    | Good              |
| `pkg/types`           | 83.7%    | Good              |
| `pkg/output`          | 91.5%    | Excellent         |
| `pkg/languages`       | 66.7%    | Needs improvement |
| `cmd/md-go-validator` | 60.8%    | Needs improvement |
| `pkg/code`            | 0.0%     | Critical gap      |
| `pkg/testutil`        | 0.0%     | Critical gap      |

## Git Log (This Session)

```
80a582d feat: add MDX file support (.mdx) with docs, changelog, and CLI tests
33a4109 docs: update CLI help header to mention MDX files
c9cf503 docs: update stale comments to reflect MDX file support
7bede34 refactor: derive verbose file count message from supportedExtensions map
e408ff2 test: add unit tests for isSupportedFile covering all extensions
49295a8 refactor: rename collectMarkdownFiles to collectSupportedFiles
edf8003 refactor: rename isMarkdownFile to isSupportedFile
1624cea refactor: extract supported file extensions into a single source of truth map
```

## Files Changed This Session

- `pkg/validator.go` — `supportedExtensions` map, `isSupportedFile()`, `collectSupportedFiles()`, `formatSupportedExtensions()`, updated comments
- `pkg/validator_test.go` — 3 new test functions (MDX file, MDX directory, isSupportedFile table-driven)
- `pkg/extractor.go` — Updated 3 doc comments
- `pkg/languages/language.go` — Updated 1 doc comment
- `pkg/types/code_block.go` — Updated 1 doc comment
- `pkg/types/result.go` — Updated 1 doc comment
- `cmd/md-go-validator/main.go` — Package comment, CLI help header, SUPPORTED FILE TYPES section
- `cmd/md-go-validator/main_test.go` — 1 new MDX subtest
- `README.md` — Description + supported file types
- `CHANGELOG.md` — MDX entry
- `AGENTS.md` — Overview
