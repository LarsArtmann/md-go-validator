package languages

import (
	"context"
	"errors"
	"go/scanner"
	"go/token"
	"slices"
	"strings"

	codeutil "github.com/larsartmann/md-go-validator/pkg/code"
)

// skipDirectiveHint is appended to validation errors to guide users toward
// the escape hatch when a snippet is intentionally illustrative.
const skipDirectiveHint = " — if this snippet is illustrative, add // skip-validate as the first line inside the code block"

const strategyCount = 6

// GoValidator validates Go code using the standard library parser.
type GoValidator struct{}

// Language returns the language this validator handles.
func (v *GoValidator) Language() Language {
	return LangGo
}

// IsAvailable always returns true for Go (stdlib is always available).
func (v *GoValidator) IsAvailable() bool {
	return true
}

// strategyAttempt records a parse attempt and its error.
type strategyAttempt struct {
	err error
}

// Validate validates Go code using multiple parsing strategies.
func (v *GoValidator) Validate(_ context.Context, code string) error {
	normalised := codeutil.NormalizeDocIdioms(code)

	if codeutil.IsPseudoModuleFile(normalised) {
		return nil
	}

	fset := token.NewFileSet()
	attempts := v.runStrategies(fset, normalised)

	if len(attempts) == 0 {
		return nil
	}

	return v.buildBestError(attempts, normalised)
}

// runStrategies tries all parsing strategies and returns error attempts
// for those that fail. Returns nil if any strategy succeeds.
func (v *GoValidator) runStrategies(fset *token.FileSet, code string) []strategyAttempt {
	attempts := make([]strategyAttempt, 0, strategyCount)

	indented := codeutil.IndentCode(code)

	strategies := []struct {
		wrapped string
	}{
		{code},
		{"package main\n\n" + code},
		{"package main\n\nfunc main() {\n" + indented + "\n}"},
		{"package main\n\nfunc _() {\n_ = " + code + "\n}"},
		{"package main\n\nfunc _() {\n" + indented + "\n}"},
	}

	for _, s := range strategies {
		err := codeutil.ParseGo(fset, s.wrapped)
		if err == nil {
			return nil
		}

		attempts = append(attempts, strategyAttempt{err: err})
	}

	// Strategy 6: Imports + statements — the dominant documentation pattern.
	importBlock, rest, ok := splitImportsAndStatements(code)
	if ok {
		wrapped6 := "package main\n\n" + importBlock + "\nfunc main() {\n" +
			codeutil.IndentCode(rest) + "\n}"

		err := codeutil.ParseGo(fset, wrapped6)
		if err == nil {
			return nil
		}

		attempts = append(attempts, strategyAttempt{err: err})
	}

	return attempts
}

// buildBestError selects the error from the strategy that parsed furthest
// (highest error line number), then enhances it with mixed-scope detection
// and a skip-directive hint.
func (v *GoValidator) buildBestError(attempts []strategyAttempt, code string) error {
	best := selectBestAttempt(attempts)

	message := extractErrorMessage(best)

	if hint := detectMixedScopeHint(code); hint != "" {
		message += " — " + hint
	}

	message += skipDirectiveHint

	line, column := extractErrorPosition(best)

	return (&ValidationError{
		Message: message,
		Line:    line,
		Column:  column,
		Code:    ErrCodeUnknown,
	}).WithCode(ErrCodeSyntax)
}

// selectBestAttempt returns the attempt whose error references the highest
// source line number — i.e. the strategy that parsed furthest before failing.
func selectBestAttempt(attempts []strategyAttempt) strategyAttempt {
	best := attempts[0]
	bestLine := errorLine(best.err)

	for _, attempt := range attempts[1:] {
		attemptLine := errorLine(attempt.err)
		if attemptLine > bestLine {
			best = attempt
			bestLine = attemptLine
		}
	}

	return best
}

// errorLine extracts the highest line number from a parse error.
func errorLine(err error) int {
	if errList, ok := errors.AsType[scanner.ErrorList](err); ok {
		maxLine := 0

		for _, e := range errList {
			if e.Pos.IsValid() && e.Pos.Line > maxLine {
				maxLine = e.Pos.Line
			}
		}

		return maxLine
	}

	return 0
}

// extractErrorMessage pulls the human-readable message from a parse error.
func extractErrorMessage(attempt strategyAttempt) string {
	var errList scanner.ErrorList

	if errors.As(attempt.err, &errList) && len(errList) > 0 {
		return "Go syntax error: " + errList[0].Msg
	}

	return "Go syntax error: " + attempt.err.Error()
}

