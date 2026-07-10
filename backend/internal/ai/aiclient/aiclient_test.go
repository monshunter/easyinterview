package aiclient_test

import (
	"context"
	"encoding/json"
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

type countingProvider struct {
	inner               aiclient.Provider
	completeCalls       int
	transcribeCalls     int
	synthesizeCalls     int
	lastCompleteProfile string
	lastCompletePayload aiclient.CompletePayload
	lastTranscribeInput aiclient.TranscriptionInput
	lastSynthesizeInput aiclient.SynthesisInput
}

func (p *countingProvider) Name() string { return p.inner.Name() }

func (p *countingProvider) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	p.completeCalls++
	p.lastCompleteProfile = profile.Name
	p.lastCompletePayload = payload
	return p.inner.Complete(ctx, profile, payload)
}

func (p *countingProvider) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	p.transcribeCalls++
	p.lastTranscribeInput = input
	return p.inner.Transcribe(ctx, profile, input)
}

func (p *countingProvider) Synthesize(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	p.synthesizeCalls++
	p.lastSynthesizeInput = input
	return p.inner.Synthesize(ctx, profile, input)
}

func (p *countingProvider) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return p.inner.Stream(ctx, profile, payload)
}

type scriptedCompleteResult struct {
	content string
	err     error
}

type scriptedProvider struct {
	name            string
	completeResults []scriptedCompleteResult
	streamEvents    []aiclient.AIStreamEvent
	completeCalls   int
	transcribeCalls int
	synthesizeCalls int
}

func (p *scriptedProvider) Name() string { return p.name }

func (p *scriptedProvider) Complete(_ context.Context, profile *aiclient.ModelProfile, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	p.completeCalls++
	idx := p.completeCalls - 1
	if idx >= len(p.completeResults) {
		return aiclient.CompleteResponse{}, scriptedMeta(p.name, profile, nil), errors.New("unexpected complete call")
	}
	result := p.completeResults[idx]
	meta := scriptedMeta(p.name, profile, result.err)
	if result.err != nil {
		return aiclient.CompleteResponse{}, meta, result.err
	}
	return aiclient.CompleteResponse{Content: result.content, FinishReason: "stop"}, meta, nil
}

func (p *scriptedProvider) Transcribe(_ context.Context, profile *aiclient.ModelProfile, _ aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	p.transcribeCalls++
	return aiclient.TranscriptionResponse{Text: "scripted transcript"}, scriptedMeta(p.name, profile, nil), nil
}

func (p *scriptedProvider) Synthesize(_ context.Context, profile *aiclient.ModelProfile, _ aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	p.synthesizeCalls++
	return aiclient.SynthesisResponse{Audio: []byte("fake-audio"), ContentType: "audio/mpeg", DurationMs: 1000, CharCount: 5}, scriptedMeta(p.name, profile, nil), nil
}

func (p *scriptedProvider) Stream(_ context.Context, _ *aiclient.ModelProfile, _ aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	ch := make(chan aiclient.AIStreamEvent, len(p.streamEvents))
	for _, ev := range p.streamEvents {
		ch <- ev
	}
	close(ch)
	return ch, nil
}

func scriptedMeta(provider string, profile *aiclient.ModelProfile, err error) aiclient.AICallMeta {
	meta := aiclient.AICallMeta{
		Provider:     provider,
		ModelFamily:  "test-family",
		ModelID:      profile.Default.Model,
		InputTokens:  10,
		OutputTokens: 5,
		LatencyMs:    1,
	}
	if err != nil {
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
		meta.ErrorCode = testErrorCode(err)
	}
	return meta
}

func testErrorCode(err error) string {
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func assertAIOutputInvalid(t *testing.T, meta aiclient.AICallMeta, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected AI_OUTPUT_INVALID error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
	if meta.ErrorCode != sharederrors.CodeAiOutputInvalid || meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected invalid meta, got %+v", meta)
	}
}

func newTestClient(t *testing.T) *aiclient.Client {
	t.Helper()
	c, _ := newTestClientWithResolver(t, defaultResolver())
	return c
}

