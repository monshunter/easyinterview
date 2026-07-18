package review

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/outputschema"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	reportGenerateFeatureKey    = string(featurekeys.ReportGenerate)
	reportGeneratePromptVersion = "v0.2.0"
	reportGenerateRubricVersion = "v0.2.0"
	reportContextStartMarker    = "<untrusted_report_context_json>"
	reportContextEndMarker      = "</untrusted_report_context_json>"
	reportMaxCallsPerAction     = 4
)

var reportActionRetryDelays = [...]time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}

var reportRepairCoordinatePathPattern = regexp.MustCompile(`^\$(?:(?:\.[A-Za-z][A-Za-z0-9_]*)|(?:\[[0-9]+\]))*$`)

var (
	ErrReviewAIOutputInvalid         = errors.New("review: ai output invalid")
	ErrReportContextTooLarge         = errors.New("review: report context too large")
	ErrReportGenerationConfigInvalid = errors.New("review: report generation configuration invalid")
	ErrReportJobLeaseInvalid         = errors.New("review: report job lease invalid")
)

type SessionSnapshot struct {
	UserID, ReportID, SessionID, TargetJobID, Language string
}

type MessageSnapshot struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	SeqNo   int    `json:"seqNo"`
}

type ReportEvidenceDraft struct {
	DimensionCode       string                 `json:"dimensionCode"`
	Evidence            string                 `json:"evidence"`
	Confidence          sharedtypes.Confidence `json:"confidence"`
	SourceMessageSeqNos []int32                `json:"sourceMessageSeqNos"`
}

type ReportNextActionDraft struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type DimensionAssessmentDraft struct {
	Code       string                      `json:"code"`
	Label      string                      `json:"label"`
	Status     sharedtypes.DimensionStatus `json:"status"`
	Confidence sharedtypes.Confidence      `json:"confidence"`
}

type ReportContentDraft struct {
	Summary                  string                     `json:"summary"`
	PreparednessLevel        sharedtypes.ReadinessTier  `json:"preparednessLevel"`
	DimensionAssessments     []DimensionAssessmentDraft `json:"dimensionAssessments"`
	Highlights               []ReportEvidenceDraft      `json:"highlights"`
	Issues                   []ReportEvidenceDraft      `json:"issues"`
	NextActions              []ReportNextActionDraft    `json:"nextActions"`
	RetryFocusDimensionCodes []string                   `json:"retryFocusDimensionCodes"`
}

type ReportGenerationResult struct {
	Content    ReportContentDraft
	Resolution registry.PromptResolution
	Meta       aiclient.AICallMeta
}

type ReportContentInvalidError struct {
	Issues []ReportValidationIssue
}

func (e *ReportContentInvalidError) Error() string {
	if e == nil || len(e.Issues) == 0 {
		return ErrReviewAIOutputInvalid.Error()
	}
	parts := make([]string, 0, len(e.Issues))
	for _, issue := range e.Issues {
		parts = append(parts, issue.Path+":"+issue.Code)
	}
	return ErrReviewAIOutputInvalid.Error() + ": " + strings.Join(parts, ",")
}

func (e *ReportContentInvalidError) Unwrap() error { return ErrReviewAIOutputInvalid }

func (s *Service) prepareReportGeneration(ctx context.Context, reportCtx ReportContext, repairIssues []ReportValidationIssue) (registry.PromptResolution, aiclient.CompletePayload, error) {
	if s == nil || s.registry == nil || s.ai == nil {
		return registry.PromptResolution{}, aiclient.CompletePayload{}, fmt.Errorf("review generation is not configured")
	}
	language, err := canonicalReportLanguage(reportCtx.Session.Language)
	if err != nil {
		return registry.PromptResolution{}, aiclient.CompletePayload{}, err
	}
	resolution, err := s.registry.ResolveActive(ctx, reportGenerateFeatureKey, language)
	if err != nil {
		return registry.PromptResolution{}, aiclient.CompletePayload{}, fmt.Errorf("resolve report.generate: %w", err)
	}
	if err := validateReportPromptResolution(resolution); err != nil {
		return resolution, aiclient.CompletePayload{}, err
	}
	if reportCtx.FrozenContext.SchemaVersion != practicedomain.ReportContextSchemaVersion {
		return resolution, aiclient.CompletePayload{}, fmt.Errorf("%w: report generation schema mismatch", ErrReportContextInvalid)
	}
	payload, err := reportCompletePayload(resolution, reportCtx, repairIssues)
	if err != nil {
		return resolution, aiclient.CompletePayload{}, err
	}
	framed, err := frameReportMessages(payload.Messages)
	if err != nil {
		return resolution, payload, err
	}
	if int64(len(framed)) > s.maxFramedInputBytes {
		return resolution, payload, fmt.Errorf("%w: framed payload is %d bytes", ErrReportContextTooLarge, len(framed))
	}
	return resolution, payload, nil
}

