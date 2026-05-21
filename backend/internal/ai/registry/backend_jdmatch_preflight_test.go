package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

func TestBackendJDMatchF3Preflight(t *testing.T) {
	prompts, rubrics := repoConfigRoots(t)
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
		requiredPrompt   []string
	}{
		{
			featureKey:       featurekeys.JdMatchRecommendation,
			modelProfileName: "jd_match.recommendation.default",
			requiredPrompt:   []string{"jobMatchId", "{{candidate_profile}}", "{{jobs_pool}}"},
		},
		{
			featureKey:       featurekeys.JdMatchSearch,
			modelProfileName: "jd_match.search.default",
			requiredPrompt:   []string{"jobMatchId", "{{query}}", "{{filters}}", "{{jobs_pool}}"},
		},
	} {
		t.Run(tc.featureKey.String(), func(t *testing.T) {
			res, err := client.ResolveActive(context.Background(), tc.featureKey.String(), "en")
			if err != nil {
				t.Fatalf("ResolveActive: %v", err)
			}
			if res.FeatureKey != tc.featureKey.String() {
				t.Fatalf("FeatureKey: want %s, got %s", tc.featureKey, res.FeatureKey)
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
			for _, want := range tc.requiredPrompt {
				if !strings.Contains(res.UserMessageTemplate, want) {
					t.Fatalf("%s prompt missing %q: %s", tc.featureKey, want, res.UserMessageTemplate)
				}
			}
		})
	}
}