func defaultResolver() staticResolver {
	return staticResolver{
		"practice.followup.default": {
			Name:       "practice.followup.default",
			Capability: aiclient.CapabilityChat,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-chat-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
		"practice.voice.stt.default": {
			Name:       "practice.voice.stt.default",
			Capability: aiclient.CapabilitySTT,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-stt-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
		"practice.voice.tts.default": {
			Name:       "practice.voice.tts.default",
			Capability: aiclient.CapabilityTts,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-tts-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
		"practice.voice.realtime.default": {
			Name:              "practice.voice.realtime.default",
			Capability:        aiclient.CapabilityRealtime,
			Status:            aiclient.ProfileStatusUnsupported,
			UnsupportedReason: "realtime voice adapter remains fail-closed",
			Default: aiclient.ProviderConfig{
				ProviderRef: stub.Name,
				Model:       "stub-realtime-1",
			},
			TimeoutMs: 5000,
			Version:   "0.1.0",
		},
	}
}

func newTestClientWithResolver(t *testing.T, resolver staticResolver) (*aiclient.Client, *countingProvider) {
	t.Helper()
	stubProv, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	counting := &countingProvider{inner: stubProv}
	c, err := aiclient.New(
		aiclient.Config{AppEnv: aiclient.AppEnvTest},
		aiclient.WithStubAllowed(true),
		aiclient.WithProfileResolver(resolver),
		aiclient.WithProvider(counting),
	)
	if err != nil {
		t.Fatalf("aiclient.New: %v", err)
	}
	return c, counting
}

func newClientWithProviders(t *testing.T, resolver staticResolver, providers ...aiclient.Provider) *aiclient.Client {
	t.Helper()
	opts := []aiclient.Option{
		aiclient.WithStubAllowed(true),
		aiclient.WithProfileResolver(resolver),
	}
	for _, provider := range providers {
		opts = append(opts, aiclient.WithProvider(provider))
	}
	c, err := aiclient.New(aiclient.Config{AppEnv: aiclient.AppEnvTest}, opts...)
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
	if meta.Capability != aiclient.CapabilityChat {
		t.Fatalf("expected meta.Capability=chat, got %q", meta.Capability)
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

func TestComplete_ToolsPayloadRemainsProviderNeutral(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())
	payload := samplePayload()
	payload.Tools = []aiclient.Tool{{
		Name:        "extract_signal",
		Description: "Extract structured interview signal.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"signal":{"type":"string"}}}`),
	}}
	payload.ToolChoice = &aiclient.ToolChoice{
		Mode: aiclient.ToolChoiceModeTool,
		Name: "extract_signal",
	}

	_, meta, err := c.Complete(context.Background(), "practice.followup.default", payload)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if provider.lastCompleteProfile != "practice.followup.default" {
		t.Fatalf("provider should receive resolved profile name only, got %q", provider.lastCompleteProfile)
	}
	if len(provider.lastCompletePayload.Tools) != 1 {
		t.Fatalf("expected one provider-neutral tool, got %+v", provider.lastCompletePayload.Tools)
	}
	if provider.lastCompletePayload.Tools[0].Name != "extract_signal" {
		t.Fatalf("tool name not propagated: %+v", provider.lastCompletePayload.Tools[0])
	}
	if provider.lastCompletePayload.ToolChoice == nil || provider.lastCompletePayload.ToolChoice.Name != "extract_signal" {
		t.Fatalf("tool choice not propagated: %+v", provider.lastCompletePayload.ToolChoice)
	}
	if meta.ModelProfileName != "practice.followup.default" || meta.Provider != stub.Name {
		t.Fatalf("meta must stay profile/provider neutral, got %+v", meta)
	}
}

func TestComplete_EmptyMessagesReturnsAIOutputInvalid(t *testing.T) {
	c := newTestClient(t)
	_, meta, err := c.Complete(context.Background(), "practice.followup.default", aiclient.CompletePayload{})
	assertAIOutputInvalid(t, meta, err)
}

func TestTranscribe_RoutesSTTProfileThroughProvider(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())
	input := aiclient.TranscriptionInput{
		Audio:       []byte("fake-webm-bytes"),
		Filename:    "answer.webm",
		ContentType: "audio/webm",
		Language:    "en",
		Prompt:      "Interview answer",
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.voice.stt",
			PromptVersion: "stt-p1",
			Language:      "en",
		},
	}

	resp, meta, err := c.Transcribe(context.Background(), "practice.voice.stt.default", input)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if resp.Text == "" {
		t.Fatalf("expected non-empty transcript")
	}
	if provider.transcribeCalls != 1 {
		t.Fatalf("expected one provider transcription call, got %d", provider.transcribeCalls)
	}
	if string(provider.lastTranscribeInput.Audio) != string(input.Audio) || provider.lastTranscribeInput.Filename != "answer.webm" {
		t.Fatalf("transcription input not propagated: %+v", provider.lastTranscribeInput)
	}
	if meta.Capability != aiclient.CapabilitySTT || meta.ModelProfileName != "practice.voice.stt.default" {
		t.Fatalf("expected stt meta for profile, got %+v", meta)
	}
}

func TestTranscribe_RequiresAudioBytesFilenameAndContentType(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())

	_, meta, err := c.Transcribe(context.Background(), "practice.voice.stt.default", aiclient.TranscriptionInput{})
	assertAIOutputInvalid(t, meta, err)
	if provider.transcribeCalls != 0 {
		t.Fatalf("invalid transcription input must not reach provider, got %d calls", provider.transcribeCalls)
	}
}

func TestTranscribe_RealtimeProfileFailsClosed(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())

	_, meta, err := c.Transcribe(context.Background(), "practice.voice.realtime.default", aiclient.TranscriptionInput{
		Audio:       []byte("fake"),
		Filename:    "voice.webm",
		ContentType: "audio/webm",
	})
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilityRealtime)
	if provider.transcribeCalls != 0 {
		t.Fatalf("realtime profile must stay fail-closed for Transcribe, got %d calls", provider.transcribeCalls)
	}
}

