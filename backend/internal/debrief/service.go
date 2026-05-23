package debrief

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type Store interface {
	CreateDebrief(ctx context.Context, in CreateDebriefStoreInput) (CreateDebriefResult, error)
	GetDebrief(ctx context.Context, userID, debriefID string) (DebriefRecord, error)
	UpdateDebriefCompleted(ctx context.Context, in UpdateDebriefCompletedInput) (DebriefRecord, error)
}

type AuditRecorder interface {
	RecordDebriefAuditEvent(ctx context.Context, event DebriefAuditEvent) error
}

type SuggestionContextStore interface {
	GetSuggestionContext(ctx context.Context, in SuggestionContextRequest) (SuggestionContext, error)
}

type PromptResolver interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type ServiceOptions struct {
	Store             Store
	SuggestionContext SuggestionContextStore
	Registry          PromptResolver
	AI                aiclient.AIClient
	AITaskRuns        aiclient.AITaskRunWriter
	Audit             AuditRecorder
	Now               func() time.Time
	NewID             func() string
}

type Service struct {
	store             Store
	suggestionContext SuggestionContextStore
	registry          PromptResolver
	ai                aiclient.AIClient
	aiTaskRuns        aiclient.AITaskRunWriter
	audit             AuditRecorder
	now               func() time.Time
	newID             func() string
}

func NewService(opts ServiceOptions) *Service {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	return &Service{
		store:             opts.Store,
		suggestionContext: opts.SuggestionContext,
		registry:          opts.Registry,
		ai:                opts.AI,
		aiTaskRuns:        opts.AITaskRuns,
		audit:             opts.Audit,
		now:               now,
		newID:             newID,
	}
}

const (
	AuditActionCreateDebrief           = "create_debrief"
	AuditActionCompleteDebrief         = "complete_debrief"
	AuditActionSuggestDebriefQuestions = "suggest_debrief_questions"
	AuditResultSuccess                 = "success"
	debriefSuggestOutputSchemaVersion  = "debrief.suggest_questions.v1"
)

var debriefSuggestFeatureKey = featurekeys.DebriefSuggestQuestions.String()

type DebriefAuditEvent struct {
	AuditEventID string
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	Metadata     map[string]any
	CreatedAt    time.Time
}

func (s *Service) CreateDebrief(ctx context.Context, in CreateDebriefRequest) (CreateDebriefResult, error) {
	if s == nil || s.store == nil {
		return CreateDebriefResult{}, fmt.Errorf("debrief service store is not configured")
	}
	now := s.now().UTC()
	storeInput := CreateDebriefStoreInput{
		DebriefID:       s.newID(),
		JobID:           s.newID(),
		OutboxEventID:   s.newID(),
		UserID:          in.UserID,
		TargetJobID:     in.TargetJobID,
		RoundType:       in.RoundType,
		InterviewerRole: in.InterviewerRole,
		Language:        in.Language,
		Notes:           in.Notes,
		Questions:       in.Questions,
		Now:             now,
	}
	result, err := s.store.CreateDebrief(ctx, storeInput)
	if err != nil {
		return CreateDebriefResult{}, err
	}
	s.recordAudit(ctx, DebriefAuditEvent{
		AuditEventID: s.newID(),
		UserID:       in.UserID,
		Action:       AuditActionCreateDebrief,
		ResourceType: string(api.ResourceTypeDebrief),
		ResourceID:   result.DebriefID,
		Result:       AuditResultSuccess,
		Metadata: map[string]any{
			"debrief_id":     result.DebriefID,
			"target_job_id":  in.TargetJobID,
			"language":       in.Language,
			"question_count": len(in.Questions),
			"status":         string(sharedtypes.DebriefStatusDraft),
		},
		CreatedAt: now,
	})
	return result, nil
}

