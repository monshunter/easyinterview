// Package featurekeys defines the canonical AI feature_key constants used by
// prompt-rubric-registry routing.
//
// feature_key is the cross-cutting AI routing key consumed by:
//   - config/prompts/<feature_key>/v*.yaml — prompt versions
//   - config/rubrics/<feature_key>/v*.yaml — rubric versions
//   - prompt-rubric-registry runtime resolver (defaultModelProfile, A3 routing)
//
// Several feature_key string values intentionally overlap with AsynqTask
// names (for example "resume.parse" matches both AsynqTaskResumeParse and
// feature_key=resume.parse). Both refer to the same cross-cutting feature
// being invoked, but the contracts are distinct: AsynqTask names belong to
// shared/jobs (queue routing), while feature_keys belong here (AI routing).
//
// These constants mirror the current config/prompts/<feature_key>/ coordinates
// and give Go callers one typed source for routing keys.
package featurekeys

// FeatureKey is the typed identifier for AI prompt-rubric routing keys.
type FeatureKey string

const (
	TargetImportParse             FeatureKey = "target.import.parse"
	PracticeSessionChat           FeatureKey = "practice.session.chat"
	ReportGenerate                FeatureKey = "report.generate"
	ResumeParse                   FeatureKey = "resume.parse"
	ResumeTailorGapReview         FeatureKey = "resume.tailor.gap_review"
	ResumeTailorBulletSuggestions FeatureKey = "resume.tailor.bullet_suggestions"
)
