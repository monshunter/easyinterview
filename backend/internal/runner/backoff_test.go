package runner

import (
	"testing"
	"time"
)

func TestBackoffPolicy_BusinessJobSchedule(t *testing.T) {
	p := DefaultBackoffPolicy()
	want := []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second, 80 * time.Second}
	for index, delay := range want {
		attempts := int32(index + 1)
		if got := p.Next(attempts); got != delay {
			t.Fatalf("Next(%d)=%s want=%s", attempts, got, delay)
		}
	}
	if got := p.Next(5); got != 80*time.Second {
		t.Fatalf("business schedule cap=%s want=80s", got)
	}
}

func TestBackoffPolicy_OutboxScheduleRemainsInfrastructurePolicy(t *testing.T) {
	p := DefaultOutboxBackoffPolicy()
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
	if got := p.Next(0); got != 10*time.Second {
		t.Fatalf("Next(0) = %s, want 10s (clamped to first)", got)
	}
	if got := p.Next(-3); got != 10*time.Second {
		t.Fatalf("Next(-3) = %s, want 10s (clamped to first)", got)
	}
	if got := p.Next(6); got != 80*time.Second {
		t.Fatalf("Next(6) = %s, want 80s (clamped to last)", got)
	}
	if got := p.Next(100); got != 80*time.Second {
		t.Fatalf("Next(100) = %s, want 80s (clamped to last)", got)
	}
}