func (s *Service) completeReportGeneration(ctx context.Context, reportCtx ReportContext, resolution registry.PromptResolution, payload aiclient.CompletePayload) (ReportGenerationResult, error) {
	result := ReportGenerationResult{Resolution: resolution}
	response, meta, err := s.ai.Complete(ctx, resolution.ModelProfileName, payload)
	result.Meta = meta
	if err != nil {
		if isTypedReportOutputInvalid(err) && strings.TrimSpace(response.FinishReason) == "stop" && validateReportInvalidCallMeta(meta, resolution, reportCtx.Session.Language) == nil {
			if content, issues, targeted := runtimeActionLabelRepairCandidate(resolution, response.Content, reportCtx); targeted {
				result.Content = content
				return result, &ReportContentInvalidError{Issues: issues}
			}
			return result, &ReportContentInvalidError{Issues: []ReportValidationIssue{OutputSchemaRepairIssue(err)}}
		}
		return result, fmt.Errorf("complete report.generate: %w", err)
	}
	if strings.TrimSpace(response.FinishReason) != "stop" {
		return result, &ReportContentInvalidError{Issues: []ReportValidationIssue{{Path: "$", Code: "finish_reason_not_stop"}}}
	}
	if err := validateReportCallMeta(meta, resolution, reportCtx.Session.Language); err != nil {
		return result, err
	}
	if resolution.OutputSchema == nil {
		return result, fmt.Errorf("%w: report output schema is missing", ErrReportGenerationConfigInvalid)
	}
	if err := outputschema.Validate(*resolution.OutputSchema, response.Content); err != nil {
		if content, issues, targeted := runtimeActionLabelRepairCandidate(resolution, response.Content, reportCtx); targeted {
			result.Content = content
			return result, &ReportContentInvalidError{Issues: issues}
		}
		return result, &ReportContentInvalidError{Issues: []ReportValidationIssue{OutputSchemaRepairIssue(err)}}
	}
	draft, issues := decodeReportContent([]byte(response.Content))
	if len(issues) == 0 {
		result.Content = draft
		issues = validateReportContent(draft, reportCtx.FrozenContext, reportCtx.Messages)
	}
	if len(issues) > 0 {
		return result, &ReportContentInvalidError{Issues: issues}
	}
	result.Content = draft
	return result, nil
}

func (s *Service) generateReportContent(ctx context.Context, reportCtx ReportContext, repairIssues []ReportValidationIssue) (ReportGenerationResult, error) {
	resolution, payload, err := s.prepareReportGeneration(ctx, reportCtx, repairIssues)
	if err != nil {
		return ReportGenerationResult{Resolution: resolution}, err
	}
	return s.completeReportGeneration(ctx, reportCtx, resolution, payload)
}

func decodeReportContent(raw []byte) (ReportContentDraft, []ReportValidationIssue) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return ReportContentDraft{}, []ReportValidationIssue{{Path: "$", Code: "empty_json"}}
	}
	var top map[string]json.RawMessage
	if err := decodeClosedJSONValue(trimmed, &top); err != nil {
		return ReportContentDraft{}, []ReportValidationIssue{{Path: "$", Code: "invalid_json"}}
	}
	required := []string{"summary", "preparednessLevel", "dimensionAssessments", "highlights", "issues", "nextActions", "retryFocusDimensionCodes"}
	for _, key := range required {
		value, ok := top[key]
		if !ok {
			return ReportContentDraft{}, []ReportValidationIssue{{Path: "$." + key, Code: "required"}}
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ReportContentDraft{}, []ReportValidationIssue{{Path: "$." + key, Code: "invalid_type"}}
		}
	}
	var draft ReportContentDraft
	if err := decodeClosedJSONValue(trimmed, &draft); err != nil {
		return ReportContentDraft{}, []ReportValidationIssue{{Path: "$", Code: "closed_schema_invalid"}}
	}
	draft.trim()
	return draft, nil
}

