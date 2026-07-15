package review

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

type targetJobReportsReadRepository struct {
	overview        TargetJobReportsOverviewRecord
	conversation    ReportConversationRecord
	err             error
	conversationErr error
	gotUserID       string
	gotTargetID     string
	gotReportID     string
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

func (r *targetJobReportsReadRepository) GetReportConversation(_ context.Context, userID, reportID string) (ReportConversationRecord, error) {
	r.gotUserID = userID
	r.gotReportID = reportID
	return r.conversation, r.conversationErr
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

func TestGetReportConversationServiceUsesOnlyNormalizedOwnerIdentity(t *testing.T) {
	want := ReportConversationRecord{ReportID: idx.SampleUUIDv7, Messages: []ReportConversationMessageRecord{}}
	repository := &targetJobReportsReadRepository{conversation: want}
	service := NewService(ServiceOptions{Repository: repository})

	got, err := service.GetReportConversation(context.Background(), " user-1 ", " "+idx.SampleUUIDv7+" ")
	if err != nil {
		t.Fatalf("GetReportConversation: %v", err)
	}
	if repository.gotUserID != "user-1" || repository.gotReportID != idx.SampleUUIDv7 {
		t.Fatalf("repository identity=(%q,%q)", repository.gotUserID, repository.gotReportID)
	}
	if got.ReportID != want.ReportID || got.Messages == nil {
		t.Fatalf("conversation=%#v", got)
	}
}

func TestGetReportConversationServiceHidesMalformedReportIDWithoutRepositoryRead(t *testing.T) {
	for _, reportID := range []string{
		"report-1",
		"tmp_0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
		"00000000-0000-4000-8000-000000000000",
	} {
		repository := &targetJobReportsReadRepository{}
		service := NewService(ServiceOptions{Repository: repository})

		got, err := service.GetReportConversation(context.Background(), "user-1", reportID)
		if !errors.Is(err, ErrReportNotFound) || got.ReportID != "" || got.Messages != nil {
			t.Fatalf("reportID=%q got=%#v err=%v", reportID, got, err)
		}
		if repository.gotReportID != "" {
			t.Fatalf("reportID=%q reached repository as %q", reportID, repository.gotReportID)
		}
	}
}

func TestGetReportConversationServiceHidesMissingIdentity(t *testing.T) {
	service := NewService(ServiceOptions{Repository: &targetJobReportsReadRepository{}})
	for _, identity := range [][2]string{{"", "report-1"}, {"user-1", ""}} {
		got, err := service.GetReportConversation(context.Background(), identity[0], identity[1])
		if !errors.Is(err, ErrReportNotFound) || got.ReportID != "" || got.Messages != nil {
			t.Fatalf("identity=%#v got=%#v err=%v", identity, got, err)
		}
	}
}
