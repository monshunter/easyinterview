package idx

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateIdempotencyKey_Format(t *testing.T) {
	key := GenerateIdempotencyKey()
	parts := strings.Split(key, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 dot-separated parts, got %d in %q", len(parts), key)
	}
	if parts[0] != IdempotencyKeyVersion {
		t.Errorf("version = %q, want %q", parts[0], IdempotencyKeyVersion)
	}
}

func TestGenerateIdempotencyKey_Unique(t *testing.T) {
	a := GenerateIdempotencyKey()
	b := GenerateIdempotencyKey()
	if a == b {
		t.Fatalf("two generated keys collided: %q", a)
	}
}

func TestParseIdempotencyKey_RoundTrip(t *testing.T) {
	key := GenerateIdempotencyKey()
	parsed, err := ParseIdempotencyKey(key)
	if err != nil {
		t.Fatalf("ParseIdempotencyKey(%q): %v", key, err)
	}
	if parsed.UUID == "" {
		t.Errorf("UUID empty for key %q", key)
	}
	if parsed.IssuedAt.IsZero() {
		t.Errorf("IssuedAt zero for key %q", key)
	}
	if time.Since(parsed.IssuedAt) > 5*time.Second {
		t.Errorf("IssuedAt %v looks far in the past for freshly generated key", parsed.IssuedAt)
	}
}

func TestParseIdempotencyKey_Rejects(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"twoParts":    "v1.123",
		"badVersion":  "v0.123.0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
		"badUnix":     "v1.notanumber.0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
		"badUUID":     "v1.123.not-a-uuid",
		"v4UUID":      "v1.123.00000000-0000-4000-8000-000000000000",
		"tmpPrefixed": "v1.123.tmp_0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
	}
	for name, in := range cases {
		if _, err := ParseIdempotencyKey(in); err == nil {
			t.Errorf("%s: ParseIdempotencyKey(%q) returned nil; want error", name, in)
		}
	}
}

func TestIsExpired_FreshKey(t *testing.T) {
	key := GenerateIdempotencyKey()
	expired, err := IsIdempotencyKeyExpired(key, time.Now())
	if err != nil {
		t.Fatalf("IsIdempotencyKeyExpired: %v", err)
	}
	if expired {
		t.Errorf("freshly generated key was reported expired")
	}
}

func TestIsExpired_PastTTL(t *testing.T) {
	// Construct a key issued 25h ago by reusing the formatter.
	past := time.Now().Add(-25 * time.Hour)
	key := formatIdempotencyKey(past, NewID())
	expired, err := IsIdempotencyKeyExpired(key, time.Now())
	if err != nil {
		t.Fatalf("IsIdempotencyKeyExpired: %v", err)
	}
	if !expired {
		t.Errorf("key issued 25h ago not reported expired")
	}
}

func TestIdempotencyKey_TTLMatchesGeneratedConstant(t *testing.T) {
	if IdempotencyKeyTTLSeconds <= 0 {
		t.Fatalf("IdempotencyKeyTTLSeconds = %d, want positive", IdempotencyKeyTTLSeconds)
	}
	// Spec §3.4 documents 24h.
	if IdempotencyKeyTTLSeconds != 86400 {
		t.Errorf("IdempotencyKeyTTLSeconds = %d, want 86400 (24h)", IdempotencyKeyTTLSeconds)
	}
}
