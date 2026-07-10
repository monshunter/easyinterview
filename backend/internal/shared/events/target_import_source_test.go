package events

import (
	"errors"
	"testing"
)

// Phase 0.2 contract: backend-targetjob 001 plan D-13 / B3 spec D-13 lock the
// event-local TargetImportSourceType to {url, text, file}. The B2 API
// sourceType variant `manual_text` maps to event `text`; `manual_form` is the
// synchronous ready fallback path and MUST NOT emit `target.import.requested`.
// The helper enforces this boundary so business code cannot accidentally
// surface `manual_form` on the wire.

func TestMapAPISourceTypeToEvent_AllowedMappings(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want TargetImportSourceType
	}{
		{"url passes through", "url", TargetImportSourceTypeUrl},
		{"manual_text maps to text", "manual_text", TargetImportSourceTypeText},
		{"file passes through", "file", TargetImportSourceTypeFile},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MapAPISourceTypeToEvent(tc.in)
			if err != nil {
				t.Fatalf("MapAPISourceTypeToEvent(%q) returned error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("MapAPISourceTypeToEvent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestMapAPISourceTypeToEvent_RejectsManualForm(t *testing.T) {
	got, err := MapAPISourceTypeToEvent("manual_form")
	if err == nil {
		t.Fatalf("expected error for manual_form, got %q", got)
	}
	if !errors.Is(err, ErrManualFormNotEventSource) {
		t.Errorf("expected ErrManualFormNotEventSource, got %T %v", err, err)
	}
	if got != "" {
		t.Errorf("expected empty source type on error, got %q", got)
	}
}

func TestMapAPISourceTypeToEvent_RejectsUnknown(t *testing.T) {
	if _, err := MapAPISourceTypeToEvent(""); err == nil {
		t.Errorf("expected error for empty string")
	}
	if _, err := MapAPISourceTypeToEvent("out_of_scope_alias"); err == nil {
		t.Errorf("expected error for unknown variant")
	}
}

func TestTargetImportSourceType_NoManualFormConstant(t *testing.T) {
	allowed := map[TargetImportSourceType]bool{
		TargetImportSourceTypeUrl:  true,
		TargetImportSourceTypeText: true,
		TargetImportSourceTypeFile: true,
	}
	for _, forbidden := range []TargetImportSourceType{"manual_form", "manual_text"} {
		if allowed[forbidden] {
			t.Errorf("TargetImportSourceType must not include %q (B3 D-13 PII boundary)", forbidden)
		}
	}
	if got := len(allowed); got != 3 {
		t.Errorf("TargetImportSourceType allowed values = %d, want 3", got)
	}
}
