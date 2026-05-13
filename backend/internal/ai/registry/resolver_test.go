package registry

import (
	"context"
	"errors"
	"testing"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()
	prompts, rubrics := repoConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	return client
}

func TestResolveExactLanguage(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	res, err := client.ResolveActive(ctx, "target.import.parse", "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if res.PromptVersion != "v0.1.0" {
		t.Errorf("PromptVersion: want v0.1.0, got %s", res.PromptVersion)
	}
	if res.RubricVersion != "v0.1.0" {
		t.Errorf("RubricVersion: want v0.1.0, got %s", res.RubricVersion)
	}
	if res.ModelProfileName != "target.import.default" {
		t.Errorf("ModelProfileName: want target.import.default, got %s", res.ModelProfileName)
	}
	if res.FeatureFlag != "none" {
		t.Errorf("FeatureFlag: want 'none', got %q", res.FeatureFlag)
	}
	if res.UserMessageTemplate == "" {
		t.Errorf("UserMessageTemplate must be populated for plan 001 baseline")
	}
	if got := client.FallbackCount(); got != 0 {
		t.Errorf("FallbackCount: exact-language hit must not increment, got %d", got)
	}
}

func TestResolveFallbackToMulti(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	res, err := client.ResolveActive(ctx, "report.generate", "fr")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if res.PromptVersion != "v0.1.0" {
		t.Errorf("PromptVersion: want v0.1.0, got %s", res.PromptVersion)
	}
	if got := client.FallbackCount(); got != 1 {
		t.Errorf("FallbackCount: language fallback must increment, got %d", got)
	}
}

func TestResolvePracticeSessionBaselineFeatures(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	cases := map[string]string{
		"practice.session.first_question":   "practice.first_question.default",
		"practice.session.follow_up":        "practice.followup.default",
		"practice.turn.lightweight_observe": "practice.turn_observe.default",
	}
	for featureKey, profileName := range cases {
		t.Run(featureKey, func(t *testing.T) {
			res, err := client.ResolveActive(ctx, featureKey, "en")
			if err != nil {
				t.Fatalf("ResolveActive: %v", err)
			}
			if res.FeatureKey != featureKey {
				t.Errorf("FeatureKey: want %s, got %s", featureKey, res.FeatureKey)
			}
			if res.PromptVersion != "v0.1.0" || res.RubricVersion != "v0.1.0" {
				t.Errorf("versions: want prompt/rubric v0.1.0, got %s/%s", res.PromptVersion, res.RubricVersion)
			}
			if res.ModelProfileName != profileName {
				t.Errorf("ModelProfileName: want %s, got %s", profileName, res.ModelProfileName)
			}
			if res.UserMessageTemplate == "" {
				t.Errorf("UserMessageTemplate must be populated for plan 001 baseline")
			}
		})
	}
}

func TestResolveUnknownFeatureKey(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	_, err := client.ResolveActive(ctx, "no.such.feature", "en")
	if !errors.Is(err, ErrPromptUnsupported) {
		t.Fatalf("want ErrPromptUnsupported, got %v", err)
	}
}

func TestResolveEmptyArgsRejected(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	if _, err := client.ResolveActive(ctx, "", "en"); !errors.Is(err, ErrPromptUnsupported) {
		t.Errorf("empty featureKey: want ErrPromptUnsupported, got %v", err)
	}
	if _, err := client.ResolveActive(ctx, "report.generate", ""); !errors.Is(err, ErrLanguageUnsupported) {
		t.Errorf("empty language: want ErrLanguageUnsupported, got %v", err)
	}
}

func TestGetPromptExact(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)

	meta, body, err := client.GetPrompt("target.import.parse", "v0.1.0", "en")
	if err != nil {
		t.Fatalf("GetPrompt: %v", err)
	}
	if meta.Version != "v0.1.0" || body == "" {
		t.Errorf("expected populated meta+body, got version=%q body_len=%d", meta.Version, len(body))
	}

	if _, _, err := client.GetPrompt("", "v0.1.0", "en"); err == nil {
		t.Error("empty featureKey must error")
	}
	if _, _, err := client.GetPrompt("target.import.parse", "v9.9.9", "en"); !errors.Is(err, ErrPromptUnsupported) {
		t.Errorf("unknown version: want ErrPromptUnsupported, got %v", err)
	}
}

func TestGetRubricExact(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)

	rs, err := client.GetRubric("target.import.parse", "v0.1.0", "en")
	if err != nil {
		t.Fatalf("GetRubric: %v", err)
	}
	if len(rs.Dimensions) == 0 {
		t.Errorf("expected at least one dimension")
	}
}
