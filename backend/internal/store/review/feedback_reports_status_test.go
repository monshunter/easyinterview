package review_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestStatusStateMachineEnforcement(t *testing.T) {
	now := time.Date(2026, 5, 15, 13, 30, 0, 0, time.UTC)
	db, mock, cleanup := newMockReviewStore(t)
	defer cleanup()
	repo := reviewstore.NewRepository(db)

	for _, tc := range []struct {
		name string
		from sharedtypes.ReportStatus
		to   sharedtypes.ReportStatus
	}{
		{name: "queued_to_ready", from: sharedtypes.ReportStatusQueued, to: sharedtypes.ReportStatusReady},
		{name: "failed_to_ready", from: sharedtypes.ReportStatusFailed, to: sharedtypes.ReportStatusReady},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.UpdateFeedbackReportStatus(context.Background(), reviewdomain.ReportStatusUpdate{
				ReportID: "0197d120-0000-7000-8000-000000000010",
				From:     tc.from,
				To:       tc.to,
				Now:      now,
			})
			if !errors.Is(err, reviewdomain.ErrIllegalTransition) {
				t.Fatalf("err = %v, want ErrIllegalTransition", err)
			}
		})
	}

	allowed := []struct {
		from sharedtypes.ReportStatus
		to   sharedtypes.ReportStatus
	}{
		{from: sharedtypes.ReportStatusQueued, to: sharedtypes.ReportStatusGenerating},
		{from: sharedtypes.ReportStatusGenerating, to: sharedtypes.ReportStatusReady},
		{from: sharedtypes.ReportStatusGenerating, to: sharedtypes.ReportStatusFailed},
		{from: sharedtypes.ReportStatusFailed, to: sharedtypes.ReportStatusQueued},
	}
	for _, tc := range allowed {
		mock.ExpectExec(regexp.QuoteMeta(`
update feedback_reports
set status = $1,
    updated_at = $2
where id = $3 and status = $4`)).
			WithArgs(string(tc.to), now, "0197d120-0000-7000-8000-000000000011", string(tc.from)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		if err := repo.UpdateFeedbackReportStatus(context.Background(), reviewdomain.ReportStatusUpdate{
			ReportID: "0197d120-0000-7000-8000-000000000011",
			From:     tc.from,
			To:       tc.to,
			Now:      now,
		}); err != nil {
			t.Fatalf("%s->%s err = %v", tc.from, tc.to, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func newMockReviewStore(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock, func() { _ = db.Close() }
}
