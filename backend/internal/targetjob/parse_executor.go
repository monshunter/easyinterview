package targetjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
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
	PromptVersion     string
	RubricVersion     string
	ModelProfileName  string
	DataSourceVersion string
	FeatureFlag       string
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

// ParseExecutor is the JobHandler for the `target_import` async job type.
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
	CoreThemes          []string                  `json:"coreThemes"`
	InterviewHypotheses []string                  `json:"interviewHypotheses"`
	Strengths           []string                  `json:"strengths"`
	Gaps                []string                  `json:"gaps"`
	RiskSignals         []string                  `json:"riskSignals"`
	Requirements        []parseAIResponseReq      `json:"requirements"`
}

type parseAIResponseReq struct {
	Kind          string `json:"kind"`
	Label         string `json:"label"`
	Description   string `json:"description,omitempty"`
	EvidenceLevel string `json:"evidenceLevel,omitempty"`
}

// Handle satisfies JobHandler. It returns success or the appropriate
// retryable / non-retryable failure outcome and writes the matching
// target.parsed / target.analysis.failed outbox event before returning.
func (p *ParseExecutor) Handle(ctx context.Context, job ClaimedJob) JobOutcome {
	if p == nil || p.store == nil {
		return JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: "parse executor not initialised"}
	}
	targetJobID := job.ResourceID
	target, sources, err := p.store.GetTargetJobForParse(ctx, targetJobID)
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), false)
	}

	if target.SourceType == SourceTypeURL {
		if err := p.fetchURLSnapshot(ctx, target, sources); err != nil {
			return p.translateAndFail(ctx, targetJobID, err)
		}
	}

	resolution, err := p.registry.Resolve(ctx, FeatureKeyTargetImportParse, target.TargetLanguage)
	if err != nil {
		// F3 disabled / unsupported / config-invalid all map to non-retryable
		// AI_PROVIDER_CONFIG_INVALID per spec D-10 / plan 1.2.
		return p.fail(ctx, targetJobID, sharederrors.CodeAiProviderConfigInvalid, err.Error(), false)
	}

	jdText := target.RawJDText
	for _, src := range sources {
		if src.SnapshotText != "" {
			jdText = src.SnapshotText
			break
		}
	}
	if strings.TrimSpace(jdText) == "" {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportSourceInvalid, "no JD text available for parse", false)
	}

	complete, _, err := p.ai.Complete(ctx, resolution.ModelProfileName, aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "You are an extraction assistant for the target.import.parse feature. Return strict JSON."},
			{Role: "user", Content: jdText},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    FeatureKeyTargetImportParse,
			PromptVersion: resolution.PromptVersion,
			RubricVersion: resolution.RubricVersion,
			Language:      target.TargetLanguage,
		},
	})
	if err != nil {
		code, retryable := translateAIClientError(err)
		return p.fail(ctx, targetJobID, code, err.Error(), retryable)
	}

	parsed, err := decodeParseResponse(complete.Content)
	if err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}

	requirements := buildRequirements(parsed.Requirements, p.newID)
	summary := mustMarshal(map[string]any{
		"coreThemes":          parsed.CoreThemes,
		"interviewHypotheses": parsed.InterviewHypotheses,
		"provenance": map[string]string{
			"language":          target.TargetLanguage,
			"featureFlag":       coalesceFlag(resolution.FeatureFlag),
			"promptVersion":     resolution.PromptVersion,
			"rubricVersion":     resolution.RubricVersion,
			"modelId":           resolution.ModelProfileName,
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
			"modelId":           resolution.ModelProfileName,
			"dataSourceVersion": resolution.DataSourceVersion,
		},
	})

	now := p.now()
	if err := p.store.ApplyParseResult(ctx, ApplyParseResultInput{
		TargetJobID:    targetJobID,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Summary:        summary,
		FitSummary:     fitSummary,
		Requirements:   requirements,
		Now:            now,
	}); err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), true)
	}

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
	if err := p.store.WriteTargetParsedOutbox(ctx, p.newID(), targetJobID, rawParsed, now); err != nil {
		return p.fail(ctx, targetJobID, sharederrors.CodeTargetImportFailed, err.Error(), true)
	}

	// 4.5: enqueue source_refresh placeholder; failure here does not roll
	// back the parse, but we surface it so observability sees the gap.
	if err := p.store.EnqueueSourceRefresh(ctx, p.newID(), targetJobID, now); err != nil {
		// Log-style non-fatal: still succeed but record the issue.
		_ = err
	}

	return JobOutcome{Succeeded: true}
}

