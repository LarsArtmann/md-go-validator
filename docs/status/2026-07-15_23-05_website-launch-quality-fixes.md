# Status Report: Website Launch Quality Fixes

**Date:** 2026-07-15 23:05  
**Session:** Post-launch quality audit and fixes  
**Trigger:** Self-assessment request after website-launch skill completion

---

## Context

The website-launch skill completed all 7 phases and the user committed the work as `69fcb10` ("feat: launch public documentation website at md-go-validator.lars.software"). This session was a brutal self-review triggered by the question "What did you forget? What could you have done better?" A comprehensive source-code audit found **16 issues** — 9 factual errors in the website content and 7 design/security improvements. This report covers all fixes applied plus the remaining open items.

---

## a) FULLY DONE

### Quality fixes applied this session (11 files, uncommitted)

| Fix                                        | File(s)                                                                                                    | Detail                                                                                                                                                               |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Go strategy count corrected (7→6)          | `README.md`, `features.ts`, `sections.ts`, `HowItWorksSection.astro`, `go-strategies.mdx`, `languages.mdx` | Source code has `strategyCount = 6` in `go_validator.go:18`. The number "7" propagated everywhere from the initial website-launch session. Fixed in all 8 locations. |
| `ValidateDirectoryFunc` callback signature | `library-api.mdx`                                                                                          | Docs showed `func(r types.Result)`. Actual signature returns `error`: `func(r types.Result) error`. Now documented correctly with abort semantics.                   |
| `ErrorCode` package corrected              | `library-api.mdx`                                                                                          | Docs said `types.ErrorCode`. Actual type is `languages.ErrorCode` (`pkg/languages/validator.go:16`).                                                                 |
| `Registry.Register` error handling         | `library-api.mdx`                                                                                          | Docs ignored return value. Actual signature returns `error`. Now shows proper error check.                                                                           |
| Missing imports in code examples           | `library-api.mdx`                                                                                          | Added `context`, `fmt`, `types` imports needed for examples to compile.                                                                                              |
| Missing builder methods                    | `library-api.mdx`                                                                                          | Added `WithMaxBlocks(n)` and `WithRegistry(r)` to the options table. Also corrected `WithExcludePatterns` parameter type to `[]ExcludePattern`.                      |
| OG description includes TSX                | `config.ts`                                                                                                | ogDescription omitted TSX (a supported language). Now lists all 7 languages.                                                                                         |
| CI workflow: `npm install` → `npm ci`      | `website.yml`                                                                                              | `npm ci` is the CI standard for reproducible builds from lockfile.                                                                                                   |
| `X-XSS-Protection` header fixed            | `firebase.json`                                                                                            | Changed from deprecated `"1; mode=block"` to `"0"`. With CSP in place, this header is unnecessary and can introduce vulnerabilities.                                 |
| CTASection icon corrected                  | `CTASection.astro`                                                                                         | "Quick Start" link used `arrow-external` icon but points to internal page. Changed to `arrow-right`.                                                                 |
| Empty `api/` directory removed             | `website/src/content/docs/api/`                                                                            | Empty directory with no content; sidebar correctly links to pkg.go.dev instead.                                                                                      |

### Previously completed (in commit `69fcb10`)

- Full website (Astro 7 + Starlight + Tailwind v4, 16 pages)
- README rewrite (badges, comparison, install/usage, library API, CI)
- LICENSE fixed (proprietary template → MIT)
- `.goreleaser.yml` homepage URLs updated to `md-go-validator.lars.software`
- GitHub metadata (description, topics, homepage URL)
- CI/CD workflow (`.github/workflows/website.yml`)
- Firebase Hosting deployed (68 files, live at `md-go-validator.web.app`)
- Firebase custom domain configured via REST API
- DNS records staged in domains repo (CNAME + ACME TXT, committed)
- `FIREBASE_SERVICE_ACCOUNT` GitHub secret set
- CHANGELOG, FEATURES, ROADMAP, CONTRIBUTING, CONSUMER_PERSPECTIVE all updated
- AGENTS.md gotreesitter version fixed (v0.21.0 → v0.37.0)
- OG meta tags fixed (removed orphaned `/og/home.png` references)

---

## b) PARTIALLY DONE

