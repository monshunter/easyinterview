package review

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestValidateGroundedReportAcceptsExactBounds(t *testing.T) {
	t.Parallel()

	content := exactBoundaryReport()
	issues := validateReportContent(content, validReportContext("en", true), validReportMessages("en"))
	if len(issues) != 0 {
		t.Fatalf("exact boundary report issues = %#v", issues)
	}
	if got := len([]rune(content.Summary)); got != 360 {
		t.Fatalf("summary rune count = %d", got)
	}
	if got := len(content.DimensionAssessments); got != 6 {
		t.Fatalf("dimension count = %d", got)
	}
	if got := len(content.Highlights); got != 4 {
		t.Fatalf("highlight count = %d", got)
	}
	if got := len(content.Highlights) + len(content.Issues); got != 6 {
		t.Fatalf("evidence count = %d", got)
	}
	if got := len([]rune(content.DimensionAssessments[0].Label)); got != 48 {
		t.Fatalf("dimension label rune count = %d", got)
	}
	if got := len([]rune(content.Highlights[0].Evidence)); got != 240 {
		t.Fatalf("evidence rune count = %d", got)
	}
	if got := len([]rune(content.NextActions[0].Label)); got != reportActionLabelSchemaRuneLimit {
		t.Fatalf("action label rune count = %d", got)
	}
	if got := len(content.NextActions); got != 2 {
		t.Fatalf("action count = %d", got)
	}

	issueBoundary := cloneReportContent(content)
	issueBoundary.Issues = append(issueBoundary.Issues, issueBoundary.Highlights[2:]...)
	issueBoundary.Highlights = issueBoundary.Highlights[:2]
	issueBoundary.DimensionAssessments[2].Status = sharedtypes.DimensionStatusNeedsWork
	issueBoundary.DimensionAssessments[3].Status = sharedtypes.DimensionStatusNeedsWork
	issueBoundary.RetryFocusDimensionCodes = []string{"d2", "d3", "d4", "d5"}
	if issues := validateReportContent(issueBoundary, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
		t.Fatalf("four-issue boundary report issues = %#v", issues)
	}
	if got := len(issueBoundary.Issues); got != 4 {
		t.Fatalf("issue count = %d", got)
	}

	unicodeBoundary := validChineseReport()
	unicodeBoundary.Summary = strings.Repeat("总", 360)
	unicodeBoundary.DimensionAssessments[0].Label = strings.Repeat("维", 48)
	unicodeBoundary.Highlights[0].Evidence = strings.Repeat("优", 240)
	unicodeBoundary.Issues[0].Evidence = strings.Repeat("缺", 240)
	unicodeBoundary.NextActions[0].Label = strings.Repeat("练", reportActionLabelChineseRuneLimit)
	unicodeBoundary.NextActions[1].Label = strings.Repeat("看", reportActionLabelChineseRuneLimit)
	if issues := validateReportContent(unicodeBoundary, validReportContext("zh-CN", true), validReportMessages("zh-CN")); len(issues) != 0 {
		t.Fatalf("Unicode boundary report issues = %#v", issues)
	}
	if got := len([]rune(unicodeBoundary.NextActions[0].Label)); got != reportActionLabelChineseRuneLimit {
		t.Fatalf("zh-CN action label code points = %d", got)
	}

	maxCodeBoundary := validEnglishReport()
	maxCode := "a" + strings.Repeat("b", 63)
	maxCodeBoundary.DimensionAssessments[0].Code = maxCode
	maxCodeBoundary.Issues[0].DimensionCode = maxCode
	maxCodeBoundary.RetryFocusDimensionCodes[0] = maxCode
	if issues := validateReportContent(maxCodeBoundary, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
		t.Fatalf("64-character code boundary issues = %#v", issues)
	}
}

func TestValidateGroundedReportActionLabelLanguageBounds(t *testing.T) {
	t.Parallel()

	englishAtLimit := strings.TrimSpace(strings.Repeat("word ", 24))
	if got := len(strings.Fields(englishAtLimit)); got != 24 {
		t.Fatalf("English fixture words = %d, want 24", got)
	}
	english := validEnglishReport()
	english.NextActions[0].Label = englishAtLimit
	if issues := validateReportContent(english, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
		t.Fatalf("24-word English label issues = %#v", issues)
	}

	englishOver := englishAtLimit + " word"
	english.NextActions[0].Label = englishOver
	assertReportIssue(t, validateReportContent(english, validReportContext("en", true), validReportMessages("en")), "$.nextActions[0].label", "max_words")
	if english.NextActions[0].Label != englishOver {
		t.Fatalf("validator rewrote English label: %q", english.NextActions[0].Label)
	}
	if got := len([]rune(englishOver)); got > reportActionLabelSchemaRuneLimit {
		t.Fatalf("English over-word fixture crossed schema fuse: %d runes", got)
	}

	feffSeparated := strings.Join([]string{
		"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten",
		"eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen",
		"nineteen", "twenty", "twentyone", "twentytwo", "twentythree", "twentyfour", "twentyfive",
	}, "\uFEFF")
	english.NextActions[0].Label = feffSeparated
	assertReportIssue(t, validateReportContent(english, validReportContext("en", true), validReportMessages("en")), "$.nextActions[0].label", "max_words")

	nelSeparated := strings.Join([]string{"one", "two", "three"}, "\u0085")
	if got := countEnglishWhitespaceWords(nelSeparated); got != 1 {
		t.Fatalf("ECMAScript U+0085 word count = %d, want 1", got)
	}

	chinese := validChineseReport()
	chineseAtLimit := strings.Repeat("练", 64)
	chinese.NextActions[0].Label = chineseAtLimit
	if issues := validateReportContent(chinese, validReportContext("zh-CN", true), validReportMessages("zh-CN")); len(issues) != 0 {
		t.Fatalf("64-code-point zh-CN label issues = %#v", issues)
	}

	chineseOver := chineseAtLimit + "习"
	chinese.NextActions[0].Label = chineseOver
	assertReportIssue(t, validateReportContent(chinese, validReportContext("zh-CN", true), validReportMessages("zh-CN")), "$.nextActions[0].label", "max_code_points")
	if chinese.NextActions[0].Label != chineseOver {
		t.Fatalf("validator rewrote zh-CN label: %q", chinese.NextActions[0].Label)
	}

	emoji := validChineseReport()
	emoji.NextActions[0].Label = strings.Repeat("🙂", 64)
	if issues := validateReportContent(emoji, validReportContext("zh-CN", true), validReportMessages("zh-CN")); len(issues) != 0 {
		t.Fatalf("64-code-point emoji label issues = %#v", issues)
	}
	emoji.NextActions[0].Label += "🙂"
	assertReportIssue(t, validateReportContent(emoji, validReportContext("zh-CN", true), validReportMessages("zh-CN")), "$.nextActions[0].label", "max_code_points")
}

