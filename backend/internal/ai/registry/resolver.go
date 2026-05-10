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

	prompt, lang, ok := selectByLanguage(prompts, language)
	if !ok {
		return PromptResolution{}, ErrLanguageUnsupported
	}
	rubric, _, ok := selectByLanguage(rubrics, lang)
	if !ok {
		// Language parity is enforced at load time, so this should be
		// unreachable in practice; return a precise error if it ever fires.
		return PromptResolution{}, ErrLanguageUnsupported
	}

	if lang != language && fallbackCounter != nil {
		fallbackCounter.Add(1)
	}

	// Plan 001 ships SystemMessage empty and UserMessageTemplate as the
	// full markdown body; targetjob's existing executor consumes the body
	// through UserMessageTemplate. Plan 002 may split body into system /
	// user sections per feature_key.
	pe := prompt.(promptEntry)
	re := rubric.(rubricEntry)
	return PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       pe.meta.Version,
		RubricVersion:       re.schema.Version,
		ModelProfileName:    defaultModelProfile(featureKey),
		DataSourceVersion:   "registry.v1",
		FeatureFlag:         "none",
		SystemMessage:       "",
		UserMessageTemplate: pe.body,
		Tools:               nil,
		OutputSchema:        nil,
		StreamWire:          nil,
	}, nil
}

// selectByLanguage returns the entry for the requested language, falling
// back to "multi" when the exact language is missing. The returned string
// is the language actually used (so callers can tell whether fallback fired).
func selectByLanguage[T any](entries map[string]T, requested string) (any, string, bool) {
	if e, ok := entries[requested]; ok {
		return e, requested, true
	}
	if e, ok := entries["multi"]; ok {
		return e, "multi", true
	}
	return *new(T), "", false
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
	case featurekeys.PracticeSessionFirstQuestion:
		return "practice.first_question.default"
	case featurekeys.PracticeSessionFollowUp:
		return "practice.followup.default"
	case featurekeys.PracticeTurnLightweightObserve:
		return "practice.turn_observe.default"
	case featurekeys.ReportGenerate:
		return "report.generate.default"
	case featurekeys.ReportQuestionAssessment:
		return "report.assessment.default"
	case featurekeys.ResumeParse:
		return "resume.parse.default"
	case featurekeys.ResumeTailorGapReview:
		return "resume.tailor.default"
	case featurekeys.ResumeTailorBulletSuggestions:
		return "resume.tailor.default"
	case featurekeys.DebriefGenerate:
		return "debrief.generate.default"
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
	pe, ok := prompts[language]
	if !ok {
		return PromptMeta{}, "", ErrLanguageUnsupported
	}
	if pe.meta.Version != version {
		return PromptMeta{}, "", ErrPromptUnsupported
	}
	return pe.meta, pe.body, nil
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
	re, ok := rubrics[language]
	if !ok {
		return RubricSchema{}, ErrLanguageUnsupported
	}
	if re.schema.Version != version {
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
