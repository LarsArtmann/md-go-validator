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

// Signature builds a unique key for a result error: "file:line".
func Signature(r types.Result) string {
	return fmt.Sprintf("%s:%d", r.File.String(), r.LineNumber.Int())
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
