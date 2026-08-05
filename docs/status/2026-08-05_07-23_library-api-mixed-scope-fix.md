# Status Report — 2026-08-05 07:23

_Session focused on fixing a validation error reported in the website docs._

---

## a) FULLY DONE

1. **Fixed mixed-scope code block in `library-api.mdx`** (line 97, block #7)
   - **Root cause:** The snippet mixed package-level declarations (`type MyValidator struct{}` + methods) with function-body statements (`registry := ...`, `if err := ...`). No single Go scope can contain both — this is a true mixed-scope error that no parsing strategy can resolve.
   - **Fix:** Split into two valid code blocks — (1) the interface implementation (package-level, handled by package-wrapper strategy) and (2) the registration code (function-body statements, handled by statements strategy). Connected with a "Then register it:" sentence.
   - **Why not `// skip-validate`:** Using a skip directive would signal "this code is broken" in the project's *own* docs — the validator should be able to validate its own documentation. Splitting produces genuinely valid Go.

2. **Verified the fix:**
   - `library-api.mdx` alone: 8 valid, 0 errors
   - Entire `website/src/content/docs` tree: 9 valid, 4 skipped, 0 errors
   - Full test suite (`go test ./...`): all 10 packages pass

3. **Change auto-committed** as `ea0459d` by the auto-git daemon.

---

## b) PARTIALLY DONE

Nothing — this was a single-file fix, fully completed.

---

## c) NOT STARTED

1. **Did NOT run `golangci-lint run ./...`** — irrelevant since no `.go` files changed, but listed for completeness.
2. **Did NOT run `nix flake check`** — the project's canonical CI gate. Would catch format issues on the `.mdx` file (treefmt formats `.go` and `.nix`, not `.mdx`, so likely a no-op, but unverified).
3. **Did NOT visually verify the rendered MDX page** — the split + "Then register it:" connector reads logically, but I didn't build the Astro site to confirm visual correctness.

---

## d) TOTALLY FUCKED UP

Nothing catastrophic. But one honest miss:

- **I declared "Done" without investigating the 4 skipped blocks.** I saw "Skipped: 4" in the report and moved on without checking *why* they were skipped. I only investigated them just now (for this report). They turned out to be legitimate (they're in `skip-directives.mdx`, intentionally demonstrating skip directives), but I got lucky. A skipped block could have been hiding a broken snippet that someone slapped `// skip-validate` on as a band-aid. **I should always investigate skips, not just errors.**

---

## e) WHAT WE SHOULD IMPROVE

### Process Improvements

1. **Always investigate skipped blocks, not just errors.** A skip is a suppressed error. Every skip in docs should have a documented, legitimate reason. Consider adding a `--fail-on-skipped` pass to the project's own dogfooding CI.

2. **The project should validate its own website docs in CI.** Currently `.github/workflows/` has `ci.yml` (Go) and `website.yml` (Astro build/deploy), but there's no step that runs `md-go-validator` on `website/src/content/docs/`. The validator should eat its own dog food — this error should have been caught by CI, not found manually.

3. **The mixed-scope error message is excellent but the hint could be better.** The current hint says "add `// skip-validate`". For snippets that *can* be split into valid blocks (like this one), a better hint might be: "consider splitting package-level declarations from function-body statements into separate code blocks." The skip directive should be the last resort, not the first suggestion.

4. **No grep-based audit for similar mixed-scope patterns in other docs.** The validator catches *syntax* errors, but there may be snippets across the 14 doc files that are technically valid yet poor quality (e.g., snippets that only pass via the weakest strategy when they could be cleaner).

### Code/Doc Quality Observations

5. **`library-api.mdx` snippets 2-5** (lines 22-63) are all function-body-only snippets that pass via the statements strategy. This is consistent and good. The registration snippet I split out now follows the same pattern.

6. **The docs use `panic(err)`** as error handling in examples (lines 35, 114). This is a docs convention for brevity but could set a bad example for library consumers. Consider a note like "In production, handle errors appropriately."

---

## f) Up to 50 Things We Should Get Done Next

### High Priority — Dogfooding & CI

1. **Add `md-go-validator` run to `website.yml` CI workflow** — validate `website/src/content/docs/` on every push. Fail the build on errors.
2. **Add `--fail-on-skipped` to the dogfooding CI step** — force every skip to be justified or fixed.
3. **Audit all 4 skipped blocks in `skip-directives.mdx`** — confirm each is intentionally demonstrating a skip directive and add inline comments if unclear.
4. **Run `nix flake check`** to confirm format/build/test all pass after the change.

### Medium Priority — Error Message UX

5. **Improve the mixed-scope error hint** — suggest splitting into separate blocks before suggesting `// skip-validate`.
6. **Add a "did you mean to split?" heuristic** — when mixed-scope is detected, check if splitting at the first function-body statement produces two valid blocks, and suggest that.
7. **Review all error messages in `pkg/languages/go.go`** for actionability per the error-handling spec (what / why / fix / escape).

### Medium Priority — Docs Quality

8. **Scan all 14 doc files for snippets that only pass via the weakest strategy** — they might be refactorable into cleaner examples.
9. **Add a contributing note** about keeping code blocks single-scope (either package-level OR function-body, not both).
10. **Review `library-api.mdx` for API accuracy** — verify `ValidateFile`, `ValidateDirectory`, `ValidateDirectoryFunc`, `HasError()` signatures match current code.
11. **Add `pkg.go.dev` link verification** — the link at line 121 may be stale if the module path changed.
12. **Consider replacing `panic(err)` in examples** with a comment about production error handling.

### Lower Priority — General Project Health

13. **Run `golangci-lint run ./...`** and address any issues.
14. **Check coverage hasn't regressed** — run `go test -cover ./...` and compare to AGENTS.md table.
15. **Verify `nix build .#` still works** after any dependency changes.
16. **Review `.github/workflows/ci.yml`** for `GOEXPERIMENT=jsonv2` consistency.
17. **Update `AGENTS.md` coverage table** if numbers have drifted.
18. **Check for stale TODOs** in `TODO_LIST.md` related to docs or website.
19. **Run `goreleaser check`** to ensure release config is valid.
20. **Audit `flake.lock`** for outdated inputs.
21. **Review `pkg/baseline/` coverage (73%)** — lowest coverage package, may need more tests.
22. **Review `cmd/` coverage (73.9%)** — second lowest, CLI flag handling may be undertested.
23. **Add integration test** that validates the project's own docs as a fixture.
24. **Consider a `make docs-validate` or flake app** for one-command docs validation.
25. **Review website DNS status** — AGENTS.md says "pending DNS propagation"; check if resolved.
26. **Verify Firebase deploy target** still works with current config.
27. **Check `website/` npm dependencies** for security advisories.
28. **Review Starlight/Astro version** for major updates.
29. **Add a `CHANGELOG.md` entry** for the docs fix if the project maintains one.
30. **Consider adding `md-go-validator.yaml` config** to the `website/` directory for custom validation settings.

---

## g) Questions I CANNOT Figure Out Myself

1. **Should the project's website CI fail on skipped blocks (`--fail-on-skipped`), or only on errors?** The `skip-directives.mdx` guide legitimately needs skips for demonstration — failing on skips would break its own CI unless those blocks are exempted. Do you want a whitelist mechanism, or should demo skips use a different strategy (e.g., a config file with `skipDirectives` scoped to that page)?

2. **Is the auto-commit message format (`docs(guides): clarify custom validator registration steps...`) acceptable, or do you want to amend it?** The daemon committed before I could review the message wording. The change is small enough that amending is trivial if you prefer different phrasing.

3. **Should I proactively scan and fix mixed-scope snippets across ALL project Markdown files (README, AGENTS.md, etc.), not just the website docs?** The validator currently only runs on `website/src/content/docs/`. There may be similar issues hiding in root-level docs that aren't covered by any validation.
