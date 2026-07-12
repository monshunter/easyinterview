package eval_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/eval"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Dir(dir) // backend/.. == repo root
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Skipf("could not locate backend go.mod from %s", wd)
		}
		dir = parent
	}
}

func repoRegistry(t *testing.T, root string) *registry.Client {
	t.Helper()
	c, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: filepath.Join(root, "config", "prompts"),
		RubricsDir: filepath.Join(root, "config", "rubrics"),
	})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	return c
}

// writeTempSuite builds a minimal but valid eval suite for follow_up so the
// runner logic is exercised without depending on the full committed suite.
func writeTempSuite(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "judge-instruction.md"), []byte("Offline judge. Score each dimension 0..1 as strict JSON."), 0o600); err != nil {
		t.Fatalf("write instruction: %v", err)
	}
	fkDir := filepath.Join(dir, "practice.session.chat")
	if err := os.MkdirAll(fkDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := `feature_key: practice.session.chat
cases:
  - id: practice.session.chat-multi-strong
    language: multi
    input: "candidate gave a shallow answer about a rollback"
    output:
      messageText: "Walk me through exactly how you verified the rollback succeeded."
    judge:
      scores:
        - dimension: followup_relevance
          value: 0.9
        - dimension: practice_depth
          value: 0.85
        - dimension: language_consistency
          value: 0.95
      reasoning:
        summary: "Targets the verification gap and stays on-language."
        evidence_quotes: []
  - id: practice.session.chat-en-fallback
    language: en
    input: "english request that should fall back to multi baseline"
    output:
      messageText: "What signal told you the incident was contained?"
    judge:
      scores:
        - dimension: followup_relevance
          value: 0.6
        - dimension: practice_depth
          value: 0.55
        - dimension: language_consistency
          value: 0.7
      reasoning:
        summary: "Reasonable follow-up with moderate depth."
        evidence_quotes: []
`
	if err := os.WriteFile(filepath.Join(fkDir, "cases.yaml"), []byte(body), 0o600); err != nil {
		t.Fatalf("write cases: %v", err)
	}
	return dir
}

func TestLoadSuite(t *testing.T) {
	suite, err := eval.LoadSuite(writeTempSuite(t))
	if err != nil {
		t.Fatalf("LoadSuite: %v", err)
	}
	if suite.Instruction == "" {
		t.Fatal("instruction must be loaded")
	}
	if suite.Count() != 2 {
		t.Fatalf("Count: want 2, got %d", suite.Count())
	}
}

func TestRunOfflineGradesEachCase(t *testing.T) {
	root := repoRoot(t)
	reg := repoRegistry(t, root)
	suite, err := eval.LoadSuite(writeTempSuite(t))
	if err != nil {
		t.Fatalf("LoadSuite: %v", err)
	}
	results, err := suite.RunOffline(context.Background(), reg)
	if err != nil {
		t.Fatalf("RunOffline: %v", err)
	}
	if len(results) != suite.Count() {
		t.Fatalf("results: want %d, got %d", suite.Count(), len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("case %s failed offline grading: %v", r.CaseID, r.Err)
		}
		if len(r.Scores) == 0 {
			t.Fatalf("case %s produced no scores", r.CaseID)
		}
	}
}

func TestResolveAllSingleSource(t *testing.T) {
	root := repoRoot(t)
	reg := repoRegistry(t, root)
	suite, err := eval.LoadSuite(writeTempSuite(t))
	if err != nil {
		t.Fatalf("LoadSuite: %v", err)
	}
	resolved, err := suite.ResolveAll(context.Background(), reg)
	if err != nil {
		t.Fatalf("ResolveAll: %v", err)
	}
	if _, ok := resolved["practice.session.chat|multi"]; !ok {
		t.Fatalf("expected resolved prompt for follow_up|multi, got keys %v", keys(resolved))
	}
	// en request must fall back to the multi baseline resolution.
	if _, ok := resolved["practice.session.chat|en"]; !ok {
		t.Fatalf("expected resolved prompt for follow_up|en fallback, got keys %v", keys(resolved))
	}
}

// TestRealSuiteOfflineGreen is the count>=24 + offline-grades-clean gate over
// the committed config/evals suite (plan 004 §4.1/§4.5). It runs with no
// AI_PROVIDER env and must not touch the network.
func TestRealSuiteOfflineGreen(t *testing.T) {
	root := repoRoot(t)
	suite, err := eval.LoadSuite(filepath.Join(root, "config", "evals"))
	if err != nil {
		t.Fatalf("LoadSuite(real): %v", err)
	}
	// Baseline is 24: 6 active feature keys each keep
	// the four quality-band cases after jd_match and debrief/profile removal.
	if suite.Count() < 24 {
		t.Fatalf("offline eval suite must have >= 24 cases, got %d", suite.Count())
	}
	reg := repoRegistry(t, root)
	results, err := suite.RunOffline(context.Background(), reg)
	if err != nil {
		t.Fatalf("RunOffline(real): %v", err)
	}
	hasEnFallback := false
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("case %s failed offline grading: %v", r.CaseID, r.Err)
		}
	}
	for _, c := range suite.Cases {
		if c.Language == "en" {
			hasEnFallback = true
		}
	}
	if !hasEnFallback {
		t.Fatal("suite must include at least one en->multi fallback case")
	}
}