func (s *Service) GetDebrief(ctx context.Context, userID, debriefID string) (DebriefRecord, error) {
	if s == nil || s.store == nil {
		return DebriefRecord{}, ErrDebriefValidation
	}
	userID = strings.TrimSpace(userID)
	debriefID = strings.TrimSpace(debriefID)
	if userID == "" || debriefID == "" {
		return DebriefRecord{}, ErrDebriefValidation
	}
	rec, err := s.store.GetDebrief(ctx, userID, debriefID)
	if err != nil {
		return DebriefRecord{}, err
	}
	if rec.Status != sharedtypes.DebriefStatusCompleted {
		rec.Provenance = nil
		return rec, nil
	}
	if rec.Provenance == nil {
		rec.Provenance = &Provenance{}
	}
	if strings.TrimSpace(rec.Provenance.FeatureFlag) == "" {
		rec.Provenance.FeatureFlag = "none"
	}
	if strings.TrimSpace(rec.Provenance.DataSourceVersion) == "" {
		rec.Provenance.DataSourceVersion = "debrief/" + rec.ID + "@v1"
	}
	return rec, nil
}

func (s *Service) recordAudit(ctx context.Context, event DebriefAuditEvent) {
	if s == nil || s.audit == nil {
		return
	}
	_ = s.audit.RecordDebriefAuditEvent(ctx, event)
}

type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

type QuestionInput struct {
	QuestionText        string `json:"questionText"`
	MyAnswerSummary     string `json:"myAnswerSummary"`
	InterviewerReaction string `json:"interviewerReaction,omitempty"`
}

type CreateDebriefRequest struct {
	UserID          string
	TargetJobID     string
	RoundType       sharedtypes.DebriefRoundType
	InterviewerRole sharedtypes.InterviewerRole
	Language        string
	Notes           string
	Questions       []QuestionInput
}

type SuggestQuestionsRequest struct {
	UserID          string
	TargetJobID     string
	SessionID       string
	ResumeVersionID string
	Language        string
	Count           int32
}

type SuggestionContextRequest struct {
	UserID          string
	TargetJobID     string
	SessionID       string
	ResumeVersionID string
}

type SuggestionContext struct {
	TargetJobID    string
	Title          string
	CompanyName    string
	Summary        string
	SessionSummary string
	ResumeSummary  string
}

type SuggestQuestionsResult struct {
	Suggestions []SuggestedQuestion
}

type SuggestedQuestion struct {
	Stage          string
	QuestionText   string
	WhyLikelyAsked string
	Source         sharedtypes.DebriefQuestionSource
}

func (s *Service) SuggestQuestions(ctx context.Context, in SuggestQuestionsRequest) (SuggestQuestionsResult, error) {
	if s == nil || s.suggestionContext == nil {
		return SuggestQuestionsResult{}, fmt.Errorf("debrief suggestion context store is not configured")
	}
	count := in.Count
	if count == 0 {
		count = 6
	}
	suggestionContext, err := s.suggestionContext.GetSuggestionContext(ctx, SuggestionContextRequest{
		UserID:          in.UserID,
		TargetJobID:     in.TargetJobID,
		SessionID:       in.SessionID,
		ResumeVersionID: in.ResumeVersionID,
	})
	if err != nil {
		return SuggestQuestionsResult{}, err
	}
	if s.registry == nil || s.ai == nil {
		return SuggestQuestionsResult{}, &ServiceError{Code: sharederrors.CodeAiProviderConfigInvalid, Message: "AI provider is not configured"}
	}
	taskCtx := aiclient.AITaskRunContext{
		ID:                  s.newID(),
		UserID:              in.UserID,
		Capability:          aiclient.AITaskRunTaskDebriefSuggestQuestions,
		ResourceType:        aiclient.AITaskRunResourceTargetJob,
		ResourceID:          in.TargetJobID,
		OutputSchemaVersion: debriefSuggestOutputSchemaVersion,
	}
	startedAt := s.now().UTC()
	resolution, err := s.registry.ResolveActive(ctx, debriefSuggestFeatureKey, in.Language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		meta := suggestFailureMeta(registry.PromptResolution{}, in.Language, code)
		s.writeSuggestTaskRun(ctx, meta, taskCtx, startedAt, s.now().UTC(), fmt.Errorf("%s", code))
		return SuggestQuestionsResult{}, &ServiceError{Code: code, Message: "debrief question suggestion prompt could not be resolved"}
	}

	resp, meta, err := s.ai.Complete(ctx, resolution.ModelProfileName, buildSuggestQuestionsPayload(resolution, suggestionContext, in, count, taskCtx))
	meta = fillSuggestMeta(meta, resolution, in.Language)
	if err != nil {
		code := sharederrors.CodeAiProviderConfigInvalid
		if mapped, ok := aiErrorCode(err); ok {
			code = mapped
		}
		if meta.ErrorCode == "" {
			meta.ErrorCode = code
		}
		s.writeSuggestTaskRun(ctx, meta, taskCtx, startedAt, s.now().UTC(), err)
		return SuggestQuestionsResult{}, &ServiceError{Code: code, Message: "debrief question suggestion AI call failed"}
	}
	suggestions, err := parseSuggestQuestions(resp.Content)
	if err != nil {
		meta.ErrorCode = sharederrors.CodeAiOutputInvalid
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
		s.writeSuggestTaskRun(ctx, meta, taskCtx, startedAt, s.now().UTC(), err)
		return SuggestQuestionsResult{}, &ServiceError{Code: sharederrors.CodeAiOutputInvalid, Message: "debrief question suggestion output is invalid"}
	}
	s.writeSuggestTaskRun(ctx, meta, taskCtx, startedAt, s.now().UTC(), nil)
	s.recordAudit(ctx, DebriefAuditEvent{
		AuditEventID: s.newID(),
		UserID:       in.UserID,
		Action:       AuditActionSuggestDebriefQuestions,
		ResourceType: string(api.ResourceTypeTargetJob),
		ResourceID:   in.TargetJobID,
		Result:       AuditResultSuccess,
		Metadata: map[string]any{
			"target_job_id":    in.TargetJobID,
			"language":         in.Language,
			"suggestion_count": len(suggestions),
		},
		CreatedAt: s.now().UTC(),
	})
	return SuggestQuestionsResult{Suggestions: suggestions}, nil
}

