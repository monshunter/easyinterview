package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	pdf "github.com/ledongthuc/pdf"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const FeatureKeyResumeParse = string(featurekeys.ResumeParse)

const defaultMaxResumeInputBytes int64 = 8 * 1024 * 1024

var ErrPromptUnsupported = errors.New("prompt registry: feature/language is not enabled")

var resumeFileNamePattern = regexp.MustCompile(`(?i)\.(pdf|txt|md|markdown)$`)

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

func (h *ParseHandler) Handle(ctx context.Context, job runner.ClaimedJob) runner.JobOutcome {
	if h == nil || h.store == nil || h.registry == nil || h.ai == nil {
		return runner.JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: "resume parse handler not initialised"}
	}
	asset, err := h.store.GetForParse(ctx, job.ResourceID)
	if err != nil {
		return runner.JobOutcome{ErrorCode: sharederrors.CodeTargetImportFailed, ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error())}
	}
	switch asset.ParseStatus {
	case sharedtypes.TargetJobParseStatusQueued, sharedtypes.TargetJobParseStatusFailed:
		if err := h.store.MarkParsing(ctx, resumestore.StatusUpdateInput{UserID: asset.UserID, AssetID: asset.ID, Now: h.now()}); err != nil {
			return runner.JobOutcome{ErrorCode: sharederrors.CodeTargetInvalidStateTransition, ErrorMessage: safeFailureMessage(sharederrors.CodeTargetInvalidStateTransition, err.Error())}
		}
	case sharedtypes.TargetJobParseStatusProcessing:
	case sharedtypes.TargetJobParseStatusReady:
		return runner.JobOutcome{Succeeded: true}
	default:
		return runner.JobOutcome{ErrorCode: sharederrors.CodeTargetInvalidStateTransition, ErrorMessage: sharederrors.CodeTargetInvalidStateTransition}
	}

	input, err := h.resumeInput(ctx, asset)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeValidationFailed, err.Error(), false, "")
	}
	parsedTextSnapshot := buildResumeMarkdownFallback(input)
	if parsedTextSnapshot == "" {
		return h.fail(ctx, asset, job, sharederrors.CodeValidationFailed, "resume snapshot is empty", false, input)
	}
	resolution, err := h.registry.Resolve(ctx, FeatureKeyResumeParse, asset.Language)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeAiProviderConfigInvalid, err.Error(), false, input)
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
		return h.fail(ctx, asset, job, code, err.Error(), retryable, input)
	}
	if strings.EqualFold(strings.TrimSpace(complete.FinishReason), "length") {
		return h.fail(ctx, asset, job, sharederrors.CodeAiOutputInvalid, "AI response reached its output limit", false, input)
	}
	parsed, displayName, err := decodeResumeParseResponse(complete.Content, input)
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeAiOutputInvalid, err.Error(), false, input)
	}
	payload, err := json.Marshal(events.ResumeParseCompletedPayload{
		ResumeID:    asset.ID,
		UserID:      asset.UserID,
		ParseStatus: sharedtypes.TargetJobParseStatusReady,
	})
	if err != nil {
		return h.fail(ctx, asset, job, sharederrors.CodeTargetImportFailed, err.Error(), true, input)
	}
	if h.newID == nil {
		return h.fail(ctx, asset, job, sharederrors.CodeTargetImportFailed, "resume parse event id generator not configured", true, input)
	}
	// D-20: parse directly produces the flat resume's structured content; the
	// parsed JSON is both the summary and the structured_profile.
	if err := h.store.CompleteParseSuccess(ctx, resumestore.CompleteParseSuccessInput{
		UserID:             asset.UserID,
		AssetID:            asset.ID,
		ParsedSummary:      parsed,
		StructuredProfile:  parsed,
		ParsedTextSnapshot: parsedTextSnapshot,
		DisplayName:        displayName,
		OutboxEventID:      h.newID(),
		OutboxEventPayload: payload,
		Now:                h.now(),
	}); err != nil {
		return runner.JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	return runner.JobOutcome{Succeeded: true}
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
		raw, err = extractUploadResumeText(asset.FileObjectKey, body)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported source_type %q", asset.SourceType)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("resume input is empty")
	}
	return raw, nil
}

