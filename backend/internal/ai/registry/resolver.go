package registry

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

// resolveSnapshot looks up the active prompt+rubric pair for a feature_key
// and language and returns the merged PromptResolution. fallback bumps the
// fallback counter when an exact-language miss falls through to multi.
func resolveSnapshot(snap *snapshot, featureKey, language string, fallbackCounter *atomic.Uint64) (PromptResolution, error) {
	if strings.TrimSpace(featureKey) == "" {
		return PromptResolution{}, ErrPromptUnsupported
	}
	if strings.TrimSpace(language) == "" {
		return PromptResolution{}, ErrLanguageUnsupported
	}

	prompts, ok := snap.prompts[featureKey]
	if !ok {
		return PromptResolution{}, ErrPromptUnsupported
	}
	rubrics, ok := snap.rubrics[featureKey]
	if !ok {
		return PromptResolution{}, ErrPromptUnsupported
	}

	promptVersions, lang, ok := selectByLanguage(prompts, language)
	if !ok {
		return PromptResolution{}, ErrLanguageUnsupported
	}
	rubricVersions, _, ok := selectByLanguage(rubrics, lang)
	if !ok {
		// Language parity is enforced at load time, so this should be
		// unreachable in practice; return a precise error if it ever fires.
		return PromptResolution{}, ErrLanguageUnsupported
	}

	if lang != language && fallbackCounter != nil {
		fallbackCounter.Add(1)
	}

	// Current baselines keep SystemMessage empty and put the full markdown body
	// in UserMessageTemplate, which existing executors consume directly.
	// Feature-specific system/user splits must be introduced by the owning spec.
	pe, ok := activePrompt(promptVersions)
	if !ok {
		return PromptResolution{}, ErrPromptUnsupported
	}
	re, ok := activeRubric(rubricVersions)
	if !ok {
		return PromptResolution{}, ErrPromptUnsupported
	}
	return PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       pe.meta.Version,
		RubricVersion:       re.schema.Version,
		ModelProfileName:    defaultModelProfile(featureKey),
		DataSourceVersion:   resolvedDataSourceVersion(featureKey, pe.meta.Version),
		FeatureFlag:         "none",
		SystemMessage:       "",
		UserMessageTemplate: pe.body,
		Tools:               nil,
		OutputSchema:        pe.outputSchema,
		StreamWire:          nil,
	}, nil
}

func resolvedDataSourceVersion(featureKey, promptVersion string) string {
	if featurekeys.FeatureKey(featureKey) == featurekeys.ReportGenerate && promptVersion == "v0.2.0" {
		return "report-context.v1"
	}
	return "registry.v1"
}

// selectByLanguage returns the entry for the requested language, falling
// back to "multi" when the exact language is missing. The returned string
// is the language actually used (so callers can tell whether fallback fired).
func selectByLanguage[T any](entries map[string]T, requested string) (T, string, bool) {
	if e, ok := entries[requested]; ok {
		return e, requested, true
	}
	if e, ok := entries["multi"]; ok {
		return e, "multi", true
	}
	return *new(T), "", false
}

func activePrompt(entries map[string]promptEntry) (promptEntry, bool) {
	for _, entry := range entries {
		if entry.meta.Status == "active" {
			return entry, true
		}
	}
	return promptEntry{}, false
}

func activeRubric(entries map[string]rubricEntry) (rubricEntry, bool) {
	for _, entry := range entries {
		if entry.schema.Status == "active" {
			return entry, true
		}
	}
	return rubricEntry{}, false
}

// defaultModelProfile maps a feature_key to its spec §3.1.1 default
// model_profile_name. Resolve does not look up the A3 profile catalog;
// callers translate the profile name through their own A3 wiring.
//
// All feature_key strings are sourced through the featurekeys package so the
// events lint gate ("naked event/job literal") stays green even when a
// feature_key value semantically overlaps with an AsynqTask name.
func defaultModelProfile(featureKey string) string {
	switch FeatureKey := featurekeys.FeatureKey(featureKey); FeatureKey {
	case featurekeys.TargetImportParse:
		return "target.import.default"
	case featurekeys.PracticeSessionChat:
		return "practice.chat.default"
	case featurekeys.ReportGenerate:
		return "report.generate.default"
	case featurekeys.ResumeParse:
		return "resume.parse.default"
	case featurekeys.ResumeTailorGapReview:
		return "resume.tailor.default"
	case featurekeys.ResumeTailorBulletSuggestions:
		return "resume.tailor.default"
	default:
		return ""
	}
}

// ResolveActive returns the active PromptResolution for (featureKey, language).
func (c *Client) ResolveActive(_ context.Context, featureKey, language string) (PromptResolution, error) {
	snap := c.cache.Load()
	return resolveSnapshot(snap, featureKey, language, &c.fallbackCount)
}

// GetPrompt returns a specific (featureKey, version, language) prompt entry.
// Used for backfill / debug; callers must supply non-empty strings.
func (c *Client) GetPrompt(featureKey, version, language string) (PromptMeta, string, error) {
	if featureKey == "" || version == "" || language == "" {
		return PromptMeta{}, "", ErrPromptUnsupported
	}
	snap := c.cache.Load()
	prompts, ok := snap.prompts[featureKey]
	if !ok {
		return PromptMeta{}, "", ErrPromptUnsupported
	}
	versions, ok := prompts[language]
	if !ok {
		return PromptMeta{}, "", ErrLanguageUnsupported
	}
	pe, ok := versions[version]
	if !ok {
		return PromptMeta{}, "", ErrPromptUnsupported
	}
	meta := pe.meta
	meta.OutputSchema = pe.outputSchema
	return meta, pe.body, nil
}

// GetRubric returns a specific (featureKey, version, language) rubric entry.
func (c *Client) GetRubric(featureKey, version, language string) (RubricSchema, error) {
	if featureKey == "" || version == "" || language == "" {
		return RubricSchema{}, ErrPromptUnsupported
	}
	snap := c.cache.Load()
	rubrics, ok := snap.rubrics[featureKey]
	if !ok {
		return RubricSchema{}, ErrPromptUnsupported
	}
	versions, ok := rubrics[language]
	if !ok {
		return RubricSchema{}, ErrLanguageUnsupported
	}
	re, ok := versions[version]
	if !ok {
		return RubricSchema{}, ErrPromptUnsupported
	}
	return re.schema, nil
}

// FallbackCount returns the cumulative number of language fallbacks
// observed since the last Reload. Tests use it to assert C-6 fallback
// behavior; F1 metrics do not consume it.
func (c *Client) FallbackCount() uint64 {
	return c.fallbackCount.Load()
}