// TestRunOfflineMakesNoNetworkCall asserts the offline path is hermetic: even
// with provider env pointed at an unroutable address, offline grading succeeds
// because it replays the recorded fixture judge and never constructs a network
// client (plan 004 §4.4).
func TestRunOfflineMakesNoNetworkCall(t *testing.T) {
	t.Setenv("AI_PROVIDER_BASE_URL", "http://203.0.113.0:9")
	t.Setenv("AI_PROVIDER_API_KEY", "should-never-be-used")
	root := repoRoot(t)
	reg := repoRegistry(t, root)
	suite, err := eval.LoadSuite(writeTempSuite(t))
	if err != nil {
		t.Fatalf("LoadSuite: %v", err)
	}
	if _, err := suite.RunOffline(context.Background(), reg); err != nil {
		t.Fatalf("offline run must not depend on the network: %v", err)
	}
}

func TestRunOfflineUsesResolveActiveVersions(t *testing.T) {
	reg := &activeVersionRegistry{promptVersion: "v9.9.9", rubricVersion: "v8.8.8"}
	suite := &eval.Suite{
		Instruction: "score each dimension",
		Cases:       []eval.Case{versionedCase()},
	}

	results, err := suite.RunOffline(context.Background(), reg)
	if err != nil {
		t.Fatalf("RunOffline: %v", err)
	}
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("RunOffline result = %+v", results)
	}
	if got := reg.promptLookups; len(got) != 1 || got[0] != "v9.9.9" {
		t.Fatalf("prompt lookup versions: got %v, want [v9.9.9]", got)
	}
	if got := reg.rubricLookups; len(got) != 1 || got[0] != "v8.8.8" {
		t.Fatalf("rubric lookup versions: got %v, want [v8.8.8]", got)
	}
}

func TestGradeOutputUsesResolveActiveVersions(t *testing.T) {
	reg := &activeVersionRegistry{promptVersion: "v9.9.9", rubricVersion: "v8.8.8"}
	suite := &eval.Suite{Instruction: "score each dimension"}
	c := versionedCase()
	model, err := c.OfflineJudgeModel()
	if err != nil {
		t.Fatalf("OfflineJudgeModel: %v", err)
	}
	output, err := c.OutputJSON()
	if err != nil {
		t.Fatalf("OutputJSON: %v", err)
	}

	if _, _, err := suite.GradeOutput(context.Background(), reg, model, c, output); err != nil {
		t.Fatalf("GradeOutput: %v", err)
	}
	if got := reg.promptLookups; len(got) != 1 || got[0] != "v9.9.9" {
		t.Fatalf("prompt lookup versions: got %v, want [v9.9.9]", got)
	}
	if got := reg.rubricLookups; len(got) != 1 || got[0] != "v8.8.8" {
		t.Fatalf("rubric lookup versions: got %v, want [v8.8.8]", got)
	}
}

type activeVersionRegistry struct {
	promptVersion string
	rubricVersion string
	promptLookups []string
	rubricLookups []string
}

func (r *activeVersionRegistry) ResolveActive(_ context.Context, featureKey, _ string) (registry.PromptResolution, error) {
	return registry.PromptResolution{
		FeatureKey:       featureKey,
		PromptVersion:    r.promptVersion,
		RubricVersion:    r.rubricVersion,
		ModelProfileName: "practice.followup.default",
	}, nil
}

func (r *activeVersionRegistry) GetPrompt(featureKey, version, language string) (registry.PromptMeta, string, error) {
	r.promptLookups = append(r.promptLookups, version)
	if version != r.promptVersion {
		return registry.PromptMeta{}, "", registry.ErrPromptUnsupported
	}
	return registry.PromptMeta{
		FeatureKey: featureKey,
		Version:    version,
		Language:   language,
		Status:     "active",
	}, "", nil
}

func (r *activeVersionRegistry) GetRubric(featureKey, version, language string) (registry.RubricSchema, error) {
	r.rubricLookups = append(r.rubricLookups, version)
	if version != r.rubricVersion {
		return registry.RubricSchema{}, registry.ErrPromptUnsupported
	}
	return registry.RubricSchema{
		FeatureKey: featureKey,
		Version:    version,
		Language:   language,
		Dimensions: []registry.RubricDimension{
			{
				Name:        "followup_relevance",
				Description: "Follow-up relevance",
				ScoreLevels: []registry.ScoreLevel{
					{Label: "weak", Threshold: 0.0, Description: "Weak"},
					{Label: "strong", Threshold: 0.8, Description: "Strong"},
				},
			},
		},
	}, nil
}

func versionedCase() eval.Case {
	c := eval.Case{
		ID:         "versioned-active-case",
		FeatureKey: "practice.session.chat",
		Language:   "multi",
		Input:      "candidate gave a shallow answer",
		Output: map[string]any{
			"questionText":   "What evidence told you the rollback worked?",
			"questionIntent": "probe-evidence",
		},
	}
	c.Judge.Scores = []eval.DimensionScore{{Dimension: "followup_relevance", Value: 0.8}}
	c.Judge.Reasoning.Summary = "Scores the active rubric version."
	return c
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
