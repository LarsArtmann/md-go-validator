// Package output provides formatting and output utilities for validation reports.
package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/larsartmann/go-output"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

// OutputFormat represents the output format for validation reports.
type OutputFormat = output.Format

const (
	// FormatTable outputs a formatted table for terminal display.
	FormatTable = output.FormatTable
	// FormatJSON outputs machine-readable JSON.
	FormatJSON = output.FormatJSON
	// FormatMarkdown outputs a Markdown table.
	FormatMarkdown = output.FormatMarkdown
	// FormatYAML outputs YAML format.
	FormatYAML = output.FormatYAML
	// FormatCSV outputs CSV format for spreadsheet import.
	FormatCSV = output.FormatCSV
	// FormatQuiet outputs only summary information.
	FormatQuiet OutputFormat = "quiet"
)

const (
	// ColorModeAuto respects NO_COLOR and CI environment detection.
	ColorModeAuto = output.ColorModeAuto
	// ColorModeAlways forces ANSI color output.
	ColorModeAlways = output.ColorModeAlways
	// ColorModeNever disables color output.
	ColorModeNever = output.ColorModeNever
)

// ColorMode determines when to use ANSI color codes in output.
type ColorMode = output.ColorMode

// ParseFormat converts a string format name to an OutputFormat.
func ParseFormat(s string) (OutputFormat, error) {
	switch s {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "csv":
		return FormatCSV, nil
	case "quiet", "q":
		return FormatQuiet, nil
	default:
		return "", fmt.Errorf(
			"invalid format: %q (allowed: table, json, markdown, yaml, csv, quiet)",
			s,
		)
	}
}

// ParseColorMode converts a string color mode to a ColorMode.
func ParseColorMode(s string) (ColorMode, error) {
	return output.ParseColorMode(s)
}

// PrintReport outputs validation results to stdout.
func PrintReport(results []types.Result, format OutputFormat, colorMode ColorMode, showCode bool) {
	//nolint:errcheck // Writing to stdout, cannot recover from write errors
	PrintReportTo(os.Stdout, results, format, colorMode, showCode)
}

// PrintReportTo outputs validation results to the specified writer.
func PrintReportTo(
	w io.Writer,
	results []types.Result,
	format OutputFormat,
	colorMode ColorMode,
	showCode bool,
) error {
	switch format {
	case FormatJSON:
		return printJSONTo(w, results, showCode)
	case FormatMarkdown:
		return printMarkdownTo(w, results, showCode)
	case FormatYAML:
		return printYAMLTo(w, results, showCode)
	case FormatCSV:
		return printCSVTo(w, results, showCode)
	case FormatQuiet:
		return printQuietTo(w, results)
	default:
		return printTableTo(w, results, colorMode, showCode)
	}
}

func printJSONTo(w io.Writer, results []types.Result, showCode bool) error {
	report := types.BuildReportData(results, showCode)
	data, err := output.MarshalJSONIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON (%d results, showCode=%t): %w", len(results), showCode, err)
	}
	_, err = fmt.Fprintln(w, string(data))
	if err != nil {
		return fmt.Errorf(
			"write JSON output (%d results, showCode=%t): %w",
			len(results),
			showCode,
			err,
		)
	}
	return nil
}

func printMarkdownTo(w io.Writer, results []types.Result, showCode bool) error {
	report := types.BuildReportData(results, showCode)

	_, _ = fmt.Fprintln(w, "# Validation Report")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "| Metric    | Count |\n")
	_, _ = fmt.Fprintf(w, "|-----------|-------|\n")
	_, _ = fmt.Fprintf(w, "| Total     | %d    |\n", report.Summary.Total)
	_, _ = fmt.Fprintf(w, "| Valid     | %d    |\n", report.Summary.Valid)
	_, _ = fmt.Fprintf(w, "| Skipped   | %d    |\n", report.Summary.Skipped)
	_, _ = fmt.Fprintf(w, "| Errors    | %d    |\n", report.Summary.Errors)
	_, _ = fmt.Fprintln(w)

	if len(report.Errors) > 0 {
		_, _ = fmt.Fprintln(w, "## Errors")
		_, _ = fmt.Fprintln(w)
		if showCode {
			_, _ = fmt.Fprintln(w, "| File | Line | Block | Error | Code |")
			_, _ = fmt.Fprintln(w, "|------|------|-------|-------|------|")
			for _, e := range report.Errors {
				code := truncateCode(e.Code, 50)
				_, _ = fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error, code)
			}
		} else {
			_, _ = fmt.Fprintln(w, "| File | Line | Block | Error |")
			_, _ = fmt.Fprintln(w, "|------|------|-------|-------|")
			for _, e := range report.Errors {
				_, _ = fmt.Fprintf(w, "| %s | %s | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error)
			}
		}
	}
	return nil
}

func printYAMLTo(w io.Writer, results []types.Result, showCode bool) error {
	report := types.BuildReportData(results, showCode)
	data, err := output.MarshalYAML(report)
	if err != nil {
		return fmt.Errorf("marshal YAML (%d results, showCode=%t): %w", len(results), showCode, err)
	}
	_, err = fmt.Fprintln(w, string(data))
	if err != nil {
		return fmt.Errorf(
			"write YAML output (%d results, showCode=%t): %w",
			len(results),
			showCode,
			err,
		)
	}
	return nil
}