func TestValidateGroundedReportRejectsUnknownJSONFields(t *testing.T) {
	t.Parallel()

	valid := `{"summary":"Structured answer with one gap.","preparednessLevel":"needs_practice","dimensionAssessments":[{"code":"risk_handling","label":"Risk handling","status":"needs_work","confidence":"high"}],"highlights":[],"issues":[{"dimensionCode":"risk_handling","evidence":"The rollback threshold was not concrete.","confidence":"medium","sourceMessageSeqNos":[2]}],"nextActions":[{"type":"retry_current_round","label":"Retry with a concrete threshold"}],"retryFocusDimensionCodes":["risk_handling"]}`
	tests := []struct {
		name       string
		raw        string
		path, code string
	}{
		{name: "top level unknown", raw: strings.TrimSuffix(valid, "}") + `,"score":4}`, path: "$", code: "closed_schema_invalid"},
		{name: "nested unknown", raw: strings.Replace(valid, `"confidence":"high"`, `"confidence":"high","score":4`, 1), path: "$", code: "closed_schema_invalid"},
		{name: "trailing JSON", raw: valid + `{}`, path: "$", code: "invalid_json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, issues := decodeReportContent([]byte(test.raw))
			assertReportIssue(t, issues, test.path, test.code)
		})
	}
}

func TestValidateGroundedReportRejectsBoundsAndEnums(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*ReportContentDraft)
		path   string
		code   string
	}{
		{name: "summary empty", mutate: func(c *ReportContentDraft) { c.Summary = "" }, path: "$.summary", code: "required"},
		{name: "summary 361 runes", mutate: func(c *ReportContentDraft) { c.Summary = strings.Repeat("界", 361) }, path: "$.summary", code: "max_length"},
		{name: "zero dimensions", mutate: func(c *ReportContentDraft) { c.DimensionAssessments = nil }, path: "$.dimensionAssessments", code: "min_items"},
		{name: "seven dimensions", mutate: func(c *ReportContentDraft) {
			c.DimensionAssessments = append(c.DimensionAssessments,
				DimensionAssessmentDraft{Code: "depth", Label: "Depth", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium},
				DimensionAssessmentDraft{Code: "scope", Label: "Scope", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium},
				DimensionAssessmentDraft{Code: "tradeoffs", Label: "Tradeoffs", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium},
				DimensionAssessmentDraft{Code: "delivery", Label: "Delivery", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium},
				DimensionAssessmentDraft{Code: "ownership", Label: "Ownership", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium})
		}, path: "$.dimensionAssessments", code: "max_items"},
		{name: "five highlights", mutate: func(c *ReportContentDraft) {
			for len(c.Highlights) < 5 {
				c.Highlights = append(c.Highlights, c.Highlights[0])
			}
		}, path: "$.highlights", code: "max_items"},
		{name: "five issues", mutate: func(c *ReportContentDraft) {
			for len(c.Issues) < 5 {
				c.Issues = append(c.Issues, c.Issues[0])
			}
		}, path: "$.issues", code: "max_items"},
		{name: "zero evidence", mutate: func(c *ReportContentDraft) { c.Highlights, c.Issues = nil, nil }, path: "$.evidence", code: "min_items"},
		{name: "seven evidence", mutate: func(c *ReportContentDraft) {
			*c = exactBoundaryReport()
			c.Issues = append(c.Issues, c.Issues[0])
		}, path: "$.evidence", code: "max_items"},
		{name: "zero actions", mutate: func(c *ReportContentDraft) { c.NextActions = nil }, path: "$.nextActions", code: "min_items"},
		{name: "three actions", mutate: func(c *ReportContentDraft) {
			c.NextActions = append(c.NextActions,
				ReportNextActionDraft{Type: "next_round", Label: "Continue to the next round"})
		}, path: "$.nextActions", code: "max_items"},
		{name: "seven focus codes", mutate: func(c *ReportContentDraft) {
			c.RetryFocusDimensionCodes = []string{"a0", "a1", "a2", "a3", "a4", "a5", "a6"}
		}, path: "$.retryFocusDimensionCodes", code: "max_items"},
		{name: "dimension label empty", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Label = "" }, path: "$.dimensionAssessments[0].label", code: "required"},
		{name: "dimension label 49 runes", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Label = strings.Repeat("界", 49) }, path: "$.dimensionAssessments[0].label", code: "max_length"},
		{name: "evidence empty", mutate: func(c *ReportContentDraft) { c.Highlights[0].Evidence = "" }, path: "$.highlights[0].evidence", code: "required"},
		{name: "evidence 241 runes", mutate: func(c *ReportContentDraft) { c.Highlights[0].Evidence = strings.Repeat("界", 241) }, path: "$.highlights[0].evidence", code: "max_length"},
		{name: "action label empty", mutate: func(c *ReportContentDraft) { c.NextActions[0].Label = "" }, path: "$.nextActions[0].label", code: "required"},
		{name: "action label over schema fuse", mutate: func(c *ReportContentDraft) {
			c.NextActions[0].Label = strings.Repeat("界", reportActionLabelSchemaRuneLimit+1)
		}, path: "$.nextActions[0].label", code: "max_length"},
		{name: "one character code", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Code = "a" }, path: "$.dimensionAssessments[0].code", code: "invalid_format"},
		{name: "uppercase code", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Code = "Risk" }, path: "$.dimensionAssessments[0].code", code: "invalid_format"},
		{name: "hyphen code", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Code = "risk-handling" }, path: "$.dimensionAssessments[0].code", code: "invalid_format"},
		{name: "65 character code", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Code = "a" + strings.Repeat("b", 64) }, path: "$.dimensionAssessments[0].code", code: "invalid_format"},
		{name: "unknown readiness", mutate: func(c *ReportContentDraft) { c.PreparednessLevel = "excellent" }, path: "$.preparednessLevel", code: "invalid_enum"},
		{name: "unknown dimension status", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Status = "average" }, path: "$.dimensionAssessments[0].status", code: "invalid_enum"},
		{name: "unknown dimension confidence", mutate: func(c *ReportContentDraft) { c.DimensionAssessments[0].Confidence = "certain" }, path: "$.dimensionAssessments[0].confidence", code: "invalid_enum"},
		{name: "unknown evidence confidence", mutate: func(c *ReportContentDraft) { c.Highlights[0].Confidence = "certain" }, path: "$.highlights[0].confidence", code: "invalid_enum"},
		{name: "unknown action", mutate: func(c *ReportContentDraft) { c.NextActions[0].Type = "schedule_interview" }, path: "$.nextActions[0].type", code: "invalid_enum"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := cloneReportContent(validEnglishReport())
			test.mutate(&content)
			assertReportIssue(t, validateReportContent(content, validReportContext("en", true), validReportMessages("en")), test.path, test.code)
		})
	}
}