func TestComplete_DisabledProfileFailsClosedWithSharedError(t *testing.T) {
	resolver := defaultResolver()
	resolver["practice.followup.default"].Status = aiclient.ProfileStatusDisabled
	resolver["practice.followup.default"].UnsupportedReason = "disabled until owner enables this capability"
	c, provider := newTestClientWithResolver(t, resolver)

	_, meta, err := c.Complete(context.Background(), "practice.followup.default", samplePayload())
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilityChat)
	if provider.completeCalls != 0 {
		t.Fatalf("disabled profile must fail before provider invocation, got %d calls", provider.completeCalls)
	}
}

func TestComplete_UnsupportedCapabilityFailsClosedWithSharedError(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())

	_, meta, err := c.Complete(context.Background(), "practice.voice.stt.default", samplePayload())
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilitySTT)
	if provider.completeCalls != 0 {
		t.Fatalf("unsupported capability must fail before provider invocation, got %d calls", provider.completeCalls)
	}
}

func TestComplete_ProfileCapabilityMismatchFailsClosedWithSharedError(t *testing.T) {
	resolver := defaultResolver()
	resolver["practice.followup.default"].Capability = aiclient.CapabilitySTT
	c, provider := newTestClientWithResolver(t, resolver)

	_, meta, err := c.Complete(context.Background(), "practice.followup.default", samplePayload())
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilitySTT)
	if provider.completeCalls != 0 {
		t.Fatalf("capability mismatch must fail before provider invocation, got %d calls", provider.completeCalls)
	}
}

func assertUnsupportedCapabilityError(t *testing.T, err error, meta aiclient.AICallMeta, capability aiclient.Capability) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected unsupported capability error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != sharederrors.CodeAiUnsupportedCapability {
		t.Fatalf("expected error code %q, got %q", sharederrors.CodeAiUnsupportedCapability, apiErr.Code)
	}
	if apiErr.Retryable {
		t.Fatalf("unsupported capability must not be retryable")
	}
	if meta.ErrorCode != sharederrors.CodeAiUnsupportedCapability {
		t.Fatalf("expected meta.ErrorCode=%q, got %q", sharederrors.CodeAiUnsupportedCapability, meta.ErrorCode)
	}
	if meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected ValidationStatusInvalid, got %q", meta.ValidationStatus)
	}
	if meta.Capability != capability {
		t.Fatalf("expected meta.Capability=%q, got %q", capability, meta.Capability)
	}
}

