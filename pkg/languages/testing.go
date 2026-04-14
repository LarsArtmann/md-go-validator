package languages

import "testing"

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
