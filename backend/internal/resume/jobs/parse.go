package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const FeatureKeyResumeParse = string(featurekeys.ResumeParse)

const defaultMaxResumeInputBytes int64 = 256 * 1024

var ErrPromptUnsupported = errors.New("prompt registry: feature/language is not enabled")

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

type PromptRegistryClient interface {
	Resolve(ctx context.Context, featureKey string, language string) (PromptResolution, error)
}

type Store interface {
	GetForParse(ctx context.Context, assetID string) (resumestore.ParseAssetRecord, error)
	MarkParsing(ctx context.Context, in resumestore.StatusUpdateInput) error
	CompleteParseSuccess(ctx context.Context, in resumestore.CompleteParseSuccessInput) error
	CompleteParseFailure(ctx context.Context, in resumestore.CompleteParseFailureInput) error
}

type ObjectReader interface {
	Read(ctx context.Context, objectKey string, maxBytes int64) ([]byte, error)
}

type ParseHandlerOptions struct {
	Store         Store
	Registry      PromptRegistryClient
	AI            aiclient.AIClient
	Objects       ObjectReader
	NewID         func() string
	Now           func() time.Time
	MaxInputBytes int64
}

type ParseHandler struct {
	store         Store
	registry      PromptRegistryClient
	ai            aiclient.AIClient
	objects       ObjectReader
	newID         func() string
	now           func() time.Time
	maxInputBytes int64
}

func NewParseHandler(opts ParseHandlerOptions) *ParseHandler {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	maxInputBytes := opts.MaxInputBytes
	if maxInputBytes <= 0 {
		maxInputBytes = defaultMaxResumeInputBytes
	}
	return &ParseHandler{
		store:         opts.Store,
		registry:      opts.Registry,
		ai:            opts.AI,
		objects:       opts.Objects,
		newID:         opts.NewID,
		now:           now,
		maxInputBytes: maxInputBytes,
	}
}