func decodeClosedJSONValue(raw []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func (d *ReportContentDraft) trim() {
	d.Summary = strings.TrimSpace(d.Summary)
	for index := range d.DimensionAssessments {
		d.DimensionAssessments[index].Code = strings.TrimSpace(d.DimensionAssessments[index].Code)
		d.DimensionAssessments[index].Label = strings.TrimSpace(d.DimensionAssessments[index].Label)
	}
	for _, items := range [][]ReportEvidenceDraft{d.Highlights, d.Issues} {
		for index := range items {
			items[index].DimensionCode = strings.TrimSpace(items[index].DimensionCode)
			items[index].Evidence = strings.TrimSpace(items[index].Evidence)
		}
	}
	for index := range d.NextActions {
		d.NextActions[index].Type = strings.TrimSpace(d.NextActions[index].Type)
		d.NextActions[index].Label = strings.TrimSpace(d.NextActions[index].Label)
	}
	for index := range d.RetryFocusDimensionCodes {
		d.RetryFocusDimensionCodes[index] = strings.TrimSpace(d.RetryFocusDimensionCodes[index])
	}
}

func validateReportPromptResolution(resolution registry.PromptResolution) error {
	if resolution.FeatureKey != reportGenerateFeatureKey || resolution.PromptVersion != reportGeneratePromptVersion || resolution.RubricVersion != reportGenerateRubricVersion {
		return fmt.Errorf("%w: active coordinate mismatch", ErrReportGenerationConfigInvalid)
	}
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		return fmt.Errorf("%w: v0.2 trust boundary must be split by the report consumer", ErrReportGenerationConfigInvalid)
	}
	if strings.TrimSpace(resolution.ModelProfileName) == "" || strings.TrimSpace(resolution.FeatureFlag) == "" || resolution.DataSourceVersion != practicedomain.ReportContextSchemaVersion || resolution.OutputSchema == nil {
		return fmt.Errorf("%w: resolution provenance is incomplete", ErrReportGenerationConfigInvalid)
	}
	return nil
}

func validateReportCallMeta(meta aiclient.AICallMeta, resolution registry.PromptResolution, rawLanguage string) error {
	if err := validateReportCallProvenance(meta, resolution, rawLanguage); err != nil {
		return err
	}
	if meta.ValidationStatus != aiclient.ValidationStatusOK || strings.TrimSpace(meta.ErrorCode) != "" {
		return fmt.Errorf("%w: call validation provenance is invalid", ErrReportGenerationConfigInvalid)
	}
	return nil
}

func validateReportInvalidCallMeta(meta aiclient.AICallMeta, resolution registry.PromptResolution, rawLanguage string) error {
	if err := validateReportCallProvenance(meta, resolution, rawLanguage); err != nil {
		return err
	}
	if meta.ValidationStatus != aiclient.ValidationStatusInvalid || meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
		return fmt.Errorf("%w: invalid call validation provenance is inconsistent", ErrReportGenerationConfigInvalid)
	}
	return nil
}

func validateReportCallProvenance(meta aiclient.AICallMeta, resolution registry.PromptResolution, rawLanguage string) error {
	language, err := canonicalReportLanguage(rawLanguage)
	if err != nil {
		return err
	}
	if strings.TrimSpace(meta.Provider) == "" || strings.TrimSpace(meta.ModelID) == "" {
		return fmt.Errorf("%w: call provenance is incomplete", ErrReportGenerationConfigInvalid)
	}
	if meta.InputTokens <= 0 || meta.OutputTokens <= 0 {
		return fmt.Errorf("%w: call token provenance is invalid", ErrReportGenerationConfigInvalid)
	}
	if meta.FeatureKey != resolution.FeatureKey || meta.PromptVersion != resolution.PromptVersion || meta.RubricVersion != resolution.RubricVersion || meta.ModelProfileName != resolution.ModelProfileName || meta.Language != language || meta.FeatureFlag != resolution.FeatureFlag || meta.DataSourceVersion != resolution.DataSourceVersion {
		return fmt.Errorf("%w: call provenance mismatch", ErrReportGenerationConfigInvalid)
	}
	return nil
}

func isTypedReportOutputInvalid(err error) bool {
	var apiErr *sharederrors.APIError
	return errors.As(err, &apiErr) && apiErr.Code == sharederrors.CodeAiOutputInvalid && !apiErr.Retryable
}

func runtimeActionLabelRepairCandidate(
	resolution registry.PromptResolution,
	responseContent string,
	reportCtx ReportContext,
) (ReportContentDraft, []ReportValidationIssue, bool) {
	if resolution.OutputSchema == nil {
		return ReportContentDraft{}, nil, false
	}
	return DetectReportActionLabelRepairCandidateWithContext(
		*resolution.OutputSchema,
		responseContent,
		reportCtx.Session.Language,
		ReportContentValidationContext{
			Language:         reportCtx.FrozenContext.Conversation.Language,
			HasNextRound:     reportCtx.FrozenContext.HasNextRound,
			LastMessageSeqNo: reportCtx.FrozenContext.Conversation.LastMessageSeqNo,
		},
		reportCtx.Messages,
	)
}