| Item                         | What's done                                                                                           | What remains                                                                      |
| ---------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| **DNS propagation**          | Records staged in `domains/lars.software.tf` (committed). Firebase custom domain API called (200 OK). | Terraform apply NOT run — Namecheap API key is a placeholder. DNS not live.       |
| **Website deployment**       | Initial deploy done (`69fcb10` version). Quality fixes built and verified locally.                    | Quality fixes not yet deployed to Firebase. Need redeploy after committing.       |
| **CI workflow verification** | Workflow file created, Firebase secret set.                                                           | Never triggered — no push to `master` since creation. Untested in GitHub Actions. |

---

## c) NOT STARTED

| Item                                    | Why                                                                                                                                                      |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Commit quality fixes**                | 11 files uncommitted. User hasn't said "commit".                                                                                                         |
| **Redeploy to Firebase**                | Waiting on commit. Need `firebase deploy` to push fixed content.                                                                                         |
| **OG image generation**                 | astro-og-canvas removed due to network restrictions (fontsource CDN blocked). No OG images. Would need network-accessible font source or a static image. |
| **Content-Security-Policy HTTP header** | Currently CSP is only via `<meta>` tag (patched by `fix-csp.mjs`). HTTP header CSP in `firebase.json` would be more robust. Not started.                 |
| **GitHub API rate-limit mitigation**    | `HeroSection.astro` fetches GitHub stars at build time (unauthenticated = 60 req/hr). Could exhaust in CI with multiple builds. Not mitigated.           |

---

## d) TOTALLY FUCKED UP

| Item                                               | Severity | Detail                                                                                                                                                                                                                                                                                      |
| -------------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Strategy count was wrong everywhere**            | HIGH     | The website, README, and docs ALL said "7-Strategy" when the code constant is `strategyCount = 6`. This was a factual lie told to every reader. The initial website-launch session wrote "7" and nobody verified against source. Fixed this session.                                        |
| **Library API docs had 4 compile-breaking errors** | HIGH     | `ValidateDirectoryFunc` callback signature was wrong (missing `error` return), `ErrorCode` was attributed to the wrong package, `Registry.Register` error return was ignored, and imports were incomplete. A user copy-pasting these examples would get compile errors. Fixed this session. |
| **ogDescription omitted a supported language**     | MEDIUM   | TSX is a supported language but wasn't listed in the OG description. Misleading for social media previews. Fixed this session.                                                                                                                                                              |
| **`npm install` in CI**                            | MEDIUM   | `npm install` can resolve different versions than the lockfile, making CI builds non-reproducible. Should always be `npm ci`. Fixed this session.                                                                                                                                           |
| **Deprecated security header**                     | LOW      | `X-XSS-Protection: 1; mode=block` is deprecated and can introduce vulnerabilities in modern browsers. Fixed this session.                                                                                                                                                                   |

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Always verify claims against source code.** The "7-strategy" error and the 4 library API errors were 100% preventable by reading the actual Go source before writing docs. This is the #1 lesson.

2. **Source-of-truth for marketing claims.** Feature counts, strategy counts, supported languages — these should be derived from the code, not invented. Consider a generated reference file.

3. **Security headers need periodic review.** The `X-XSS-Protection` issue is a known industry deprecation. Firebase security configs should be audited against current OWASP recommendations.

4. **CI should use `npm ci` always.** This is table stakes. The initial workflow used `npm install` — a basic CI mistake.

5. **Self-review should happen BEFORE commit, not after.** The user committed `69fcb10` with all 16 issues. A 10-minute source verification pass before declaring "done" would have caught 9 of them.

### Technical improvements

6. **Add Content-Security-Policy as an HTTP header** in `firebase.json`, not just via `<meta>` tag. HTTP header CSP is more robust and respected by all user agents.

7. **Add rate-limit handling for GitHub API fetch** in `HeroSection.astro`. Either use a build-time env var, cache the result, or gracefully degrade.

8. **Consider re-adding OG image generation** with a local font file or a network-accessible font source. Social media previews currently have no preview image.

9. **The Docker Hub reference in ci-integration.mdx** (`LarsArtmann/md-go-validator:latest`) may not have a published image. Either publish one or remove the reference.

---

## f) Up to 50 things we should get done next

### Immediate (this session's uncommitted work)

1. Commit the 11 quality fix files
2. Redeploy website to Firebase Hosting with fixed content
3. Verify the deployed site shows correct strategy count and API docs

### DNS and go-live

4. Apply DNS via terraform (requires real Namecheap API key)
5. Verify `md-go-validator.lars.software` resolves
6. Verify SSL certificate is provisioned by Firebase
7. Verify custom domain serves the correct Firebase target

### CI/CD

