// Package eval is the F3 offline evaluation harness backend (plan
// prompt-rubric-registry/004 §4). It loads the config/evals suite, resolves
// each case's business prompt through the registry single source, and grades
// recorded golden outputs with the real registry.LLMJudge. The default path is
// deterministic and makes no network call: the judge verdict comes from the
// recorded fixture, not a live provider. EVAL_LIVE wiring lives in the
// cmd/evalkit live path, not here.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"gopkg.in/yaml.v3"
)

// DimensionScore is one recorded per-dimension verdict in a case fixture.
type DimensionScore struct {
	Dimension string  `yaml:"dimension"`
	Value     float64 `yaml:"value"`
}

// JudgeFixture is the recorded judge transcript for a case: the verdict the
// judge model would have returned. Offline grading replays it deterministically.
type JudgeFixture struct {
	Scores    []DimensionScore `yaml:"scores"`
	Reasoning struct {
		Summary        string   `yaml:"summary"`
		EvidenceQuotes []string `yaml:"evidence_quotes"`
	} `yaml:"reasoning"`
	ItemVerdicts            []ItemVerdictFixture `yaml:"item_verdicts,omitempty"`
	CausalChecks            []CausalCheckFixture `yaml:"causal_checks,omitempty"`
	ZeroToleranceViolations []string             `yaml:"zero_tolerance_violations,omitempty"`
	CriticalSafetyPass      *bool                `yaml:"critical_safety_pass,omitempty"`
}

type ItemVerdictFixture struct {
	Path                    string `yaml:"path"`
	Kind                    string `yaml:"kind"`
	Support                 string `yaml:"support"`
	EvidenceLimitedExplicit bool   `yaml:"evidence_limited_explicit"`
	UsedForNegativeClaim    bool   `yaml:"used_for_negative_claim"`
	Reason                  string `yaml:"reason"`
}

type CausalCheckFixture struct {
	DimensionCode   string `yaml:"dimension_code"`
	IssueSupported  bool   `yaml:"issue_supported"`
	FocusSupported  bool   `yaml:"focus_supported"`
	ActionSupported bool   `yaml:"action_supported"`
	Reason          string `yaml:"reason"`
}

// Case is one offline evaluation case.
type Case struct {
	ID            string       `yaml:"id"`
	FeatureKey    string       `yaml:"-"`
	Language      string       `yaml:"language"`
	PromptVersion string       `yaml:"prompt_version,omitempty"`
	RubricVersion string       `yaml:"rubric_version,omitempty"`
	Input         string       `yaml:"input"`
	Context       any          `yaml:"context,omitempty"`
	Transcript    any          `yaml:"transcript,omitempty"`
	Critical      bool         `yaml:"critical,omitempty"`
	Redacted      bool         `yaml:"redacted,omitempty"`
	Output        any          `yaml:"output"`
	Judge         JudgeFixture `yaml:"judge"`
}

// Suite is the loaded eval suite.
type Suite struct {
	Instruction string
	Cases       []Case
}

type caseFile struct {
	FeatureKey string `yaml:"feature_key"`
	Cases      []Case `yaml:"cases"`
}

// Count returns the number of loaded cases.
func (s *Suite) Count() int { return len(s.Cases) }

// LoadSuite reads judge-instruction.md plus every <feature_key>/cases.yaml under
// dir. Case IDs must be unique and feature_key non-empty.
func LoadSuite(dir string) (*Suite, error) {
	instructionPath := filepath.Join(dir, "judge-instruction.md")
	instruction, err := os.ReadFile(instructionPath)
	if err != nil {
		return nil, fmt.Errorf("eval: read judge instruction: %w", err)
	}
	suite := &Suite{Instruction: string(instruction)}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("eval: read evals dir: %w", err)
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		casesPath := filepath.Join(dir, entry.Name(), "cases.yaml")
		raw, err := os.ReadFile(casesPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("eval: read %s: %w", casesPath, err)
		}
		var cf caseFile
		if err := yaml.Unmarshal(raw, &cf); err != nil {
			return nil, fmt.Errorf("eval: parse %s: %w", casesPath, err)
		}
		if cf.FeatureKey == "" {
			return nil, fmt.Errorf("eval: %s missing feature_key", casesPath)
		}
		if cf.FeatureKey != entry.Name() {
			return nil, fmt.Errorf("eval: %s feature_key %q does not match directory %q", casesPath, cf.FeatureKey, entry.Name())
		}
		for i := range cf.Cases {
			c := cf.Cases[i]
			c.FeatureKey = cf.FeatureKey
			if c.ID == "" {
				return nil, fmt.Errorf("eval: %s has a case with empty id", casesPath)
			}
			if seen[c.ID] {
				return nil, fmt.Errorf("eval: duplicate case id %q", c.ID)
			}
			if c.Language == "" {
				return nil, fmt.Errorf("eval: case %q missing language", c.ID)
			}
			seen[c.ID] = true
			suite.Cases = append(suite.Cases, c)
		}
	}
	sort.Slice(suite.Cases, func(i, j int) bool { return suite.Cases[i].ID < suite.Cases[j].ID })
	if err := validateGroundedReportCases(suite.Cases); err != nil {
		return nil, err
	}
	return suite, nil
}