func extractUploadResumeText(objectKey string, body []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(objectKey)))
	var (
		raw string
		err error
	)
	switch ext {
	case ".pdf":
		raw, err = extractPDFText(body)
	case ".md", ".markdown", ".txt", "":
		raw = strings.ToValidUTF8(string(body), "")
	default:
		return "", fmt.Errorf("unsupported resume upload type %q", ext)
	}
	if err != nil {
		return "", err
	}
	raw = normalizeExtractedResumeText(raw)
	if raw == "" {
		return "", fmt.Errorf("upload resume text is empty")
	}
	return raw, nil
}

func extractPDFText(body []byte) (string, error) {
	if text, err := extractPDFTextWithPdftotext(body); err == nil && isReadableExtractedResumeText(text) {
		return text, nil
	}
	reader, err := pdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err == nil {
		plain, plainErr := reader.GetPlainText()
		if plainErr == nil {
			data, readErr := io.ReadAll(plain)
			if readErr == nil {
				if text := normalizeExtractedResumeText(string(data)); isReadableExtractedResumeText(text) {
					return text, nil
				}
			}
		}
	}
	if text := extractPDFLiteralText(body); isReadableExtractedResumeText(text) {
		return text, nil
	}
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w", err)
	}
	return "", fmt.Errorf("extract pdf text: no readable text found")
}

func extractPDFTextWithPdftotext(body []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", "-", "-")
	cmd.Stdin = bytes.NewReader(body)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("pdftotext timeout")
		}
		return "", fmt.Errorf("pdftotext unavailable or failed")
	}
	text := normalizeExtractedResumeText(stdout.String())
	if text == "" {
		return "", fmt.Errorf("pdftotext returned empty text")
	}
	return text, nil
}

var pdfLiteralTextPattern = regexp.MustCompile(`\((?:\\.|[^\\()])*\)`)

func extractPDFLiteralText(body []byte) string {
	source := string(body)
	if !strings.Contains(source, "Tj") && !strings.Contains(source, "TJ") {
		return ""
	}
	matches := pdfLiteralTextPattern.FindAllString(source, -1)
	parts := make([]string, 0, len(matches))
	for _, match := range matches {
		unescaped := unescapePDFLiteral(match[1 : len(match)-1])
		if strings.TrimSpace(unescaped) != "" {
			parts = append(parts, unescaped)
		}
	}
	return normalizeExtractedResumeText(strings.Join(parts, "\n"))
}

func unescapePDFLiteral(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\`,
		`\(`, `(`,
		`\)`, `)`,
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\b`, "\b",
		`\f`, "\f",
	)
	return replacer.Replace(value)
}

func normalizeExtractedResumeText(value string) string {
	lines := strings.Split(strings.ToValidUTF8(value, ""), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			normalized = append(normalized, line)
		}
	}
	return strings.TrimSpace(strings.Join(normalized, "\n"))
}

func isReadableExtractedResumeText(value string) bool {
	value = normalizeExtractedResumeText(value)
	runes := []rune(value)
	if len(runes) < 8 {
		return false
	}
	readable := 0
	nonPrintable := 0
	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			readable++
		}
		if r == unicode.ReplacementChar || (!unicode.IsPrint(r) && !unicode.IsSpace(r)) {
			nonPrintable++
		}
	}
	return readable >= 8 && nonPrintable == 0
}

func (h *ParseHandler) fail(ctx context.Context, asset resumestore.ParseAssetRecord, job runner.ClaimedJob, code, message string, retryable bool, parsedTextSnapshot string) runner.JobOutcome {
	markdownSnapshot := buildResumeMarkdownFallback(parsedTextSnapshot)
	if err := h.store.CompleteParseFailure(ctx, resumestore.CompleteParseFailureInput{
		UserID:             asset.UserID,
		AssetID:            asset.ID,
		ErrorCode:          code,
		ParsedTextSnapshot: markdownSnapshot,
		DisplayName:        deriveDisplayNameFromResumeText(markdownSnapshot),
		Now:                h.now(),
	}); err != nil {
		return runner.JobOutcome{
			ErrorCode:    sharederrors.CodeTargetImportFailed,
			ErrorMessage: safeFailureMessage(sharederrors.CodeTargetImportFailed, err.Error()),
			Retryable:    true,
		}
	}
	return runner.JobOutcome{
		ErrorCode:    code,
		ErrorMessage: safeFailureMessage(code, message),
		Retryable:    retryable,
	}
}

