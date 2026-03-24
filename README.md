# md-go-validator

A CLI tool and Go library that validates Go code blocks in Markdown files.

Ensures code examples in your documentation are syntactically valid Go code. Uses multiple parsing strategies to handle partial code snippets (imports, type declarations, function bodies) commonly found in technical documentation.

- **CLI tool** - Validate entire documentation repositories
- **Go library** - Integrate validation into your own tools
- **CI-friendly** - Exit code 1 on validation errors
- **Pure Go** - No external dependencies

## Features

- **Multiple Parsing Strategies** - Handles partial code:
  - Complete files
  - Package declarations only
  - Function bodies
  - Expressions
  - Statements
- **Skip Directives** - Mark intentionally incomplete code
- **Recursive Scanning** - Validates entire documentation trees
- **CI-Friendly** - Exit code 1 on errors, structured output
- **Fast** - Pure Go, no external dependencies

## Installation

```bash
go install github.com/larsartmann/md-go-validator@latest
```

## Usage

```bash
# Validate all markdown in current directory
md-go-validator .

# Validate specific file
md-go-validator README.md

# Validate multiple paths
md-go-validator docs/ README.md

# Verbose output (show each block)
md-go-validator -v .

# Quiet mode (summary only)
md-go-validator -q .
```

## Options

| Option          | Description                              |
| --------------- | ---------------------------------------- |
| `-v, --verbose` | Show progress for each code block        |
| `-q, --quiet`   | Only show summary (no code in errors)    |
| `--no-code`     | Don't show code snippets in error output |
| `-h, --help`    | Show help message                        |

## Skip Directives

For intentionally incomplete code snippets (common in documentation), place these **before** the code block or **inside** it:

````markdown
<!-- skip-validate -->

`+ "```go" +`
// This is intentionally partial
type MyStruct struct {
Name string
}
`+ "```" +`
````

### Available Directives

- `<!-- skip-validate -->`
- `<!-- skip-md-validate -->`
- `<!-- md-skip -->`
- `<!-- no-validate -->`
- `// skip-validate`
- `//nolint`

## How It Works

### Parsing Strategies

The validator tries multiple approaches to parse code:

1. **Complete File** - Parse as-is
2. **Package Wrapper** - Wrap in `package main`
3. **Function Wrapper** - Wrap in `func main() { ... }`
4. **Expression** - Wrap as `_ = <code>`
5. **Statements** - Wrap as function body

This handles:

- Complete programs
- Type declarations
- Function signatures
- Import statements
- Variable declarations
- Expressions

### Example Validations

**Valid (passes):**

```go
package main

func main() {
    fmt.Println("Hello")
}
```

```go
type User struct {
    Name string
    Age  int
}
```

```go
import "github.com/example/pkg"
```

```go
result, err := doSomething()
if err != nil {
    return err
}
```

**Invalid (fails):**

<!-- skip-validate -->

```go
require (
    github.com/pkg v1.0.0
)
```

(This is go.mod syntax, not Go code)

## Sample Output

```
============================================================
📊 VALIDATION REPORT
============================================================
Total code blocks: 863
✅ Valid: 738
⏭️  Skipped: 0
❌ Invalid: 125

------------------------------------------------------------
❌ ERRORS FOUND:
------------------------------------------------------------

📍 docs/example.md:42 (block #3)
   Error: snippet.go:2:1: expected 'package', found 'import'

   Code:
   --------------------------------------------------
     1 | import "fmt"
     2 |
     3 | func main() {}
   --------------------------------------------------

============================================================
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Validate Docs

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Validate Go code in markdown
        run: go run github.com/larsartmann/md-go-validator@latest --no-code .
```

### Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: md-go-validator
        name: Validate Go in Markdown
        entry: md-go-validator
        language: system
        types: [markdown]
```

## Comparison with Alternatives

| Tool                | Partial Code           | Skip Directives | CI-Friendly   | Dependencies |
| ------------------- | ---------------------- | --------------- | ------------- | ------------ |
| **md-go-validator** | ✅ Multiple strategies | ✅ 6 directives | ✅ Exit codes | ❌ None      |
| godown              | ❌ Requires package    | ❌ No           | ✅ Yes        | ❌ None      |
| Runme               | ⚠️ Limited             | ⚠️ Annotations  | ✅ Yes        | ❌ None      |
| embedmd             | ❌ No execution        | ❌ No           | ✅ Yes        | ❌ None      |

## Library Usage

```go
package main

import (
    "fmt"

    mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
)

func main() {
    v := mdgovalidator.New(false)

    // Validate a single file
    results, err := v.ValidateFile("README.md")
    if err != nil {
        panic(err)
    }

    // Check for errors
    if mdgovalidator.HasErrors(results) {
        mdgovalidator.PrintReport(results, true)
    }
}
```

## License

MIT
