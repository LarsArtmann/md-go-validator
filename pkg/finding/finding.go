// Package finding provides a bridge between md-go-validator's Result type
// and the neutral go-finding Finding type. This allows embedders to convert
// validation results into SARIF, LSP diagnostics, or any other finding format
// with a single function call.
package finding

import (
	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// ToolName is the tool name used in generated findings.
const ToolName = "md-go-validator"

// RuleName is the rule name for syntax validation findings.
const RuleName = "md-codeblock-syntax"

// FromResult converts a single validation Result into a neutral Finding.
// Non-error results (valid or skipped) return a zero-value Finding and false.
func FromResult(r types.Result) (finding.Finding, bool) {
	if !r.HasError() {
		return finding.Finding{}, false //nolint:exhaustruct // zero-value is intentional for non-error results
	}

	pos := finding.Position{
		File:   r.File.String(),
		Line:   r.LineNumber.Int(),
		Column: 1,
		Offset: -1,
	}

	result := finding.NewFinding(
		finding.RuleName(RuleName),
		finding.ToolName(ToolName),
		r.ErrorMessage(),
		finding.SeverityError,
		pos,
		finding.Confidence(1.0),
	)

	return result, true
}

// FromResults converts a slice of validation Results into Findings.
// Only error results are included; valid and skipped results are filtered out.
func FromResults(results []types.Result) []finding.Finding {
	findings := make([]finding.Finding, 0, len(results))

	for _, r := range results {
		if f, ok := FromResult(r); ok {
			findings = append(findings, f)
		}
	}

	return findings
}