// extractErrorPosition pulls the line and column from a parse error.
func extractErrorPosition(attempt strategyAttempt) (int, int) {
	var errList scanner.ErrorList

	if errors.As(attempt.err, &errList) && len(errList) > 0 {
		firstErr := errList[0]

		if firstErr.Pos.IsValid() {
			return firstErr.Pos.Line, firstErr.Pos.Column
		}
	}

	return 0, 0
}

// detectMixedScopeHint checks whether a code snippet mixes package-level
// declarations (import/type/func/var/const) with bare function-body
// statements, which is the #1 documentation pattern that no single strategy
// can handle.
func detectMixedScopeHint(code string) string {
	hasPackageLevel := false
	hasStatements := false

	for line := range strings.SplitSeq(code, "\n") {
		trimmed := strings.TrimSpace(line)

		if isPackageLevelDecl(trimmed) {
			hasPackageLevel = true
		}

		if isLikelyStatement(trimmed) {
			hasStatements = true
		}
	}

	if hasPackageLevel && hasStatements {
		return "snippet mixes package-level declarations with function-body statements"
	}

	return ""
}

// isPackageLevelDecl returns true for lines that start package-level declarations.
func isPackageLevelDecl(line string) bool {
	prefixes := []string{"import ", "type ", "var ", "const "}

	return slices.ContainsFunc(prefixes, func(prefix string) bool {
		return strings.HasPrefix(line, prefix)
	})
}

// isLikelyStatement returns true for lines that look like function-body
// statements (assignments, calls, returns) but not package-level declarations.
func isLikelyStatement(line string) bool {
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
		return false
	}

	if isPackageLevelDecl(line) || strings.HasPrefix(line, "package ") ||
		strings.HasPrefix(line, "func ") {
		return false
	}

	return isAssignmentStatement(line) ||
		strings.HasPrefix(line, "return ") ||
		strings.HasPrefix(line, "defer ") ||
		isFunctionCall(line)
}

// isAssignmentStatement checks if a line is an assignment (x = ... or x := ...).
func isAssignmentStatement(line string) bool {
	idx := assignmentIndex(line)
	if idx <= 0 {
		return false
	}

	left := strings.TrimSpace(line[:idx])

	return left != "" && !strings.Contains(left, " ")
}

// isFunctionCall checks if a line looks like a function call.
func isFunctionCall(line string) bool {
	idx := strings.Index(line, "(")
	if idx <= 0 {
		return false
	}

	before := strings.TrimSpace(line[:idx])

	return before != "" && !strings.ContainsAny(before, "=\"'")
}

// assignmentIndex returns the index of the assignment operator (= or :=)
// in a line, handling := correctly. Returns -1 if no assignment is found.
func assignmentIndex(line string) int {
	if idx := strings.Index(line, ":="); idx > 0 {
		return idx
	}

	return indexSingleEquals(line)
}

// indexSingleEquals finds a standalone = (not ==, <=, >=, !=).
func indexSingleEquals(line string) int {
	for i, r := range line {
		if r != '=' {
			continue
		}

		prev := byte(0)
		if i > 0 {
			prev = line[i-1]
		}

		if prev == '<' || prev == '>' || prev == '!' || prev == '=' {
			continue
		}

		next := byte(0)
		if i+1 < len(line) {
			next = line[i+1]
		}

		if next == '=' {
			continue
		}

		return i
	}

	return -1
}

// splitImportsAndStatements separates a snippet into a leading import block
// (lines starting with `import`) and the remaining statements. Returns
// ok=false if no import declarations are found.
func splitImportsAndStatements(code string) (string, string, bool) {
	lines := strings.Split(code, "\n")

	var importLines, otherLines []string

	inMultilineImport := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inMultilineImport {
			importLines = append(importLines, line)

			if strings.Contains(trimmed, ")") {
				inMultilineImport = false
			}

			continue
		}

		if strings.HasPrefix(trimmed, "import ") || trimmed == "import" {
			importLines = append(importLines, line)

			if strings.Contains(trimmed, "(") && !strings.Contains(trimmed, ")") {
				inMultilineImport = true
			}

			continue
		}

		otherLines = append(otherLines, line)
	}

	if len(importLines) == 0 {
		return "", "", false
	}

	return strings.Join(importLines, "\n"), strings.Join(otherLines, "\n"), true
}