8. Push to master to trigger `website.yml` workflow
9. Verify build-website job passes (astro check, build, html-validate)
10. Verify deploy-website job passes (Firebase deploy)
11. Add CSP HTTP header to `firebase.json` (defense in depth)
12. Consider adding a CI step to verify no broken internal links

### Content quality

13. Verify ALL code examples in docs compile (write a CI test that extracts and validates them — dogfooding!)
14. Verify Docker Hub image reference or remove it from ci-integration.mdx
15. Add OG image (static SVG or re-enable astro-og-canvas with local font)
16. Review every doc page for factual accuracy against current source code
17. Add a "Stability" section to library-api.mdx documenting which APIs are stable vs experimental
18. Write tests for the library API examples to ensure they stay in sync with the code

### Website polish

19. Add rate-limit handling/fallback for GitHub stars fetch in HeroSection
20. Add a "Last updated" or version indicator to the docs
21. Consider adding analytics (privacy-respecting: Plausible, Umami)
22. Add structured data (JSON-LD) to doc pages, not just the landing page
23. Add an Algolia/Pagefind search experience customization
24. Add a "Edit on GitHub" link to each doc page
25. Consider dark mode screenshots for social sharing
26. Add open graph images for each doc page
27. Add a sitemap submission to Google Search Console
28. Monitor Core Web Vitals after custom domain goes live

### README and project presence

29. Add a "Used by" section if any projects adopt the tool
30. Add all-contributors bot for contributor recognition
31. Create a GitHub Discussion template for Q&A
32. Add issue templates (bug report, feature request)
33. Add a SECURITY.md for vulnerability reporting
34. Consider a GitHub Sponsors button

### Build and distribution

35. Publish Homebrew tap (goreleaser has `skip_upload: true`)
36. Publish Docker image to GitHub Container Registry
37. Add AUR package
38. Consider npm wrapper for JS-heavy teams
39. Generate shell completions (bash/zsh/fish)
40. Add `--watch` mode for development feedback

### Code quality (pre-existing TODO items)

41. Resolve `postPatch` replace directive decision (blocked on user design call)
42. Add drift guard: go.mod vs flake input versions
43. Add `--dry-run` flag
44. Add progress indicator for large directories
45. Document API stability (stable vs experimental packages)
46. Add finding round-trip integration test
47. Run `nix flake check --all-systems`
48. Add Python language support (roadmap)
49. Add Java language support (roadmap)
50. Add shell/bash validation (roadmap)

---

## g) Top 2 questions I cannot answer myself

### 1. Should the quality fixes be a separate commit or squashed into `69fcb10`?

The prior commit `69fcb10` contains the website launch WITH the 16 bugs. My 11 files are the fixes. Options:

- **Amend** `69fcb10` to include the fixes (clean history, but rewrites published commit)
- **New commit** on top (honest history: "fix: correct strategy count and API docs after self-review")
- **Squash** both if nothing has been pushed yet

I need to know: **Has `69fcb10` been pushed to the remote?** If yes, a new fix commit is the only safe option. If no, amending is cleaner.

### 2. Is there a real Namecheap API key available to apply the DNS?

The terraform staging is done. The Firebase custom domain API call succeeded. But the actual DNS records can't go live without `terraform apply`, which needs a real API key in `terraform.tfvars`. This is the only thing standing between the project and a live custom domain at `md-go-validator.lars.software`. **Is there a way to get the API key, or should the DNS be applied manually through the Namecheap web UI?**

---

## Build verification (this session)

```
astro check:    0 errors, 0 warnings, 0 hints
astro build:    16 pages, CSP patched 16/16
html-validate:  exit 0 (clean)
grep audit:     0 remaining "7-strategy" references
```

---

## File inventory (this session's uncommitted changes)

```
 .github/workflows/website.yml                     |  2 +-
 README.md                                         |  4 +--
 website/firebase.json                             |  2 +-
 website/src/components/CTASection.astro           |  2 +-
 website/src/components/HowItWorksSection.astro    |  2 +-
 website/src/content/docs/guides/go-strategies.mdx |  4 +--
 website/src/content/docs/guides/languages.mdx     |  4 +--
 website/src/content/docs/guides/library-api.mdx   | 38 ++++++++++++++---------
 website/src/data/config.ts                        |  2 +-
 website/src/data/features.ts                      |  2 +-
 website/src/data/sections.ts                      |  2 +-
 11 files changed, 37 insertions(+), 27 deletions(-)
```
