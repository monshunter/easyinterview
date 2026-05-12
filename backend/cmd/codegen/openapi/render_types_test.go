package main

import "testing"

func TestTSTypeForNullableStringIncludesNull(t *testing.T) {
	got := tsTypeFor(map[string]any{
		"type":     "string",
		"nullable": true,
	})

	if want := "string | null"; got != want {
		t.Fatalf("tsTypeFor(nullable string) = %q, want %q", got, want)
	}
}

func TestGoTypeForNullableStringUsesPointer(t *testing.T) {
	got := goTypeFor(map[string]any{
		"type":     "string",
		"nullable": true,
	})

	if want := "*string"; got != want {
		t.Fatalf("goTypeFor(nullable string) = %q, want %q", got, want)
	}
}
