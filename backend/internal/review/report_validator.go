package review

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const (
	copiedMessageRuneLimit            = 120
	reportActionLabelSchemaRuneLimit  = 200
	reportActionLabelEnglishWordLimit = 24
	reportActionLabelChineseRuneLimit = 64
)

var (
	reportDimensionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	forbiddenEnglishClaims     = []*regexp.Regexp{
		regexp.MustCompile(`\bhiring\s+(probability|chance|likelihood)\b`),
		regexp.MustCompile(`\b(probability|chance|likelihood)\s+of\s+(being\s+)?hired\b`),
		regexp.MustCompile(`\b(candidate\s+)?ranking\b`),
		regexp.MustCompile(`\b(rank(ed|ing)?\s+(among|against)\s+candidates?|percentile)\b`),
		regexp.MustCompile(`\b(beat|beats|beating|outperform|outperforms|outperformed|outperforming)\s+(other\s+)?candidates?\b`),
		regexp.MustCompile(`\bpercentage\s+of\s+candidates?\s+(beat|beaten|outperformed)\b`),
		regexp.MustCompile(`\b(candidate|readiness)\s+(score|rating)\b`),
		regexp.MustCompile(`\breadiness\s*(is|:)?\s*[0-9]+\s*(/|out\s+of)\s*[0-9]+\b`),
		regexp.MustCompile(`\b(speech\s+rate|speaking\s+speed|pace\s+of\s+speech)\b`),
		regexp.MustCompile(`\b((speaking|speech)\s+)?pauses?\b`),
		regexp.MustCompile(`\b(emotion|emotional\s+state|sentiment)\b`),
		regexp.MustCompile(`\bpersonality(\s+(trait|assessment|profile))?\b`),
	}
)

type ReportValidationIssue struct {
	Path string
	Code string
}

type ReportValidationIssues []ReportValidationIssue

func (issues ReportValidationIssues) Error() string {
	parts := make([]string, 0, len(issues))
	for _, issue := range issues {
		parts = append(parts, issue.Path+":"+issue.Code)
	}
	return strings.Join(parts, ",")
}

// ReportContentValidationContext is the minimal frozen coordinate set needed
// by the product report validator. Runtime and evaluation callers supply these
// coordinates from their respective frozen context representations while
// sharing every business invariant below.
type ReportContentValidationContext struct {
	Language         string
	HasNextRound     bool
	LastMessageSeqNo int32
}

// ValidateReportContentJSON closed-decodes one model response and applies the
// complete product report validator. Returned issues contain only bounded
// paths and codes, never model-authored values.
func ValidateReportContentJSON(raw string, validationContext ReportContentValidationContext, messages []MessageSnapshot) (ReportContentDraft, []ReportValidationIssue) {
	draft, issues := decodeReportContent([]byte(raw))
	if len(issues) > 0 {
		return ReportContentDraft{}, issues
	}
	issues = ValidateReportContent(draft, validationContext, messages)
	return draft, issues
}

// ValidateReportContent applies the complete product report validator to an
// already decoded draft.
func ValidateReportContent(content ReportContentDraft, validationContext ReportContentValidationContext, messages []MessageSnapshot) []ReportValidationIssue {
	validator := reportContentValidator{
		language:    validationContext.Language,
		messages:    messages,
		messageRole: make(map[int]string, len(messages)),
		lastSeqNo:   validationContext.LastMessageSeqNo,
		hasNext:     validationContext.HasNextRound,
	}
	for _, message := range messages {
		validator.messageRole[message.SeqNo] = message.Role
	}
	validator.validate(content)
	return validator.issues
}

// ValidateReportActionLabelLimits applies the product's language-specific
// action-label limits without requiring a persisted report context. Callers
// must validate the surrounding report output schema separately.
func ValidateReportActionLabelLimits(language string, actions []ReportNextActionDraft) []ReportValidationIssue {
	issues := make([]ReportValidationIssue, 0)
	for index, action := range actions {
		path := indexedPath("$.nextActions", index) + ".label"
		switch language {
		case "en":
			if countEnglishWhitespaceWords(action.Label) > reportActionLabelEnglishWordLimit {
				issues = append(issues, ReportValidationIssue{Path: path, Code: "max_words"})
			}
		case "zh-CN":
			if utf8.RuneCountInString(action.Label) > reportActionLabelChineseRuneLimit {
				issues = append(issues, ReportValidationIssue{Path: path, Code: "max_code_points"})
			}
		}
	}
	return issues
}

