package aiclient_test

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// staticResolver is the simplest ProfileResolver: a name → *ModelProfile map.
// Production code uses the YAML loader instead (Phase 2).
type staticResolver map[string]*aiclient.ModelProfile

func (r staticResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	p, ok := r[name]
	if !ok {
		return nil, errors.New("profile not found: " + name)
	}
	return p, nil
}

func newTestClient(t *testing.T) *aiclient.Client {
	t.Helper()
	stubProv, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	resolver := staticResolver{
		"practice.followup.default": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default: aiclient.ProviderConfig{
				Provider: stub.Name,
				Model:    "stub-chat-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
		"review.embed.default": {
			Name:     "review.embed.default",
			TaskType: aiclient.TaskTypeEmbed,
			Default: aiclient.ProviderConfig{
				Provider: stub.Name,
				Model:    "stub-embed-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
	}
	c, err := aiclient.New(
		aiclient.Config{AppEnv: aiclient.AppEnvTest},
		aiclient.WithProfileResolver(resolver),
		aiclient.WithProvider(stubProv),
	)
	if err != nil {
		t.Fatalf("aiclient.New: %v", err)
	}
	return c
}

func samplePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "You are an interviewer."},
			{Role: "user", Content: "Tell me about a time you led a project."},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	}
}

func TestComplete_RoutesToStubAndReturnsMeta(t *testing.T) {
	c := newTestClient(t)
	resp, meta, err := c.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content == "" {
		t.Fatalf("expected non-empty content, got %q", resp.Content)
	}
	if meta.Provider != stub.Name {
		t.Fatalf("expected meta.Provider=%q, got %q", stub.Name, meta.Provider)
	}
	if meta.ModelProfileName != "practice.followup.default" {
		t.Fatalf("expected meta.ModelProfileName=practice.followup.default, got %q", meta.ModelProfileName)
	}
	if meta.ModelProfileVersion != "1.0.0" {
		t.Fatalf("expected meta.ModelProfileVersion=1.0.0, got %q", meta.ModelProfileVersion)
	}
	if meta.TaskType != aiclient.TaskTypeChat {
		t.Fatalf("expected meta.TaskType=chat, got %q", meta.TaskType)
	}
	if meta.PromptVersion != "p1" || meta.RubricVersion != "r1" || meta.Language != "en" {
		t.Fatalf("call metadata not propagated to meta: %+v", meta)
	}
	if meta.ValidationStatus != aiclient.ValidationStatusOK {
		t.Fatalf("expected ValidationStatusOK on success, got %q", meta.ValidationStatus)
	}
}

func TestComplete_DeterministicForSameInput(t *testing.T) {
	c := newTestClient(t)
	first, _, err := c.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("first Complete: %v", err)
	}
	second, _, err := c.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("second Complete: %v", err)
	}
	if first.Content != second.Content {
		t.Fatalf("expected deterministic output across calls, got %q vs %q", first.Content, second.Content)
	}
}

func TestComplete_EmptyMessagesReturnsAIOutputInvalid(t *testing.T) {
	c := newTestClient(t)
	_, meta, err := c.Complete(context.Background(), "practice.followup.default", aiclient.CompletePayload{})
	if err == nil {
		t.Fatalf("expected error for empty messages")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected error code %q, got %q", sharederrors.CodeAiOutputInvalid, apiErr.Code)
	}
	if meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected meta.ErrorCode=%q, got %q", sharederrors.CodeAiOutputInvalid, meta.ErrorCode)
	}
	if meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected ValidationStatusInvalid, got %q", meta.ValidationStatus)
	}
}

func TestEmbed_ReturnsVectors(t *testing.T) {
	c := newTestClient(t)
	resp, meta, err := c.Embed(context.Background(), "review.embed.default", aiclient.EmbedInput{
		Texts: []string{"hello", "world"},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "review.embed",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
		},
	})
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if len(resp.Vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(resp.Vectors))
	}
	if meta.TaskType != aiclient.TaskTypeEmbed {
		t.Fatalf("expected meta.TaskType=embed, got %q", meta.TaskType)
	}
}

func TestStream_DoneEventAndChannelClose(t *testing.T) {
	c := newTestClient(t)
	ch, err := c.Stream(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var events []aiclient.AIStreamEvent
	for ev := range ch {
		events = append(events, ev)
	}
	if len(events) == 0 {
		t.Fatalf("expected at least one event before channel close")
	}
	last := events[len(events)-1]
	if last.Type != aiclient.StreamEventDone {
		t.Fatalf("expected last event type 'done', got %q", last.Type)
	}
	if last.Meta == nil {
		t.Fatalf("expected done event to carry AICallMeta")
	}
	if last.Meta.Provider != stub.Name {
		t.Fatalf("expected done meta.Provider=%q, got %q", stub.Name, last.Meta.Provider)
	}
}

func TestNew_ProductionWithoutGatewayConfigFails(t *testing.T) {
	_, err := aiclient.New(aiclient.Config{AppEnv: "production"})
	if !errors.Is(err, aiclient.ErrMissingGatewayConfig) {
		t.Fatalf("expected ErrMissingGatewayConfig, got %v", err)
	}
}
