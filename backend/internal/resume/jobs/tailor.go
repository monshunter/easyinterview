package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const (
	FeatureKeyResumeTailorGapReview         = string(featurekeys.ResumeTailorGapReview)
	FeatureKeyResumeTailorBulletSuggestions = string(featurekeys.ResumeTailorBulletSuggestions)
)

type TailorStore interface {
	MarkTailorRunGenerating(ctx context.Context, in resumestore.TailorRunStatusInput) (resumestore.TailorRunRecord, error)
	GetForTailor(ctx context.Context, tailorRunID string, resumeVersionID string) (resumestore.TailorJobContext, error)
	CompleteTailorRunSuccess(ctx context.Context, in resumestore.CompleteTailorRunSuccessInput) error
	MarkTailorRunFailed(ctx context.Context, in resumestore.TailorRunFailureInput) (resumestore.TailorRunRecord, error)
}

type TailorHandlerOptions struct {
	Store      TailorStore
	Registry   PromptRegistryClient
	AI         aiclient.AIClient
	AITaskRuns aiclient.AITaskRunWriter
	NewID      func() string
	Now        func() time.Time
}

type TailorHandler struct {
	store      TailorStore
	registry   PromptRegistryClient
	ai         aiclient.AIClient
	aiTaskRuns aiclient.AITaskRunWriter
	newID      func() string
	now        func() time.Time
}

func NewTailorHandler(opts TailorHandlerOptions) *TailorHandler {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	return &TailorHandler{
		store:      opts.Store,
		registry:   opts.Registry,
		ai:         opts.AI,
		aiTaskRuns: opts.AITaskRuns,
		newID:      newID,
		now:        now,
	}
}