// Keep backend/evalkit admission byte-for-byte aligned with the frontend's
// ECMAScript /\s/u delimiter set. In particular, it includes U+FEFF but not
// Unicode White_Space's U+0085 NEXT LINE.
func countEnglishWhitespaceWords(value string) int {
	return len(strings.FieldsFunc(value, isECMAScriptWhitespace))
}

func isECMAScriptWhitespace(r rune) bool {
	switch r {
	case '\u0009', '\u000B', '\u000C', '\uFEFF', '\u000A', '\u000D', '\u2028', '\u2029':
		return true
	default:
		return unicode.In(r, unicode.Zs)
	}
}

type reportContentValidator struct {
	issues      []ReportValidationIssue
	language    string
	messages    []MessageSnapshot
	messageRole map[int]string
	lastSeqNo   int32
	hasNext     bool
}

func validateReportContent(content ReportContentDraft, frozen practicedomain.ReportContextSnapshot, messages []MessageSnapshot) []ReportValidationIssue {
	return ValidateReportContent(content, ReportContentValidationContext{
		Language:         frozen.Conversation.Language,
		HasNextRound:     frozen.HasNextRound,
		LastMessageSeqNo: frozen.Conversation.LastMessageSeqNo,
	}, messages)
}

func (v *reportContentValidator) validate(content ReportContentDraft) {
	if v.language != "en" && v.language != "zh-CN" {
		v.add("$.language", "unsupported_language")
	}
	v.validateText("$.summary", content.Summary, 360)
	v.validateReadiness(content.PreparednessLevel)

	if len(content.DimensionAssessments) < 1 {
		v.add("$.dimensionAssessments", "min_items")
	}
	if len(content.DimensionAssessments) > 6 {
		v.add("$.dimensionAssessments", "max_items")
	}
	if len(content.Highlights) > 4 {
		v.add("$.highlights", "max_items")
	}
	if len(content.Issues) > 4 {
		v.add("$.issues", "max_items")
	}
	evidenceCount := len(content.Highlights) + len(content.Issues)
	if evidenceCount < 1 {
		v.add("$.evidence", "min_items")
	}
	if evidenceCount > 6 {
		v.add("$.evidence", "max_items")
	}
	if len(content.NextActions) < 1 {
		v.add("$.nextActions", "min_items")
	}
	if len(content.NextActions) > 2 {
		v.add("$.nextActions", "max_items")
	}
	if len(content.RetryFocusDimensionCodes) > 6 {
		v.add("$.retryFocusDimensionCodes", "max_items")
	}

	dimensions := make(map[string]DimensionAssessmentDraft, len(content.DimensionAssessments))
	seenDimensionCodes := make(map[string]struct{}, len(content.DimensionAssessments))
	for index, dimension := range content.DimensionAssessments {
		path := indexedPath("$.dimensionAssessments", index)
		if !reportDimensionCodePattern.MatchString(dimension.Code) {
			v.add(path+".code", "invalid_format")
		}
		if _, duplicate := seenDimensionCodes[dimension.Code]; duplicate {
			v.add(path+".code", "duplicate")
		} else {
			seenDimensionCodes[dimension.Code] = struct{}{}
			dimensions[dimension.Code] = dimension
		}
		v.validateText(path+".label", dimension.Label, 48)
		if !validDimensionStatus(dimension.Status) {
			v.add(path+".status", "invalid_enum")
		}
		if !validConfidence(dimension.Confidence) {
			v.add(path+".confidence", "invalid_enum")
		}
		if dimension.Confidence == sharedtypes.ConfidenceLow &&
			(dimension.Status == sharedtypes.DimensionStatusStrong || dimension.Status == sharedtypes.DimensionStatusNeedsWork) {
			v.add(path+".confidence", "low_confidence_status_invalid")
		}
	}

	evidenceByDimension := make(map[string]int, len(dimensions))
	highlightsByDimension := make(map[string]int, len(dimensions))
	issuesByDimension := make(map[string]int, len(dimensions))
	v.validateEvidenceItems("$.highlights", content.Highlights, dimensions, evidenceByDimension, highlightsByDimension)
	v.validateEvidenceItems("$.issues", content.Issues, dimensions, evidenceByDimension, issuesByDimension)
	for index, highlight := range content.Highlights {
		dimension, exists := dimensions[highlight.DimensionCode]
		if exists && highlight.Confidence == sharedtypes.ConfidenceLow &&
			(dimension.Status == sharedtypes.DimensionStatusStrong || dimension.Status == sharedtypes.DimensionStatusNeedsWork) {
			v.add(indexedPath("$.highlights", index)+".confidence", "low_confidence_status_invalid")
		}
	}
	for index, issue := range content.Issues {
		path := indexedPath("$.issues", index)
		if dimension, exists := dimensions[issue.DimensionCode]; exists && dimension.Status != sharedtypes.DimensionStatusNeedsWork {
			v.add(path+".dimensionCode", "not_needs_work")
		}
		if issue.Confidence == sharedtypes.ConfidenceLow {
			v.add(path+".confidence", "low_confidence_issue_invalid")
		}
	}

	hasNeedsWork := false
	for index, dimension := range content.DimensionAssessments {
		path := indexedPath("$.dimensionAssessments", index)
		if evidenceByDimension[dimension.Code] == 0 {
			v.add(path, "missing_evidence")
		}
		switch dimension.Status {
		case sharedtypes.DimensionStatusStrong:
			if highlightsByDimension[dimension.Code] == 0 {
				v.add(path+".status", "strong_requires_highlight")
			}
		case sharedtypes.DimensionStatusNeedsWork:
			hasNeedsWork = true
			if issuesByDimension[dimension.Code] == 0 {
				v.add(path+".status", "needs_work_requires_issue")
			}
		}
	}
	retryPresent := v.validateActions(content.NextActions)
	v.validateReadinessConsistency(content, hasNeedsWork)
	v.validateRetryFocus(content.RetryFocusDimensionCodes, retryPresent, dimensions, issuesByDimension)
}

