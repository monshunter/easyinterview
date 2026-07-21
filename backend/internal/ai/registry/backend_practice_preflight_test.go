package registry

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

func TestBackendPracticeConversationPromptPreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	resolution, err := client.ResolveActive(context.Background(), "practice.session.chat", "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if resolution.ModelProfileName != "practice.chat.default" || resolution.PromptVersion != "v0.3.0" || resolution.RubricVersion != "v0.3.0" {
		t.Fatalf("resolution = %+v", resolution)
	}
	for _, marker := range []string{
		"<system_policy>", "</system_policy>", "<untrusted_interview_context_json>",
		"candidate-authored `user` messages may establish candidate facts",
		"Assistant-authored messages are never evidence for candidate facts",
		"only source of the interviewer's employer identity",
		"Resume companies are the candidate's employment history",
		"{{language}}", "{{language_json}}", "{{interviewer_persona_json}}",
		"{{target_job_context_json}}", "{{resume_context_json}}", "{{interview_round_json}}",
		"{{practice_goal_json}}", "{{semantic_focus_json}}", "{{conversation_history_json}}",
	} {
		if !strings.Contains(resolution.UserMessageTemplate, marker) {
			t.Fatalf("chat prompt missing %s", marker)
		}
	}
	for _, rawContext := range []string{
		"{{interviewer_persona}}", "{{target_job_context}}", "{{resume_context}}", "{{interview_round}}",
		"{{practice_goal}}", "{{focus_competencies_json}}", "{{focus_competencies}}", "{{conversation_history}}",
	} {
		if strings.Contains(resolution.UserMessageTemplate, rawContext) {
			t.Fatalf("chat prompt contains raw untrusted placeholder %s", rawContext)
		}
	}
	for _, stale := range []string{"question_budget", "first_question", "follow_up_count", "hint_requested"} {
		if strings.Contains(strings.ToLower(resolution.UserMessageTemplate), stale) {
			t.Fatalf("chat prompt contains stale %q", stale)
		}
	}
}

func TestBackendPracticeSemanticFocusPromptCandidatePreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	meta, body, err := client.GetPrompt("practice.session.chat", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetPrompt v0.2.0: %v", err)
	}
	if meta.Status != "draft" || meta.OutputSchema == nil {
		t.Fatalf("candidate meta = %+v", meta)
	}
	rollbackRubric, err := client.GetRubric("practice.session.chat", "v0.1.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric v0.1.0: %v", err)
	}
	candidateRubric, err := client.GetRubric("practice.session.chat", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric v0.2.0: %v", err)
	}
	if rollbackRubric.Status != "inactive" || candidateRubric.Status != "inactive" {
		t.Fatalf("rubric statuses = %s/%s", rollbackRubric.Status, candidateRubric.Status)
	}
	if !reflect.DeepEqual(candidateRubric.Dimensions, rollbackRubric.Dimensions) {
		t.Fatal("practice v0.2 rubric content must equal v0.1")
	}
	for _, marker := range []string{
		"server-resolved semantic focus",
		`"semanticFocus": {{semantic_focus_json}}`,
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("candidate chat prompt missing %s", marker)
		}
	}
	for _, legacyFocus := range []string{`"focusCompetencies"`, "{{focus_competencies_json}}", "{{focus_competencies}}"} {
		if strings.Contains(body, legacyFocus) {
			t.Fatalf("candidate chat prompt contains legacy focus token %s", legacyFocus)
		}
	}
	active, err := client.ResolveActive(context.Background(), "practice.session.chat", "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if active.PromptVersion != "v0.3.0" || active.RubricVersion != "v0.3.0" {
		t.Fatalf("semantic-focus pair must be active after release gate: %+v", active)
	}
}