func (h *TailorHandler) Handle(ctx context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
	if h == nil || h.store == nil || h.registry == nil || h.ai == nil || h.newID == nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: "resume tailor handler not initialised"}
	}
	payload, err := decodeTailorJobPayload(job)
	if err != nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeValidationFailed, ErrorMessage: sharederrors.CodeValidationFailed}
	}
	if _, err := h.store.MarkTailorRunGenerating(ctx, resumestore.TailorRunStatusInput{TailorRunID: payload.TailorRunID, Now: h.now()}); err != nil {
		return targetjob.JobOutcome{
			ErrorCode:    sharederrors.CodeTargetInvalidStateTransition,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetInvalidStateTransition, err.Error()),
		}
	}
	tailorCtx, err := h.store.GetForTailor(ctx, payload.TailorRunID, payload.ResumeVersionID)
	if err != nil {
		return h.fail(ctx, payload.TailorRunID, sharederrors.CodeTargetImportFailed, err.Error(), false)
	}
	featureKey, err := tailorFeatureKey(tailorCtx.Mode)
	if err != nil {
		return h.fail(ctx, payload.TailorRunID, sharederrors.CodeValidationFailed, err.Error(), false)
	}
	resolution, err := h.registry.Resolve(ctx, featureKey, tailorCtx.Language)
	if err != nil {
		return h.fail(ctx, payload.TailorRunID, sharederrors.CodeAiProviderConfigInvalid, err.Error(), false)
	}

	taskCtx := aiclient.AITaskRunContext{
		ID:                  h.newID(),
		UserID:              tailorCtx.UserID,
		Capability:          aiclient.AITaskRunTaskResumeTailor,
		ResourceType:        aiclient.AITaskRunResourceResumeTailorRun,
		ResourceID:          tailorCtx.TailorRunID,
		OutputSchemaVersion: "resume.tailor.v1",
	}
	startedAt := h.now().UTC()
	complete, meta, err := h.ai.Complete(ctx, resolution.ModelProfileName, aiclient.CompletePayload{
		Messages: buildTailorPromptMessages(resolution, tailorCtx),
		Metadata: aiclient.CallMetadata{
			FeatureKey:        featureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     resolution.RubricVersion,
			Language:          tailorCtx.Language,
			FeatureFlag:       coalesceFlag(resolution.FeatureFlag),
			DataSourceVersion: resolution.DataSourceVersion,
			TaskRun:           taskCtx,
		},
	})
	completedAt := h.now().UTC()
	meta = enrichTailorMeta(meta, resolution, featureKey, tailorCtx.Language, "")
	if err != nil {
		code, retryable := translateAIClientError(err)
		meta = enrichTailorMeta(meta, resolution, featureKey, tailorCtx.Language, code)
		h.writeTaskRun(ctx, meta, taskCtx, startedAt, completedAt, err)
		return h.fail(ctx, payload.TailorRunID, code, err.Error(), retryable)
	}
	parsed, err := decodeTailorAIResponse(complete.Content, tailorCtx)
	if err != nil {
		meta = enrichTailorMeta(meta, resolution, featureKey, tailorCtx.Language, sharederrors.CodeAiOutputInvalid)
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
		h.writeTaskRun(ctx, meta, taskCtx, startedAt, completedAt, fmt.Errorf("%s: %w", sharederrors.CodeAiOutputInvalid, err))
		return h.fail(ctx, payload.TailorRunID, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}
	h.writeTaskRun(ctx, meta, taskCtx, startedAt, completedAt, nil)

	outboxPayload, err := json.Marshal(events.ResumeTailorCompletedPayload{
		TailorRunID:   tailorCtx.TailorRunID,
		ResumeAssetID: tailorCtx.ResumeAssetID,
		TargetJobID:   tailorCtx.TargetJobID,
		Mode:          events.ResumeTailorMode(tailorCtx.Mode),
		Status:        sharedtypes.ReportStatusReady,
	})
	if err != nil {
		return h.fail(ctx, payload.TailorRunID, sharederrors.CodeTargetImportFailed, err.Error(), true)
	}
	suggestions := make([]resumestore.TailorSuggestionInput, 0, len(parsed.Suggestions))
	for _, suggestion := range parsed.Suggestions {
		suggestions = append(suggestions, resumestore.TailorSuggestionInput{
			ID:              h.newID(),
			OriginalBullet:  suggestion.OriginalBullet,
			SuggestedBullet: suggestion.SuggestedBullet,
			Reason:          suggestion.Reason,
		})
	}
	if err := h.store.CompleteTailorRunSuccess(ctx, resumestore.CompleteTailorRunSuccessInput{
		TailorRunID:        tailorCtx.TailorRunID,
		ResumeVersionID:    tailorCtx.ResumeVersionID,
		MatchSummary:       parsed.MatchSummary,
		Suggestions:        suggestions,
		Provenance:         tailorProvenance(resolution, meta, tailorCtx.Language),
		OutboxEventID:      h.newID(),
		OutboxEventPayload: outboxPayload,
		Now:                h.now(),
	}); err != nil {
		return targetjob.JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	return targetjob.JobOutcome{Succeeded: true}
}

func (h *TailorHandler) fail(ctx context.Context, tailorRunID string, code, message string, retryable bool) targetjob.JobOutcome {
	if _, err := h.store.MarkTailorRunFailed(ctx, resumestore.TailorRunFailureInput{
		TailorRunID: tailorRunID,
		ErrorCode:   code,
		Now:         h.now(),
	}); err != nil {
		return targetjob.JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	return targetjob.JobOutcome{
		ErrorCode:    code,
		ErrorMessage: safeFailureMessage(code, message),
		Retryable:    retryable,
	}
}

func (h *TailorHandler) writeTaskRun(ctx context.Context, meta aiclient.AICallMeta, taskCtx aiclient.AITaskRunContext, startedAt, completedAt time.Time, callErr error) {
	if h == nil || h.aiTaskRuns == nil {
		return
	}
	row, err := observability.AITaskRunRowFromMeta(meta, taskCtx, aiclient.AuditMetadata{}, startedAt, completedAt, callErr)
	if err != nil {
		return
	}
	_ = h.aiTaskRuns.WriteAITaskRun(ctx, row)
}

type tailorJobPayload struct {
	TailorRunID     string `json:"tailorRunId"`
	ResumeVersionID string `json:"resumeVersionId,omitempty"`
}

func decodeTailorJobPayload(job targetjob.ClaimedJob) (tailorJobPayload, error) {
	payload := tailorJobPayload{TailorRunID: strings.TrimSpace(job.ResourceID)}
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return tailorJobPayload{}, fmt.Errorf("decode resume tailor job payload: %w", err)
		}
		payload.TailorRunID = strings.TrimSpace(payload.TailorRunID)
		payload.ResumeVersionID = strings.TrimSpace(payload.ResumeVersionID)
	}
	if payload.TailorRunID == "" {
		return tailorJobPayload{}, fmt.Errorf("tailorRunId is required")
	}
	return payload, nil
}

