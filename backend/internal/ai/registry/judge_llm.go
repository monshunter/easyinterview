package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
)

// ErrJudgeOutputInvalid is returned when the judge model response cannot be
// parsed into a valid per-dimension verdict. It is a hard fail-closed signal:
// the LLMJudge never silently zero-fills missing dimensions.
var ErrJudgeOutputInvalid = errors.New("registry: LLM judge response is invalid")

// ErrJudgeNotConfigured is returned by NewLLMJudge when a required dependency
// (registry, judge model client, or scoring instruction) is missing.
var ErrJudgeNotConfigured = errors.New("registry: LLM judge is not configured")

// JudgeModelClient is the narrow A3 surface the LLMJudge depends on. It is
// satisfied structurally by *aiclient.Client.CompleteJudge so judge calls go
// through the same profile/provider/fallback machinery as chat calls while
// staying off the business-facing AIClient interface (plan 004 §2.2). The
// LLMJudge depends only on this interface, never on the concrete client.
type JudgeModelClient interface {
	CompleteJudge(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error)
}

// RubricProvider is the narrow registry surface the LLMJudge reads: the rubric
// dimensions to score against, and the prompt entry that carries the
// language-independent output schema used to fail-close on an invalid
// evaluated output. *Client satisfies it.
type RubricProvider interface {
	GetRubric(featureKey, version, language string) (RubricSchema, error)
	GetPrompt(featureKey, version, language string) (PromptMeta, string, error)
}

// LLMJudge is the F3 real LLM Judge (spec D-9 v2.8). It loads the rubric and
// output schema from the registry single source, asks the judge model to score
// each rubric dimension, and parses a per-dimension []Score + Reasoning. The
// business prompt under evaluation is always resolved through the registry;
// the judge scoring instruction is injected from the config/evals eval asset
// (never hardcoded in this package).
type LLMJudge struct {
	reg         RubricProvider
	model       JudgeModelClient
	instruction string
	profileName string
	language    string
}

// Compile-time assertion that LLMJudge satisfies the Judge contract. The eval
// runner obtains an LLMJudge by constructor injection (NewLLMJudge), never via
// a global singleton; NotImplementedJudge remains the fail-closed default for
// callers that have not injected a judge dependency.
var _ Judge = (*LLMJudge)(nil)

// LLMJudgeOption tunes LLMJudge construction.
type LLMJudgeOption func(*LLMJudge)

// WithJudgeProfile overrides the judge model profile name (default
// "judge.default").
func WithJudgeProfile(name string) LLMJudgeOption {
	return func(j *LLMJudge) {
		if name != "" {
			j.profileName = name
		}
	}
}

// WithJudgeRubricLanguage overrides the rubric/schema language coordinate
// (default canonical "multi").
func WithJudgeRubricLanguage(language string) LLMJudgeOption {
	return func(j *LLMJudge) {
		if language != "" {
			j.language = language
		}
	}
}

// NewLLMJudge constructs an LLMJudge. instruction is the scoring instruction
// template loaded from the config/evals eval asset by the caller; it must be
// non-empty so this package never embeds a hardcoded judge prompt.
func NewLLMJudge(reg RubricProvider, model JudgeModelClient, instruction string, opts ...LLMJudgeOption) (*LLMJudge, error) {
	if reg == nil || model == nil || instruction == "" {
		return nil, ErrJudgeNotConfigured
	}
	j := &LLMJudge{
		reg:         reg,
		model:       model,
		instruction: instruction,
		profileName: "judge.default",
		language:    "multi",
	}
	for _, opt := range opts {
		opt(j)
	}
	return j, nil
}

// judgeRequest is the structured data block sent to the judge model. The
// scoring instruction prose lives in the system message (from config/evals);
// the dimensions and the output under evaluation are passed as data.
type judgeRequest struct {
	FeatureKey    string               `json:"feature_key"`
	PromptVersion string               `json:"prompt_version"`
	RubricVersion string               `json:"rubric_version"`
	Dimensions    []judgeRequestDim `json:"dimensions"`
	Output        json.RawMessage   `json:"output_to_evaluate"`
}

type judgeRequestDim struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// judgeResponseEnvelope is the strict JSON shape the judge model must return.
type judgeResponseEnvelope struct {
	Scores []struct {
		Dimension string  `json:"dimension"`
		Value     float64 `json:"value"`
	} `json:"scores"`
	Reasoning struct {
		Summary        string   `json:"summary"`
		EvidenceQuotes []string `json:"evidence_quotes"`
	} `json:"reasoning"`
}

