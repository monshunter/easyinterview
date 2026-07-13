package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreateTailorRunInsertsAsyncJobWithResumePayload(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select 1 from resumes`)).
		WithArgs("resume-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`select 1 from target_jobs`)).
		WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WithArgs(
			"job-1", "resume_tailor", "resume_tailor_run", "tailor-1", "dedupe-1",
			string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := repo.CreateTailorRun(context.Background(), resumestore.CreateTailorRunInput{
		TailorRunID: "tailor-1",
		JobID:       "job-1",
		UserID:      "user-1",
		TargetJobID: "target-1",
		ResumeID:    "resume-1",
		Mode:        "gap_review",
		DedupeKey:   "dedupe-1",
		Now:         now,
	})
	if err != nil {
		t.Fatalf("CreateTailorRun: %v", err)
	}
	if got.TailorRunID != "tailor-1" || got.JobID != "job-1" || got.JobStatus != sharedtypes.JobStatusQueued {
		t.Fatalf("result = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateTailorRunWithoutTargetJobSkipsTargetCheck(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select 1 from resumes`)).
		WithArgs("resume-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WithArgs(
			"job-1", "resume_tailor", "resume_tailor_run", "tailor-1", nil,
			string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if _, err := repo.CreateTailorRun(context.Background(), resumestore.CreateTailorRunInput{
		TailorRunID: "tailor-1",
		JobID:       "job-1",
		UserID:      "user-1",
		ResumeID:    "resume-1",
		Mode:        "bullet_suggestions",
		Now:         now,
	}); err != nil {
		t.Fatalf("CreateTailorRun: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateTailorRunResumeNotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select 1 from resumes`)).
		WithArgs("resume-1", "user-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.CreateTailorRun(context.Background(), resumestore.CreateTailorRunInput{
		TailorRunID: "tailor-1",
		JobID:       "job-1",
		UserID:      "user-1",
		ResumeID:    "resume-1",
		Mode:        "gap_review",
		Now:         time.Now(),
	})
	if !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("err = %v, want ErrAssetNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetTailorRunMapsStatusFromAsyncJob(t *testing.T) {
	tests := []struct {
		name       string
		jobStatus  string
		result     []byte
		wantStatus string
		wantReady  bool
	}{
		{name: "queued", jobStatus: "queued", result: []byte(`{}`), wantStatus: "queued"},
		{name: "running", jobStatus: "running", result: []byte(`{}`), wantStatus: "generating"},
		{name: "failed", jobStatus: "failed", result: []byte(`{}`), wantStatus: "failed"},
		{
			name:       "ready with result",
			jobStatus:  "succeeded",
			result:     []byte(`{"matchSummary":{"strengths":["Go"],"gaps":["k8s"]},"suggestions":[{"originalBullet":"a","suggestedBullet":"b","reason":"impact"}],"provenance":{"promptVersion":"p","rubricVersion":"r","modelId":"m","provider":"provider","language":"en","featureFlag":"flag","dataSourceVersion":"source"}}`),
			wantStatus: "ready",
			wantReady:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepository(t)
			defer cleanup()
			now := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)
			mock.ExpectQuery(regexp.QuoteMeta(`select aj.resource_id, rs.user_id, aj.status, aj.payload, aj.result`)).
				WithArgs("tailor-1", "user-1", "resume_tailor_run").
				WillReturnRows(sqlmock.NewRows([]string{
					"resource_id", "user_id", "status", "payload", "result", "error_code", "created_at", "updated_at",
				}).AddRow(
					"tailor-1", "user-1", tc.jobStatus,
					[]byte(`{"resumeId":"resume-1","targetJobId":"target-1","mode":"gap_review"}`),
					tc.result, nil, now, now,
				))

			got, err := repo.GetTailorRun(context.Background(), "user-1", "tailor-1")
			if err != nil {
				t.Fatalf("GetTailorRun: %v", err)
			}
			if got.Status != tc.wantStatus {
				t.Fatalf("status = %q, want %q", got.Status, tc.wantStatus)
			}
			if got.ResumeID != "resume-1" || got.TargetJobID != "target-1" || got.Mode != "gap_review" {
				t.Fatalf("run = %+v", got)
			}
			if tc.wantReady {
				if string(got.MatchSummary) == "" || string(got.MatchSummary) == "{}" {
					t.Fatalf("match summary = %s", got.MatchSummary)
				}
				var suggestions []map[string]string
				if err := json.Unmarshal(got.Suggestions, &suggestions); err != nil || len(suggestions) != 1 {
					t.Fatalf("suggestions = %s (err=%v)", got.Suggestions, err)
				}
				wantProvenance := resumestore.VersionProvenance{
					PromptVersion:     "p",
					RubricVersion:     "r",
					ModelID:           "m",
					Provider:          "provider",
					Language:          "en",
					FeatureFlag:       "flag",
					DataSourceVersion: "source",
				}
				if got.Provenance != wantProvenance {
					t.Fatalf("provenance = %+v, want %+v", got.Provenance, wantProvenance)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("sql expectations: %v", err)
			}
		})
	}
}

func TestGetTailorRunCrossUserNotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectQuery(regexp.QuoteMeta(`select aj.resource_id, rs.user_id, aj.status, aj.payload, aj.result`)).
		WithArgs("tailor-1", "user-2", "resume_tailor_run").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetTailorRun(context.Background(), "user-2", "tailor-1")
	if !errors.Is(err, resumestore.ErrTailorRunNotFound) {
		t.Fatalf("err = %v, want ErrTailorRunNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCompleteTailorRunSuccessWritesResultAndOutbox(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 11, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*where id = \$1.*resource_type = \$2.*resource_id = \$3.*status = 'running'.*attempts = \$4.*for update`).
		WithArgs("job-1", "resume_tailor_run", "tailor-1", int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"attempts"}).AddRow(2))
	mock.ExpectExec(regexp.QuoteMeta(`update async_jobs`)).
		WithArgs(sqlmock.AnyArg(), now, "job-1", "resume_tailor_run", "tailor-1", int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into outbox_events`)).
		WithArgs("event-1", "resume.tailor.completed", "resume", "resume-1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.CompleteTailorRunSuccess(context.Background(), resumestore.CompleteTailorRunSuccessInput{
		JobID:           "job-1",
		ClaimedAttempts: 2,
		TailorRunID:     "tailor-1",
		ResumeID:        "resume-1",
		TargetJobID:     "target-1",
		Mode:            "gap_review",
		MatchSummary:    []byte(`{"strengths":["Go"],"gaps":[]}`),
		Suggestions: []resumestore.TailorSuggestionInput{
			{ID: "sug-1", OriginalBullet: "a", SuggestedBullet: "b", Reason: "impact"},
		},
		Provenance:         resumestore.VersionProvenance{PromptVersion: "p", ModelID: "m"},
		OutboxEventID:      "event-1",
		OutboxEventPayload: []byte(`{"tailorRunId":"tailor-1","resumeId":"resume-1","targetJobId":"target-1","mode":"gap_review","status":"ready"}`),
		Now:                now,
	}); err != nil {
		t.Fatalf("CompleteTailorRunSuccess: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCompleteTailorRunSuccessInvalidStateTransition(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 11, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)select attempts.*from async_jobs.*for update`).
		WithArgs("job-1", "resume_tailor_run", "tailor-1", int32(1)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := repo.CompleteTailorRunSuccess(context.Background(), resumestore.CompleteTailorRunSuccessInput{
		JobID:              "job-1",
		ClaimedAttempts:    1,
		TailorRunID:        "tailor-1",
		ResumeID:           "resume-1",
		OutboxEventID:      "event-1",
		OutboxEventPayload: []byte(`{}`),
		Now:                now,
	})
	if !errors.Is(err, runner.ErrStaleLease) {
		t.Fatalf("err = %v, want ErrStaleLease", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
