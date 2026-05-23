package debrief

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

type GenerateStore interface {
	LoadGenerateContext(ctx context.Context, payload GenerateJobPayload) (GenerateContext, error)
	UpdateDebriefCompleted(ctx context.Context, in UpdateDebriefCompletedInput) (DebriefRecord, error)
}

type GenerateHandlerOptions struct {
	Store      GenerateStore
	Registry   PromptResolver
	AI         aiclient.AIClient
	AITaskRuns aiclient.AITaskRunWriter
	Audit      AuditRecorder
	Now        func() time.Time
	NewID      func() string
}

type GenerateHandler struct {
	store      GenerateStore
	registry   PromptResolver
	ai         aiclient.AIClient
	aiTaskRuns aiclient.AITaskRunWriter
	audit      AuditRecorder
	now        func() time.Time
	newID      func() string
}

type GenerateJobPayload struct {
	DebriefID     string `json:"debriefId"`
	TargetJobID   string `json:"targetJobId"`
	Language      string `json:"language"`
	QuestionCount int    `json:"questionCount"`
}

type GenerateContext struct {
	UserID        string
	DebriefID     string
	TargetJobID   string
	Language      string
	TargetTitle   string
	TargetSummary string
	Questions     []QuestionInput
}

func NewGenerateHandler(opts GenerateHandlerOptions) *GenerateHandler {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	return &GenerateHandler{
		store:      opts.Store,
		registry:   opts.Registry,
		ai:         opts.AI,
		aiTaskRuns: opts.AITaskRuns,
		audit:      opts.Audit,
		now:        now,
		newID:      newID,
	}
}

func (h *GenerateHandler) Handle(ctx context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
	if h == nil || h.store == nil || h.registry == nil || h.ai == nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeAiProviderConfigInvalid, ErrorMessage: "debrief generate handler is not configured", Retryable: retryableForAttempt(job)}
	}
	payload, err := parseGenerateJobPayload(job)
	if err != nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: false}
	}
	generateContext, err := h.store.LoadGenerateContext(ctx, payload)
	if err != nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: retryableForAttempt(job)}
	}
	language := fallbackString(payload.Language, generateContext.Language)
	resolution, err := h.registry.ResolveActive(ctx, featurekeys.DebriefGenerate.String(), language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		h.writeGenerateTaskRun(ctx, generateFailureMeta(resolution, language, code), generateTaskContext(h.newID(), generateContext), h.now().UTC(), h.now().UTC(), fmt.Errorf("%s", code))
		return targetjob.JobOutcome{ErrorCode: code, ErrorMessage: err.Error(), Retryable: retryableForAttempt(job)}
	}
	taskCtx := generateTaskContext(h.newID(), generateContext)
	startedAt := h.now().UTC()
	resp, meta, err := h.ai.Complete(ctx, resolution.ModelProfileName, buildGenerateCompletePayload(resolution, generateContext, language, taskCtx))
	meta = fillGenerateMeta(meta, resolution, language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		if meta.ErrorCode == "" {
			meta.ErrorCode = code
		}
		h.writeGenerateTaskRun(ctx, meta, taskCtx, startedAt, h.now().UTC(), err)
		return targetjob.JobOutcome{ErrorCode: code, ErrorMessage: err.Error(), Retryable: retryableForAttempt(job)}
	}
	parsed, err := parseGenerateResponse(resp.Content)
	if err != nil {
		meta.ErrorCode = sharederrors.CodeAiOutputInvalid
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
		h.writeGenerateTaskRun(ctx, meta, taskCtx, startedAt, h.now().UTC(), err)
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: retryableForAttempt(job)}
	}
	h.writeGenerateTaskRun(ctx, meta, taskCtx, startedAt, h.now().UTC(), nil)
	if _, err := h.store.UpdateDebriefCompleted(ctx, UpdateDebriefCompletedInput{
		UserID:        generateContext.UserID,
		DebriefID:     generateContext.DebriefID,
		OutboxEventID: h.newID(),
		Questions:     parsed.Questions,
		RiskItems:     parsed.RiskItems,
		Provenance: Provenance{
			PromptVersion:     fallbackString(meta.PromptVersion, resolution.PromptVersion),
			RubricVersion:     fallbackString(meta.RubricVersion, resolution.RubricVersion),
			ModelID:           fallbackString(meta.ModelID, "model-profile:"+resolution.ModelProfileName),
			Language:          language,
			FeatureFlag:       fallbackString(meta.FeatureFlag, resolution.FeatureFlag),
			DataSourceVersion: fallbackString(meta.DataSourceVersion, resolution.DataSourceVersion),
		},
		Now: h.now().UTC(),
	}); err != nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: retryableForAttempt(job)}
	}
	h.recordCompleteAudit(ctx, generateContext, len(parsed.Questions), h.now().UTC())
	return targetjob.JobOutcome{Succeeded: true}
}

func (h *GenerateHandler) recordCompleteAudit(ctx context.Context, generateContext GenerateContext, questionCount int, now time.Time) {
	if h == nil || h.audit == nil {
		return
	}
	_ = h.audit.RecordDebriefAuditEvent(ctx, DebriefAuditEvent{
		AuditEventID: h.newID(),
		UserID:       generateContext.UserID,
		Action:       AuditActionCompleteDebrief,
		ResourceType: "debrief",
		ResourceID:   generateContext.DebriefID,
		Result:       AuditResultSuccess,
		Metadata: map[string]any{
			"debrief_id":     generateContext.DebriefID,
			"target_job_id":  generateContext.TargetJobID,
			"language":       generateContext.Language,
			"question_count": questionCount,
			"status":         string(sharedtypes.DebriefStatusCompleted),
		},
		CreatedAt: now,
	})
}

