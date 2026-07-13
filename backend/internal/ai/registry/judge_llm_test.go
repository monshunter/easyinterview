package registry

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

// fakeJudgeModel is a recorded-fixture JudgeModelClient: it never makes a
// network call and returns a canned judge transcript. The eval harness injects
// a real *aiclient.Client (CompleteJudge) in production.
type fakeJudgeModel struct {
	content     string
	err         error
	steps       []fakeJudgeStep
	calls       int
	lastProfile string
	lastPayload aiclient.CompletePayload
}

type fakeJudgeStep struct {
	content string
	meta    aiclient.AICallMeta
	err     error
}

func (f *fakeJudgeModel) CompleteJudge(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.calls++
	f.lastProfile = profileName
	f.lastPayload = payload
	if len(f.steps) > 0 {
		step := f.steps[0]
		f.steps = f.steps[1:]
		return aiclient.CompleteResponse{Content: step.content}, step.meta, step.err
	}
	if f.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, f.err
	}
	return aiclient.CompleteResponse{Content: f.content}, aiclient.AICallMeta{Capability: aiclient.CapabilityJudge}, nil
}

func TestLLMJudgeRetryBudget(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("practice.session.chat", "v0.1.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	valid := recordedJudgeTranscript(t, rubric, 0.8, "grounded")

	t.Run("retryable provider failure then success", func(t *testing.T) {
		model := &fakeJudgeModel{steps: []fakeJudgeStep{
			{err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "temporary", true)},
			{content: valid},
		}}
		judge := mustJudge(t, reg, model)
		if _, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0"); err != nil {
			t.Fatalf("Judge: %v", err)
		}
		if model.calls != 2 {
			t.Fatalf("calls=%d, want 2", model.calls)
		}
	})

	t.Run("protocol invalid then success", func(t *testing.T) {
		model := &fakeJudgeModel{steps: []fakeJudgeStep{{content: "not-json"}, {content: valid}}}
		judge := mustJudge(t, reg, model)
		if _, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0"); err != nil {
			t.Fatalf("Judge: %v", err)
		}
		if model.calls != 2 {
			t.Fatalf("calls=%d, want 2", model.calls)
		}
	})

	t.Run("four protocol invalid attempts fail closed", func(t *testing.T) {
		model := &fakeJudgeModel{content: "not-json"}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeProtocolInvalid) || model.calls != 4 {
			t.Fatalf("error=%v calls=%d, want typed protocol error after 4 calls", err, model.calls)
		}
	})

	t.Run("nonretryable provider failure is terminal", func(t *testing.T) {
		tests := []struct {
			name string
			ctx  context.Context
			err  error
		}{
			{name: "config", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiProviderConfigInvalid, "bad config", false)},
			{name: "secret", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "missing secret", false)},
			{name: "unsupported", ctx: context.Background(), err: sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "unsupported", false)},
			{name: "cancelled", ctx: cancelledContext(), err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "cancelled", true)},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				model := &fakeJudgeModel{err: tc.err}
				judge := mustJudge(t, reg, model)
				_, _, err := judge.Judge(tc.ctx, "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
				if err == nil || model.calls != 1 {
					t.Fatalf("error=%v calls=%d, want terminal failure after 1 call", err, model.calls)
				}
			})
		}
	})
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func newRepoRegistryClient(t *testing.T) *Client {
	t.Helper()
	prompts, rubrics := testsupport.ConfigRoots(t)
	c, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	return c
}

// testJudgeInstruction stands in for the config/evals scoring instruction
// template injected in production (Phase 4). Test-only literal: the production
// LLMJudge never hardcodes the scoring prompt.
const testJudgeInstruction = "Offline evaluation judge. Score each rubric dimension from 0 to 1 and return strict JSON."

// recordedJudgeTranscript builds a deterministic one-score-per-dimension judge
// response for the given rubric.
func recordedJudgeTranscript(t *testing.T, rubric RubricSchema, value float64, summary string) string {
	t.Helper()
	scores := make([]map[string]any, 0, len(rubric.Dimensions))
	for _, d := range rubric.Dimensions {
		scores = append(scores, map[string]any{"dimension": d.Name, "value": value})
	}
	raw, err := json.Marshal(map[string]any{
		"scores":    scores,
		"reasoning": map[string]any{"summary": summary, "evidence_quotes": []string{}},
	})
	if err != nil {
		t.Fatalf("marshal transcript: %v", err)
	}
	return string(raw)
}

const validPracticeChatOutput = `{"messageText":"Tell me more about the rollback you mentioned."}`

