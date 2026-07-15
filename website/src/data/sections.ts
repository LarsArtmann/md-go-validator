import type { ComparisonItem, UseCase } from "./types";

export const comparisons: ComparisonItem[] = [
  {
    variant: "markdownlint",
    pros: ["Markdown structure linting", "Wide ecosystem", "Configurable rules"],
    cons: ["No code block syntax validation", "No multi-language parsing", "No SARIF output"],
    accent: false,
  },
  {
    variant: "prettier",
    pros: ["Code formatting", "Markdown formatting", "Wide language support"],
    cons: ["No syntax validation", "Modifies files (not read-only)", "Requires Node.js"],
    accent: false,
  },
  {
    variant: "md-go-validator",
    pros: [
      "Validates code block syntax",
      "6 Go parsing strategies",
      "7 languages via tree-sitter",
      "Pure Go (no CGO)",
      "SARIF + JSON output",
      "Baseline regression mode",
    ],
    cons: [],
    accent: true,
  },
];

export const useCases: UseCase[] = [
  {
    title: "CI Pipelines",
    desc: "Block PRs that break documentation code examples",
    icon: "git-branch",
  },
  {
    title: "Pre-commit Hooks",
    desc: "Catch errors locally before they reach CI",
    icon: "terminal",
  },
  {
    title: "GitHub Actions",
    desc: "Built-in action with zero install needed",
    icon: "globe",
  },
  {
    title: "Library Integration",
    desc: "Embed validation in your own Go tools",
    icon: "zap",
  },
  {
    title: "SARIF Reports",
    desc: "Upload results to GitHub Code Scanning",
    icon: "shield",
  },
];
