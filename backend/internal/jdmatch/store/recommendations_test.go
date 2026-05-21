package store_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

func recRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "user_id", "title", "company", "company_tag", "level", "location", "comp", "posted_label",
		"score", "fit", "reasons", "risks", "highlights", "seen", "dismissed_at", "dismiss_reason",
		"dismiss_free_note", "source_url", "source_label", "network_note", "similar_interviewers",
		"interview_hypotheses", "prompt_version", "rubric_version", "model_id", "language",
		"feature_flag", "data_source_version", "recommended_at", "updated_at", "deleted_at",
	})
}

func recRow(id string, score int, recommendedAt time.Time) []driver.Value {
	fit := []byte(`{"must":4,"total":5,"plus":3,"totalPlus":4}`)
	reasons := "{a,b}"
	risks := "{x}"
	highlights := "{h1}"
	hypos := "{ih1}"
	return []driver.Value{
		id, "user-A", "Title", "Acme", nil, nil, "Shanghai", nil, nil,
		score, fit, reasons, risks, highlights, false, nil, nil,
		nil, nil, nil, nil, nil,
		hypos, nil, nil, nil, "zh-CN",
		"none", "jd_match.v1", recommendedAt, recommendedAt, nil,
	}
}

var _ = regexp.MustCompile // keep import alive across test edits

func expectRecommendationSavedChecks(mock sqlmock.Sqlmock, userID string, ids ...string) {
	for _, id := range ids {
		mock.ExpectQuery(`select exists \(\s+select 1 from watchlist_items`).
			WithArgs(userID, id).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	}
}

func TestListRecommendationsByUserCursor(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	t0 := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	rows := recRows().
		AddRow(recRow("rec-1", 92, t0)...).
		AddRow(recRow("rec-2", 78, t0.Add(-time.Hour))...).
		AddRow(recRow("rec-3", 64, t0.Add(-2*time.Hour))...)
	mock.ExpectQuery(`SELECT .* FROM jd_match_recommendations.*WHERE user_id = \$1 AND dismissed_at IS NULL`).
		WithArgs("user-A", 3).
		WillReturnRows(rows)
	expectRecommendationSavedChecks(mock, "user-A", "rec-1", "rec-2", "rec-3")
	res, err := repo.ListRecommendationsByUser(context.Background(), "user-A", store.ListRecommendationsFilter{PageSize: 2})
	if err != nil {
		t.Fatalf("ListRecommendationsByUser: %v", err)
	}
	if len(res.Items) != 2 {
		t.Fatalf("len=%d, want 2", len(res.Items))
	}
	if !res.HasMore || res.NextCursor == "" {
		t.Fatalf("expected hasMore + cursor, got %+v", res)
	}
	// Validate scored DESC order.
	if res.Items[0].Score != 92 || res.Items[1].Score != 78 {
		t.Fatalf("order wrong: %+v", res.Items)
	}
}

func TestListRecommendationsByUserCrossUser(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT .* FROM jd_match_recommendations`).
		WithArgs("user-B", 21).
		WillReturnRows(recRows())
	res, err := repo.ListRecommendationsByUser(context.Background(), "user-B", store.ListRecommendationsFilter{PageSize: 20})
	if err != nil {
		t.Fatalf("ListRecommendationsByUser: %v", err)
	}
	if len(res.Items) != 0 || res.HasMore {
		t.Fatalf("expected empty + hasMore=false, got %+v", res)
	}
}

func TestGetRecommendationByIDForUserHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	t0 := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT .* FROM jd_match_recommendations WHERE id = \$1 AND user_id = \$2 AND deleted_at IS NULL`).
		WithArgs("rec-1", "user-A").
		WillReturnRows(recRows().AddRow(recRow("rec-1", 92, t0)...))
	rec, err := repo.GetRecommendationByIDForUser(context.Background(), "user-A", "rec-1")
	if err != nil {
		t.Fatalf("GetRecommendationByIDForUser: %v", err)
	}
	if rec.ID != "rec-1" || rec.Score != 92 {
		t.Fatalf("rec = %+v", rec)
	}
}