func (v *reportContentValidator) validateReadinessConsistency(content ReportContentDraft, hasNeedsWork bool) {
	switch content.PreparednessLevel {
	case sharedtypes.ReadinessTierNotReady, sharedtypes.ReadinessTierNeedsPractice:
		if !hasNeedsWork {
			v.add("$.preparednessLevel", "requires_needs_work")
		}
		if len(content.NextActions) == 0 || content.NextActions[0].Type != string(NextActionRetryCurrentRound) {
			v.add("$.nextActions[0].type", "retry_current_round_required")
		}
		for index, action := range content.NextActions {
			if action.Type == string(NextActionNextRound) {
				v.add(indexedPath("$.nextActions", index)+".type", "next_round_inconsistent_with_readiness")
			}
		}
	case sharedtypes.ReadinessTierBasicallyReady:
		if hasNeedsWork {
			v.add("$.preparednessLevel", "forbids_needs_work")
		}
		hasSupportedUsableDimension := false
		for _, dimension := range content.DimensionAssessments {
			if (dimension.Status == sharedtypes.DimensionStatusMeetsBar || dimension.Status == sharedtypes.DimensionStatusStrong) &&
				(dimension.Confidence == sharedtypes.ConfidenceMedium || dimension.Confidence == sharedtypes.ConfidenceHigh) {
				hasSupportedUsableDimension = true
				break
			}
		}
		if !hasSupportedUsableDimension {
			v.add("$.preparednessLevel", "requires_supported_usable_dimension")
		}
	case sharedtypes.ReadinessTierWellPrepared:
		if len(content.DimensionAssessments) < 2 {
			v.add("$.dimensionAssessments", "well_prepared_min_items")
		}
		for index, dimension := range content.DimensionAssessments {
			path := indexedPath("$.dimensionAssessments", index)
			if dimension.Status != sharedtypes.DimensionStatusStrong {
				v.add(path+".status", "well_prepared_requires_strong")
			}
			if dimension.Confidence != sharedtypes.ConfidenceHigh {
				v.add(path+".confidence", "well_prepared_requires_high_confidence")
			}
		}
		if len(content.Issues) > 0 {
			v.add("$.issues", "well_prepared_forbids_issues")
		}
		userMessages := make(map[int32]struct{})
		for _, highlight := range content.Highlights {
			for _, seqNo := range highlight.SourceMessageSeqNos {
				if seqNo > 0 && seqNo <= v.lastSeqNo && v.messageRole[int(seqNo)] == "user" {
					userMessages[seqNo] = struct{}{}
				}
			}
		}
		if len(userMessages) < 2 {
			v.add("$.highlights", "well_prepared_requires_two_user_messages")
		}
	}
}