func TestComplete_CentralFallbackRetriesMatchingTimeout(t *testing.T) {
	primaryTimeout := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "primary timeout", true)
	primary := &scriptedProvider{name: "primary", completeResults: []scriptedCompleteResult{{err: primaryTimeout}}}
	fallback := &scriptedProvider{name: "fallback", completeResults: []scriptedCompleteResult{{content: "fallback ok"}}}
	profile := fallbackProfile(aiclient.CapabilityChat, []aiclient.FallbackEntry{{
		ProviderConfig: aiclient.ProviderConfig{ProviderRef: fallback.Name(), Model: "fallback-model"},
		When:           []string{"timeout"},
	}})
	c := newClientWithProviders(t, staticResolver{profile.Name: profile}, primary, fallback)

	resp, meta, err := c.Complete(context.Background(), profile.Name, samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content != "fallback ok" {
		t.Fatalf("expected fallback response, got %q", resp.Content)
	}
	if primary.completeCalls != 1 || fallback.completeCalls != 1 {
		t.Fatalf("expected primary and fallback once, got primary=%d fallback=%d", primary.completeCalls, fallback.completeCalls)
	}
	if meta.Provider != fallback.Name() || meta.ModelID != "fallback-model" {
		t.Fatalf("expected final provider/model from fallback, got %+v", meta)
	}
	wantChain := []string{"primary/primary-model", "fallback/fallback-model"}
	if !sameStrings(meta.FallbackChain, wantChain) {
		t.Fatalf("fallback chain mismatch: got %+v want %+v", meta.FallbackChain, wantChain)
	}
}

func TestComplete_FallbackConditionMissReturnsPrimaryError(t *testing.T) {
	primaryErr := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "bad output", false)
	primary := &scriptedProvider{name: "primary", completeResults: []scriptedCompleteResult{{err: primaryErr}}}
	fallback := &scriptedProvider{name: "fallback", completeResults: []scriptedCompleteResult{{content: "must not run"}}}
	profile := fallbackProfile(aiclient.CapabilityChat, []aiclient.FallbackEntry{{
		ProviderConfig: aiclient.ProviderConfig{ProviderRef: fallback.Name(), Model: "fallback-model"},
		When:           []string{"timeout"},
	}})
	c := newClientWithProviders(t, staticResolver{profile.Name: profile}, primary, fallback)

	_, meta, err := c.Complete(context.Background(), profile.Name, samplePayload())
	assertAPIErrorCode(t, err, sharederrors.CodeAiOutputInvalid, false)
	if fallback.completeCalls != 0 {
		t.Fatalf("fallback condition miss must not invoke fallback, got %d calls", fallback.completeCalls)
	}
	if meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected primary error in meta, got %+v", meta)
	}
}

func TestComplete_FallbackExhaustedReturnsSharedErrorWithChain(t *testing.T) {
	primaryTimeout := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "primary timeout", true)
	fallbackTimeout := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "fallback timeout", true)
	primary := &scriptedProvider{name: "primary", completeResults: []scriptedCompleteResult{{err: primaryTimeout}}}
	fallback := &scriptedProvider{name: "fallback", completeResults: []scriptedCompleteResult{{err: fallbackTimeout}}}
	profile := fallbackProfile(aiclient.CapabilityChat, []aiclient.FallbackEntry{{
		ProviderConfig: aiclient.ProviderConfig{ProviderRef: fallback.Name(), Model: "fallback-model"},
		When:           []string{"timeout"},
	}})
	c := newClientWithProviders(t, staticResolver{profile.Name: profile}, primary, fallback)

	_, meta, err := c.Complete(context.Background(), profile.Name, samplePayload())
	assertAPIErrorCode(t, err, sharederrors.CodeAiFallbackExhausted, true)
	wantChain := []string{"primary/primary-model", "fallback/fallback-model"}
	if !sameStrings(meta.FallbackChain, wantChain) {
		t.Fatalf("fallback chain mismatch: got %+v want %+v", meta.FallbackChain, wantChain)
	}
	if meta.ErrorCode != sharederrors.CodeAiFallbackExhausted {
		t.Fatalf("expected fallback exhausted in meta, got %+v", meta)
	}
}

