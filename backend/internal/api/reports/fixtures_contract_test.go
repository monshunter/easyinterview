package reports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReportsFixturesIncludeFailureAndEmptyVariants(t *testing.T) {
	reportFixture := loadJSONFixture(t, "openapi/fixtures/Reports/getFeedbackReport.json")
	reportFailed := lookupMap(t, reportFixture, "scenarios", "report-failed", "response")
	if status := reportFailed["status"]; status != float64(200) {
		t.Fatalf("report-failed response status = %v, want 200", status)
	}
	reportBody := lookupMap(t, reportFailed, "body")
	if got := reportBody["status"]; got != "failed" {
		t.Fatalf("report-failed body.status = %v, want failed", got)
	}
	if got := reportBody["errorCode"]; got != "AI_PROVIDER_TIMEOUT" {
		t.Fatalf("report-failed errorCode = %v, want AI_PROVIDER_TIMEOUT", got)
	}
	if got := reportBody["provenance"]; got != nil {
		t.Fatalf("report-failed provenance = %v, want nil", got)
	}
	for _, key := range []string{"highlights", "issues", "nextActions", "questionAssessments", "retryFocusTurnIds"} {
		values, ok := reportBody[key].([]any)
		if !ok || len(values) != 0 {
			t.Fatalf("report-failed %s = %#v, want empty array", key, reportBody[key])
		}
	}

	listFixture := loadJSONFixture(t, "openapi/fixtures/Reports/listTargetJobReports.json")
	empty := lookupMap(t, listFixture, "scenarios", "empty", "response", "body")
	items, ok := empty["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("empty list items = %#v, want empty array", empty["items"])
	}
	pageInfo := mapValue(t, empty, "pageInfo")
	if got := pageInfo["hasMore"]; got != false {
		t.Fatalf("empty pageInfo.hasMore = %v, want false", got)
	}
	if got := pageInfo["nextCursor"]; got != nil {
		t.Fatalf("empty pageInfo.nextCursor = %v, want nil", got)
	}
}

func loadJSONFixture(t *testing.T, rel string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse %s: %v", rel, err)
	}
	return doc
}
