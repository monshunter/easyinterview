package registry

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

// fakeJudgeModel is a recorded-fixture JudgeModelClient: it never makes a
// network call and returns a canned judge transcript. The eval harness injects
// a real *aiclient.Client (CompleteJudge) in production.
type fakeJudgeModel struct {
	content     string
	err         error
	lastProfile string
	lastPayload aiclient.CompletePayload
}

func (f *fakeJudgeModel) CompleteJudge(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.lastProfile = profileName
	f.lastPayload = payload
	if f.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, f.err
	}
	return aiclient.CompleteResponse{Content: f.content}, aiclient.AICallMeta{Capability: aiclient.CapabilityJudge}, nil
}

func newRepoRegistryClient(t *testing.T) *Client {
	t.Helper()
	prompts, rubrics := repoConfigRoots(t)
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

const validFollowupOutput = `{"questionText":"Tell me more about the rollback you mentioned.","questionIntent":"probe-depth"}`

func TestLLMJudgeReturnsPerDimensionScores(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("practice.session.follow_up", "v0.1.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	model := &fakeJudgeModel{content: recordedJudgeTranscript(t, rubric, 0.8, "meets the dimensions")}

	judge, err := NewLLMJudge(reg, model, testJudgeInstruction)
	if err != nil {
		t.Fatalf("NewLLMJudge: %v", err)
	}
	var _ Judge = judge // compile-time interface assertion (item 2.6)

	scores, reasoning, err := judge.Judge(context.Background(), "practice.session.follow_up", "v0.1.0", []byte(validFollowupOutput), "v0.1.0")
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
