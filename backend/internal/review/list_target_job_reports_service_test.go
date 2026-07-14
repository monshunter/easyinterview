package review

import (
	"context"
	"errors"
	"testing"
)

type targetJobReportsReadRepository struct {
	overview    TargetJobReportsOverviewRecord
	err         error
	gotUserID   string
	gotTargetID string
}

func (r *targetJobReportsReadRepository) LoadReportContext(context.Context, string) (ReportContext, error) {
	return ReportContext{}, nil
}

func (r *targetJobReportsReadRepository) AssertCurrentReportJobLease(context.Context, string, int32) error {
	return nil
}

func (r *targetJobReportsReadRepository) PersistReportResult(context.Context, ReportResultPersistence) error {
	return nil
}

func (r *targetJobReportsReadRepository) PersistReportFailure(context.Context, ReportFailurePersistence) error {
	return nil
}

func (r *targetJobReportsReadRepository) GetFeedbackReport(context.Context, string, string) (FeedbackReportRecord, error) {
	return FeedbackReportRecord{}, nil
}

func (r *targetJobReportsReadRepository) ListTargetJobReports(_ context.Context, userID, targetJobID string) (TargetJobReportsOverviewRecord, error) {
	r.gotUserID = userID
	r.gotTargetID = targetJobID
	return r.overview, r.err
}

func TestListTargetJobReportsServiceUsesOnlyNormalizedOwnerIdentity(t *testing.T) {
	want := TargetJobReportsOverviewRecord{TargetJobID: "target-1", Rounds: []TargetJobReportRoundOverviewRecord{}}
	repository := &targetJobReportsReadRepository{overview: want}
	service := NewService(ServiceOptions{Repository: repository})

	got, err := service.ListTargetJobReports(context.Background(), ListTargetJobReportsRequest{
		UserID: " user-1 ", TargetJobID: " target-1 ",
	})
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if repository.gotUserID != "user-1" || repository.gotTargetID != "target-1" {
		t.Fatalf("repository identity=(%q,%q)", repository.gotUserID, repository.gotTargetID)
	}
	if got.TargetJobID != want.TargetJobID || got.Rounds == nil {
		t.Fatalf("overview=%#v", got)
	}
}

func TestListTargetJobReportsServiceHidesMissingIdentity(t *testing.T) {
	service := NewService(ServiceOptions{Repository: &targetJobReportsReadRepository{}})
	for _, request := range []ListTargetJobReportsRequest{
		{TargetJobID: "target-1"},
		{UserID: "user-1"},
	} {
		got, err := service.ListTargetJobReports(context.Background(), request)
		if !errors.Is(err, ErrReportNotFound) || got.TargetJobID != "" || got.Rounds != nil {
			t.Fatalf("request=%#v got=%#v err=%v", request, got, err)
		}
	}
}
