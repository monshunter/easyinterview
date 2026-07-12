// Command evalkit is the F3 offline evaluation harness CLI (plan
// prompt-rubric-registry/004 §4). It is the Go backend the pinned Promptfoo
// runner shells out to: Promptfoo orchestrates and reports, while evalkit owns
// registry single-source resolution, schema-validated grading via the single
// registry.LLMJudge, the >=24 count gate, and the registry-single-source drift
// gate. The default path is deterministic and makes no network call; EVAL_LIVE=1
// opts into real provider/judge calls.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/bootstrap"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/eval"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
	"gopkg.in/yaml.v3"
)

const minCases = 24

type paths struct {
	evals    string
	prompts  string
	rubrics  string
	artifact string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: evalkit <resolve|drift-check|run|complete|grade|prompts-tests|version> [flags]")
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	var p paths
	var live liveOpts
	fs.StringVar(&p.evals, "evals", "config/evals", "eval suite directory")
	fs.StringVar(&p.prompts, "prompts", "config/prompts", "prompts truth source directory")
	fs.StringVar(&p.rubrics, "rubrics", "config/rubrics", "rubrics truth source directory")
	fs.StringVar(&p.artifact, "artifact", "config/evals/resolved-prompts.json", "registry-resolved single-source export artifact")
	caseID := fs.String("case", "", "case id (complete/grade)")
	outFile := fs.String("out", "", "output file (resolve/prompts-tests); default stdout")
	outputArg := fs.String("output", "", "candidate output JSON to grade (grade); default stdin")
	// --live opts into real provider/judge calls. The EVAL_LIVE env opt-in is
	// translated to this flag by the Makefile / Promptfoo bridge so this binary
	// never reads os.Getenv directly (secrets-and-config §4.1 boundary).
	fs.BoolVar(&live.enabled, "live", false, "opt into real provider/judge calls (EVAL_LIVE)")
	fs.StringVar(&live.appEnv, "app-env", "dev", "APP_ENV for the live runtime")
	fs.StringVar(&live.configDir, "config-dir", "config", "config directory for the live runtime")
	_ = fs.Parse(os.Args[2:])

	if err := run(cmd, p, *caseID, *outFile, *outputArg, live); err != nil {
		fmt.Fprintf(os.Stderr, "evalkit %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

// liveOpts carries the EVAL_LIVE opt-in and the live runtime config inputs.
type liveOpts struct {
	enabled   bool
	appEnv    string
	configDir string
}

func run(cmd string, p paths, caseID, outFile, outputArg string, live liveOpts) error {
	switch cmd {
	case "version":
		fmt.Println("evalkit 1.0.0")
		return nil
	case "resolve":
		return cmdResolve(p, outFile)
	case "drift-check":
		return cmdDriftCheck(p)
	case "run":
		return cmdRun(p, live)
	case "complete":
		return cmdComplete(p, caseID, live)
	case "grade":
		return cmdGrade(p, caseID, outputArg, live)
	case "prompts-tests":
		return cmdPromptsTests(p, outFile)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func loadSuiteAndRegistry(p paths) (*eval.Suite, *registry.Client, error) {
	suite, err := eval.LoadSuite(p.evals)
	if err != nil {
		return nil, nil, err
	}
	reg, err := registry.NewRegistryClient(registry.RegistryOptions{PromptsDir: p.prompts, RubricsDir: p.rubrics})
	if err != nil {
		return nil, nil, err
	}
	return suite, reg, nil
}

func cmdResolve(p paths, outFile string) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	resolved, err := suite.ResolveAll(context.Background(), reg)
	if err != nil {
		return err
	}
	data, err := eval.MarshalResolved(resolved)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	target := outFile
	if target == "" {
		target = p.artifact
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644)
}

func cmdDriftCheck(p paths) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	if suite.Count() < minCases {
		return fmt.Errorf("offline eval suite has %d cases, need >= %d", suite.Count(), minCases)
	}
	resolved, err := suite.ResolveAll(context.Background(), reg)
	if err != nil {
		return err
	}
	current, err := eval.MarshalResolved(resolved)
	if err != nil {
		return err
	}
	committed, err := os.ReadFile(p.artifact)
	if err != nil {
		return fmt.Errorf("read export artifact (run 'evalkit resolve' first): %w", err)
	}
	if !bytes.Equal(bytes.TrimSpace(committed), bytes.TrimSpace(current)) {
		return fmt.Errorf("registry single-source drift: %s does not match current registry resolution; re-run 'make eval-offline-resolve'", p.artifact)
	}
	fmt.Printf("evalkit drift-check: OK (%d cases, %d resolved prompts, single-source clean)\n", suite.Count(), len(resolved))
	return nil
}

func cmdRun(p paths, live liveOpts) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	if suite.Count() < minCases {
		return fmt.Errorf("offline eval suite has %d cases, need >= %d", suite.Count(), minCases)
	}
	ctx := context.Background()
	if live.enabled {
		return runLive(ctx, suite, reg, live)
	}
	results, err := suite.RunOffline(ctx, reg)
	if err != nil {
		return err
	}
	fmt.Printf("evalkit run: OK (offline, %d cases graded, no network)\n", len(results))
	return nil
}

func cmdComplete(p paths, caseID string, live liveOpts) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	c, ok := suite.CaseByID(caseID)
	if !ok {
		return fmt.Errorf("case %q not found", caseID)
	}
	if live.enabled {
		out, err := liveComplete(context.Background(), reg, c, live)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	}
	out, err := c.OutputJSON()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func cmdGrade(p paths, caseID, outputArg string, live liveOpts) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	c, ok := suite.CaseByID(caseID)
	if !ok {
		return fmt.Errorf("case %q not found", caseID)
	}
	output := []byte(outputArg)
	if len(output) == 0 {
		output, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}
	ctx := context.Background()
	var model registry.JudgeModelClient
	if live.enabled {
		model, err = liveJudgeModel(live)
	} else {
		model, err = c.OfflineJudgeModel()
	}
	if err != nil {
		return err
	}
	scores, reasoning, gradeErr := suite.GradeOutput(ctx, reg, model, c, bytes.TrimSpace(output))
	verdict := map[string]any{
		"pass":   gradeErr == nil,
		"caseId": c.ID,
	}
	if gradeErr != nil {
		verdict["reason"] = gradeErr.Error()
	} else {
		verdict["scores"] = scores
		verdict["reason"] = reasoning.Summary
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(verdict); err != nil {
		return err
	}
	if gradeErr != nil {
		return gradeErr
	}
	return nil
}

func cmdPromptsTests(p paths, outFile string) error {
	suite, _, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	type testVars struct {
		CaseID string `yaml:"caseId"`
		Input  string `yaml:"input"`
	}
	type test struct {
		Description string   `yaml:"description"`
		Vars        testVars `yaml:"vars"`
	}
	tests := make([]test, 0, suite.Count())
	for _, c := range suite.Cases {
		tests = append(tests, test{
			Description: c.ID,
			Vars:        testVars{CaseID: c.ID, Input: c.Input},
		})
	}
	data, err := yaml.Marshal(tests)
	if err != nil {
		return err
	}
	target := outFile
	if target == "" {
		fmt.Print(string(data))
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644)
}

// --- live (EVAL_LIVE -> --live) wiring; opt-in, not exercised by make test/eval-offline ---
//
// The live runtime loads provider config and secrets through the allowlisted
// platform/config + platform/secrets composition root, so this command never
// reads os.Getenv itself (secrets-and-config §4.1 boundary).

func liveRuntime(live liveOpts) (*bootstrap.Runtime, error) {
	loader, err := config.LoadCanonical(config.CanonicalOptions{
		AppEnv:       live.appEnv,
		ConfigDir:    live.configDir,
		SecretSource: secrets.EnvSecretSource{},
	})
	if err != nil {
		return nil, err
	}
	return bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               loader.AppEnv(),
			ProviderRegistryPath: loader.GetString("ai.providerRegistryPath"),
			ModelProfilePath:     loader.GetString("ai.modelProfilePath"),
		},
		SecretSource: secrets.EnvSecretSource{},
	})
}