func TestComplete_FallbackOverTwoHopsRejectedBeforeProviderCall(t *testing.T) {
	primary := &scriptedProvider{name: "primary", completeResults: []scriptedCompleteResult{{content: "must not run"}}}
	profile := fallbackProfile(aiclient.CapabilityChat, []aiclient.FallbackEntry{
		{ProviderConfig: aiclient.ProviderConfig{ProviderRef: "fb-1", Model: "m1"}},
		{ProviderConfig: aiclient.ProviderConfig{ProviderRef: "fb-2", Model: "m2"}},
		{ProviderConfig: aiclient.ProviderConfig{ProviderRef: "fb-3", Model: "m3"}},
	})
	c := newClientWithProviders(t, staticResolver{profile.Name: profile}, primary)

	_, meta, err := c.Complete(context.Background(), profile.Name, samplePayload())
	assertAPIErrorCode(t, err, sharederrors.CodeAiFallbackExhausted, true)
	if primary.completeCalls != 0 {
		t.Fatalf("fallback chain over max must fail before primary provider call, got %d calls", primary.completeCalls)
	}
	if meta.ErrorCode != sharederrors.CodeAiFallbackExhausted {
		t.Fatalf("expected meta fallback exhausted, got %+v", meta)
	}
}

func fallbackProfile(capability aiclient.Capability, fallback []aiclient.FallbackEntry) *aiclient.ModelProfile {
	return &aiclient.ModelProfile{
		Name:       "fallback.profile.default",
		Capability: capability,
		Status:     aiclient.ProfileStatusActive,
		Default: aiclient.ProviderConfig{
			ProviderRef: "primary",
			Model:       "primary-model",
		},
		Fallback:  fallback,
		TimeoutMs: 5000,
		Route:     "fallback.test",
		Version:   "1.0.0",
	}
}

func assertAPIErrorCode(t *testing.T, err error, code string, retryable bool) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %q", code)
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Code != code {
		t.Fatalf("expected error code %q, got %q", code, apiErr.Code)
	}
	if apiErr.Retryable != retryable {
		t.Fatalf("expected retryable=%v, got %v", retryable, apiErr.Retryable)
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
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
	if last.Meta.Capability != aiclient.CapabilityChat ||
		last.Meta.ModelProfileName != "practice.followup.default" ||
		last.Meta.ModelProfileVersion != "1.0.0" ||
		last.Meta.PromptVersion != "p1" ||
		last.Meta.RubricVersion != "r1" ||
		last.Meta.Language != "en" ||
		last.Meta.ValidationStatus != aiclient.ValidationStatusOK {
		t.Fatalf("stream done meta was not canonical merged: %+v", last.Meta)
	}
}

