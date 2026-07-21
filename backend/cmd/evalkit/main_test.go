package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/eval"
	"github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

func TestReportJudgeInstructionOmitsVerdictsForEmptyCollections(t *testing.T) {
	t.Parallel()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", ".."))
	body, err := os.ReadFile(filepath.Join(repoRoot, "config", "evals", "judge-instruction.md"))
	if err != nil {
		t.Fatal(err)
	}
	normalized := strings.Join(strings.Fields(strings.ToLower(string(body))), " ")
	for _, marker := range []string{
		"emit no collection-level verdict for `$.highlights`, `$.issues`, or `$.nextactions`",
		"an empty highlights or issues array produces zero verdicts for that array",
		"`$.retryfocusdimensioncodes` remains the only array-level verdict",
		"copy those `path` and `kind` pairs exactly once in the provided order",
		"emit exactly one causal check for each listed code and no other code",
		"`retry_current_round` advice may only turn cited missing behavior into something to add",
		"`review_evidence` may ask the user to revisit cited positive or explicitly evidence-limited content",
		"without inventing an artifact, corrective gap, new scenario, or transfer task",
		"`next_round` is supported only when frozen `hasnextround` is true",
		"a focused retry is unsupported when its label uses only an umbrella term instead of one directly cited missing behavior per selected focus code",
	} {
		if !strings.Contains(normalized, marker) {
			t.Fatalf("judge instruction missing %q", marker)
		}
	}
}

type auditJudgeStub struct {
	response aiclient.CompleteResponse
	meta     aiclient.AICallMeta
}

type liveCompletionStep struct {
	response aiclient.CompleteResponse
	meta     aiclient.AICallMeta
	err      error
}

type scriptedLiveCompleter struct {
	steps    []liveCompletionStep
	payloads []aiclient.CompletePayload
}

func (s *scriptedLiveCompleter) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	s.payloads = append(s.payloads, payload)
	if len(s.steps) == 0 {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, assertError("unexpected completion call")
	}
	step := s.steps[0]
	s.steps = s.steps[1:]
	return step.response, step.meta, step.err
}

type assertError string

func (e assertError) Error() string { return string(e) }

func (s auditJudgeStub) CompleteJudge(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return s.response, s.meta, nil
}

func TestValidateCaseCountRequiresExactCurrentSuite(t *testing.T) {
	if err := validateCaseCount(32); err != nil {
		t.Fatalf("exact current suite must pass: %v", err)
	}
	for _, count := range []int{31, 33} {
		if err := validateCaseCount(count); err == nil {
			t.Fatalf("case count %d must fail", count)
		}
	}
}

func TestValidateLiveReportOutputSchemaRejectsOversizedActionLabel(t *testing.T) {
	t.Parallel()
	max := 200
	schema := json.RawMessage(`{"type":"object","required":["label"],"additionalProperties":false,"properties":{"label":{"type":"string","maxLength":200}}}`)
	resolution := registry.PromptResolution{OutputSchema: &schema}
	response := aiclient.CompleteResponse{Content: `{"label":"` + strings.Repeat("x", max+1) + `"}`}
	if err := validateLiveReportOutputSchema(response, resolution); err == nil || !strings.Contains(err.Error(), "output schema invalid") {
		t.Fatalf("oversized live report output must fail schema validation, got %v", err)
	}
	response.Content = `{"label":"` + strings.Repeat("x", max) + `"}`
	if err := validateLiveReportOutputSchema(response, resolution); err != nil {
		t.Fatalf("exact-boundary live report output must pass: %v", err)
	}
}

