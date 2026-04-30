package events

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	fixtures := readEnvelopeFixtures(t)
	if got, want := len(fixtures), 3; got != want {
		t.Fatalf("fixture count = %d, want %d", got, want)
	}

	for _, fixture := range fixtures {
		encoded, err := json.Marshal(fixture)
		if err != nil {
			t.Fatalf("Marshal(%s): %v", fixture.EventName, err)
		}
		var decoded Envelope
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("Unmarshal(%s): %v", fixture.EventName, err)
		}
		if decoded.EventID != fixture.EventID ||
			decoded.EventName != fixture.EventName ||
			decoded.EventVersion != fixture.EventVersion ||
			decoded.AggregateType != fixture.AggregateType ||
			decoded.AggregateID != fixture.AggregateID ||
			decoded.Producer != fixture.Producer ||
			!decoded.OccurredAt.Equal(fixture.OccurredAt) {
			t.Fatalf("round-trip mismatch for %s: got %#v want %#v", fixture.EventName, decoded, fixture)
		}
		if (decoded.TraceID == nil) != (fixture.TraceID == nil) {
			t.Fatalf("traceId nil mismatch for %s", fixture.EventName)
		}
		if decoded.TraceID != nil && *decoded.TraceID != *fixture.TraceID {
			t.Fatalf("traceId mismatch for %s: got %q want %q", fixture.EventName, *decoded.TraceID, *fixture.TraceID)
		}
		if !jsonEqual(decoded.Payload, fixture.Payload) {
			t.Fatalf("payload mismatch for %s: got %s want %s", fixture.EventName, decoded.Payload, fixture.Payload)
		}
	}
}

func TestTraceIDSoftRequired(t *testing.T) {
	fixtures := readEnvelopeFixtures(t)
	missing := findEnvelope(t, fixtures, EventNameReportGenerated)
	warnings := missing.ValidateForPublish()
	if got, want := len(warnings), 1; got != want {
		t.Fatalf("missing traceId warnings = %d, want %d", got, want)
	}
	if warnings[0].Field != "traceId" || warnings[0].EventName != EventNameReportGenerated {
		t.Fatalf("unexpected warning: %#v", warnings[0])
	}

	present := findEnvelope(t, fixtures, EventNameTargetImportRequested)
	if present.TraceID == nil || *present.TraceID != "trace-target-import-1" {
		t.Fatalf("traceId not preserved: %#v", present.TraceID)
	}
	if warnings := present.ValidateForPublish(); len(warnings) != 0 {
		t.Fatalf("present traceId warnings = %#v, want none", warnings)
	}
}

func readEnvelopeFixtures(t *testing.T) []Envelope {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	path := filepath.Join(wd, "..", "..", "..", "..", "shared", "events", "__fixtures__", "envelopes.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixtures []Envelope
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	return fixtures
}

func findEnvelope(t *testing.T, fixtures []Envelope, name EventName) Envelope {
	t.Helper()
	for _, fixture := range fixtures {
		if fixture.EventName == name {
			return fixture
		}
	}
	t.Fatalf("missing fixture for %s", name)
	return Envelope{}
}

func jsonEqual(left, right json.RawMessage) bool {
	var leftValue any
	var rightValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return false
	}
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return false
	}
	leftBytes, _ := json.Marshal(leftValue)
	rightBytes, _ := json.Marshal(rightValue)
	return bytes.Equal(leftBytes, rightBytes)
}
