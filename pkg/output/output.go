package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/larsartmann/go-output"
	mdgovalidator "github.com/larsartmann/md-go-validator/pkg"
)

type OutputFormat = output.Format

const (
	FormatTable    = output.FormatTable
	FormatJSON     = output.FormatJSON
	FormatMarkdown = output.FormatMarkdown
	FormatYAML     = output.FormatYAML
	FormatCSV   = output.FormatCSV
	FormatQuiet OutputFormat = "quiet"
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

type ReportData struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

type ResultData struct {
	File       string `json:"file"`
	LineNumber int    `json:"line"`
	Block      int    `json:"block"`
	Code       string `json:"code,omitempty"`
	Skipped    bool   `json:"skipped"`
	Error      string `json:"error,omitempty"`
}

type ReportOutput struct {
	Summary ReportSummary `json:"summary"`
	Errors  []ErrorEntry  `json:"errors,omitempty"`
}

type ReportSummary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Skipped int `json:"skipped"`
	Errors  int `json:"errors"`
}

type ErrorEntry struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Block      int    `json:"block"`
	Error      string `json:"error"`
	Code       string `json:"code,omitempty"`
}

func buildReportData(results []mdgovalidator.Result, showCode bool) ReportOutput {
	var valid, skipped int
	var errorEntries []ErrorEntry

	for _, r := range results {
		switch {
		case r.Skipped:
			skipped++
		case r.Error != nil:
			entry := ErrorEntry{
				File:  r.File,
				Line:  r.LineNumber,
				Block: r.CodeBlock,
			}
			if r.Error != nil {
				entry.Error = r.Error.Error()
			}
			if showCode {
				entry.Code = r.Code
			}
			errorEntries = append(errorEntries, entry)
		default:
			valid++
		}
	}

	return ReportOutput{
		Summary: ReportSummary{
			Total:   len(results),
			Valid:   valid,
			Skipped: skipped,
			Errors:  len(errorEntries),
		},
		Errors: errorEntries,
	}
}

func PrintReport(results []mdgovalidator.Result, format OutputFormat, colorMode ColorMode, showCode bool) {
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

func printJSON(results []mdgovalidator.Result, showCode bool) {
	report := buildReportData(results, showCode)
	data, err := output.MarshalJSONIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printMarkdown(results []mdgovalidator.Result, showCode bool) {
	report := buildReportData(results, showCode)

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
				code := strings.ReplaceAll(e.Code, "\n", "\\n")
				if len(code) > 50 {
					code = code[:47] + "..."
				}
				fmt.Printf("| %s | %d | %d | %s | %s |\n",
					e.File, e.Line, e.Block, e.Error, code)
			}
		} else {
			fmt.Println("| File | Line | Block | Error |")
			fmt.Println("|------|------|-------|-------|")
			for _, e := range report.Errors {
				fmt.Printf("| %s | %d | %d | %s |\n",
					e.File, e.Line, e.Block, e.Error)
			}
		}
	}
}

func printYAML(results []mdgovalidator.Result, showCode bool) {
	report := buildReportData(results, showCode)
	data, err := output.MarshalYAML(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func printCSV(results []mdgovalidator.Result, showCode bool) {
	fmt.Println("file,line,block,skipped,error,code")
	for _, r := range results {
		status := "valid"
		var errMsg string
		if r.Skipped {
			status = "skipped"
		} else if r.Error != nil {
			status = "error"
			errMsg = r.Error.Error()
		}
		code := ""
		if showCode {
			code = r.Code
		}
		fmt.Printf("%s,%d,%d,%s,%s,%s\n",
			escapeCSV(r.File),
			r.LineNumber,
			r.CodeBlock,
			status,
			escapeCSV(errMsg),
			escapeCSV(code))
	}
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func printQuiet(results []mdgovalidator.Result) {
	report := buildReportData(results, false)
	if report.Summary.Errors > 0 {
		fmt.Printf("%d errors found\n", report.Summary.Errors)
	} else {
		fmt.Printf("All %d code blocks valid\n", report.Summary.Valid)
	}
}

// printTable prints results in a formatted table.
//colorMode is kept for future use with lipgloss styling.
//nolint:unparam // colorMode reserved for future color support
func printTable(results []mdgovalidator.Result, colorMode ColorMode, showCode bool) {
	report := buildReportData(results, showCode)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("VALIDATION REPORT")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total code blocks: %d\n", report.Summary.Total)
	fmt.Printf("Valid: %d\n", report.Summary.Valid)
	fmt.Printf("Skipped: %d\n", report.Summary.Skipped)
	fmt.Printf("Errors: %d\n", report.Summary.Errors)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	if len(report.Errors) > 0 {
		fmt.Println("ERRORS FOUND:")
		fmt.Println(strings.Repeat("-", 60))
		for _, e := range report.Errors {
			fmt.Printf("\n%s:%d (block #%d)\n", e.File, e.Line, e.Block)
			fmt.Printf("  Error: %s\n", e.Error)
			if showCode && e.Code != "" {
				fmt.Println("\n  Code:")
				fmt.Println("  " + strings.Repeat("-", 40))
				for i, line := range strings.Split(e.Code, "\n") {
					fmt.Printf("  %3d | %s\n", i+1, line)
				}
				fmt.Println("  " + strings.Repeat("-", 40))
			}
		}
		fmt.Println()
	}
}