func (v *reportContentValidator) validateReadiness(readiness sharedtypes.ReadinessTier) {
	switch readiness {
	case sharedtypes.ReadinessTierNotReady,
		sharedtypes.ReadinessTierNeedsPractice,
		sharedtypes.ReadinessTierBasicallyReady,
		sharedtypes.ReadinessTierWellPrepared:
	default:
		v.add("$.preparednessLevel", "invalid_enum")
	}
}

func (v *reportContentValidator) validateEvidenceItems(
	basePath string,
	items []ReportEvidenceDraft,
	dimensions map[string]DimensionAssessmentDraft,
	evidenceByDimension map[string]int,
	kindByDimension map[string]int,
) {
	for index, item := range items {
		path := indexedPath(basePath, index)
		if !reportDimensionCodePattern.MatchString(item.DimensionCode) {
			v.add(path+".dimensionCode", "invalid_format")
		}
		if _, exists := dimensions[item.DimensionCode]; !exists {
			v.add(path+".dimensionCode", "unknown_reference")
		} else {
			evidenceByDimension[item.DimensionCode]++
			kindByDimension[item.DimensionCode]++
		}
		v.validateText(path+".evidence", item.Evidence, 240)
		if !validConfidence(item.Confidence) {
			v.add(path+".confidence", "invalid_enum")
		}
		v.validateAnchors(path+".sourceMessageSeqNos", item.SourceMessageSeqNos)
	}
}

func (v *reportContentValidator) validateAnchors(path string, anchors []int32) {
	if len(anchors) == 0 {
		v.add(path, "min_items")
		return
	}
	seen := make(map[int32]struct{}, len(anchors))
	var previous int32
	for index, seqNo := range anchors {
		itemPath := indexedPath(path, index)
		if seqNo <= 0 {
			v.add(itemPath, "not_positive")
		}
		if index > 0 && seqNo < previous {
			v.add(itemPath, "not_ascending")
		}
		if _, duplicate := seen[seqNo]; duplicate {
			v.add(itemPath, "duplicate")
		} else {
			seen[seqNo] = struct{}{}
		}
		if seqNo > v.lastSeqNo {
			v.add(itemPath, "out_of_range")
		}
		if role, exists := v.messageRole[int(seqNo)]; !exists || role != "user" {
			v.add(itemPath, "not_user_message")
		}
		previous = seqNo
	}
}

func (v *reportContentValidator) validateActions(actions []ReportNextActionDraft) bool {
	seen := make(map[string]struct{}, len(actions))
	retryPresent := false
	for index, action := range actions {
		path := indexedPath("$.nextActions", index)
		switch action.Type {
		case string(NextActionRetryCurrentRound):
			retryPresent = true
		case string(NextActionNextRound):
			if !v.hasNext {
				v.add(path+".type", "next_round_unavailable")
			}
		case string(NextActionReviewEvidence):
		default:
			v.add(path+".type", "invalid_enum")
		}
		if _, duplicate := seen[action.Type]; duplicate {
			v.add(path+".type", "duplicate")
		} else {
			seen[action.Type] = struct{}{}
		}
		labelPath := path + ".label"
		v.validateText(labelPath, action.Label, reportActionLabelSchemaRuneLimit)
	}
	v.issues = append(v.issues, ValidateReportActionLabelLimits(v.language, actions)...)
	return retryPresent
}

