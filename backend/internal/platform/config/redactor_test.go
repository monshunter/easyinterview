package config_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

const plaintext = "ultra-secret-value"

func TestRedactedStringFmtPaths(t *testing.T) {
	rs := config.NewRedactedString(plaintext)

	cases := map[string]string{
		"%s":  fmt.Sprintf("%s", rs),
		"%v":  fmt.Sprintf("%v", rs),
		"%q":  fmt.Sprintf("%q", rs),
		"%+v": fmt.Sprintf("%+v", rs),
		"%#v": fmt.Sprintf("%#v", rs),
		"Println": func() string {
			var b strings.Builder
			fmt.Fprintln(&b, rs)
			return b.String()
		}(),
		"String": rs.String(),
		"GoString": rs.GoString(),
	}
	for name, got := range cases {
		if strings.Contains(got, plaintext) {
			t.Errorf("[%s] leaked plaintext: %q", name, got)
		}
		if !strings.Contains(got, "***") {
			t.Errorf("[%s] missing redaction marker: %q", name, got)
		}
	}
}

func TestRedactedStringJSON(t *testing.T) {
	rs := config.NewRedactedString(plaintext)
	out, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	if strings.Contains(string(out), plaintext) {
		t.Errorf("JSON leaked plaintext: %s", out)
	}
	if string(out) != `"***"` {
		t.Errorf("JSON not redacted: %s", out)
	}

	type wrapper struct {
		Secret config.RedactedString `json:"secret"`
		Public string                `json:"public"`
	}
	w := wrapper{Secret: rs, Public: "ok"}
	out, err = json.Marshal(w)
	if err != nil {
		t.Fatalf("MarshalJSON nested: %v", err)
	}
	if strings.Contains(string(out), plaintext) {
		t.Errorf("nested JSON leaked plaintext: %s", out)
	}

	mt, err := rs.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(mt) != "***" {
		t.Errorf("MarshalText not redacted: %s", mt)
	}
}

func TestRedactedStringErrorWrapping(t *testing.T) {
	rs := config.NewRedactedString(plaintext)
	wrapped := fmt.Errorf("loading secret: %w", fmt.Errorf("value=%s", rs))
	if strings.Contains(wrapped.Error(), plaintext) {
		t.Errorf("error wrapping leaked plaintext: %v", wrapped)
	}
	if !errors.Is(wrapped, wrapped) {
		t.Errorf("errors.Is sanity")
	}
}

func TestRedactedStringRevealOnly(t *testing.T) {
	rs := config.NewRedactedString(plaintext)
	if rs.Reveal() != plaintext {
		t.Errorf("Reveal mismatch")
	}
}
