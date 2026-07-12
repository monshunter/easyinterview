package registry

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// TestLLMJudgeFailClose covers the negative fail-close paths: the judge never
// silently zero-fills, and a structurally invalid evaluated output or judge
// transcript is rejected (plan 004 §2.5).
func TestLLMJudgeFailClose(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("practice.session.chat", "v0.1.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	goodTranscript := recordedJudgeTranscript(t, rubric, 0.7, "ok")

	t.Run("judge profile unavailable propagates error", func(t *testing.T) {
		model := &fakeJudgeModel{err: errors.New("AI_UNSUPPORTED_CAPABILITY: judge.default not active")}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if err == nil {
			t.Fatal("expected error when judge model fails")
		}
	})

	t.Run("evaluated output failing schema fail-closes", func(t *testing.T) {
		model := &fakeJudgeModel{content: goodTranscript}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(`{"foo":"bar"}`), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for invalid evaluated output, got %v", err)
		}
		if model.lastProfile != "" {
			t.Fatal("judge model must not be called when evaluated output is invalid")
		}
	})

	t.Run("unparseable judge response fail-closes", func(t *testing.T) {
		model := &fakeJudgeModel{content: "not json at all"}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for unparseable response, got %v", err)
		}
	})

	t.Run("dimension count mismatch fail-closes", func(t *testing.T) {
		partial, _ := json.Marshal(map[string]any{
			"scores":    []map[string]any{{"dimension": rubric.Dimensions[0].Name, "value": 0.5}},
			"reasoning": map[string]any{"summary": "partial"},
		})
		model := &fakeJudgeModel{content: string(partial)}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for dimension count mismatch, got %v", err)
		}
	})

	t.Run("unknown dimension fail-closes", func(t *testing.T) {
		scores := make([]map[string]any, 0, len(rubric.Dimensions))
		for i, d := range rubric.Dimensions {
			name := d.Name
			if i == 0 {
				name = "not_a_real_dimension"
			}
			scores = append(scores, map[string]any{"dimension": name, "value": 0.5})
		}
		raw, _ := json.Marshal(map[string]any{"scores": scores, "reasoning": map[string]any{"summary": "x"}})
		model := &fakeJudgeModel{content: string(raw)}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for unknown dimension, got %v", err)
		}
	})

	t.Run("value out of range fail-closes", func(t *testing.T) {
		scores := make([]map[string]any, 0, len(rubric.Dimensions))
		for i, d := range rubric.Dimensions {
			v := 0.5
			if i == 0 {
				v = 1.7
			}
			scores = append(scores, map[string]any{"dimension": d.Name, "value": v})
		}
		raw, _ := json.Marshal(map[string]any{"scores": scores, "reasoning": map[string]any{"summary": "x"}})
		model := &fakeJudgeModel{content: string(raw)}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for out-of-range value, got %v", err)
		}
	})

	t.Run("empty reasoning summary fail-closes", func(t *testing.T) {
		model := &fakeJudgeModel{content: recordedJudgeTranscript(t, rubric, 0.6, "")}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "practice.session.chat", "v0.1.0", []byte(validPracticeChatOutput), "v0.1.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) {
			t.Fatalf("want ErrJudgeOutputInvalid for empty summary, got %v", err)
		}
	})
}

func TestNewLLMJudgeRejectsMissingDeps(t *testing.T) {
	reg := newRepoRegistryClient(t)
	if _, err := NewLLMJudge(nil, &fakeJudgeModel{}, testJudgeInstruction); !errors.Is(err, ErrJudgeNotConfigured) {
		t.Fatalf("want ErrJudgeNotConfigured for nil registry, got %v", err)
	}
	if _, err := NewLLMJudge(reg, nil, testJudgeInstruction); !errors.Is(err, ErrJudgeNotConfigured) {
		t.Fatalf("want ErrJudgeNotConfigured for nil model, got %v", err)
	}
	if _, err := NewLLMJudge(reg, &fakeJudgeModel{}, ""); !errors.Is(err, ErrJudgeNotConfigured) {
		t.Fatalf("want ErrJudgeNotConfigured for empty instruction, got %v", err)
	}
}

func mustJudge(t *testing.T, reg RubricProvider, model JudgeModelClient) *LLMJudge {
	t.Helper()
	j, err := NewLLMJudge(reg, model, testJudgeInstruction)
	if err != nil {
		t.Fatalf("NewLLMJudge: %v", err)
	}
	return j
}