func isActionLabelLimitCode(language, code string) bool {
	if code == "max_length" {
		return true
	}
	return (language == "en" && code == "max_words") || (language == "zh-CN" && code == "max_code_points")
}

func reportCompletePayload(resolution registry.PromptResolution, reportCtx ReportContext, repairIssues []ReportValidationIssue) (aiclient.CompletePayload, error) {
	language, err := canonicalReportLanguage(reportCtx.Session.Language)
	if err != nil {
		return aiclient.CompletePayload{}, err
	}
	assessmentMessages := reportAssessmentMessages(reportCtx.Messages)
	frozenRaw, err := json.Marshal(reportCtx.FrozenContext)
	if err != nil {
		return aiclient.CompletePayload{}, fmt.Errorf("marshal frozen report context: %w", err)
	}
	messagesRaw, err := json.Marshal(assessmentMessages)
	if err != nil {
		return aiclient.CompletePayload{}, fmt.Errorf("marshal report messages: %w", err)
	}
	messages, err := buildReportPromptMessages(
		resolution.UserMessageTemplate,
		language,
		frozenRaw,
		messagesRaw,
		repairIssues,
		candidateUserMessageSeqNos(assessmentMessages),
	)
	if err != nil {
		return aiclient.CompletePayload{}, err
	}
	payload := aiclient.CompletePayload{
		Messages: messages,
		Metadata: aiclient.CallMetadata{
			FeatureKey:        reportGenerateFeatureKey,
			PromptVersion:     resolution.PromptVersion,
			RubricVersion:     resolution.RubricVersion,
			Language:          language,
			FeatureFlag:       resolution.FeatureFlag,
			DataSourceVersion: resolution.DataSourceVersion,
			OutputSchema:      append(json.RawMessage(nil), (*resolution.OutputSchema)...),
			TaskRun: aiclient.AITaskRunContext{
				UserID: reportCtx.Session.UserID, Capability: aiclient.AITaskRunTaskReportGenerate,
				ResourceType: aiclient.AITaskRunResourceFeedbackReport, ResourceID: reportCtx.Session.ReportID,
			},
		},
	}
	return payload, nil
}

func reportAssessmentMessages(messages []MessageSnapshot) []MessageSnapshot {
	ordered := append([]MessageSnapshot(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SeqNo < ordered[j].SeqNo })
	if len(ordered) > 0 && ordered[len(ordered)-1].Role == "assistant" {
		ordered = ordered[:len(ordered)-1]
	}
	return ordered
}

// BuildReportPromptMessages is the single report.generate trust-boundary
// serializer shared by the product runtime and the live evaluation harness.
// The caller supplies private full-content valid JSON data. This function
// never promotes any part of that data into the trusted system role; callers
// must not log it or persist it outside explicitly content-bearing surfaces.
func BuildReportPromptMessages(template, language string, frozenContext, conversationMessages json.RawMessage) ([]aiclient.Message, error) {
	return buildReportPromptMessages(template, language, frozenContext, conversationMessages, nil, nil)
}

// BuildReportRepairPromptMessages is the narrow live-evaluation handoff for
// the product report repair serializer. It preserves the exact untrusted
// context message and adds bounded path/code coordinates plus, when needed,
// trusted candidate-user message sequence numbers to the policy. Callers must
// never pass raw model output or unvalidated role claims as repair coordinates.
func BuildReportRepairPromptMessages(
	template, language string,
	frozenContext, conversationMessages json.RawMessage,
	repairIssues []ReportValidationIssue,
	candidateUserSeqNos []int,
) ([]aiclient.Message, error) {
	if len(repairIssues) == 0 {
		return nil, fmt.Errorf("report.generate repair requires at least one path/code coordinate")
	}
	return buildReportPromptMessages(template, language, frozenContext, conversationMessages, repairIssues, candidateUserSeqNos)
}

