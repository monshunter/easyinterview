package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

func TestReportGenerateConversationContractPreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	resolution, err := client.ResolveActive(context.Background(), string(featurekeys.ReportGenerate), "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if resolution.ModelProfileName != "report.generate.default" || resolution.OutputSchema == nil {
		t.Fatalf("resolution = %+v", resolution)
	}
	lower := strings.ToLower(resolution.UserMessageTemplate)
	for _, required := range []string{"conversation_messages", "dimension_scores", "highlights", "issues", "next_actions", "retry_focus_competency_codes"} {
		if !strings.Contains(lower, required) {
			t.Fatalf("report prompt missing %q", required)
		}
	}
	for _, stale := range []string{"question_assessment", "retry_focus_turn_ids", "turn_summaries"} {
		if strings.Contains(lower, stale) {
			t.Fatalf("report prompt contains stale %q", stale)
		}
	}
}
