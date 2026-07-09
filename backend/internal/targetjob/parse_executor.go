package targetjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

// FeatureKeyTargetImportParse is the F3 feature_key consumed by the parse
// pipeline. Spec §3.1 D-4 fixes this string; F3 is the owner of the
// underlying prompt / rubric / model_profile binding.
const FeatureKeyTargetImportParse = "target.import.parse"

// PromptResolution is the F3 RegistryClient response shape that the parse
// executor consumes. Spec D-4 fixes the three required fields plus the
// language echo and the data_source_version slot.
type PromptResolution struct {
	PromptVersion       string
	RubricVersion       string
	ModelProfileName    string
	DataSourceVersion   string
	FeatureFlag         string
	SystemMessage       string
	UserMessageTemplate string
	OutputSchema        *json.RawMessage
}

// PromptRegistryClient is the F3 boundary. The targetjob domain references
// it through this interface so F3 can ship its own client implementation
// without us depending on its package directly.
type PromptRegistryClient interface {
	Resolve(ctx context.Context, featureKey string, language string) (PromptResolution, error)
}

// ErrPromptUnsupported is returned by F3 implementations when the requested
// (featureKey, language) tuple is not enabled for the active profile.
var ErrPromptUnsupported = errors.New("prompt registry: feature/language is not enabled")

// URLFetcher is the urlfetch boundary used by the parse executor. The
// production implementation is *urlfetch.Fetcher.
type URLFetcher interface {
	Fetch(ctx context.Context, rawURL string) (urlfetch.FetchResult, error)
}

// ParseExecutorOptions wires the parse executor.
type ParseExecutorOptions struct {
	Store    Store
	Registry PromptRegistryClient
	AI       aiclient.AIClient
	Fetcher  URLFetcher
	NewID    IDGenerator
	Now      func() time.Time
}

// ParseExecutor is the JobHandler for the TargetJob import async job type.
type ParseExecutor struct {
	store    Store
	registry PromptRegistryClient
	ai       aiclient.AIClient
	fetcher  URLFetcher
	newID    IDGenerator
	now      func() time.Time
}

// NewParseExecutor constructs a ParseExecutor.
func NewParseExecutor(opts ParseExecutorOptions) *ParseExecutor {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	return &ParseExecutor{
		store:    opts.Store,
		registry: opts.Registry,
		ai:       opts.AI,
		fetcher:  opts.Fetcher,
		newID:    opts.NewID,
		now:      opts.Now,
	}
}

// parseAIResponse is the structured shape the executor expects from A3. Any
// drift from this shape surfaces as AI_OUTPUT_INVALID (non-retryable). The
// upstream prompt is owned by F3 so the schema lives here as the consumer.
type parseAIResponse struct {
	Title           string                 `json:"title"`
	CompanyName     string                 `json:"companyName"`
	CoreThemes      []string               `json:"coreThemes"`
	InterviewRounds []parseAIResponseRound `json:"interviewRounds"`
	Strengths       []string               `json:"strengths"`
	Gaps            []string               `json:"gaps"`
	RiskSignals     []string               `json:"riskSignals"`
	Requirements    []parseAIResponseReq   `json:"requirements"`
}

type parseAIResponseRound struct {
	Sequence        int    `json:"sequence"`
	Type            string `json:"type"`
	Name            string `json:"name"`
	DurationMinutes int    `json:"durationMinutes"`
	Focus           string `json:"focus"`
}

type parseAIResponseReq struct {
	Kind          string `json:"kind"`
	Label         string `json:"label"`
	Description   string `json:"description,omitempty"`
	EvidenceLevel string `json:"evidenceLevel,omitempty"`
}