var resumeFallbackSectionHeadings = []string{
	"个人摘要",
	"核心能力",
	"工作经历",
	"项目经历",
	"教育经历",
	"教育背景",
	"专业技能",
	"技能",
	"Summary",
	"Profile",
	"Core Skills",
	"Skills",
	"Experience",
	"Work Experience",
	"Projects",
	"Education",
}

func buildResumeMarkdownFallback(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	lines := splitFallbackResumeLines(value)
	if len(lines) == 0 {
		return ""
	}

	title := fallbackTitleLine(lines[0])
	out := []string{title}
	currentSection := ""
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			out = appendWithBlank(out, normalizeMarkdownHeading(line))
			currentSection = strings.TrimSpace(strings.TrimLeft(line, "#"))
			continue
		}
		if isKnownResumeFallbackSection(line) {
			out = appendWithBlank(out, "## "+line)
			currentSection = line
			continue
		}
		if heading, rest, ok := splitFallbackSectionPrefix(line); ok {
			out = appendWithBlank(out, "## "+heading)
			currentSection = heading
			if rest != "" {
				out = appendFallbackContent(out, currentSection, rest)
			}
			continue
		}
		out = appendFallbackContent(out, currentSection, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func splitFallbackResumeLines(value string) []string {
	normalized := normalizeExtractedResumeText(value)
	normalized = regexp.MustCompile(`(?i)(Phone:|Email:|GitHub:)`).ReplaceAllString(normalized, "\n$1")
	rawLines := strings.Split(normalized, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func fallbackTitleLine(line string) string {
	line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
	if name, headline := splitNameHeadlineLine(line); name != "" && headline != "" {
		return "# " + name + " - " + headline
	}
	return "# " + line
}

func normalizeMarkdownHeading(line string) string {
	level := 0
	for _, r := range line {
		if r != '#' {
			break
		}
		level++
	}
	if level <= 0 {
		return line
	}
	if level > 3 {
		level = 3
	}
	text := strings.TrimSpace(strings.TrimLeft(line, "#"))
	if text == "" {
		return ""
	}
	return strings.Repeat("#", level) + " " + text
}

func isKnownResumeFallbackSection(line string) bool {
	for _, heading := range resumeFallbackSectionHeadings {
		if line == heading {
			return true
		}
	}
	return false
}

func splitFallbackSectionPrefix(line string) (string, string, bool) {
	if isContactLikeLine(line) {
		return "", "", false
	}
	for _, sep := range []string{"：", ":"} {
		parts := strings.SplitN(line, sep, 2)
		if len(parts) != 2 {
			continue
		}
		heading := strings.TrimSpace(parts[0])
		rest := strings.TrimSpace(parts[1])
		if heading == "" || rest == "" || len([]rune(heading)) > 48 {
			continue
		}
		if isKnownResumeFallbackSection(heading) {
			return heading, rest, true
		}
	}
	return "", "", false
}

func appendFallbackContent(out []string, currentSection string, line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return out
	}
	line = strings.TrimSpace(strings.TrimLeft(line, "：:"))
	if line == "" {
		return out
	}
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return append(out, line)
	}
	if isSkillsLikeSection(currentSection) {
		if label, rest, sep, ok := splitFallbackLabelValue(line); ok {
			return append(out, "- **"+label+"**"+sep+" "+rest)
		}
		return append(out, "- "+line)
	}
	return append(out, line)
}

func appendWithBlank(out []string, line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return out
	}
	if len(out) > 0 && out[len(out)-1] != "" {
		out = append(out, "")
	}
	return append(out, line)
}