func TestGetRecommendationByIDForUserCrossUserNotFound(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT .* FROM jd_match_recommendations WHERE id = \$1 AND user_id = \$2`).
		WithArgs("rec-1", "user-B").
		WillReturnRows(recRows())
	_, err := repo.GetRecommendationByIDForUser(context.Background(), "user-B", "rec-1")
	if !errors.Is(err, jdmatch.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestMarkRecommendationDismissedHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	t0 := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	updated := recRow("rec-1", 92, t0)
	updated[15] = fixedNow() // dismissed_at column index in recRows
	updated[16] = "wrong_level"
	updated[17] = "too senior"
	mock.ExpectQuery(`UPDATE jd_match_recommendations`).
		WithArgs("rec-1", "user-A", fixedNow(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(recRows().AddRow(updated...))
	rec, err := repo.MarkRecommendationDismissed(context.Background(), store.MarkRecommendationDismissedInput{
		ID: "rec-1", UserID: "user-A", Reason: "wrong_level", FreeNote: "too senior",
	})
	if err != nil {
		t.Fatalf("MarkRecommendationDismissed: %v", err)
	}
	if rec.DismissedAt == nil {
		t.Fatalf("dismissedAt should be set: %+v", rec)
	}
	if rec.DismissReason == nil || *rec.DismissReason != "wrong_level" {
		t.Fatalf("dismissReason = %v", rec.DismissReason)
	}
}

func TestMarkRecommendationDismissedAlreadyDismissed(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	// UPDATE returns no rows because dismissed_at IS NOT NULL.
	mock.ExpectQuery(`UPDATE jd_match_recommendations`).
		WithArgs("rec-1", "user-A", fixedNow(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(recRows())
	// Probe reports already-dismissed.
	mock.ExpectQuery(`SELECT dismissed_at IS NOT NULL FROM jd_match_recommendations`).
		WithArgs("rec-1", "user-A").
		WillReturnRows(sqlmock.NewRows([]string{"already"}).AddRow(true))
	_, err := repo.MarkRecommendationDismissed(context.Background(), store.MarkRecommendationDismissedInput{
		ID: "rec-1", UserID: "user-A", Reason: "wrong_level",
	})
	if !errors.Is(err, jdmatch.ErrAlreadyDismissed) {
		t.Fatalf("err = %v, want ErrAlreadyDismissed", err)
	}
}

func TestMarkRecommendationDismissedNotFound(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(`UPDATE jd_match_recommendations`).
		WithArgs("rec-x", "user-A", fixedNow(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(recRows())
	// Probe returns no rows (row doesn't exist at all).
	mock.ExpectQuery(`SELECT dismissed_at IS NOT NULL FROM jd_match_recommendations`).
		WithArgs("rec-x", "user-A").
		WillReturnRows(sqlmock.NewRows([]string{"already"}))
	_, err := repo.MarkRecommendationDismissed(context.Background(), store.MarkRecommendationDismissedInput{
		ID: "rec-x", UserID: "user-A",
	})
	if !errors.Is(err, jdmatch.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteRecommendationsForUserHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM jd_match_recommendations WHERE user_id = $1`)).
		WithArgs("user-A").
		WillReturnResult(sqlmock.NewResult(0, 10))
	n, err := repo.DeleteRecommendationsForUser(context.Background(), "user-A")
	if err != nil {
		t.Fatalf("DeleteRecommendationsForUser: %v", err)
	}
	if n != 10 {
		t.Fatalf("n = %d, want 10", n)
	}
}

func TestCountActiveRecommendationsByUser(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM jd_match_recommendations WHERE user_id = $1 AND dismissed_at IS NULL AND deleted_at IS NULL`)).
		WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	n, err := repo.CountActiveRecommendationsByUser(context.Background(), "user-A")
	if err != nil {
		t.Fatalf("CountActiveRecommendationsByUser: %v", err)
	}
	if n != 7 {
		t.Fatalf("n = %d, want 7", n)
	}
}
