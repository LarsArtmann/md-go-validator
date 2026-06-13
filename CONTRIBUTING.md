# Contributing to md-go-validator

> **Thank you for contributing!** This guide covers everything you need to know to contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Quick Start](#quick-start)
- [Development Setup](#development-setup)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Architecture](#architecture)
- [Linting & Quality](#linting--quality)
- [Security](#security)
- [Pull Request Process](#pull-request-process)
- [Commit Messages](#commit-messages)

---

## Code of Conduct

We are committed to providing a welcoming and respectful environment. All contributors are expected to:

- **Be respectful** — Treat others with kindness and professionalism
- **Be inclusive** — Welcome diverse perspectives and experiences
- **Be constructive** — Provide feedback that helps improve the project
- **Be collaborative** — Work together to achieve the best outcomes

**Unacceptable behavior** will not be tolerated.

---

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/LarsArtmann/md-go-validator.git
cd md-go-validator

# 2. Enter the Nix development shell
nix develop

# 3. Verify everything works
go test -race ./...
golangci-lint run ./...
nix flake check
```

---

## Development Setup

### Prerequisites

| Tool          | Version | Purpose                  |
| ------------- | ------- | ------------------------ |
| Nix           | latest  | Build system / dev shell |
| Go            | 1.26+   | Language runtime         |
| golangci-lint | v2.6.0  | Code linting             |

### Nix-Based Workflow (Recommended)

All build, test, lint, and format tasks are driven by `flake.nix`.

```bash
# Enter a shell with Go, gopls, golangci-lint, and goreleaser
nix develop

# Build the binary
nix build .#
# or
go build ./cmd/md-go-validator

# Run tests
nix run .#test
# or
go test -race ./...

# Run the linter
nix run .#lint
# or
golangci-lint run ./...

# Run all checks (build + test + format)
nix flake check

# Format code
nix fmt
```

### Manual Setup (Without Nix)

If you do not use Nix, install the prerequisites manually:

```bash
# Install Go 1.26.3+ (see https://go.dev/dl/)
# Install golangci-lint (see https://golangci-lint.run/usage/install/)

go test -race ./...
golangci-lint run ./...
```

---

## Code Standards

### Mandatory Rules

1. **Composition Over Inheritance**
   - Prefer interfaces and struct embedding over class hierarchies
   - Use dependency injection for testability

2. **Strong ID Types**
   - No raw string/numeric IDs — use branded types
   - Example: `type FileID = brsdt.Type[string, "file_id"]`

3. **Early Returns**
   - Use guard clauses to reduce nesting
   - Keep functions small and focused

### File Organization

```
├── cmd/                    # Application entry points
│   └── md-go-validator/main.go
├── pkg/                    # Public packages
│   ├── code/               # Code utilities
│   ├── languages/          # Language validators and registry
│   ├── output/             # Report formatting
│   ├── types/              # Domain types (branded IDs, results, etc.)
│   ├── validator.go        # File/directory validation orchestration
│   ├── extractor.go        # Markdown code-block extraction
│   └── context.go          # Context lifecycle management
├── docs/                   # Project documentation
│   ├── status/             # Status reports
│   └── planning/           # Execution plans
├── flake.nix               # Nix flake (build, test, dev shell)
└── package.nix             # Nix package expression for overlays
```

### Naming Conventions

| Type       | Convention                  | Example                        |
| ---------- | --------------------------- | ------------------------------ |
| Packages   | lowercase, single word      | `domain`, `handlers`           |
| Interfaces | PascalCase with "er" suffix | `Repository`, `Service`        |
| Functions  | PascalCase                  | `CreateUser`, `GetByID`        |
| Variables  | camelCase                   | `userID`, `isActive`           |
| Constants  | PascalCase                  | `MaxRetries`, `DefaultTimeout` |
| Files      | lowercase, descriptive      | `user_repository.go`           |

---

## Testing

### Test Structure

<!-- skip-validate -->

```go
// Package naming: same package or package_test for black-box
package domain_test  // Preferred for integration

// or

package domain  // For white-box testing
```

### Test Organization

<!-- skip-validate -->

```go
// Use descriptive names with Given-When-Then pattern
func TestCreateUser_GivenValidInput_WhenUserDoesNotExist_ShouldCreateUser(t *testing.T)
```

### Coverage Requirements

| Metric          | Minimum | Target |
| --------------- | ------- | ------ |
| Line Coverage   | 70%     | 85%    |
| Branch Coverage | 60%     | 75%    |
| Critical Paths  | 100%    | 100%   |

```bash
# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Detailed coverage analysis
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Architecture

### Package Layout

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI                                  │
│   cmd/md-go-validator/main.go                                │
└─────────────────────────┬───────────────────────────────────┘
                          │ imports
┌─────────────────────────▼───────────────────────────────────┐
│                      pkg/validator                           │
│   (File/directory validation orchestration)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │ imports
┌─────────────────────────▼───────────────────────────────────┐
│              pkg/extractor + pkg/languages                   │
│   (Code-block extraction + language validators)              │
└─────────────────────────┬───────────────────────────────────┘
                          │ imports
┌─────────────────────────▼───────────────────────────────────┐
│                       pkg/types                              │
│   (Branded IDs, Result, CodeBlock, ValidationStatus)         │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Rules

- `cmd/` → `pkg/*`
- `pkg/validator` → `pkg/extractor`, `pkg/languages`, `pkg/types`
- `pkg/languages` → `pkg/types`, `pkg/code`
- `pkg/output` → `pkg/types`

### Current Known Cycle

`pkg/types` currently imports `pkg/languages` (for the `Language` type used in `CodeBlock`), and `pkg/languages` imports `pkg/types` for result types. This is documented in `docs/modularization/PROPOSAL.md` and should be resolved by moving the `Language` type to a dedicated package.

---

## Linting & Quality

### Quick Commands

```bash
# Run the Go linter
golangci-lint run ./...

# Run all Nix checks (format, build, test)
nix flake check

# Format .nix and .go files
nix fmt
```

### Quality Gates

Before merging, all checks must pass:

- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run ./...` passes
- [ ] `nix flake check` passes
- [ ] Code is formatted (`gofmt` / `nixfmt`)
- [ ] Imports are organized (`goimports`)

---

## Security

### Security Best Practices

1. **Input Validation** — Validate all inputs at boundaries
2. **No Secrets in Code** — Use environment variables
3. **Principle of Least Privilege** — Request only required capabilities
4. **Error Messages** — Don't leak sensitive information

---

## Pull Request Process

### PR Requirements

1. **Branch Naming**

   ```
   feat/description
   fix/description
   docs/description
   refactor/description
   test/description
   ```

2. **PR Description Template**

   ```markdown
   ## Summary

   Brief description of changes

   ## Type

   - [ ] Feature
   - [ ] Bug fix
   - [ ] Refactoring
   - [ ] Documentation

   ## Test Plan

   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing performed

   ## Checklist

   - [ ] Code follows style guidelines
   - [ ] Tests pass
   - [ ] `nix flake check` passes
   - [ ] Documentation updated
   ```

3. **Self-review first** — Run `go test -race ./...`, `golangci-lint run ./...`, and `nix flake check` locally
4. **Small PRs** — Keep changes focused and digestible
5. **Explain "why"** — Not just "what", but rationale
6. **Be responsive** — Address feedback promptly

---

## Commit Messages

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

| Type     | Description              |
| -------- | ------------------------ |
| feat     | New feature              |
| fix      | Bug fix                  |
| docs     | Documentation changes    |
| style    | Formatting, whitespace   |
| refactor | Code restructuring       |
| test     | Adding/updating tests    |
| chore    | Build, tooling, CI       |
| perf     | Performance improvements |
| ci       | CI/CD changes            |
| revert   | Reverting changes        |

### Examples

```bash
# Good
feat(cli): add --version flag

Adds --version and -V flags so users can verify the installed
binary version. Version is injected at build time via ldflags.

# Bad
fix stuff

# Good
docs(readme): update installation instructions

Added Nix-based development instructions and removed stale
`just` references.

# Bad
updated README
```

---

## Getting Help

### Resources

- [Go Documentation](https://go.dev/doc/)
- [Uber Go Style Guide](https://github.com/uber-go/guide)
- [Effective Go](https://go.dev/doc/effective_go)
- [Project docs](./docs)

### Getting Unblocked

1. **Read the docs** — Check `docs/` folder first
2. **Check existing issues** — Someone may have solved it
3. **Ask questions** — Open a discussion, don't struggle alone

---

_Thank you for contributing to md-go-validator!_
