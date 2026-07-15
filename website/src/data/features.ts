import type { Feature } from "./types";

export const features: Feature[] = [
  {
    icon: "check",
    title: "Multi-Language Validation",
    desc: "Go (built-in parser), TypeScript, TSX, Rust, Nix, HCL/Terraform, and Templ — all embedded, no external tools.",
  },
  {
    icon: "code",
    title: "7-Strategy Go Parser",
    desc: "Progressive parsing handles partial snippets: complete files, package wrappers, function bodies, expressions, statements, and imports.",
  },
  {
    icon: "terminal",
    title: "CI-Friendly Output",
    desc: "Table, JSON, YAML, CSV, Markdown, SARIF, and quiet formats. Structured exit codes (0/1/2) for pipeline integration.",
  },
  {
    icon: "package",
    title: "Pure Go, No CGO",
    desc: "Uses gotreesitter (pure-Go tree-sitter) for multi-language parsing. Compiles to any GOOS/GOARCH with zero C dependencies.",
  },
  {
    icon: "layers",
    title: "Baseline Regression Mode",
    desc: "Generate a baseline of known errors, then fail only on new issues. Perfect for incremental adoption in existing projects.",
  },
  {
    icon: "bolt",
    title: "Skip Directives",
    desc: "Mark intentionally incomplete snippets with <!-- skip-validate --> or //nolint. Custom directives configurable via YAML.",
  },
];
