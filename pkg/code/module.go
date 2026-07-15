package code

import (
	"slices"
	"strings"
)

// IsPseudoModuleFile returns true if the code block looks like go.mod content
// rather than Go source code. Documentation often includes module directives
// (require, replace, module, go) inside Go fenced code blocks.
func IsPseudoModuleFile(code string) bool {
	lines := strings.Split(code, "\n")

	moduleDirectives := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		if isModuleDirective(trimmed) {
			moduleDirectives++
		}
	}

	// If more than half the non-empty lines are module directives, it's a
	// pseudo go.mod file.
	nonEmpty := 0

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}

	if nonEmpty == 0 {
		return false
	}

	return moduleDirectives*2 >= nonEmpty
}

// isModuleDirective checks if a line is a go.mod directive.
func isModuleDirective(line string) bool {
	directives := []string{
		"module ",
		"require ",
		"require(",
		"replace ",
		"exclude ",
		"retract ",
	}

	if slices.ContainsFunc(directives, func(d string) bool {
		return strings.HasPrefix(line, d)
	}) {
		return true
	}

	// "go 1.21" version directive.
	if rest, ok := strings.CutPrefix(line, "go "); ok {
		if rest != "" && (rest[0] >= '0' && rest[0] <= '9') {
			return true
		}
	}

	// Version strings inside require blocks: "github.com/owner/repo v1.2.3"
	if looksLikeModuleVersion(line) {
		return true
	}

	// Replace target: "github.com/owner/repo => ../local"
	if strings.Contains(line, " => ") {
		return true
	}

	return false
}

// looksLikeModuleVersion checks if a line looks like a module version line
// inside a require block (e.g. "github.com/acme/core v1.6.0").
func looksLikeModuleVersion(line string) bool {
	// Must contain a space-separated version string.
	idx := strings.IndexByte(line, ' ')
	if idx <= 0 {
		return false
	}

	path := line[:idx]
	version := strings.TrimSpace(line[idx:])

	// Path must look like a module path (contains / or starts with domain).
	if !strings.Contains(path, "/") && !strings.Contains(path, ".") {
		return false
	}

	// Version must start with 'v' followed by a digit.
	const minVersionLen = 2
	if len(version) < minVersionLen {
		return false
	}

	return version[0] == 'v' && version[1] >= '0' && version[1] <= '9'
}