func buildReportPromptMessages(
	template, language string,
	frozenContext, conversationMessages json.RawMessage,
	repairIssues []ReportValidationIssue,
	candidateUserSeqNos []int,
) ([]aiclient.Message, error) {
	language, err := canonicalReportLanguage(language)
	if err != nil {
		return nil, err
	}
	if !json.Valid(frozenContext) || !json.Valid(conversationMessages) {
		return nil, fmt.Errorf("report.generate untrusted context must be valid JSON")
	}
	systemTemplate, userTemplate, err := splitReportPromptTemplate(template)
	if err != nil {
		return nil, err
	}
	if err := validateReportPromptTokens(systemTemplate, userTemplate); err != nil {
		return nil, err
	}
	system := strings.ReplaceAll(systemTemplate, "{{language}}", language)
	if len(repairIssues) > 0 {
		if err := validateReportRepairInputs(repairIssues, candidateUserSeqNos); err != nil {
			return nil, err
		}
		repairRaw, err := json.Marshal(repairIssueCoordinates(repairIssues))
		if err != nil {
			return nil, fmt.Errorf("marshal repair coordinates: %w", err)
		}
		system += "\n\nThe previous attempt violated the output contract. Generate the complete report again from the same untrusted context. Do not repeat these path/code violations: " + string(repairRaw)
		intent, err := buildReportRepairIntent(language, repairIssues, candidateUserSeqNos)
		if err != nil {
			return nil, err
		}
		system += " " + intent
	}
	user := strings.ReplaceAll(userTemplate, "{{frozen_context}}", escapeUntrustedJSONForPrompt(frozenContext))
	user = strings.ReplaceAll(user, "{{conversation_messages}}", escapeUntrustedJSONForPrompt(conversationMessages))
	return []aiclient.Message{{Role: "system", Content: strings.TrimSpace(system)}, {Role: "user", Content: strings.TrimSpace(user)}}, nil
}

func escapeUntrustedJSONForPrompt(raw json.RawMessage) string {
	return strings.NewReplacer("&", `\u0026`, "<", `\u003c`, ">", `\u003e`).Replace(string(raw))
}

func candidateUserMessageSeqNos(messages []MessageSnapshot) []int {
	seqNos := make([]int, 0, len(messages))
	for _, message := range messages {
		if message.Role == "user" && message.SeqNo > 0 {
			seqNos = append(seqNos, message.SeqNo)
		}
	}
	sort.Ints(seqNos)
	return seqNos
}

func validateReportRepairInputs(issues []ReportValidationIssue, candidateUserSeqNos []int) error {
	hasAnchorIssue := false
	for _, issue := range issues {
		if len(issue.Path) > 256 || !reportRepairCoordinatePathPattern.MatchString(issue.Path) {
			return fmt.Errorf("%w: unsafe repair coordinate path", ErrReviewAIOutputInvalid)
		}
		family, err := reportRepairFamilyForIssue(issue)
		if err != nil {
			return err
		}
		switch family {
		case reportRepairFamilyAnchors:
			hasAnchorIssue = true
			if !strings.Contains(issue.Path, ".sourceMessageSeqNos") {
				return fmt.Errorf("%w: anchor repair code has incompatible path", ErrReviewAIOutputInvalid)
			}
		case reportRepairFamilyEvidence:
			if !hasAnyReportRepairPathPrefix(issue.Path, "$.dimensionAssessments", "$.highlights", "$.issues") {
				return fmt.Errorf("%w: evidence repair code has incompatible path", ErrReviewAIOutputInvalid)
			}
		case reportRepairFamilyReadiness:
			if issue.Path != "$.preparednessLevel" && !hasAnyReportRepairPathPrefix(issue.Path, "$.dimensionAssessments", "$.highlights", "$.issues") {
				return fmt.Errorf("%w: readiness repair code has incompatible path", ErrReviewAIOutputInvalid)
			}
		case reportRepairFamilyActions:
			if !hasAnyReportRepairPathPrefix(issue.Path, "$.nextActions", "$.retryFocusDimensionCodes") {
				return fmt.Errorf("%w: action repair code has incompatible path", ErrReviewAIOutputInvalid)
			}
		case reportRepairFamilyText:
			if issue.Path != "$.language" && issue.Path != "$.summary" && !hasAnyReportRepairPathPrefix(issue.Path, "$.dimensionAssessments", "$.highlights", "$.issues", "$.nextActions") {
				return fmt.Errorf("%w: text repair code has incompatible path", ErrReviewAIOutputInvalid)
			}
		}
	}
	previous := 0
	for index, seqNo := range candidateUserSeqNos {
		if seqNo <= 0 || (index > 0 && seqNo <= previous) {
			return fmt.Errorf("%w: candidate user message sequence numbers must be positive, unique and ascending", ErrReviewAIOutputInvalid)
		}
		previous = seqNo
	}
	if hasAnchorIssue && len(candidateUserSeqNos) == 0 {
		return fmt.Errorf("%w: anchor repair requires at least one validated candidate user message sequence number", ErrReviewAIOutputInvalid)
	}
	return nil
}

func hasAnyReportRepairPathPrefix(path string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"[") || strings.HasPrefix(path, prefix+".") {
			return true
		}
	}
	return false
}