// Handle satisfies JobHandler. It returns success or the appropriate
// retryable / non-retryable failure outcome and writes the matching
// parsed / analysis-failed outbox event before returning.
func (p *ParseExecutor) Handle(ctx context.Context, job ClaimedJob) JobOutcome {
	if p == nil || p.store == nil {
		return JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: "parse executor not initialised"}
	}
	targetJobID := job.ResourceID
	target, sources, err := p.store.GetTargetJobForParse(ctx, targetJobID)
	if err != nil {
		if errors.Is(err, ErrTargetJobNotFound) {
			return JobOutcome{
				ErrorCode:    sharederrors.CodeTargetJobNotFound,
				ErrorMessage: safeFailureMessage(sharederrors.CodeTargetJobNotFound, err.Error()),
				Retryable:    false,
			}
		}
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), false)
	}

	var fetchedURLBody string
	var sourceURLForPrompt string
	if target.SourceType == SourceTypeURL {
		fetched, err := p.fetchURLSnapshot(ctx, target, sources)
		if err != nil {
			return p.translateAndFail(ctx, targetJobID, err)
		}
		fetchedURLBody = fetched.Body
		sourceURLForPrompt = fetched.SanitizedURL
	}

	resolution, err := p.registry.Resolve(ctx, FeatureKeyTargetImportParse, target.TargetLanguage)
	if err != nil {
		// F3 disabled / unsupported / config-invalid all map to non-retryable
		// AI_PROVIDER_CONFIG_INVALID per spec D-10 / plan 1.2.
		return p.fail(ctx, targetJobID, sharederrors.CodeAiProviderConfigInvalid, err.Error(), false)
	}

	jdText := fetchedURLBody
	if strings.TrimSpace(jdText) == "" {
		jdText = target.RawJDText
		for _, src := range sources {
			if src.SnapshotText != "" {
				jdText = src.SnapshotText
				break
			}
		}
	}
	if strings.TrimSpace(jdText) == "" {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportSourceInvalid, "no JD text available for parse", false)
	}

	metadata := aiclient.CallMetadata{
		FeatureKey:        FeatureKeyTargetImportParse,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          target.TargetLanguage,
		FeatureFlag:       coalesceFlag(resolution.FeatureFlag),
		DataSourceVersion: resolution.DataSourceVersion,
		TaskRun: aiclient.AITaskRunContext{
			Capability:   aiclient.AITaskRunTaskJDParse,
			ResourceType: aiclient.AITaskRunResourceTargetJob,
			ResourceID:   targetJobID,
		},
	}
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	complete, aiMeta, err := p.ai.Complete(ctx, resolution.ModelProfileName, aiclient.CompletePayload{
		Messages: buildPromptMessages(resolution, target.TargetLanguage, jdText, sourceURLForPrompt),
		Metadata: metadata,
	})
	if err != nil {
		code, retryable := translateAIClientError(err)
		return p.fail(ctx, targetJobID, code, err.Error(), retryable)
	}
	modelID := strings.TrimSpace(aiMeta.ModelID)
	if modelID == "" {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiProviderConfigInvalid, "AI call meta missing model_id", false)
	}

	parsed, err := decodeParseResponse(complete.Content)
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}
	parsed.Title = strings.TrimSpace(parsed.Title)
	parsed.CompanyName = coalesceCompanyName(parsed.CompanyName, target.TargetLanguage)
	if parsed.Title == "" {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiOutputInvalid, "AI output missing title", false)
	}
	parsed.Requirements, err = ensureHiddenSignalRequirements(parsed.Requirements, parsed.RiskSignals)
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}

	requirements, err := buildRequirements(parsed.Requirements, p.newID)
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}
	summary := mustMarshal(map[string]any{
		"coreThemes":      parsed.CoreThemes,
		"interviewRounds": parsed.InterviewRounds,
		"provenance": map[string]string{
			"language":          target.TargetLanguage,
			"featureFlag":       coalesceFlag(resolution.FeatureFlag),
			"promptVersion":     resolution.PromptVersion,
			"rubricVersion":     resolution.RubricVersion,
			"modelId":           modelID,
			"dataSourceVersion": resolution.DataSourceVersion,
		},
	})
	fitSummary := mustMarshal(map[string]any{
		"strengths":   parsed.Strengths,
		"gaps":        parsed.Gaps,
		"riskSignals": parsed.RiskSignals,
		"provenance": map[string]string{
			"language":          target.TargetLanguage,
			"featureFlag":       coalesceFlag(resolution.FeatureFlag),
			"promptVersion":     resolution.PromptVersion,
			"rubricVersion":     "not_applicable",
			"modelId":           modelID,
			"dataSourceVersion": resolution.DataSourceVersion,
		},
	})

	now := p.now()
	parsedPayload, err := BuildTargetParsedPayload(TargetParsedInput{
		TargetJobID:      targetJobID,
		UserID:           target.UserID,
		AnalysisStatus:   sharedtypes.TargetJobParseStatusReady,
		RequirementCount: len(requirements),
		CoreThemes:       parsed.CoreThemes,
	})
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), false)
	}
	rawParsed, _ := json.Marshal(parsedPayload)
	if err := p.store.CompleteParseSuccess(ctx, CompleteParseSuccessInput{
		TargetJobID:        targetJobID,
		Title:              parsed.Title,
		CompanyName:        parsed.CompanyName,
		AnalysisStatus:     sharedtypes.TargetJobParseStatusReady,
		Summary:            summary,
		FitSummary:         fitSummary,
		Requirements:       requirements,
		ParsedEventID:      p.newID(),
		ParsedEventPayload: rawParsed,
		SourceRefreshJobID: p.newID(),
		Now:                now,
	}); err != nil {
		return JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}

	return JobOutcome{Succeeded: true}
}