func liveComplete(ctx context.Context, reg *registry.Client, c eval.Case, live liveOpts) (string, error) {
	res, err := reg.ResolveActive(ctx, c.FeatureKey, c.Language)
	if err != nil {
		return "", err
	}
	rt, err := liveRuntime(live)
	if err != nil {
		return "", err
	}
	defer rt.Close()
	payload := aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: res.SystemMessage},
			{Role: "user", Content: res.UserMessageTemplate + "\n\nInput:\n" + c.Input},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    res.FeatureKey,
			PromptVersion: res.PromptVersion,
			RubricVersion: res.RubricVersion,
			Language:      c.Language,
		},
	}
	resp, _, err := rt.Client.Complete(ctx, res.ModelProfileName, payload)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func liveJudgeModel(live liveOpts) (registry.JudgeModelClient, error) {
	rt, err := liveRuntime(live)
	if err != nil {
		return nil, err
	}
	// Runtime is intentionally not closed here: the judge model is used for the
	// duration of a single grade call in the process.
	return rt.Client, nil
}

func runLive(ctx context.Context, suite *eval.Suite, reg *registry.Client, live liveOpts) error {
	rt, err := liveRuntime(live)
	if err != nil {
		return err
	}
	defer rt.Close()
	for _, c := range suite.Cases {
		out, err := liveComplete(ctx, reg, c, live)
		if err != nil {
			return fmt.Errorf("case %q live complete: %w", c.ID, err)
		}
		if _, _, err := suite.GradeOutput(ctx, reg, rt.Client, c, []byte(out)); err != nil {
			return fmt.Errorf("case %q live grade: %w", c.ID, err)
		}
	}
	fmt.Printf("evalkit run: OK (live, %d cases)\n", suite.Count())
	return nil
}