type reportRepairFamily string

const (
	reportRepairFamilyStructural reportRepairFamily = "structural"
	reportRepairFamilyAnchors    reportRepairFamily = "anchors"
	reportRepairFamilyEvidence   reportRepairFamily = "evidence"
	reportRepairFamilyReadiness  reportRepairFamily = "readiness"
	reportRepairFamilyActions    reportRepairFamily = "actions"
	reportRepairFamilyText       reportRepairFamily = "text"
)

func buildReportRepairIntent(language string, issues []ReportValidationIssue, candidateUserSeqNos []int) (string, error) {
	families := make(map[reportRepairFamily]struct{}, len(issues))
	for _, issue := range issues {
		family, err := reportRepairFamilyForIssue(issue)
		if err != nil {
			return "", err
		}
		families[family] = struct{}{}
	}
	intents := make([]string, 0, len(families))
	for _, family := range []reportRepairFamily{
		reportRepairFamilyStructural,
		reportRepairFamilyAnchors,
		reportRepairFamilyEvidence,
		reportRepairFamilyReadiness,
		reportRepairFamilyActions,
		reportRepairFamilyText,
	} {
		if _, present := families[family]; !present {
			continue
		}
		switch family {
		case reportRepairFamilyStructural:
			intent := "Return exactly one complete, stop-terminated JSON object that satisfies the rendered closed output contract at every listed path: include every required non-blank field, use the required types, enums and formats, remove duplicates and unknown fields, and do not return a patch or a targeted-repair envelope. Produce 1 to 6 dimensionAssessments, at most 4 highlights, at most 4 issues, 1 to 6 combined highlight/issue evidence items, 1 or 2 nextActions, and at most 6 retryFocusDimensionCodes. Dimension and focus codes must match ^[a-z][a-z0-9_]{1,63}$."
			if hasReportValidationCode(issues, "output_schema_invalid") {
				intent += " Recheck every required field against the rendered output contract, including type, enum, string-length, and array-item bounds." + reportActionLabelRepairGuidance(language)
			}
			if hasReportValidationCode(issues, "max_length") {
				intent += " The field at the supplied path exceeds its maximum length; shorten it to the bound in the rendered output contract."
			}
			intents = append(intents, intent)
		case reportRepairFamilyAnchors:
			allowedRaw, err := json.Marshal(candidateUserSeqNos)
			if err != nil {
				return "", fmt.Errorf("marshal candidate user message sequence numbers: %w", err)
			}
			intents = append(intents, "Every highlight and issue must contain at least one evidence anchor. Evidence anchors must be positive, unique, ascending, within the frozen transcript, and must reference candidate user messages rather than assistant messages. Valid candidate user message sequence numbers for evidence anchors: "+string(allowedRaw)+". Use only supported numbers from this list; remove an unsupported evidence item and its unsupported dimension instead of guessing or deterministically changing an anchor.")
		case reportRepairFamilyEvidence:
			intents = append(intents, "Make every evidence item reference an existing dimensionCode exactly. Every dimensionAssessment must be referenced by at least one highlight or issue using the exact same dimensionCode. A strong dimension needs a highlight, a needs_work dimension needs an issue, and an issue must not target a non-needs_work dimension. If a dimension cannot be supported by such an evidence item, remove that dimension instead of inventing evidence.")
		case reportRepairFamilyReadiness:
			intents = append(intents, "preparednessLevel must be one of not_ready, needs_practice, basically_ready, or well_prepared. Every dimension status must be strong, meets_bar, or needs_work, and every confidence must be high, medium, or low. A basically_ready report cannot contain any needs_work dimension. It must contain at least one meets_bar or strong dimension with medium or high confidence. Choose preparedness from supported candidate evidence: if a material gap exists or no supported usable dimension exists, use a lower tier, keep an issue for the same needs_work dimension, and make retry_current_round the first action; otherwise remove unsupported needs_work dimensions and issues. A well_prepared report requires at least two strong, high-confidence dimensions, no issues, and highlights grounded in at least two candidate user messages. Low confidence cannot support strong, needs_work, or issue claims. Do not change the tier mechanically just to pass validation.")
		case reportRepairFamilyActions:
			intents = append(intents, "Return 1 or 2 nextActions with unique types chosen only from retry_current_round, next_round, and review_evidence. Make nextActions and retryFocusDimensionCodes consistent with readiness and hasNextRound: lower tiers start with retry_current_round; do not use next_round when unavailable or with a lower tier; include retry focus only with a retry action and make it the unique ascending exact set of issue-backed needs_work dimension codes, except the documented single generic answer-depth/relevance case requires an empty focus list.")
		case reportRepairFamilyText:
			intent := "Rewrite text at the listed paths in " + language + " using concise original wording. Do not copy transcript text or make forbidden hiring-probability, ranking, speech, emotion, or personality claims."
			switch {
			case hasReportValidationCode(issues, "max_words"):
				intent += fmt.Sprintf(" For the English action label at the supplied path, count words with whitespace delimiters and rewrite it to at most %d words while preserving the same supported action meaning.", reportActionLabelEnglishWordLimit)
			case hasReportValidationCode(issues, "max_code_points"):
				intent += fmt.Sprintf(" For the zh-CN action label at the supplied path, count Unicode code points and rewrite it to at most %d characters while preserving the same supported action meaning.", reportActionLabelChineseRuneLimit)
			case hasActionLabelLimitIssue(language, issues):
				intent += reportActionLabelRepairGuidance(language)
			}
			intents = append(intents, intent)
		}
	}
	return strings.Join(intents, " "), nil
}