func (h *ParseHandler) Handle(ctx context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
	if h == nil || h.store == nil || h.registry == nil || h.ai == nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: "resume parse handler not initialised"}
	}
	asset, err := h.store.GetForParse(ctx, job.ResourceID)
	if err != nil {
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error())}
	}
	switch asset.ParseStatus {
	case sharedtypes.TargetJobParseStatusQueued, sharedtypes.TargetJobParseStatusFailed:
		if err := h.store.MarkParsing(ctx, resumestore.StatusUpdateInput{UserID: asset.UserID, AssetID: asset.ID, Now: h.now()}); err != nil {
			return targetjob.JobOutcome{ErrorCode: sharederrors.CodeTargetInvalidStateTransition, ErrorMessage: safeFailureMessage(sharederrors.CodeTargetInvalidStateTransition, err.Error())}
		}
	case sharedtypes.TargetJobParseStatusProcessing:
	case sharedtypes.TargetJobParseStatusReady:
		return targetjob.JobOutcome{Succeeded: true}
	default:
		return targetjob.JobOutcome{ErrorCode: sharederrors.CodeTargetInvalidStateTransition, ErrorMessage: sharederrors.CodeTargetInvalidStateTransition}
	}

	input, err := h.resumeInput(ctx, asset)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeValidationFailed, err.Error(), false)
	}
	resolution, err := h.registry.Resolve(ctx, FeatureKeyResumeParse, asset.Language)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeAiProviderConfigInvalid, err.Error(), false)
	}
	metadata := aiclient.CallMetadata{
		FeatureKey:        FeatureKeyResumeParse,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          asset.Language,
		FeatureFlag:       coalesceFlag(resolution.FeatureFlag),
		DataSourceVersion: resolution.DataSourceVersion,
		TaskRun: aiclient.AITaskRunContext{
			UserID:              asset.UserID,
			Capability:          aiclient.AITaskRunTaskResumeParse,
			ResourceType:        aiclient.AITaskRunResourceResumeAsset,
			ResourceID:          asset.ID,
			OutputSchemaVersion: "resume.parse.v1",
		},
	}
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	complete, _, err := h.ai.Complete(ctx, resolution.ModelProfileName, aiclient.CompletePayload{
		Messages: buildPromptMessages(resolution, input),
		Metadata: metadata,
	})
	if err != nil {
		code, retryable := translateAIClientError(err)
		return h.fail(ctx, asset, job, code, err.Error(), retryable)
	}
	parsed, displayName, err := decodeResumeParseResponse(complete.Content)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeAiOutputInvalid, err.Error(), false)
	}
	payload, err := json.Marshal(events.ResumeParseCompletedPayload{
		ResumeID:    asset.ID,
		UserID:      asset.UserID,
		ParseStatus: sharedtypes.TargetJobParseStatusReady,
	})
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeTargetImportFailed, err.Error(), true)
	}
	if h.newID == nil {
		return h.fail(ctx, asset, job, sharederrors.CodeTargetImportFailed, "resume parse event id generator not configured", true)
	}
	// D-20: parse directly produces the flat resume's structured content; the
	// parsed JSON is both the summary and the structured_profile.
	if err := h.store.CompleteParseSuccess(ctx, resumestore.CompleteParseSuccessInput{
		UserID:             asset.UserID,
		AssetID:            asset.ID,
		ParsedSummary:      parsed,
		StructuredProfile:  parsed,
		ParsedTextSnapshot: input,
		DisplayName:        displayName,
		OutboxEventID:      h.newID(),
		OutboxEventPayload: payload,
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

func (h *ParseHandler) resumeInput(ctx context.Context, asset resumestore.ParseAssetRecord) (string, error) {
	var raw string
	switch asset.SourceType {
	case "paste":
		raw = asset.OriginalText
	case "upload":
		if h.objects == nil {
			return "", fmt.Errorf("object reader is not configured")
		}
		if strings.TrimSpace(asset.FileObjectKey) == "" {
			return "", fmt.Errorf("file object key is empty")
		}
		body, err := h.objects.Read(ctx, asset.FileObjectKey, h.maxInputBytes)
		if err != nil {
			return "", err
		}
		raw = string(body)
	default:
		return "", fmt.Errorf("unsupported source_type %q", asset.SourceType)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("resume input is empty")
	}
	return raw, nil
}

func (h *ParseHandler) fail(ctx context.Context, asset resumestore.ParseAssetRecord, job targetjob.ClaimedJob, code, message string, retryable bool) targetjob.JobOutcome {
	if err := h.store.CompleteParseFailure(ctx, resumestore.CompleteParseFailureInput{
		UserID:    asset.UserID,
		AssetID:   asset.ID,
		ErrorCode: code,
		Now:       h.now(),
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

func buildPromptMessages(resolution PromptResolution, resumeText string) []aiclient.Message {
	messages := make([]aiclient.Message, 0, 2)
	if system := strings.TrimSpace(resolution.SystemMessage); system != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: system})
	}
	user := strings.TrimSpace(resumeText)
	if template := strings.TrimSpace(resolution.UserMessageTemplate); template != "" {
		user = strings.ReplaceAll(template, "{{resume_text}}", resumeText)
		user = strings.TrimSpace(user)
	}
	if user != "" {
		messages = append(messages, aiclient.Message{Role: "user", Content: user})
	}
	return messages
}

func decodeResumeParseResponse(content string) (json.RawMessage, *string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil, fmt.Errorf("AI response content was empty")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, nil, fmt.Errorf("AI response is not valid JSON: %v", err)
	}
	for _, key := range []string{"basics", "experiences", "projects", "education", "skills", "languages"} {
		if _, ok := parsed[key]; !ok {
			return nil, nil, fmt.Errorf("AI response missing %s", key)
		}
	}
	raw, err := json.Marshal(parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal parsed resume summary: %w", err)
	}
	return raw, deriveResumeDisplayName(parsed), nil
}

func deriveResumeDisplayName(parsed map[string]any) *string {
	name := fieldString(objectField(parsed, "basics"), "name")
	headline := firstNonEmpty(
		fieldString(objectField(parsed, "basics"), "headline"),
		fieldString(objectField(parsed, "basics"), "title"),
		fieldString(parsed, "headline"),
		fieldString(parsed, "title"),
		fieldString(parsed, "summary"),
		firstRecordField(parsed["experiences"], "title", "role"),
		firstRecordField(parsed["projects"], "name", "title"),
	)
	name = normalizeDisplayNamePart(name)
	headline = normalizeDisplayNamePart(headline)
	if name != "" && headline != "" && !strings.EqualFold(name, headline) {
		return stringPtr(truncateDisplayName(name + " - " + headline))
	}
	if name != "" {
		return stringPtr(truncateDisplayName(name))
	}
	if headline != "" {
		return stringPtr(truncateDisplayName(headline))
	}
	return nil
}

func objectField(values map[string]any, key string) map[string]any {
	value, ok := values[key].(map[string]any)
	if !ok {
		return nil
	}
	return value
}

func fieldString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key].(string)
	if !ok {
		return ""
	}
	return value
}