func TestValidateGroundedReportRejectsCrossFieldViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mutate     func(*ReportContentDraft, *practicedomain.ReportContextSnapshot)
		path, code string
	}{
		{name: "duplicate dimension code", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.DimensionAssessments[1].Code = c.DimensionAssessments[0].Code
		}, path: "$.dimensionAssessments[1].code", code: "duplicate"},
		{name: "unknown evidence dimension", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.Highlights[0].DimensionCode = "unknown_dimension"
		}, path: "$.highlights[0].dimensionCode", code: "unknown_reference"},
		{name: "invalid evidence dimension code", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.Highlights[0].DimensionCode = "Risk-Handling"
		}, path: "$.highlights[0].dimensionCode", code: "invalid_format"},
		{name: "dimension without evidence", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.DimensionAssessments = append(c.DimensionAssessments, DimensionAssessmentDraft{Code: "delivery", Label: "Delivery", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium})
		}, path: "$.dimensionAssessments[2]", code: "missing_evidence"},
		{name: "strong without highlight", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.Highlights[0].DimensionCode = "risk_handling"
		}, path: "$.dimensionAssessments[1].status", code: "strong_requires_highlight"},
		{name: "needs work without issue", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.Issues[0].DimensionCode = "communication"
			c.Highlights[0].DimensionCode = "risk_handling"
		}, path: "$.dimensionAssessments[0].status", code: "needs_work_requires_issue"},
		{name: "not ready without needs work", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.PreparednessLevel = sharedtypes.ReadinessTierNotReady
			c.DimensionAssessments[0].Status = sharedtypes.DimensionStatusMeetsBar
			c.RetryFocusDimensionCodes = nil
		}, path: "$.preparednessLevel", code: "requires_needs_work"},
		{name: "needs practice without needs work", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.PreparednessLevel = sharedtypes.ReadinessTierNeedsPractice
			c.DimensionAssessments[0].Status = sharedtypes.DimensionStatusMeetsBar
			c.RetryFocusDimensionCodes = nil
		}, path: "$.preparednessLevel", code: "requires_needs_work"},
		{name: "duplicate action type", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.NextActions = append(c.NextActions, ReportNextActionDraft{Type: "retry_current_round", Label: "Retry again"})
		}, path: "$.nextActions[2].type", code: "duplicate"},
		{name: "next round without successor", mutate: func(c *ReportContentDraft, f *practicedomain.ReportContextSnapshot) {
			f.HasNextRound = false
			c.NextActions = append(c.NextActions, ReportNextActionDraft{Type: "next_round", Label: "Continue to the next round"})
		}, path: "$.nextActions[2].type", code: "next_round_unavailable"},
		{name: "lower readiness forbids next round even with successor", mutate: func(c *ReportContentDraft, f *practicedomain.ReportContextSnapshot) {
			f.HasNextRound = true
			c.NextActions = append(c.NextActions, ReportNextActionDraft{Type: "next_round", Label: "Continue to the next round"})
		}, path: "$.nextActions[2].type", code: "next_round_inconsistent_with_readiness"},
		{name: "focus without retry", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.NextActions = []ReportNextActionDraft{{Type: "review_evidence", Label: "Review the cited evidence"}}
		}, path: "$.retryFocusDimensionCodes", code: "retry_action_required"},
		{name: "unknown focus", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.RetryFocusDimensionCodes = []string{"unknown_dimension"}
		}, path: "$.retryFocusDimensionCodes[0]", code: "unknown_reference"},
		{name: "invalid focus code", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.RetryFocusDimensionCodes = []string{"Risk-Handling"}
		}, path: "$.retryFocusDimensionCodes[0]", code: "invalid_format"},
		{name: "focus is not needs work", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.RetryFocusDimensionCodes = []string{"communication"}
		}, path: "$.retryFocusDimensionCodes[0]", code: "not_needs_work"},
		{name: "focus is not issue backed", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.Issues[0].DimensionCode = "communication"
			c.Highlights[0].DimensionCode = "risk_handling"
		}, path: "$.retryFocusDimensionCodes[0]", code: "not_issue_backed"},
		{name: "duplicate focus", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.RetryFocusDimensionCodes = []string{"risk_handling", "risk_handling"}
		}, path: "$.retryFocusDimensionCodes[1]", code: "duplicate"},
		{name: "unsorted focus", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			*c = exactBoundaryReport()
			c.RetryFocusDimensionCodes = []string{"d5", "d4"}
		}, path: "$.retryFocusDimensionCodes[1]", code: "not_ascending"},
		{name: "multiple supported issues require exact focused retry", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.DimensionAssessments = append(c.DimensionAssessments, DimensionAssessmentDraft{Code: "answer_depth", Label: "Answer depth", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh})
			c.Issues = append(c.Issues, ReportEvidenceDraft{DimensionCode: "answer_depth", Evidence: "A second concrete gap was stated by the candidate.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}})
			c.RetryFocusDimensionCodes = nil
		}, path: "$.retryFocusDimensionCodes", code: "retry_focus_mismatch"},
		{name: "focused retry must cover every needs work issue", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.DimensionAssessments = append(c.DimensionAssessments, DimensionAssessmentDraft{Code: "answer_depth", Label: "Answer depth", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh})
			c.Issues = append(c.Issues, ReportEvidenceDraft{DimensionCode: "answer_depth", Evidence: "A second concrete gap was stated by the candidate.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}})
		}, path: "$.retryFocusDimensionCodes", code: "retry_focus_mismatch"},
		{name: "exact generic exception requires empty focus", mutate: func(c *ReportContentDraft, _ *practicedomain.ReportContextSnapshot) {
			c.DimensionAssessments[0].Code = "answer_depth"
			c.DimensionAssessments[0].Label = "Answer depth"
			c.Issues[0].DimensionCode = "answer_depth"
			c.RetryFocusDimensionCodes = []string{"answer_depth"}
		}, path: "$.retryFocusDimensionCodes", code: "generic_exception_requires_empty_focus"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := cloneReportContent(validEnglishReport())
			frozen := validReportContext("en", true)
			test.mutate(&content, &frozen)
			assertReportIssue(t, validateReportContent(content, frozen, validReportMessages("en")), test.path, test.code)
		})
	}

	t.Run("ordinary single issue retry rejects empty focus", func(t *testing.T) {
		content := validEnglishReport()
		content.RetryFocusDimensionCodes = nil
		assertReportIssue(t, validateReportContent(content, validReportContext("en", true), validReportMessages("en")), "$.retryFocusDimensionCodes", "retry_focus_mismatch")
	})

	for _, code := range []string{"answer_depth", "answer_relevance"} {
		t.Run("exact generic exception accepts empty focus "+code, func(t *testing.T) {
			content := validEnglishReport()
			content.DimensionAssessments[0].Code = code
			content.DimensionAssessments[0].Label = "Answer limitation"
			content.Issues[0].DimensionCode = code
			content.RetryFocusDimensionCodes = nil
			if issues := validateReportContent(content, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
				t.Fatalf("exact generic exception issues = %#v", issues)
			}
		})
	}

	t.Run("non-retry action accepts empty focus", func(t *testing.T) {
		content := validBasicallyReadyReport()
		content.NextActions = []ReportNextActionDraft{{Type: "review_evidence", Label: "Review the cited evidence"}}
		if issues := validateReportContent(content, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
			t.Fatalf("empty focus without retry issues = %#v", issues)
		}
	})
}

