package practice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryGetSessionScopesByUserAndLoadsCurrentTurn(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	createdAt := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)

	mock.ExpectQuery(`from practice_sessions s`).
		WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count",
			"created_at", "updated_at", "turn_id", "turn_index", "question_text", "question_intent", "turn_status", "asked_at",
		}).AddRow(
			"session-1",
			"plan-1",
			"target-1",
			string(sharedtypes.SessionStatusRunning),
			"zh-CN",
			true,
			1,
			createdAt,
			updatedAt,
			"turn-1",
			1,
			"Question?",
			"behavioral",
			"asked",
			updatedAt,
		))

	session, err := repo.GetSession(context.Background(), "user-1", "session-1")
	if err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}
	if session.ID != "session-1" || session.CurrentTurn == nil || session.CurrentTurn.ID != "turn-1" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryGetSessionMapsNoRowsToNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)

	mock.ExpectQuery(`from practice_sessions s`).
		WithArgs("user-b", "session-owned-by-user-a").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetSession(context.Background(), "user-b", "session-owned-by-user-a")
	if !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("error = %v, want ErrSessionNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
