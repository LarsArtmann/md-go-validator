package code

import (
	"testing"
)

func TestIsPseudoModuleFile_RequireBlock(t *testing.T) {
	t.Parallel()

	code := `require (
    github.com/acme/core v1.6.0
    github.com/acme/lib v2.3.1
)`

	if !IsPseudoModuleFile(code) {
		t.Error("expected true for require block")
	}
}

func TestIsPseudoModuleFile_ReplaceDirective(t *testing.T) {
	t.Parallel()

	code := `replace github.com/acme/core => ../core
replace github.com/acme/lib => ../lib`

	if !IsPseudoModuleFile(code) {
		t.Error("expected true for replace directives")
	}
}

func TestIsPseudoModuleFile_ModuleFile(t *testing.T) {
	t.Parallel()

	code := `module github.com/acme/myproject

go 1.21

require github.com/acme/core v1.6.0`

	if !IsPseudoModuleFile(code) {
		t.Error("expected true for module file content")
	}
}

func TestIsPseudoModuleFile_GoSource(t *testing.T) {
	t.Parallel()

	code := `package main

import "fmt"

func main() {
    fmt.Println("hello")
}`

	if IsPseudoModuleFile(code) {
		t.Error("expected false for Go source code")
	}
}

func TestIsPseudoModuleFile_Empty(t *testing.T) {
	t.Parallel()

	if IsPseudoModuleFile("") {
		t.Error("expected false for empty string")
	}
}

func TestIsPseudoModuleFile_MixedContent(t *testing.T) {
	t.Parallel()

	// Mostly Go code with one line that looks like a directive — should not
	// be classified as pseudo module file.
	code := `package main

func main() {
    module := "test"
}`

	if IsPseudoModuleFile(code) {
		t.Error("expected false for Go code with variable named 'module'")
	}
}
