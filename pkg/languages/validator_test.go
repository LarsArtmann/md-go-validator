package languages

import (
	"context"
	"testing"
)

// MockValidator is a mock validator for testing.
type MockValidator struct {
	lang      Language
	available bool
}

func (m *MockValidator) Language() Language {
	return m.lang
}

func (m *MockValidator) Validate(_ context.Context, _ string) error {
	return nil
}

func (m *MockValidator) IsAvailable() bool {
	return m.available
}

// registerValidator registers a validator and fails the test if it errors.
func registerValidator(t *testing.T, r *Registry, v *MockValidator) {
	t.Helper()

	err := r.Register(v)
	if err != nil {
		t.Fatalf("register %s: %v", v.lang, err)
	}
}

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if len(r.Languages()) != 0 {
		t.Errorf("New registry should have no languages, got %d", len(r.Languages()))
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	t.Run("register valid validator", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		v := &MockValidator{lang: LangGo, available: true}

		err := r.Register(v)
		if err != nil {
			t.Errorf("Register() error = %v", err)
		}

		if got := r.Get(LangGo); got != v {
			t.Error("Get() did not return registered validator")
		}
	})

	t.Run("register nil validator", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()

		err := r.Register(nil)
		if err == nil {
			t.Error("Register(nil) should return error")
		}
	})

	t.Run("register invalid language", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		v := &MockValidator{lang: Language("invalid"), available: true}

		err := r.Register(v)
		if err == nil {
			t.Error("Register() with invalid language should return error")
		}
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	v := &MockValidator{lang: LangGo, available: true}
	registerValidator(t, r, v)

	testRegistryLookup(t, "Get", r, r.Get, []struct {
		name     string
		key      Language
		expected Validator
	}{
		{"registered", LangGo, v},
		{"not registered", LangTypeScript, nil},
	})
}

func TestRegistry_GetByString(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	v := &MockValidator{lang: LangGo, available: true}
	registerValidator(t, r, v)

	testRegistryLookup(t, "GetByString", r, r.GetByString, []struct {
		name     string
		key      string
		expected Validator
	}{
		{"go", "go", v},
		{"Go uppercase", "Go", v},
		{"unknown", "python", nil},
	})
}

func testRegistryLookup[T any](
	t *testing.T,
	testName string,
	_ *Registry,
	lookup func(T) Validator,
	tests []struct {
		name     string
		key      T
		expected Validator
	},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := lookup(tt.key); got != tt.expected {
				t.Errorf("%s() = %v, want %v", testName, got, tt.expected)
			}
		})
	}
}

func TestRegistry_GetAvailable(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	v1 := &MockValidator{lang: LangGo, available: true}
	v2 := &MockValidator{lang: LangTypeScript, available: false}

	registerValidator(t, r, v1)
	registerValidator(t, r, v2)

	available := r.GetAvailable()
	if len(available) != 1 {
		t.Errorf("GetAvailable() returned %d validators, want 1", len(available))
	}

	if available[0] != v1 {
		t.Error("GetAvailable() did not return the available validator")
	}
}

func TestRegistry_Languages(t *testing.T) {
	t.Parallel()

	r := NewRegistry()
	v1 := &MockValidator{lang: LangGo, available: true}
	v2 := &MockValidator{lang: LangTypeScript, available: true}

	registerValidator(t, r, v1)
	registerValidator(t, r, v2)

	langs := r.Languages()
	if len(langs) != 2 {
		t.Errorf("Languages() returned %d languages, want 2", len(langs))
	}
}

func TestRegistry_Validate(t *testing.T) {
	t.Parallel()

	t.Run("validate with registered validator", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		v := &MockValidator{lang: LangGo, available: true}
		registerValidator(t, r, v)

		ctx := context.Background()

		err := r.Validate(ctx, LangGo, "code")
		if err != nil {
			t.Errorf("Validate() error = %v", err)
		}
	})

	t.Run("validate unregistered language", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()

		ctx := context.Background()

		err := r.Validate(ctx, LangGo, "code")
		if err == nil {
			t.Error("Validate() with unregistered language should return error")
		}
	})

	t.Run("validate unavailable validator", func(t *testing.T) {
		t.Parallel()

		r := NewRegistry()
		v := &MockValidator{lang: LangTypeScript, available: false}
		registerValidator(t, r, v)

		ctx := context.Background()

		err := r.Validate(ctx, LangTypeScript, "code")
		if err == nil {
			t.Error("Validate() with unavailable validator should return error")
		}
	})
}

func TestDefaultRegistry(t *testing.T) {
	t.Parallel()

	r := DefaultRegistry()
	if r == nil {
		t.Fatal("DefaultRegistry() returned nil")
	}

	// Should have at least Go validator
	if got := r.Get(LangGo); got == nil {
		t.Error("DefaultRegistry() missing Go validator")
	}
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	testStringMethod(t, "Error", []struct {
		name     string
		value    ValidationError
		expected string
	}{
		{
			name: "with line and column",
			value: ValidationError{
				Message: "syntax error",
				Line:    10,
				Column:  5,
				Code:    ErrCodeSyntax,
			},
			expected: "10:5: syntax error",
		},
		{
			name: "without line and column",
			//nolint:exhaustruct // Intentionally testing partial initialization
			value:    ValidationError{Message: "syntax error"},
			expected: "syntax error",
		},
	}, func(e ValidationError) string { return e.Error() })
}
