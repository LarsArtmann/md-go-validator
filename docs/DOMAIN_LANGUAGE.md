# Domain Language

A **Unified Language** for md-go-validator — shared across maintainers, contributors, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.

## Glossary

| Term            | Definition                                                           | Context                              |
| --------------- | -------------------------------------------------------------------- | ------------------------------------ |
| Code Block      | A fenced code region in Markdown/MDX (e.g., ` ```go ... ``` `)       | Extraction, validation               |
| Code Block Info | The language tag on a fenced code block (e.g., `go`, `typescript`)   | Extraction, language identification  |
| Skip Directive  | A comment that marks a code block as intentionally invalid           | Extraction, validation               |
| Strategy        | One of the parsing approaches tried by the Go validator              | Go validation                        |
| Best-attempt    | The error from the strategy that parsed furthest before failing      | Error reporting                      |
| Baseline        | A file of known error signatures; only new errors fail the build     | Baseline regression mode             |
| Signature       | `file:line:errorcode` tuple identifying a specific known error       | Baseline                             |
| Finding         | A neutral validation result convertible to SARIF/LSP/JSON            | Finding interchange (`pkg/finding/`) |
| Exclude Pattern | A glob pattern that filters files/dirs from validation               | File processing                      |
| Branded Type    | A named type wrapping a primitive to prevent mixing (e.g., `FileID`) | Type safety (`pkg/types/`)           |

## Entities

Objects with identity and lifecycle.

| Term                | Definition                                            | Context            |
| ------------------- | ----------------------------------------------------- | ------------------ |
| FileValidator       | The main validator orchestrating file/directory scans | `pkg/validator.go` |
| Registry            | Maps languages to their validators                    | `pkg/languages/`   |
| GoValidator         | stdlib-based Go parser with multi-strategy approach   | `pkg/languages/`   |
| TreeSitterValidator | tree-sitter-based validator for non-Go languages      | `pkg/languages/`   |

## Value Objects

Immutable objects defined by attributes.

| Term             | Definition                                          | Context          |
| ---------------- | --------------------------------------------------- | ---------------- |
| CodeBlock        | Extracted code block with language, content, line   | `pkg/types/`     |
| Result           | Validation outcome for a single code block          | `pkg/types/`     |
| ValidationStatus | Enum: unknown/valid/skipped/error                   | `pkg/types/`     |
| ErrorCode        | Enum: syntax/not_available/not_registered/unknown   | `pkg/types/`     |
| FileID           | Branded string identifying a source file            | `pkg/types/`     |
| LineNumber       | Branded uint for 1-based line numbers               | `pkg/types/`     |
| BlockIndex       | Branded uint for code block position in a file      | `pkg/types/`     |
| ExcludePattern   | Branded string with encapsulated glob matching      | `pkg/types/`     |
| Language         | Branded string for a supported programming language | `pkg/languages/` |

## Bounded Contexts

| Context    | Description                                                |
| ---------- | ---------------------------------------------------------- |
| Extraction | Parsing Markdown/MDX to find and classify code blocks      |
| Validation | Running language-specific parsers against extracted code   |
| Output     | Formatting and emitting results in various formats         |
| Config     | Loading and merging `.md-go-validator.yaml` with CLI flags |
| Baseline   | Tracking known errors to support incremental adoption      |
| Finding    | Converting results to neutral Finding type for interchange |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
