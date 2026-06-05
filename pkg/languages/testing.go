package languages

import "testing"

func assertExtensionsEqual(t *testing.T, got, expected []string) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("expected %d extensions, got %d", len(expected), len(got))
	}

	for i := range got {
		if got[i] != expected[i] {
			t.Errorf("extension[%d] = %q, want %q", i, got[i], expected[i])
		}
	}
}

func testStringMethod[T any](t *testing.T, methodName string, cases []struct {
	name     string
	value    T
	expected string
}, getString func(T) string,
) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := getString(tt.value); got != tt.expected {
				t.Errorf("%s() = %v, want %v", methodName, got, tt.expected)
			}
		})
	}
}