func validateGroundedReportCases(cases []Case) error {
	count := 0
	critical := 0
	outputs := map[string]string{}
	for _, c := range cases {
		if c.FeatureKey != string(featurekeys.ReportGenerate) {
			continue
		}
		count++
		if c.PromptVersion != "v0.2.0" || c.RubricVersion != "v0.2.0" {
			return fmt.Errorf("eval: report case %q must pin prompt/rubric v0.2.0", c.ID)
		}
		if c.Context == nil || c.Transcript == nil {
			return fmt.Errorf("eval: report case %q requires context and transcript", c.ID)
		}
		if c.Language != "en" && c.Language != "zh-CN" {
			return fmt.Errorf("eval: report case %q runtime language must be en or zh-CN", c.ID)
		}
		contextRaw, err := json.Marshal(c.Context)
		if err != nil {
			return fmt.Errorf("eval: marshal report case %q context: %w", c.ID, err)
		}
		var reportContext struct {
			Language string `json:"language"`
		}
		if err := json.Unmarshal(contextRaw, &reportContext); err != nil {
			return fmt.Errorf("eval: parse report case %q context: %w", c.ID, err)
		}
		if reportContext.Language != c.Language {
			return fmt.Errorf("eval: report case %q runtime language must match context language", c.ID)
		}
		transcript, err := json.Marshal(c.Transcript)
		var messages []json.RawMessage
		if err != nil || json.Unmarshal(transcript, &messages) != nil || len(messages) == 0 {
			return fmt.Errorf("eval: report case %q transcript must be a non-empty array", c.ID)
		}
		if !c.Redacted {
			return fmt.Errorf("eval: report case %q must declare redacted tracked data", c.ID)
		}
		if c.Critical {
			critical++
		}
		output, err := json.Marshal(c.Output)
		if err != nil {
			return fmt.Errorf("eval: marshal report case %q output: %w", c.ID, err)
		}
		if previous := outputs[string(output)]; previous != "" {
			return fmt.Errorf("eval: report cases %q and %q reuse the same output", previous, c.ID)
		}
		outputs[string(output)] = c.ID
	}
	if count == 0 {
		return nil
	}
	if count != 5 || critical != 3 {
		return fmt.Errorf("eval: grounded report suite requires exactly 5 cases and 3 critical cases; got %d/%d", count, critical)
	}
	return nil
}

// Result is the offline grading outcome for one case.
type Result struct {
	CaseID     string
	FeatureKey string
	Scores     []registry.Score
	Reasoning  registry.Reasoning
	Critical   bool
	Err        error
}

