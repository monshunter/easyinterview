package review

import (
	"strings"
	"testing"
)

func TestDecodeReportContentRejectsClosedShapeAndRequiredTypeDrift(t *testing.T) {
	valid := validDirectReportJSON("en")
	dimensionsArray := `[{"code":"technical_depth","label":"Technical depth","status":"needs_work","confidence":"high"}]`
	issuesArray := `[{"dimensionCode":"technical_depth","evidence":"The candidate explained queue backpressure but did not provide concrete rollback steps.","confidence":"high","sourceMessageSeqNos":[2]}]`
	actionsArray := `[{"type":"retry_current_round","label":"Add executable rollback steps and replay this round"}]`
	tests := []struct {
		name string
		raw  string
		path string
		code string
	}{
		{name: "empty", raw: "", path: "$", code: "empty_json"},
		{name: "trailing", raw: valid + `{}`, path: "$", code: "invalid_json"},
		{name: "unknown top field", raw: strings.TrimSuffix(valid, "}") + `,"unexpected":true}`, path: "$", code: "closed_schema_invalid"},
		{name: "unknown nested field", raw: strings.Replace(valid, `"confidence":"high"}`, `"confidence":"high","score":5}`, 1), path: "$", code: "closed_schema_invalid"},
		{name: "missing required focus", raw: strings.Replace(valid, `,"retryFocusDimensionCodes":["technical_depth"]`, "", 1), path: "$.retryFocusDimensionCodes", code: "required"},
		{name: "null summary", raw: strings.Replace(valid, `"summary":"The answer explained key tradeoffs, but the rollback plan still needs concrete steps."`, `"summary":null`, 1), path: "$.summary", code: "invalid_type"},
		{name: "null preparedness", raw: strings.Replace(valid, `"preparednessLevel":"needs_practice"`, `"preparednessLevel":null`, 1), path: "$.preparednessLevel", code: "invalid_type"},
		{name: "null dimensions", raw: strings.Replace(valid, `"dimensionAssessments":`+dimensionsArray, `"dimensionAssessments":null`, 1), path: "$.dimensionAssessments", code: "invalid_type"},
		{name: "null highlights", raw: strings.Replace(valid, `"highlights":[]`, `"highlights":null`, 1), path: "$.highlights", code: "invalid_type"},
		{name: "null issues", raw: strings.Replace(valid, `"issues":`+issuesArray, `"issues":null`, 1), path: "$.issues", code: "invalid_type"},
		{name: "null actions", raw: strings.Replace(valid, `"nextActions":`+actionsArray, `"nextActions":null`, 1), path: "$.nextActions", code: "invalid_type"},
		{name: "null focus", raw: strings.Replace(valid, `"retryFocusDimensionCodes":["technical_depth"]`, `"retryFocusDimensionCodes":null`, 1), path: "$.retryFocusDimensionCodes", code: "invalid_type"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, issues := decodeReportContent([]byte(tc.raw))
			if !hasReportIssue(issues, tc.path, tc.code) {
				t.Fatalf("issues=%#v want %s:%s", issues, tc.path, tc.code)
			}
		})
	}
}

func TestDecodeReportContentAcceptsExplicitEmptyGenericFocusArray(t *testing.T) {
	raw := strings.Replace(validDirectReportJSON("en"), `"retryFocusDimensionCodes":["technical_depth"]`, `"retryFocusDimensionCodes":[]`, 1)
	content, issues := decodeReportContent([]byte(raw))
	if len(issues) != 0 {
		t.Fatalf("issues=%#v", issues)
	}
	if content.RetryFocusDimensionCodes == nil || len(content.RetryFocusDimensionCodes) != 0 {
		t.Fatalf("explicit empty focus was not preserved: %#v", content.RetryFocusDimensionCodes)
	}
}
