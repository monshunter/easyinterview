package registry

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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

func TestLLMJudgeGroundedReportFailClose(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("report.generate", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	validTranscript := recordedGroundedJudgeTranscript(t, rubric)
	evidence := JudgeContext{
		FrozenContext: json.RawMessage(`{"targetJobTitle":"Backend Engineer"}`),
		Transcript:    json.RawMessage(`[{"seqNo":2,"role":"user","content":"I added retry limits."}]`),
	}

	t.Run("missing context fails before model", func(t *testing.T) {
		model := &fakeJudgeModel{content: validTranscript}
		judge := mustJudge(t, reg, model)
		_, _, err := judge.Judge(context.Background(), "report.generate", "v0.2.0", []byte(validGroundedReportOutput), "v0.2.0")
		if !errors.Is(err, ErrJudgeOutputInvalid) || model.lastProfile != "" {
			t.Fatalf("missing context error=%v modelProfile=%q", err, model.lastProfile)
		}
	})

	mutations := map[string]func(map[string]any){
		"missing item verdict": func(env map[string]any) {
			items := env["item_verdicts"].([]any)
			env["item_verdicts"] = items[:len(items)-1]
		},
		"unsupported item": func(env map[string]any) {
			env["item_verdicts"].([]any)[0].(map[string]any)["support"] = "unsupported"
		},
		"partial used negatively": func(env map[string]any) {
			item := env["item_verdicts"].([]any)[0].(map[string]any)
			item["support"] = "partial"
			item["evidence_limited_explicit"] = true
			item["used_for_negative_claim"] = true
		},
		"causal mismatch": func(env map[string]any) {
			env["causal_checks"].([]any)[0].(map[string]any)["action_supported"] = false
		},
		"zero tolerance violation": func(env map[string]any) {
			env["zero_tolerance_violations"] = []any{"fabricated_fact"}
		},
		"critical safety miss": func(env map[string]any) {
			env["critical_safety_pass"] = false
		},
		"unknown judge field": func(env map[string]any) {
			env["unexpected"] = true
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var env map[string]any
			if err := json.Unmarshal([]byte(validTranscript), &env); err != nil {
				t.Fatalf("parse valid transcript: %v", err)
			}
			mutate(env)
			raw, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal mutated transcript: %v", err)
			}
			model := &fakeJudgeModel{content: string(raw)}
			judge := mustJudge(t, reg, model)
			_, _, err = judge.JudgeWithContext(context.Background(), "report.generate", "v0.2.0", []byte(validGroundedReportOutput), "v0.2.0", evidence)
			if !errors.Is(err, ErrJudgeOutputInvalid) {
				t.Fatalf("want ErrJudgeOutputInvalid, got %v", err)
			}
		})
	}

	preparednessCases := []struct {
		name    string
		mutate  func(map[string]any)
		wantErr string
	}{
		{
			name: "missing preparedness verdict",
			mutate: func(env map[string]any) {
				items := env["item_verdicts"].([]any)
				for i, item := range items {
					if item.(map[string]any)["path"] == "$.preparednessLevel" {
						env["item_verdicts"] = append(items[:i], items[i+1:]...)
						return
					}
				}
				t.Fatal("preparedness verdict fixture is missing")
			},
			wantErr: "does not cover",
		},
		{
			name: "unsupported preparedness verdict",
			mutate: func(env map[string]any) {
				groundedVerdictByPath(t, env, "$.preparednessLevel")["support"] = "unsupported"
			},
			wantErr: `unsupported report item "$.preparednessLevel"`,
		},
		{
			name: "partial negative preparedness verdict",
			mutate: func(env map[string]any) {
				item := groundedVerdictByPath(t, env, "$.preparednessLevel")
				item["support"] = "partial"
				item["evidence_limited_explicit"] = true
				item["used_for_negative_claim"] = true
			},
			wantErr: `partial item "$.preparednessLevel" must be explicitly evidence-limited and non-negative`,
		},
	}
	for _, tc := range preparednessCases {
		t.Run(tc.name, func(t *testing.T) {
			var env map[string]any
			if err := json.Unmarshal([]byte(validTranscript), &env); err != nil {
				t.Fatalf("parse valid transcript: %v", err)
			}
			tc.mutate(env)
			raw, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal mutated transcript: %v", err)
			}
			model := &fakeJudgeModel{content: string(raw)}
			judge := mustJudge(t, reg, model)
			_, _, err = judge.JudgeWithContext(context.Background(), "report.generate", "v0.2.0", []byte(validGroundedReportOutput), "v0.2.0", evidence)
			if !errors.Is(err, ErrJudgeOutputInvalid) || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want ErrJudgeOutputInvalid containing %q, got %v", tc.wantErr, err)
			}
		})
	}

	t.Run("empty retry focus still requires verdict", func(t *testing.T) {
		var env map[string]any
		if err := json.Unmarshal([]byte(validTranscript), &env); err != nil {
			t.Fatalf("parse valid transcript: %v", err)
		}
		removeGroundedVerdictByPath(t, env, "$.retryFocusDimensionCodes")
		var output map[string]any
		if err := json.Unmarshal([]byte(validGroundedReportOutput), &output); err != nil {
			t.Fatalf("parse report output: %v", err)
		}
		output["retryFocusDimensionCodes"] = []any{}
		rawOutput, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("marshal report output: %v", err)
		}
		rawJudge, err := json.Marshal(env)
		if err != nil {
			t.Fatalf("marshal judge transcript: %v", err)
		}
		model := &fakeJudgeModel{content: string(rawJudge)}
		judge := mustJudge(t, reg, model)
		_, _, err = judge.JudgeWithContext(context.Background(), "report.generate", "v0.2.0", rawOutput, "v0.2.0", evidence)
		if !errors.Is(err, ErrJudgeOutputInvalid) || !strings.Contains(err.Error(), "does not cover") {
			t.Fatalf("want missing retry focus verdict to fail-close, got %v", err)
		}
	})

	retryFocusCases := []struct {
		name    string
		mutate  func(map[string]any)
		wantErr string
	}{
		{
			name: "unsupported retry focus verdict",
			mutate: func(env map[string]any) {
				groundedVerdictByPath(t, env, "$.retryFocusDimensionCodes")["support"] = "unsupported"
			},
			wantErr: `unsupported report item "$.retryFocusDimensionCodes"`,
		},
		{
			name: "partial negative retry focus verdict",
			mutate: func(env map[string]any) {
				item := groundedVerdictByPath(t, env, "$.retryFocusDimensionCodes")
				item["support"] = "partial"
				item["evidence_limited_explicit"] = true
				item["used_for_negative_claim"] = true
			},
			wantErr: `partial item "$.retryFocusDimensionCodes" must be explicitly evidence-limited and non-negative`,
		},
	}
	for _, tc := range retryFocusCases {
		t.Run(tc.name, func(t *testing.T) {
			var env map[string]any
			if err := json.Unmarshal([]byte(validTranscript), &env); err != nil {
				t.Fatalf("parse valid transcript: %v", err)
			}
			tc.mutate(env)
			raw, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal mutated transcript: %v", err)
			}
			model := &fakeJudgeModel{content: string(raw)}
			judge := mustJudge(t, reg, model)
			_, _, err = judge.JudgeWithContext(context.Background(), "report.generate", "v0.2.0", []byte(validGroundedReportOutput), "v0.2.0", evidence)
			if !errors.Is(err, ErrJudgeOutputInvalid) || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want ErrJudgeOutputInvalid containing %q, got %v", tc.wantErr, err)
			}
		})
	}

	t.Run("unknown report output field fails before model", func(t *testing.T) {
		var output map[string]any
		if err := json.Unmarshal([]byte(validGroundedReportOutput), &output); err != nil {
			t.Fatalf("parse output: %v", err)
		}
		output["dimension_scores"] = []any{}
		raw, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("marshal output: %v", err)
		}
		model := &fakeJudgeModel{content: validTranscript}
		judge := mustJudge(t, reg, model)
		_, _, err = judge.JudgeWithContext(context.Background(), "report.generate", "v0.2.0", raw, "v0.2.0", evidence)
		if !errors.Is(err, ErrJudgeOutputInvalid) || model.lastProfile != "" {
			t.Fatalf("unknown output error=%v modelProfile=%q", err, model.lastProfile)
		}
	})
}