const validGroundedReportOutput = `{
  "summary":"The answer explained queue backpressure but left rollback verification unclear.",
  "preparednessLevel":"needs_practice",
  "dimensionAssessments":[{"code":"risk_handling","label":"Risk handling","status":"needs_work","confidence":"high"}],
  "highlights":[{"dimensionCode":"risk_handling","evidence":"Explained queue backpressure and retry tradeoffs.","confidence":"high","sourceMessageSeqNos":[2]}],
  "issues":[{"dimensionCode":"risk_handling","evidence":"Did not identify a concrete rollback verification signal.","confidence":"medium","sourceMessageSeqNos":[2]}],
  "nextActions":[{"type":"retry_current_round","label":"Practice rollback verification with a measurable signal"}],
  "retryFocusDimensionCodes":["risk_handling"]
}`

func recordedGroundedJudgeTranscript(t *testing.T, rubric RubricSchema) string {
	t.Helper()
	scores := make([]map[string]any, 0, len(rubric.Dimensions))
	for _, dimension := range rubric.Dimensions {
		scores = append(scores, map[string]any{"dimension": dimension.Name, "value": 0.85})
	}
	raw, err := json.Marshal(map[string]any{
		"scores": scores,
		"reasoning": map[string]any{
			"summary":         "All report claims are grounded and the retry advice follows the issue.",
			"evidence_quotes": []string{},
		},
		"item_verdicts": []map[string]any{
			{"path": "$.summary", "kind": "judgment", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": false, "reason": "Summary matches cited evidence."},
			{"path": "$.preparednessLevel", "kind": "judgment", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": true, "reason": "Readiness follows the supported answer gap."},
			{"path": "$.dimensionAssessments[0]", "kind": "judgment", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": true, "reason": "Needs-work status follows the answer gap."},
			{"path": "$.highlights[0]", "kind": "fact", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": false, "reason": "Highlight is present in message 2."},
			{"path": "$.issues[0]", "kind": "judgment", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": true, "reason": "Issue is present in message 2."},
			{"path": "$.nextActions[0]", "kind": "advice", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": false, "reason": "Action repairs the supported issue."},
			{"path": "$.retryFocusDimensionCodes", "kind": "advice", "support": "supported", "evidence_limited_explicit": false, "used_for_negative_claim": false, "reason": "Retry focus follows the supported needs-work dimension."},
		},
		"causal_checks": []map[string]any{
			{"dimension_code": "risk_handling", "issue_supported": true, "focus_supported": true, "action_supported": true, "reason": "Issue, focus, and action align."},
		},
		"zero_tolerance_violations": []string{},
		"critical_safety_pass":      true,
	})
	if err != nil {
		t.Fatalf("marshal grounded judge transcript: %v", err)
	}
	return string(raw)
}

func TestLLMJudgeReturnsPerDimensionScores(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("practice.session.chat", "v0.1.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	model := &fakeJudgeModel{content: recordedJudgeTranscript(t, rubric, 0.8, "meets the dimensions")}

	judge, err := NewLLMJudge(reg, model, testJudgeInstruction)
	if err != nil {
		t.Fatalf("NewLLMJudge: %v", err)
	}
	var _ Judge = judge // compile-time interface assertion (item 2.6)

	scores, reasoning, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
	if err != nil {
		t.Fatalf("Judge: %v", err)
	}
	if len(scores) != len(rubric.Dimensions) {
		t.Fatalf("scores len: want %d, got %d", len(rubric.Dimensions), len(scores))
	}
	names := map[string]bool{}
	for _, s := range scores {
		names[s.Dimension] = true
		if s.Value < 0 || s.Value > 1 {
			t.Fatalf("score value out of [0,1]: %+v", s)
		}
	}
	for _, d := range rubric.Dimensions {
		if !names[d.Name] {
			t.Fatalf("missing dimension %q in returned scores", d.Name)
		}
	}
	if reasoning.Summary == "" {
		t.Fatalf("reasoning.Summary must be non-empty")
	}
	if model.lastProfile != "judge.default" {
		t.Fatalf("expected judge.default profile, got %q", model.lastProfile)
	}
	if len(model.lastPayload.Messages) == 0 {
		t.Fatalf("expected judge payload messages to be built")
	}

}

func TestLLMJudgeRequestCarriesCompleteRubricDimensions(t *testing.T) {
	rubric := RubricSchema{Dimensions: []RubricDimension{
		{
			Name: "evidence", Weight: 0.65, Description: "Evidence quality",
			ScoreLevels: []ScoreLevel{
				{Label: "weak", Threshold: 0, Description: "Unsupported"},
				{Label: "strong", Threshold: 0.9, Description: "Fully grounded"},
			},
		},
		{
			Name: "action", Weight: 0.35, Description: "Action quality",
			ScoreLevels: []ScoreLevel{
				{Label: "weak", Threshold: 0, Description: "Not executable"},
				{Label: "strong", Threshold: 0.9, Description: "Immediately executable"},
			},
		},
	}}
	judge := &LLMJudge{instruction: testJudgeInstruction, language: "multi"}
	payload, err := judge.buildPayload("report.generate", "v0.2.0", "v0.2.0", rubric, []byte(validGroundedReportOutput), JudgeContext{})
	if err != nil {
		t.Fatalf("buildPayload: %v", err)
	}
	var request struct {
		Dimensions []struct {
			Name        string       `json:"name"`
			Weight      float64      `json:"weight"`
			Description string       `json:"description"`
			ScoreLevels []ScoreLevel `json:"score_levels"`
		} `json:"dimensions"`
	}
	if err := json.Unmarshal([]byte(payload.Messages[1].Content), &request); err != nil {
		t.Fatalf("parse judge request: %v", err)
	}
	if len(request.Dimensions) != len(rubric.Dimensions) {
		t.Fatalf("request dimensions len: want %d, got %d", len(rubric.Dimensions), len(request.Dimensions))
	}
	for i, want := range rubric.Dimensions {
		got := request.Dimensions[i]
		if got.Name != want.Name || got.Weight != want.Weight || got.Description != want.Description || !reflect.DeepEqual(got.ScoreLevels, want.ScoreLevels) {
			t.Fatalf("request dimension[%d]: want name=%q weight=%v description=%q score_levels=%+v, got name=%q weight=%v description=%q score_levels=%+v", i, want.Name, want.Weight, want.Description, want.ScoreLevels, got.Name, got.Weight, got.Description, got.ScoreLevels)
		}
	}
}

func TestLLMJudgeGroundedReportCarriesContextTranscriptAndOutput(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("report.generate", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	model := &fakeJudgeModel{content: recordedGroundedJudgeTranscript(t, rubric)}
	judge, err := NewLLMJudge(reg, model, testJudgeInstruction)
	if err != nil {
		t.Fatalf("NewLLMJudge: %v", err)
	}

	contextJSON := json.RawMessage(`{"targetJobTitle":"Backend Engineer","hasNextRound":false}`)
	transcriptJSON := json.RawMessage(`[{"seqNo":2,"role":"user","content":"I used queue depth, but did not define the rollback verification signal."}]`)
	_, reasoning, err := judge.JudgeWithContext(
		context.Background(),
		"report.generate",
		"v0.2.0",
		[]byte(validGroundedReportOutput),
		"v0.2.0",
		JudgeContext{FrozenContext: contextJSON, Transcript: transcriptJSON},
	)
	if err != nil {
		t.Fatalf("JudgeWithContext: %v", err)
	}
	if len(reasoning.ItemVerdicts) != 7 || reasoning.ItemVerdicts[1].Path != "$.preparednessLevel" || reasoning.ItemVerdicts[1].Kind != "judgment" || reasoning.ItemVerdicts[6].Path != "$.retryFocusDimensionCodes" || reasoning.ItemVerdicts[6].Kind != "advice" || len(reasoning.CausalChecks) != 1 || !reasoning.CriticalSafetyPass {
		t.Fatalf("grounded reasoning = %+v", reasoning)
	}

	var request map[string]any
	if err := json.Unmarshal([]byte(model.lastPayload.Messages[1].Content), &request); err != nil {
		t.Fatalf("parse judge request: %v", err)
	}
	for _, key := range []string{"frozen_context", "transcript", "output_to_evaluate", "expected_item_verdicts", "expected_causal_dimension_codes"} {
		if _, ok := request[key]; !ok {
			t.Fatalf("judge request missing %q: %v", key, request)
		}
	}
	itemRaw, ok := request["expected_item_verdicts"].([]any)
	if !ok {
		t.Fatalf("expected_item_verdicts = %#v", request["expected_item_verdicts"])
	}
	gotItems := make([]string, 0, len(itemRaw))
	for _, raw := range itemRaw {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected item = %#v", raw)
		}
		gotItems = append(gotItems, item["path"].(string)+":"+item["kind"].(string))
	}
	wantItems := []string{
		"$.summary:judgment",
		"$.preparednessLevel:judgment",
		"$.dimensionAssessments[0]:judgment",
		"$.highlights[0]:fact",
		"$.issues[0]:judgment",
		"$.nextActions[0]:advice",
		"$.retryFocusDimensionCodes:advice",
	}
	if !reflect.DeepEqual(gotItems, wantItems) {
		t.Fatalf("expected item verdict coordinates = %v, want %v", gotItems, wantItems)
	}
	if got := request["expected_causal_dimension_codes"]; !reflect.DeepEqual(got, []any{"risk_handling"}) {
		t.Fatalf("expected causal codes = %#v", got)
	}
}
