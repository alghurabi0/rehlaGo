package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	if actual != expected {
		t.Errorf("want %v got %v", actual, expected)
	}
}