func TestLLMJudgeGroundedContentRejectionIsTerminalWithoutRetry(t *testing.T) {
	reg := newRepoRegistryClient(t)
	rubric, err := reg.GetRubric("report.generate", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	mutations := map[string]func(map[string]any){
		"unsupported item": func(env map[string]any) {
			groundedVerdictByPath(t, env, "$.issues[0]")["support"] = "unsupported"
		},
		"causal false": func(env map[string]any) {
			env["causal_checks"].([]any)[0].(map[string]any)["action_supported"] = false
		},
		"zero tolerance": func(env map[string]any) {
			env["zero_tolerance_violations"] = []any{"fabricated_fact"}
		},
		"critical false": func(env map[string]any) {
			env["critical_safety_pass"] = false
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var env map[string]any
			if err := json.Unmarshal([]byte(recordedGroundedJudgeTranscript(t, rubric)), &env); err != nil {
				t.Fatalf("parse judge transcript: %v", err)
			}
			mutate(env)
			raw, err := json.Marshal(env)
			if err != nil {
				t.Fatalf("marshal judge transcript: %v", err)
			}
			model := &fakeJudgeModel{content: string(raw)}
			judge := mustJudge(t, reg, model)
			_, _, err = judge.JudgeWithContext(
				context.Background(),
				"report.generate",
				"v0.2.0",
				[]byte(validGroundedReportOutput),
				"v0.2.0",
				JudgeContext{
					FrozenContext: json.RawMessage(`{"targetJobTitle":"Backend Engineer"}`),
					Transcript:    json.RawMessage(`[{"seqNo":2,"role":"user","content":"I added retry limits."}]`),
				},
			)
			if !errors.Is(err, ErrJudgeContentRejected) || model.calls != 1 {
				t.Fatalf("error=%v calls=%d, want typed content rejection after exactly 1 call", err, model.calls)
			}
		})
	}
}

