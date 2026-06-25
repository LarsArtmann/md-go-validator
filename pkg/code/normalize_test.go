package code

import (
	"testing"
)

func TestNormalizeDocIdioms_BodyOmitted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spaced body",
			input: "type Foo struct { ... }",
			want:  "type Foo struct {}",
		},
		{
			name:  "tight body",
			input: "type Foo struct{...}",
			want:  "type Foo struct{}",
		},
		{
			name:  "function body",
			input: "func DoSomething() error { ... }",
			want:  "func DoSomething() error {}",
		},
		{
			name:  "interface body",
			input: "type Reader interface { ... }",
			want:  "type Reader interface {}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NormalizeDocIdioms(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeDocIdioms(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeDocIdioms_EllipsisLine(t *testing.T) {
	t.Parallel()

	input := "func main() {\n\tresult := doThing()\n\t...\n\treturn result\n}"
	want := "func main() {\n\tresult := doThing()\n\treturn result\n}"

	got := NormalizeDocIdioms(input)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeDocIdioms_MultipleIdioms(t *testing.T) {
	t.Parallel()

	input := `type Service struct { ... }

func NewService(db *sql.DB, opts ...Option) *Service { ... }`
	want := `type Service struct {}

func NewService(db *sql.DB, opts ...Option) *Service {}`

	got := NormalizeDocIdioms(input)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeDocIdioms_NoChange(t *testing.T) {
	t.Parallel()

	input := "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"
	want := input

	got := NormalizeDocIdioms(input)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeDocIdioms_Empty(t *testing.T) {
	t.Parallel()

	got := NormalizeDocIdioms("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestNormalizeDocIdioms_CollapsesBlankLines(t *testing.T) {
	t.Parallel()

	input := "line1\n\n\n\nline2"
	want := "line1\n\nline2"

	got := NormalizeDocIdioms(input)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCollapseBlankLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single blank", "a\n\nb", "a\n\nb"},
		{"double blank", "a\n\n\nb", "a\n\nb"},
		{"triple blank", "a\n\n\n\nb", "a\n\nb"},
		{"trailing blanks", "a\n\n\n", "a"},
		{"no trailing newline", "a\nb", "a\nb"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := collapseBlankLines(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
