package languages

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestGoValidator_Language(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}

	if v.Language() != LangGo {
		t.Errorf("expected %s, got %s", LangGo, v.Language())
	}
}

func TestGoValidator_IsAvailable(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}

	if !v.IsAvailable() {
		t.Error("GoValidator should always be available")
	}
}

func TestGoValidator_Validate(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}
	ctx := context.Background()

	validCases := goValidatorValidCases()

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := v.Validate(ctx, tc.code)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func goValidatorValidCases() []struct {
	name string
	code string
} {
	return []struct {
		name string
		code string
	}{
		{"complete file", "package main\n\nfunc main() {}"},
		{"type declaration", "type User struct {\n\tName string\n}"},
		{"function signature", "func DoSomething() error"},
		{"import statement", "import \"fmt\""},
		{"variable declaration", "var x = 42"},
		{"expression", "x + y"},
		{"method declaration", "func (s *Server) Start() error"},
		{"interface", "type Reader interface {\n\tRead(p []byte) (n int, err error)\n}"},
		{"const block", "const (\n\tA = 1\n\tB = 2\n)"},
		// Elision idioms — { ... } → {}
		{"elided struct body", "type Foo struct { ... }"},
		{"elided func body", "func DoSomething() error { ... }"},
		{"elided interface body", "type Reader interface { ... }"},
		// Imports + statements (strategy 6 — the dominant docs pattern)
		{"imports and statements", "import \"fmt\"\n\nfmt.Println(\"hello\")"},
		{"multi-import and statements", "import (\n\t\"fmt\"\n\t\"os\"\n)\n\nfmt.Println(os.Args[0])"},
		{"import and assignment", "import \"strings\"\n\nresult := strings.ToUpper(\"hello\")"},
	}
}

func TestGoValidator_Validate_Invalid(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}
	ctx := context.Background()

	invalidCases := []struct {
		name string
		code string
	}{
		{"broken func", "func broken {"},
		{"invalid syntax", "func () {\n\treturn\n}("},
		{"missing bracket", "func main() {\n\tfmt.Println("},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := v.Validate(ctx, tc.code)
			if err == nil {
				t.Error("expected error for invalid syntax")
			}

			// Verify it's a ValidationError with syntax code
			var valErr *ValidationError
			if !errors.As(err, &valErr) {
				t.Errorf("expected ValidationError, got %T: %v", err, err)
			}

			if valErr.Code != ErrCodeSyntax {
				t.Errorf("expected ErrCodeSyntax, got %d", valErr.Code)
			}
		})
	}
}

func TestValidationError_WithCode(t *testing.T) {
	t.Parallel()

	original := &ValidationError{
		Message: "test error",
		Line:    10,
		Column:  5,
		Code:    ErrCodeUnknown,
	}

	withCode := original.WithCode(ErrCodeSyntax)

	if withCode.Code != ErrCodeSyntax {
		t.Errorf("expected code %d, got %d", ErrCodeSyntax, withCode.Code)
	}

	if withCode.Message != original.Message {
		t.Errorf("expected message %q, got %q", original.Message, withCode.Message)
	}

	if withCode.Line != original.Line {
		t.Errorf("expected line %d, got %d", original.Line, withCode.Line)
	}
}

func TestTreeSitterValidator_UnavailableLanguage(t *testing.T) {
	t.Parallel()

	// Use the test seam to inject a grammar name that no library recognizes,
	// since all valid Languages ship with embedded grammars.
	validator := newTreeSitterValidatorWithGrammarName(LangGo, "nonexistent_language_xyz")

	if validator.IsAvailable() {
		t.Skip("unexpected: nonexistent language is available")
	}

	err := validator.Validate(context.Background(), "code")
	if err == nil {
		t.Error("expected error for unavailable language")
	}
}

func TestValidationError_Error_NoPosition(t *testing.T) {
	t.Parallel()

	e := &ValidationError{Message: "simple error"}

	got := e.Error()
	if got != "simple error" {
		t.Errorf("expected %q, got %q", "simple error", got)
	}
}

