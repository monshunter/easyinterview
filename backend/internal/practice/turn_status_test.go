package practice

import "testing"

func TestTurnStatusRoundTripFiveWireValues(t *testing.T) {
	values := []TurnStatus{
		TurnStatusAsked,
		TurnStatusAnswered,
		TurnStatusFollowUpRequested,
		TurnStatusAssessed,
		TurnStatusSkipped,
	}
	for _, status := range values {
		t.Run(string(status), func(t *testing.T) {
			fromDB, err := ParseTurnStatus(string(status))
			if err != nil {
				t.Fatalf("ParseTurnStatus: %v", err)
			}
			if fromDB != status {
				t.Fatalf("ParseTurnStatus = %s, want %s", fromDB, status)
			}
			wire, err := status.WireValue()
			if err != nil {
				t.Fatalf("WireValue: %v", err)
			}
			if wire != string(status) {
				t.Fatalf("WireValue = %s, want %s", wire, status)
			}
			fromWire, err := ParseWireTurnStatus(wire)
			if err != nil {
				t.Fatalf("ParseWireTurnStatus: %v", err)
			}
			if fromWire != status {
				t.Fatalf("ParseWireTurnStatus = %s, want %s", fromWire, status)
			}
		})
	}
}

func TestTurnStatusRejectsUnknownAndDoesNotCompressRuntimeValues(t *testing.T) {
	for _, raw := range []string{"", "followup_requested", "done", "unknown"} {
		if _, err := ParseTurnStatus(raw); err == nil {
			t.Fatalf("ParseTurnStatus(%q) expected error", raw)
		}
		if _, err := ParseWireTurnStatus(raw); err == nil {
			t.Fatalf("ParseWireTurnStatus(%q) expected error", raw)
		}
	}

	if wire, err := TurnStatusFollowUpRequested.WireValue(); err != nil || wire != "follow_up_requested" {
		t.Fatalf("follow_up_requested must stay visible on wire, got %q err=%v", wire, err)
	}
	if wire, err := TurnStatusAssessed.WireValue(); err != nil || wire != "assessed" {
		t.Fatalf("assessed must stay visible on wire, got %q err=%v", wire, err)
	}
}
