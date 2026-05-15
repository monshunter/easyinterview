//go:build integration

package review_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestPersistReportWritesQuestionAssessments(t *testing.T) {
	db := openReviewStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ids := reviewPersistIDs("040")
	setupReviewPersistRows(t, ctx, db, ids, 1)
	repo := reviewstore.NewRepository(db)
	now := time.Date(2026, 5, 15, 17, 0, 0, 0, time.UTC)

	err := repo.PersistReport(ctx, reviewstore.PersistReportInput{
		UserID:             ids.userID,
		ReportID:           ids.reportID,
		SessionID:          ids.sessionID,
		TargetJobID:        ids.targetJobID,
		AsyncJobID:         ids.jobID,
		OutboxEventID:      ids.outboxID,
		AuditEventID:       ids.auditID,
		PreparednessLevel:  sharedtypes.ReadinessTierBasicallyReady,
		PromptVersion:      "v0.1.0",
		RubricVersion:      "v0.1.0",
		ModelID:            "model-profile:report.generate.default",
		Provider:           "stub",
		Language:           "en",
		FeatureFlag:        "none",
		DataSourceVersion:  "registry.v1",
		RetryFocusTurnIDs:  []string{ids.turnID},
		QuestionIssueCount: 1,
		Now:                now,
		Content: reviewdomain.ReportContentDraft{
			Summary:     "ready with one issue",
			Highlights:  []reviewdomain.ReportEvidenceDraft{{Dimension: "depth", Evidence: "clear summary", Confidence: 0.8}},
			Issues:      []reviewdomain.ReportEvidenceDraft{{Dimension: "depth", Evidence: "missed tradeoff", Confidence: 0.7}},
			NextActions: []reviewdomain.ReportNextActionDraft{{Type: string(reviewdomain.NextActionRetryCurrentRound), Label: "Retry tradeoff"}},
		},
		Assessments: []reviewdomain.QuestionAssessmentDraft{{
			TurnID:               ids.turnID,
			TurnIndex:            1,
			QuestionIntent:       "architecture",
			OverallStatus:        sharedtypes.DimensionStatusNeedsWork,
			Confidence:           0.7,
			Strengths:            []string{"clear structure"},
			Gaps:                 []string{"missed tradeoff"},
			RecommendedFramework: "STAR",
			ReviewStatus:         sharedtypes.QuestionReviewStatusQueuedForRetry,
			IncludedInRetryPlan:  true,
			DimensionResults: map[string]reviewdomain.DimensionResultDraft{
				"depth": {Status: sharedtypes.DimensionStatusNeedsWork, Confidence: 0.7},
			},
		}},
	})
	if err != nil {
		t.Fatalf("PersistReport: %v", err)
	}

	var status, language, featureFlag, dataSourceVersion string
	var retryRaw []byte
	if err := db.QueryRowContext(ctx, `select status, language, feature_flag, data_source_version, retry_focus_turn_ids from feedback_reports where id=$1`, ids.reportID).
		Scan(&status, &language, &featureFlag, &dataSourceVersion, &retryRaw); err != nil {
		t.Fatalf("select feedback_report: %v", err)
	}
	if status != "ready" || language != "en" || featureFlag != "none" || dataSourceVersion != "registry.v1" {
		t.Fatalf("feedback_report status/provenance = %s/%s/%s/%s", status, language, featureFlag, dataSourceVersion)
	}
	var retryIDs []string
	if err := json.Unmarshal(retryRaw, &retryIDs); err != nil || len(retryIDs) != 1 || retryIDs[0] != ids.turnID {
		t.Fatalf("retry ids raw=%s err=%v", string(retryRaw), err)
	}
	assertReviewPersistCount(t, ctx, db, `select count(*) from question_assessments where report_id=$1`, ids.reportID, 1)
	assertReviewPersistCount(t, ctx, db, `select count(*) from outbox_events where aggregate_id=$1 and event_name='report.generated'`, ids.reportID, 1)
	assertReviewPersistCount(t, ctx, db, `select count(*) from audit_events where resource_id=$1 and action='feedback_report.generated'`, ids.reportID, 1)
	var lockedAt sql.NullTime
	if err := db.QueryRowContext(ctx, `select locked_at from async_jobs where id=$1 and status='succeeded'`, ids.jobID).Scan(&lockedAt); err != nil {
		t.Fatalf("select async job: %v", err)
	}
	if lockedAt.Valid {
		t.Fatalf("locked_at after success = %v", lockedAt)
	}
}

