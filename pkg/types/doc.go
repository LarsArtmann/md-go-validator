// Package types provides domain types for md-go-validator.
// This package centralizes all domain types to ensure type safety
// and prevent split-brain data structures.
//
// Key principles:
// - Branded types prevent mixing unrelated values (e.g., FileID with BlockID)
// - Immutability where possible
// - Clear semantics through type naming
package types