func TestGroundedReportEmptyRetryFocusRequiresVerdict(t *testing.T) {
	report, err := decodeGroundedReportOutput([]byte(validGroundedReportOutput))
	if err != nil {
		t.Fatalf("decode report: %v", err)
	}
	report.RetryFocusDimensionCodes = []string{}

	var transcript map[string]any
	if err := json.Unmarshal([]byte(recordedGroundedJudgeTranscript(t, RubricSchema{})), &transcript); err != nil {
		t.Fatalf("parse judge transcript: %v", err)
	}
	removeGroundedVerdictByPath(t, transcript, "$.retryFocusDimensionCodes")
	raw, err := json.Marshal(transcript)
	if err != nil {
		t.Fatalf("marshal judge transcript: %v", err)
	}
	var env judgeResponseEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode judge envelope: %v", err)
	}

	err = validateGroundedReportVerdict(env, report, &Reasoning{})
	if !errors.Is(err, ErrJudgeOutputInvalid) || !strings.Contains(err.Error(), "does not cover") {
		t.Fatalf("want empty retry focus to require its verdict, got %v", err)
	}
}

func groundedVerdictByPath(t *testing.T, env map[string]any, path string) map[string]any {
	t.Helper()
	for _, item := range env["item_verdicts"].([]any) {
		verdict := item.(map[string]any)
		if verdict["path"] == path {
			return verdict
		}
	}
	t.Fatalf("item verdict %q is missing", path)
	return nil
}

func removeGroundedVerdictByPath(t *testing.T, env map[string]any, path string) {
	t.Helper()
	items := env["item_verdicts"].([]any)
	for i, item := range items {
		if item.(map[string]any)["path"] == path {
			env["item_verdicts"] = append(items[:i], items[i+1:]...)
			return
		}
	}
	t.Fatalf("item verdict %q is missing", path)
}

func mustJudge(t *testing.T, reg RubricProvider, model JudgeModelClient) *LLMJudge {
	t.Helper()
	j, err := NewLLMJudge(reg, model, testJudgeInstruction)
	if err != nil {
		t.Fatalf("NewLLMJudge: %v", err)
	}
	return j
}