// RubricProvider is the registry surface the harness needs (satisfied by
// *registry.Client).
type RubricProvider interface {
	registry.RubricProvider
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

// fixtureJudgeModel is the offline JudgeModelClient: it returns the recorded
// judge transcript for one case and never makes a network call.
type fixtureJudgeModel struct {
	response string
}

func (f fixtureJudgeModel) CompleteJudge(_ context.Context, _ string, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return aiclient.CompleteResponse{Content: f.response}, aiclient.AICallMeta{Capability: aiclient.CapabilityJudge}, nil
}

// judgeWire mirrors the strict JSON shape registry.LLMJudge expects from a
// judge model so a recorded fixture replays exactly like a live response.
type judgeWire struct {
	Scores                  []wireScore       `json:"scores"`
	Reasoning               wireReasoning     `json:"reasoning"`
	ItemVerdicts            []wireItemVerdict `json:"item_verdicts,omitempty"`
	CausalChecks            []wireCausalCheck `json:"causal_checks,omitempty"`
	ZeroToleranceViolations []string          `json:"zero_tolerance_violations,omitempty"`
	CriticalSafetyPass      *bool             `json:"critical_safety_pass,omitempty"`
}

type wireScore struct {
	Dimension string  `json:"dimension"`
	Value     float64 `json:"value"`
}

type wireReasoning struct {
	Summary        string   `json:"summary"`
	EvidenceQuotes []string `json:"evidence_quotes"`
}

type wireItemVerdict struct {
	Path                    string `json:"path"`
	Kind                    string `json:"kind"`
	Support                 string `json:"support"`
	EvidenceLimitedExplicit bool   `json:"evidence_limited_explicit"`
	UsedForNegativeClaim    bool   `json:"used_for_negative_claim"`
	Reason                  string `json:"reason"`
}

type wireCausalCheck struct {
	DimensionCode   string `json:"dimension_code"`
	IssueSupported  bool   `json:"issue_supported"`
	FocusSupported  bool   `json:"focus_supported"`
	ActionSupported bool   `json:"action_supported"`
	Reason          string `json:"reason"`
}

func toWireScores(scores []DimensionScore) []wireScore {
	out := make([]wireScore, len(scores))
	for i, s := range scores {
		out[i] = wireScore(s)
	}
	return out
}

func toWireItemVerdicts(items []ItemVerdictFixture) []wireItemVerdict {
	out := make([]wireItemVerdict, len(items))
	for i, item := range items {
		out[i] = wireItemVerdict(item)
	}
	return out
}

func toWireCausalChecks(checks []CausalCheckFixture) []wireCausalCheck {
	out := make([]wireCausalCheck, len(checks))
	for i, check := range checks {
		out[i] = wireCausalCheck(check)
	}
	return out
}

// RunOffline grades every case with the real registry.LLMJudge fed by the
// recorded fixture judge transcript. It returns one Result per case; a case
// whose evaluated output or recorded transcript is invalid fails closed and is
// reported via Result.Err and a non-nil aggregate error.
func (s *Suite) RunOffline(ctx context.Context, reg RubricProvider) ([]Result, error) {
	results := make([]Result, 0, len(s.Cases))
	var firstErr error
	for _, c := range s.Cases {
		res := s.gradeOffline(ctx, reg, c)
		if res.Err != nil && firstErr == nil {
			firstErr = fmt.Errorf("eval: case %q failed: %w", c.ID, res.Err)
		}
		results = append(results, res)
	}
	return results, firstErr
}

func (s *Suite) gradeOffline(ctx context.Context, reg RubricProvider, c Case) Result {
	res := Result{CaseID: c.ID, FeatureKey: c.FeatureKey, Critical: c.Critical}
	outputBytes, err := json.Marshal(c.Output)
	if err != nil {
		res.Err = fmt.Errorf("marshal output: %w", err)
		return res
	}
	transcript, err := json.Marshal(judgeWire{
		Scores:                  toWireScores(c.Judge.Scores),
		Reasoning:               wireReasoning{Summary: c.Judge.Reasoning.Summary, EvidenceQuotes: c.Judge.Reasoning.EvidenceQuotes},
		ItemVerdicts:            toWireItemVerdicts(c.Judge.ItemVerdicts),
		CausalChecks:            toWireCausalChecks(c.Judge.CausalChecks),
		ZeroToleranceViolations: c.Judge.ZeroToleranceViolations,
		CriticalSafetyPass:      c.Judge.CriticalSafetyPass,
	})
	if err != nil {
		res.Err = fmt.Errorf("marshal judge fixture: %w", err)
		return res
	}
	resolution, err := ResolveCase(ctx, reg, c)
	if err != nil {
		res.Err = fmt.Errorf("resolve active prompt: %w", err)
		return res
	}
	judge, err := registry.NewLLMJudge(reg, fixtureJudgeModel{response: string(transcript)}, s.Instruction, registry.WithJudgeRubricLanguage("multi"))
	if err != nil {
		res.Err = err
		return res
	}
	// PromptVersion/RubricVersion follow the active registry coordinate so the
	// eval gate tracks future prompt/rubric upgrades instead of pinning v0.1.0.
	if c.FeatureKey == string(featurekeys.ReportGenerate) {
		evidence, evidenceErr := judgeContext(c)
		if evidenceErr != nil {
			res.Err = evidenceErr
			return res
		}
		res.Scores, res.Reasoning, res.Err = judge.JudgeWithContext(ctx, c.FeatureKey, resolution.PromptVersion, outputBytes, resolution.RubricVersion, evidence)
		if res.Err == nil {
			res.Err = validateReportScores(reg, resolution.RubricVersion, res.Scores)
		}
		return res
	}
	res.Scores, res.Reasoning, res.Err = judge.Judge(ctx, c.FeatureKey, resolution.PromptVersion, outputBytes, resolution.RubricVersion)
	return res
}

// ResolveCase returns the active coordinate for ordinary cases or the exact
// inactive candidate coordinate pinned by a pre-activation evaluation case.
func ResolveCase(ctx context.Context, reg RubricProvider, c Case) (registry.PromptResolution, error) {
	resolution, err := reg.ResolveActive(ctx, c.FeatureKey, c.Language)
	if err != nil {
		return registry.PromptResolution{}, err
	}
	if c.PromptVersion == "" && c.RubricVersion == "" {
		return resolution, nil
	}
	if c.PromptVersion == "" || c.RubricVersion == "" {
		return registry.PromptResolution{}, fmt.Errorf("eval: case %q must pin prompt and rubric together", c.ID)
	}
	meta, body, err := reg.GetPrompt(c.FeatureKey, c.PromptVersion, "multi")
	if err != nil {
		return registry.PromptResolution{}, fmt.Errorf("eval: case %q prompt version: %w", c.ID, err)
	}
	if _, err := reg.GetRubric(c.FeatureKey, c.RubricVersion, "multi"); err != nil {
		return registry.PromptResolution{}, fmt.Errorf("eval: case %q rubric version: %w", c.ID, err)
	}
	resolution.PromptVersion = c.PromptVersion
	resolution.RubricVersion = c.RubricVersion
	resolution.UserMessageTemplate = body
	resolution.OutputSchema = meta.OutputSchema
	return resolution, nil
}

func judgeContext(c Case) (registry.JudgeContext, error) {
	contextJSON, err := json.Marshal(c.Context)
	if err != nil {
		return registry.JudgeContext{}, fmt.Errorf("eval: marshal case %q context: %w", c.ID, err)
	}
	transcriptJSON, err := json.Marshal(c.Transcript)
	if err != nil {
		return registry.JudgeContext{}, fmt.Errorf("eval: marshal case %q transcript: %w", c.ID, err)
	}
	return registry.JudgeContext{FrozenContext: contextJSON, Transcript: transcriptJSON}, nil
}

func validateReportScores(reg RubricProvider, rubricVersion string, scores []registry.Score) error {
	_, err := ReportWeightedScore(reg, rubricVersion, scores)
	return err
}

// ReportWeightedScore validates the locked report score thresholds and returns
// the weighted result using the registry rubric as the sole weight source.
// Live UAT consumers use this value instead of copying evaluator weights.
func ReportWeightedScore(reg RubricProvider, rubricVersion string, scores []registry.Score) (float64, error) {
	rubric, err := reg.GetRubric(string(featurekeys.ReportGenerate), rubricVersion, "multi")
	if err != nil {
		return 0, err
	}
	weights := make(map[string]float64, len(rubric.Dimensions))
	for _, dimension := range rubric.Dimensions {
		weights[dimension.Name] = dimension.Weight
	}
	weighted := 0.0
	for _, score := range scores {
		if score.Value < 0.70 {
			return 0, fmt.Errorf("eval: report dimension %s score %.3f is below 0.70", score.Dimension, score.Value)
		}
		weighted += score.Value * weights[score.Dimension]
	}
	if weighted < 0.80 {
		return 0, fmt.Errorf("eval: report weighted score %.3f is below 0.80", weighted)
	}
	return weighted, nil
}

// CaseByID returns the case with the given id.
func (s *Suite) CaseByID(id string) (Case, bool) {
	for _, c := range s.Cases {
		if c.ID == id {
			return c, true
		}
	}
	return Case{}, false
}

// OutputJSON returns the recorded golden output of a case as canonical JSON.
func (c Case) OutputJSON() ([]byte, error) {
	return json.Marshal(c.Output)
}

// JudgeTranscript returns the recorded judge transcript of a case as the strict
// JSON a judge model would return.
func (c Case) JudgeTranscript() ([]byte, error) {
	return json.Marshal(judgeWire{
		Scores:                  toWireScores(c.Judge.Scores),
		Reasoning:               wireReasoning{Summary: c.Judge.Reasoning.Summary, EvidenceQuotes: c.Judge.Reasoning.EvidenceQuotes},
		ItemVerdicts:            toWireItemVerdicts(c.Judge.ItemVerdicts),
		CausalChecks:            toWireCausalChecks(c.Judge.CausalChecks),
		ZeroToleranceViolations: c.Judge.ZeroToleranceViolations,
		CriticalSafetyPass:      c.Judge.CriticalSafetyPass,
	})
}

// GradeOutput grades a candidate output for a case. In offline mode model is a
// fixtureJudgeModel; live callers pass a real JudgeModelClient. It reuses the
// single registry.LLMJudge implementation.
func (s *Suite) GradeOutput(ctx context.Context, reg RubricProvider, model registry.JudgeModelClient, c Case, output []byte) ([]registry.Score, registry.Reasoning, error) {
	resolution, err := ResolveCase(ctx, reg, c)
	if err != nil {
		return nil, registry.Reasoning{}, fmt.Errorf("eval: resolve %s/%s: %w", c.FeatureKey, c.Language, err)
	}
	judge, err := registry.NewLLMJudge(reg, model, s.Instruction, registry.WithJudgeRubricLanguage("multi"))
	if err != nil {
		return nil, registry.Reasoning{}, err
	}
	if c.FeatureKey == string(featurekeys.ReportGenerate) {
		evidence, err := judgeContext(c)
		if err != nil {
			return nil, registry.Reasoning{}, err
		}
		scores, reasoning, err := judge.JudgeWithContext(ctx, c.FeatureKey, resolution.PromptVersion, output, resolution.RubricVersion, evidence)
		if err != nil {
			return nil, registry.Reasoning{}, err
		}
		if err := validateReportScores(reg, resolution.RubricVersion, scores); err != nil {
			return nil, registry.Reasoning{}, err
		}
		return scores, reasoning, nil
	}
	return judge.Judge(ctx, c.FeatureKey, resolution.PromptVersion, output, resolution.RubricVersion)
}

// OfflineJudgeModel returns the recorded-fixture JudgeModelClient for a case.
func (c Case) OfflineJudgeModel() (registry.JudgeModelClient, error) {
	transcript, err := c.JudgeTranscript()
	if err != nil {
		return nil, err
	}
	return fixtureJudgeModel{response: string(transcript)}, nil
}

// MarshalResolved serializes a ResolveAll map to stable JSON for the
// single-source export artifact and drift gate.
func MarshalResolved(resolved map[string]registry.PromptResolution) ([]byte, error) {
	type entry struct {
		FeatureKey          string `json:"feature_key"`
		PromptVersion       string `json:"prompt_version"`
		RubricVersion       string `json:"rubric_version"`
		ModelProfileName    string `json:"model_profile_name"`
		DataSourceVersion   string `json:"data_source_version"`
		SystemMessage       string `json:"system_message"`
		UserMessageTemplate string `json:"user_message_template"`
	}
	out := map[string]entry{}
	for k, r := range resolved {
		out[k] = entry{
			FeatureKey:          r.FeatureKey,
			PromptVersion:       r.PromptVersion,
			RubricVersion:       r.RubricVersion,
			ModelProfileName:    r.ModelProfileName,
			DataSourceVersion:   r.DataSourceVersion,
			SystemMessage:       r.SystemMessage,
			UserMessageTemplate: r.UserMessageTemplate,
		}
	}
	return json.MarshalIndent(out, "", "  ")
}

// ResolveAll returns the registry-resolved prompt for every distinct
// (feature_key, language) used by the suite. The map key is
// "<feature_key>|<language>". This is the single-source export the Promptfoo
// runner and the drift gate consume; the prompt text is never copied into the
// eval assets.
func (s *Suite) ResolveAll(ctx context.Context, reg RubricProvider) (map[string]registry.PromptResolution, error) {
	out := map[string]registry.PromptResolution{}
	for _, c := range s.Cases {
		key := c.FeatureKey + "|" + c.Language
		if _, ok := out[key]; ok {
			continue
		}
		res, err := ResolveCase(ctx, reg, c)
		if err != nil {
			return nil, fmt.Errorf("eval: resolve %s/%s: %w", c.FeatureKey, c.Language, err)
		}
		out[key] = res
	}
	return out, nil
}
