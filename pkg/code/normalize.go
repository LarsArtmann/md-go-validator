package code

import (
	"regexp"
	"strings"
)

// rxBodyOmitted matches brace bodies that are just an ellipsis placeholder,
// e.g. `{ ... }`, `{...}`, `{ ... }` — the documentation idiom for "body elided".
var rxBodyOmitted = regexp.MustCompile(`\{\s*\.\.\.\s*\}`)

// rxEllipsisLine matches a line that contains only an ellipsis (optionally
// surrounded by whitespace) including the trailing newline. These lines
// represent "code omitted" in docs.
var rxEllipsisLine = regexp.MustCompile(`(?m)^[ \t]*\.\.\.[ \t]*\r?\n`)

// NormalizeDocIdioms transforms common documentation elision idioms into
// syntactically valid Go so the existing parsing strategies can handle them.
//
// Recognised idioms:
//
//   - `{ ... }` or `{...}` → `{}`  (body omitted)
//   - A line containing only `...` → dropped entirely
func NormalizeDocIdioms(code string) string {
	// 1. "body omitted" — { ... } → {}
	normalised := rxBodyOmitted.ReplaceAllString(code, "{}")

	// 2. Lines that are only "..." → remove the entire line
	normalised = rxEllipsisLine.ReplaceAllString(normalised, "")

	// 3. Collapse multiple consecutive blank lines into one, trim trailing blanks
	normalised = collapseBlankLines(normalised)

	return normalised
}

// collapseBlankLines reduces runs of 2+ blank lines to a single blank line
// and trims trailing blank lines.
func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")

	var kept []string

	prevBlank := false

	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""

		if isBlank && prevBlank {
			continue
		}

		kept = append(kept, line)

		prevBlank = isBlank
	}

	// Trim trailing blank lines.
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}

	return strings.Join(kept, "\n")
}
