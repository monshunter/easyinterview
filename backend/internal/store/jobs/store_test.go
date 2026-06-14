package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestRepositoryGetJobScopesAsyncJobToOwningUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 17, 10, 45, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)from async_jobs j.*resource_type = 'resume_asset'.*from resumes rs.*resource_type = 'resume_tailor_run'.*payload->>'resumeId'.*resource_type = 'debrief'.*debriefs d.*d.user_id = \$2`).
		WithArgs("job-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at",
		}).AddRow(
			"job-1",
			"debrief_generate",
			"debrief",
			"debrief-1",
			string(sharedtypes.JobStatusQueued),
			nil,
			now,
			now,
		))

	got, err := repo.GetJob(context.Background(), "user-1", "job-1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.ID != "job-1" || got.ResourceID != "debrief-1" || got.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("job drifted: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRepositoryGetJobReturnsNotFoundWhenUserDoesNotOwnResource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	repo := NewRepository(db)

	mock.ExpectQuery(`from async_jobs j`).
		WithArgs("job-1", "other-user").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at",
		}))

	_, err = repo.GetJob(context.Background(), "other-user", "job-1")
	if !errors.Is(err, domain.ErrJobNotFound) {
		t.Fatalf("err = %v, want ErrJobNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
