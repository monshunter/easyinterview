package observability_test

import (
	"context"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
)

// sensitiveTokens are unique substrings the test plants into the prompt
// and response. The decorator MUST NOT let any of them surface in metric
// labels, log fields, ai_task_runs rows, or audit_events metadata.
var sensitiveTokens = []string{
	"SSN-123-45-6789",
	"OFFER-LETTER-ACME",
	"INTERVIEWER-NAME-Alice",
	"COMPANY-PROPRIETARY-XYZ",
}

var sensitiveTTSTokens = []string{
	"TTS-PRIVATE-FEEDBACK",
	"VOICE-OUTPUT-SECRET",
}

func sensitivePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "Background: candidate " + sensitiveTokens[0] + " applied to " + sensitiveTokens[1] + "."},
			{Role: "user", Content: "Interviewer notes: " + sensitiveTokens[2] + " regarding " + sensitiveTokens[3]},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
			TaskRun: aiclient.AITaskRunContext{
				Capability:   aiclient.AITaskRunTaskFollowupGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   "018f0d59-0f7a-7b58-9f2f-65cc4d8e8b1d",
			},
		},
	}
}

func TestPrivacy_NoPlaintextLeaksAnywhere(t *testing.T) {
	wrap, registry, logger, runs, audit := newTestStack(t)

	if _, _, err := wrap.Complete(context.Background(), "practice.followup.default", sensitivePayload()); err != nil {
		t.Fatalf("Complete: %v", err)
	}

	assertNoPlaintextLeaks(t, sensitiveTokens, registry, logger, runs, audit)

	// Sanity check: hashes must be present and non-empty (the decorator
	// is not silently dropping the audit row).
	rows := audit.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(rows))
	}
	if rows[0].Metadata.PromptHash == "" || rows[0].Metadata.ResponseHash == "" {
		t.Fatalf("audit hashes missing: %+v", rows[0].Metadata)
	}
	if rows[0].Metadata.PromptCharLength == 0 || rows[0].Metadata.ResponseCharLength == 0 {
		t.Fatalf("audit char lengths zero: %+v", rows[0].Metadata)
	}
}

func TestPrivacy_SynthesizeNoPlaintextLeaksAndUsesTTSLabels(t *testing.T) {
	registry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	runs := &memTaskRunWriter{}
	audit := &memAuditWriter{}
	resolver := staticResolver{
		"practice.voice.tts.default": {
			Name:       "practice.voice.tts.default",
			Capability: aiclient.CapabilityTts,
			Status:     aiclient.ProfileStatusActive,
			Default: aiclient.ProviderConfig{
				ProviderRef: "tts-provider",
				Model:       "speech-model",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
			Route:     "practice.voice.tts",
		},
	}
	wrap, err := observability.New(&sensitiveSynthesisInner{},
		observability.WithRegisterer(registry),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runs),
		observability.WithAuditEventWriter(audit),
		observability.WithProfileResolver(resolver),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	input := sampleSynthesisInput()
	input.Text = "spoken coaching note " + sensitiveTTSTokens[0]

	_, meta, err := wrap.Synthesize(context.Background(), "practice.voice.tts.default", input)
	if err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	if meta.Capability != aiclient.CapabilityTts {
		t.Fatalf("expected capability=%q, got %q", aiclient.CapabilityTts, meta.Capability)
	}

	successLabels := []string{"tts-provider", "speech-family", "practice.voice.tts.default", "practice.voice.tts", string(aiclient.CapabilityTts), "en", "success"}
	if got := registry.CounterValue(observability.MetricRunsTotal, successLabels...); got != 1 {
		t.Fatalf("expected tts run counter=1, got %v", got)
	}

	assertNoPlaintextLeaks(t, sensitiveTTSTokens, registry, logger, runs, audit)

	auditRows := audit.Rows()
	if len(auditRows) != 1 {
		t.Fatalf("expected 1 audit row, got %d", len(auditRows))
	}
	if auditRows[0].Metadata.PromptHash == "" || auditRows[0].Metadata.ResponseHash == "" {
		t.Fatalf("audit hashes missing: %+v", auditRows[0].Metadata)
	}
	if auditRows[0].Metadata.PromptCharLength != len(input.Text) {
		t.Fatalf("expected prompt length=%d, got %+v", len(input.Text), auditRows[0].Metadata)
	}
}

func assertNoPlaintextLeaks(
	t *testing.T,
	tokens []string,
	registry *observability.InMemoryRegistry,
	logger *observability.MemoryLogger,
	runs *memTaskRunWriter,
	audit *memAuditWriter,
) {
	t.Helper()
	metricFamilies := []string{
		observability.MetricRunsTotal,
		observability.MetricInputTokensTotal,
		observability.MetricOutputTokensTotal,
		observability.MetricCostUSDTotal,
		observability.MetricOutputValidationFailures,
		observability.MetricFallbackTotal,
	}
	for _, token := range tokens {
		for _, family := range metricFamilies {
			for _, labels := range registry.CounterLabelValues(family) {
				for _, label := range labels {
					if strings.Contains(label, token) {
						t.Errorf("plaintext token %q leaked into metric %q label %q", token, family, label)
					}
				}
			}
		}
		for _, entry := range logger.Entries() {
			if anyContains(entry.Fields, token) {
				t.Errorf("plaintext token %q leaked into log fields: %+v", token, entry.Fields)
			}
		}
		for _, row := range runs.Rows() {
			if anyTaskRunContains(row, token) {
				t.Errorf("plaintext token %q leaked into ai_task_runs row: %+v", token, row)
			}
		}
		for _, row := range audit.Rows() {
			meta := row.Metadata
			if strings.Contains(meta.PromptHash, token) ||
				strings.Contains(meta.ResponseHash, token) ||
				strings.Contains(meta.ProfileName, token) {
				t.Errorf("plaintext token %q leaked into audit metadata: %+v", token, meta)
			}
		}
	}
}

type sensitiveSynthesisInner struct{}

func (s *sensitiveSynthesisInner) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, nil
}

