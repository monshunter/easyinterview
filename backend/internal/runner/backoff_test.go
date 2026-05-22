package runner

import (
	"testing"
	"time"
)

func TestBackoffPolicy_Next_Table(t *testing.T) {
	p := DefaultBackoffPolicy()
	want := []time.Duration{
		30 * time.Second,
		2 * time.Minute,
		10 * time.Minute,
		1 * time.Hour,
		6 * time.Hour,
	}
	for i, w := range want {
		attempts := int32(i + 1)
		if got := p.Next(attempts); got != w {
			t.Fatalf("Next(%d) = %s, want %s", attempts, got, w)
		}
	}
	if MaxAttempts != 5 {
		t.Fatalf("MaxAttempts = %d, want 5", MaxAttempts)
	}
}

func TestBackoffPolicy_BoundaryAttempts(t *testing.T) {
	p := DefaultBackoffPolicy()
	if got := p.Next(0); got != 30*time.Second {
		t.Fatalf("Next(0) = %s, want 30s (clamped to first)", got)
	}
	if got := p.Next(-3); got != 30*time.Second {
		t.Fatalf("Next(-3) = %s, want 30s (clamped to first)", got)
	}
	if got := p.Next(6); got != 6*time.Hour {
		t.Fatalf("Next(6) = %s, want 6h (clamped to last)", got)
	}
	if got := p.Next(100); got != 6*time.Hour {
		t.Fatalf("Next(100) = %s, want 6h (clamped to last)", got)
	}
}
