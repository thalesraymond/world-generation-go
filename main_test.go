package main

import "testing"

func TestGreet(t *testing.T) {
	got := greet()
	want := "world-generation-go is ready"

	if got != want {
		t.Fatalf("greet() = %q, want %q", got, want)
	}
}
