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
// Hand-written today; the truth source remains config/prompts/<feature_key>/.
// A future codegen pass may replace this file with generated output sourced
// from config/prompts/. Until then, every feature_key string literal in
// non-allowlisted Go source must be sourced through this package so the
// scripts/lint/lint_events.py "naked event/job literal" gate stays green.
package featurekeys

// FeatureKey is the typed identifier for AI prompt-rubric routing keys.
type FeatureKey string

const (
	TargetImportParse              FeatureKey = "target.import.parse"
	PracticeSessionFirstQuestion   FeatureKey = "practice.session.first_question"
	PracticeSessionFollowUp        FeatureKey = "practice.session.follow_up"
	PracticeTurnLightweightObserve FeatureKey = "practice.turn.lightweight_observe"
	ReportGenerate                 FeatureKey = "report.generate"
	ReportQuestionAssessment       FeatureKey = "report.question_assessment"
	ResumeParse                    FeatureKey = "resume.parse"
	ResumeTailorGapReview          FeatureKey = "resume.tailor.gap_review"
	ResumeTailorBulletSuggestions  FeatureKey = "resume.tailor.bullet_suggestions"
)

// All returns every known feature_key constant in declaration order. Useful
// for codegen drift checks that compare this list against the contents of
// config/prompts/.
func All() []FeatureKey {
	return []FeatureKey{
		TargetImportParse,
		PracticeSessionFirstQuestion,
		PracticeSessionFollowUp,
		PracticeTurnLightweightObserve,
		ReportGenerate,
		ReportQuestionAssessment,
		ResumeParse,
		ResumeTailorGapReview,
		ResumeTailorBulletSuggestions,
	}
}

// String returns the wire string value of a FeatureKey for use in switch
// cases or comparisons against untyped string parameters.
func (f FeatureKey) String() string {
	return string(f)
}