func TestCompleteLiveReportRepairsExactlyOnceAfterSchemaInvalid(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	invalidSummary := strings.Repeat("x", 361)
	repairedOutput := liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail")
	firstMeta := liveReportMeta(resolution, c.Language, 100, 20, 300)
	repairMeta := liveReportMeta(resolution, c.Language, 140, 30, 500)
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{response: aiclient.CompleteResponse{Content: liveReportSummaryJSON(t, "en", invalidSummary), FinishReason: "stop"}, meta: firstMeta},
		{response: aiclient.CompleteResponse{Content: repairedOutput, FinishReason: "stop"}, meta: repairMeta},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err != nil {
		t.Fatalf("completeLiveReportWithRepair: %v", err)
	}
	if !result.repairUsed || len(client.payloads) != 2 {
		t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
	}
	if result.repairScope != repairScopeWholeReport {
		t.Fatalf("repairScope=%q, want %q", result.repairScope, repairScopeWholeReport)
	}
	if result.response.Content != repairedOutput {
		t.Fatalf("final response = %s", result.response.Content)
	}
	if result.meta.InputTokens != 240 || result.meta.OutputTokens != 50 || result.meta.LatencyMs != 800 {
		t.Fatalf("aggregated meta = %+v", result.meta)
	}
	if result.attemptCount != 2 || !reflect.DeepEqual(result.retryReasons, []string{"output_schema_invalid"}) || !reflect.DeepEqual(result.repairScopes, []string{repairScopeWholeReport}) {
		t.Fatalf("retry audit = attempts=%d reasons=%v scopes=%v", result.attemptCount, result.retryReasons, result.repairScopes)
	}
	if client.payloads[0].Messages[1].Content != client.payloads[1].Messages[1].Content {
		t.Fatal("repair must keep the exact untrusted context")
	}
	if strings.Contains(client.payloads[1].Messages[0].Content, invalidSummary) {
		t.Fatal("repair guidance leaked the invalid model output")
	}
	for _, want := range []string{`"path":"$.summary"`, `"code":"max_length"`, "The field at the supplied path exceeds its maximum length"} {
		if !strings.Contains(client.payloads[1].Messages[0].Content, want) {
			t.Fatalf("repair guidance missing %s: %s", want, client.payloads[1].Messages[0].Content)
		}
	}
}

func TestCompleteLiveReportFourthSchemaInvalidFailsClosed(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{response: aiclient.CompleteResponse{Content: liveReportSummaryJSON(t, "en", strings.Repeat("x", 361)), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 100, 20, 300)},
		{response: aiclient.CompleteResponse{Content: liveReportSummaryJSON(t, "en", strings.Repeat("y", 361)), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 140, 30, 500)},
		{response: aiclient.CompleteResponse{Content: liveReportSummaryJSON(t, "en", strings.Repeat("z", 361)), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 160, 40, 700)},
		{response: aiclient.CompleteResponse{Content: liveReportSummaryJSON(t, "en", strings.Repeat("q", 361)), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 180, 50, 900)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err == nil || !strings.Contains(err.Error(), "$.summary:max_length") {
		t.Fatalf("second invalid output must fail closed, got %v", err)
	}
	if !result.repairUsed || len(client.payloads) != 4 {
		t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
	}
	if result.repairScope != repairScopeWholeReport {
		t.Fatalf("repairScope=%q, want %q", result.repairScope, repairScopeWholeReport)
	}
	if result.meta.InputTokens != 580 || result.meta.OutputTokens != 140 || result.meta.LatencyMs != 2400 {
		t.Fatalf("failed repair meta must retain aggregate usage: %+v", result.meta)
	}
	if result.attemptCount != 4 || len(result.retryReasons) != 3 || len(result.repairScopes) != 3 {
		t.Fatalf("retry audit = attempts=%d reasons=%v scopes=%v", result.attemptCount, result.retryReasons, result.repairScopes)
	}
}

func TestCompleteLiveReportRetriesProviderFailureThenSucceeds(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{meta: liveReportMeta(resolution, c.Language, 10, 2, 100), err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "temporary", true)},
		{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail"), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 100, 20, 300)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err != nil {
		t.Fatalf("completeLiveReportWithRepair: %v", err)
	}
	if len(client.payloads) != 2 || result.attemptCount != 2 || result.repairUsed {
		t.Fatalf("calls=%d attempts=%d repairUsed=%t", len(client.payloads), result.attemptCount, result.repairUsed)
	}
	if !reflect.DeepEqual(result.retryReasons, []string{"provider_retryable"}) || !reflect.DeepEqual(result.repairScopes, []string{repairScopeNone}) {
		t.Fatalf("reasons=%v scopes=%v", result.retryReasons, result.repairScopes)
	}
	if result.meta.InputTokens != 110 || result.meta.OutputTokens != 22 || result.meta.LatencyMs != 400 {
		t.Fatalf("aggregate meta=%+v", result.meta)
	}
}

func TestCompleteLiveReportNonretryableProviderFailureStopsAfterOneCall(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	tests := []struct {
		name string
		ctx  context.Context
		err  error
	}{
		{name: "config", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiProviderConfigInvalid, "bad config", false)},
		{name: "secret", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "missing secret", false)},
		{name: "unsupported", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "unsupported", false)},
		{name: "cancelled", ctx: cancelledEvalContext(), err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "cancelled", true)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &scriptedLiveCompleter{steps: []liveCompletionStep{{err: tc.err}}}
			result, err := completeLiveReportWithRepair(tc.ctx, client, resolution, c, payload)
			if err == nil || len(client.payloads) != 1 || result.attemptCount != 1 || len(result.retryReasons) != 0 {
				t.Fatalf("error=%v calls=%d result=%+v", err, len(client.payloads), result)
			}
		})
	}
}

func cancelledEvalContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestCompleteLiveReportSchemaValidDoesNotRepair(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{{
		response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail"), FinishReason: "stop"},
		meta:     liveReportMeta(resolution, c.Language, 100, 20, 300),
	}}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err != nil {
		t.Fatalf("completeLiveReportWithRepair: %v", err)
	}
	if result.repairUsed || len(client.payloads) != 1 {
		t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
	}
	if result.repairScope != repairScopeNone {
		t.Fatalf("repairScope=%q, want %q", result.repairScope, repairScopeNone)
	}
}

func TestCompleteLiveReportRepairsSchemaValidActionLabelLimitViolation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		language       string
		invalidLabel   string
		repairedLabel  string
		code           string
		repairGuidance string
		decoratedError bool
	}{
		{
			name:           "english",
			language:       "en",
			invalidLabel:   strings.TrimSpace(strings.Repeat("word ", 25)),
			repairedLabel:  "Add one supported failure scenario and explain its bounded fallback",
			code:           "max_words",
			repairGuidance: "hard target of 4-18 whitespace-delimited words",
		},
		{
			name:           "english schema fuse within word limit",
			language:       "en",
			invalidLabel:   strings.TrimSpace(strings.Repeat("extraordinaryword ", 14)),
			repairedLabel:  "Add bounded rollback steps and replay this round",
			code:           "max_length",
			repairGuidance: "hard target of 4-18 whitespace-delimited words",
			decoratedError: true,
		},
		{
			name:           "chinese",
			language:       "zh-CN",
			invalidLabel:   strings.Repeat("改", 65),
			repairedLabel:  "补充失败场景与回退边界",
			code:           "max_code_points",
			repairGuidance: "hard target of at most 52 Unicode code points",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolution, c, payload := liveReportActionLimitFixture(t, tc.language)
			firstMeta := liveReportMeta(resolution, c.Language, 100, 20, 300)
			firstMeta.CostUSDMicros = 11
			repairMeta := liveReportMeta(resolution, c.Language, 140, 30, 500)
			repairMeta.CostUSDMicros = 17
			initialStep := liveCompletionStep{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSONForLanguage(t, tc.language, tc.invalidLabel), FinishReason: "stop"}, meta: firstMeta}
			if tc.decoratedError {
				initialStep.meta.ValidationStatus = aiclient.ValidationStatusInvalid
				initialStep.meta.ErrorCode = sharederrors.CodeAiOutputInvalid
				initialStep.err = sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "output failed schema validation: $.nextActions[0].label length exceeds 200", false)
			}
			client := &scriptedLiveCompleter{steps: []liveCompletionStep{
				initialStep,
				{response: aiclient.CompleteResponse{Content: liveReportLabelRepairJSON(t, 0, tc.repairedLabel), FinishReason: "stop"}, meta: repairMeta},
			}}

			result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
			if err != nil {
				t.Fatalf("completeLiveReportWithRepair: %v", err)
			}
			if !result.repairUsed || len(client.payloads) != 2 {
				t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
			}
			if result.repairScope != repairScopeActionLabels {
				t.Fatalf("repairScope=%q, want %q", result.repairScope, repairScopeActionLabels)
			}
			var final struct {
				Summary     string                         `json:"summary"`
				NextActions []review.ReportNextActionDraft `json:"nextActions"`
			}
			if err := json.Unmarshal([]byte(result.response.Content), &final); err != nil || len(final.NextActions) != 1 || final.NextActions[0].Label != tc.repairedLabel || final.Summary != liveReportDefaultSummary(tc.language) {
				t.Fatalf("final response = %s err=%v", result.response.Content, err)
			}
			if result.meta.InputTokens != 240 || result.meta.OutputTokens != 50 || result.meta.LatencyMs != 800 || result.meta.CostUSDMicros != 28 {
				t.Fatalf("aggregated meta = %+v", result.meta)
			}
			repairSystem := client.payloads[1].Messages[0].Content
			repairUser := client.payloads[1].Messages[1].Content
			if strings.Contains(repairSystem, tc.invalidLabel) {
				t.Fatal("repair guidance leaked the invalid model output")
			}
			if strings.Contains(repairSystem+repairUser, "PRIVATE-ANSWER") {
				t.Fatal("targeted repair leaked the original completion prompt context")
			}
			for _, want := range []string{`"path":"$.nextActions[0].label"`, `"code":"` + tc.code + `"`, `"type":"retry_current_round"`, tc.invalidLabel, `"relatedIssues"`, `"retryFocusDimensionCodes":["technical_depth"]`} {
				if !strings.Contains(repairUser, want) {
					t.Fatalf("repair untrusted data missing %s: %s", want, repairUser)
				}
			}
			if !strings.Contains(repairSystem, tc.repairGuidance) {
				t.Fatalf("repair guidance missing %s: %s", tc.repairGuidance, repairSystem)
			}
		})
	}
}

