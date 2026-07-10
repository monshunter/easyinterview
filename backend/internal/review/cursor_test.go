package review

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestCursorEncodeDecodeRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 5, 15, 10, 20, 30, 123456789, time.UTC)
	id := "0197d120-0000-7000-8000-000000000501"

	cursor := EncodeCursor(createdAt, id)
	if cursor == "" {
		t.Fatal("EncodeCursor returned empty cursor")
	}
	if strings.Contains(cursor, "=") {
		t.Fatalf("cursor %q used padded encoding", cursor)
	}

	gotCreatedAt, gotID, err := DecodeCursor(cursor)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if !gotCreatedAt.Equal(createdAt) || gotID != id {
		t.Fatalf("decoded cursor = %s/%s, want %s/%s", gotCreatedAt, gotID, createdAt, id)
	}
}

func TestCursorRejectsTampered(t *testing.T) {
	for _, cursor := range []string{
		"",
		"not-base64url",
		base64.RawURLEncoding.EncodeToString([]byte(`{"createdAt":"2026-05-15T10:20:30Z","id":"0197d120-0000-7000-8000-000000000501","extra":true}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"createdAt":"2026-05-15T10:20:30+00:00","id":"0197d120-0000-7000-8000-000000000501"}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"createdAt":"2026-05-15T10:20:30Z","id":"not-a-uuid"}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"createdAt":"2026-05-15T10:20:30Z","id":"0197d120-0000-7000-8000-000000000501"} trailing`)),
	} {
		t.Run(cursor, func(t *testing.T) {
			if _, _, err := DecodeCursor(cursor); err == nil {
				t.Fatal("DecodeCursor succeeded, want ErrInvalidCursor")
			}
		})
	}
}

func TestCursorRejectsOutOfScopeFormat(t *testing.T) {
	outOfScope := base64.RawURLEncoding.EncodeToString([]byte("2026-05-15T10:20:30Z|0197d120-0000-7000-8000-000000000501"))
	if _, _, err := DecodeCursor(outOfScope); err == nil {
		t.Fatal("DecodeCursor accepted out-of-scope pipe-delimited cursor")
	}
}