func firstRecordField(value any, keys ...string) string {
	records, ok := value.([]any)
	if !ok || len(records) == 0 {
		return ""
	}
	record, ok := records[0].(map[string]any)
	if !ok {
		return ""
	}
	for _, key := range keys {
		if value := fieldString(record, key); value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizeDisplayNamePart(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	switch strings.ToLower(value) {
	case "", "pasted resume", "paste resume", "pasted text", "paste text", "uploaded resume", "upload resume", "uploaded file", "upload file", "粘贴的简历", "粘帖的简历", "上传的简历":
		return ""
	default:
		return value
	}
}

func truncateDisplayName(value string) string {
	runes := []rune(value)
	if len(runes) <= 96 {
		return value
	}
	return strings.TrimSpace(string(runes[:96]))
}

func stringPtr(value string) *string {
	return &value
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
	return sharederrors.CodeAiFallbackExhausted, true
}

func safeFailureMessage(code, msg string) string {
	switch code {
	case sharederrors.CodeAiProviderTimeout,
		sharederrors.CodeAiFallbackExhausted,
		sharederrors.CodeAiOutputInvalid,
		sharederrors.CodeAiUnsupportedCapability,
		sharederrors.CodeAiProviderSecretMissing,
		sharederrors.CodeAiProviderConfigInvalid,
		sharederrors.CodeValidationFailed,
		sharederrors.CodeTargetInvalidStateTransition:
		return code
	default:
		return redactErrorMessage(msg)
	}
}

func redactErrorMessage(msg string) string {
	if len(msg) > 240 {
		msg = msg[:240]
	}
	lower := strings.ToLower(msg)
	for _, kw := range []string{
		"resume body",
		"resume text",
		"prompt body",
		"response body",
		"provider secret",
		"authorization:",
		"bearer ",
		"sk-",
	} {
		if strings.Contains(lower, kw) {
			return "redacted error message containing forbidden token"
		}
	}
	return msg
}

func coalesceFlag(v string) string {
	if strings.TrimSpace(v) == "" {
		return "none"
	}
	return v
}

type RegistryAdapter struct {
	client *registry.Client
}

func NewRegistryAdapter(client *registry.Client) *RegistryAdapter {
	if client == nil {
		return nil
	}
	return &RegistryAdapter{client: client}
}

func (a *RegistryAdapter) Resolve(ctx context.Context, featureKey string, language string) (PromptResolution, error) {
	if a == nil || a.client == nil {
		return PromptResolution{}, ErrPromptUnsupported
	}
	resolved, err := a.client.ResolveActive(ctx, featureKey, language)
	if err != nil {
		if errors.Is(err, registry.ErrPromptUnsupported) || errors.Is(err, registry.ErrLanguageUnsupported) {
			return PromptResolution{}, ErrPromptUnsupported
		}
		return PromptResolution{}, fmt.Errorf("resume parse registry resolve: %w", err)
	}
	if resolved.FeatureKey != featureKey {
		return PromptResolution{}, fmt.Errorf("resume parse registry returned feature_key %q, expected %q", resolved.FeatureKey, featureKey)
	}
	return PromptResolution{
		PromptVersion:       resolved.PromptVersion,
		RubricVersion:       resolved.RubricVersion,
		ModelProfileName:    resolved.ModelProfileName,
		DataSourceVersion:   resolved.DataSourceVersion,
		FeatureFlag:         resolved.FeatureFlag,
		SystemMessage:       resolved.SystemMessage,
		UserMessageTemplate: resolved.UserMessageTemplate,
		OutputSchema:        resolved.OutputSchema,
	}, nil
}