func TestCompleteLiveReportFourthActionLabelLimitViolationFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		language     string
		invalidLabel string
		code         string
	}{
		{
			name:         "english",
			language:     "en",
			invalidLabel: strings.TrimSpace(strings.Repeat("word ", 25)),
			code:         "max_words",
		},
		{
			name:         "chinese",
			language:     "zh-CN",
			invalidLabel: strings.Repeat("改", 65),
			code:         "max_code_points",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resolution, c, payload := liveReportActionLimitFixture(t, tc.language)
			firstMeta := liveReportMeta(resolution, c.Language, 100, 20, 300)
			firstMeta.CostUSDMicros = 11
			repairMeta := liveReportMeta(resolution, c.Language, 140, 30, 500)
			repairMeta.CostUSDMicros = 17
			client := &scriptedLiveCompleter{steps: []liveCompletionStep{
				{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSONForLanguage(t, tc.language, tc.invalidLabel), FinishReason: "stop"}, meta: firstMeta},
				{response: aiclient.CompleteResponse{Content: liveReportLabelRepairJSON(t, 0, tc.invalidLabel), FinishReason: "stop"}, meta: repairMeta},
				{response: aiclient.CompleteResponse{Content: liveReportLabelRepairJSON(t, 0, tc.invalidLabel), FinishReason: "stop"}, meta: repairMeta},
				{response: aiclient.CompleteResponse{Content: liveReportLabelRepairJSON(t, 0, tc.invalidLabel), FinishReason: "stop"}, meta: repairMeta},
			}}

			result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
			if err == nil || !strings.Contains(err.Error(), "$.nextActions[0].label:"+tc.code) {
				t.Fatalf("second semantic violation must fail closed, got %v", err)
			}
			if strings.Contains(err.Error(), tc.invalidLabel) {
				t.Fatal("failure leaked the invalid model output")
			}
			if !result.repairUsed || len(client.payloads) != 4 {
				t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
			}
			if result.repairScope != repairScopeActionLabels {
				t.Fatalf("repairScope=%q, want %q", result.repairScope, repairScopeActionLabels)
			}
			if result.meta.InputTokens != 520 || result.meta.OutputTokens != 110 || result.meta.LatencyMs != 1800 || result.meta.CostUSDMicros != 62 {
				t.Fatalf("failed repair meta must retain aggregate usage: %+v", result.meta)
			}
			if result.attemptCount != 4 || len(result.retryReasons) != 3 || len(result.repairScopes) != 3 {
				t.Fatalf("retry audit = attempts=%d reasons=%v scopes=%v", result.attemptCount, result.retryReasons, result.repairScopes)
			}
		})
	}
}

func TestCompleteLiveReportRoutesReadinessInconsistencyToWholeReportRepair(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportActionLimitFixture(t, "en")
	repairedOutput := liveReportGenericRetryJSON(t, false, "")
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{response: aiclient.CompleteResponse{Content: liveReportGenericRetryJSON(t, true, "Continue to the next round"), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 100, 20, 300)},
		{response: aiclient.CompleteResponse{Content: repairedOutput, FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 140, 30, 500)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err != nil {
		t.Fatalf("completeLiveReportWithRepair: %v", err)
	}
	if !result.repairUsed || result.repairScope != repairScopeWholeReport || len(client.payloads) != 2 {
		t.Fatalf("repairUsed=%t scope=%q calls=%d", result.repairUsed, result.repairScope, len(client.payloads))
	}
	if result.response.Content != repairedOutput {
		t.Fatalf("final response = %s", result.response.Content)
	}
}

func TestCompleteLiveReportRoutesMixedLabelAndReadinessViolationsToWholeReportRepair(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportActionLimitFixture(t, "en")
	longLabel := strings.TrimSpace(strings.Repeat("extraordinaryword ", 14))
	invalidMeta := liveReportMeta(resolution, c.Language, 100, 20, 300)
	invalidMeta.ValidationStatus = aiclient.ValidationStatusInvalid
	invalidMeta.ErrorCode = sharederrors.CodeAiOutputInvalid
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{
			response: aiclient.CompleteResponse{Content: liveReportGenericRetryJSON(t, true, longLabel), FinishReason: "stop"},
			meta:     invalidMeta,
			err:      sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "output failed schema validation: $.nextActions[1].label length exceeds 200", false),
		},
		{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail"), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 140, 30, 500)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err != nil {
		t.Fatalf("completeLiveReportWithRepair: %v", err)
	}
	if !result.repairUsed || result.repairScope != repairScopeWholeReport || len(client.payloads) != 2 {
		t.Fatalf("repairUsed=%t scope=%q calls=%d", result.repairUsed, result.repairScope, len(client.payloads))
	}
	if strings.Contains(client.payloads[1].Messages[0].Content, longLabel) {
		t.Fatal("whole-report repair guidance leaked the invalid label")
	}
}

