package registry

import (
	"context"
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
	if resolution.ModelProfileName != "practice.chat.default" || resolution.PromptVersion != "v0.1.0" || resolution.RubricVersion != "v0.1.0" {
		t.Fatalf("resolution = %+v", resolution)
	}
	for _, marker := range []string{
		"<system_policy>", "</system_policy>", "<untrusted_interview_context_json>",
		"candidate-authored `user` messages may establish candidate facts",
		"Assistant-authored messages are never evidence for candidate facts",
		"{{language}}", "{{language_json}}", "{{interviewer_persona_json}}",
		"{{target_job_context_json}}", "{{resume_context_json}}", "{{interview_round_json}}",
		"{{practice_goal_json}}", "{{focus_competencies_json}}", "{{conversation_history_json}}",
	} {
		if !strings.Contains(resolution.UserMessageTemplate, marker) {
			t.Fatalf("chat prompt missing %s", marker)
		}
	}
	for _, rawContext := range []string{
		"{{interviewer_persona}}", "{{target_job_context}}", "{{resume_context}}", "{{interview_round}}",
		"{{practice_goal}}", "{{focus_competencies}}", "{{conversation_history}}",
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
