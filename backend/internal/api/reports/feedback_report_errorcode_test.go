package reports

import (
	"encoding/json"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestFeedbackReportSchemaIncludesErrorCode(t *testing.T) {
	code := generated.ApiErrorCodeAIPROVIDERTIMEOUT
	report := generated.FeedbackReport{
		Id:          "01918fa0-0000-7000-8000-00000000a001",
		SessionId:   "01918fa0-0000-7000-8000-00000000a002",
		TargetJobId: "01918fa0-0000-7000-8000-00000000a003",
		Status:      types.ReportStatusFailed,
		ErrorCode:   &code,
		CreatedAt:   "2026-05-15T00:00:00Z",
		UpdatedAt:   "2026-05-15T00:00:00Z",
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal FeedbackReport: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal FeedbackReport: %v", err)
	}
	if got["errorCode"] != string(code) {
		t.Fatalf("errorCode = %v, want %s", got["errorCode"], code)
	}
}