func reportRepairFamilyForIssue(issue ReportValidationIssue) (reportRepairFamily, error) {
	anchorPath := strings.Contains(issue.Path, ".sourceMessageSeqNos")
	focusPath := strings.HasPrefix(issue.Path, "$.retryFocusDimensionCodes")
	actionPath := strings.HasPrefix(issue.Path, "$.nextActions")
	readinessPath := issue.Path == "$.preparednessLevel" || strings.HasSuffix(issue.Path, ".status") || strings.HasSuffix(issue.Path, ".confidence")

	switch issue.Code {
	case "not_user_message", "not_positive", "out_of_range":
		return reportRepairFamilyAnchors, nil
	case "not_ascending":
		if anchorPath {
			return reportRepairFamilyAnchors, nil
		}
		if focusPath {
			return reportRepairFamilyActions, nil
		}
		return reportRepairFamilyStructural, nil
	case "duplicate", "min_items":
		if anchorPath {
			return reportRepairFamilyAnchors, nil
		}
		if focusPath || actionPath {
			return reportRepairFamilyActions, nil
		}
		return reportRepairFamilyStructural, nil
	case "invalid_enum":
		if actionPath {
			return reportRepairFamilyActions, nil
		}
		if readinessPath {
			return reportRepairFamilyReadiness, nil
		}
		return reportRepairFamilyStructural, nil
	case "max_length":
		if strings.Contains(issue.Path, "nextActions") {
			return reportRepairFamilyText, nil
		}
		return reportRepairFamilyStructural, nil
	case "empty_json", "invalid_json", "closed_schema_invalid", "output_schema_invalid",
		"required", "invalid_type", "unknown_field", "min_length", "max_items",
		"invalid_format", "finish_reason_not_stop", "invalid_utf8", "replacement_set_mismatch":
		return reportRepairFamilyStructural, nil
	case "missing_evidence", "strong_requires_highlight", "needs_work_requires_issue":
		return reportRepairFamilyEvidence, nil
	case "unknown_reference", "not_needs_work", "not_issue_backed":
		if focusPath {
			return reportRepairFamilyActions, nil
		}
		return reportRepairFamilyEvidence, nil
	case "requires_needs_work", "forbids_needs_work", "requires_supported_usable_dimension",
		"low_confidence_status_invalid", "low_confidence_issue_invalid", "well_prepared_min_items",
		"well_prepared_requires_strong", "well_prepared_requires_high_confidence",
		"well_prepared_forbids_issues", "well_prepared_requires_two_user_messages":
		return reportRepairFamilyReadiness, nil
	case "retry_current_round_required", "next_round_inconsistent_with_readiness",
		"next_round_unavailable", "retry_action_required", "retry_focus_mismatch",
		"generic_exception_requires_empty_focus":
		return reportRepairFamilyActions, nil
	case "forbidden_claim", "copied_message_text", "language_mismatch", "max_words",
		"max_code_points", "unsupported_language":
		return reportRepairFamilyText, nil
	default:
		return "", fmt.Errorf("%w: no safe repair intent for validation code %q", ErrReviewAIOutputInvalid, issue.Code)
	}
}

func hasActionLabelLimitIssue(language string, issues []ReportValidationIssue) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Path, "nextActions") && isActionLabelLimitCode(language, issue.Code) {
			return true
		}
	}
	return false
}

