# md-go-validator - Project Instructions

## Overview

A Go tool that validates Go code blocks embedded in Markdown documentation files.
Uses multiple parsing strategies to handle partial code snippets commonly found in technical documentation.

## Tech Stack

- Go 1.26+
- Pure stdlib (no external dependencies)
- Library code in `pkg/`
- CLI entry point in `cmd/md-go-validator/`

## Build Commands

```bash
# Build
go build ./cmd/md-go-validator

# Test
go test ./...

# Test with coverage
go test -cover ./...

# Run
go run ./cmd/md-go-validator [options] [path...]

# Validate project
buildflow --semantic --fix
go-structure-linter . --fix
```

## Architecture

### pkg/extractor.go

- `ExtractGoCodeBlocks(content string) []CodeBlock` - Extracts Go code blocks from markdown
- `SkipDirectives` - List of directives that skip validation

### pkg/parser.go

- `ValidateGoCode(code string) error` - Multi-strategy Go code validation

### pkg/validator.go

- `Validator` type - Main validator with file/directory validation
- `PrintReport()` - Output formatting
- `HasErrors()` - Error checking

### cmd/md-go-validator/main.go

- CLI entry point
- Argument parsing
- Output orchestration

## Key Patterns

### Multi-Strategy Parsing

The validator tries 5 approaches:

1. Complete file parsing
2. Package wrapper (`package main`)
3. Function wrapper (`func main()`)
4. Expression (`_ = <code>`)
5. Statements (function body)

### Skip Directives

Users can skip validation with:

- `<!-- skip-validate -->`
- `<!-- skip-md-validate -->`
- `<!-- md-skip -->`
- `<!-- no-validate -->`
- `// skip-validate`
- `//nolint`

## Code Style

- Functional programming: immutability, pure functions
- Early returns over nested conditionals
- Small, focused functions (max 10 complexity)
- All tests must use `t.Parallel()`
- All struct fields must be explicit in tests (exhaustruct)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`

## Release

```bash
goreleaser release
```