func TestPersistReportFailureRetryAndPermanent(t *testing.T) {
	db := openReviewStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo := reviewstore.NewRepository(db)
	now := time.Date(2026, 5, 15, 18, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		suffix     string
		attempts   int
		wantStatus string
	}{
		{suffix: "050", attempts: 1, wantStatus: "queued"},
		{suffix: "060", attempts: 5, wantStatus: "failed"},
	} {
		ids := reviewPersistIDs(tc.suffix)
		setupReviewPersistRows(t, ctx, db, ids, tc.attempts)
		err := repo.PersistReportFailure(ctx, reviewstore.PersistReportFailureInput{
			UserID:        ids.userID,
			ReportID:      ids.reportID,
			SessionID:     ids.sessionID,
			AsyncJobID:    ids.jobID,
			OutboxEventID: ids.outboxID,
			AuditEventID:  ids.auditID,
			ErrorCode:     "AI_PROVIDER_TIMEOUT",
			Retryable:     true,
			Now:           now,
		})
		if err != nil {
			t.Fatalf("PersistReportFailure attempts=%d: %v", tc.attempts, err)
		}
		var jobStatus string
		var lockedAt sql.NullTime
		if err := db.QueryRowContext(ctx, `select status, locked_at from async_jobs where id=$1`, ids.jobID).Scan(&jobStatus, &lockedAt); err != nil {
			t.Fatalf("select async job: %v", err)
		}
		if jobStatus != tc.wantStatus || lockedAt.Valid {
			t.Fatalf("attempts=%d job status=%s locked_at=%v", tc.attempts, jobStatus, lockedAt)
		}
		assertReviewPersistCount(t, ctx, db, `select count(*) from outbox_events where aggregate_id=$1 and event_name='report.generation.failed'`, ids.reportID, 1)
	}
}

type reviewIDs struct {
	userID      string
	targetJobID string
	planID      string
	sessionID   string
	turnID      string
	reportID    string
	jobID       string
	outboxID    string
	auditID     string
}

func reviewPersistIDs(suffix string) reviewIDs {
	return reviewIDs{
		userID:      "0197d120-0000-7000-8000-000000000" + suffix,
		targetJobID: "0197d120-0000-7000-8000-000000001" + suffix,
		planID:      "0197d120-0000-7000-8000-000000002" + suffix,
		sessionID:   "0197d120-0000-7000-8000-000000003" + suffix,
		turnID:      "0197d120-0000-7000-8000-000000004" + suffix,
		reportID:    "0197d120-0000-7000-8000-000000005" + suffix,
		jobID:       "0197d120-0000-7000-8000-000000006" + suffix,
		outboxID:    "0197d120-0000-7000-8000-000000007" + suffix,
		auditID:     "0197d120-0000-7000-8000-000000008" + suffix,
	}
}

func setupReviewPersistRows(t *testing.T, ctx context.Context, db *sql.DB, ids reviewIDs, attempts int) {
	t.Helper()
	cleanupReviewPersistRows(t, db, ids)
	now := time.Date(2026, 5, 15, 16, 0, 0, 0, time.UTC)
	mustExecReview(t, ctx, db, `insert into users(id, email, status) values ($1, $2, 'active')`, ids.userID, ids.userID+"@example.test")
	mustExecReview(t, ctx, db, `insert into target_jobs(id, user_id, source_type, target_language, title) values ($1, $2, 'manual_text', 'en', 'Backend Engineer')`, ids.targetJobID, ids.userID)
	mustExecReview(t, ctx, db, `insert into practice_plans(id, user_id, target_job_id, goal, mode, interviewer_persona, language, time_budget_minutes, question_budget) values ($1,$2,$3,'baseline','strict','technical_manager','en',30,3)`, ids.planID, ids.userID, ids.targetJobID)
	mustExecReview(t, ctx, db, `insert into practice_sessions(id, user_id, plan_id, target_job_id, status, language, turn_count) values ($1,$2,$3,$4,'completed','en',1)`, ids.sessionID, ids.userID, ids.planID, ids.targetJobID)
	mustExecReview(t, ctx, db, `insert into practice_turns(id, session_id, turn_index, question_text, question_intent, interviewer_persona, status, asked_at) values ($1,$2,1,'Question?','architecture','technical_manager','answered',$3)`, ids.turnID, ids.sessionID, now)
	mustExecReview(t, ctx, db, `insert into feedback_reports(id, user_id, session_id, target_job_id, status, created_at, updated_at) values ($1,$2,$3,$4,'generating',$5,$5)`, ids.reportID, ids.userID, ids.sessionID, ids.targetJobID, now)
	mustExecReview(t, ctx, db, `insert into async_jobs(id, job_type, resource_type, resource_id, status, attempts, payload, available_at, locked_at, created_at, updated_at) values ($1,'report_generate','feedback_report',$2,'running',$3,'{}',$4,$4,$4,$4)`, ids.jobID, ids.reportID, attempts, now)
}

func cleanupReviewPersistRows(t *testing.T, db *sql.DB, ids reviewIDs) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.Exec(`delete from audit_events where resource_id = $1`, ids.reportID)
		_, _ = db.Exec(`delete from outbox_events where aggregate_id = $1`, ids.reportID)
		_, _ = db.Exec(`delete from async_jobs where id = $1 or resource_id = $2`, ids.jobID, ids.reportID)
		_, _ = db.Exec(`delete from question_assessments where report_id = $1`, ids.reportID)
		_, _ = db.Exec(`delete from feedback_reports where id = $1`, ids.reportID)
		_, _ = db.Exec(`delete from practice_turns where id = $1`, ids.turnID)
		_, _ = db.Exec(`delete from practice_sessions where id = $1`, ids.sessionID)
		_, _ = db.Exec(`delete from practice_plans where id = $1`, ids.planID)
		_, _ = db.Exec(`delete from target_jobs where id = $1`, ids.targetJobID)
		_, _ = db.Exec(`delete from users where id = $1`, ids.userID)
	})
}

func assertReviewPersistCount(t *testing.T, ctx context.Context, db *sql.DB, query string, arg any, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(ctx, query, arg).Scan(&got); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("count query %q = %d, want %d", query, got, want)
	}
}
