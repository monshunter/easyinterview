package practice

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryCreateDerivedPlanProjectsDimensionFocus(t *testing.T) {
	tests := []struct {
		name         string
		goal         sharedtypes.PracticeGoal
		retryCodes   string
		wantRoundID  string
		wantSequence int32
		wantDuration int32
		wantFocus    []string
		marker       string
	}{
		{name: "generic retry", goal: sharedtypes.PracticeGoalRetryCurrentRound, retryCodes: `{}`, wantRoundID: "round-1-technical", wantSequence: 1, wantDuration: 45, wantFocus: []string{}, marker: "REPORT_GENERIC_RETRY_PASS"},
		{name: "issue backed retry", goal: sharedtypes.PracticeGoalRetryCurrentRound, retryCodes: `{system_design}`, wantRoundID: "round-1-technical", wantSequence: 1, wantDuration: 45, wantFocus: []string{"system_design"}, marker: "REPORT_DERIVED_FOCUS_PASS"},
		{name: "next round ignores report retry focus", goal: sharedtypes.PracticeGoalNextRound, retryCodes: `{system_design}`, wantRoundID: "round-2-manager", wantSequence: 2, wantDuration: 30, wantFocus: []string{}, marker: "REPORT_NEXT_EMPTY_FOCUS_PASS"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			now := time.Unix(10, 0).UTC()
			in := domain.CreatePlanStoreInput{
				PlanID: "plan-derived", AuditEventID: "audit-derived", UserID: "user-1",
				SourceReportID: "report-1", Goal: tc.goal, Now: now,
			}

			mock.ExpectBegin()
			mock.ExpectQuery(`(?s)from feedback_reports fr.*join practice_sessions ps.*join practice_plans pp.*fr\.status = 'ready'.*for update`).
				WithArgs(in.SourceReportID, in.UserID).
				WillReturnRows(sqlmock.NewRows([]string{
					"target_job_id", "session_id", "generation_context", "dimension_assessments", "issues", "retry_focus_dimension_codes",
					"plan_id", "plan_target_job_id", "resume_id", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes",
				}).AddRow(
					"target-1", "session-1", derivedReportContextJSON(t),
					`[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`,
					`[{"dimensionCode":"system_design","evidence":"未说明容量估算与故障恢复取舍","confidence":"high","sourceMessageSeqNos":[2]}]`,
					tc.retryCodes,
					"source-plan-1", "target-1", "resume-1", "round-1-technical", 1,
					string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", 45,
				))
			mock.ExpectQuery(`(?s)insert into practice_plans .*focus_dimension_codes.*returning .*focus_dimension_codes`).
				WithArgs(
					in.PlanID, in.UserID, "target-1", in.SourceReportID, string(in.Goal), tc.wantRoundID, tc.wantSequence,
					string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", tc.wantDuration, "resume-1",
					sqlmock.AnyArg(), in.Now,
				).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "target_job_id", "source_report_id", "goal", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "status", "created_at",
				}).AddRow(
					in.PlanID, "target-1", in.SourceReportID, string(in.Goal), tc.wantRoundID, tc.wantSequence,
					string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", tc.wantDuration, "resume-1", textArray(tc.wantFocus), "ready", now,
				))
			mock.ExpectExec(`insert into audit_events`).WithArgs(in.AuditEventID, in.UserID, in.UserID, in.PlanID, sqlmock.AnyArg(), in.Now).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			plan, err := NewSQLRepository(db).CreatePlan(context.Background(), in)
			if err != nil {
				t.Fatalf("CreatePlan: %v", err)
			}
			if plan.TargetJobID != "target-1" || plan.ResumeID != "resume-1" || plan.RoundID != tc.wantRoundID ||
				plan.RoundSequence != tc.wantSequence || plan.TimeBudgetMinutes != tc.wantDuration ||
				!reflect.DeepEqual(plan.FocusDimensionCodes, tc.wantFocus) {
				t.Fatalf("derived plan = %+v", plan)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			t.Log(tc.marker)
		})
	}
}