// Judge implements the Judge interface (spec D-9 v2.8).
func (j *LLMJudge) Judge(ctx context.Context, featureKey, promptVersion string, output []byte, rubricVersion string) ([]Score, Reasoning, error) {
	// 1. Fail-close if the evaluated output does not satisfy the
	//    (featureKey, promptVersion) output schema. Reuses the single A3
	//    validator subset (plan 004 §2.5).
	meta, _, err := j.reg.GetPrompt(featureKey, promptVersion, j.language)
	if err != nil {
		return nil, Reasoning{}, err
	}
	if meta.OutputSchema != nil && len(*meta.OutputSchema) > 0 {
		if vErr := outputschema.Validate(*meta.OutputSchema, string(output)); vErr != nil {
			return nil, Reasoning{}, fmt.Errorf("%w: evaluated output failed schema: %v", ErrJudgeOutputInvalid, vErr)
		}
	}

	// 2. Load rubric dimensions to score against.
	rubric, err := j.reg.GetRubric(featureKey, rubricVersion, j.language)
	if err != nil {
		return nil, Reasoning{}, err
	}
	if len(rubric.Dimensions) == 0 {
		return nil, Reasoning{}, fmt.Errorf("%w: rubric %s@%s has no dimensions", ErrJudgeOutputInvalid, featureKey, rubricVersion)
	}

	// 3. Build the judge payload: scoring instruction (system) + structured
	//    data (user). The business prompt itself is not copied here; only the
	//    evaluated output is passed.
	payload, err := j.buildPayload(featureKey, promptVersion, rubricVersion, rubric, output)
	if err != nil {
		return nil, Reasoning{}, err
	}

	// 4. Call the judge model through the narrow A3 surface.
	resp, _, err := j.model.CompleteJudge(ctx, j.profileName, payload)
	if err != nil {
		return nil, Reasoning{}, err
	}

	// 5. Parse and validate the per-dimension verdict (fail-close).
	return parseJudgeResponse(resp.Content, rubric)
}

func (j *LLMJudge) buildPayload(featureKey, promptVersion, rubricVersion string, rubric RubricSchema, output []byte) (aiclient.CompletePayload, error) {
	dims := make([]judgeRequestDim, 0, len(rubric.Dimensions))
	for _, d := range rubric.Dimensions {
		dims = append(dims, judgeRequestDim{Name: d.Name, Description: d.Description})
	}
	req := judgeRequest{
		FeatureKey:    featureKey,
		PromptVersion: promptVersion,
		RubricVersion: rubricVersion,
		Dimensions:    dims,
		Output:        json.RawMessage(output),
	}
	userBody, err := json.Marshal(req)
	if err != nil {
		return aiclient.CompletePayload{}, fmt.Errorf("registry: marshal judge request: %w", err)
	}
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: j.instruction},
			{Role: "user", Content: string(userBody)},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    featureKey,
			PromptVersion: promptVersion,
			RubricVersion: rubricVersion,
			Language:      j.language,
		},
	}, nil
}

func parseJudgeResponse(content string, rubric RubricSchema) ([]Score, Reasoning, error) {
	if content == "" {
		return nil, Reasoning{}, fmt.Errorf("%w: empty judge response", ErrJudgeOutputInvalid)
	}
	var env judgeResponseEnvelope
	if err := json.Unmarshal([]byte(content), &env); err != nil {
		return nil, Reasoning{}, fmt.Errorf("%w: parse judge response: %v", ErrJudgeOutputInvalid, err)
	}
	if len(env.Scores) != len(rubric.Dimensions) {
		return nil, Reasoning{}, fmt.Errorf("%w: judge returned %d scores for %d rubric dimensions", ErrJudgeOutputInvalid, len(env.Scores), len(rubric.Dimensions))
	}
	if env.Reasoning.Summary == "" {
		return nil, Reasoning{}, fmt.Errorf("%w: judge reasoning summary is empty", ErrJudgeOutputInvalid)
	}

	want := make(map[string]bool, len(rubric.Dimensions))
	for _, d := range rubric.Dimensions {
		want[d.Name] = true
	}
	scores := make([]Score, 0, len(env.Scores))
	seen := make(map[string]bool, len(env.Scores))
	for _, s := range env.Scores {
		if !want[s.Dimension] {
			return nil, Reasoning{}, fmt.Errorf("%w: judge scored unknown dimension %q", ErrJudgeOutputInvalid, s.Dimension)
		}
		if seen[s.Dimension] {
			return nil, Reasoning{}, fmt.Errorf("%w: judge scored dimension %q twice", ErrJudgeOutputInvalid, s.Dimension)
		}
		if s.Value < 0 || s.Value > 1 {
			return nil, Reasoning{}, fmt.Errorf("%w: dimension %q value %v outside [0,1]", ErrJudgeOutputInvalid, s.Dimension, s.Value)
		}
		seen[s.Dimension] = true
		scores = append(scores, Score{Dimension: s.Dimension, Value: s.Value})
	}

	reasoning := Reasoning{
		Summary:        env.Reasoning.Summary,
		EvidenceQuotes: env.Reasoning.EvidenceQuotes,
	}
	return scores, reasoning, nil
}