func (p *ParseExecutor) fetchURLSnapshot(ctx context.Context, target TargetJobRecord, sources []SourceRecord) (urlfetch.FetchResult, error) {
	if p.fetcher == nil {
		return urlfetch.FetchResult{}, fmt.Errorf("%w: url fetcher not configured", urlfetch.ErrSourceUnavailable)
	}
	if target.SourceURL == "" {
		return urlfetch.FetchResult{}, fmt.Errorf("%w: no source url recorded", urlfetch.ErrInvalidSource)
	}
	res, err := p.fetcher.Fetch(ctx, target.SourceURL)
	if err != nil {
		return urlfetch.FetchResult{}, err
	}
	// Find the most recent url source row to update.
	var sourceID string
	for _, s := range sources {
		if s.SourceType == SourceTypeURL {
			sourceID = s.ID
			break
		}
	}
	if sourceID == "" {
		// Nothing to update — proceed; the fetch result is in memory but
		// will not be persisted in target_job_sources. Phase 1 ImportTargetJob
		// always inserts a source row for url, so this is a defensive no-op.
		return res, nil
	}
	if err := p.store.UpdateSourceSnapshot(ctx, sourceID, res.SanitizedURL, res.Body, res.FetchedAt, p.now()); err != nil {
		return urlfetch.FetchResult{}, err
	}
	return res, nil
}

func (p *ParseExecutor) fail(ctx context.Context, targetJobID, code, message string, retryable bool) JobOutcome {
	now := p.now()
	payload, err := BuildTargetAnalysisFailedPayload(TargetAnalysisFailedInput{
		TargetJobID: targetJobID,
		ErrorCode:   code,
		Retryable:   retryable,
	})
	if err != nil {
		return JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	raw, _ := json.Marshal(payload)
	if err := p.store.CompleteParseFailure(ctx, CompleteParseFailureInput{
		TargetJobID:        targetJobID,
		FailedEventID:      p.newID(),
		FailedEventPayload: raw,
		Now:                now,
	}); err != nil {
		return JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	return JobOutcome{
		ErrorCode:    code,
		ErrorMessage: safeFailureMessage(code, message),
		Retryable:    retryable,
	}
}

func (p *ParseExecutor) translateAndFail(ctx context.Context, targetJobID string, err error) JobOutcome {
	switch {
	case errors.Is(err, urlfetch.ErrInvalidSource):
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportSourceInvalid, err.Error(), false)
	case errors.Is(err, urlfetch.ErrSourceUnavailable):
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportSourceUnavailable, err.Error(), true)
	default:
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), true)
	}
}

