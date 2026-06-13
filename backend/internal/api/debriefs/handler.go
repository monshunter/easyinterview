package debriefs

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"net/http"
	"strings"

	domain "github.com/monshunter/easyinterview/backend/internal/debrief"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
)

type SessionResolver func(ctx context.Context) (userID string, ok bool)

type debriefService interface {
	CreateDebrief(ctx context.Context, in domain.CreateDebriefRequest) (domain.CreateDebriefResult, error)
	SuggestQuestions(ctx context.Context, in domain.SuggestQuestionsRequest) (domain.SuggestQuestionsResult, error)
	GetDebrief(ctx context.Context, userID, debriefID string) (domain.DebriefRecord, error)
}

type HandlerOptions struct {
	Service debriefService
	Session SessionResolver
}

type Handler struct {
	service debriefService
	session SessionResolver
}

func NewHandler(opts HandlerOptions) *Handler {
	return &Handler{service: opts.Service, session: opts.Session}
}

func (h *Handler) CreateDebrief(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "debrief service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.CreateDebriefRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	if field, message := validateCreateDebriefRequest(body); field != "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, message, map[string]any{"field": field})
		return
	}

	result, err := h.service.CreateDebrief(r.Context(), domain.CreateDebriefRequest{
		UserID:          userID,
		TargetJobID:     strings.TrimSpace(body.TargetJobId),
		RoundType:       body.RoundType,
		InterviewerRole: interviewerRoleValue(body.InterviewerRole),
		Language:        strings.TrimSpace(body.Language),
		Notes:           stringValue(body.Notes),
		Questions:       debriefQuestionsFromAPI(body.Questions),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	idempotency.SetResponseResource(w, string(api.ResourceTypeDebrief), result.DebriefID)
	writeJSON(w, http.StatusAccepted, toAPIDebriefWithJob(result))
}

func (h *Handler) SuggestDebriefQuestions(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "debrief service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	var body api.SuggestDebriefQuestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "request body is malformed", nil)
		return
	}
	count := suggestQuestionCount(body.Count)
	if field, message := validateSuggestDebriefQuestionsRequest(body, count); field != "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, message, map[string]any{"field": field})
		return
	}
	result, err := h.service.SuggestQuestions(r.Context(), domain.SuggestQuestionsRequest{
		UserID:          userID,
		TargetJobID:     strings.TrimSpace(body.TargetJobId),
		SessionID:       stringValue(body.SessionId),
		ResumeID:        stringValue(body.ResumeId),
		Language:        strings.TrimSpace(body.Language),
		Count:           count,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPISuggestDebriefQuestionsResponse(result))
}

func (h *Handler) GetDebrief(w http.ResponseWriter, r *http.Request, debriefID string) {
	if h == nil || h.service == nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "debrief service is not configured", nil)
		return
	}
	userID, ok := h.resolveUser(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required", nil)
		return
	}
	debriefID = strings.TrimSpace(debriefID)
	if debriefID == "" {
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "debriefId is required", map[string]any{"field": "debriefId"})
		return
	}
	result, err := h.service.GetDebrief(r.Context(), userID, debriefID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAPIDebrief(result))
}

func (h *Handler) resolveUser(r *http.Request) (string, bool) {
	if h.session == nil {
		return "", false
	}
	userID, ok := h.session(r.Context())
	userID = strings.TrimSpace(userID)
	return userID, ok && userID != ""
}

