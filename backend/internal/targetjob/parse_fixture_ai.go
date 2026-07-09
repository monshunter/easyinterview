package targetjob

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// DeterministicParseAIClient wraps the A3 runtime only for APP_ENV=test.
// It returns a schema-valid target.import.parse fixture so cmd/api drainer
// paths can exercise the real TargetJob pipeline without relying on a live
// provider or the generic stub provider's non-JSON chat response.
type DeterministicParseAIClient struct {
	inner aiclient.AIClient
}

// NewDeterministicParseAIClient returns an AIClient that intercepts only
// target.import.parse Complete calls and delegates every other capability.
func NewDeterministicParseAIClient(inner aiclient.AIClient) *DeterministicParseAIClient {
	return &DeterministicParseAIClient{inner: inner}
}

// Complete implements aiclient.AIClient.
func (c *DeterministicParseAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if payload.Metadata.FeatureKey == FeatureKeyTargetImportParse {
		content := deterministicParseFixtureContent(payload.Metadata.Language)
		return aiclient.CompleteResponse{Content: content, FinishReason: "stop"}, aiclient.AICallMeta{
			Provider:         "targetjob-test-fixture",
			ModelFamily:      "fixture",
			ModelID:          "target-import-parse-fixture",
			ModelProfileName: profileName,
			Language:         payload.Metadata.Language,
			InputTokens:      len(payload.Messages),
			OutputTokens:     len(content),
			LatencyMs:        1,
		}, nil
	}
	if c == nil || c.inner == nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, errors.New("targetjob deterministic parse AI client has no delegate")
	}
	return c.inner.Complete(ctx, profileName, payload)
}

// Transcribe implements aiclient.AIClient.
func (c *DeterministicParseAIClient) Transcribe(ctx context.Context, profileName string, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	if c == nil || c.inner == nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("targetjob deterministic parse AI client has no delegate")
	}
	return c.inner.Transcribe(ctx, profileName, input)
}

// Stream implements aiclient.AIClient.
func (c *DeterministicParseAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	if c == nil || c.inner == nil {
		return nil, errors.New("targetjob deterministic parse AI client has no delegate")
	}
	return c.inner.Stream(ctx, profileName, payload)
}

// Synthesize implements aiclient.AIClient.
func (c *DeterministicParseAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	if c == nil || c.inner == nil {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("targetjob deterministic parse AI client has no delegate")
	}
	return c.inner.Synthesize(ctx, profileName, input)
}

func deterministicParseFixtureContent(language string) string {
	if language == "" {
		language = "en"
	}
	raw, _ := json.Marshal(parseAIResponse{
		Title:       "Backend Platform Engineer",
		CompanyName: "Acme",
		CoreThemes:  []string{"backend", "ownership"},
		InterviewRounds: []parseAIResponseRound{
			{
				Sequence:        1,
				Type:            "technical",
				Name:            "Backend architecture deep dive",
				DurationMinutes: 45,
				Focus:           "Discuss API design and async pipelines.",
			},
			{
				Sequence:        2,
				Type:            "manager",
				Name:            "Hiring manager ownership interview",
				DurationMinutes: 40,
				Focus:           "Assess service ownership, incident response, and cross-team collaboration.",
			},
		},
		Strengths:   []string{"Backend service experience"},
		Gaps:        []string{"Clarify production scale evidence"},
		RiskSignals: []string{"The JD implies production ownership without naming on-call or incident response expectations."},
		Requirements: []parseAIResponseReq{
			{
				Kind:          string(RequirementMustHave),
				Label:         "Backend service ownership",
				Description:   "Own APIs, persistence, and asynchronous job processing.",
				EvidenceLevel: string(EvidenceExplicit),
			},
			{
				Kind:          string(RequirementHiddenSignal),
				Label:         "Production ownership expectations",
				Description:   "The role likely screens for incident response and operational ownership evidence.",
				EvidenceLevel: string(EvidenceInferred),
			},
		},
	})
	return string(raw)
}