func TestCompleteLiveReportRevalidatesFullSemanticsAfterTargetedLabelMerge(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportActionLimitFixture(t, "en")
	invalidLabel := strings.TrimSpace(strings.Repeat("word ", 25))
	rawReplacement := "RAW-PRIVATE hiring probability improves after this retry"
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, invalidLabel), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 100, 20, 300)},
		{response: aiclient.CompleteResponse{Content: liveReportLabelRepairJSON(t, 0, rawReplacement), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 140, 30, 500)},
		{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, rawReplacement), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 160, 40, 700)},
		{response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, rawReplacement), FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 180, 50, 900)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err == nil || !strings.Contains(err.Error(), "$.nextActions[0].label:forbidden_claim") {
		t.Fatalf("targeted merge semantic violation must fail closed, got %v", err)
	}
	if strings.Contains(err.Error(), rawReplacement) {
		t.Fatal("targeted merge failure leaked repaired raw content")
	}
	if !result.repairUsed || result.repairScope != repairScopeWholeReport || len(client.payloads) != 4 {
		t.Fatalf("repairUsed=%t scope=%q calls=%d", result.repairUsed, result.repairScope, len(client.payloads))
	}
}

func TestCompleteLiveReportFailsClosedWhenWholeReportRepairRemainsSemanticallyInvalid(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportActionLimitFixture(t, "en")
	invalid := liveReportGenericRetryJSON(t, true, "Continue to the next round")
	client := &scriptedLiveCompleter{steps: []liveCompletionStep{
		{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 100, 20, 300)},
		{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 140, 30, 500)},
		{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 160, 40, 700)},
		{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: liveReportMeta(resolution, c.Language, 180, 50, 900)},
	}}

	result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
	if err == nil || !strings.Contains(err.Error(), "next_round_inconsistent_with_readiness") {
		t.Fatalf("second semantic violation must fail closed, got %v", err)
	}
	if strings.Contains(err.Error(), "Continue to the next round") {
		t.Fatal("whole-report repair failure leaked raw output")
	}
	if !result.repairUsed || result.repairScope != repairScopeWholeReport || len(client.payloads) != 4 {
		t.Fatalf("repairUsed=%t scope=%q calls=%d", result.repairUsed, result.repairScope, len(client.payloads))
	}
}

func TestCompleteLiveReportDoesNotRepairInvalidFinishOrMetadata(t *testing.T) {
	t.Parallel()
	resolution, c, payload := liveReportRepairFixture(t)
	tests := []struct {
		name     string
		response aiclient.CompleteResponse
		meta     aiclient.AICallMeta
	}{
		{
			name:     "finish reason",
			response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail"), FinishReason: "length"},
			meta:     liveReportMeta(resolution, c.Language, 100, 20, 300),
		},
		{
			name:     "metadata",
			response: aiclient.CompleteResponse{Content: liveReportActionOnlyJSON(t, "Retry this round with the cited rollback detail"), FinishReason: "stop"},
			meta: func() aiclient.AICallMeta {
				meta := liveReportMeta(resolution, c.Language, 100, 20, 300)
				meta.InputTokens = 0
				return meta
			}(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &scriptedLiveCompleter{steps: []liveCompletionStep{{response: tc.response, meta: tc.meta}}}
			result, err := completeLiveReportWithRepair(context.Background(), client, resolution, c, payload)
			if err == nil {
				t.Fatal("invalid first-call provenance must fail")
			}
			if result.repairUsed || len(client.payloads) != 1 {
				t.Fatalf("repairUsed=%t calls=%d", result.repairUsed, len(client.payloads))
			}
		})
	}
}