func validateCreateDebriefRequest(body api.CreateDebriefRequest) (field string, message string) {
	if strings.TrimSpace(body.TargetJobId) == "" {
		return "targetJobId", "targetJobId is required"
	}
	if strings.TrimSpace(body.Language) == "" {
		return "language", "language is required"
	}
	if len(body.Questions) == 0 {
		return "questions", "at least one debrief question is required"
	}
	for i, q := range body.Questions {
		prefix := "questions"
		if i >= 0 {
			prefix = "questions[" + itoa(i) + "]"
		}
		if strings.TrimSpace(q.QuestionText) == "" {
			return prefix + ".questionText", "questionText is required"
		}
		if runeLen(q.QuestionText) > 4000 {
			return prefix + ".questionText", "questionText must be 4000 characters or fewer"
		}
		if strings.TrimSpace(q.MyAnswerSummary) == "" {
			return prefix + ".myAnswerSummary", "myAnswerSummary is required"
		}
		if runeLen(q.MyAnswerSummary) > 4000 {
			return prefix + ".myAnswerSummary", "myAnswerSummary must be 4000 characters or fewer"
		}
		if q.InterviewerReaction != nil && runeLen(*q.InterviewerReaction) > 1000 {
			return prefix + ".interviewerReaction", "interviewerReaction must be 1000 characters or fewer"
		}
	}
	return "", ""
}

func validateSuggestDebriefQuestionsRequest(body api.SuggestDebriefQuestionsRequest, count int32) (field string, message string) {
	if strings.TrimSpace(body.TargetJobId) == "" {
		return "targetJobId", "targetJobId is required"
	}
	if strings.TrimSpace(body.Language) == "" {
		return "language", "language is required"
	}
	if count < 1 || count > 10 {
		return "count", "count must be between 1 and 10"
	}
	return "", ""
}

func suggestQuestionCount(value *int32) int32 {
	if value == nil {
		return 6
	}
	return *value
}

func debriefQuestionsFromAPI(in []api.DebriefQuestionInput) []domain.QuestionInput {
	out := make([]domain.QuestionInput, 0, len(in))
	for _, q := range in {
		out = append(out, domain.QuestionInput{
			QuestionText:        strings.TrimSpace(q.QuestionText),
			MyAnswerSummary:     strings.TrimSpace(q.MyAnswerSummary),
			InterviewerReaction: stringValue(q.InterviewerReaction),
		})
	}
	return out
}

