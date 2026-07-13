package review

import (
	"errors"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
)

func TestBuildReportActionLabelRepairPayloadPinsExactClosedEnvelopeWithoutRawLeak(t *testing.T) {
	invalidLabel := "PRIVATE-LABEL " + strings.TrimSpace(strings.Repeat("word ", 25))
	privateEvidence := "PRIVATE-EVIDENCE cited gap"
	content := ReportContentDraft{
		Issues:                   []ReportEvidenceDraft{{DimensionCode: "private_focus", Evidence: privateEvidence}},
		NextActions:              []ReportNextActionDraft{{Type: "retry_current_round", Label: invalidLabel}},
		RetryFocusDimensionCodes: []string{"private_focus"},
	}
	payload, err := BuildReportActionLabelRepairPayload(
		validReportResolution(),
		"en",
		content,
		[]ReportValidationIssue{{Path: "$.nextActions[0].label", Code: "max_words"}},
		aiclient.AITaskRunContext{},
	)
	if err != nil {
		t.Fatalf("BuildReportActionLabelRepairPayload: %v", err)
	}
	if len(payload.Messages) != 2 || payload.Messages[0].Role != "system" || payload.Messages[1].Role != "user" {
		t.Fatalf("repair messages=%#v", payload.Messages)
	}
	system := payload.Messages[0].Content
	user := payload.Messages[1].Content
	exactShape := `{"labels":[{"index":<integer copied from violation>,"label":"<rewritten label>"}]}`
	for _, want := range []string{
		exactShape,
		"one item for every violation",
		"Copy each violation index unchanged",
		"labels must be an array",
		"only index and label",
		"Do not return nextActions, explanations, or any other field",
		"at most 200 Unicode code points",
		"hard target of 4-18 whitespace-delimited words",
		"Count the label's whitespace-delimited words before returning it",
		"if it has more than 18 words, rewrite it",
		"Omit introductions, articles, and framing",
		"For multiple focus codes, use compact semicolon-separated fragments",
	} {
		if !strings.Contains(system, want) {
			t.Fatalf("trusted repair contract missing %q: %s", want, system)
		}
	}
	for _, private := range []string{invalidLabel, privateEvidence, "private_focus"} {
		if strings.Contains(system, private) {
			t.Fatalf("trusted repair contract leaked %q", private)
		}
		if !strings.Contains(user, private) {
			t.Fatalf("untrusted repair input missing %q", private)
		}
	}
}

func TestBuildReportActionLabelRepairPayloadUsesChineseRepairMarginWithoutRawLeak(t *testing.T) {
	invalidLabel := strings.Repeat("私", 41)
	content := ReportContentDraft{NextActions: []ReportNextActionDraft{{Type: "retry_current_round", Label: invalidLabel}}}
	payload, err := BuildReportActionLabelRepairPayload(
		validReportResolution(),
		"zh-CN",
		content,
		[]ReportValidationIssue{{Path: "$.nextActions[0].label", Code: "max_code_points"}},
		aiclient.AITaskRunContext{},
	)
	if err != nil {
		t.Fatalf("BuildReportActionLabelRepairPayload: %v", err)
	}
	system := payload.Messages[0].Content
	for _, want := range []string{
		"hard target of at most 52 Unicode code points",
		"Count Unicode code points before returning the label",
		"if it has more than 52 code points, rewrite it",
		"Omit introductions and framing",
		"For multiple focus codes, use compact semicolon-separated fragments",
	} {
		if !strings.Contains(system, want) {
			t.Fatalf("zh-CN repair margin missing %q: %s", want, system)
		}
	}
	if strings.Contains(system, invalidLabel) || !strings.Contains(payload.Messages[1].Content, invalidLabel) {
		t.Fatal("zh-CN invalid label crossed the trusted/untrusted boundary")
	}
}

func TestReportActionLabelRepairMarginsStayInsideProductLimits(t *testing.T) {
	if reportActionLabelRepairEnglishWordTarget != 18 || reportActionLabelRepairChineseRuneTarget != 52 {
		t.Fatalf("repair margins en=%d zh=%d", reportActionLabelRepairEnglishWordTarget, reportActionLabelRepairChineseRuneTarget)
	}
	if reportActionLabelEnglishWordLimit != 24 || reportActionLabelChineseRuneLimit != 64 {
		t.Fatalf("product limits en=%d zh=%d", reportActionLabelEnglishWordLimit, reportActionLabelChineseRuneLimit)
	}
	if reportActionLabelRepairEnglishWordTarget >= reportActionLabelEnglishWordLimit || reportActionLabelRepairChineseRuneTarget >= reportActionLabelChineseRuneLimit {
		t.Fatalf("repair margins must remain below product limits: repair=%d/%d product=%d/%d", reportActionLabelRepairEnglishWordTarget, reportActionLabelRepairChineseRuneTarget, reportActionLabelEnglishWordLimit, reportActionLabelChineseRuneLimit)
	}
	if reportActionLabelSchemaRuneLimit != 200 {
		t.Fatalf("wire fuse=%d, want 200", reportActionLabelSchemaRuneLimit)
	}
}

func TestMergeReportActionLabelRepairRejectsWrongEnvelopeAndReplacementSet(t *testing.T) {
	content := ReportContentDraft{NextActions: []ReportNextActionDraft{
		{Type: "retry_current_round", Label: strings.TrimSpace(strings.Repeat("word ", 25))},
		{Type: "review_evidence", Label: strings.TrimSpace(strings.Repeat("evidence ", 25))},
	}}
	issues := []ReportValidationIssue{
		{Path: "$.nextActions[0].label", Code: "max_words"},
		{Path: "$.nextActions[1].label", Code: "max_words"},
	}
	tests := []struct {
		name, response, code string
	}{
		{
			name:     "nextActions envelope",
			response: `{"nextActions":[{"index":0,"label":"Retry with cited rollback steps"}]}`,
			code:     "closed_schema_invalid",
		},
		{
			name:     "missing replacement",
			response: `{"labels":[{"index":0,"label":"Retry with cited rollback steps"}]}`,
			code:     "replacement_set_mismatch",
		},
		{
			name:     "duplicate replacement index",
			response: `{"labels":[{"index":0,"label":"Retry with cited rollback steps"},{"index":0,"label":"Review the cited positive evidence"}]}`,
			code:     "replacement_set_mismatch",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := MergeReportActionLabelRepair(content, "en", issues, tc.response)
			var invalid *ReportContentInvalidError
			if err == nil || !strings.Contains(err.Error(), tc.code) || !errors.As(err, &invalid) {
				t.Fatalf("error=%v invalid=%+v, want code %s", err, invalid, tc.code)
			}
			if strings.Contains(err.Error(), content.NextActions[0].Label) {
				t.Fatal("repair failure leaked the raw invalid label")
			}
		})
	}
}