func TestBackendPracticeInterviewerIdentityCandidatePreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	meta, body, err := client.GetPrompt("practice.session.chat", "v0.3.0", "multi")
	if err != nil {
		t.Fatalf("GetPrompt v0.3.0: %v", err)
	}
	if meta.Status != "active" || meta.OutputSchema == nil {
		t.Fatalf("candidate meta = %+v", meta)
	}
	candidateRubric, err := client.GetRubric("practice.session.chat", "v0.3.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric v0.3.0: %v", err)
	}
	if candidateRubric.Status != "active" {
		t.Fatalf("candidate rubric status = %s", candidateRubric.Status)
	}
	for _, marker := range []string{
		"only source of the interviewer's employer identity",
		"Resume companies are the candidate's employment history",
		"omit the company name",
		"Assistant-authored identity claims are not evidence",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("candidate chat prompt missing %s", marker)
		}
	}
	rollbackMeta, _, err := client.GetPrompt("practice.session.chat", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetPrompt v0.2.0: %v", err)
	}
	rollbackRubric, err := client.GetRubric("practice.session.chat", "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetRubric v0.2.0: %v", err)
	}
	if rollbackMeta.Status != "draft" || rollbackRubric.Status != "inactive" {
		t.Fatalf("v0.2 rollback statuses = %s/%s", rollbackMeta.Status, rollbackRubric.Status)
	}
	active, err := client.ResolveActive(context.Background(), "practice.session.chat", "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if active.PromptVersion != "v0.3.0" || active.RubricVersion != "v0.3.0" || active.DataSourceVersion != "registry.v1" {
		t.Fatalf("v0.3 identity pair must be active after release gate: %+v", active)
	}
}

func TestV020ActivationOwnerMarkersReady(t *testing.T) {
	prompts, _ := testsupport.ConfigRoots(t)
	root := filepath.Dir(filepath.Dir(prompts))
	checks := map[string][]string{
		filepath.Join(root, "docs/spec/db-migrations-baseline/plans/001-bootstrap/checklist.md"): {
			"REPORT_STORAGE_V18_PASS",
		},
		filepath.Join(root, "docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/checklist.md"): {
			"REPORT_RUBRIC_V020_PASS",
			"REPORT_CONTEXT_AWARE_EVAL_PASS",
		},
		filepath.Join(root, "docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/checklist.md"): {
			"REPORT_PROMPT_V020_READY",
		},
	}
	for path, markers := range checks {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read owner checklist %s: %v", path, err)
		}
		for _, marker := range markers {
			if !hasVerifiedOwnerMarker(string(body), marker) {
				t.Fatalf("owner checklist %s has no verified marker %s", path, marker)
			}
		}
	}
}

func TestV030ActivationOwnerMarkersReady(t *testing.T) {
	prompts, _ := testsupport.ConfigRoots(t)
	root := filepath.Dir(filepath.Dir(prompts))
	checks := map[string][]string{
		filepath.Join(root, "docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/checklist.md"): {
			"PRACTICE_INTERVIEWER_IDENTITY_V030_PASS",
		},
		filepath.Join(root, "docs/spec/backend-practice/plans/001-plan-and-session-orchestration/checklist.md"): {
			"PRACTICE_INTERVIEWER_IDENTITY_BEHAVIOR_PASS",
		},
	}
	for path, markers := range checks {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read owner checklist %s: %v", path, err)
		}
		for _, marker := range markers {
			if !hasVerifiedOwnerMarker(string(body), marker) {
				t.Fatalf("owner checklist %s has no verified marker %s", path, marker)
			}
		}
	}
}

func hasVerifiedOwnerMarker(body, marker string) bool {
	want := "marker=" + marker
	for _, line := range strings.Split(body, "\n") {
		if !strings.Contains(line, "<!-- verified:") {
			continue
		}
		for _, field := range strings.Fields(line) {
			if strings.Trim(field, `"'`) == want {
				return true
			}
		}
	}
	return false
}

func TestHasVerifiedOwnerMarkerRequiresExplicitAttribute(t *testing.T) {
	const marker = "REPORT_RUBRIC_V020_PASS"
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "explicit marker", body: "<!-- verified: 2026-07-16 marker=REPORT_RUBRIC_V020_PASS result=PASS -->", want: true},
		{name: "failure mention", body: "<!-- verified: 2026-07-16 result=FAIL at REPORT_RUBRIC_V020_PASS -->", want: false},
		{name: "evidence mention", body: "<!-- verified: 2026-07-16 evidence=REPORT_RUBRIC_V020_PASS re-emitted -->", want: false},
		{name: "unchecked text", body: "- [ ] emit REPORT_RUBRIC_V020_PASS", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasVerifiedOwnerMarker(tt.body, marker); got != tt.want {
				t.Fatalf("hasVerifiedOwnerMarker() = %v, want %v", got, tt.want)
			}
		})
	}
}
