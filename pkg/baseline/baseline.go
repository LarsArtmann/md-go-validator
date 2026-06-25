// Package baseline provides regression-mode filtering for validation results.
// A baseline file contains known error signatures (file:line). When a baseline
// is loaded, only NEW errors (not in the baseline) are reported as failures.
// This enables incremental adoption on noisy repositories.
package baseline

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/larsartmann/md-go-validator/pkg/types"
)

// Set represents a collection of known error signatures.
type Set struct {
	signatures map[string]bool
}

// Signature builds a unique key for a result error: "file:line:code".
// Includes the error code so that an error changing type on the same line
// is treated as a new error rather than being suppressed.
func Signature(r types.Result) string {
	return fmt.Sprintf("%s:%d:%s", r.File.String(), r.LineNumber.Int(), r.ErrorCode)
}

// Load reads a baseline file containing one signature per line (file:line).
// Blank lines and lines starting with # are ignored.
func Load(path string) (Set, error) {
	file, err := os.Open(path) //nolint:gosec // G304: path is user-controlled baseline file
	if err != nil {
		return Set{}, fmt.Errorf("open baseline file %s: %w", path, err)
	}

	defer func() {
		_ = file.Close()
	}()

	signatures := make(map[string]bool)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		signatures[line] = true
	}

	err = scanner.Err()
	if err != nil {
		return Set{}, fmt.Errorf("read baseline file %s: %w", path, err)
	}

	return Set{signatures: signatures}, nil
}

// Contains returns true if the given result's error is in the baseline.
func (s Set) Contains(r types.Result) bool {
	return s.signatures[Signature(r)]
}

// IsEmpty returns true if the baseline set has no entries.
func (s Set) IsEmpty() bool {
	return len(s.signatures) == 0
}

// FilterNew returns only results whose errors are NOT in the baseline.
func (s Set) FilterNew(results []types.Result) []types.Result {
	filtered := make([]types.Result, 0, len(results))

	for _, r := range results {
		if !r.HasError() {
			filtered = append(filtered, r)

			continue
		}

		if !s.Contains(r) {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

// Count returns the number of signatures in the baseline.
func (s Set) Count() int {
	return len(s.signatures)
}

// Save writes error signatures from results to a file, one per line.
// Only error results are included; each line is "file:line:errorcode".
// Lines starting with # are treated as comments when loaded.
func Save(path string, results []types.Result) error {
	var lines []string

	for _, r := range results {
		if r.HasError() {
			lines = append(lines, Signature(r))
		}
	}

	content := strings.Join(lines, "\n") + "\n"

	err := os.WriteFile(path, []byte(content), baselineFilePerms)
	if err != nil {
		return fmt.Errorf("write baseline file %s: %w", path, err)
	}

	return nil
}

const baselineFilePerms = 0o600
