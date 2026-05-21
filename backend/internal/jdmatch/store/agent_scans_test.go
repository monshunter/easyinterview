package store_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

func fixedNow() time.Time {
	return time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
}

func newAgentScanRepo(t *testing.T) (*store.Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	repo := store.NewRepository(db, fixedNow)
	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func TestGetLatestAgentScanForUserHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	last := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	next := last.Add(4 * time.Hour)
	mock.ExpectQuery(`SELECT id, user_id, status, started_at, finished_at, last_scan_at, next_scan_at, error_message, recommendation_count, created_at, updated_at`).
		WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "started_at", "finished_at", "last_scan_at", "next_scan_at", "error_message", "recommendation_count", "created_at", "updated_at",
		}).AddRow(
			"scan-1", "user-A", "idle", nil, nil, last, next, nil, 5, last, last,
		))
	rec, err := repo.GetLatestAgentScanForUser(context.Background(), "user-A")
	if err != nil {
		t.Fatalf("GetLatestAgentScanForUser: %v", err)
	}
	if rec.Status != jdmatch.AgentScanStatusIdle {
		t.Fatalf("status = %s, want idle", rec.Status)
	}
	if rec.LastScanAt == nil || !rec.LastScanAt.Equal(last) {
		t.Fatalf("lastScanAt = %v, want %v", rec.LastScanAt, last)
	}
	if rec.NextScanAt == nil || !rec.NextScanAt.Equal(next) {
		t.Fatalf("nextScanAt = %v, want %v", rec.NextScanAt, next)
	}
	if rec.RecommendationCount != 5 {
		t.Fatalf("recommendationCount = %d, want 5", rec.RecommendationCount)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetLatestAgentScanForUserNotFound(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT id, user_id, status`).
		WithArgs("user-B").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "started_at", "finished_at", "last_scan_at", "next_scan_at", "error_message", "recommendation_count", "created_at", "updated_at",
		}))
	_, err := repo.GetLatestAgentScanForUser(context.Background(), "user-B")
	if !errors.Is(err, jdmatch.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestGetLatestAgentScanForUserCrossUser(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	// The query is bound to user_id = $1, so user-C returns 0 rows
	// despite user-A having a row. This proves cross-user isolation
	// at the SQL boundary.
	mock.ExpectQuery(`SELECT id, user_id, status`).
		WithArgs("user-C").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "started_at", "finished_at", "last_scan_at", "next_scan_at", "error_message", "recommendation_count", "created_at", "updated_at",
		}))
	if _, err := repo.GetLatestAgentScanForUser(context.Background(), "user-C"); !errors.Is(err, jdmatch.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateAgentScanHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	started := fixedNow()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_scans")).
		WithArgs("scan-1", "user-A", "scanning", &started, nil, nil, fixedNow()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	rec, err := repo.CreateAgentScan(context.Background(), store.CreateAgentScanInput{
		ID:        "scan-1",
		UserID:    "user-A",
		Status:    jdmatch.AgentScanStatusScanning,
		StartedAt: &started,
	})
	if err != nil {
		t.Fatalf("CreateAgentScan: %v", err)
	}
	if rec.ID != "scan-1" || rec.Status != jdmatch.AgentScanStatusScanning {
		t.Fatalf("rec = %#v", rec)
	}
}

func TestCreateAgentScanRejectsInvalidStatus(t *testing.T) {
	repo, _, cleanup := newAgentScanRepo(t)
	defer cleanup()
	_, err := repo.CreateAgentScan(context.Background(), store.CreateAgentScanInput{
		ID:     "scan-1",
		UserID: "user-A",
		Status: jdmatch.AgentScanStatus("bogus"),
	})
	if !errors.Is(err, jdmatch.ErrInvalidStatus) {
		t.Fatalf("err = %v, want ErrInvalidStatus", err)
	}
}

func TestUpdateAgentScanStatusHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	last := fixedNow()
	next := last.Add(4 * time.Hour)
	finished := last.Add(15 * time.Minute)
	count := 5
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE agent_scans")).
		WithArgs("scan-1", "user-A", "idle", (*time.Time)(nil), &finished, &last, &next, (*string)(nil), &count, fixedNow()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "started_at", "finished_at", "last_scan_at", "next_scan_at", "error_message", "recommendation_count", "created_at", "updated_at",
		}).AddRow(
			"scan-1", "user-A", "idle", nil, finished, last, next, nil, 5, fixedNow().Add(-time.Hour), fixedNow(),
		))
	rec, err := repo.UpdateAgentScanStatus(context.Background(), store.UpdateAgentScanStatusInput{
		ID:                  "scan-1",
		UserID:              "user-A",
		Status:              jdmatch.AgentScanStatusIdle,
		FinishedAt:          &finished,
		LastScanAt:          &last,
		NextScanAt:          &next,
		RecommendationCount: &count,
	})
	if err != nil {
		t.Fatalf("UpdateAgentScanStatus: %v", err)
	}
	if rec.Status != jdmatch.AgentScanStatusIdle || rec.RecommendationCount != 5 {
		t.Fatalf("rec = %#v", rec)
	}
}

func TestUpdateAgentScanStatusCrossUserReturnsNotFound(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE agent_scans")).
		WithArgs("scan-1", "user-B", "idle", (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*time.Time)(nil), (*string)(nil), (*int)(nil), fixedNow()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "status", "started_at", "finished_at", "last_scan_at", "next_scan_at", "error_message", "recommendation_count", "created_at", "updated_at",
		}))
	_, err := repo.UpdateAgentScanStatus(context.Background(), store.UpdateAgentScanStatusInput{
		ID:     "scan-1",
		UserID: "user-B",
		Status: jdmatch.AgentScanStatusIdle,
	})
	if !errors.Is(err, jdmatch.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDeleteAgentScansForUserHappyPath(t *testing.T) {
	repo, mock, cleanup := newAgentScanRepo(t)
	defer cleanup()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM agent_scans WHERE user_id = $1")).
		WithArgs("user-A").
		WillReturnResult(sqlmock.NewResult(0, 3))
	count, err := repo.DeleteAgentScansForUser(context.Background(), "user-A")
	if err != nil {
		t.Fatalf("DeleteAgentScansForUser: %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestRepositoryRejectsEmptyUserID(t *testing.T) {
	repo, _, cleanup := newAgentScanRepo(t)
	defer cleanup()
	if _, err := repo.GetLatestAgentScanForUser(context.Background(), "  "); !errors.Is(err, jdmatch.ErrUserIDRequired) {
		t.Fatalf("Get: err = %v, want ErrUserIDRequired", err)
	}
	if _, err := repo.CreateAgentScan(context.Background(), store.CreateAgentScanInput{ID: "x", Status: jdmatch.AgentScanStatusIdle}); !errors.Is(err, jdmatch.ErrUserIDRequired) {
		t.Fatalf("Create: err = %v, want ErrUserIDRequired", err)
	}
	if _, err := repo.UpdateAgentScanStatus(context.Background(), store.UpdateAgentScanStatusInput{ID: "x", Status: jdmatch.AgentScanStatusIdle}); !errors.Is(err, jdmatch.ErrUserIDRequired) {
		t.Fatalf("Update: err = %v, want ErrUserIDRequired", err)
	}
	if _, err := repo.DeleteAgentScansForUser(context.Background(), ""); !errors.Is(err, jdmatch.ErrUserIDRequired) {
		t.Fatalf("Delete: err = %v, want ErrUserIDRequired", err)
	}
}