func parseGenerateJobPayload(job targetjob.ClaimedJob) (GenerateJobPayload, error) {
	var payload GenerateJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return GenerateJobPayload{}, fmt.Errorf("parse debrief_generate payload: %w", err)
	}
	if strings.TrimSpace(payload.DebriefID) == "" {
		payload.DebriefID = strings.TrimSpace(job.ResourceID)
	}
	if strings.TrimSpace(payload.DebriefID) == "" || strings.TrimSpace(payload.TargetJobID) == "" {
		return GenerateJobPayload{}, fmt.Errorf("debrief_generate payload missing debriefId or targetJobId")
	}
	return payload, nil
}

func buildGenerateCompletePayload(resolution registry.PromptResolution, ctx GenerateContext, language string, taskCtx aiclient.AITaskRunContext) aiclient.CompletePayload {
	user := strings.TrimSpace(resolution.UserMessageTemplate)
	replacements := map[string]string{
		"{{language}}":      language,
		"{{targetTitle}}":   ctx.TargetTitle,
		"{{role_title}}":    ctx.TargetTitle,
		"{{targetSummary}}": ctx.TargetSummary,
		"{{questions}}":     mustJSONString(ctx.Questions),
		"{{notes}}":         mustJSONString(ctx.Questions),
		"{{transcript}}":    mustJSONString(ctx.Questions),
	}
	for marker, value := range replacements {
		user = strings.ReplaceAll(user, marker, value)
	}
	if strings.TrimSpace(user) == "" {
		user = "Analyze the debrief questions and return strict JSON with questions and riskItems."
	}
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: user})
	metadata := aiclient.CallMetadata{
		FeatureKey:        featurekeys.DebriefGenerate.String(),
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          language,
		FeatureFlag:       resolution.FeatureFlag,
		DataSourceVersion: resolution.DataSourceVersion,
		TaskRun:           taskCtx,
	}
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: metadata,
	}
}

func generateTaskContext(id string, ctx GenerateContext) aiclient.AITaskRunContext {
	return aiclient.AITaskRunContext{
		ID:                  id,
		UserID:              ctx.UserID,
		Capability:          aiclient.AITaskRunTaskDebriefGenerate,
		ResourceType:        aiclient.AITaskRunResourceDebrief,
		ResourceID:          ctx.DebriefID,
		OutputSchemaVersion: "debrief.generate.v1",
	}
}

func fillGenerateMeta(meta aiclient.AICallMeta, resolution registry.PromptResolution, language string) aiclient.AICallMeta {
	if meta.FeatureKey == "" {
		meta.FeatureKey = featurekeys.DebriefGenerate.String()
	}
	if meta.PromptVersion == "" {
		meta.PromptVersion = resolution.PromptVersion
	}
	if meta.RubricVersion == "" {
		meta.RubricVersion = resolution.RubricVersion
	}
	if meta.ModelProfileName == "" {
		meta.ModelProfileName = resolution.ModelProfileName
	}
	if meta.FeatureFlag == "" {
		meta.FeatureFlag = fallbackString(resolution.FeatureFlag, "none")
	}
	if meta.DataSourceVersion == "" {
		meta.DataSourceVersion = fallbackString(resolution.DataSourceVersion, "not_applicable")
	}
	if meta.Language == "" {
		meta.Language = fallbackString(language, "en")
	}
	if meta.ValidationStatus == "" {
		meta.ValidationStatus = aiclient.ValidationStatusOK
	}
	return meta
}

func generateFailureMeta(resolution registry.PromptResolution, language string, code string) aiclient.AICallMeta {
	return fillGenerateMeta(aiclient.AICallMeta{
		Provider:         "not_applicable",
		ModelFamily:      "not_applicable",
		ModelID:          "model-profile:unknown",
		ValidationStatus: aiclient.ValidationStatusInvalid,
		ErrorCode:        code,
	}, resolution, language)
}

func retryableForAttempt(job targetjob.ClaimedJob) bool {
	return job.MaxAttempts <= 0 || job.Attempts <= 0 || job.Attempts < job.MaxAttempts
}

func (h *GenerateHandler) writeGenerateTaskRun(ctx context.Context, meta aiclient.AICallMeta, taskCtx aiclient.AITaskRunContext, startedAt, completedAt time.Time, callErr error) {
	if h == nil || h.aiTaskRuns == nil {
		return
	}
	row, err := observability.AITaskRunRowFromMeta(meta, taskCtx, aiclient.AuditMetadata{}, startedAt, completedAt, callErr)
	if err != nil {
		return
	}
	_ = h.aiTaskRuns.WriteAITaskRun(ctx, row)
}

type generateAIResponse struct {
	Questions []QuestionRecord `json:"questions"`
	RiskItems []RiskItem       `json:"riskItems"`
}

func parseGenerateResponse(content string) (generateAIResponse, error) {
	var out generateAIResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &out); err != nil {
		return generateAIResponse{}, fmt.Errorf("parse debrief.generate response: %w", err)
	}
	if len(out.Questions) == 0 {
		return generateAIResponse{}, fmt.Errorf("parse debrief.generate response: empty questions")
	}
	for _, question := range out.Questions {
		if strings.TrimSpace(question.QuestionText) == "" || strings.TrimSpace(question.MyAnswerSummary) == "" || strings.TrimSpace(question.AIAnalysis) == "" {
			return generateAIResponse{}, fmt.Errorf("parse debrief.generate response: invalid question")
		}
	}
	return out, nil
}

func mustJSONString(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(raw)
}
