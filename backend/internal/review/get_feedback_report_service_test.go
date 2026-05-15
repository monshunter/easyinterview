package review

import (
	"context"
	"errors"
	"testing"
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGetFeedbackReportServiceUserScopedNotFound(t *testing.T) {
	repo := &fakeReadRepository{getErr: ErrReportNotFound}
	svc := NewService(ServiceOptions{Repository: repo})

	_, err := svc.GetFeedbackReport(context.Background(), "user-1", "report-1")
	if !errors.Is(err, ErrReportNotFound) {
		t.Fatalf("err = %v, want ErrReportNotFound", err)
	}
	if repo.getUserID != "user-1" || repo.getReportID != "report-1" {
		t.Fatalf("request = %s/%s", repo.getUserID, repo.getReportID)
	}
}

func TestListTargetJobReportsServicePagination(t *testing.T) {
	createdAt := time.Date(2026, 5, 15, 10, 20, 30, 0, time.UTC)
	report := sampleFeedbackReportRecord(createdAt)
	repo := &fakeReadRepository{listResult: ListTargetJobReportsResult{
		Items:      []FeedbackReportRecord{report},
		NextCursor: EncodeCursor(createdAt, report.ID),
		HasMore:    true,
		PageSize:   50,
	}}
	svc := NewService(ServiceOptions{Repository: repo})

	got, err := svc.ListTargetJobReports(context.Background(), ListTargetJobReportsRequest{
		UserID:      "user-1",
		TargetJobID: "target-1",
		PageSize:    500,
	})
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if repo.listInput.UserID != "user-1" || repo.listInput.TargetJobID != "target-1" || repo.listInput.PageSize != 50 {
		t.Fatalf("list input = %+v", repo.listInput)
	}
	if !got.PageInfo.HasMore || got.PageInfo.PageSize != 50 || got.PageInfo.NextCursor == "" || len(got.Items) != 1 {
		t.Fatalf("list result = %+v", got)
	}
}

func sampleFeedbackReportRecord(createdAt time.Time) FeedbackReportRecord {
	tier := sharedtypes.ReadinessTierBasicallyReady
	errorCode := ""
	return FeedbackReportRecord{
		ID:                  "0197d120-0000-7000-8000-000000000501",
		SessionID:           "0197d120-0000-7000-8000-000000000502",
		TargetJobID:         "0197d120-0000-7000-8000-000000000503",
		Status:              sharedtypes.ReportStatusReady,
		PreparednessLevel:   &tier,
		Highlights:          []ReportEvidenceRecord{{Dimension: "depth", Evidence: "clear", Confidence: sharedtypes.ConfidenceHigh}},
		Issues:              []ReportEvidenceRecord{},
		NextActions:         []ReportNextActionRecord{{Type: string(NextActionNextRound), Label: "Next round"}},
		QuestionAssessments: []QuestionAssessmentRecord{{TurnID: "0197d120-0000-7000-8000-000000000504", QuestionIntent: "architecture", DimensionResults: map[string]DimensionResultRecord{"depth": {Status: sharedtypes.DimensionStatusMeetsBar, Confidence: sharedtypes.ConfidenceHigh}}, ReviewStatus: sharedtypes.QuestionReviewStatusOpen}},
		RetryFocusTurnIDs:   []string{},
		Provenance:          &GenerationProvenanceRecord{PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelID: "model-profile:report.generate.default", Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1"},
		ErrorCode:           &errorCode,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
	}
}

type fakeReadRepository struct {
	fakeReportRepository
	getReport   FeedbackReportRecord
	getErr      error
	getUserID   string
	getReportID string
	listInput   ListTargetJobReportsInput
	listResult  ListTargetJobReportsResult
	listErr     error
}

func (f *fakeReadRepository) GetFeedbackReport(ctx context.Context, userID, reportID string) (FeedbackReportRecord, error) {
	f.getUserID = userID
	f.getReportID = reportID
	return f.getReport, f.getErr
}

func (f *fakeReadRepository) ListTargetJobReports(ctx context.Context, in ListTargetJobReportsInput) (ListTargetJobReportsResult, error) {
	f.listInput = in
	return f.listResult, f.listErr
}