func tailorFeatureKey(mode string) (string, error) {
	switch strings.TrimSpace(mode) {
	case "gap_review":
		return FeatureKeyResumeTailorGapReview, nil
	case "bullet_suggestions":
		return FeatureKeyResumeTailorBulletSuggestions, nil
	default:
		return "", fmt.Errorf("unsupported resume tailor mode %q", mode)
	}
}

func buildTailorPromptMessages(resolution PromptResolution, ctx resumestore.TailorJobContext) []aiclient.Message {
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	user := strings.TrimSpace(resolution.UserMessageTemplate)
	replacements := map[string]string{
		"{{resume_summary}}":     compactJSONString(ctx.ResumeSummary),
		"{{structured_profile}}": compactJSONString(ctx.StructuredProfile),
		"{{jd_summary}}":         compactJSONString(ctx.TargetSummary),
		"{{jd_context}}":         strings.TrimSpace(ctx.TargetTitle + "\n" + ctx.TargetCompany + "\n" + compactJSONString(ctx.TargetSummary) + "\n" + ctx.RawJDText),
		"{{target_seniority}}":   ctx.TargetSeniority,
		"{{original_bullet}}":    ctx.OriginalBullet,
		"{{tone}}":               "truthful, concise, impact-driven",
		"{{language}}":           ctx.Language,
	}
	for placeholder, value := range replacements {
		user = strings.ReplaceAll(user, placeholder, value)
	}
	if strings.TrimSpace(user) == "" {
		user = compactJSONString(mustMarshalTailorPromptContext(ctx))
	}
	if strings.TrimSpace(user) != "" {
		messages = append(messages, aiclient.Message{Role: "user", Content: strings.TrimSpace(user)})
	}
	return messages
}

func compactJSONString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return strings.TrimSpace(string(raw))
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return strings.TrimSpace(string(raw))
	}
	return string(encoded)
}

func mustMarshalTailorPromptContext(ctx resumestore.TailorJobContext) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"resumeSummary":     json.RawMessage(compactJSONString(ctx.ResumeSummary)),
		"structuredProfile": json.RawMessage(compactJSONString(ctx.StructuredProfile)),
		"targetSummary":     json.RawMessage(compactJSONString(ctx.TargetSummary)),
		"targetTitle":       ctx.TargetTitle,
		"targetCompany":     ctx.TargetCompany,
		"targetSeniority":   ctx.TargetSeniority,
		"originalBullet":    ctx.OriginalBullet,
		"language":          ctx.Language,
	})
	return raw
}

type decodedTailorOutput struct {
	MatchSummary json.RawMessage
	Suggestions  []decodedTailorSuggestion
}

type decodedTailorSuggestion struct {
	OriginalBullet  string
	SuggestedBullet string
	Reason          string
}

func decodeTailorAIResponse(content string, ctx resumestore.TailorJobContext) (decodedTailorOutput, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return decodedTailorOutput{}, fmt.Errorf("AI response content was empty")
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(content), &root); err != nil {
		return decodedTailorOutput{}, fmt.Errorf("AI response is not valid JSON: %v", err)
	}
	matchSummary, err := normalizeMatchSummary(root)
	if err != nil {
		return decodedTailorOutput{}, err
	}
	suggestions := normalizeSuggestions(root["suggestions"], ctx.OriginalBullet)
	if len(matchSummary) == 0 && len(suggestions) == 0 {
		return decodedTailorOutput{}, fmt.Errorf("AI response missing match summary and suggestions")
	}
	if len(matchSummary) == 0 {
		matchSummary = json.RawMessage(`{}`)
	}
	return decodedTailorOutput{MatchSummary: matchSummary, Suggestions: suggestions}, nil
}

