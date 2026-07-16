package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

// ErrJudgeOutputInvalid is returned when the judge model response cannot be
// parsed into a valid per-dimension verdict. It is a hard fail-closed signal:
// the LLMJudge never silently zero-fills missing dimensions.
var ErrJudgeOutputInvalid = errors.New("registry: LLM judge response is invalid")

// ErrJudgeProtocolInvalid identifies a malformed, schema-invalid, or
// coverage-incomplete judge response. These failures are safe to retry because
// no valid content verdict was produced.
var ErrJudgeProtocolInvalid = errors.New("registry: LLM judge protocol is invalid")

// ErrJudgeContentRejected identifies a structurally valid terminal negative
// verdict. It must never be retried into a pass: unsupported report items,
// failed causal checks, zero-tolerance violations, and critical-safety failure
// are valid evaluation results rather than response-protocol failures.
var ErrJudgeContentRejected = errors.New("registry: LLM judge rejected content")

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

// LLMJudge is the F3 real LLM Judge (spec D-9). It loads the rubric and
// output schema from the registry single source, asks the judge model to score
// each rubric dimension, and parses a per-dimension []Score + Reasoning. The
// business prompt under evaluation is always resolved through the registry;
// the judge scoring instruction is injected from the config/evals eval asset
// (never hardcoded in this package).
type LLMJudge struct {
	reg         RubricProvider
	model       JudgeModelClient
	instruction string
	language    string
}

const maxJudgeAttempts = 4

func judgeProtocolInvalidf(format string, args ...any) error {
	return fmt.Errorf("%w: %w: %s", ErrJudgeOutputInvalid, ErrJudgeProtocolInvalid, fmt.Sprintf(format, args...))
}

func judgeContentRejectedf(format string, args ...any) error {
	return fmt.Errorf("%w: %w: %s", ErrJudgeOutputInvalid, ErrJudgeContentRejected, fmt.Sprintf(format, args...))
}

// Compile-time assertion that LLMJudge satisfies the Judge contract. The eval
// runner obtains an LLMJudge by constructor injection (NewLLMJudge), never via
// a global singleton; FailClosedJudge remains the fail-closed default for
// callers that have not injected a judge dependency.
var _ Judge = (*LLMJudge)(nil)

// LLMJudgeOption tunes LLMJudge construction.
type LLMJudgeOption func(*LLMJudge)

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
	FeatureKey                   string                    `json:"feature_key"`
	PromptVersion                string                    `json:"prompt_version"`
	RubricVersion                string                    `json:"rubric_version"`
	Dimensions                   []judgeRequestDim         `json:"dimensions"`
	ExpectedItemVerdicts         []judgeRequestItemVerdict `json:"expected_item_verdicts,omitempty"`
	ExpectedCausalDimensionCodes []string                  `json:"expected_causal_dimension_codes,omitempty"`
	FrozenContext                json.RawMessage           `json:"frozen_context,omitempty"`
	Transcript                   json.RawMessage           `json:"transcript,omitempty"`
	Output                       json.RawMessage           `json:"output_to_evaluate"`
}

type judgeRequestItemVerdict struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type judgeRequestDim struct {
	Name        string                   `json:"name"`
	Weight      float64                  `json:"weight"`
	Description string                   `json:"description"`
	ScoreLevels []judgeRequestScoreLevel `json:"score_levels"`
}

type judgeRequestScoreLevel struct {
	Label       string  `json:"label"`
	Threshold   float64 `json:"threshold"`
	Description string  `json:"description"`
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
	ItemVerdicts []struct {
		Path                    string `json:"path"`
		Kind                    string `json:"kind"`
		Support                 string `json:"support"`
		EvidenceLimitedExplicit bool   `json:"evidence_limited_explicit"`
		UsedForNegativeClaim    bool   `json:"used_for_negative_claim"`
		Reason                  string `json:"reason"`
	} `json:"item_verdicts,omitempty"`
	CausalChecks []struct {
		DimensionCode   string `json:"dimension_code"`
		IssueSupported  bool   `json:"issue_supported"`
		FocusSupported  bool   `json:"focus_supported"`
		ActionSupported bool   `json:"action_supported"`
		Reason          string `json:"reason"`
	} `json:"causal_checks,omitempty"`
	ZeroToleranceViolations []string `json:"zero_tolerance_violations,omitempty"`
	CriticalSafetyPass      *bool    `json:"critical_safety_pass,omitempty"`
}

// Judge implements the Judge interface (spec D-9).
func (j *LLMJudge) Judge(ctx context.Context, featureKey, promptVersion string, output []byte, rubricVersion string) ([]Score, Reasoning, error) {
	return j.judge(ctx, featureKey, promptVersion, output, rubricVersion, JudgeContext{})
}