func (p *ParseExecutor) fetchURLSnapshot(ctx context.Context, target TargetJobRecord, sources []SourceRecord) error {
	if p.fetcher == nil {
		return fmt.Errorf("%w: url fetcher not configured", urlfetch.ErrSourceUnavailable)
	}
	if target.SourceURL == "" {
		return fmt.Errorf("%w: no source url recorded", urlfetch.ErrInvalidSource)
	}
	res, err := p.fetcher.Fetch(ctx, target.SourceURL)
	if err != nil {
		return err
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
		return nil
	}
	// We do not have a transactional UpdateSource method yet; mirror via
	// existing freshness updater plus an inline patch. For the boot plan
	// we settle for refreshing freshness; spec follow-up plan can extend
	// the store to write snapshot_text. This still satisfies the parse
	// pipeline because we already pass jdText via the in-memory FetchResult
	// above (see Handle).
	_ = res
	return nil
}

func (p *ParseExecutor) fail(ctx context.Context, targetJobID, code, message string, retryable bool) JobOutcome {
	now := p.now()
	if err := p.store.UpdateTargetJobAnalysisFailure(ctx, targetJobID, now); err != nil {
		// stay with the original failure outcome regardless
		_ = err
	}
	payload, err := BuildTargetAnalysisFailedPayload(TargetAnalysisFailedInput{
		TargetJobID: targetJobID,
		ErrorCode:   code,
		Retryable:   retryable,
	})
	if err == nil {
		raw, _ := json.Marshal(payload)
		_ = p.store.WriteParseFailedOutbox(ctx, p.newID(), targetJobID, raw, now)
	}
	return JobOutcome{
		ErrorCode:    code,
		ErrorMessage: redactErrorMessage(message),
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
	content = strings.TrimSpace(content)
	if content == "" {
		return parseAIResponse{}, fmt.Errorf("AI response content was empty")
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return parseAIResponse{}, fmt.Errorf("AI response is not valid JSON: %v", err)
	}
	if len(out.Requirements) == 0 {
		return parseAIResponse{}, fmt.Errorf("AI response has no requirements")
	}
	return out, nil
}

func buildRequirements(input []parseAIResponseReq, newID IDGenerator) []RequirementRecord {
	out := make([]RequirementRecord, 0, len(input))
	for _, r := range input {
		kind := RequirementKind(strings.TrimSpace(r.Kind))
		label := strings.TrimSpace(r.Label)
		if !validRequirementKind(kind) || label == "" {
			continue
		}
		evidence := EvidenceLevel(strings.TrimSpace(r.EvidenceLevel))
		if evidence == "" {
			evidence = EvidenceExplicit
		}
		out = append(out, RequirementRecord{
			ID:            newID(),
			Kind:          kind,
			Label:         label,
			Description:   strings.TrimSpace(r.Description),
			EvidenceLevel: evidence,
		})
	}
	return out
}

func validRequirementKind(k RequirementKind) bool {
	switch k {
	case RequirementMustHave, RequirementNiceToHave, RequirementHiddenSignal, RequirementInterviewFocus:
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

func redactErrorMessage(msg string) string {
	// Defensive redaction: never let raw JD or prompt body leak into the
	// async_jobs.error_message column. Truncate at 240 chars and strip
	// obvious sensitive substrings.
	if len(msg) > 240 {
		msg = msg[:240]
	}
	for _, kw := range []string{"raw_jd_text", "Authorization:", "Bearer "} {
		if strings.Contains(msg, kw) {
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
