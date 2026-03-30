# Clone Analysis Report

Based on art-dupl analysis with semantic sorting and threshold of 30 tokens, the following clone groups were identified:

## Clone Groups Summary

| Group | Files | Lines | Clone Count | Description |
|-------|-------|-------|-------------|-------------|
| 1 | pkg/output/output_test.go | 189,197,210,218,299,307,341,349,368,376,401,409 | 6 | Similar NewErrorResult test patterns |
| 2 | pkg/validator_test.go | 328,333,411,416,461,466,487,492 | 4 | Similar test setup patterns |
| 3 | pkg/context_test.go | 136,143,175,182,196,203,221,228 | 4 | Similar test setup patterns |
| 4 | pkg/types/types_test.go | 50,56,66,72,90,96,106,112 | 4 | Similar test setup patterns |
| 5 | cmd/md-go-validator/main_test.go | 338,347,419,428 | 2 | Similar test setup patterns |
| 6 | pkg/output/output_test.go, pkg/validator_test.go | 95,108,243,256 | 2 | Cross-file similar ValidResult patterns |
| 7 | pkg/output/output_test.go | 427,438,446,457 | 2 | Similar PrintReport test patterns |
| 8 | cmd/md-go-validator/main_test.go | 349,358,387,396 | 2 | Similar test setup patterns |
| 9 | pkg/output/output.go | 246,259,259,266 | 2 | Similar code in output.go |
| 10 | pkg/validator_test.go | 26,36,38,48 | 2 | Similar test setup patterns |
| 11 | pkg/validator_test.go | 50,60,90,101 | 2 | Similar test setup patterns |

## Detailed Analysis

### Group 1: output_test.go (6 clones)
Lines: 189-197, 210-218, 299-307, 341-349, 368-376, 401-409
Pattern: All contain nearly identical `types.NewErrorResult` structures with minor variations in the code string and boolean parameter to BuildReportData

### Group 2: validator_test.go (4 clones)
Lines: 328-333, 411-416, 461-466, 487-492
Pattern: All contain similar for-loop test setup patterns creating multiple temporary files

### Group 3: context_test.go (4 clones)
Lines: 136-143, 175-182, 196-203, 221-228
Pattern: Similar test setup patterns

### Group 4: types_test.go (4 clones)
Lines: 50-56, 66-72, 90-96, 106-112
Pattern: Similar test setup patterns

### Group 5: main_test.go (2 clones)
Lines: 338-347, 419-428
Pattern: Similar test setup patterns

### Group 6: Cross-file (output_test.go & validator_test.go)
Lines: output_test.go:95-108, validator_test.go:243-256
Pattern: Similar ValidResult test patterns

### Group 7: output_test.go (2 clones)
Lines: 427-438, 446-457
Pattern: Similar PrintReport test patterns

### Group 8: main_test.go (2 clones)
Lines: 349-358, 387-396
Pattern: Similar test setup patterns

### Group 9: output.go (2 clones)
Lines: 246-259, 259-266
Pattern: Similar code in output.go

### Group 10: validator_test.go (2 clones)
Lines: 26-36, 38-48
Pattern: Similar test setup patterns

### Group 11: validator_test.go (2 clones)
Lines: 50-60, 90-101
Pattern: Similar test setup patterns