func TestSQLRepositoryGetPlanReplaysExactDimensionFocus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Unix(10, 0).UTC()
	mock.ExpectQuery(`(?s)select id, target_job_id, source_report_id::text, goal, round_id, round_sequence.*focus_dimension_codes.*from practice_plans`).
		WithArgs("user-1", "plan-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "source_report_id", "goal", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes", "resume_id", "focus_dimension_codes", "status", "created_at",
		}).AddRow(
			"plan-1", "target-1", "report-1", string(sharedtypes.PracticeGoalRetryCurrentRound), "round-1-technical", 1,
			string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", 45, "resume-1", `{system_design}`, "ready", now,
		))

	plan, err := NewSQLRepository(db).GetPlan(context.Background(), "user-1", "plan-1")
	if err != nil {
		t.Fatalf("GetPlan: %v", err)
	}
	if !reflect.DeepEqual(plan.FocusDimensionCodes, []string{"system_design"}) {
		t.Fatalf("focus replay = %+v", plan.FocusDimensionCodes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLRepositoryCreateDerivedPlanFailsClosedOnSourceMatrix(t *testing.T) {
	for _, kind := range []string{
		"missing", "cross-user", "non-ready", "missing-context", "wrong-target", "wrong-resume", "wrong-round",
		"wrong-persona", "wrong-language", "wrong-budget", "unsupported-focus", "duplicate-focus",
	} {
		t.Run(kind, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			in := domain.CreatePlanStoreInput{
				PlanID: "plan-derived", AuditEventID: "audit-derived", UserID: "user-1",
				SourceReportID: "report-private", Goal: sharedtypes.PracticeGoalRetryCurrentRound, Now: time.Unix(10, 0).UTC(),
			}
			mock.ExpectBegin()
			expectation := mock.ExpectQuery(`(?s)from feedback_reports fr.*join practice_sessions ps.*join practice_plans pp.*fr\.status = 'ready'.*for update`).
				WithArgs(in.SourceReportID, in.UserID)
			if kind == "missing" || kind == "cross-user" || kind == "non-ready" {
				expectation.WillReturnRows(sqlmock.NewRows(derivedSourceColumns()))
			} else {
				contextRaw := derivedReportContextJSON(t)
				if kind == "missing-context" {
					contextRaw = []byte(`{}`)
				} else if strings.HasPrefix(kind, "wrong-") {
					contextRaw = mutateDerivedReportContext(t, contextRaw, kind)
				}
				focus := `{system_design}`
				if kind == "unsupported-focus" {
					focus = `{delivery}`
				}
				if kind == "duplicate-focus" {
					focus = `{system_design,system_design}`
				}
				expectation.WillReturnRows(sqlmock.NewRows(derivedSourceColumns()).AddRow(
					"target-1", "session-1", contextRaw,
					`[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`,
					`[{"dimensionCode":"system_design","evidence":"缺少容量估算","confidence":"high","sourceMessageSeqNos":[2]}]`,
					focus, "source-plan-1", "target-1", "resume-1", "round-1-technical", 1,
					string(sharedtypes.InterviewerRoleHiringManager), "standard", "zh-CN", 45,
				))
			}
			mock.ExpectRollback()

			_, err = NewSQLRepository(db).CreatePlan(context.Background(), in)
			if !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
				t.Fatalf("error=%v want generic prerequisite failure", err)
			}
			if strings.Contains(err.Error(), in.SourceReportID) || strings.Contains(err.Error(), "target-1") || strings.Contains(err.Error(), "resume-1") {
				t.Fatalf("error leaks private source identity: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("invalid source performed insert or other side effect: %v", err)
			}
		})
	}
}

func derivedReportContextJSON(t *testing.T) []byte {
	t.Helper()
	snapshot := domain.ReportContextSnapshot{
		SchemaVersion: domain.ReportContextSchemaVersion,
		TargetJob: domain.ReportTargetJobSnapshot{
			ID: "target-1", Title: "平台工程师", Company: "Example", Language: "zh-CN", RawJD: "完整 JD",
			Summary:      json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","durationMinutes":45,"focus":"系统设计"},{"sequence":2,"type":"manager","name":"经理面","durationMinutes":30,"focus":"影响力"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture","language":"zh-CN","dataSourceVersion":"target-job.v1"}}`),
			Requirements: []domain.ReportRequirementSnapshot{},
		},
		Resume: domain.ReportResumeSnapshot{ID: "resume-1", DisplayName: "主简历", Language: "zh-CN", SourceSnapshot: "完整简历", StructuredProfile: json.RawMessage(`{}`)},
		Round:  domain.ReportRoundSnapshot{ID: "round-1-technical", Sequence: 1, Type: "technical", Name: "技术面", Focus: "系统设计", DurationMinutes: 45},
		CanonicalRounds: []domain.ReportRoundSnapshot{
			{ID: "round-1-technical", Sequence: 1, Type: "technical", Name: "技术面", Focus: "系统设计", DurationMinutes: 45},
			{ID: "round-2-manager", Sequence: 2, Type: "manager", Name: "经理面", Focus: "影响力", DurationMinutes: 30},
		},
		Plan: domain.ReportPlanSnapshot{
			ID: "source-plan-1", Goal: "baseline", InterviewerPersona: string(sharedtypes.InterviewerRoleHiringManager),
			Difficulty: "standard", Language: "zh-CN", TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1, FocusDimensionCodes: []string{},
		},
		Conversation: domain.ReportConversationCoordinate{SessionID: "session-1", Language: "zh-CN", MessageCount: 3, LastMessageSeqNo: 3},
		HasNextRound: true,
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func textArray(values []string) string {
	if len(values) == 0 {
		return `{}`
	}
	return `{` + values[0] + `}`
}

func derivedSourceColumns() []string {
	return []string{
		"target_job_id", "session_id", "generation_context", "dimension_assessments", "issues", "retry_focus_dimension_codes",
		"plan_id", "plan_target_job_id", "resume_id", "round_id", "round_sequence", "interviewer_persona", "difficulty", "language", "time_budget_minutes",
	}
}

func mutateDerivedReportContext(t *testing.T, raw []byte, kind string) []byte {
	t.Helper()
	var snapshot domain.ReportContextSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	switch kind {
	case "wrong-target":
		snapshot.TargetJob.ID = "target-other"
	case "wrong-resume":
		snapshot.Resume.ID = "resume-other"
		snapshot.Plan.ResumeID = "resume-other"
	case "wrong-round":
		snapshot.Round = snapshot.CanonicalRounds[1]
		snapshot.Plan.RoundID = snapshot.Round.ID
		snapshot.Plan.RoundSequence = snapshot.Round.Sequence
		snapshot.HasNextRound = false
	case "wrong-persona":
		snapshot.Plan.InterviewerPersona = string(sharedtypes.InterviewerRoleGeneralist)
	case "wrong-language":
		snapshot.Plan.Language = "en"
		snapshot.Conversation.Language = "en"
	case "wrong-budget":
		snapshot.Plan.TimeBudgetMinutes = 46
	default:
		t.Fatalf("unknown context mutation %q", kind)
	}
	out, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
