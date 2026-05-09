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
// output / streaming hints); plan 001 does not consume them but downstream
// callers can already see their shape through this struct.
type PromptResolution struct {
	FeatureKey          string
	PromptVersion       string
	RubricVersion       string
	ModelProfileName    string
	DataSourceVersion   string
	FeatureFlag         string
	SystemMessage       string
	UserMessageTemplate string

	// D-12 reserved fields (plan 001: declared, not consumed).
	Tools        []ToolDescriptor
	OutputSchema *json.RawMessage
	StreamWire   *string
}

// ToolDescriptor is the provider-neutral tool descriptor reserved by D-12
// for future Resolve outputs. Plan 002 may attach a non-empty Tools slice;
// plan 001 always returns an empty slice.
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
}

// RubricSchema is the parsed YAML body of a single rubric baseline entry.
type RubricSchema struct {
	FeatureKey string
	Version    string
	Language   string
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

// Score is the F3 LLM Judge numeric verdict per dimension.
type Score struct {
	Dimension string
	Value     float64
}

// Reasoning is the F3 LLM Judge structured reasoning trail.
type Reasoning struct {
	Summary        string
	EvidenceQuotes []string
}

// Judge is the F3 LLM Judge contract. The signature mirrors spec D-9
// `Judge(featureKey, prompt_version, output, rubric_version) → (score,
// reasoning)` with the Go-idiomatic context.Context prepended and an error
// trailing return. Plan 001 ships only NotImplementedJudge.
type Judge interface {
	Judge(
		ctx context.Context,
		featureKey string,
		promptVersion string,
		output []byte,
		rubricVersion string,
	) (Score, Reasoning, error)
}

// ErrPromptUnsupported is returned when the requested feature_key has no
// active baseline in the loaded snapshot.
var ErrPromptUnsupported = errors.New("registry: feature_key has no active baseline")

// ErrLanguageUnsupported is returned when the requested language coordinate
// (and the multi fallback) are both absent for a feature_key.
var ErrLanguageUnsupported = errors.New("registry: language coordinate unavailable for feature_key")

// ErrJudgeNotImplemented is returned by NotImplementedJudge for every call.
// Plan 002 replaces the implementation; plan 001 callers must treat the
// error as a hard fail-closed signal.
var ErrJudgeNotImplemented = errors.New("registry: LLM Judge is not implemented in plan 001")
