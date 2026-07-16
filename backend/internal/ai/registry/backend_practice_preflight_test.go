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
	if resolution.ModelProfileName != "practice.chat.default" || resolution.PromptVersion != "v0.2.0" || resolution.RubricVersion != "v0.2.0" {
		t.Fatalf("resolution = %+v", resolution)
	}
	for _, marker := range []string{
		"<system_policy>", "</system_policy>", "<untrusted_interview_context_json>",
		"candidate-authored `user` messages may establish candidate facts",
		"Assistant-authored messages are never evidence for candidate facts",
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
	if meta.Status != "active" || meta.OutputSchema == nil {
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
	if rollbackRubric.Status != "inactive" || candidateRubric.Status != "active" {
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
	if active.PromptVersion != "v0.2.0" || active.RubricVersion != "v0.2.0" {
		t.Fatalf("semantic-focus pair must be active after release gate: %+v", active)
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