func interviewerRoleValue(value *sharedtypes.InterviewerRole) sharedtypes.InterviewerRole {
	if value == nil {
		return ""
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func runeLen(value string) int {
	return len([]rune(strings.TrimSpace(value)))
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	n := value
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return string(digits[i:])
}

func toAPIDebriefWithJob(result domain.CreateDebriefResult) api.DebriefWithJob {
	job := api.Job{
		Id:           result.Job.ID,
		JobType:      result.Job.JobType,
		ResourceType: result.Job.ResourceType,
		ResourceId:   result.Job.ResourceID,
		Status:       result.Job.Status,
		CreatedAt:    result.Job.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:    result.Job.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}
	if strings.TrimSpace(result.Job.ErrorCode) != "" {
		code := api.ApiErrorCode(result.Job.ErrorCode)
		job.ErrorCode = &code
	}
	return api.DebriefWithJob{DebriefId: result.DebriefID, Job: job}
}

func toAPISuggestDebriefQuestionsResponse(result domain.SuggestQuestionsResult) api.SuggestDebriefQuestionsResponse {
	suggestions := make([]api.SuggestedDebriefQuestion, 0, len(result.Suggestions))
	for _, suggestion := range result.Suggestions {
		item := api.SuggestedDebriefQuestion{
			QuestionText:   suggestion.QuestionText,
			WhyLikelyAsked: suggestion.WhyLikelyAsked,
			Source:         suggestion.Source,
		}
		if strings.TrimSpace(suggestion.Stage) != "" {
			stage := strings.TrimSpace(suggestion.Stage)
			item.Stage = &stage
		}
		suggestions = append(suggestions, item)
	}
	return api.SuggestDebriefQuestionsResponse{Suggestions: suggestions}
}

func toAPIDebrief(record domain.DebriefRecord) api.Debrief {
	out := api.Debrief{
		Id:          record.ID,
		TargetJobId: record.TargetJobID,
		Status:      record.Status,
		RoundType:   record.RoundType,
		CreatedAt:   record.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:   record.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}
	if strings.TrimSpace(record.InterviewerRole) != "" {
		role := sharedtypes.InterviewerRole(record.InterviewerRole)
		out.InterviewerRole = &role
	}
	out.Questions = toAPIDebriefQuestions(record.Questions)
	out.RiskItems = toAPIDebriefRiskItems(record.RiskItems)
	out.NextRoundChecklist = toAPIDebriefChecklist(record.NextRoundChecklist)
	if strings.TrimSpace(record.ThankYouDraft) != "" {
		thankYouDraft := strings.TrimSpace(record.ThankYouDraft)
		out.ThankYouDraft = &thankYouDraft
	}
	if record.Provenance != nil {
		out.Provenance = &api.GenerationProvenance{
			PromptVersion:     record.Provenance.PromptVersion,
			RubricVersion:     record.Provenance.RubricVersion,
			ModelId:           record.Provenance.ModelID,
			Language:          record.Provenance.Language,
			FeatureFlag:       record.Provenance.FeatureFlag,
			DataSourceVersion: record.Provenance.DataSourceVersion,
		}
	}
	return out
}

func toAPIDebriefQuestions(in []domain.QuestionRecord) []api.DebriefQuestion {
	out := make([]api.DebriefQuestion, 0, len(in))
	for _, question := range in {
		item := api.DebriefQuestion{
			QuestionText:    question.QuestionText,
			MyAnswerSummary: question.MyAnswerSummary,
		}
		if strings.TrimSpace(question.InterviewerReaction) != "" {
			interviewerReaction := question.InterviewerReaction
			item.InterviewerReaction = &interviewerReaction
		}
		if strings.TrimSpace(question.AIAnalysis) != "" {
			aiAnalysis := question.AIAnalysis
			item.AiAnalysis = &aiAnalysis
		}
		out = append(out, item)
	}
	return out
}

func toAPIDebriefRiskItems(in []domain.RiskItem) []api.DebriefRiskItem {
	out := make([]api.DebriefRiskItem, 0, len(in))
	for _, risk := range in {
		out = append(out, api.DebriefRiskItem{Label: risk.Label, Severity: risk.Severity})
	}
	return out
}

func toAPIDebriefChecklist(in []domain.NextRoundChecklistItem) []api.DebriefNextRoundChecklistItem {
	out := make([]api.DebriefNextRoundChecklistItem, 0, len(in))
	for _, item := range in {
		apiItem := api.DebriefNextRoundChecklistItem{Label: item.Label}
		if strings.TrimSpace(item.Rationale) != "" {
			rationale := item.Rationale
			apiItem.Rationale = &rationale
		}
		out = append(out, apiItem)
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	raw, err := json.Marshal(body)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "response encoding failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func writeServiceError(w http.ResponseWriter, err error) {
	var svcErr *domain.ServiceError
	if stderrs.As(err, &svcErr) {
		writeServiceAPIError(w, svcErr)
		return
	}
	switch {
	case stderrs.Is(err, domain.ErrDebriefNotFound):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeDebriefNotFound, "debrief not found or not accessible", nil)
	case stderrs.Is(err, domain.ErrDebriefPrerequisite):
		writeAPIError(w, http.StatusNotFound, sharederrors.CodeTargetJobNotFound, "target job not found or not accessible", nil)
	case stderrs.Is(err, domain.ErrDebriefValidation):
		writeAPIError(w, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed, "debrief request is invalid", nil)
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "debrief request failed", nil)
	}
}

func writeServiceAPIError(w http.ResponseWriter, err *domain.ServiceError) {
	status := http.StatusInternalServerError
	switch err.Code {
	case sharederrors.CodeAiFallbackExhausted:
		status = http.StatusServiceUnavailable
	case sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiUnsupportedCapability:
		status = http.StatusBadGateway
	}
	writeAPIError(w, status, err.Code, err.Message, nil)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(api.ApiErrorResponse{
		Error: api.ApiError{
			Code:      code,
			Message:   message,
			RequestID: "",
			Retryable: meta.Retryable,
			Details:   details,
		},
	})
	_, _ = w.Write(raw)
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
