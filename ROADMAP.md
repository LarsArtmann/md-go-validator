# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to TODO_LIST.md.

## Recently Shipped

| Item                  | Details                                                                                                  |
| --------------------- | -------------------------------------------------------------------------------------------------------- |
| Documentation website | [md-go-validator.lars.software](https://md-go-validator.lars.software) — Astro + Starlight + Tailwind v4 |
| CI/CD auto-deploy     | `.github/workflows/website.yml` — Firebase Hosting on push to master                                     |

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
- npm wrapper for JS-heavy teams
- Docker image publication for CI pipelines without Go/Nix

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Runtime validation of executed code:** This is a syntax validator, not a test runner. Running code blocks is out of scope.
- **Linting/style enforcement:** Syntax validity is the contract. Opinionated style rules belong in linters (golangci-lint, eslint, etc.).
- **Markdown linting:** We validate embedded code blocks, not markdown structure. Tools like markdownlint own that space.
