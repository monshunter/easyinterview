package stub_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
)

func chatProfile() *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "practice.followup.default",
		Capability: aiclient.CapabilityChat,
		Default: aiclient.ProviderConfig{
			ProviderRef: stub.Name,
			Model:       "stub-chat-1",
		},
		TimeoutMs: 5000,
		Version:   "1.0.0",
	}
}

func chatPayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "user", Content: "deterministic input"},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	}
}

func TestStubFactoryAllowedInTestEnv(t *testing.T) {
	p, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New(test): %v", err)
	}
	if p.Name() != stub.Name {
		t.Fatalf("expected name=%q, got %q", stub.Name, p.Name())
	}
}

func TestStubFactoryRejectedOutsideTestEnv(t *testing.T) {
	if _, err := stub.New(stub.WithAppEnv("production")); !errors.Is(err, stub.ErrNotAllowed) {
		t.Fatalf("expected ErrNotAllowed, got %v", err)
	}
}

func TestStubFactoryAllowedWithExplicitOverride(t *testing.T) {
	if _, err := stub.New(stub.WithAppEnv("production"), stub.WithAllowed(true)); err != nil {
		t.Fatalf("stub.New(WithAllowed): %v", err)
	}
}

func TestStubCompleteIsDeterministic(t *testing.T) {
	p, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	a, _, err := p.Complete(context.Background(), chatProfile(), chatPayload())
	if err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	b, _, err := p.Complete(context.Background(), chatProfile(), chatPayload())
	if err != nil {
		t.Fatalf("second Complete: %v", err)
	}
	if a.Content != b.Content {
		t.Fatalf("expected deterministic stub output, got %q vs %q", a.Content, b.Content)
	}
}

func TestStubCompleteWithToolsIsDeterministic(t *testing.T) {
	p, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	payload := chatPayload()
	payload.Tools = []aiclient.Tool{{
		Name:        "extract_signal",
		Description: "Extract structured signal.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"signal":{"type":"string"}}}`),
	}}
	payload.ToolChoice = &aiclient.ToolChoice{Mode: aiclient.ToolChoiceModeTool, Name: "extract_signal"}

	a, metaA, err := p.Complete(context.Background(), chatProfile(), payload)
	if err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	b, metaB, err := p.Complete(context.Background(), chatProfile(), payload)
	if err != nil {
		t.Fatalf("second Complete: %v", err)
	}
	if a.Content != b.Content {
		t.Fatalf("expected deterministic tool stub output, got %q vs %q", a.Content, b.Content)
	}
	if len(a.ToolCalls) != 1 || len(b.ToolCalls) != 1 {
		t.Fatalf("expected deterministic stub tool call replay, got first=%+v second=%+v", a.ToolCalls, b.ToolCalls)
	}
	if a.ToolCalls[0].Name != "extract_signal" || b.ToolCalls[0].Name != "extract_signal" {
		t.Fatalf("stub tool call name mismatch: first=%+v second=%+v", a.ToolCalls[0], b.ToolCalls[0])
	}
	if string(a.ToolCalls[0].Arguments) != string(b.ToolCalls[0].Arguments) {
		t.Fatalf("expected deterministic tool arguments, got %s vs %s", a.ToolCalls[0].Arguments, b.ToolCalls[0].Arguments)
	}
	if strings.Contains(string(a.ToolCalls[0].Arguments), "deterministic input") {
		t.Fatalf("stub tool arguments leaked prompt plaintext: %s", a.ToolCalls[0].Arguments)
	}
	if len(metaA.ToolInvocations) != 1 || len(metaB.ToolInvocations) != 1 {
		t.Fatalf("expected tool invocation summaries, got first=%+v second=%+v", metaA.ToolInvocations, metaB.ToolInvocations)
	}
	if metaA.ToolInvocations[0].Name != "extract_signal" ||
		metaA.ToolInvocations[0].ArgumentsHash == "" ||
		metaA.ToolInvocations[0].ArgumentsLength != len(a.ToolCalls[0].Arguments) {
		t.Fatalf("tool invocation summary must contain name/hash/length only: %+v", metaA.ToolInvocations[0])
	}
	if metaA.ToolInvocations[0] != metaB.ToolInvocations[0] {
		t.Fatalf("expected deterministic tool invocation summary, got %+v vs %+v", metaA.ToolInvocations[0], metaB.ToolInvocations[0])
	}
}

func TestStubStreamEmitsDoneAndCloses(t *testing.T) {
	p, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	ch, err := p.Stream(context.Background(), chatProfile(), chatPayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	gotDone := false
	count := 0
	for ev := range ch {
		count++
		if ev.Type == aiclient.StreamEventDone {
			gotDone = true
			if ev.Meta == nil {
				t.Fatalf("expected meta on done event")
			}
		}
	}
	if !gotDone {
		t.Fatalf("expected at least one done event, saw %d events", count)
	}
}

func TestStubFactoryRejectsWhenNoAppEnvProvided(t *testing.T) {
	if _, err := stub.New(); !errors.Is(err, stub.ErrNotAllowed) {
		t.Fatalf("expected ErrNotAllowed when no AppEnv option set, got %v", err)
	}
}
