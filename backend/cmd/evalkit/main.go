// Command evalkit is the F3 offline evaluation harness CLI (plan
// prompt-rubric-registry/004 §4). It is the Go backend the pinned Promptfoo
// runner shells out to: Promptfoo orchestrates and reports, while evalkit owns
// registry single-source resolution, schema-validated grading via the single
// registry.LLMJudge, the exact-28 count gate, and the registry-single-source drift
// gate. The default path is deterministic and makes no network call; EVAL_LIVE=1
// opts into real provider/judge calls.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/bootstrap"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	"github.com/monshunter/easyinterview/backend/internal/eval"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/platform/secrets"
	"github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"gopkg.in/yaml.v3"
)

const exactCases = 28

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
	auditOut := fs.String("audit-out", "", "write a redacted live-call audit JSON file (complete/grade --live)")
	// --live opts into real provider/judge calls. The EVAL_LIVE env opt-in is
	// translated to this flag by the Makefile / Promptfoo bridge so this binary
	// never reads os.Getenv directly (secrets-and-config §4.1 boundary).
	fs.BoolVar(&live.enabled, "live", false, "opt into real provider/judge calls (EVAL_LIVE)")
	fs.StringVar(&live.appEnv, "app-env", "dev", "APP_ENV for the live runtime")
	fs.StringVar(&live.configDir, "config-dir", "config", "config directory for the live runtime")
	_ = fs.Parse(os.Args[2:])

	if err := run(cmd, p, *caseID, *outFile, *outputArg, *auditOut, live); err != nil {
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

func run(cmd string, p paths, caseID, outFile, outputArg, auditOut string, live liveOpts) error {
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
		return cmdComplete(p, caseID, auditOut, live)
	case "grade":
		return cmdGrade(p, caseID, outputArg, auditOut, live)
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
	if err := validateCaseCount(suite.Count()); err != nil {
		return err
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
	if err := validateCaseCount(suite.Count()); err != nil {
		return err
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

func validateCaseCount(count int) error {
	if count != exactCases {
		return fmt.Errorf("offline eval suite has %d cases, need exactly %d", count, exactCases)
	}
	return nil
}

func cmdComplete(p paths, caseID, auditOut string, live liveOpts) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	c, ok := suite.CaseByID(caseID)
	if !ok {
		return fmt.Errorf("case %q not found", caseID)
	}
	if auditOut != "" && !live.enabled {
		return fmt.Errorf("--audit-out requires --live")
	}
	if live.enabled {
		result, callErr := liveComplete(context.Background(), reg, c, live)
		if auditOut != "" {
			if err := writeLiveAudit(auditOut, buildCompletionAudit(c, result, callErr)); err != nil {
				return err
			}
		}
		if callErr != nil {
			return callErr
		}
		fmt.Println(result.response.Content)
		return nil
	}
	out, err := c.OutputJSON()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func cmdGrade(p paths, caseID, outputArg, auditOut string, live liveOpts) error {
	suite, reg, err := loadSuiteAndRegistry(p)
	if err != nil {
		return err
	}
	c, ok := suite.CaseByID(caseID)
	if !ok {
		return fmt.Errorf("case %q not found", caseID)
	}
	if auditOut != "" && !live.enabled {
		return fmt.Errorf("--audit-out requires --live")
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
	var capture *capturingJudgeModel
	var runtime *bootstrap.Runtime
	if live.enabled {
		runtime, err = liveRuntime(live)
		if err == nil {
			defer runtime.Close()
			capture = &capturingJudgeModel{delegate: runtime.Client}
			model = capture
		}
	} else {
		model, err = c.OfflineJudgeModel()
	}
	if err != nil {
		return err
	}
	scores, reasoning, gradeErr := suite.GradeOutput(ctx, reg, model, c, bytes.TrimSpace(output))
	if live.enabled && gradeErr == nil {
		resolution, resolveErr := eval.ResolveCase(ctx, reg, c)
		if resolveErr != nil {
			gradeErr = resolveErr
		} else {
			gradeErr = validateLiveJudgeCall(capture.response, capture.meta, resolution)
		}
	}
	var weightedScore *float64
	if gradeErr == nil && c.FeatureKey == string(featurekeys.ReportGenerate) {
		resolution, resolveErr := eval.ResolveCase(ctx, reg, c)
		if resolveErr != nil {
			gradeErr = resolveErr
		} else {
			weighted, weightErr := eval.ReportWeightedScore(reg, resolution.RubricVersion, scores)
			if weightErr != nil {
				gradeErr = weightErr
			} else {
				weightedScore = &weighted
			}
		}
	}
	if auditOut != "" {
		if err := writeLiveAudit(auditOut, buildJudgeAudit(c, capture, gradeErr)); err != nil {
			return err
		}
	}
	verdict := buildGradeVerdict(c, scores, reasoning, weightedScore, gradeErr)
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
	limits, err := loader.ContentLimits()
	if err != nil {
		return nil, err
	}
	return bootstrap.NewClient(bootstrap.Options{
		Config: aiclient.Config{
			AppEnv:               loader.AppEnv(),
			ProviderRegistryPath: loader.GetString("ai.providerRegistryPath"),
			ModelProfilePath:     loader.GetString("ai.modelProfilePath"),
		},
		SecretSource:         secrets.EnvSecretSource{},
		MaxResponseBodyBytes: limits.AIProviderMaxResponseBodyBytes,
	})
}

type liveCompletionResult struct {
	resolution   registry.PromptResolution
	response     aiclient.CompleteResponse
	meta         aiclient.AICallMeta
	repairUsed   bool
	repairScope  string
	attemptCount int
	retryReasons []string
	repairScopes []string
}

const (
	repairScopeNone         = "none"
	repairScopeWholeReport  = "whole_report"
	repairScopeActionLabels = "action_labels"
	maxLiveReportAttempts   = 4

	retryReasonProviderRetryable = "provider_retryable"
	retryReasonOutputSchema      = "output_schema_invalid"
	retryReasonOutputSemantic    = "output_semantic_invalid"
)

type liveReportCompleter interface {
	Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error)
}

func liveComplete(ctx context.Context, reg *registry.Client, c eval.Case, live liveOpts) (liveCompletionResult, error) {
	res, err := eval.ResolveCase(ctx, reg, c)
	if err != nil {
		return liveCompletionResult{}, err
	}
	result := liveCompletionResult{resolution: res, repairScope: repairScopeNone}
	rt, err := liveRuntime(live)
	if err != nil {
		return result, err
	}
	defer rt.Close()
	messages, err := renderCaseMessages(res, c)
	if err != nil {
		return result, err
	}
	var outputSchema json.RawMessage
	if res.OutputSchema != nil {
		outputSchema = append(json.RawMessage(nil), (*res.OutputSchema)...)
	}
	payload := aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey:        res.FeatureKey,
			PromptVersion:     res.PromptVersion,
			RubricVersion:     res.RubricVersion,
			Language:          c.Language,
			FeatureFlag:       res.FeatureFlag,
			DataSourceVersion: res.DataSourceVersion,
			OutputSchema:      outputSchema,
		},
	}
	if c.FeatureKey == string(featurekeys.ReportGenerate) {
		return completeLiveReportWithRepair(ctx, rt.Client, res, c, payload)
	}
	result.response, result.meta, err = rt.Client.Complete(ctx, res.ModelProfileName, payload)
	result.attemptCount = 1
	if err != nil {
		return result, err
	}
	return result, nil
}

func completeLiveReportWithRepair(
	ctx context.Context,
	client liveReportCompleter,
	resolution registry.PromptResolution,
	c eval.Case,
	payload aiclient.CompletePayload,
) (liveCompletionResult, error) {
	result := liveCompletionResult{resolution: resolution, repairScope: repairScopeNone}
	validationContext, reportMessages, err := liveReportValidationInputs(c)
	if err != nil {
		return result, err
	}

	currentPayload := payload
	currentScope := repairScopeNone
	var targetedBase *review.ReportContentDraft
	var targetedIssues []review.ReportValidationIssue

	for attempt := 1; attempt <= maxLiveReportAttempts; attempt++ {
		response, meta, callErr := client.Complete(ctx, resolution.ModelProfileName, currentPayload)
		result.attemptCount = attempt
		result.response = response
		if attempt == 1 {
			result.meta = meta
		} else {
			result.meta = review.AggregateReportRepairMeta(result.meta, meta)
		}

		if callErr != nil && !isLiveReportOutputInvalid(callErr) {
			if attempt < maxLiveReportAttempts && retryableLiveProviderError(ctx, callErr) {
				result.retryReasons = append(result.retryReasons, retryReasonProviderRetryable)
				result.repairScopes = append(result.repairScopes, currentScope)
				continue
			}
			return result, callErr
		}
		if callErr != nil {
			if invalidErr := validateLiveReportInvalidCall(response, meta, resolution, c.Language); invalidErr != nil {
				return result, invalidErr
			}
		} else if callErr = validateLiveReportCall(response, meta, resolution, c.Language); callErr != nil {
			return result, callErr
		}

		if targetedBase != nil {
			merged, mergeErr := review.MergeReportActionLabelRepair(*targetedBase, c.Language, targetedIssues, response.Content)
			if mergeErr != nil {
				if attempt == maxLiveReportAttempts {
					return result, mergeErr
				}
				result.retryReasons = append(result.retryReasons, retryReasonOutputSchema)
				result.repairScopes = append(result.repairScopes, repairScopeActionLabels)
				continue
			}
			mergedRaw, marshalErr := json.Marshal(merged)
			if marshalErr != nil {
				return result, fmt.Errorf("marshal targeted action label repair: %w", marshalErr)
			}
			result.response.Content = string(mergedRaw)
		}

		validation := validateLiveReportProductOutput(result.response, resolution, validationContext, reportMessages)
		if validation.valid() {
			return result, nil
		}
		if attempt == maxLiveReportAttempts {
			return result, validation.err()
		}

		repairPayload, scope, base, issues, buildErr := buildLiveReportRetryPayload(
			payload,
			resolution,
			c,
			result.response.Content,
			validation,
			validationContext,
			reportMessages,
		)
		if buildErr != nil {
			return result, buildErr
		}
		currentPayload = repairPayload
		currentScope = scope
		targetedBase = base
		targetedIssues = issues
		result.repairUsed = true
		result.repairScope = scope
		result.retryReasons = append(result.retryReasons, validation.retryReason())
		result.repairScopes = append(result.repairScopes, scope)
	}
	return result, fmt.Errorf("live report completion attempt budget exhausted")
}

func buildLiveReportRetryPayload(
	originalPayload aiclient.CompletePayload,
	resolution registry.PromptResolution,
	c eval.Case,
	content string,
	validation liveReportProductValidation,
	validationContext review.ReportContentValidationContext,
	reportMessages []review.MessageSnapshot,
) (aiclient.CompletePayload, string, *review.ReportContentDraft, []review.ReportValidationIssue, error) {
	issues := validation.issues
	if resolution.OutputSchema != nil {
		if candidate, candidateIssues, ok := review.DetectReportActionLabelRepairCandidateWithContext(
			*resolution.OutputSchema,
			content,
			c.Language,
			validationContext,
			reportMessages,
		); ok {
			repairPayload, err := review.BuildReportActionLabelRepairPayload(resolution, c.Language, candidate, candidateIssues, aiclient.AITaskRunContext{})
			if err != nil {
				return aiclient.CompletePayload{}, "", nil, nil, err
			}
			return repairPayload, repairScopeActionLabels, &candidate, candidateIssues, nil
		}
	}
	repairPayload := originalPayload
	repairMessages, err := renderCaseRepairMessages(resolution, c, issues)
	if err != nil {
		return aiclient.CompletePayload{}, "", nil, nil, err
	}
	repairPayload.Messages = repairMessages
	return repairPayload, repairScopeWholeReport, nil, nil, nil
}

func retryableLiveProviderError(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var apiErr *sharederrors.APIError
	return errors.As(err, &apiErr) && apiErr.Retryable
}

type liveReportProductValidation struct {
	issues        []review.ReportValidationIssue
	schemaInvalid bool
}

func (validation liveReportProductValidation) valid() bool {
	return len(validation.issues) == 0
}

func (validation liveReportProductValidation) err() error {
	if validation.valid() {
		return nil
	}
	return fmt.Errorf("live report completion invalid: %w", review.ReportValidationIssues(validation.issues))
}

func (validation liveReportProductValidation) retryReason() string {
	if validation.schemaInvalid {
		return retryReasonOutputSchema
	}
	return retryReasonOutputSemantic
}

func validateLiveReportProductOutput(
	response aiclient.CompleteResponse,
	resolution registry.PromptResolution,
	validationContext review.ReportContentValidationContext,
	messages []review.MessageSnapshot,
) liveReportProductValidation {
	issues := make([]review.ReportValidationIssue, 0)
	schemaInvalid := false
	if err := validateLiveReportOutputSchema(response, resolution); err != nil {
		schemaInvalid = true
		issues = append(issues, review.OutputSchemaRepairIssue(err))
	}
	_, semanticIssues := review.ValidateReportContentJSON(response.Content, validationContext, messages)
	issues = append(issues, semanticIssues...)
	return liveReportProductValidation{issues: deduplicateReportValidationIssues(issues), schemaInvalid: schemaInvalid}
}

func deduplicateReportValidationIssues(issues []review.ReportValidationIssue) []review.ReportValidationIssue {
	seen := make(map[review.ReportValidationIssue]struct{}, len(issues))
	result := make([]review.ReportValidationIssue, 0, len(issues))
	for _, issue := range issues {
		if _, duplicate := seen[issue]; duplicate {
			continue
		}
		seen[issue] = struct{}{}
		result = append(result, issue)
	}
	return result
}

func liveReportValidationInputs(c eval.Case) (review.ReportContentValidationContext, []review.MessageSnapshot, error) {
	contextRaw, err := json.Marshal(c.Context)
	if err != nil {
		return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation context cannot be encoded")
	}
	var contextWire struct {
		Language     string `json:"language"`
		HasNextRound *bool  `json:"hasNextRound"`
	}
	if err := json.Unmarshal(contextRaw, &contextWire); err != nil || contextWire.HasNextRound == nil {
		return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation context coordinates are incomplete")
	}
	if contextWire.Language != c.Language || (c.Language != "en" && c.Language != "zh-CN") {
		return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation language coordinate is invalid")
	}
	transcriptRaw, err := json.Marshal(c.Transcript)
	if err != nil {
		return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation transcript cannot be encoded")
	}
	var messages []review.MessageSnapshot
	if err := json.Unmarshal(transcriptRaw, &messages); err != nil || len(messages) == 0 {
		return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation transcript coordinates are incomplete")
	}
	seen := make(map[int]struct{}, len(messages))
	lastSeqNo := 0
	for _, message := range messages {
		if message.SeqNo <= 0 || (message.Role != "user" && message.Role != "assistant") {
			return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation transcript coordinate is invalid")
		}
		if _, duplicate := seen[message.SeqNo]; duplicate {
			return review.ReportContentValidationContext{}, nil, fmt.Errorf("live report validation transcript coordinate is duplicated")
		}
		seen[message.SeqNo] = struct{}{}
		if message.SeqNo > lastSeqNo {
			lastSeqNo = message.SeqNo
		}
	}
	return review.ReportContentValidationContext{
		Language:         c.Language,
		HasNextRound:     *contextWire.HasNextRound,
		LastMessageSeqNo: int32(lastSeqNo),
	}, messages, nil
}

func validateLiveReportOutputSchema(response aiclient.CompleteResponse, resolution registry.PromptResolution) error {
	if resolution.OutputSchema == nil || len(*resolution.OutputSchema) == 0 {
		return fmt.Errorf("live report completion output schema is missing")
	}
	if err := outputschema.Validate(*resolution.OutputSchema, response.Content); err != nil {
		return fmt.Errorf("live report completion output schema invalid: %w", err)
	}
	return nil
}

func renderCaseMessages(res registry.PromptResolution, c eval.Case) ([]aiclient.Message, error) {
	if c.FeatureKey != string(featurekeys.ReportGenerate) {
		return []aiclient.Message{
			{Role: "system", Content: res.SystemMessage},
			{Role: "user", Content: res.UserMessageTemplate + "\n\nInput:\n" + c.Input},
		}, nil
	}
	contextJSON, transcriptJSON, err := marshalReportCaseContext(c)
	if err != nil {
		return nil, err
	}
	return review.BuildReportPromptMessages(res.UserMessageTemplate, c.Language, contextJSON, transcriptJSON)
}

func renderCaseRepairMessages(res registry.PromptResolution, c eval.Case, issues []review.ReportValidationIssue) ([]aiclient.Message, error) {
	contextJSON, transcriptJSON, err := marshalReportCaseContext(c)
	if err != nil {
		return nil, err
	}
	_, messages, err := liveReportValidationInputs(c)
	if err != nil {
		return nil, err
	}
	candidateUserSeqNos := make([]int, 0, len(messages))
	for _, message := range messages {
		if message.Role == "user" {
			candidateUserSeqNos = append(candidateUserSeqNos, message.SeqNo)
		}
	}
	sort.Ints(candidateUserSeqNos)
	return review.BuildReportRepairPromptMessages(res.UserMessageTemplate, c.Language, contextJSON, transcriptJSON, issues, candidateUserSeqNos)
}

func marshalReportCaseContext(c eval.Case) (json.RawMessage, json.RawMessage, error) {
	contextJSON, err := json.Marshal(c.Context)
	if err != nil {
		return nil, nil, err
	}
	transcriptJSON, err := json.Marshal(c.Transcript)
	if err != nil {
		return nil, nil, err
	}
	return contextJSON, transcriptJSON, nil
}

type liveCallAudit struct {
	SchemaVersion       string   `json:"schemaVersion"`
	Stage               string   `json:"stage"`
	CaseID              string   `json:"caseId"`
	Critical            bool     `json:"critical"`
	Pass                bool     `json:"pass"`
	RepairUsed          bool     `json:"repairUsed"`
	RepairScope         string   `json:"repairScope"`
	AttemptCount        int      `json:"attemptCount"`
	RetryCount          int      `json:"retryCount"`
	RetryReasons        []string `json:"retryReasons"`
	RepairScopes        []string `json:"repairScopes"`
	ErrorClass          string   `json:"errorClass,omitempty"`
	FeatureKey          string   `json:"featureKey"`
	PromptVersion       string   `json:"promptVersion"`
	RubricVersion       string   `json:"rubricVersion"`
	Language            string   `json:"language"`
	FeatureFlag         string   `json:"featureFlag,omitempty"`
	DataSourceVersion   string   `json:"dataSourceVersion,omitempty"`
	Provider            string   `json:"provider,omitempty"`
	ModelID             string   `json:"modelId,omitempty"`
	ModelProfileName    string   `json:"modelProfileName,omitempty"`
	ModelProfileVersion string   `json:"modelProfileVersion,omitempty"`
	FinishReason        string   `json:"finishReason,omitempty"`
	InputTokens         int      `json:"inputTokens"`
	OutputTokens        int      `json:"outputTokens"`
	LatencyMs           int64    `json:"latencyMs"`
	ValidationStatus    string   `json:"validationStatus,omitempty"`
	OutputSHA256        string   `json:"outputSha256"`
	OutputBytes         int      `json:"outputBytes"`
}

func buildCompletionAudit(c eval.Case, result liveCompletionResult, callErr error) liveCallAudit {
	return newLiveCallAudit(
		"completion", c,
		result.resolution.FeatureKey, result.resolution.PromptVersion, result.resolution.RubricVersion, c.Language,
		result.resolution.FeatureFlag, result.resolution.DataSourceVersion,
		result.response, result.meta, result.repairUsed, result.repairScope,
		result.attemptCount, result.retryReasons, result.repairScopes, callErr,
	)
}

func buildJudgeAudit(c eval.Case, capture *capturingJudgeModel, callErr error) liveCallAudit {
	response, meta := capture.response, capture.meta
	promptVersion := meta.PromptVersion
	if promptVersion == "" {
		promptVersion = c.PromptVersion
	}
	rubricVersion := meta.RubricVersion
	if rubricVersion == "" {
		rubricVersion = c.RubricVersion
	}
	language := meta.Language
	if language == "" {
		language = "multi"
	}
	return newLiveCallAudit(
		"judge", c, c.FeatureKey, promptVersion, rubricVersion, language, meta.FeatureFlag, meta.DataSourceVersion,
		response, meta, false, repairScopeNone,
		capture.attemptCount, capture.retryReasons, capture.repairScopes, callErr,
	)
}

func newLiveCallAudit(
	stage string,
	c eval.Case,
	featureKey, promptVersion, rubricVersion, language, featureFlag, dataSourceVersion string,
	response aiclient.CompleteResponse,
	meta aiclient.AICallMeta,
	repairUsed bool,
	repairScope string,
	attemptCount int,
	retryReasons []string,
	repairScopes []string,
	callErr error,
) liveCallAudit {
	if repairScope == "" {
		repairScope = repairScopeNone
	}
	digest := sha256.Sum256([]byte(response.Content))
	audit := liveCallAudit{
		SchemaVersion: "evalkit-live-call-audit.v2", Stage: stage, CaseID: c.ID, Critical: c.Critical, Pass: callErr == nil, RepairUsed: repairUsed, RepairScope: repairScope,
		AttemptCount: attemptCount, RetryCount: len(retryReasons), RetryReasons: nonNilStrings(retryReasons), RepairScopes: nonNilStrings(repairScopes),
		FeatureKey: featureKey, PromptVersion: promptVersion, RubricVersion: rubricVersion, Language: language,
		FeatureFlag: featureFlag, DataSourceVersion: dataSourceVersion,
		Provider: meta.Provider, ModelID: meta.ModelID, ModelProfileName: meta.ModelProfileName, ModelProfileVersion: meta.ModelProfileVersion,
		FinishReason: response.FinishReason, InputTokens: meta.InputTokens, OutputTokens: meta.OutputTokens, LatencyMs: meta.LatencyMs,
		ValidationStatus: string(meta.ValidationStatus), OutputSHA256: fmt.Sprintf("%x", digest), OutputBytes: len(response.Content),
	}
	if callErr != nil {
		audit.ErrorClass = liveCallErrorClass(callErr)
	}
	return audit
}

func nonNilStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func liveCallErrorClass(err error) string {
	if errors.Is(err, registry.ErrJudgeContentRejected) {
		return "judge_content_rejected"
	}
	if errors.Is(err, registry.ErrJudgeProtocolInvalid) {
		return "judge_protocol_invalid"
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Retryable {
			return "provider_retryable_exhausted"
		}
		return "provider_nonretryable"
	}
	return "validation_failed"
}

func writeLiveAudit(path string, audit liveCallAudit) error {
	data, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal live audit: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create live audit directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write live audit: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("protect live audit: %w", err)
	}
	return nil
}

type gradeScore struct {
	Dimension string  `json:"dimension"`
	Value     float64 `json:"value"`
}

type gradeItemVerdict struct {
	Path                    string `json:"path"`
	Kind                    string `json:"kind"`
	Support                 string `json:"support"`
	EvidenceLimitedExplicit bool   `json:"evidence_limited_explicit"`
	UsedForNegativeClaim    bool   `json:"used_for_negative_claim"`
	Reason                  string `json:"reason"`
}

type gradeCausalCheck struct {
	DimensionCode   string `json:"dimension_code"`
	IssueSupported  bool   `json:"issue_supported"`
	FocusSupported  bool   `json:"focus_supported"`
	ActionSupported bool   `json:"action_supported"`
	Reason          string `json:"reason"`
}

func buildGradeVerdict(c eval.Case, scores []registry.Score, reasoning registry.Reasoning, weightedScore *float64, gradeErr error) map[string]any {
	verdict := map[string]any{"pass": gradeErr == nil, "caseId": c.ID, "critical": c.Critical}
	if gradeErr != nil {
		verdict["reason"] = gradeErr.Error()
		return verdict
	}
	gradeScores := make([]gradeScore, 0, len(scores))
	for _, score := range scores {
		gradeScores = append(gradeScores, gradeScore{Dimension: score.Dimension, Value: score.Value})
	}
	items := make([]gradeItemVerdict, 0, len(reasoning.ItemVerdicts))
	for _, item := range reasoning.ItemVerdicts {
		items = append(items, gradeItemVerdict{
			Path: item.Path, Kind: item.Kind, Support: item.Support, EvidenceLimitedExplicit: item.EvidenceLimitedExplicit,
			UsedForNegativeClaim: item.UsedForNegativeClaim, Reason: item.Reason,
		})
	}
	causal := make([]gradeCausalCheck, 0, len(reasoning.CausalChecks))
	for _, check := range reasoning.CausalChecks {
		causal = append(causal, gradeCausalCheck{
			DimensionCode: check.DimensionCode, IssueSupported: check.IssueSupported, FocusSupported: check.FocusSupported,
			ActionSupported: check.ActionSupported, Reason: check.Reason,
		})
	}
	verdict["scores"] = gradeScores
	if weightedScore != nil {
		verdict["weighted_score"] = *weightedScore
	}
	verdict["reason"] = reasoning.Summary
	verdict["item_verdicts"] = items
	verdict["causal_checks"] = causal
	violations := reasoning.ZeroToleranceViolations
	if violations == nil {
		violations = []string{}
	}
	verdict["zero_tolerance_violations"] = violations
	verdict["critical_safety_pass"] = reasoning.CriticalSafetyPass
	return verdict
}

type capturingJudgeModel struct {
	delegate           registry.JudgeModelClient
	response           aiclient.CompleteResponse
	meta               aiclient.AICallMeta
	attemptCount       int
	retryReasons       []string
	repairScopes       []string
	pendingRetryReason string
}

func (m *capturingJudgeModel) CompleteJudge(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	if m.attemptCount > 0 {
		reason := m.pendingRetryReason
		if reason == "" {
			reason = "judge_protocol_invalid"
		}
		m.retryReasons = append(m.retryReasons, reason)
		m.repairScopes = append(m.repairScopes, repairScopeNone)
	}
	response, meta, err := m.delegate.CompleteJudge(ctx, profileName, payload)
	m.attemptCount++
	m.response = response
	if m.attemptCount == 1 {
		m.meta = meta
	} else {
		m.meta = review.AggregateReportRepairMeta(m.meta, meta)
	}
	if retryableLiveProviderError(ctx, err) {
		m.pendingRetryReason = retryReasonProviderRetryable
	} else if err == nil {
		m.pendingRetryReason = "judge_protocol_invalid"
	} else {
		m.pendingRetryReason = ""
	}
	return response, meta, err
}

func validateLiveReportCall(response aiclient.CompleteResponse, meta aiclient.AICallMeta, resolution registry.PromptResolution, language string) error {
	if strings.TrimSpace(response.FinishReason) != "stop" {
		return fmt.Errorf("live report completion finish reason must be stop")
	}
	return validateLiveReportCallMeta(meta, resolution, language, aiclient.ValidationStatusOK, "")
}

func validateLiveReportInvalidCall(response aiclient.CompleteResponse, meta aiclient.AICallMeta, resolution registry.PromptResolution, language string) error {
	if strings.TrimSpace(response.FinishReason) != "stop" {
		return fmt.Errorf("live report completion invalid finish reason must be stop")
	}
	return validateLiveReportCallMeta(meta, resolution, language, aiclient.ValidationStatusInvalid, sharederrors.CodeAiOutputInvalid)
}

func validateLiveReportCallMeta(meta aiclient.AICallMeta, resolution registry.PromptResolution, language string, validationStatus aiclient.ValidationStatus, errorCode string) error {
	if meta.InputTokens <= 0 || meta.OutputTokens <= 0 {
		return fmt.Errorf("live report completion token usage is missing")
	}
	if meta.ValidationStatus != validationStatus || meta.ErrorCode != errorCode {
		return fmt.Errorf("live report completion validation provenance is invalid")
	}
	if meta.FeatureKey != resolution.FeatureKey || meta.PromptVersion != resolution.PromptVersion || meta.RubricVersion != resolution.RubricVersion ||
		meta.ModelProfileName != resolution.ModelProfileName || meta.Language != language || meta.FeatureFlag != resolution.FeatureFlag || meta.DataSourceVersion != resolution.DataSourceVersion {
		return fmt.Errorf("live report completion provenance mismatch")
	}
	return nil
}

func isLiveReportOutputInvalid(err error) bool {
	var apiErr *sharederrors.APIError
	return errors.As(err, &apiErr) && apiErr.Code == sharederrors.CodeAiOutputInvalid && !apiErr.Retryable
}

func validateLiveJudgeCall(response aiclient.CompleteResponse, meta aiclient.AICallMeta, resolution registry.PromptResolution) error {
	if strings.TrimSpace(response.FinishReason) != "stop" {
		return fmt.Errorf("live judge finish reason must be stop")
	}
	if meta.InputTokens <= 0 || meta.OutputTokens <= 0 {
		return fmt.Errorf("live judge token usage is missing")
	}
	if meta.ValidationStatus != aiclient.ValidationStatusOK || meta.ErrorCode != "" {
		return fmt.Errorf("live judge validation provenance is invalid")
	}
	if meta.FeatureKey != resolution.FeatureKey || meta.PromptVersion != resolution.PromptVersion || meta.RubricVersion != resolution.RubricVersion || meta.ModelProfileName != "judge.default" || meta.Language != "multi" {
		return fmt.Errorf("live judge provenance mismatch")
	}
	return nil
}

func runLive(ctx context.Context, suite *eval.Suite, reg *registry.Client, live liveOpts) error {
	rt, err := liveRuntime(live)
	if err != nil {
		return err
	}
	defer rt.Close()
	for _, c := range suite.Cases {
		completion, err := liveComplete(ctx, reg, c, live)
		if err != nil {
			return fmt.Errorf("case %q live complete: %w", c.ID, err)
		}
		if _, _, err := suite.GradeOutput(ctx, reg, rt.Client, c, []byte(completion.response.Content)); err != nil {
			return fmt.Errorf("case %q live grade: %w", c.ID, err)
		}
	}
	fmt.Printf("evalkit run: OK (live, %d cases)\n", suite.Count())
	return nil
}