// JudgeWithContext evaluates an output against frozen context and transcript.
// report.generate v0.2 requires this path; older/non-report evaluations retain
// the output-only Judge contract.
func (j *LLMJudge) JudgeWithContext(
	ctx context.Context,
	featureKey string,
	promptVersion string,
	output []byte,
	rubricVersion string,
	evidence JudgeContext,
) ([]Score, Reasoning, error) {
	return j.judge(ctx, featureKey, promptVersion, output, rubricVersion, evidence)
}

func (j *LLMJudge) judge(
	ctx context.Context,
	featureKey string,
	promptVersion string,
	output []byte,
	rubricVersion string,
	evidence JudgeContext,
) ([]Score, Reasoning, error) {
	requireGroundedVerdict := featureKey == string(featurekeys.ReportGenerate) && promptVersion == "v0.2.0" && rubricVersion == "v0.2.0"
	var groundedOutput *groundedReportOutput
	if requireGroundedVerdict {
		if err := validateJudgeContext(evidence); err != nil {
			return nil, Reasoning{}, err
		}
		decoded, err := decodeGroundedReportOutput(output)
		if err != nil {
			return nil, Reasoning{}, err
		}
		groundedOutput = &decoded
	}
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
	payload, err := j.buildPayload(featureKey, promptVersion, rubricVersion, rubric, output, evidence)
	if err != nil {
		return nil, Reasoning{}, err
	}

	// 4. Call the judge model through the narrow A3 surface. The bounded budget
	//    is initial + at most three retries. Only typed retryable provider errors
	//    and protocol-invalid judge responses may consume another attempt.
	for attempt := 1; attempt <= maxJudgeAttempts; attempt++ {
		resp, _, callErr := j.model.CompleteJudge(ctx, "judge.default", payload)
		if callErr != nil {
			if attempt < maxJudgeAttempts && retryableJudgeProviderError(ctx, callErr) {
				continue
			}
			return nil, Reasoning{}, callErr
		}

		scores, reasoning, verdictErr := parseJudgeResponse(resp.Content, rubric, groundedOutput)
		if verdictErr == nil {
			return scores, reasoning, nil
		}
		if errors.Is(verdictErr, ErrJudgeContentRejected) || !errors.Is(verdictErr, ErrJudgeProtocolInvalid) || attempt == maxJudgeAttempts {
			return nil, Reasoning{}, verdictErr
		}
	}
	return nil, Reasoning{}, judgeProtocolInvalidf("judge attempt budget exhausted")
}

func retryableJudgeProviderError(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *sharederrors.APIError
	return errors.As(err, &apiErr) && apiErr.Retryable
}

func validateJudgeContext(evidence JudgeContext) error {
	if len(evidence.FrozenContext) == 0 || !json.Valid(evidence.FrozenContext) {
		return fmt.Errorf("%w: grounded report frozen context is required", ErrJudgeOutputInvalid)
	}
	var frozenContext map[string]json.RawMessage
	if err := json.Unmarshal(evidence.FrozenContext, &frozenContext); err != nil || len(frozenContext) == 0 {
		return fmt.Errorf("%w: grounded report frozen context must be a non-empty object", ErrJudgeOutputInvalid)
	}
	if len(evidence.Transcript) == 0 || !json.Valid(evidence.Transcript) {
		return fmt.Errorf("%w: grounded report transcript is required", ErrJudgeOutputInvalid)
	}
	var transcript []json.RawMessage
	if err := json.Unmarshal(evidence.Transcript, &transcript); err != nil || len(transcript) == 0 {
		return fmt.Errorf("%w: grounded report transcript must be a non-empty array", ErrJudgeOutputInvalid)
	}
	return nil
}

