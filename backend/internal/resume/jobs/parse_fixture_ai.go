package jobs

import (
	"context"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// DeterministicParseAIClient returns a schema-valid resume.parse response in
// APP_ENV=test while delegating every other AI capability to the inner client.
type DeterministicParseAIClient struct {
	inner aiclient.AIClient
}

func NewDeterministicParseAIClient(inner aiclient.AIClient) *DeterministicParseAIClient {
	return &DeterministicParseAIClient{inner: inner}
}

func (c *DeterministicParseAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if payload.Metadata.FeatureKey == FeatureKeyResumeParse {
		return aiclient.CompleteResponse{Content: `{"displayName":"Fixture Candidate - Engineer","basics":{"name":"Fixture Candidate","headline":"Engineer"},"experiences":[{"company":"Fixture Co","title":"Engineer","start":"2024-01","end":"","summary":"Built interview preparation systems.","bullets":["Led reliable product delivery"]}],"projects":[],"education":[],"skills":["Go","React"],"languages":["en"]}`, FinishReason: "stop"}, aiclient.AICallMeta{
			Provider:         "stub",
			ModelFamily:      "stub",
			ModelID:          "resume-parse-fixture",
			FallbackChain:    []string{"stub/resume-parse-fixture"},
			ValidationStatus: aiclient.ValidationStatusOK,
		}, nil
	}
	return c.inner.Complete(ctx, profileName, payload)
}

func (c *DeterministicParseAIClient) Transcribe(ctx context.Context, profileName string, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return c.inner.Transcribe(ctx, profileName, input)
}

func (c *DeterministicParseAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return c.inner.Stream(ctx, profileName, payload)
}

func (c *DeterministicParseAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return c.inner.Synthesize(ctx, profileName, input)
}
