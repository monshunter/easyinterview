package registry

import (
	"context"
	"encoding/json"
	"errors"
)

// PromptResolution is the runtime answer to a Resolve call. The first eight
// fields mirror the targetjob.PromptResolution shape so RegistryAdapter can
// translate between the two without losing context. The last three fields
// reserve room for spec §3.1 D-12 (provider-neutral tools / structured
// output / streaming hints). Callers can already see their shape through this
// struct, but producers only populate them when the owning implementation is
// wired.
type PromptResolution struct {
	FeatureKey          string
	PromptVersion       string
	RubricVersion       string
	ModelProfileName    string
	DataSourceVersion   string
	FeatureFlag         string
	SystemMessage       string
	UserMessageTemplate string

	// D-12 reserved fields. Producers leave them empty until the owning
	// implementation is wired.
	Tools        []ToolDescriptor
	OutputSchema *json.RawMessage
	StreamWire   *string
}

// ToolDescriptor is the provider-neutral tool descriptor reserved by D-12
// for optional Resolve outputs.
type ToolDescriptor struct {
	Name        string
	Description string
	// Parameters is a JSON-Schema object; consumers must validate it before
	// passing it to a provider-side tool channel.
	Parameters json.RawMessage
}

// PromptMeta is the parsed YAML meta for a single prompt baseline entry.
// Field names and order mirror config/prompts/README.md §2.
type PromptMeta struct {
	FeatureKey   string
	Version      string
	Language     string
	TemplateHash string
	Status       string
	CreatedAt    string
	// OutputSchema is the language-independent JSON schema attached at
	// load time. It is not part of YAML metadata or template_hash input.
	OutputSchema *json.RawMessage
}

// RubricSchema is the parsed YAML body of a single rubric baseline entry.
type RubricSchema struct {
	FeatureKey string
	Version    string
	Language   string
	Status     string
	Dimensions []RubricDimension
}

// RubricDimension is one dimension entry inside a RubricSchema.
type RubricDimension struct {
	Name        string
	Weight      float64
	Description string
	ScoreLevels []ScoreLevel
}

// ScoreLevel is one threshold band inside a RubricDimension.
type ScoreLevel struct {
	Label       string
	Threshold   float64
	Description string
}

// Score is the F3 LLM Judge numeric verdict for one rubric dimension. A Judge
// call returns one Score per dimension in the resolved rubric (spec D-9):
// Dimension matches the rubric `dimensions[].name`, and Value ∈ [0,1] can be
// mapped back to a label via the rubric `score_levels[].threshold` bands.
type Score struct {
	Dimension string
	Value     float64
}

// Reasoning is the F3 LLM Judge structured reasoning trail.
type Reasoning struct {
	Summary                 string
	EvidenceQuotes          []string
	ItemVerdicts            []ItemVerdict
	CausalChecks            []CausalCheck
	ZeroToleranceViolations []string
	CriticalSafetyPass      bool
}

// ItemVerdict is the context-aware report judge verdict for one report fact,
// judgment, or advice item.
type ItemVerdict struct {
	Path                    string
	Kind                    string
	Support                 string
	EvidenceLimitedExplicit bool
	UsedForNegativeClaim    bool
	Reason                  string
}

// CausalCheck records whether one needs-work dimension has a supported issue,
// retry focus decision, and executable action chain.
type CausalCheck struct {
	DimensionCode   string
	IssueSupported  bool
	FocusSupported  bool
	ActionSupported bool
	Reason          string
}

// JudgeContext carries the frozen, redacted evaluation context for a grounded
// report verdict. It is data in the judge request, never judge instruction.
type JudgeContext struct {
	FrozenContext json.RawMessage
	Transcript    json.RawMessage
}

// Judge is the F3 LLM Judge contract. The signature mirrors spec D-9
// `Judge(featureKey, prompt_version, output, rubric_version) → ([]score,
// reasoning)` with the Go-idiomatic context.Context prepended and an error
// trailing return. The first return evolved from a single Score to a
// per-dimension []Score so multi-dimension rubrics are scored
// dimension-by-dimension.
type Judge interface {
	Judge(
		ctx context.Context,
		featureKey string,
		promptVersion string,
		output []byte,
		rubricVersion string,
	) ([]Score, Reasoning, error)
}

// ErrPromptUnsupported is returned when the requested feature_key has no
// active baseline in the loaded snapshot.
var ErrPromptUnsupported = errors.New("registry: feature_key has no active baseline")

// ErrLanguageUnsupported is returned when the requested language coordinate
// (and the multi fallback) are both absent for a feature_key.
var ErrLanguageUnsupported = errors.New("registry: language coordinate unavailable for feature_key")

// ErrJudgeUnavailable is returned by FailClosedJudge for every call.
// Callers must treat it as a hard fail-closed signal.
var ErrJudgeUnavailable = errors.New("registry: LLM Judge is not configured")
