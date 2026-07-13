//go:build integration

package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const derivedIntegrationDSN = "postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable"

type derivedIntegrationFixture struct {
	db       *sql.DB
	ctx      context.Context
	userID   string
	resumeID string
	targetID string
	now      time.Time
	summary  string
}

func TestE2EP0070PracticeDerivedPlanCreateReadReplay(t *testing.T) {
	fixture := newDerivedIntegrationFixture(t, "70")
	fixture.cleanup(t)
	t.Cleanup(func() {
		fixture.cleanup(t)
		fixture.db.Close()
	})
	fixture.seedBase(t)

	const (
		emptySourcePlanID = "019f5700-0000-7000-8000-000000000101"
		emptySessionID    = "019f5700-0000-7000-8000-000000000102"
		emptyEventID      = "019f5700-0000-7000-8000-000000000103"
		emptyReportID     = "019f5700-0000-7000-8000-000000000104"
		focusSourcePlanID = "019f5700-0000-7000-8000-000000000111"
		focusSessionID    = "019f5700-0000-7000-8000-000000000112"
		focusEventID      = "019f5700-0000-7000-8000-000000000113"
		focusReportID     = "019f5700-0000-7000-8000-000000000114"
	)
	fixture.seedSourceReport(t, emptySourcePlanID, emptySessionID, emptyEventID, emptyReportID, []string{})
	fixture.seedSourceReport(t, focusSourcePlanID, focusSessionID, focusEventID, focusReportID, []string{"system_design"})

	repo := NewSQLRepository(fixture.db)
	create := func(planID, auditID, reportID string, goal sharedtypes.PracticeGoal) domain.PlanRecord {
		t.Helper()
		plan, err := repo.CreatePlan(fixture.ctx, domain.CreatePlanStoreInput{
			PlanID: planID, AuditEventID: auditID, UserID: fixture.userID,
			SourceReportID: reportID, Goal: goal, Now: fixture.now,
		})
		if err != nil {
			t.Fatalf("CreatePlan(%s): %v", goal, err)
		}
		return plan
	}

	generic := create("019f5700-0000-7000-8000-000000000121", "019f5700-0000-7000-8000-000000000122", emptyReportID, sharedtypes.PracticeGoalRetryCurrentRound)
	if generic.TargetJobID != fixture.targetID || generic.ResumeID != fixture.resumeID || generic.RoundID != "round-1-technical" ||
		generic.RoundSequence != 1 || generic.TimeBudgetMinutes != 45 || generic.FocusDimensionCodes == nil || len(generic.FocusDimensionCodes) != 0 {
		t.Fatalf("generic retry projection = %+v", generic)
	}
	t.Log("REPORT_GENERIC_RETRY_PASS")

	focused := create("019f5700-0000-7000-8000-000000000131", "019f5700-0000-7000-8000-000000000132", focusReportID, sharedtypes.PracticeGoalRetryCurrentRound)
	if !reflect.DeepEqual(focused.FocusDimensionCodes, []string{"system_design"}) || focused.RoundID != "round-1-technical" {
		t.Fatalf("focused retry projection = %+v", focused)
	}
	readback, err := repo.GetPlan(fixture.ctx, fixture.userID, focused.ID)
	if err != nil || !reflect.DeepEqual(readback.FocusDimensionCodes, focused.FocusDimensionCodes) || readback.SourceReportID != focusReportID {
		t.Fatalf("focused readback=%+v err=%v", readback, err)
	}
	t.Log("REPORT_DERIVED_FOCUS_PASS")

	reservation, err := repo.ReserveSessionStart(fixture.ctx, domain.StartSessionReservationInput{
		IdempotencyRecordID: "019f5700-0000-7000-8000-000000000141",
		SessionID:           "019f5700-0000-7000-8000-000000000142",
		UserID:              fixture.userID,
		PlanID:              focused.ID,
		IdempotencyKeyHash:  "p0-070-start-hash",
		RequestFingerprint:  "p0-070-start-fingerprint",
		ExpiresAt:           fixture.now.Add(time.Hour),
		Now:                 fixture.now,
	})
	if err != nil {
		t.Fatalf("ReserveSessionStart: %v", err)
	}
	wantFocus := []domain.SemanticFocusDimension{{Code: "system_design", Label: "系统设计", Issues: []string{"未说明容量估算与故障恢复取舍"}}}
	if !reflect.DeepEqual(reservation.SemanticFocus, wantFocus) {
		t.Fatalf("start semantic focus = %+v", reservation.SemanticFocus)
	}
	if _, err := fixture.db.ExecContext(fixture.ctx, `update practice_sessions set status='waiting_user_input' where id=$1`, reservation.SessionID); err != nil {
		t.Fatalf("make session message-ready: %v", err)
	}
	if _, err := fixture.db.ExecContext(fixture.ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,created_at)
values ($1,$2,1,'assistant','请介绍你的系统设计方案。',$3)`, "019f5700-0000-7000-8000-000000000143", reservation.SessionID, fixture.now); err != nil {
		t.Fatalf("insert opening message: %v", err)
	}
	messageReservation, err := repo.ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5700-0000-7000-8000-000000000144", UserID: fixture.userID,
		SessionID: reservation.SessionID, ClientMessageID: "019f5700-0000-7000-8000-000000000145",
		Text: "我会先估算容量。", Now: fixture.now,
	})
	if err != nil || !reflect.DeepEqual(messageReservation.Session.SemanticFocus, wantFocus) {
		t.Fatalf("send semantic focus=%+v err=%v", messageReservation.Session.SemanticFocus, err)
	}
	t.Log("REPORT_DERIVED_SEMANTIC_PROMPT_PASS")

	next := create("019f5700-0000-7000-8000-000000000151", "019f5700-0000-7000-8000-000000000152", focusReportID, sharedtypes.PracticeGoalNextRound)
	if next.RoundID != "round-2-manager" || next.RoundSequence != 2 || next.TimeBudgetMinutes != 30 || next.FocusDimensionCodes == nil || len(next.FocusDimensionCodes) != 0 {
		t.Fatalf("next-round projection = %+v", next)
	}
	t.Log("REPORT_NEXT_EMPTY_FOCUS_PASS")
	t.Log("REPORT_DERIVED_POSTGRES_PASS")
}

func TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy(t *testing.T) {
	fixture := newDerivedIntegrationFixture(t, "72")
	fixture.cleanup(t)
	t.Cleanup(func() {
		fixture.cleanup(t)
		fixture.db.Close()
	})
	fixture.seedBase(t)

	const (
		sourcePlanID = "019f5720-0000-7000-8000-000000000101"
		sessionID    = "019f5720-0000-7000-8000-000000000102"
		eventID      = "019f5720-0000-7000-8000-000000000103"
		reportID     = "019f5720-0000-7000-8000-000000000104"
	)
	fixture.seedSourceReport(t, sourcePlanID, sessionID, eventID, reportID, []string{"system_design"})
	repo := NewSQLRepository(fixture.db)
	originalContext := fixture.reportContext(t, sourcePlanID, sessionID)
	originalRaw, _ := json.Marshal(originalContext)

	type invalidCase struct {
		name     string
		userID   string
		reportID string
		prepare  func(t *testing.T)
		restore  func(t *testing.T)
	}
	cases := []invalidCase{
		{name: "missing", userID: fixture.userID, reportID: "019f5720-0000-7000-8000-000000000199"},
		{name: "cross-user", userID: "019f5720-0000-7000-8000-000000000299", reportID: reportID},
		{name: "non-ready", userID: fixture.userID, reportID: reportID,
			prepare: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set status='generating' where id=$1`, reportID)
			},
			restore: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set status='ready' where id=$1`, reportID)
			}},
		{name: "missing-context", userID: fixture.userID, reportID: reportID,
			prepare: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set generation_context='{}'::jsonb where id=$1`, reportID)
			},
			restore: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set generation_context=$1::jsonb where id=$2`, originalRaw, reportID)
			}},
	}
	for _, kind := range []string{"wrong-target", "wrong-resume", "wrong-round", "wrong-persona", "wrong-language", "wrong-budget"} {
		kind := kind
		cases = append(cases, invalidCase{name: kind, userID: fixture.userID, reportID: reportID,
			prepare: func(t *testing.T) {
				mutated := mutateDerivedContextForIntegration(t, originalContext, kind)
				raw, _ := json.Marshal(mutated)
				fixture.exec(t, `update feedback_reports set generation_context=$1::jsonb where id=$2`, raw, reportID)
			},
			restore: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set generation_context=$1::jsonb where id=$2`, originalRaw, reportID)
			},
		})
	}
	cases = append(cases,
		invalidCase{name: "unsupported-focus", userID: fixture.userID, reportID: reportID,
			prepare: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set retry_focus_dimension_codes=array['delivery']::text[] where id=$1`, reportID)
			},
			restore: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set retry_focus_dimension_codes=array['system_design']::text[] where id=$1`, reportID)
			}},
		invalidCase{name: "duplicate-focus", userID: fixture.userID, reportID: reportID,
			prepare: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set retry_focus_dimension_codes=array['system_design','system_design']::text[] where id=$1`, reportID)
			},
			restore: func(t *testing.T) {
				fixture.exec(t, `update feedback_reports set retry_focus_dimension_codes=array['system_design']::text[] where id=$1`, reportID)
			}},
	)

	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}
			if tc.restore != nil {
				defer tc.restore(t)
			}
			planID := derivedIsolationUUID(index, 1)
			auditID := derivedIsolationUUID(index, 2)
			_, err := repo.CreatePlan(fixture.ctx, domain.CreatePlanStoreInput{
				PlanID: planID, AuditEventID: auditID, UserID: tc.userID,
				SourceReportID: tc.reportID, Goal: sharedtypes.PracticeGoalRetryCurrentRound, Now: fixture.now,
			})
			if !errors.Is(err, domain.ErrPlanPrerequisiteNotFound) {
				t.Fatalf("error=%v want generic prerequisite failure", err)
			}
			var count int
			if err := fixture.db.QueryRowContext(fixture.ctx, `select count(*) from practice_plans where id=$1`, planID).Scan(&count); err != nil || count != 0 {
				t.Fatalf("invalid source inserted plan count=%d err=%v", count, err)
			}
		})
	}
	t.Log("REPORT_DERIVED_ISOLATION_PASS")
	t.Log("REPORT_DERIVED_PRIVACY_PASS")
	t.Log("REPORT_DERIVED_POSTGRES_PASS")
}

func newDerivedIntegrationFixture(t *testing.T, suffix string) derivedIntegrationFixture {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = derivedIntegrationDSN
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("postgres ping: %v", err)
	}
	base := "019f57" + suffix
	return derivedIntegrationFixture{
		db: db, ctx: ctx,
		userID:   base + "-0000-7000-8000-000000000001",
		resumeID: base + "-0000-7000-8000-000000000002",
		targetID: base + "-0000-7000-8000-000000000003",
		now:      time.Now().UTC().Truncate(time.Microsecond),
		summary:  `{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","durationMinutes":45,"focus":"系统设计"},{"sequence":2,"type":"manager","name":"经理面","durationMinutes":30,"focus":"影响力"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"zh-CN","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`,
	}
}

func (f derivedIntegrationFixture) cleanup(t *testing.T) {
	t.Helper()
	if _, err := f.db.ExecContext(context.Background(), `delete from audit_events where user_id=$1 or actor_id=$1`, f.userID); err != nil {
		t.Fatalf("cleanup derived integration audit events: %v", err)
	}
	if _, err := f.db.ExecContext(context.Background(), `delete from users where id=$1`, f.userID); err != nil {
		t.Fatalf("cleanup derived integration user: %v", err)
	}
}

func (f derivedIntegrationFixture) seedBase(t *testing.T) {
	t.Helper()
	f.exec(t, `insert into users (id,email,display_name,created_at,updated_at) values ($1,$2,$3,$4,$4)`, f.userID, "derived-"+f.userID+"@example.test", "Derived Integration", f.now)
	f.exec(t, `
insert into resumes (
  id,user_id,title,display_name,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Integration resume','Integration resume','zh-CN','ready','{}'::jsonb,$3,'paste',$3,$3,'{}'::jsonb,$4,$4)`, f.resumeID, f.userID, "完整简历", f.now)
	f.exec(t, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,company_name,target_language,source_type,
  raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','Example','zh-CN','manual_text',$4,$5::jsonb,'{}'::jsonb,$6,$6)`, f.targetID, f.userID, f.resumeID, "完整 JD", f.summary, f.now)
}

func (f derivedIntegrationFixture) seedSourceReport(t *testing.T, planID, sessionID, eventID, reportID string, focus []string) {
	t.Helper()
	f.exec(t, `
insert into practice_plans (
  id,user_id,target_job_id,goal,round_id,round_sequence,interviewer_persona,difficulty,language,
  time_budget_minutes,resume_id,focus_dimension_codes,status,created_at,updated_at
) values ($1,$2,$3,'baseline','round-1-technical',1,'hiring_manager','standard','zh-CN',45,$4,'{}'::text[],'ready',$5,$5)`, planID, f.userID, f.targetID, f.resumeID, f.now)
	f.exec(t, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,completed_at,created_at,updated_at)
values ($1,$2,$3,$4,'completed','zh-CN',$5,$5,$5)`, sessionID, f.userID, planID, f.targetID, f.now)
	f.exec(t, `insert into practice_session_events (id,session_id,seq_no,event_type,payload,created_at) values ($1,$2,1,'session_completed','{}'::jsonb,$3)`, eventID, sessionID, f.now)
	contextRaw, err := json.Marshal(f.reportContext(t, planID, sessionID))
	if err != nil {
		t.Fatal(err)
	}
	f.exec(t, `
insert into feedback_reports (
  id,user_id,session_id,target_job_id,status,summary,generation_context,preparedness_level,
  dimension_assessments,issues,next_actions,retry_focus_dimension_codes,language,created_at,updated_at
) values ($1,$2,$3,$4,'ready','需要进一步练习',$5::jsonb,'needs_practice',$6::jsonb,$7::jsonb,'[]'::jsonb,$8,'zh-CN',$9,$9)`,
		reportID, f.userID, sessionID, f.targetID, contextRaw,
		`[{"code":"system_design","label":"系统设计","status":"needs_work","confidence":"high"}]`,
		`[{"dimensionCode":"system_design","evidence":"未说明容量估算与故障恢复取舍","confidence":"high","sourceMessageSeqNos":[2]}]`,
		pq.Array(focus), f.now,
	)
}

func (f derivedIntegrationFixture) reportContext(t *testing.T, planID, sessionID string) domain.ReportContextSnapshot {
	t.Helper()
	context, err := domain.BuildReportContextSnapshot(domain.ReportContextSnapshotInput{
		TargetJob: domain.ReportTargetJobSnapshot{
			ID: f.targetID, Title: "Platform Engineer", Company: "Example", Language: "zh-CN", RawJD: "完整 JD",
			Summary: json.RawMessage(f.summary), Requirements: []domain.ReportRequirementSnapshot{},
		},
		Resume: domain.ReportResumeSnapshot{ID: f.resumeID, DisplayName: "Integration resume", Language: "zh-CN", SourceSnapshot: "完整简历", StructuredProfile: json.RawMessage(`{}`)},
		Plan: domain.ReportPlanSnapshot{
			ID: planID, Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "zh-CN",
			TimeBudgetMinutes: 45, ResumeID: f.resumeID, RoundID: "round-1-technical", RoundSequence: 1, FocusDimensionCodes: []string{},
		},
		Conversation: domain.ReportConversationCoordinate{SessionID: sessionID, Language: "zh-CN", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatalf("BuildReportContextSnapshot: %v", err)
	}
	return context
}

func (f derivedIntegrationFixture) exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := f.db.ExecContext(f.ctx, query, args...); err != nil {
		t.Fatalf("exec derived integration fixture: %v", err)
	}
}

func mutateDerivedContextForIntegration(t *testing.T, source domain.ReportContextSnapshot, kind string) domain.ReportContextSnapshot {
	t.Helper()
	mutated := source
	mutated.TargetJob.Summary = append(json.RawMessage(nil), source.TargetJob.Summary...)
	mutated.Resume.StructuredProfile = append(json.RawMessage(nil), source.Resume.StructuredProfile...)
	mutated.CanonicalRounds = append([]domain.ReportRoundSnapshot(nil), source.CanonicalRounds...)
	switch kind {
	case "wrong-target":
		mutated.TargetJob.ID = "019f5720-0000-7000-8000-000000000401"
	case "wrong-resume":
		mutated.Resume.ID = "019f5720-0000-7000-8000-000000000402"
		mutated.Plan.ResumeID = mutated.Resume.ID
	case "wrong-round":
		mutated.Round = mutated.CanonicalRounds[1]
		mutated.Plan.RoundID = mutated.Round.ID
		mutated.Plan.RoundSequence = mutated.Round.Sequence
		mutated.HasNextRound = false
	case "wrong-persona":
		mutated.Plan.InterviewerPersona = "generalist"
	case "wrong-language":
		mutated.Plan.Language = "en"
		mutated.Conversation.Language = "en"
	case "wrong-budget":
		mutated.Plan.TimeBudgetMinutes = 46
	default:
		t.Fatalf("unknown integration context mutation %q", kind)
	}
	return mutated
}

func derivedIsolationUUID(index, offset int) string {
	return "019f5720-0000-7000-8001-" + leftPad12(index*10+offset)
}

func leftPad12(value int) string {
	digits := "000000000000" + fmt.Sprintf("%d", value)
	return digits[len(digits)-12:]
}