func TestValidateGroundedReportRejectsReadinessConfidenceCrossFieldViolations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    func() ReportContentDraft
		context    func() practicedomain.ReportContextSnapshot
		messages   func() []MessageSnapshot
		path, code string
	}{
		{
			name: "issue references strong dimension",
			content: func() ReportContentDraft {
				content := validEnglishReport()
				content.Issues[0].DimensionCode = "communication"
				return content
			},
			path: "$.issues[0].dimensionCode", code: "not_needs_work",
		},
		{
			name: "lower readiness does not put retry first",
			content: func() ReportContentDraft {
				content := validEnglishReport()
				content.NextActions[0], content.NextActions[1] = content.NextActions[1], content.NextActions[0]
				return content
			},
			path: "$.nextActions[0].type", code: "retry_current_round_required",
		},
		{
			name: "basically ready has needs work",
			content: func() ReportContentDraft {
				content := validEnglishReport()
				content.PreparednessLevel = sharedtypes.ReadinessTierBasicallyReady
				return content
			},
			path: "$.preparednessLevel", code: "forbids_needs_work",
		},
		{
			name: "basically ready lacks supported usable dimension",
			content: func() ReportContentDraft {
				content := validBasicallyReadyReport()
				for index := range content.DimensionAssessments {
					content.DimensionAssessments[index].Status = sharedtypes.DimensionStatusMeetsBar
					content.DimensionAssessments[index].Confidence = sharedtypes.ConfidenceLow
				}
				return content
			},
			path: "$.preparednessLevel", code: "requires_supported_usable_dimension",
		},
		{
			name: "well prepared has one dimension",
			content: func() ReportContentDraft {
				content := validWellPreparedReport()
				content.DimensionAssessments = content.DimensionAssessments[:1]
				content.Highlights = content.Highlights[:1]
				return content
			},
			context: validWellPreparedContext, messages: validWellPreparedMessages,
			path: "$.dimensionAssessments", code: "well_prepared_min_items",
		},
		{
			name: "well prepared has meets bar dimension",
			content: func() ReportContentDraft {
				content := validWellPreparedReport()
				content.DimensionAssessments[1].Status = sharedtypes.DimensionStatusMeetsBar
				return content
			},
			context: validWellPreparedContext, messages: validWellPreparedMessages,
			path: "$.dimensionAssessments[1].status", code: "well_prepared_requires_strong",
		},
		{
			name: "well prepared has medium confidence dimension",
			content: func() ReportContentDraft {
				content := validWellPreparedReport()
				content.DimensionAssessments[1].Confidence = sharedtypes.ConfidenceMedium
				return content
			},
			context: validWellPreparedContext, messages: validWellPreparedMessages,
			path: "$.dimensionAssessments[1].confidence", code: "well_prepared_requires_high_confidence",
		},
		{
			name: "well prepared has issue",
			content: func() ReportContentDraft {
				content := validWellPreparedReport()
				content.Issues = append(content.Issues, ReportEvidenceDraft{
					DimensionCode: "communication", Evidence: "One part could still be clearer.",
					Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2},
				})
				return content
			},
			context: validWellPreparedContext, messages: validWellPreparedMessages,
			path: "$.issues", code: "well_prepared_forbids_issues",
		},
		{
			name: "well prepared highlights cover one user message",
			content: func() ReportContentDraft {
				content := validWellPreparedReport()
				content.Highlights[1].SourceMessageSeqNos = []int32{2}
				return content
			},
			context: validWellPreparedContext, messages: validWellPreparedMessages,
			path: "$.highlights", code: "well_prepared_requires_two_user_messages",
		},
		{
			name: "strong dimension has low confidence",
			content: func() ReportContentDraft {
				content := validBasicallyReadyReport()
				content.DimensionAssessments[1].Confidence = sharedtypes.ConfidenceLow
				return content
			},
			path: "$.dimensionAssessments[1].confidence", code: "low_confidence_status_invalid",
		},
		{
			name: "needs work dimension has low confidence",
			content: func() ReportContentDraft {
				content := validEnglishReport()
				content.DimensionAssessments[0].Confidence = sharedtypes.ConfidenceLow
				return content
			},
			path: "$.dimensionAssessments[0].confidence", code: "low_confidence_status_invalid",
		},
		{
			name: "strong highlight has low confidence",
			content: func() ReportContentDraft {
				content := validBasicallyReadyReport()
				content.Highlights[1].Confidence = sharedtypes.ConfidenceLow
				return content
			},
			path: "$.highlights[1].confidence", code: "low_confidence_status_invalid",
		},
		{
			name: "issue has low confidence",
			content: func() ReportContentDraft {
				content := validEnglishReport()
				content.Issues[0].Confidence = sharedtypes.ConfidenceLow
				return content
			},
			path: "$.issues[0].confidence", code: "low_confidence_issue_invalid",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			content := test.content()
			context := validReportContext("en", true)
			if test.context != nil {
				context = test.context()
			}
			messages := validReportMessages("en")
			if test.messages != nil {
				messages = test.messages()
			}
			assertReportIssue(t, validateReportContent(content, context, messages), test.path, test.code)
		})
	}
}