func translateAIClientError(err error) (string, bool) {
	msg := err.Error()
	for _, code := range []string{
		sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiFallbackExhausted,
	} {
		if strings.Contains(msg, code) {
			return code, true
		}
	}
	for _, code := range []string{
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiUnsupportedCapability,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiProviderConfigInvalid,
	} {
		if strings.Contains(msg, code) {
			return code, false
		}
	}
	// Default: treat as retryable transient AI failure.
	return sharederrors.CodeAiFallbackExhausted, true
}

// decodeParseResponse parses the AI completion content as the structured
// JSON shape the F3 prompt is configured to emit. Schema drift surfaces
// as AI_OUTPUT_INVALID upstream.
func decodeParseResponse(content string) (parseAIResponse, error) {
	var out parseAIResponse
	content = outputschema.NormalizeJSONContent(content)
	if content == "" {
		return parseAIResponse{}, fmt.Errorf("AI response content was empty")
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return parseAIResponse{}, fmt.Errorf("AI response is not valid JSON: %v", err)
	}
	if len(out.Requirements) == 0 {
		return parseAIResponse{}, fmt.Errorf("AI response has no requirements")
	}
	rounds, err := validateInterviewRounds(out.InterviewRounds)
	if err != nil {
		return parseAIResponse{}, err
	}
	out.InterviewRounds = rounds
	return out, nil
}

func buildPromptMessages(resolution PromptResolution, language string, jdText string, sourceURL string) []aiclient.Message {
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	user := strings.TrimSpace(jdText)
	if template := strings.TrimSpace(resolution.UserMessageTemplate); template != "" {
		user = strings.ReplaceAll(template, "{{jd_text}}", jdText)
		user = strings.ReplaceAll(user, "{{language}}", language)
		user = strings.ReplaceAll(user, "{{jd_source_url}}", sourceURL)
		user = strings.TrimSpace(user)
	}
	if user != "" {
		messages = append(messages, aiclient.Message{Role: "user", Content: user})
	}
	return messages
}

