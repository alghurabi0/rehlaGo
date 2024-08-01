package main

import "testing"

func TestSubtract(t *testing.T) {
	n1 := 5
	n2 := 8

	res := subtract(n1, n2)

	if res != -3 {
		t.Errorf("want %d got %d", -3, res)
	}
}
