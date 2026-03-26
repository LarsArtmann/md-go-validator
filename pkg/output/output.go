package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/larsartmann/go-output"
	"github.com/larsartmann/md-go-validator/pkg/types"
)

type OutputFormat = output.Format

const (
	FormatTable                 = output.FormatTable
	FormatJSON                  = output.FormatJSON
	FormatMarkdown              = output.FormatMarkdown
	FormatYAML                  = output.FormatYAML
	FormatCSV                   = output.FormatCSV
	FormatQuiet    OutputFormat = "quiet"
)

const (
	ColorModeAuto   = output.ColorModeAuto
	ColorModeAlways = output.ColorModeAlways
	ColorModeNever  = output.ColorModeNever
)

type ColorMode = output.ColorMode

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
		return "", fmt.Errorf("invalid format: %q (allowed: table, json, markdown, yaml, csv, quiet)", s)
	}
}

func ParseColorMode(s string) (ColorMode, error) {
	return output.ParseColorMode(s)
}

func PrintReport(results []types.Result, format OutputFormat, colorMode ColorMode, showCode bool) {
	PrintReportTo(os.Stdout, results, format, colorMode, showCode)
}

func PrintReportTo(w io.Writer, results []types.Result, format OutputFormat, colorMode ColorMode, showCode bool) error {
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
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func printMarkdownTo(w io.Writer, results []types.Result, showCode bool) error {
	report := types.BuildReportData(results, showCode)

	fmt.Fprintln(w, "# Validation Report")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "| Metric    | Count |\n")
	fmt.Fprintf(w, "|-----------|-------|\n")
	fmt.Fprintf(w, "| Total     | %d    |\n", report.Summary.Total)
	fmt.Fprintf(w, "| Valid     | %d    |\n", report.Summary.Valid)
	fmt.Fprintf(w, "| Skipped   | %d    |\n", report.Summary.Skipped)
	fmt.Fprintf(w, "| Errors    | %d    |\n", report.Summary.Errors)
	fmt.Fprintln(w)

	if len(report.Errors) > 0 {
		fmt.Fprintln(w, "## Errors")
		fmt.Fprintln(w)
		if showCode {
			fmt.Fprintln(w, "| File | Line | Block | Error | Code |")
			fmt.Fprintln(w, "|------|------|-------|-------|------|")
			for _, e := range report.Errors {
				code := truncateCode(e.Code, 50)
				fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error, code)
			}
		} else {
			fmt.Fprintln(w, "| File | Line | Block | Error |")
			fmt.Fprintln(w, "|------|------|-------|-------|")
			for _, e := range report.Errors {
				fmt.Fprintf(w, "| %s | %s | %s | %s |\n",
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
		return fmt.Errorf("marshal YAML: %w", err)
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

func printCSVTo(w io.Writer, results []types.Result, showCode bool) error {
	csvWriter := output.NewCSVWriter(w)
	if err := csvWriter.WriteHeader([]string{"file", "line", "block", "status", "error", "code"}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
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
			return fmt.Errorf("write CSV row: %w", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	return nil
}

func printQuietTo(w io.Writer, results []types.Result) error {
	report := types.BuildReportData(results, false)
	if report.Summary.Errors > 0 {
		_, err := fmt.Fprintf(w, "%d errors found\n", report.Summary.Errors)
		return err
	}
	_, err := fmt.Fprintf(w, "All %d code blocks valid\n", report.Summary.Valid)
	return err
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
		fmt.Fprintln(w, "\033[1;36m============================================================\033[0m")
		fmt.Fprintln(w, "\033[1;36m📊 VALIDATION REPORT\033[0m")
		fmt.Fprintln(w, "\033[1;36m============================================================\033[0m")
		fmt.Fprintf(w, "\033[1;32m✅ Valid:\033[0m %d\n", summary.Valid)
		fmt.Fprintf(w, "\033[33m⏭️  Skipped:\033[0m %d\n", summary.Skipped)
		fmt.Fprintf(w, "\033[1;31m❌ Errors:\033[0m %d\n", summary.Errors)
	} else {
		fmt.Fprintln(w, "\n============================================================")
		fmt.Fprintln(w, "VALIDATION REPORT")
		fmt.Fprintln(w, "============================================================")
		fmt.Fprintf(w, "Valid: %d\n", summary.Valid)
		fmt.Fprintf(w, "Skipped: %d\n", summary.Skipped)
		fmt.Fprintf(w, "Errors: %d\n", summary.Errors)
	}
	fmt.Fprintln(w, "============================================================")
}

func printTableErrorsTo(w io.Writer, errors []types.ErrorEntry, showCode, shouldColor bool) {
	if len(errors) == 0 {
		return
	}

	fmt.Fprintln(w)
	if shouldColor {
		fmt.Fprintln(w, "\033[1;31mERRORS FOUND:\033[0m")
	} else {
		fmt.Fprintln(w, "ERRORS FOUND:")
	}
	fmt.Fprintln(w, "------------------------------------------------------------")

	for _, e := range errors {
		fileLoc := fmt.Sprintf("%s:%s (block #%s)", e.File, e.Line, e.Block)
		if shouldColor {
			fmt.Fprintf(w, "\n\033[1;33m📍 %s\033[0m\n", fileLoc)
			fmt.Fprintf(w, "   \033[1;31mError:\033[0m %s\n", e.Error)
		} else {
			fmt.Fprintf(w, "\n📍 %s\n", fileLoc)
			fmt.Fprintf(w, "   Error: %s\n", e.Error)
		}

		if showCode && e.Code != "" {
			fmt.Fprintln(w, "\n   Code:")
			fmt.Fprintln(w, "   "+"------------------------------------------------")
			for i, line := range strings.Split(e.Code, "\n") {
				fmt.Fprintf(w, "   %3d | %s\n", i+1, line)
			}
			fmt.Fprintln(w, "   "+"------------------------------------------------")
		}
	}
	fmt.Fprintln(w)
}

func truncateCode(code string, maxLen uint) string {
	if code == "" {
		return ""
	}
	truncated := code
	if len(truncated) > int(maxLen) {
		truncated = truncated[:maxLen-3] + "..."
	}
	return truncated
}