func TestValidateGroundedReportAcceptsReadinessConfidenceCrossFieldBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("basically ready with one supported usable dimension", func(t *testing.T) {
		content := validBasicallyReadyReport()
		content.DimensionAssessments[0].Confidence = sharedtypes.ConfidenceLow
		content.Highlights[0].Confidence = sharedtypes.ConfidenceLow
		if issues := validateReportContent(content, validReportContext("en", true), validReportMessages("en")); len(issues) != 0 {
			t.Fatalf("basically-ready report issues = %#v", issues)
		}
	})

	t.Run("well prepared with two strong high-confidence dimensions and distinct user messages", func(t *testing.T) {
		if issues := validateReportContent(validWellPreparedReport(), validWellPreparedContext(), validWellPreparedMessages()); len(issues) != 0 {
			t.Fatalf("well-prepared report issues = %#v", issues)
		}
	})
}

func TestValidateGroundedReportRejectsInvalidAnchors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		anchors    []int32
		messages   []MessageSnapshot
		lastSeq    int32
		path, code string
	}{
		{name: "empty", anchors: nil, messages: validReportMessages("en"), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos", code: "min_items"},
		{name: "not positive", anchors: []int32{0}, messages: validReportMessages("en"), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[0]", code: "not_positive"},
		{name: "descending", anchors: []int32{4, 2}, messages: append(validReportMessages("en"), MessageSnapshot{Role: "user", Content: "Another candidate answer.", SeqNo: 4}), lastSeq: 4, path: "$.highlights[0].sourceMessageSeqNos[1]", code: "not_ascending"},
		{name: "duplicate", anchors: []int32{2, 2}, messages: validReportMessages("en"), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[1]", code: "duplicate"},
		{name: "beyond frozen last", anchors: []int32{4}, messages: append(validReportMessages("en"), MessageSnapshot{Role: "user", Content: "Another candidate answer.", SeqNo: 4}), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[0]", code: "out_of_range"},
		{name: "assistant anchor", anchors: []int32{1}, messages: validReportMessages("en"), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[0]", code: "not_user_message"},
		{name: "terminal unanswered assistant anchor", anchors: []int32{3}, messages: validReportMessages("en"), lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[0]", code: "not_user_message"},
		{name: "missing message", anchors: []int32{2}, messages: []MessageSnapshot{{Role: "assistant", Content: "Question", SeqNo: 1}, {Role: "assistant", Content: "Pending", SeqNo: 3}}, lastSeq: 3, path: "$.highlights[0].sourceMessageSeqNos[0]", code: "not_user_message"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := validEnglishReport()
			content.Highlights[0].SourceMessageSeqNos = test.anchors
			frozen := validReportContext("en", true)
			frozen.Conversation.LastMessageSeqNo = test.lastSeq
			assertReportIssue(t, validateReportContent(content, frozen, test.messages), test.path, test.code)
		})
	}
}

func TestValidateGroundedReportRejectsSafetyCopyAndLanguageViolations(t *testing.T) {
	t.Parallel()

	for _, phrase := range []string{
		"录用概率", "候选人排名", "候选人评分", "准备度评分", "击败比例", "语速", "停顿", "情绪", "人格",
		"hiring probability", "ranking", "candidate score", "readiness score", "beat other candidates", "percentage of candidates beaten", "speech rate", "pauses", "emotional state", "personality trait",
	} {
		phrase := phrase
		t.Run("forbidden_"+phrase, func(t *testing.T) {
			content := validEnglishReport()
			frozen := validReportContext("en", true)
			messages := validReportMessages("en")
			if hasHan(phrase) {
				content = validChineseReport()
				frozen = validReportContext("zh-CN", true)
				messages = validReportMessages("zh-CN")
			}
			content.Summary += " " + phrase
			assertReportIssue(t, validateReportContent(content, frozen, messages), "$.summary", "forbidden_claim")
		})
	}

	t.Run("forbidden phrase in action", func(t *testing.T) {
		content := validEnglishReport()
		content.NextActions[0].Label = "Review the candidate ranking"
		assertReportIssue(t, validateReportContent(content, validReportContext("en", true), validReportMessages("en")), "$.nextActions[0].label", "forbidden_claim")
	})

	t.Run("119 copied characters remain below threshold", func(t *testing.T) {
		copied := strings.Repeat("abcdefghij", 12)
		messages := validReportMessages("en")
		messages[1].Content = "prefix " + copied + " suffix"
		content := validEnglishReport()
		content.Summary = copied[:119]
		if issues := validateReportContent(content, validReportContext("en", true), messages); hasReportIssue(issues, "$.summary", "copied_message_text") {
			t.Fatalf("119-rune copy was rejected: %#v", issues)
		}
	})

	t.Run("120 copied characters are rejected", func(t *testing.T) {
		copied := strings.Repeat("abcdefghij", 12)
		messages := validReportMessages("en")
		messages[1].Content = "prefix " + copied + " suffix"
		content := validEnglishReport()
		content.Summary = copied
		assertReportIssue(t, validateReportContent(content, validReportContext("en", true), messages), "$.summary", "copied_message_text")
	})

	t.Run("assistant copy is also rejected", func(t *testing.T) {
		copied := strings.Repeat("qrstuvwxyz", 12)
		messages := validReportMessages("en")
		messages[0].Content = "prefix " + copied + " suffix"
		content := validEnglishReport()
		content.Summary = copied
		assertReportIssue(t, validateReportContent(content, validReportContext("en", true), messages), "$.summary", "copied_message_text")
	})

	t.Run("English report rejects Han text", func(t *testing.T) {
		content := validEnglishReport()
		content.Issues[0].Evidence = "回滚触发条件仍不够具体。"
		assertReportIssue(t, validateReportContent(content, validReportContext("en", true), validReportMessages("en")), "$.issues[0].evidence", "language_mismatch")
	})

	t.Run("English report allows a short Chinese proper noun", func(t *testing.T) {
		content := validEnglishReport()
		content.Summary = "The answer used 字节跳动 as a concise deployment example."
		if issues := validateReportContent(content, validReportContext("en", true), validReportMessages("en")); hasReportIssue(issues, "$.summary", "language_mismatch") {
			t.Fatalf("proper noun was rejected: %#v", issues)
		}
	})

	t.Run("Chinese report rejects English-only sentence", func(t *testing.T) {
		content := validChineseReport()
		content.Summary = "The rollback trigger is still missing."
		assertReportIssue(t, validateReportContent(content, validReportContext("zh-CN", true), validReportMessages("zh-CN")), "$.summary", "language_mismatch")
	})

	t.Run("Chinese report allows a technical product label", func(t *testing.T) {
		content := validChineseReport()
		content.DimensionAssessments[0].Label = "Kafka"
		if issues := validateReportContent(content, validReportContext("zh-CN", true), validReportMessages("zh-CN")); hasReportIssue(issues, "$.dimensionAssessments[0].label", "language_mismatch") {
			t.Fatalf("technical label was rejected: %#v", issues)
		}
	})

	t.Run("unsupported frozen language", func(t *testing.T) {
		assertReportIssue(t, validateReportContent(validEnglishReport(), validReportContext("fr", true), validReportMessages("en")), "$.language", "unsupported_language")
	})
}

func TestValidateGroundedReportIssuesAreStructuredAndAsable(t *testing.T) {
	t.Parallel()

	issues := ReportValidationIssues{{Path: "$.summary", Code: "forbidden_claim"}, {Path: "$.nextActions", Code: "min_items"}}
	err := fmt.Errorf("repairable output: %w", issues)
	var extracted ReportValidationIssues
	if !errors.As(err, &extracted) {
		t.Fatalf("errors.As did not extract ReportValidationIssues from %T", err)
	}
	if got, want := extracted[0], issues[0]; got != want {
		t.Fatalf("first issue = %#v, want %#v", got, want)
	}
	if got := issues.Error(); got != "$.summary:forbidden_claim,$.nextActions:min_items" {
		t.Fatalf("Error() = %q", got)
	}
}

func validEnglishReport() ReportContentDraft {
	return ReportContentDraft{
		Summary:           "The answer explained key tradeoffs, but the rollback trigger needs more detail.",
		PreparednessLevel: sharedtypes.ReadinessTierNeedsPractice,
		DimensionAssessments: []DimensionAssessmentDraft{
			{Code: "risk_handling", Label: "Risk handling", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh},
			{Code: "communication", Label: "Communication", Status: sharedtypes.DimensionStatusStrong, Confidence: sharedtypes.ConfidenceHigh},
		},
		Highlights: []ReportEvidenceDraft{{DimensionCode: "communication", Evidence: "The response explained the main tradeoffs in a clear sequence.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
		Issues:     []ReportEvidenceDraft{{DimensionCode: "risk_handling", Evidence: "The rollback trigger was mentioned without a concrete threshold.", Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}}},
		NextActions: []ReportNextActionDraft{
			{Type: "retry_current_round", Label: "Retry this round with a concrete rollback trigger"},
			{Type: "review_evidence", Label: "Review the cited evidence before retrying"},
		},
		RetryFocusDimensionCodes: []string{"risk_handling"},
	}
}

func validChineseReport() ReportContentDraft {
	return ReportContentDraft{
		Summary:           "回答清楚说明了主要取舍，但回滚触发条件仍需进一步具体化。",
		PreparednessLevel: sharedtypes.ReadinessTierNeedsPractice,
		DimensionAssessments: []DimensionAssessmentDraft{
			{Code: "risk_handling", Label: "风险处理", Status: sharedtypes.DimensionStatusNeedsWork, Confidence: sharedtypes.ConfidenceHigh},
			{Code: "communication", Label: "沟通表达", Status: sharedtypes.DimensionStatusStrong, Confidence: sharedtypes.ConfidenceHigh},
		},
		Highlights: []ReportEvidenceDraft{{DimensionCode: "communication", Evidence: "回答按清晰顺序说明了主要取舍。", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}}},
		Issues:     []ReportEvidenceDraft{{DimensionCode: "risk_handling", Evidence: "回答提到回滚，但没有给出具体触发阈值。", Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}}},
		NextActions: []ReportNextActionDraft{
			{Type: "retry_current_round", Label: "复练当前轮并补充具体回滚触发条件"},
			{Type: "review_evidence", Label: "复练前先复查报告引用的证据"},
		},
		RetryFocusDimensionCodes: []string{"risk_handling"},
	}
}

func validBasicallyReadyReport() ReportContentDraft {
	return ReportContentDraft{
		Summary:           "The answer gave a usable approach with clearly explained tradeoffs.",
		PreparednessLevel: sharedtypes.ReadinessTierBasicallyReady,
		DimensionAssessments: []DimensionAssessmentDraft{
			{Code: "risk_handling", Label: "Risk handling", Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceMedium},
			{Code: "communication", Label: "Communication", Status: sharedtypes.DimensionStatusStrong, Confidence: sharedtypes.ConfidenceHigh},
		},
		Highlights: []ReportEvidenceDraft{
			{DimensionCode: "risk_handling", Evidence: "The response described a staged rollout with monitoring.", Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}},
			{DimensionCode: "communication", Evidence: "The response explained the main tradeoffs in a clear sequence.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}},
		},
		NextActions: []ReportNextActionDraft{
			{Type: "next_round", Label: "Continue to the next round"},
			{Type: "review_evidence", Label: "Review the cited evidence"},
		},
	}
}

func validWellPreparedReport() ReportContentDraft {
	return ReportContentDraft{
		Summary:           "Two answers gave specific, consistently strong evidence across the assessed dimensions.",
		PreparednessLevel: sharedtypes.ReadinessTierWellPrepared,
		DimensionAssessments: []DimensionAssessmentDraft{
			{Code: "risk_handling", Label: "Risk handling", Status: sharedtypes.DimensionStatusStrong, Confidence: sharedtypes.ConfidenceHigh},
			{Code: "communication", Label: "Communication", Status: sharedtypes.DimensionStatusStrong, Confidence: sharedtypes.ConfidenceHigh},
		},
		Highlights: []ReportEvidenceDraft{
			{DimensionCode: "risk_handling", Evidence: "The response gave a specific staged rollout and rollback threshold.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{2}},
			{DimensionCode: "communication", Evidence: "The follow-up explained the tradeoff and decision in a clear sequence.", Confidence: sharedtypes.ConfidenceHigh, SourceMessageSeqNos: []int32{4}},
		},
		NextActions: []ReportNextActionDraft{
			{Type: "next_round", Label: "Continue to the next round"},
			{Type: "review_evidence", Label: "Review the cited evidence"},
		},
	}
}

func validWellPreparedContext() practicedomain.ReportContextSnapshot {
	context := validReportContext("en", true)
	context.Conversation.LastMessageSeqNo = 4
	return context
}

func validWellPreparedMessages() []MessageSnapshot {
	return []MessageSnapshot{
		{Role: "assistant", Content: "Explain your approach and its main tradeoffs.", SeqNo: 1},
		{Role: "user", Content: "I would use a staged rollout and roll back when errors exceed two percent.", SeqNo: 2},
		{Role: "assistant", Content: "How would you communicate that decision?", SeqNo: 3},
		{Role: "user", Content: "I would state the signal, decision owner, impact, and rollback timing in that order.", SeqNo: 4},
	}
}

func exactBoundaryReport() ReportContentDraft {
	dimensions := make([]DimensionAssessmentDraft, 0, 6)
	highlights := make([]ReportEvidenceDraft, 0, 4)
	issues := make([]ReportEvidenceDraft, 0, 2)
	for index := 0; index < 6; index++ {
		status := sharedtypes.DimensionStatusMeetsBar
		if index >= 4 {
			status = sharedtypes.DimensionStatusNeedsWork
		}
		code := fmt.Sprintf("d%d", index)
		dimensions = append(dimensions, DimensionAssessmentDraft{Code: code, Label: strings.Repeat("l", 48), Status: status, Confidence: sharedtypes.ConfidenceMedium})
		evidence := ReportEvidenceDraft{DimensionCode: code, Evidence: strings.Repeat(string(rune('a'+index)), 240), Confidence: sharedtypes.ConfidenceMedium, SourceMessageSeqNos: []int32{2}}
		if index < 4 {
			highlights = append(highlights, evidence)
		} else {
			issues = append(issues, evidence)
		}
	}
	return ReportContentDraft{
		Summary:              strings.Repeat("s", 360),
		PreparednessLevel:    sharedtypes.ReadinessTierNeedsPractice,
		DimensionAssessments: dimensions,
		Highlights:           highlights,
		Issues:               issues,
		NextActions: []ReportNextActionDraft{
			{Type: "retry_current_round", Label: strings.Repeat("r", reportActionLabelSchemaRuneLimit)},
			{Type: "review_evidence", Label: strings.Repeat("v", reportActionLabelSchemaRuneLimit)},
		},
		RetryFocusDimensionCodes: []string{"d4", "d5"},
	}
}

func validReportContext(language string, hasNextRound bool) practicedomain.ReportContextSnapshot {
	return practicedomain.ReportContextSnapshot{
		Conversation: practicedomain.ReportConversationCoordinate{Language: language, LastMessageSeqNo: 3},
		HasNextRound: hasNextRound,
	}
}

func validReportMessages(language string) []MessageSnapshot {
	if language == "zh-CN" {
		return []MessageSnapshot{
			{Role: "assistant", Content: "请说明你的方案和主要取舍。", SeqNo: 1},
			{Role: "user", Content: "我会先灰度发布，监控错误率，并在指标超过阈值时回滚。", SeqNo: 2},
			{Role: "assistant", Content: "你会如何确定回滚阈值？", SeqNo: 3},
		}
	}
	return []MessageSnapshot{
		{Role: "assistant", Content: "Explain your approach and its main tradeoffs.", SeqNo: 1},
		{Role: "user", Content: "I would use a staged rollout, monitor errors, and roll back when the threshold is exceeded.", SeqNo: 2},
		{Role: "assistant", Content: "How would you define the rollback threshold?", SeqNo: 3},
	}
}

func cloneReportContent(in ReportContentDraft) ReportContentDraft {
	out := in
	out.DimensionAssessments = append([]DimensionAssessmentDraft(nil), in.DimensionAssessments...)
	out.Highlights = cloneReportEvidence(in.Highlights)
	out.Issues = cloneReportEvidence(in.Issues)
	out.NextActions = append([]ReportNextActionDraft(nil), in.NextActions...)
	out.RetryFocusDimensionCodes = append([]string(nil), in.RetryFocusDimensionCodes...)
	return out
}

func cloneReportEvidence(in []ReportEvidenceDraft) []ReportEvidenceDraft {
	out := append([]ReportEvidenceDraft(nil), in...)
	for index := range out {
		out[index].SourceMessageSeqNos = append([]int32(nil), in[index].SourceMessageSeqNos...)
	}
	return out
}

func assertReportIssue(t *testing.T, issues []ReportValidationIssue, path, code string) {
	t.Helper()
	if !hasReportIssue(issues, path, code) {
		t.Fatalf("issues = %#v, want %s:%s", issues, path, code)
	}
}

func hasReportIssue(issues []ReportValidationIssue, path, code string) bool {
	for _, issue := range issues {
		if issue.Path == path && issue.Code == code {
			return true
		}
	}
	return false
}

func hasHan(text string) bool {
	for _, r := range text {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}
