package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

func TestBackendPracticeF3Preflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	repoRoot := filepath.Dir(filepath.Dir(prompts))
	assertCompletedDocHeader(t, filepath.Join(repoRoot, "docs", "spec", "prompt-rubric-registry", "plans", "001-baseline", "plan.md"))
	assertCompletedDocHeader(t, filepath.Join(repoRoot, "docs", "spec", "prompt-rubric-registry", "plans", "001-baseline", "checklist.md"))

	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}

	for _, tc := range []struct {
		featureKey       string
		modelProfileName string
	}{
		{featureKey: "practice.session.first_question", modelProfileName: "practice.first_question.default"},
		{featureKey: "practice.session.follow_up", modelProfileName: "practice.followup.default"},
		{featureKey: "practice.turn.lightweight_observe", modelProfileName: "practice.turn_observe.default"},
	} {
		t.Run(tc.featureKey, func(t *testing.T) {
			res, err := client.ResolveActive(context.Background(), tc.featureKey, "en")
			if err != nil {
				t.Fatalf("ResolveActive: %v", err)
			}
			if res.FeatureKey != tc.featureKey {
				t.Fatalf("FeatureKey: want %s, got %s", tc.featureKey, res.FeatureKey)
			}
			if res.PromptVersion == "" || res.RubricVersion == "" || res.ModelProfileName == "" {
				t.Fatalf("ResolveActive returned incomplete triple: %+v", res)
			}
			if res.ModelProfileName != tc.modelProfileName {
				t.Fatalf("ModelProfileName: want %s, got %s", tc.modelProfileName, res.ModelProfileName)
			}
			if res.PromptVersion != "v0.1.0" || res.RubricVersion != "v0.1.0" {
				t.Fatalf("ResolveActive versions: prompt=%s rubric=%s", res.PromptVersion, res.RubricVersion)
			}
			if res.DataSourceVersion != "registry.v1" || res.FeatureFlag != "none" {
				t.Fatalf("ResolveActive provenance: dataSourceVersion=%s featureFlag=%s", res.DataSourceVersion, res.FeatureFlag)
			}
			if res.UserMessageTemplate == "" {
				t.Fatal("ResolveActive returned empty prompt body")
			}
			if tc.featureKey == "practice.session.follow_up" {
				for _, marker := range []string{
					"{{language}}",
					"{{generation_kind}}",
					"{{attempt_mode}}",
					"{{practice_goal}}",
					"{{practice_mode}}",
					"{{turn_status}}",
					"{{target_job_id}}",
					"{{last_question}}",
					"{{question_intent}}",
					"{{last_answer}}",
					"{{follow_up_count}}",
					"{{covered_dimensions}}",
					"{{remaining_dimensions}}",
					"{{committed_context}}",
				} {
					if !strings.Contains(res.UserMessageTemplate, marker) {
						t.Fatalf("follow-up prompt missing canonical runtime marker %s", marker)
					}
				}
			}
		})
	}
}

func assertCompletedDocHeader(t *testing.T, path string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "> **状态**: completed") {
		t.Fatalf("%s header must be completed", path)
	}
}