func (j *LLMJudge) buildPayload(featureKey, promptVersion, rubricVersion string, rubric RubricSchema, output []byte, evidence JudgeContext) (aiclient.CompletePayload, error) {
	dims := make([]judgeRequestDim, 0, len(rubric.Dimensions))
	for _, d := range rubric.Dimensions {
		levels := make([]judgeRequestScoreLevel, 0, len(d.ScoreLevels))
		for _, level := range d.ScoreLevels {
			levels = append(levels, judgeRequestScoreLevel(level))
		}
		dims = append(dims, judgeRequestDim{Name: d.Name, Weight: d.Weight, Description: d.Description, ScoreLevels: levels})
	}
	req := judgeRequest{
		FeatureKey:    featureKey,
		PromptVersion: promptVersion,
		RubricVersion: rubricVersion,
		Dimensions:    dims,
		FrozenContext: evidence.FrozenContext,
		Transcript:    evidence.Transcript,
		Output:        json.RawMessage(output),
	}
	if featureKey == string(featurekeys.ReportGenerate) && promptVersion == "v0.2.0" {
		report, decodeErr := decodeGroundedReportOutput(output)
		if decodeErr != nil {
			return aiclient.CompletePayload{}, decodeErr
		}
		req.ExpectedItemVerdicts, req.ExpectedCausalDimensionCodes = groundedJudgeCoordinates(report)
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

func groundedJudgeCoordinates(report groundedReportOutput) ([]judgeRequestItemVerdict, []string) {
	items := []judgeRequestItemVerdict{
		{Path: "$.summary", Kind: "judgment"},
		{Path: "$.preparednessLevel", Kind: "judgment"},
	}
	causalCodes := make([]string, 0, len(report.DimensionAssessments))
	for index, dimension := range report.DimensionAssessments {
		items = append(items, judgeRequestItemVerdict{Path: fmt.Sprintf("$.dimensionAssessments[%d]", index), Kind: "judgment"})
		if dimension.Status == "needs_work" {
			causalCodes = append(causalCodes, dimension.Code)
		}
	}
	for index := range report.Highlights {
		items = append(items, judgeRequestItemVerdict{Path: fmt.Sprintf("$.highlights[%d]", index), Kind: "fact"})
	}
	for index := range report.Issues {
		items = append(items, judgeRequestItemVerdict{Path: fmt.Sprintf("$.issues[%d]", index), Kind: "judgment"})
	}
	for index := range report.NextActions {
		items = append(items, judgeRequestItemVerdict{Path: fmt.Sprintf("$.nextActions[%d]", index), Kind: "advice"})
	}
	items = append(items, judgeRequestItemVerdict{Path: "$.retryFocusDimensionCodes", Kind: "advice"})
	return items, causalCodes
}

func parseJudgeResponse(content string, rubric RubricSchema, groundedOutput *groundedReportOutput) ([]Score, Reasoning, error) {
	if content == "" {
		return nil, Reasoning{}, judgeProtocolInvalidf("empty judge response")
	}
	var env judgeResponseEnvelope
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&env); err != nil {
		return nil, Reasoning{}, judgeProtocolInvalidf("parse judge response: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, Reasoning{}, judgeProtocolInvalidf("trailing judge response content")
	}
	if len(env.Scores) != len(rubric.Dimensions) {
		return nil, Reasoning{}, judgeProtocolInvalidf("judge returned %d scores for %d rubric dimensions", len(env.Scores), len(rubric.Dimensions))
	}
	if env.Reasoning.Summary == "" {
		return nil, Reasoning{}, judgeProtocolInvalidf("judge reasoning summary is empty")
	}

	want := make(map[string]bool, len(rubric.Dimensions))
	for _, d := range rubric.Dimensions {
		want[d.Name] = true
	}
	scores := make([]Score, 0, len(env.Scores))
	seen := make(map[string]bool, len(env.Scores))
	for _, s := range env.Scores {
		if !want[s.Dimension] {
			return nil, Reasoning{}, judgeProtocolInvalidf("judge scored unknown dimension %q", s.Dimension)
		}
		if seen[s.Dimension] {
			return nil, Reasoning{}, judgeProtocolInvalidf("judge scored dimension %q twice", s.Dimension)
		}
		if s.Value < 0 || s.Value > 1 {
			return nil, Reasoning{}, judgeProtocolInvalidf("dimension %q value %v outside [0,1]", s.Dimension, s.Value)
		}
		seen[s.Dimension] = true
		scores = append(scores, Score{Dimension: s.Dimension, Value: s.Value})
	}

	reasoning := Reasoning{
		Summary:        env.Reasoning.Summary,
		EvidenceQuotes: env.Reasoning.EvidenceQuotes,
	}
	if groundedOutput != nil {
		if err := validateGroundedReportVerdict(env, *groundedOutput, &reasoning); err != nil {
			return nil, Reasoning{}, err
		}
	}
	return scores, reasoning, nil
}

type groundedReportOutput struct {
	Summary              string `json:"summary"`
	PreparednessLevel    string `json:"preparednessLevel"`
	DimensionAssessments []struct {
		Code       string `json:"code"`
		Label      string `json:"label"`
		Status     string `json:"status"`
		Confidence string `json:"confidence"`
	} `json:"dimensionAssessments"`
	Highlights []struct {
		DimensionCode       string `json:"dimensionCode"`
		Evidence            string `json:"evidence"`
		Confidence          string `json:"confidence"`
		SourceMessageSeqNos []int  `json:"sourceMessageSeqNos"`
	} `json:"highlights"`
	Issues []struct {
		DimensionCode       string `json:"dimensionCode"`
		Evidence            string `json:"evidence"`
		Confidence          string `json:"confidence"`
		SourceMessageSeqNos []int  `json:"sourceMessageSeqNos"`
	} `json:"issues"`
	NextActions []struct {
		Type  string `json:"type"`
		Label string `json:"label"`
	} `json:"nextActions"`
	RetryFocusDimensionCodes []string `json:"retryFocusDimensionCodes"`
}

func decodeGroundedReportOutput(output []byte) (groundedReportOutput, error) {
	var report groundedReportOutput
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&report); err != nil {
		return groundedReportOutput{}, fmt.Errorf("%w: strict grounded report output: %v", ErrJudgeOutputInvalid, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return groundedReportOutput{}, fmt.Errorf("%w: grounded report output has trailing content", ErrJudgeOutputInvalid)
	}
	return report, nil
}

func validateGroundedReportVerdict(env judgeResponseEnvelope, report groundedReportOutput, reasoning *Reasoning) error {
	expectedItems := map[string]string{
		"$.summary":                  "judgment",
		"$.preparednessLevel":        "judgment",
		"$.retryFocusDimensionCodes": "advice",
	}
	needsWork := map[string]bool{}
	for i, dimension := range report.DimensionAssessments {
		expectedItems[fmt.Sprintf("$.dimensionAssessments[%d]", i)] = "judgment"
		if dimension.Status == "needs_work" {
			needsWork[dimension.Code] = true
		}
	}
	for i := range report.Highlights {
		expectedItems[fmt.Sprintf("$.highlights[%d]", i)] = "fact"
	}
	for i := range report.Issues {
		expectedItems[fmt.Sprintf("$.issues[%d]", i)] = "judgment"
	}
	for i := range report.NextActions {
		expectedItems[fmt.Sprintf("$.nextActions[%d]", i)] = "advice"
	}

	seenItems := map[string]bool{}
	for _, verdict := range env.ItemVerdicts {
		expectedKind, ok := expectedItems[verdict.Path]
		if !ok || verdict.Kind != expectedKind || seenItems[verdict.Path] {
			return judgeProtocolInvalidf("invalid or duplicate item verdict %q/%q", verdict.Path, verdict.Kind)
		}
		if strings.TrimSpace(verdict.Reason) == "" {
			return judgeProtocolInvalidf("item verdict %q reason is required", verdict.Path)
		}
		switch verdict.Support {
		case "supported":
		case "partial":
			if !verdict.EvidenceLimitedExplicit || verdict.UsedForNegativeClaim {
				return judgeProtocolInvalidf("partial item %q must be explicitly evidence-limited and non-negative", verdict.Path)
			}
		case "unsupported":
			return judgeContentRejectedf("unsupported report item %q", verdict.Path)
		default:
			return judgeProtocolInvalidf("item verdict %q has invalid support %q", verdict.Path, verdict.Support)
		}
		seenItems[verdict.Path] = true
		reasoning.ItemVerdicts = append(reasoning.ItemVerdicts, ItemVerdict{
			Path: verdict.Path, Kind: verdict.Kind, Support: verdict.Support,
			EvidenceLimitedExplicit: verdict.EvidenceLimitedExplicit,
			UsedForNegativeClaim:    verdict.UsedForNegativeClaim,
			Reason:                  verdict.Reason,
		})
	}
	if len(seenItems) != len(expectedItems) {
		return judgeProtocolInvalidf("item verdict count %d does not cover %d report items", len(seenItems), len(expectedItems))
	}

	seenCausal := map[string]bool{}
	for _, check := range env.CausalChecks {
		if !needsWork[check.DimensionCode] || seenCausal[check.DimensionCode] {
			return judgeProtocolInvalidf("invalid or duplicate causal check %q", check.DimensionCode)
		}
		if strings.TrimSpace(check.Reason) == "" {
			return judgeProtocolInvalidf("causal check %q reason is required", check.DimensionCode)
		}
		if !check.IssueSupported || !check.FocusSupported || !check.ActionSupported {
			return judgeContentRejectedf("causal check %q failed", check.DimensionCode)
		}
		seenCausal[check.DimensionCode] = true
		reasoning.CausalChecks = append(reasoning.CausalChecks, CausalCheck{
			DimensionCode: check.DimensionCode, IssueSupported: check.IssueSupported,
			FocusSupported: check.FocusSupported, ActionSupported: check.ActionSupported,
			Reason: check.Reason,
		})
	}
	if len(seenCausal) != len(needsWork) {
		return judgeProtocolInvalidf("causal checks cover %d of %d needs-work dimensions", len(seenCausal), len(needsWork))
	}
	if len(env.ZeroToleranceViolations) != 0 {
		return judgeContentRejectedf("zero-tolerance violation present")
	}
	if env.CriticalSafetyPass == nil {
		return judgeProtocolInvalidf("critical safety verdict is missing")
	}
	if !*env.CriticalSafetyPass {
		return judgeContentRejectedf("critical safety verdict did not pass")
	}
	reasoning.ZeroToleranceViolations = append([]string(nil), env.ZeroToleranceViolations...)
	reasoning.CriticalSafetyPass = true
	return nil
}