func (v *reportContentValidator) validateRetryFocus(
	codes []string,
	retryPresent bool,
	dimensions map[string]DimensionAssessmentDraft,
	issuesByDimension map[string]int,
) {
	issueCount := 0
	expectedSet := make(map[string]struct{}, len(issuesByDimension))
	for _, count := range issuesByDimension {
		issueCount += count
	}
	for code, count := range issuesByDimension {
		if count > 0 && dimensions[code].Status == sharedtypes.DimensionStatusNeedsWork {
			expectedSet[code] = struct{}{}
		}
	}
	expected := make([]string, 0, len(expectedSet))
	for code := range expectedSet {
		expected = append(expected, code)
	}
	sort.Strings(expected)
	exactGenericException := issueCount == 1 &&
		(issuesByDimension["answer_depth"] == 1 || issuesByDimension["answer_relevance"] == 1)
	if retryPresent {
		switch {
		case exactGenericException && len(codes) > 0:
			v.add("$.retryFocusDimensionCodes", "generic_exception_requires_empty_focus")
		case !exactGenericException && (len(expected) == 0 || !equalStringLists(codes, expected)):
			v.add("$.retryFocusDimensionCodes", "retry_focus_mismatch")
		}
	}
	if len(codes) > 0 && !retryPresent {
		v.add("$.retryFocusDimensionCodes", "retry_action_required")
	}
	seen := make(map[string]struct{}, len(codes))
	previous := ""
	for index, code := range codes {
		path := indexedPath("$.retryFocusDimensionCodes", index)
		if !reportDimensionCodePattern.MatchString(code) {
			v.add(path, "invalid_format")
		}
		if index > 0 && code < previous {
			v.add(path, "not_ascending")
		}
		if _, duplicate := seen[code]; duplicate {
			v.add(path, "duplicate")
		} else {
			seen[code] = struct{}{}
		}
		dimension, exists := dimensions[code]
		if !exists {
			v.add(path, "unknown_reference")
		} else {
			if dimension.Status != sharedtypes.DimensionStatusNeedsWork {
				v.add(path, "not_needs_work")
			}
			if issuesByDimension[code] == 0 {
				v.add(path, "not_issue_backed")
			}
		}
		previous = code
	}
}

func equalStringLists(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func (v *reportContentValidator) validateText(path, text string, maxRunes int) {
	runeCount := utf8.RuneCountInString(text)
	if runeCount < 1 {
		v.add(path, "required")
	}
	if runeCount > maxRunes {
		v.add(path, "max_length")
	}
	if !utf8.ValidString(text) {
		v.add(path, "invalid_utf8")
	}
	if containsForbiddenReportClaim(text) {
		v.add(path, "forbidden_claim")
	}
	if containsCopiedMessageText(text, v.messages) {
		v.add(path, "copied_message_text")
	}
	if reportTextLanguageMismatch(text, v.language) {
		v.add(path, "language_mismatch")
	}
}

func (v *reportContentValidator) add(path, code string) {
	v.issues = append(v.issues, ReportValidationIssue{Path: path, Code: code})
}

func validDimensionStatus(status sharedtypes.DimensionStatus) bool {
	switch status {
	case sharedtypes.DimensionStatusStrong, sharedtypes.DimensionStatusMeetsBar, sharedtypes.DimensionStatusNeedsWork:
		return true
	default:
		return false
	}
}

func validConfidence(confidence sharedtypes.Confidence) bool {
	switch confidence {
	case sharedtypes.ConfidenceHigh, sharedtypes.ConfidenceMedium, sharedtypes.ConfidenceLow:
		return true
	default:
		return false
	}
}

func containsForbiddenReportClaim(text string) bool {
	lower := strings.ToLower(text)
	for _, phrase := range []string{"录用概率", "候选人排名", "候选人排行", "候选人评分", "准备度评分", "准备度得分", "击败比例", "击败百分比", "击败其他候选人", "语速", "停顿", "情绪", "人格"} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	for _, pattern := range forbiddenEnglishClaims {
		if pattern.MatchString(lower) {
			return true
		}
	}
	return false
}

func containsCopiedMessageText(text string, messages []MessageSnapshot) bool {
	runes := []rune(text)
	if len(runes) < copiedMessageRuneLimit {
		return false
	}
	for start := 0; start+copiedMessageRuneLimit <= len(runes); start++ {
		candidate := string(runes[start : start+copiedMessageRuneLimit])
		for _, message := range messages {
			if strings.Contains(message.Content, candidate) {
				return true
			}
		}
	}
	return false
}

func reportTextLanguageMismatch(text, language string) bool {
	if text == "" || (language != "en" && language != "zh-CN") {
		return false
	}
	hanCount := 0
	latinCount := 0
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hanCount++
		}
		if unicode.Is(unicode.Latin, r) {
			latinCount++
		}
	}
	if language == "en" {
		return hanCount >= 4 && hanCount*2 > latinCount
	}
	return hanCount == 0 && latinCount >= 12 && len(strings.Fields(text)) >= 3
}

func indexedPath(base string, index int) string {
	return base + "[" + strconv.Itoa(index) + "]"
}