func normalizeMatchSummary(root map[string]any) (json.RawMessage, error) {
	if raw, ok := root["matchSummary"]; ok {
		encoded, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshal matchSummary: %w", err)
		}
		var typed struct {
			Strengths []string `json:"strengths"`
			Gaps      []string `json:"gaps"`
		}
		if err := json.Unmarshal(encoded, &typed); err != nil {
			return nil, fmt.Errorf("matchSummary schema invalid: %w", err)
		}
		return encoded, nil
	}
	strengths := stringListFromObjects(root["strengths_to_amplify"], "topic", "evidence")
	gaps := stringListFromObjects(root["gaps"], "topic", "why")
	if len(strengths) == 0 && len(gaps) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(map[string][]string{"strengths": strengths, "gaps": gaps})
	if err != nil {
		return nil, fmt.Errorf("marshal normalized match summary: %w", err)
	}
	return encoded, nil
}

func normalizeSuggestions(raw any, fallbackOriginal string) []decodedTailorSuggestion {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]decodedTailorSuggestion, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		original := firstNonEmptyString(obj["originalBullet"], obj["original_bullet"], fallbackOriginal)
		suggested := firstNonEmptyString(obj["suggestedBullet"], obj["suggested_bullet"], obj["re"+"write"])
		reason := firstNonEmptyString(obj["reason"], obj["why_better"], obj["whyBetter"])
		if strings.TrimSpace(original) == "" || strings.TrimSpace(suggested) == "" {
			continue
		}
		out = append(out, decodedTailorSuggestion{
			OriginalBullet:  strings.TrimSpace(original),
			SuggestedBullet: strings.TrimSpace(suggested),
			Reason:          strings.TrimSpace(reason),
		})
	}
	return out
}

func stringListFromObjects(raw any, fields ...string) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		switch typed := item.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				out = append(out, strings.TrimSpace(typed))
			}
		case map[string]any:
			parts := make([]string, 0, len(fields))
			for _, field := range fields {
				if value := firstNonEmptyString(typed[field]); value != "" {
					parts = append(parts, value)
				}
			}
			if len(parts) > 0 {
				out = append(out, strings.Join(parts, ": "))
			}
		}
	}
	return out
}

func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		}
	}
	return ""
}

func enrichTailorMeta(meta aiclient.AICallMeta, resolution PromptResolution, featureKey string, language string, errorCode string) aiclient.AICallMeta {
	if meta.PromptVersion == "" {
		meta.PromptVersion = resolution.PromptVersion
	}
	if meta.RubricVersion == "" {
		meta.RubricVersion = resolution.RubricVersion
	}
	if meta.ModelProfileName == "" {
		meta.ModelProfileName = resolution.ModelProfileName
	}
	if meta.Language == "" {
		meta.Language = language
	}
	if meta.FeatureKey == "" {
		meta.FeatureKey = featureKey
	}
	if meta.FeatureFlag == "" {
		meta.FeatureFlag = coalesceFlag(resolution.FeatureFlag)
	}
	if meta.DataSourceVersion == "" {
		meta.DataSourceVersion = resolution.DataSourceVersion
	}
	if errorCode != "" {
		meta.ErrorCode = errorCode
		if meta.ValidationStatus == "" {
			meta.ValidationStatus = aiclient.ValidationStatusInvalid
		}
	}
	if meta.ValidationStatus == "" {
		meta.ValidationStatus = aiclient.ValidationStatusOK
	}
	return meta
}

func tailorProvenance(resolution PromptResolution, meta aiclient.AICallMeta, language string) resumestore.VersionProvenance {
	return resumestore.VersionProvenance{
		PromptVersion:     firstNonEmptyString(meta.PromptVersion, resolution.PromptVersion),
		RubricVersion:     firstNonEmptyString(meta.RubricVersion, resolution.RubricVersion),
		ModelID:           meta.ModelID,
		Provider:          meta.Provider,
		Language:          firstNonEmptyString(meta.Language, language),
		FeatureFlag:       coalesceFlag(firstNonEmptyString(meta.FeatureFlag, resolution.FeatureFlag)),
		DataSourceVersion: firstNonEmptyString(meta.DataSourceVersion, resolution.DataSourceVersion),
	}
}
