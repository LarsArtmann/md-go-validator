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

// Format represents the output format for validation reports.
type Format = output.Format

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
	FormatQuiet Format = "quiet"
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

// ParseFormat converts a string format name to a Format.
func ParseFormat(s string) (Format, error) {
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
	cm, err := output.ParseColorMode(s)
	if err != nil {
		return "", fmt.Errorf("parse color mode %q: %w", s, err)
	}

	return cm, nil
}

// PrintReport outputs validation results to stdout.
func PrintReport(results []types.Result, format Format, colorMode ColorMode, showCode bool) {
	//nolint:errcheck,gosec // Writing to stdout, cannot recover from write errors
	PrintReportTo(os.Stdout, results, format, colorMode, showCode)
}

// PrintReportTo outputs validation results to the specified writer.
func PrintReportTo(
	w io.Writer,
	results []types.Result,
	format Format,
	colorMode ColorMode,
	showCode bool,
) error {
	switch format {
	case FormatJSON:
		return marshalReport(w, results, showCode, func(r any) ([]byte, error) {
			return output.MarshalJSONIndent(r, "", "  ")
		}, "JSON")
	case FormatMarkdown:
		return printMarkdownTo(w, results, showCode)
	case FormatYAML:
		return marshalReport(w, results, showCode, output.MarshalYAML, "YAML")
	case FormatCSV:
		return printCSVTo(w, results, showCode)
	case FormatQuiet:
		return printQuietTo(w, results)
	default:
		return printTableTo(w, results, colorMode, showCode)
	}
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

func marshalReport(
	w io.Writer,
	results []types.Result,
	showCode bool,
	marshalFn func(any) ([]byte, error),
	formatName string,
) error {
	report := types.BuildReportData(results, showCode)

	data, err := marshalFn(report)
	if err != nil {
		return fmt.Errorf(
			"marshal %s (%d results, showCode=%t): %w",
			formatName,
			len(results),
			showCode,
			err,
		)
	}

	return writeOutput(w, data, len(results), showCode, formatName)
}

// writeOutput writes marshaled data to the writer with consistent error handling.
func writeOutput(
	w io.Writer,
	data []byte,
	resultCount int,
	showCode bool,
	formatName string,
) error {
	_, err := fmt.Fprintln(w, string(data))
	if err != nil {
		return fmt.Errorf(
			"write %s output (%d results, showCode=%t): %w",
			formatName,
			resultCount,
			showCode,
			err,
		)
	}

	return nil
}

// newOutputError creates a formatted error for output write operations.
func newOutputError(action string, results []types.Result, showCode bool, err error) error {
	return fmt.Errorf("write %s (%d results, showCode=%t): %w", action, len(results), showCode, err)
}

func printCSVTo(writer io.Writer, results []types.Result, showCode bool) error {
	csvWriter := output.NewCSVWriter(writer)

	err := csvWriter.WriteHeader(
		[]string{"file", "line", "block", "status", "error", "code"},
	)
	if err != nil {
		return newOutputError("CSV header", results, showCode, err)
	}

	writeErr := writeCSVRows(csvWriter, results, showCode)
	if writeErr != nil {
		return writeErr
	}

	csvWriter.Flush()

	flushErr := csvWriter.Error()
	if flushErr != nil {
		return newOutputError("CSV flush", results, showCode, flushErr)
	}

	return nil
}

func writeCSVRows(csvWriter *output.CSVWriter, results []types.Result, showCode bool) error {
	for _, r := range results {
		var errMsg, code string
		if r.Error != nil {
			errMsg = r.Error.Error()
		}

		if showCode {
			code = r.Code
		}

		row := []string{
			r.File.String(),
			r.LineNumber.String(),
			r.Block.String(),
			r.Status.String(),
			errMsg,
			code,
		}

		err := csvWriter.WriteRow(row)
		if err != nil {
			return fmt.Errorf(
				"write CSV row (file=%s, line=%s, block=%s): %w",
				r.File, r.LineNumber, r.Block, err,
			)
		}
	}

	return nil
}

func printQuietTo(w io.Writer, results []types.Result) error {
	report := types.BuildReportData(results, false)
	if report.Summary.Errors > 0 {
		_, err := fmt.Fprintf(w, "%d errors found\n", report.Summary.Errors)
		if err != nil {
			return newOutputError("quiet output", results, false, err)
		}

		return nil
	}

	_, err := fmt.Fprintf(w, "All %d code blocks valid\n", report.Summary.Valid)
	if err != nil {
		return newOutputError("quiet output", results, false, err)
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
	divider := "============================================================"
	reportLabel := "VALIDATION REPORT"
	validLabel := fmt.Sprintf("Valid: %d", summary.Valid)
	skippedLabel := fmt.Sprintf("Skipped: %d", summary.Skipped)
	errorsLabel := fmt.Sprintf("Errors: %d", summary.Errors)

	if shouldColor {
		_, _ = fmt.Fprintln(w, "\033[1;36m"+divider+"\033[0m")
		_, _ = fmt.Fprintln(w, "\033[1;36m📊 "+reportLabel+"\033[0m")
		_, _ = fmt.Fprintln(w, "\033[1;36m"+divider+"\033[0m")
		_, _ = fmt.Fprintf(w, "\033[1;32m✅ %s\033[0m\n", validLabel)
		_, _ = fmt.Fprintf(w, "\033[33m⏭️  %s\033[0m\n", skippedLabel)
		_, _ = fmt.Fprintf(w, "\033[1;31m❌ %s\033[0m\n", errorsLabel)
	} else {
		_, _ = fmt.Fprintln(w, "\n"+divider)
		_, _ = fmt.Fprintln(w, reportLabel)
		_, _ = fmt.Fprintln(w, divider)
		_, _ = fmt.Fprintln(w, validLabel)
		_, _ = fmt.Fprintln(w, skippedLabel)
		_, _ = fmt.Fprintln(w, errorsLabel)
	}

	_, _ = fmt.Fprintln(w, divider)
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
		printErrorEntry(w, fileLoc, e.Error, shouldColor)

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

const (
	ansiReset    = "\033[0m"
	ansiBold     = "\033[1m"
	ansiYellow   = "\033[1;33m"
	ansiRed      = "\033[1;31m"
	ansiLocation = "📍"
	ansiError    = "Error:"
)

func printErrorEntry(w io.Writer, fileLoc, errMsg string, shouldColor bool) {
	if shouldColor {
		_, _ = fmt.Fprintf(w, "\n%s%s %s%s\n", ansiBold, ansiYellow, fileLoc, ansiReset)
		_, _ = fmt.Fprintf(w, "   %s%s%s %s%s\n", ansiBold, ansiRed, ansiError, ansiReset, errMsg)
	} else {
		_, _ = fmt.Fprintf(w, "\n%s %s\n", ansiLocation, fileLoc)
		_, _ = fmt.Fprintf(w, "   %s %s\n", ansiError, errMsg)
	}
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
