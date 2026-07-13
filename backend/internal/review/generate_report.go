package review

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	reportPayloadByteLimit      = 48_000
	reportContextStartMarker    = "<untrusted_report_context_json>"
	reportContextEndMarker      = "</untrusted_report_context_json>"
	reportMaxCallsPerAction     = 4
)

var reportActionRetryDelays = [...]time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}

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
	if len(framed) > reportPayloadByteLimit {
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
	ordered := append([]MessageSnapshot(nil), reportCtx.Messages...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].SeqNo < ordered[j].SeqNo })
	frozenRaw, err := json.Marshal(reportCtx.FrozenContext)
	if err != nil {
		return aiclient.CompletePayload{}, fmt.Errorf("marshal frozen report context: %w", err)
	}
	messagesRaw, err := json.Marshal(ordered)
	if err != nil {
		return aiclient.CompletePayload{}, fmt.Errorf("marshal report messages: %w", err)
	}
	messages, err := buildReportPromptMessages(resolution.UserMessageTemplate, language, frozenRaw, messagesRaw, repairIssues)
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

// BuildReportPromptMessages is the single report.generate trust-boundary
// serializer shared by the product runtime and the live evaluation harness.
// The caller supplies private full-content valid JSON data. This function
// never promotes any part of that data into the trusted system role; callers
// must not log it or persist it outside explicitly content-bearing surfaces.
func BuildReportPromptMessages(template, language string, frozenContext, conversationMessages json.RawMessage) ([]aiclient.Message, error) {
	return buildReportPromptMessages(template, language, frozenContext, conversationMessages, nil)
}

// BuildReportRepairPromptMessages is the narrow live-evaluation handoff for
// the product report repair serializer. It preserves the exact untrusted
// context message and adds only bounded path/code coordinates to the trusted
// policy. Callers must never pass raw model output as a repair coordinate.
func BuildReportRepairPromptMessages(
	template, language string,
	frozenContext, conversationMessages json.RawMessage,
	repairIssues []ReportValidationIssue,
) ([]aiclient.Message, error) {
	if len(repairIssues) == 0 {
		return nil, fmt.Errorf("report.generate repair requires at least one path/code coordinate")
	}
	return buildReportPromptMessages(template, language, frozenContext, conversationMessages, repairIssues)
}

func buildReportPromptMessages(template, language string, frozenContext, conversationMessages json.RawMessage, repairIssues []ReportValidationIssue) ([]aiclient.Message, error) {
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
		repairRaw, err := json.Marshal(repairIssueCoordinates(repairIssues))
		if err != nil {
			return nil, fmt.Errorf("marshal repair coordinates: %w", err)
		}
		system += "\n\nThe previous attempt violated the output contract. Generate the complete report again from the same untrusted context. Do not repeat these path/code violations: " + string(repairRaw)
		for _, issue := range repairIssues {
			if issue.Code == "missing_evidence" {
				system += " Every dimensionAssessment must be referenced by at least one highlight or issue using the exact same dimensionCode. If a dimension cannot be supported by such an evidence item, remove that dimension instead of inventing evidence."
				break
			}
			if issue.Code == "output_schema_invalid" {
				system += " Recheck every required field against the rendered output contract, including type, enum, string-length, and array-item bounds." + reportActionLabelRepairGuidance(language) + " Return a fully compliant report."
				break
			}
			if issue.Code == "max_length" {
				system += " The field at the supplied path exceeds its maximum length; shorten it to the bound in the rendered output contract."
				if strings.Contains(issue.Path, "nextActions") {
					system += reportActionLabelRepairGuidance(language)
				}
				break
			}
			if issue.Code == "max_words" {
				system += fmt.Sprintf(" For the English action label at the supplied path, count words with whitespace delimiters and rewrite it to at most %d words while preserving the same supported action meaning.", reportActionLabelEnglishWordLimit)
				break
			}
			if issue.Code == "max_code_points" {
				system += fmt.Sprintf(" For the zh-CN action label at the supplied path, count Unicode code points and rewrite it to at most %d characters while preserving the same supported action meaning.", reportActionLabelChineseRuneLimit)
				break
			}
		}
	}
	user := strings.ReplaceAll(userTemplate, "{{frozen_context}}", string(frozenContext))
	user = strings.ReplaceAll(user, "{{conversation_messages}}", string(conversationMessages))
	return []aiclient.Message{{Role: "system", Content: strings.TrimSpace(system)}, {Role: "user", Content: strings.TrimSpace(user)}}, nil
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