func TestStream_PartialDoneMetaIsCanonicalMerged(t *testing.T) {
	partial := aiclient.AICallMeta{
		Provider:          "primary",
		ModelFamily:       "test-family",
		ModelID:           "primary-model",
		InputTokens:       3,
		OutputTokens:      7,
		ValidationStatus:  aiclient.ValidationStatusInvalid,
		ErrorCode:         sharederrors.CodeAiProviderTimeout,
		PartialMetaReason: "context_cancelled",
	}
	provider := &scriptedProvider{
		name: "primary",
		streamEvents: []aiclient.AIStreamEvent{{
			Type: aiclient.StreamEventDone,
			Meta: &partial,
		}},
	}
	profile := fallbackProfile(aiclient.CapabilityChat, nil)
	c := newClientWithProviders(t, staticResolver{profile.Name: profile}, provider)

	ch, err := c.Stream(context.Background(), profile.Name, samplePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	events := collectClientStreamEvents(ch)
	if len(events) != 1 || events[0].Type != aiclient.StreamEventDone || events[0].Meta == nil {
		t.Fatalf("expected one done event, got %+v", events)
	}
	meta := events[0].Meta
	if meta.Capability != aiclient.CapabilityChat ||
		meta.ModelProfileName != profile.Name ||
		meta.ModelProfileVersion != "1.0.0" ||
		meta.PromptVersion != "p1" ||
		meta.RubricVersion != "r1" ||
		meta.Language != "en" {
		t.Fatalf("partial done meta missing canonical fields: %+v", meta)
	}
	if meta.ErrorCode != sharederrors.CodeAiProviderTimeout ||
		meta.ValidationStatus != aiclient.ValidationStatusInvalid ||
		meta.PartialMetaReason != "context_cancelled" ||
		meta.OutputTokens != 7 {
		t.Fatalf("partial done meta did not preserve provider fields: %+v", meta)
	}
}

func TestNew_ProductionWithoutProviderConfigFails(t *testing.T) {
	_, err := aiclient.New(aiclient.Config{AppEnv: "production"})
	if !errors.Is(err, aiclient.ErrMissingProviderConfig) {
		t.Fatalf("expected ErrMissingProviderConfig, got %v", err)
	}
}

func collectClientStreamEvents(ch <-chan aiclient.AIStreamEvent) []aiclient.AIStreamEvent {
	var events []aiclient.AIStreamEvent
	for ev := range ch {
		events = append(events, ev)
	}
	return events
}

func TestSynthesize_RoutesTTSProfileThroughProvider(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())
	input := aiclient.SynthesisInput{
		Text:         "你好，欢迎参加面试",
		Voice:        "zh_female_qingxin",
		Format:       "mp3",
		SpeakingRate: 1.0,
		Language:     "zh-CN",
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.voice.tts",
			PromptVersion: "tts-p1",
			Language:      "zh-CN",
		},
	}

	resp, meta, err := c.Synthesize(context.Background(), "practice.voice.tts.default", input)
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if len(resp.Audio) == 0 {
		t.Fatal("expected non-empty audio bytes")
	}
	if resp.ContentType == "" {
		t.Fatal("expected non-empty content type")
	}
	if provider.synthesizeCalls != 1 {
		t.Fatalf("expected one provider synthesizes call, got %d", provider.synthesizeCalls)
	}
	if provider.lastSynthesizeInput.Text != input.Text || provider.lastSynthesizeInput.Voice != "zh_female_qingxin" {
		t.Fatalf("synthesis input not propagated: %+v", provider.lastSynthesizeInput)
	}
	if meta.Capability != aiclient.CapabilityTts || meta.ModelProfileName != "practice.voice.tts.default" {
		t.Fatalf("expected tts meta for profile, got %+v", meta)
	}
}

func TestSynthesize_RequiresNonEmptyText(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())

	_, meta, err := c.Synthesize(context.Background(), "practice.voice.tts.default", aiclient.SynthesisInput{})
	assertAIOutputInvalid(t, meta, err)
	if provider.synthesizeCalls != 0 {
		t.Fatalf("invalid synthesize input must not reach provider, got %d calls", provider.synthesizeCalls)
	}
}

func TestSynthesize_UnsupportedCapabilityFailsClosedWithSharedError(t *testing.T) {
	c, provider := newTestClientWithResolver(t, defaultResolver())

	_, meta, err := c.Synthesize(context.Background(), "practice.voice.stt.default", aiclient.SynthesisInput{
		Text: "test",
	})
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilitySTT)
	if provider.synthesizeCalls != 0 {
		t.Fatalf("capability mismatch must fail before provider invocation, got %d calls", provider.synthesizeCalls)
	}
}

func TestSynthesize_DisabledProfileFailsClosedWithSharedError(t *testing.T) {
	resolver := defaultResolver()
	resolver["practice.voice.tts.default"].Status = aiclient.ProfileStatusDisabled
	resolver["practice.voice.tts.default"].UnsupportedReason = "disabled until TTS adapter lands"
	c, provider := newTestClientWithResolver(t, resolver)

	_, meta, err := c.Synthesize(context.Background(), "practice.voice.tts.default", aiclient.SynthesisInput{
		Text: "test",
	})
	assertUnsupportedCapabilityError(t, err, meta, aiclient.CapabilityTts)
	if provider.synthesizeCalls != 0 {
		t.Fatalf("disabled profile must fail before provider invocation, got %d calls", provider.synthesizeCalls)
	}
}