func TestGoValidator_ErrorIncludesSkipDirectiveHint(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}
	ctx := context.Background()

	err := v.Validate(ctx, "func broken {")
	if err == nil {
		t.Fatal("expected error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	if !strings.Contains(valErr.Message, "skip-validate") {
		t.Errorf("error message should mention skip-validate, got: %s", valErr.Message)
	}
}

func TestGoValidator_ErrorIncludesMixedScopeHint(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}
	ctx := context.Background()

	// Mixed-scope: import (package-level) + bare statement (function-body)
	code := "import \"fmt\"\n\nresult := &&&broken"

	err := v.Validate(ctx, code)
	if err == nil {
		t.Fatal("expected error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	if !strings.Contains(valErr.Message, "mix") {
		t.Errorf("error message should mention mixed-scope, got: %s", valErr.Message)
	}
}

func TestGoValidator_BestAttemptReporting(t *testing.T) {
	t.Parallel()

	v := &GoValidator{}
	ctx := context.Background()

	// A snippet that fails all strategies but strategy 3 (func wrapper) gets
	// further than strategy 1 (raw) — the error should reference a later line.
	code := "func broken {"

	err := v.Validate(ctx, code)
	if err == nil {
		t.Fatal("expected error")
	}

	var valErr *ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}

	// The best-attempt error should have a line number > 0 (from a wrapped strategy).
	if valErr.Line == 0 {
		t.Errorf("expected non-zero line from best attempt, got line=0")
	}
}

func TestSplitImportsAndStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		code       string
		wantImport string
		wantRest   string
		wantOk     bool
	}{
		{
			name:       "single import + statement",
			code:       "import \"fmt\"\n\nfmt.Println(\"hi\")",
			wantImport: "import \"fmt\"",
			wantRest:   "\nfmt.Println(\"hi\")",
			wantOk:     true,
		},
		{
			name:       "multi-line import + statement",
			code:       "import (\n\t\"fmt\"\n\t\"os\"\n)\n\nfmt.Println(os.Args[0])",
			wantImport: "import (\n\t\"fmt\"\n\t\"os\"\n)",
			wantRest:   "\nfmt.Println(os.Args[0])",
			wantOk:     true,
		},
		{
			name:       "no imports",
			code:       "fmt.Println(\"hi\")",
			wantImport: "",
			wantRest:   "",
			wantOk:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			imp, rest, ok := splitImportsAndStatements(tc.code)
			if ok != tc.wantOk {
				t.Errorf("ok = %v, want %v", ok, tc.wantOk)

				return
			}

			if imp != tc.wantImport {
				t.Errorf("import = %q, want %q", imp, tc.wantImport)
			}

			if rest != tc.wantRest {
				t.Errorf("rest = %q, want %q", rest, tc.wantRest)
			}
		})
	}
}

func TestDetectMixedScopeHint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "import + statement",
			code: "import \"fmt\"\n\nfmt.Println(\"hi\")",
			want: "snippet mixes package-level declarations with function-body statements",
		},
		{
			name: "only declarations",
			code: "type Foo struct {\n\tName string\n}",
			want: "",
		},
		{
			name: "only statements",
			code: "x := 42\nfmt.Println(x)",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detectMixedScopeHint(tc.code)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLanguage_Extensions_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lang     Language
		expected []string
	}{
		{"Go", LangGo, []string{".go"}},
		{"Templ", LangTempl, []string{".templ"}},
		{"TypeScript", LangTypeScript, []string{".ts"}},
		{"TSX", LangTSX, []string{".tsx"}},
		{"Nix", LangNix, []string{".nix"}},
		{"Rust", LangRust, []string{".rs"}},
		{"HCL", LangHCL, []string{".hcl", ".tf", ".tfvars"}},
		{"Terraform", LangTerraform, []string{".hcl", ".tf", ".tfvars"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.lang.Extensions()
			assertExtensionsEqual(t, got, tt.expected)
		})
	}
}