func liveReportRepairFixture(t *testing.T) (registry.PromptResolution, eval.Case, aiclient.CompletePayload) {
	t.Helper()
	schema := json.RawMessage(`{"type":"object","properties":{"summary":{"type":"string","maxLength":360},"nextActions":{"type":"array","items":{"type":"object","properties":{"type":{"type":"string"},"label":{"type":"string","maxLength":200}}}}}}`)
	resolution := registry.PromptResolution{
		FeatureKey: string(featurekeys.ReportGenerate), PromptVersion: "v0.2.0", RubricVersion: "v0.2.0",
		ModelProfileName: "report.generate.default", FeatureFlag: "none", DataSourceVersion: "report-context.v1",
		UserMessageTemplate: `Trusted report policy in {{language}}.

<untrusted_report_context_json>
{"context":{{frozen_context}},"messages":{{conversation_messages}}}
</untrusted_report_context_json>

Return strict JSON.`,
		OutputSchema: &schema,
	}
	c := eval.Case{
		ID: "report.generate-repair", FeatureKey: string(featurekeys.ReportGenerate), Language: "en",
		Context:    map[string]any{"roundName": "Technical", "language": "en", "hasNextRound": false},
		Transcript: []map[string]any{{"seqNo": 2, "role": "user", "content": "PRIVATE-ANSWER"}},
	}
	messages, err := renderCaseMessages(resolution, c)
	if err != nil {
		t.Fatal(err)
	}
	return resolution, c, aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey: resolution.FeatureKey, PromptVersion: resolution.PromptVersion, RubricVersion: resolution.RubricVersion,
			Language: c.Language, FeatureFlag: resolution.FeatureFlag, DataSourceVersion: resolution.DataSourceVersion,
			OutputSchema: append(json.RawMessage(nil), schema...),
		},
	}
}

func liveReportActionLimitFixture(t *testing.T, language string) (registry.PromptResolution, eval.Case, aiclient.CompletePayload) {
	t.Helper()
	schema := json.RawMessage(`{"type":"object","properties":{"nextActions":{"type":"array","items":{"type":"object","properties":{"type":{"type":"string"},"label":{"type":"string","maxLength":200}}}}}}`)
	resolution := registry.PromptResolution{
		FeatureKey: string(featurekeys.ReportGenerate), PromptVersion: "v0.2.0", RubricVersion: "v0.2.0",
		ModelProfileName: "report.generate.default", FeatureFlag: "none", DataSourceVersion: "report-context.v1",
		UserMessageTemplate: `Trusted report policy in {{language}}.

<untrusted_report_context_json>
{"context":{{frozen_context}},"messages":{{conversation_messages}}}
</untrusted_report_context_json>

Return strict JSON.`,
		OutputSchema: &schema,
	}
	c := eval.Case{
		ID: "report.generate-action-limit-repair", FeatureKey: string(featurekeys.ReportGenerate), Language: language,
		Context:    map[string]any{"roundName": "Technical", "language": language, "hasNextRound": true},
		Transcript: []map[string]any{{"seqNo": 2, "role": "user", "content": "PRIVATE-ANSWER"}},
	}
	messages, err := renderCaseMessages(resolution, c)
	if err != nil {
		t.Fatal(err)
	}
	return resolution, c, aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey: resolution.FeatureKey, PromptVersion: resolution.PromptVersion, RubricVersion: resolution.RubricVersion,
			Language: c.Language, FeatureFlag: resolution.FeatureFlag, DataSourceVersion: resolution.DataSourceVersion,
			OutputSchema: append(json.RawMessage(nil), schema...),
		},
	}
}

func liveReportActionOnlyJSON(t *testing.T, label string) string {
	t.Helper()
	return liveReportActionOnlyJSONForLanguage(t, "en", label)
}

func liveReportActionOnlyJSONForLanguage(t *testing.T, language, label string) string {
	t.Helper()
	return liveReportJSON(t, language, liveReportDefaultSummary(language), "retry_current_round", label)
}

func liveReportSummaryJSON(t *testing.T, language, summary string) string {
	t.Helper()
	label := "Retry this round with the cited rollback detail"
	if language == "zh-CN" {
		label = "复练当前轮并补充已引用的回滚细节"
	}
	return liveReportJSON(t, language, summary, "retry_current_round", label)
}

func liveReportDefaultSummary(language string) string {
	if language == "zh-CN" {
		return "本次回答仍需补充关键细节。"
	}
	return "Grounded summary"
}

