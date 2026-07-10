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
}

// Case is one offline evaluation case.
type Case struct {
	ID         string       `yaml:"id"`
	FeatureKey string       `yaml:"-"`
	Language   string       `yaml:"language"`
	Input      string       `yaml:"input"`
	Output     any          `yaml:"output"`
	Judge      JudgeFixture `yaml:"judge"`
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
	return suite, nil
}

// Result is the offline grading outcome for one case.
type Result struct {
	CaseID     string
	FeatureKey string
	Scores     []registry.Score
	Reasoning  registry.Reasoning
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
	Scores    []wireScore   `json:"scores"`
	Reasoning wireReasoning `json:"reasoning"`
}

type wireScore struct {
	Dimension string  `json:"dimension"`
	Value     float64 `json:"value"`
}

type wireReasoning struct {
	Summary        string   `json:"summary"`
	EvidenceQuotes []string `json:"evidence_quotes"`
}

func toWireScores(scores []DimensionScore) []wireScore {
	out := make([]wireScore, len(scores))
	for i, s := range scores {
		out[i] = wireScore(s)
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
	res := Result{CaseID: c.ID, FeatureKey: c.FeatureKey}
	outputBytes, err := json.Marshal(c.Output)
	if err != nil {
		res.Err = fmt.Errorf("marshal output: %w", err)
		return res
	}
	transcript, err := json.Marshal(judgeWire{
		Scores:    toWireScores(c.Judge.Scores),
		Reasoning: wireReasoning{Summary: c.Judge.Reasoning.Summary, EvidenceQuotes: c.Judge.Reasoning.EvidenceQuotes},
	})
	if err != nil {
		res.Err = fmt.Errorf("marshal judge fixture: %w", err)
		return res
	}
	resolution, err := reg.ResolveActive(ctx, c.FeatureKey, c.Language)
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
	res.Scores, res.Reasoning, res.Err = judge.Judge(ctx, c.FeatureKey, resolution.PromptVersion, outputBytes, resolution.RubricVersion)
	return res
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
		Scores:    toWireScores(c.Judge.Scores),
		Reasoning: wireReasoning{Summary: c.Judge.Reasoning.Summary, EvidenceQuotes: c.Judge.Reasoning.EvidenceQuotes},
	})
}

// GradeOutput grades a candidate output for a case. In offline mode model is a
// fixtureJudgeModel; live callers pass a real JudgeModelClient. It reuses the
// single registry.LLMJudge implementation.
func (s *Suite) GradeOutput(ctx context.Context, reg RubricProvider, model registry.JudgeModelClient, c Case, output []byte) ([]registry.Score, registry.Reasoning, error) {
	resolution, err := reg.ResolveActive(ctx, c.FeatureKey, c.Language)
	if err != nil {
		return nil, registry.Reasoning{}, fmt.Errorf("eval: resolve %s/%s: %w", c.FeatureKey, c.Language, err)
	}
	judge, err := registry.NewLLMJudge(reg, model, s.Instruction, registry.WithJudgeRubricLanguage("multi"))
	if err != nil {
		return nil, registry.Reasoning{}, err
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
		res, err := reg.ResolveActive(ctx, c.FeatureKey, c.Language)
		if err != nil {
			return nil, fmt.Errorf("eval: resolve %s/%s: %w", c.FeatureKey, c.Language, err)
		}
		out[key] = res
	}
	return out, nil
}
