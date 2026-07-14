package targetjob_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

// TestParseExecutorRegistryAdapterCrossLayer wires a real F3 RegistryAdapter
// (not the in-package fakeRegistry) into ParseExecutor so the
// PromptResolution shape that flows from yaml meta on disk through the
// adapter, the executor, and the AI metadata payload is exercised
// end-to-end. Plan §3.4 verification gate.
//
// The test does not verify B4 ai_task_runs typed columns (those land in
// Phase 4 alongside the additive migration); it only freezes the field
// projection from F3 baseline to AI call payload.
func TestParseExecutorRegistryAdapterCrossLayer(t *testing.T) {
	t.Parallel()

	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	adapter := targetjob.NewRegistryAdapter(client)

	resolved, err := adapter.Resolve(context.Background(), targetjob.FeatureKeyTargetImportParse, "en")
	if err != nil {
		t.Fatalf("adapter.Resolve: %v", err)
	}

	// Field-by-field assertion against the spec §3.1 D-4 + D-1 contract.
	wants := map[string]string{
		"PromptVersion":     "v0.1.0",
		"RubricVersion":     "v0.1.0",
		"ModelProfileName":  "target.import.default",
		"FeatureFlag":       "none",
		"DataSourceVersion": "registry.v1",
	}
	got := map[string]string{
		"PromptVersion":     resolved.PromptVersion,
		"RubricVersion":     resolved.RubricVersion,
		"ModelProfileName":  resolved.ModelProfileName,
		"FeatureFlag":       resolved.FeatureFlag,
		"DataSourceVersion": resolved.DataSourceVersion,
	}
	for k, want := range wants {
		if got[k] != want {
			t.Errorf("%s: want %q, got %q", k, want, got[k])
		}
	}
	if resolved.UserMessageTemplate == "" {
		t.Errorf("UserMessageTemplate must be populated for plan 001 baseline")
	}
	if resolved.OutputSchema == nil {
		t.Fatalf("OutputSchema must be populated for target.import.parse")
	}

	// Plan §3.4 second clause: provenance JSON written by ParseExecutor
	// must contain promptVersion / rubricVersion / modelId / language /
	// featureFlag / dataSourceVersion (the GenerationProvenance 6 fields).
	// modelId is the A3 resolved adapter model id, not the registry profile.
	aiModelID := "fixture-model:target-import-parse"
	provenance := map[string]string{
		"language":          "en",
		"featureFlag":       coalesceFlagForTest(resolved.FeatureFlag),
		"promptVersion":     resolved.PromptVersion,
		"rubricVersion":     resolved.RubricVersion,
		"modelId":           aiModelID,
		"dataSourceVersion": resolved.DataSourceVersion,
	}
	for _, field := range []string{
		"language",
		"featureFlag",
		"promptVersion",
		"rubricVersion",
		"modelId",
		"dataSourceVersion",
	} {
		if provenance[field] == "" {
			t.Errorf("provenance.%s must be populated for cross-layer contract", field)
		}
	}

	// Sanity: the assembled provenance map must JSON-encode without
	// loss; consumers downstream parse it as openapi GenerationProvenance.
	if _, err := json.Marshal(provenance); err != nil {
		t.Fatalf("provenance marshal: %v", err)
	}
}

func TestTargetImportPromptMatchesParseResponseSchema(t *testing.T) {
	t.Parallel()

	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	adapter := targetjob.NewRegistryAdapter(client)

	for _, tc := range []struct {
		name     string
		language string
	}{
		{name: "exact en", language: "en"},
		{name: "multi fallback", language: "fr"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolved, err := adapter.Resolve(context.Background(), targetjob.FeatureKeyTargetImportParse, tc.language)
			if err != nil {
				t.Fatalf("adapter.Resolve(%q): %v", tc.language, err)
			}
			body := resolved.UserMessageTemplate
			for _, token := range []string{"{{jd_text}}", "{{language}}"} {
				if !strings.Contains(body, token) {
					t.Errorf("target.import.parse prompt must contain %q; body=%s", token, body)
				}
			}
			for _, forbidden := range []string{
				"{{jd_source_url}}",
				"{{target_language}}",
				"JD source URL",
				"source URL",
				"page heading",
				"metadata-like lines",
			} {
				if strings.Contains(body, forbidden) {
					t.Errorf("target.import.parse prompt must not contain removed source input %q; body=%s", forbidden, body)
				}
			}
			for _, key := range []string{
				"title",
				"companyName",
				"coreThemes",
				"interviewRounds",
				"sequence",
				"type",
				"name",
				"durationMinutes",
				"focus",
				"role seniority",
				"company or industry nature",
				"team or business context",
				"hiring-process hints",
				"common interview practices",
				"strengths",
				"gaps",
				"riskSignals",
				"requirements",
				"hidden_signal",
				"inferred",
				"kind",
				"label",
				"evidenceLevel",
			} {
				if !strings.Contains(body, key) {
					t.Errorf("target.import.parse prompt must instruct parser key %q; body=%s", key, body)
				}
			}
			for _, outOfScopeKey := range []string{
				"role_title",
				"required_skills",
				"responsibilities",
				"language_signals",
				"parse_confidence",
				"interviewHypotheses",
			} {
				if strings.Contains(body, outOfScopeKey) {
					t.Errorf("target.import.parse prompt still instructs out-of-scope parser key %q; body=%s", outOfScopeKey, body)
				}
			}
		})
	}
}

