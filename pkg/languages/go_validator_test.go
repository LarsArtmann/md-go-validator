package languages

import (
	"context"
	"errors"
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

	validator := NewTreeSitterValidator(LangGo, "nonexistent_language_xyz")

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
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d extensions, got %d", len(tt.expected), len(got))
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("extension[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
