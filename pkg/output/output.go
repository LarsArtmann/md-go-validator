package output

import (
	"fmt"
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
	switch format {
	case FormatJSON:
		printJSON(results, showCode)
	case FormatMarkdown:
		printMarkdown(results, showCode)
	case FormatYAML:
		printYAML(results, showCode)
	case FormatCSV:
		printCSV(results, showCode)
	case FormatQuiet:
		printQuiet(results)
	default:
		printTable(results, colorMode, showCode)
	}
}

func printJSON(results []types.Result, showCode bool) {
	report := types.BuildReportData(results, showCode)
	data, err := output.MarshalJSONIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printMarkdown(results []types.Result, showCode bool) {
	report := types.BuildReportData(results, showCode)

	fmt.Println("# Validation Report")
	fmt.Println()
	fmt.Printf("| Metric    | Count |\n")
	fmt.Printf("|-----------|-------|\n")
	fmt.Printf("| Total     | %d    |\n", report.Summary.Total)
	fmt.Printf("| Valid     | %d    |\n", report.Summary.Valid)
	fmt.Printf("| Skipped   | %d    |\n", report.Summary.Skipped)
	fmt.Printf("| Errors    | %d    |\n", report.Summary.Errors)
	fmt.Println()

	if len(report.Errors) > 0 {
		fmt.Println("## Errors")
		fmt.Println()
		if showCode {
			fmt.Println("| File | Line | Block | Error | Code |")
			fmt.Println("|------|------|-------|-------|------|")
			for _, e := range report.Errors {
				code := truncateCode(e.Code, 50)
				fmt.Printf("| %s | %s | %s | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error, code)
			}
		} else {
			fmt.Println("| File | Line | Block | Error |")
			fmt.Println("|------|------|-------|-------|")
			for _, e := range report.Errors {
				fmt.Printf("| %s | %s | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error)
			}
		}
	}
}

func printYAML(results []types.Result, showCode bool) {
	report := types.BuildReportData(results, showCode)
	data, err := output.MarshalYAML(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printCSV(results []types.Result, showCode bool) {
	csvWriter := output.NewCSVWriter(os.Stdout)
	csvWriter.WriteHeader([]string{"file", "line", "block", "status", "error", "code"})

	for _, r := range results {
		var errMsg, code string
		if r.Error != nil {
			errMsg = r.Error.Error()
		}
		if showCode {
			code = r.Code
		}
		csvWriter.WriteRow([]string{
			r.File.String(),
			r.LineNumber.String(),
			r.Block.String(),
			r.Status.String(),
			errMsg,
			code,
		})
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
	}
}

func printQuiet(results []types.Result) {
	report := types.BuildReportData(results, false)
	if report.Summary.Errors > 0 {
		fmt.Printf("%d errors found\n", report.Summary.Errors)
	} else {
		fmt.Printf("All %d code blocks valid\n", report.Summary.Valid)
	}
}

func printTable(results []types.Result, colorMode ColorMode, showCode bool) {
	report := types.BuildReportData(results, showCode)
	shouldColor := colorMode.ShouldColor()

	printTableHeader(report.Summary, shouldColor)
	printTableErrors(report.Errors, showCode, shouldColor)
}

func printTableHeader(summary types.ReportSummary, shouldColor bool) {
	if shouldColor {
		fmt.Println("\033[1;36m============================================================\033[0m")
		fmt.Println("\033[1;36m📊 VALIDATION REPORT\033[0m")
		fmt.Println("\033[1;36m============================================================\033[0m")
		fmt.Printf("\033[1;32m✅ Valid:\033[0m %d\n", summary.Valid)
		fmt.Printf("\033[33m⏭️  Skipped:\033[0m %d\n", summary.Skipped)
		fmt.Printf("\033[1;31m❌ Errors:\033[0m %d\n", summary.Errors)
	} else {
		fmt.Println("\n============================================================")
		fmt.Println("VALIDATION REPORT")
		fmt.Println("============================================================")
		fmt.Printf("Valid: %d\n", summary.Valid)
		fmt.Printf("Skipped: %d\n", summary.Skipped)
		fmt.Printf("Errors: %d\n", summary.Errors)
	}
	fmt.Println("============================================================")
}

func printTableErrors(errors []types.ErrorEntry, showCode, shouldColor bool) {
	if len(errors) == 0 {
		return
	}

	fmt.Println()
	if shouldColor {
		fmt.Println("\033[1;31mERRORS FOUND:\033[0m")
	} else {
		fmt.Println("ERRORS FOUND:")
	}
	fmt.Println("------------------------------------------------------------")

	for _, e := range errors {
		fileLoc := fmt.Sprintf("%s:%s (block #%s)", e.File, e.Line, e.Block)
		if shouldColor {
			fmt.Printf("\n\033[1;33m📍 %s\033[0m\n", fileLoc)
			fmt.Printf("   \033[1;31mError:\033[0m %s\n", e.Error)
		} else {
			fmt.Printf("\n📍 %s\n", fileLoc)
			fmt.Printf("   Error: %s\n", e.Error)
		}

		if showCode && e.Code != "" {
			fmt.Println("\n   Code:")
			fmt.Println("   " + "------------------------------------------------")
			for i, line := range strings.Split(e.Code, "\n") {
				fmt.Printf("   %3d | %s\n", i+1, line)
			}
			fmt.Println("   " + "------------------------------------------------")
		}
	}
	fmt.Println()
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