func buildRequirements(input []parseAIResponseReq, newID IDGenerator) ([]RequirementRecord, error) {
	out := make([]RequirementRecord, 0, len(input))
	for i, r := range input {
		kind := RequirementKind(strings.TrimSpace(r.Kind))
		label := strings.TrimSpace(r.Label)
		if !validRequirementKind(kind) {
			return nil, fmt.Errorf("AI response requirement %d has invalid kind", i)
		}
		if label == "" {
			return nil, fmt.Errorf("AI response requirement %d has empty label", i)
		}
		evidence := EvidenceLevel(strings.TrimSpace(r.EvidenceLevel))
		if evidence == "" {
			evidence = EvidenceExplicit
		}
		if !validEvidenceLevel(evidence) {
			return nil, fmt.Errorf("AI response requirement %d has invalid evidence level", i)
		}
		out = append(out, RequirementRecord{
			ID:            newID(),
			Kind:          kind,
			Label:         label,
			Description:   strings.TrimSpace(r.Description),
			EvidenceLevel: evidence,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("AI response has no valid requirements")
	}
	return out, nil
}

func ensureHiddenSignalRequirements(input []parseAIResponseReq, riskSignals []string) ([]parseAIResponseReq, error) {
	for _, r := range input {
		if strings.TrimSpace(r.Kind) == string(RequirementHiddenSignal) && strings.TrimSpace(r.Label) != "" {
			return input, nil
		}
	}

	out := append([]parseAIResponseReq{}, input...)
	for _, signal := range riskSignals {
		label := strings.TrimSpace(signal)
		if label == "" {
			continue
		}
		out = append(out, parseAIResponseReq{
			Kind:          string(RequirementHiddenSignal),
			Label:         label,
			EvidenceLevel: string(EvidenceInferred),
		})
	}
	if len(out) == len(input) {
		return nil, fmt.Errorf("AI response must include at least one hidden_signal requirement or riskSignals item")
	}
	return out, nil
}

func validateInterviewRounds(input []parseAIResponseRound) ([]parseAIResponseRound, error) {
	if len(input) < 2 || len(input) > 5 {
		return nil, fmt.Errorf("AI response must include 2 to 5 interview rounds")
	}
	out := make([]parseAIResponseRound, 0, len(input))
	seenSequence := map[int]struct{}{}
	for i, round := range input {
		round.Type = strings.TrimSpace(round.Type)
		round.Name = strings.TrimSpace(round.Name)
		round.Focus = strings.TrimSpace(round.Focus)
		if round.Sequence < 1 {
			return nil, fmt.Errorf("AI response interview round %d has invalid sequence", i)
		}
		if _, ok := seenSequence[round.Sequence]; ok {
			return nil, fmt.Errorf("AI response interview round %d has duplicate sequence", i)
		}
		seenSequence[round.Sequence] = struct{}{}
		if !validInterviewRoundType(round.Type) {
			return nil, fmt.Errorf("AI response interview round %d has invalid type", i)
		}
		if round.Name == "" {
			return nil, fmt.Errorf("AI response interview round %d has empty name", i)
		}
		if round.DurationMinutes < 10 || round.DurationMinutes > 180 {
			return nil, fmt.Errorf("AI response interview round %d has invalid durationMinutes", i)
		}
		if round.Focus == "" {
			return nil, fmt.Errorf("AI response interview round %d has empty focus", i)
		}
		out = append(out, round)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Sequence < out[j].Sequence
	})
	return out, nil
}

func validInterviewRoundType(value string) bool {
	switch value {
	case "hr", "technical", "manager", "cross_functional", "culture", "final", "other":
		return true
	default:
		return false
	}
}

func validRequirementKind(k RequirementKind) bool {
	switch k {
	case RequirementMustHave, RequirementNiceToHave, RequirementHiddenSignal, RequirementInterviewFocus:
		return true
	default:
		return false
	}
}

func validEvidenceLevel(e EvidenceLevel) bool {
	switch e {
	case EvidenceExplicit, EvidenceInferred:
		return true
	default:
		return false
	}
}

func mustMarshal(v any) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func coalesceFlag(v string) string {
	if v == "" {
		return "none"
	}
	return v
}

func coalesceCompanyName(companyName string, language string) string {
	if companyName = strings.TrimSpace(companyName); companyName != "" {
		return companyName
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(language)), "zh") {
		return "未提供"
	}
	return "Unknown company"
}

func safeFailureMessage(code, msg string) string {
	switch code {
	case sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiFallbackExhausted,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiUnsupportedCapability,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeTargetImportSourceInvalid,
		sharederrors.CodeTargetImportSourceUnavailable:
		return code
	default:
		return redactErrorMessage(msg)
	}
}

func redactErrorMessage(msg string) string {
	// Defensive redaction: never let raw JD or prompt body leak into the
	// async_jobs.error_message column. Truncate at 240 chars and strip
	// obvious sensitive substrings.
	if len(msg) > 240 {
		msg = msg[:240]
	}
	lower := strings.ToLower(msg)
	for _, kw := range []string{
		"raw_jd_text",
		"authorization:",
		"bearer ",
		"provider secret",
		"prompt body",
		"response body",
		"private jd body",
	} {
		if strings.Contains(lower, kw) {
			return "redacted error message containing forbidden token"
		}
	}
	return msg
}

// SourceRefreshHandler is the placeholder JobHandler for source_refresh
// rows enqueued in 4.5. The actual refresh is deferred to a future plan;
// today we just flip target_job_sources.freshness_status to 'stale' so
// observability has a marker.
type SourceRefreshHandler struct {
	Store Store
	Now   func() time.Time
}

// Handle satisfies JobHandler.
func (h *SourceRefreshHandler) Handle(ctx context.Context, job ClaimedJob) JobOutcome {
	now := time.Now().UTC()
	if h.Now != nil {
		now = h.Now()
	}
	if err := h.Store.UpdateSourceFreshness(ctx, job.ResourceID, FreshnessStale, now); err != nil {
		return JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: err.Error(), Retryable: true}
	}
	return JobOutcome{Succeeded: true}
}
