package registry

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

func TestF3ReportGenerateAndAssessmentPreflight(t *testing.T) {
	prompts, rubrics := repoConfigRoots(t)
	repoRoot := filepath.Dir(filepath.Dir(prompts))
	assertFileContains(t, filepath.Join(repoRoot, "docs", "spec", "prompt-rubric-registry", "spec.md"), "Prompt Rubric Registry Spec", "> **版本**: 2.12")
	assertCompletedDocHeader(t, filepath.Join(repoRoot, "docs", "spec", "prompt-rubric-registry", "plans", "001-baseline", "plan.md"))
	assertCompletedDocHeader(t, filepath.Join(repoRoot, "docs", "spec", "prompt-rubric-registry", "plans", "001-baseline", "checklist.md"))
	assertWorkJournalContains(t, filepath.Join(repoRoot, "docs", "work-journal"), "docs(prompt-rubric-registry): close 001-baseline lifecycle and record ac self-check")

	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}

	for _, tc := range []struct {
		featureKey       featurekeys.FeatureKey
		modelProfileName string
	}{
		{featureKey: featurekeys.ReportGenerate, modelProfileName: "report.generate.default"},
		{featureKey: featurekeys.ReportQuestionAssessment, modelProfileName: "report.assessment.default"},
	} {
		t.Run(string(tc.featureKey), func(t *testing.T) {
			res, err := client.ResolveActive(context.Background(), string(tc.featureKey), "en")
			if err != nil {
				t.Fatalf("ResolveActive: %v", err)
			}
			if res.FeatureKey != string(tc.featureKey) {
				t.Fatalf("FeatureKey: want %s, got %s", tc.featureKey, res.FeatureKey)
			}
			if res.PromptVersion != "v0.1.0" || res.RubricVersion != "v0.1.0" {
				t.Fatalf("ResolveActive versions: prompt=%s rubric=%s", res.PromptVersion, res.RubricVersion)
			}
			if res.ModelProfileName != tc.modelProfileName {
				t.Fatalf("ModelProfileName: want %s, got %s", tc.modelProfileName, res.ModelProfileName)
			}
			if res.DataSourceVersion != "registry.v1" || res.FeatureFlag != "none" {
				t.Fatalf("ResolveActive provenance: dataSourceVersion=%s featureFlag=%s", res.DataSourceVersion, res.FeatureFlag)
			}
			assertReportPromptSafeInputContract(t, res.UserMessageTemplate)
			assertReportPromptOutputContract(t, string(tc.featureKey), res.UserMessageTemplate)
			if tc.featureKey == featurekeys.ReportQuestionAssessment {
				assertQuestionAssessmentReviewStatusEnum(t, res.OutputSchema)
			}

			rubric, err := client.GetRubric(string(tc.featureKey), "v0.1.0", "multi")
			if err != nil {
				t.Fatalf("GetRubric: %v", err)
			}
			for _, dim := range rubric.Dimensions {
				assertFourScoreLevels(t, dim.ScoreLevels)
			}
		})
	}
}

func assertReportPromptOutputContract(t *testing.T, featureKey string, prompt string) {
	t.Helper()
	lower := strings.ToLower(prompt)
	var required []string
	switch featureKey {
	case string(featurekeys.ReportGenerate):
		required = []string{
			"summary",
			"dimension_scores",
			"highlights",
			"issues",
			"next_actions",
			"retry_focus_turn_ids",
		}
	case string(featurekeys.ReportQuestionAssessment):
		required = []string{
			"dimension_results",
			"overall_status",
			"confidence",
			"score_level",
			"strengths",
			"gaps",
			"recommended_framework",
			"review_status",
			"open",
			"queued_for_retry",
			"resolved",
		}
	default:
		t.Fatalf("unexpected report feature key %q", featureKey)
	}
	for _, want := range required {
		if !strings.Contains(lower, want) {
			t.Fatalf("%s prompt missing required output key %q: %s", featureKey, want, prompt)
		}
	}
	if featureKey == string(featurekeys.ReportQuestionAssessment) && strings.Contains(lower, `"review_status": "ready"`) {
		t.Fatalf("%s prompt must not use non-B1 review_status example %q: %s", featureKey, "ready", prompt)
	}
}

func assertQuestionAssessmentReviewStatusEnum(t *testing.T, raw *json.RawMessage) {
	t.Helper()
	if raw == nil {
		t.Fatal("report.question_assessment schema must be populated")
	}
	var schema struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(*raw, &schema); err != nil {
		t.Fatalf("unmarshal report.question_assessment schema: %v", err)
	}
	got := schema.Properties["review_status"].Enum
	want := []string{"open", "queued_for_retry", "resolved"}
	if len(got) != len(want) {
		t.Fatalf("review_status enum = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("review_status enum = %#v, want %#v", got, want)
		}
	}
}

func assertFileContains(t *testing.T, path string, wants ...string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("%s missing %q", path, want)
		}
	}
}

func assertWorkJournalContains(t *testing.T, root string, want string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), want) {
			return
		}
	}
	t.Fatalf("%s missing work-journal evidence %q", root, want)
}

func assertReportPromptSafeInputContract(t *testing.T, prompt string) {
	t.Helper()
	if strings.TrimSpace(prompt) == "" {
		t.Fatal("prompt body is empty")
	}
	lower := strings.ToLower(prompt)
	for _, forbidden := range []string{
		"{{transcript}}",
		"{{question}}",
		"{{answer}}",
		"transcript:",
		"answer transcript:",
		"evidence_quotes",
		"verbatim",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("report prompt must not require raw interview input or quote output %q: %s", forbidden, prompt)
		}
	}
	for _, required := range []string{
		"session metadata",
		"turn summaries",
		"rubric",
		"strict json",
	} {
		if !strings.Contains(lower, required) {
			t.Fatalf("report prompt missing safe-input/output contract %q: %s", required, prompt)
		}
	}
}

func assertFourScoreLevels(t *testing.T, levels []ScoreLevel) {
	t.Helper()
	if len(levels) != 4 {
		t.Fatalf("score_levels must contain exactly four levels, got %d", len(levels))
	}
	wantLabels := []string{"weak", "developing", "proficient", "strong"}
	for i, want := range wantLabels {
		if levels[i].Label != want {
			t.Fatalf("score_levels[%d].label: want %s, got %s", i, want, levels[i].Label)
		}
	}
}