func liveReportJSON(t *testing.T, language, summary, actionType, label string) string {
	t.Helper()
	dimensionLabel := "Technical depth"
	evidence := "The cited answer omitted rollback steps."
	if language == "zh-CN" {
		dimensionLabel = "技术深度"
		evidence = "回答没有说明已引用的回滚步骤。"
	}
	encoded, err := json.Marshal(map[string]any{
		"summary":           summary,
		"preparednessLevel": "needs_practice",
		"dimensionAssessments": []map[string]string{{
			"code": "technical_depth", "label": dimensionLabel, "status": "needs_work", "confidence": "high",
		}},
		"highlights": []any{},
		"issues": []map[string]any{{
			"dimensionCode": "technical_depth", "evidence": evidence, "confidence": "high", "sourceMessageSeqNos": []int{2},
		}},
		"nextActions":              []map[string]string{{"type": actionType, "label": label}},
		"retryFocusDimensionCodes": []string{"technical_depth"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func liveReportLabelRepairJSON(t *testing.T, index int, label string) string {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{"labels": []map[string]any{{"index": index, "label": label}}})
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func liveReportGenericRetryJSON(t *testing.T, includeNextRound bool, nextRoundLabel string) string {
	t.Helper()
	actions := []map[string]string{{"type": "retry_current_round", "label": "Retry this round with concrete supporting detail"}}
	if includeNextRound {
		actions = append(actions, map[string]string{"type": "next_round", "label": nextRoundLabel})
	}
	encoded, err := json.Marshal(map[string]any{
		"summary":           "The short answer identified one mechanism but did not provide concrete supporting detail.",
		"preparednessLevel": "needs_practice",
		"dimensionAssessments": []map[string]string{{
			"code": "answer_depth", "label": "Answer depth", "status": "needs_work", "confidence": "medium",
		}},
		"highlights": []map[string]any{{
			"dimensionCode": "answer_depth", "evidence": "The response identified one mechanism.", "confidence": "high", "sourceMessageSeqNos": []int{2},
		}},
		"issues": []map[string]any{{
			"dimensionCode": "answer_depth", "evidence": "The response provided no concrete supporting detail.", "confidence": "medium", "sourceMessageSeqNos": []int{2},
		}},
		"nextActions":              actions,
		"retryFocusDimensionCodes": []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func liveReportMeta(resolution registry.PromptResolution, language string, inputTokens, outputTokens int, latencyMs int64) aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider: "deepseek", ModelID: "deepseek-v4-pro", ModelProfileName: resolution.ModelProfileName,
		FeatureKey: resolution.FeatureKey, PromptVersion: resolution.PromptVersion, RubricVersion: resolution.RubricVersion,
		Language: language, FeatureFlag: resolution.FeatureFlag, DataSourceVersion: resolution.DataSourceVersion,
		InputTokens: inputTokens, OutputTokens: outputTokens, LatencyMs: latencyMs, ValidationStatus: aiclient.ValidationStatusOK,
	}
}

func TestRenderCaseMessagesPreservesGroundedReportTrustBoundary(t *testing.T) {
	messages, err := renderCaseMessages(
		registry.PromptResolution{UserMessageTemplate: `Trusted policy in {{language}}.

<untrusted_report_context_json>
{"frozenContext":{{frozen_context}},"conversationMessages":{{conversation_messages}}}
</untrusted_report_context_json>

Return strict JSON.`},
		eval.Case{
			FeatureKey: string(featurekeys.ReportGenerate),
			Language:   "en",
			Context:    map[string]any{"roundName": "Technical"},
			Transcript: []map[string]any{{"seqNo": 2, "role": "user", "content": "bounded retries"}},
		},
	)
	if err != nil {
		t.Fatalf("renderCaseMessages: %v", err)
	}
	if len(messages) != 2 || messages[0].Role != "system" || messages[1].Role != "user" {
		t.Fatalf("messages must be one trusted system plus one untrusted user: %#v", messages)
	}
	if !strings.Contains(messages[0].Content, "Trusted policy in en") {
		t.Fatalf("system policy missing canonical language: %s", messages[0].Content)
	}
	for _, untrusted := range []string{`"roundName":"Technical"`, `"content":"bounded retries"`} {
		if strings.Contains(messages[0].Content, untrusted) {
			t.Fatalf("untrusted value leaked into system role: %q", untrusted)
		}
		if !strings.Contains(messages[1].Content, untrusted) {
			t.Fatalf("untrusted user message missing %q: %s", untrusted, messages[1].Content)
		}
	}
	for _, stale := range []string{"{{frozen_context}}", "{{conversation_messages}}", "{{rubric_dimensions}}"} {
		if strings.Contains(messages[0].Content+messages[1].Content, stale) {
			t.Fatalf("rendered messages contain stale token %q", stale)
		}
	}
}

func TestBuildCompletionAuditContainsProvenanceWithoutRawContent(t *testing.T) {
	c := eval.Case{ID: "report.generate-redacted", FeatureKey: string(featurekeys.ReportGenerate), Language: "en", Critical: true}
	resolution := registry.PromptResolution{
		FeatureKey: "report.generate", PromptVersion: "v0.2.0", RubricVersion: "v0.2.0",
		ModelProfileName: "report.generate.default", FeatureFlag: "none", DataSourceVersion: "report-context.v1",
	}
	response := aiclient.CompleteResponse{Content: `{"summary":"RAW-CANDIDATE-SECRET"}`, FinishReason: "stop"}
	meta := aiclient.AICallMeta{
		Provider: "deepseek", ModelID: "deepseek-v4-pro", ModelProfileName: "report.generate.default", ModelProfileVersion: "1.2.0",
		FeatureKey: "report.generate", PromptVersion: "v0.2.0", RubricVersion: "v0.2.0", Language: "en",
		FeatureFlag: "none", DataSourceVersion: "report-context.v1", InputTokens: 123, OutputTokens: 45,
		LatencyMs: 678, ValidationStatus: aiclient.ValidationStatusOK,
	}
	audit := buildCompletionAudit(c, liveCompletionResult{
		resolution: resolution, response: response, meta: meta, repairUsed: true, repairScope: repairScopeActionLabels,
		attemptCount: 2, retryReasons: []string{retryReasonOutputSemantic}, repairScopes: []string{repairScopeActionLabels},
	}, nil)
	raw, err := json.Marshal(audit)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, forbidden := range []string{"RAW-CANDIDATE-SECRET", "prompt body", "response body"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("completion audit leaked %q: %s", forbidden, text)
		}
	}
	for _, want := range []string{`"schemaVersion":"evalkit-live-call-audit.v2"`, `"finishReason":"stop"`, `"inputTokens":123`, `"outputTokens":45`, `"repairUsed":true`, `"repairScope":"action_labels"`, `"attemptCount":2`, `"retryCount":1`, `"retryReasons":["output_semantic_invalid"]`, `"repairScopes":["action_labels"]`, `"outputSha256"`, `"featureFlag":"none"`, `"dataSourceVersion":"report-context.v1"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("completion audit missing %s: %s", want, text)
		}
	}
}

func TestGradeVerdictAndJudgeAuditExposeGroundedResultWithoutPrompt(t *testing.T) {
	reasoning := registry.Reasoning{
		Summary: "Grounded and executable.",
		ItemVerdicts: []registry.ItemVerdict{{
			Path: "$.issues[0]", Kind: "judgment", Support: "supported", Reason: "Candidate evidence supports the issue.",
		}},
		CausalChecks:       []registry.CausalCheck{{DimensionCode: "risk_handling", IssueSupported: true, FocusSupported: true, ActionSupported: true}},
		CriticalSafetyPass: true,
	}
	weightedScore := 0.91
	verdict := buildGradeVerdict(
		eval.Case{ID: "report.generate-redacted", Critical: true},
		[]registry.Score{{Dimension: "report_evidence", Value: 0.91}},
		reasoning,
		&weightedScore,
		nil,
	)
	raw, err := json.Marshal(verdict)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, want := range []string{`"weighted_score":0.91`, `"item_verdicts"`, `"causal_checks"`, `"zero_tolerance_violations":[]`, `"critical_safety_pass":true`, `"path":"$.issues[0]"`} {
		if !strings.Contains(text, want) {
			t.Fatalf("grade verdict missing %s: %s", want, text)
		}
	}

	stub := auditJudgeStub{
		response: aiclient.CompleteResponse{Content: `{"reasoning":"RAW-JUDGE-BODY"}`, FinishReason: "stop"},
		meta:     aiclient.AICallMeta{Provider: "judge-deepseek", ModelID: "deepseek-v4-pro", ModelProfileName: "judge.default", InputTokens: 321, OutputTokens: 67},
	}
	capture := &capturingJudgeModel{delegate: stub}
	if _, _, err := capture.CompleteJudge(context.Background(), "judge.default", aiclient.CompletePayload{}); err != nil {
		t.Fatal(err)
	}
	audit := buildJudgeAudit(eval.Case{ID: "report.generate-redacted", FeatureKey: "report.generate", Language: "en"}, capture, nil)
	auditRaw, err := json.Marshal(audit)
	if err != nil {
		t.Fatal(err)
	}
	auditText := string(auditRaw)
	if strings.Contains(auditText, "RAW-JUDGE-BODY") {
		t.Fatalf("judge audit leaked raw response: %s", auditText)
	}
	for _, want := range []string{`"modelProfileName":"judge.default"`, `"finishReason":"stop"`, `"inputTokens":321`, `"outputTokens":67`, `"repairUsed":false`, `"repairScope":"none"`, `"attemptCount":1`, `"retryCount":0`, `"retryReasons":[]`, `"repairScopes":[]`} {
		if !strings.Contains(auditText, want) {
			t.Fatalf("judge audit missing %s: %s", want, auditText)
		}
	}
}