func (s *sensitiveSynthesisInner) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, nil
}

func (s *sensitiveSynthesisInner) Synthesize(_ context.Context, _ string, _ aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{
			Audio:       []byte("rendered " + sensitiveTTSTokens[1]),
			ContentType: "audio/mpeg",
			DurationMs:  320,
			CharCount:   24,
		}, aiclient.AICallMeta{
			Provider:            "tts-provider",
			ModelFamily:         "speech-family",
			ModelID:             "speech-model",
			Capability:          aiclient.CapabilityTts,
			ModelProfileName:    "practice.voice.tts.default",
			ModelProfileVersion: "1.0.0",
			Language:            "en",
			InputTokens:         24,
			OutputTokens:        320,
			LatencyMs:           1,
			ValidationStatus:    aiclient.ValidationStatusOK,
			Route:               "practice.voice.tts",
		}, nil
}

func (s *sensitiveSynthesisInner) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	ch := make(chan aiclient.AIStreamEvent)
	close(ch)
	return ch, nil
}

func anyContains(fields observability.LogFields, token string) bool {
	candidates := []string{
		fields.Provider, fields.ModelID, fields.ModelProfileName, fields.ModelProfileVersion,
		fields.PromptVersion, fields.RubricVersion, fields.Capability, fields.Language,
		fields.Route, fields.ValidationStatus, fields.ErrorCode,
	}
	for _, c := range candidates {
		if strings.Contains(c, token) {
			return true
		}
	}
	for _, fb := range fields.FallbackChain {
		if strings.Contains(fb, token) {
			return true
		}
	}
	return false
}

func anyTaskRunContains(row aiclient.AITaskRunRow, token string) bool {
	candidates := []string{
		row.Provider, row.ModelFamily, row.ModelID, string(row.Capability),
		row.PromptVersion, row.RubricVersion, row.ModelProfileName, row.ModelProfileVersion,
		row.Language, row.Route, string(row.ValidationStatus), row.ErrorCode,
		row.Metadata.PromptHash, row.Metadata.ResponseHash, row.Metadata.ProfileName,
	}
	for _, c := range candidates {
		if strings.Contains(c, token) {
			return true
		}
	}
	for _, fb := range row.FallbackChain {
		if strings.Contains(fb, token) {
			return true
		}
	}
	return false
}
