package stub_test

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
)

func chatProfile() *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:     "practice.followup.default",
		TaskType: aiclient.TaskTypeChat,
		Default: aiclient.ProviderConfig{
			Provider: stub.Name,
			Model:    "stub-chat-1",
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