func hasReportValidationCode(issues []ReportValidationIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func reportActionLabelRepairGuidance(language string) string {
	schemaFuse := fmt.Sprintf(" Keep every action label within the %d Unicode-character schema safety fuse.", reportActionLabelSchemaRuneLimit)
	switch language {
	case "en":
		return schemaFuse + fmt.Sprintf(" For English, count words with whitespace delimiters and use at most %d words while preserving the same supported action meaning.", reportActionLabelEnglishWordLimit)
	case "zh-CN":
		return schemaFuse + fmt.Sprintf(" For zh-CN, count Unicode code points and use at most %d characters while preserving the same supported action meaning.", reportActionLabelChineseRuneLimit)
	default:
		return schemaFuse
	}
}

func validateReportPromptTokens(systemTemplate, userTemplate string) error {
	if strings.Count(systemTemplate, "{{language}}") != 1 {
		return fmt.Errorf("%w: trusted policy must contain one language token", ErrReportGenerationConfigInvalid)
	}
	trustedRemainder := strings.ReplaceAll(systemTemplate, "{{language}}", "")
	if strings.Contains(trustedRemainder, "{{") {
		return fmt.Errorf("%w: trusted policy contains an unsupported template token", ErrReportGenerationConfigInvalid)
	}
	for _, token := range []string{"{{frozen_context}}", "{{conversation_messages}}"} {
		if strings.Count(userTemplate, token) != 1 {
			return fmt.Errorf("%w: untrusted block must contain one %s token", ErrReportGenerationConfigInvalid, token)
		}
		userTemplate = strings.ReplaceAll(userTemplate, token, "")
	}
	if strings.Contains(userTemplate, "{{") {
		return fmt.Errorf("%w: untrusted block contains an unsupported template token", ErrReportGenerationConfigInvalid)
	}
	return nil
}

func splitReportPromptTemplate(template string) (string, string, error) {
	start, startEnd, startCount := standalonePromptMarker(template, reportContextStartMarker)
	end, endEnd, endCount := standalonePromptMarker(template, reportContextEndMarker)
	if startCount != 1 || endCount != 1 {
		return "", "", fmt.Errorf("%w: prompt must contain one untrusted context block", ErrReportGenerationConfigInvalid)
	}
	if start < 0 || end <= startEnd {
		return "", "", fmt.Errorf("%w: untrusted context block is malformed", ErrReportGenerationConfigInvalid)
	}
	trustedBefore := strings.TrimSpace(template[:start])
	trustedAfter := strings.TrimSpace(template[endEnd:])
	untrusted := strings.TrimSpace(template[start:endEnd])
	if trustedBefore == "" || trustedAfter == "" || untrusted == "" {
		return "", "", fmt.Errorf("%w: trust boundary is incomplete", ErrReportGenerationConfigInvalid)
	}
	system := trustedBefore + "\n\nThe untrusted report context is supplied as a separate user message.\n\n" + trustedAfter
	return system, untrusted, nil
}

func standalonePromptMarker(template, marker string) (start, end, count int) {
	start = -1
	end = -1
	for lineStart := 0; lineStart <= len(template); {
		lineEnd := strings.IndexByte(template[lineStart:], '\n')
		next := len(template)
		if lineEnd >= 0 {
			lineEnd += lineStart
			next = lineEnd + 1
		} else {
			lineEnd = len(template)
		}
		line := strings.TrimSpace(template[lineStart:lineEnd])
		if line == marker {
			count++
			start = lineStart
			end = lineEnd
		}
		if next >= len(template) {
			break
		}
		lineStart = next
	}
	return start, end, count
}

func frameReportMessages(messages []aiclient.Message) ([]byte, error) {
	if len(messages) != 2 || messages[0].Role != "system" || messages[1].Role != "user" {
		return nil, fmt.Errorf("report.generate requires one trusted system and one untrusted user message")
	}
	framed, err := json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("marshal report message frame: %w", err)
	}
	if !utf8.Valid(framed) {
		return nil, fmt.Errorf("report.generate message frame is not valid UTF-8")
	}
	return framed, nil
}

func repairIssueCoordinates(issues []ReportValidationIssue) []map[string]string {
	out := make([]map[string]string, 0, len(issues))
	for _, issue := range issues {
		out = append(out, map[string]string{"path": issue.Path, "code": issue.Code})
	}
	return out
}

func canonicalReportLanguage(language string) (string, error) {
	switch strings.TrimSpace(language) {
	case "en":
		return "en", nil
	case "zh-CN":
		return "zh-CN", nil
	default:
		return "", fmt.Errorf("report.generate session language is unsupported")
	}
}

func fallbackLanguage(language string) string {
	canonical, err := canonicalReportLanguage(language)
	if err != nil {
		return strings.TrimSpace(language)
	}
	return canonical
}