func buildSuggestQuestionsPayload(resolution registry.PromptResolution, ctx SuggestionContext, in SuggestQuestionsRequest, count int32, taskCtx aiclient.AITaskRunContext) aiclient.CompletePayload {
	userContent := renderSuggestTemplate(resolution.UserMessageTemplate, ctx, in, count)
	if userContent == "" {
		userContent = fmt.Sprintf("Suggest %d likely debrief questions for target job %s.", count, ctx.TargetJobID)
	}
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: userContent})
	metadata := aiclient.CallMetadata{
		FeatureKey:        debriefSuggestFeatureKey,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          in.Language,
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

func renderSuggestTemplate(template string, ctx SuggestionContext, in SuggestQuestionsRequest, count int32) string {
	content := strings.TrimSpace(template)
	if content == "" {
		return ""
	}
	replacements := map[string]string{
		"{{targetTitle}}":    ctx.Title,
		"{{companyName}}":    ctx.CompanyName,
		"{{targetSummary}}":  ctx.Summary,
		"{{sessionSummary}}": ctx.SessionSummary,
		"{{resumeSummary}}":  ctx.ResumeSummary,
		"{{language}}":       in.Language,
		"{{count}}":          fmt.Sprintf("%d", count),
	}
	for marker, value := range replacements {
		content = strings.ReplaceAll(content, marker, value)
	}
	return content
}

func fillSuggestMeta(meta aiclient.AICallMeta, resolution registry.PromptResolution, language string) aiclient.AICallMeta {
	if meta.FeatureKey == "" {
		meta.FeatureKey = debriefSuggestFeatureKey
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

func suggestFailureMeta(resolution registry.PromptResolution, language string, code string) aiclient.AICallMeta {
	return fillSuggestMeta(aiclient.AICallMeta{
		Provider:         "not_applicable",
		ModelFamily:      "not_applicable",
		ModelID:          "model-profile:unknown",
		ValidationStatus: aiclient.ValidationStatusInvalid,
		ErrorCode:        code,
	}, resolution, language)
}

func (s *Service) writeSuggestTaskRun(ctx context.Context, meta aiclient.AICallMeta, taskCtx aiclient.AITaskRunContext, startedAt, completedAt time.Time, callErr error) {
	if s == nil || s.aiTaskRuns == nil {
		return
	}
	row, err := observability.AITaskRunRowFromMeta(meta, taskCtx, aiclient.AuditMetadata{}, startedAt, completedAt, callErr)
	if err != nil {
		return
	}
	_ = s.aiTaskRuns.WriteAITaskRun(ctx, row)
}

type suggestQuestionsAIResponse struct {
	Suggestions []suggestedQuestionAI `json:"suggestions"`
}

type suggestedQuestionAI struct {
	Stage          string `json:"stage"`
	QuestionText   string `json:"questionText"`
	WhyLikelyAsked string `json:"whyLikelyAsked"`
	Source         string `json:"source"`
}

func parseSuggestQuestions(content string) ([]SuggestedQuestion, error) {
	var out suggestQuestionsAIResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &out); err != nil {
		return nil, fmt.Errorf("parse debrief suggestions: %w", err)
	}
	if len(out.Suggestions) == 0 {
		return nil, fmt.Errorf("parse debrief suggestions: empty suggestions")
	}
	suggestions := make([]SuggestedQuestion, 0, len(out.Suggestions))
	for _, item := range out.Suggestions {
		questionText := strings.TrimSpace(item.QuestionText)
		why := strings.TrimSpace(item.WhyLikelyAsked)
		source := sharedtypes.DebriefQuestionSource(strings.TrimSpace(item.Source))
		if questionText == "" || why == "" || !validDebriefQuestionSource(source) {
			return nil, fmt.Errorf("parse debrief suggestions: invalid item")
		}
		suggestions = append(suggestions, SuggestedQuestion{
			Stage:          strings.TrimSpace(item.Stage),
			QuestionText:   questionText,
			WhyLikelyAsked: why,
			Source:         source,
		})
	}
	return suggestions, nil
}

func validDebriefQuestionSource(source sharedtypes.DebriefQuestionSource) bool {
	for _, allowed := range sharedtypes.AllDebriefQuestionSources {
		if source == allowed {
			return true
		}
	}
	return false
}

func aiErrorCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	var apiErr *sharederrors.APIError
	if stderrs.As(err, &apiErr) && apiErr.Code != "" {
		return apiErr.Code, true
	}
	if stderrs.Is(err, context.DeadlineExceeded) {
		return sharederrors.CodeAiProviderTimeout, true
	}
	msg := err.Error()
	for _, code := range []string{
		sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiFallbackExhausted,
		sharederrors.CodeAiUnsupportedCapability,
		sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeAiProviderSecretMissing,
	} {
		if strings.Contains(msg, code) {
			return code, true
		}
	}
	return "", false
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

type CreateDebriefStoreInput struct {
	DebriefID       string
	JobID           string
	OutboxEventID   string
	UserID          string
	TargetJobID     string
	RoundType       sharedtypes.DebriefRoundType
	InterviewerRole sharedtypes.InterviewerRole
	Language        string
	Notes           string
	Questions       []QuestionInput
	Now             time.Time
}

type CreateDebriefResult struct {
	DebriefID string
	Job       JobRecord
}

type JobRecord struct {
	ID           string
	JobType      api.JobType
	ResourceType api.ResourceType
	ResourceID   string
	Status       sharedtypes.JobStatus
	ErrorCode    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DebriefRecord struct {
	ID                 string
	TargetJobID        string
	Status             sharedtypes.DebriefStatus
	RoundType          sharedtypes.DebriefRoundType
	InterviewerRole    string
	Questions          []QuestionRecord
	RiskItems          []RiskItem
	NextRoundChecklist []NextRoundChecklistItem
	ThankYouDraft      string
	Provenance         *Provenance
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type QuestionRecord struct {
	QuestionText        string `json:"questionText"`
	MyAnswerSummary     string `json:"myAnswerSummary"`
	InterviewerReaction string `json:"interviewerReaction,omitempty"`
	AIAnalysis          string `json:"aiAnalysis,omitempty"`
}

type RiskItem struct {
	Label    string `json:"label"`
	Severity string `json:"severity"`
}

type NextRoundChecklistItem struct {
	Label     string `json:"label"`
	Rationale string `json:"rationale,omitempty"`
}

type Provenance struct {
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
}

type UpdateDebriefCompletedInput struct {
	UserID        string
	DebriefID     string
	OutboxEventID string
	Questions     []QuestionRecord
	RiskItems     []RiskItem
	Provenance    Provenance
	Now           time.Time
}