func printCSVTo(w io.Writer, results []types.Result, showCode bool) error {
	csvWriter := output.NewCSVWriter(w)
	if err := csvWriter.WriteHeader(
		[]string{"file", "line", "block", "status", "error", "code"},
	); err != nil {
		return fmt.Errorf(
			"write CSV header (%d results, showCode=%t): %w",
			len(results),
			showCode,
			err,
		)
	}

	for _, r := range results {
		var errMsg, code string
		if r.Error != nil {
			errMsg = r.Error.Error()
		}
		if showCode {
			code = r.Code
		}
		if err := csvWriter.WriteRow([]string{
			r.File.String(),
			r.LineNumber.String(),
			r.Block.String(),
			r.Status.String(),
			errMsg,
			code,
		}); err != nil {
			return fmt.Errorf("write CSV row (file=%s, line=%s, showCode=%t, errMsg=%q): %w",
				r.File, r.LineNumber, showCode, errMsg, err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("flush CSV (%d rows, showCode=%t): %w", len(results), showCode, err)
	}
	return nil
}

func printQuietTo(w io.Writer, results []types.Result) error {
	report := types.BuildReportData(results, false)
	if report.Summary.Errors > 0 {
		_, err := fmt.Fprintf(w, "%d errors found\n", report.Summary.Errors)
		if err != nil {
			return fmt.Errorf("write quiet output (%d results, %d errors): %w",
				len(results), report.Summary.Errors, err)
		}
		return nil
	}
	_, err := fmt.Fprintf(w, "All %d code blocks valid\n", report.Summary.Valid)
	if err != nil {
		return fmt.Errorf("write quiet output (%d valid results): %w",
			len(results), err)
	}
	return nil
}

func printTableTo(w io.Writer, results []types.Result, colorMode ColorMode, showCode bool) error {
	report := types.BuildReportData(results, showCode)
	shouldColor := colorMode.ShouldColor()

	printTableHeaderTo(w, report.Summary, shouldColor)
	printTableErrorsTo(w, report.Errors, showCode, shouldColor)
	return nil
}

func printTableHeaderTo(w io.Writer, summary types.ReportSummary, shouldColor bool) {
	if shouldColor {
		_, _ = fmt.Fprintln(
			w,
			"\033[1;36m============================================================\033[0m",
		)
		_, _ = fmt.Fprintln(w, "\033[1;36m📊 VALIDATION REPORT\033[0m")
		_, _ = fmt.Fprintln(
			w,
			"\033[1;36m============================================================\033[0m",
		)
		_, _ = fmt.Fprintf(w, "\033[1;32m✅ Valid:\033[0m %d\n", summary.Valid)
		_, _ = fmt.Fprintf(w, "\033[33m⏭️  Skipped:\033[0m %d\n", summary.Skipped)
		_, _ = fmt.Fprintf(w, "\033[1;31m❌ Errors:\033[0m %d\n", summary.Errors)
	} else {
		_, _ = fmt.Fprintln(w, "\n============================================================")
		_, _ = fmt.Fprintln(w, "VALIDATION REPORT")
		_, _ = fmt.Fprintln(w, "============================================================")
		_, _ = fmt.Fprintf(w, "Valid: %d\n", summary.Valid)
		_, _ = fmt.Fprintf(w, "Skipped: %d\n", summary.Skipped)
		_, _ = fmt.Fprintf(w, "Errors: %d\n", summary.Errors)
	}
	_, _ = fmt.Fprintln(w, "============================================================")
}

func printTableErrorsTo(w io.Writer, errors []types.ErrorEntry, showCode, shouldColor bool) {
	if len(errors) == 0 {
		return
	}

	_, _ = fmt.Fprintln(w)
	if shouldColor {
		_, _ = fmt.Fprintln(w, "\033[1;31mERRORS FOUND:\033[0m")
	} else {
		_, _ = fmt.Fprintln(w, "ERRORS FOUND:")
	}
	_, _ = fmt.Fprintln(w, "------------------------------------------------------------")

	for _, e := range errors {
		fileLoc := fmt.Sprintf("%s:%s (block #%s)", e.File, e.Line, e.Block)
		if shouldColor {
			_, _ = fmt.Fprintf(w, "\n\033[1;33m📍 %s\033[0m\n", fileLoc)
			_, _ = fmt.Fprintf(w, "   \033[1;31mError:\033[0m %s\n", e.Error)
		} else {
			_, _ = fmt.Fprintf(w, "\n📍 %s\n", fileLoc)
			_, _ = fmt.Fprintf(w, "   Error: %s\n", e.Error)
		}

		if showCode && e.Code != "" {
			_, _ = fmt.Fprintln(w, "\n   Code:")
			_, _ = fmt.Fprintln(w, "   "+"------------------------------------------------")
			for i, line := range strings.Split(e.Code, "\n") {
				_, _ = fmt.Fprintf(w, "   %3d | %s\n", i+1, line)
			}
			_, _ = fmt.Fprintln(w, "   "+"------------------------------------------------")
		}
	}
	_, _ = fmt.Fprintln(w)
}

func truncateCode(code string, maxLen uint) string {
	if code == "" {
		return ""
	}
	if uint(len(code)) <= maxLen {
		return code
	}
	if maxLen <= 3 {
		return "..."
	}
	// maxLen > 3 and len(code) > maxLen, so this is safe
	return code[:maxLen-3] + "..."
}