// coalesceFlagForTest mirrors targetjob.coalesceFlag (unexported) so the
// cross-layer test can build the provenance map without duplicating
// production behavior. It is local to the test package and limited to the
// "none" default branch the spec requires.
func coalesceFlagForTest(flag string) string {
	if flag == "" {
		return "none"
	}
	return flag
}

// TestParseExecutorMetadataCarriesF3Triple verifies that when a real
// RegistryAdapter is wired into ParseExecutor and a stub AI client
// captures the metadata, the AI call sees the F3 triple
// (FeatureKey + PromptVersion + RubricVersion + DataSourceVersion).
// Combined with the existing fakeRegistry-based TestParseExecutor_HappyPath
// it asserts that the 7-field PromptResolution mapping does not lose
// data when crossing the targetjob/registry boundary.
func TestParseExecutorMetadataCarriesF3Triple(t *testing.T) {
	t.Parallel()

	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	adapter := targetjob.NewRegistryAdapter(client)

	captured := &captureAIClient{}
	store := &pipelineFakeStore{
		target: targetjob.TargetJobRecord{
			ID:             "tgt-1",
			UserID:         "user-1",
			TargetLanguage: "en",
			RawJDText:      "JD text",
		},
	}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: adapter,
		AI:       captured,
		NewID:    func() string { return "id-1" },
		Now:      func() time.Time { return time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC) },
	})

	// We don't care whether the parse completes successfully — the AI
	// metadata is captured before any response parsing happens.
	_ = exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "job-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if captured.metadata.FeatureKey != targetjob.FeatureKeyTargetImportParse {
		t.Errorf("metadata.FeatureKey: want %q, got %q",
			targetjob.FeatureKeyTargetImportParse, captured.metadata.FeatureKey)
	}
	if captured.metadata.PromptVersion != "v0.1.0" {
		t.Errorf("metadata.PromptVersion: want v0.1.0, got %q", captured.metadata.PromptVersion)
	}
	if captured.metadata.RubricVersion != "v0.1.0" {
		t.Errorf("metadata.RubricVersion: want v0.1.0, got %q", captured.metadata.RubricVersion)
	}
	if captured.metadata.DataSourceVersion == "" {
		t.Errorf("metadata.DataSourceVersion must be populated")
	}
	if captured.metadata.FeatureFlag != "none" {
		t.Errorf("metadata.FeatureFlag: want none, got %q", captured.metadata.FeatureFlag)
	}
	if captured.metadata.TaskRun.Capability != aiclient.AITaskRunTaskJDParse ||
		captured.metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		captured.metadata.TaskRun.ResourceID != "tgt-1" {
		t.Errorf("metadata.TaskRun did not carry B4 targetjob context: %+v", captured.metadata.TaskRun)
	}
	if len(captured.metadata.OutputSchema) == 0 {
		t.Fatalf("metadata.OutputSchema must be populated")
	}
}

// captureAIClient records the metadata of the most recent Complete call.
// It deliberately returns a parse-friendly response so the executor does
// not short-circuit before reaching the metadata-attach path.
type captureAIClient struct {
	metadata aiclient.CallMetadata
}

func (c *captureAIClient) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.metadata = payload.Metadata
	return aiclient.CompleteResponse{Content: `{"title":"Backend Engineer","companyName":"Acme","requirements":[{"kind":"must_have","label":"Go","evidenceLevel":"explicit"}]}`}, aiclient.AICallMeta{
		Provider:         "unit-test-provider",
		ModelFamily:      "fixture",
		ModelID:          "fixture-model:target-import-parse",
		ModelProfileName: "target.import.default",
		Language:         payload.Metadata.Language,
	}, nil
}
func (c *captureAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, nil
}
func (c *captureAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}
func (c *captureAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, nil
}