func isSkillsLikeSection(section string) bool {
	switch section {
	case "核心能力", "专业技能", "技能", "Core Skills", "Skills":
		return true
	default:
		return false
	}
}

func splitFallbackLabelValue(line string) (string, string, string, bool) {
	for _, sep := range []string{"：", ":"} {
		parts := strings.SplitN(line, sep, 2)
		if len(parts) != 2 {
			continue
		}
		label := strings.TrimSpace(parts[0])
		rest := strings.TrimSpace(parts[1])
		if label == "" || rest == "" || len([]rune(label)) > 64 || strings.Contains(label, "@") {
			continue
		}
		return label, rest, sep, true
	}
	return "", "", "", false
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

func decodeResumeParseResponse(content string, resumeText string) (json.RawMessage, *string, error) {
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
	delete(parsed, "markdownText")
	raw, err := json.Marshal(parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal parsed resume summary: %w", err)
	}
	return raw, deriveResumeDisplayName(parsed, resumeText), nil
}

func deriveResumeDisplayName(parsed map[string]any, resumeText string) *string {
	if displayName := normalizeDisplayNameCandidate(fieldString(parsed, "displayName"), resumeText); displayName != "" {
		return stringPtr(truncateDisplayName(displayName))
	}
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
	return normalizeDisplayNameCandidate(value, "")
}

func normalizeDisplayNameCandidate(value string, resumeText string) string {
	value = strings.TrimSpace(strings.TrimPrefix(strings.Join(strings.Fields(value), " "), "#"))
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	switch lower {
	case "", "pasted resume", "paste resume", "pasted text", "paste text", "uploaded resume", "upload resume", "uploaded file", "upload file", "粘贴的简历", "粘帖的简历", "上传的简历":
		return ""
	}
	if looksLikeResumeFileName(value) {
		return ""
	}
	if resumeText != "" && isSameDisplayName(value, firstResumeTextLine(resumeText)) {
		return ""
	}
	return value
}

func looksLikeResumeFileName(value string) bool {
	return resumeFileNamePattern.MatchString(strings.TrimSpace(value))
}

func isSameDisplayName(left string, right string) bool {
	return strings.EqualFold(strings.Join(strings.Fields(left), " "), strings.Join(strings.Fields(right), " "))
}

func firstResumeTextLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func deriveDisplayNameFromResumeText(value string) *string {
	lines := resumeSignalLines(value)
	if len(lines) == 0 {
		return nil
	}
	if name, headline := splitNameHeadlineLine(lines[0]); name != "" && headline != "" {
		return stringPtr(truncateDisplayName(name + " - " + headline))
	}
	if len(lines) >= 2 {
		name := normalizeDisplayNameCandidate(lines[0], "")
		headline := normalizeDisplayNameCandidate(lines[1], "")
		if name != "" && headline != "" && !isContactLikeLine(headline) && !isSameDisplayName(name, headline) {
			return stringPtr(truncateDisplayName(name + " - " + headline))
		}
	}
	return nil
}

func resumeSignalLines(value string) []string {
	rawLines := strings.Split(normalizeExtractedResumeText(value), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		line = strings.TrimSpace(line)
		if line == "" || isContactLikeLine(line) {
			continue
		}
		lines = append(lines, line)
		if len(lines) >= 4 {
			break
		}
	}
	return lines
}

func splitNameHeadlineLine(line string) (string, string) {
	for _, sep := range []string{" | ", "｜", " - ", " — ", " – ", " · "} {
		parts := strings.Split(line, sep)
		if len(parts) < 2 {
			continue
		}
		name := normalizeDisplayNameCandidate(parts[0], "")
		headline := normalizeDisplayNameCandidate(strings.Join(parts[1:], " "), "")
		if name != "" && headline != "" && !isContactLikeLine(headline) {
			return name, headline
		}
	}
	return "", ""
}

func isContactLikeLine(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "@") ||
		strings.Contains(lower, "github") ||
		strings.Contains(lower, "phone") ||
		strings.Contains(lower, "email") ||
		strings.Contains(value, "电话") ||
		strings.Contains(value, "邮箱")
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
