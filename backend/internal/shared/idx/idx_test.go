package idx

import (
	"regexp"
	"strings"
	"testing"
)

func TestNewID_ReturnsValidUUIDv7(t *testing.T) {
	pattern := regexp.MustCompile(UUIDv7RegexExpr)
	id := NewID()
	if !pattern.MatchString(id) {
		t.Fatalf("NewID() = %q, does not match UUIDv7 regex", id)
	}
}

func TestNewID_Unique(t *testing.T) {
	a := NewID()
	b := NewID()
	if a == b {
		t.Fatalf("NewID() returned the same value twice: %q", a)
	}
}

func TestRequireServerID_AcceptsValidUUIDv7(t *testing.T) {
	if err := RequireServerID(SampleUUIDv7); err != nil {
		t.Fatalf("RequireServerID(sample) returned error: %v", err)
	}
	id := NewID()
	if err := RequireServerID(id); err != nil {
		t.Fatalf("RequireServerID(NewID()) returned error: %v", err)
	}
}

func TestRequireServerID_RejectsTmpPrefix(t *testing.T) {
	cases := []string{
		"tmp_abc",
		"tmp_0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
		TmpIDPrefix + "anything",
	}
	for _, in := range cases {
		err := RequireServerID(in)
		if err == nil {
			t.Errorf("RequireServerID(%q) returned nil; want error", in)
			continue
		}
		if !strings.Contains(err.Error(), "tmp_") {
			t.Errorf("RequireServerID(%q) error = %v; expected message mentioning tmp_", in, err)
		}
	}
}

func TestRequireServerID_RejectsEmpty(t *testing.T) {
	if err := RequireServerID(""); err == nil {
		t.Fatal("RequireServerID(\"\") returned nil; want error")
	}
}

func TestRequireServerID_RejectsInvalidUUID(t *testing.T) {
	cases := []string{
		"not-a-uuid",
		"0195f2d0",
		"00000000-0000-0000-0000-000000000000", // valid UUID but not v7 (version digit 0)
	}
	for _, in := range cases {
		if err := RequireServerID(in); err == nil {
			t.Errorf("RequireServerID(%q) returned nil; want error", in)
		}
	}
}
