package review

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const targetJobReportsTargetQuery = `(?s)select tj\.summary\s+from target_jobs tj\s+where tj\.id = \$1 and tj\.user_id = \$2 and tj\.deleted_at is null`

const targetJobReportsRowsQuery = `(?s)select fr\.id::text, fr\.user_id::text, fr\.session_id::text, fr\.target_job_id::text,\s+ps\.user_id::text, ps\.target_job_id::text, fr\.status, fr\.error_code,\s+fr\.generation_context, fr\.generated_at, fr\.created_at\s+from feedback_reports fr\s+left join practice_sessions ps on ps\.id = fr\.session_id\s+where fr\.target_job_id = \$1 or ps\.target_job_id = \$1`

func TestListTargetJobReportsOverviewEnumeratesCanonicalRoundsAndSelectsIndependentPointers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	summary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "technical", Name: "Technical", Focus: "design", Duration: 45},
		targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Manager", Focus: "ownership", Duration: 30},
		targetJobReportsRound{Sequence: 3, Type: "culture", Name: "Culture", Focus: "values", Duration: 25},
	)

	mock.ExpectBegin()
	mock.ExpectQuery(targetJobReportsTargetQuery).
		WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"summary"}).AddRow(summary))
	rows := targetJobReportsRows()
	rows.AddRow("report-ready-old", "user-1", "session-ready-old", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, summary, 0, "session-ready-old"), now.Add(2*time.Minute), now)
	rows.AddRow("report-failed-new", "user-1", "session-failed-new", "target-1", "user-1", "target-1", "failed", sharederrors.CodeAiOutputInvalid,
		targetJobReportsFrozenContext(t, summary, 0, "session-failed-new"), nil, now.Add(4*time.Minute))
	rows.AddRow("report-generating", "user-1", "session-generating", "target-1", "user-1", "target-1", "generating", "STALE_INTERNAL_CODE",
		targetJobReportsFrozenContext(t, summary, 1, "session-generating"), nil, now.Add(3*time.Minute))
	rows.AddRow("report-ready-latest", "user-1", "session-ready-latest", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, summary, 2, "session-ready-latest"), now.Add(5*time.Minute), now.Add(5*time.Minute))
	mock.ExpectQuery(targetJobReportsRowsQuery).WithArgs("target-1").WillReturnRows(rows)
	mock.ExpectCommit()

	got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if got.TargetJobID != "target-1" || len(got.Rounds) != 3 {
		t.Fatalf("overview=%#v", got)
	}
	if got.Rounds[0].Round.RoundID != "round-1-technical" || got.Rounds[0].CurrentReport == nil || got.Rounds[0].CurrentReport.ID != "report-ready-old" || got.Rounds[0].LatestAttempt == nil || got.Rounds[0].LatestAttempt.ID != "report-failed-new" || got.Rounds[0].LatestAttempt.ErrorCode == nil || *got.Rounds[0].LatestAttempt.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("round 1 selection=%#v", got.Rounds[0])
	}
	if got.Rounds[1].CurrentReport != nil || got.Rounds[1].LatestAttempt == nil || got.Rounds[1].LatestAttempt.ID != "report-generating" || got.Rounds[1].LatestAttempt.ErrorCode != nil {
		t.Fatalf("round 2 selection=%#v", got.Rounds[1])
	}
	if got.Rounds[2].CurrentReport == nil || got.Rounds[2].LatestAttempt == nil || got.Rounds[2].CurrentReport.ID != "report-ready-latest" || got.Rounds[2].LatestAttempt.ID != "report-ready-latest" {
		t.Fatalf("round 3 selection=%#v", got.Rounds[2])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTargetJobReportsOverviewUsesDeterministicIndependentTieBreaks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC)
	summary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "technical", Name: "Technical", Focus: "design", Duration: 45},
		targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Manager", Focus: "ownership", Duration: 30},
	)

	mock.ExpectBegin()
	mock.ExpectQuery(targetJobReportsTargetQuery).WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"summary"}).AddRow(summary))
	rows := targetJobReportsRows()
	rows.AddRow("report-b", "user-1", "session-b", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, summary, 0, "session-b"), now.Add(2*time.Minute), now)
	rows.AddRow("report-a", "user-1", "session-a", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, summary, 0, "session-a"), now.Add(2*time.Minute), now)
	rows.AddRow("report-c", "user-1", "session-c", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, summary, 0, "session-c"), now.Add(time.Minute), now.Add(3*time.Minute))
	rows.AddRow("report-d", "user-1", "session-d", "target-1", "user-1", "target-1", "failed", sharederrors.CodeAiOutputInvalid,
		targetJobReportsFrozenContext(t, summary, 0, "session-d"), nil, now.Add(4*time.Minute))
	rows.AddRow("report-e", "user-1", "session-e", "target-1", "user-1", "target-1", "failed", sharederrors.CodeAiOutputInvalid,
		targetJobReportsFrozenContext(t, summary, 0, "session-e"), nil, now.Add(4*time.Minute))
	mock.ExpectQuery(targetJobReportsRowsQuery).WithArgs("target-1").WillReturnRows(rows)
	mock.ExpectCommit()

	got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if got.Rounds[0].CurrentReport == nil || got.Rounds[0].CurrentReport.ID != "report-b" {
		t.Fatalf("current tie-break=%#v", got.Rounds[0].CurrentReport)
	}
	if got.Rounds[0].LatestAttempt == nil || got.Rounds[0].LatestAttempt.ID != "report-e" {
		t.Fatalf("latest tie-break=%#v", got.Rounds[0].LatestAttempt)
	}
	if got.Rounds[1].CurrentReport != nil || got.Rounds[1].LatestAttempt != nil {
		t.Fatalf("empty canonical round=%#v", got.Rounds[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTargetJobReportsOverviewEnumeratesFiveEmptyCanonicalRounds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	summary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "hr", Name: "HR", Focus: "fit", Duration: 20},
		targetJobReportsRound{Sequence: 2, Type: "technical", Name: "Technical", Focus: "design", Duration: 45},
		targetJobReportsRound{Sequence: 3, Type: "manager", Name: "Manager", Focus: "ownership", Duration: 30},
		targetJobReportsRound{Sequence: 4, Type: "culture", Name: "Culture", Focus: "values", Duration: 25},
		targetJobReportsRound{Sequence: 5, Type: "final", Name: "Final", Focus: "decision", Duration: 30},
	)
	mock.ExpectBegin()
	mock.ExpectQuery(targetJobReportsTargetQuery).WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"summary"}).AddRow(summary))
	mock.ExpectQuery(targetJobReportsRowsQuery).WithArgs("target-1").WillReturnRows(targetJobReportsRows())
	mock.ExpectCommit()

	got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if len(got.Rounds) != 5 {
		t.Fatalf("round count=%d", len(got.Rounds))
	}
	for i, round := range got.Rounds {
		if round.Round.RoundSequence != int32(i+1) || round.CurrentReport != nil || round.LatestAttempt != nil {
			t.Fatalf("round[%d]=%#v", i, round)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTargetJobReportsOverviewAllowsCurrentDisplayEvolutionWhenFrozenPairRemains(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 14, 9, 30, 0, 0, time.UTC)
	frozenSummary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "technical", Name: "Old Technical", Focus: "old focus", Duration: 45},
		targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Old Manager", Focus: "old ownership", Duration: 30},
	)
	currentSummary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "technical", Name: "Current Technical", Focus: "current focus", Duration: 60},
		targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Current Manager", Focus: "current ownership", Duration: 40},
	)
	mock.ExpectBegin()
	mock.ExpectQuery(targetJobReportsTargetQuery).WithArgs("target-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"summary"}).AddRow(currentSummary))
	mock.ExpectQuery(targetJobReportsRowsQuery).WithArgs("target-1").WillReturnRows(targetJobReportsRows().AddRow(
		"report-1", "user-1", "session-1", "target-1", "user-1", "target-1", "ready", nil,
		targetJobReportsFrozenContext(t, frozenSummary, 0, "session-1"), now, now,
	))
	mock.ExpectCommit()

	got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
	if err != nil {
		t.Fatalf("ListTargetJobReports: %v", err)
	}
	if len(got.Rounds) != 2 || got.Rounds[0].CurrentReport == nil || got.Rounds[0].CurrentReport.ID != "report-1" {
		t.Fatalf("overview=%#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListTargetJobReportsOverviewHidesMissingDeletedAndNonOwnedTarget(t *testing.T) {
	for _, name := range []string{"missing", "deleted", "non-owned"} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			mock.ExpectQuery(targetJobReportsTargetQuery).WithArgs("target-1", "user-1").
				WillReturnRows(sqlmock.NewRows([]string{"summary"}))
			mock.ExpectRollback()

			got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
			if !errors.Is(err, reviewdomain.ErrReportNotFound) || got.TargetJobID != "" || got.Rounds != nil {
				t.Fatalf("got=%#v err=%v", got, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestListTargetJobReportsOverviewRejectsAnyInvalidAttemptWithoutPartialResponse(t *testing.T) {
	validSummary := targetJobReportsSummaryJSON(
		targetJobReportsRound{Sequence: 1, Type: "technical", Name: "Technical", Focus: "design", Duration: 45},
		targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Manager", Focus: "ownership", Duration: 30},
	)
	now := time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
	type invalidRow struct {
		name           string
		currentSummary []byte
		row            func(*testing.T) []driver.Value
		wantMissing    bool
	}
	valid := func(t *testing.T) []driver.Value {
		return []driver.Value{"report-1", "user-1", "session-1", "target-1", "user-1", "target-1", "ready", nil,
			targetJobReportsFrozenContext(t, validSummary, 0, "session-1"), now, now}
	}
	tests := []invalidRow{
		{name: "invalid current summary", currentSummary: []byte(`{"interviewRounds":[]}`), row: valid},
		{name: "missing frozen context", currentSummary: validSummary, wantMissing: true, row: func(*testing.T) []driver.Value {
			row := valid(t)
			row[8] = nil
			return row
		}},
		{name: "invalid frozen context", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[8] = append(targetJobReportsFrozenContext(t, validSummary, 0, "session-1"), []byte(` {}`)...)
			return row
		}},
		{name: "unknown frozen context field", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			var value map[string]any
			if err := json.Unmarshal(targetJobReportsFrozenContext(t, validSummary, 0, "session-1"), &value); err != nil {
				t.Fatal(err)
			}
			value["unexpected"] = true
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			row[8] = raw
			return row
		}},
		{name: "report user mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[1] = "user-2"
			return row
		}},
		{name: "report target mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[3] = "target-2"
			return row
		}},
		{name: "session user mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[4] = "user-2"
			return row
		}},
		{name: "session target mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[5] = "target-2"
			return row
		}},
		{name: "frozen session mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[8] = targetJobReportsFrozenContext(t, validSummary, 0, "session-2")
			return row
		}},
		{name: "frozen target mismatch", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[8] = targetJobReportsMutateFrozenContext(t, targetJobReportsFrozenContext(t, validSummary, 0, "session-1"), func(snapshot *practicedomain.ReportContextSnapshot) {
				snapshot.TargetJob.ID = "target-2"
			})
			return row
		}},
		{name: "round pair absent from current catalog", currentSummary: targetJobReportsSummaryJSON(
			targetJobReportsRound{Sequence: 1, Type: "hr", Name: "HR", Focus: "fit", Duration: 20},
			targetJobReportsRound{Sequence: 2, Type: "manager", Name: "Manager", Focus: "ownership", Duration: 30},
		), row: valid},
		{name: "ready without generated at", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[9] = nil
			return row
		}},
		{name: "invalid status", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[6] = "cancelled"
			return row
		}},
		{name: "failed without error code", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[6] = "failed"
			row[7] = nil
			row[9] = nil
			return row
		}},
		{name: "failed with unknown error code", currentSummary: validSummary, row: func(t *testing.T) []driver.Value {
			row := valid(t)
			row[6] = "failed"
			row[7] = "UNKNOWN_CODE"
			row[9] = nil
			return row
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.ExpectBegin()
			mock.ExpectQuery(targetJobReportsTargetQuery).WithArgs("target-1", "user-1").
				WillReturnRows(sqlmock.NewRows([]string{"summary"}).AddRow(tc.currentSummary))
			if tc.name != "invalid current summary" {
				mock.ExpectQuery(targetJobReportsRowsQuery).WithArgs("target-1").
					WillReturnRows(targetJobReportsRows().AddRow(tc.row(t)...))
			}
			mock.ExpectRollback()

			got, err := NewRepository(db).ListTargetJobReports(context.Background(), "user-1", "target-1")
			if !errors.Is(err, reviewdomain.ErrReportContextInvalid) || got.TargetJobID != "" || got.Rounds != nil {
				t.Fatalf("got=%#v err=%v", got, err)
			}
			if errors.Is(err, reviewdomain.ErrReportContextMissing) != tc.wantMissing {
				t.Fatalf("missing-context classification=%v want=%v err=%v", errors.Is(err, reviewdomain.ErrReportContextMissing), tc.wantMissing, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func targetJobReportsRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"report_id", "report_user_id", "report_session_id", "report_target_job_id",
		"session_user_id", "session_target_job_id", "status", "error_code",
		"generation_context", "generated_at", "created_at",
	})
}

type targetJobReportsRound struct {
	Sequence int32
	Type     string
	Name     string
	Focus    string
	Duration int32
}

func targetJobReportsSummaryJSON(rounds ...targetJobReportsRound) []byte {
	items := make([]map[string]any, 0, len(rounds))
	for _, round := range rounds {
		items = append(items, map[string]any{
			"sequence": round.Sequence, "type": round.Type, "name": round.Name,
			"focus": round.Focus, "durationMinutes": round.Duration,
		})
	}
	raw, _ := json.Marshal(map[string]any{
		"coreThemes":      []string{"reliability"},
		"interviewRounds": items,
		"provenance": map[string]string{
			"promptVersion": "v1", "rubricVersion": "v1", "modelId": "model-1",
			"language": "en", "featureFlag": "none", "dataSourceVersion": "target-summary.v1",
		},
	})
	return raw
}

func targetJobReportsFrozenContext(t *testing.T, summary []byte, roundIndex int, sessionID string) []byte {
	t.Helper()
	rounds, err := practicedomain.ParseCanonicalReportRounds(summary)
	if err != nil {
		t.Fatalf("ParseCanonicalReportRounds: %v", err)
	}
	snapshot, err := practicedomain.BuildReportContextSnapshot(practicedomain.ReportContextSnapshotInput{
		TargetJob: practicedomain.ReportTargetJobSnapshot{
			ID: "target-1", Title: "Platform Engineer", Company: "Acme", Language: "en", RawJD: "complete jd", Summary: summary,
		},
		Resume: practicedomain.ReportResumeSnapshot{
			ID: "resume-1", DisplayName: "Resume", Language: "en", SourceSnapshot: "complete resume", StructuredProfile: json.RawMessage(`{}`),
		},
		Plan: practicedomain.ReportPlanSnapshot{
			ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en",
			TimeBudgetMinutes: rounds[roundIndex].DurationMinutes, ResumeID: "resume-1",
			RoundID: rounds[roundIndex].ID, RoundSequence: rounds[roundIndex].Sequence,
		},
		Conversation: practicedomain.ReportConversationCoordinate{SessionID: sessionID, Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatalf("BuildReportContextSnapshot: %v", err)
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func targetJobReportsMutateFrozenContext(t *testing.T, raw []byte, mutate func(*practicedomain.ReportContextSnapshot)) []byte {
	t.Helper()
	var snapshot practicedomain.ReportContextSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	mutate(&snapshot)
	updated, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return updated
}
