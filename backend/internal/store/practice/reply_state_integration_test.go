//go:build integration

package practice

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestIntegrationPracticeReplyStateRecovery(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for the practice reply-state integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	const (
		userID     = "019f5b70-0000-7000-8000-000000000001"
		resumeID   = "019f5b70-0000-7000-8000-000000000002"
		targetID   = "019f5b70-0000-7000-8000-000000000003"
		plan1ID    = "019f5b70-0000-7000-8000-000000000004"
		plan2ID    = "019f5b70-0000-7000-8000-000000000005"
		session1ID = "019f5b70-0000-7000-8000-000000000006"
		session2ID = "019f5b70-0000-7000-8000-000000000007"
		opening1ID = "019f5b70-0000-7000-8000-000000000008"
		opening2ID = "019f5b70-0000-7000-8000-000000000009"
		user1ID    = "019f5b70-0000-7000-8000-00000000000a"
		user2ID    = "019f5b70-0000-7000-8000-00000000000b"
		assistant  = "019f5b70-0000-7000-8000-00000000000c"
		clientID   = "019f5b70-0000-7000-8000-00000000000d"
	)
	cleanup := func() {
		_, _ = db.ExecContext(context.Background(), `delete from users where id=$1`, userID)
	}
	cleanup()
	t.Cleanup(cleanup)
	now := time.Now().UTC().Truncate(time.Microsecond)
	mustExecReplyState(t, ctx, db, `insert into users (id,email,display_name,created_at,updated_at) values ($1,'reply-state@example.test','Reply State',$2,$2)`, userID, now)
	mustExecReplyState(t, ctx, db, `
insert into resumes (
  id,user_id,title,display_name,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Reply State Resume','Reply State Resume','zh-CN','ready','{}'::jsonb,'完整简历','paste','完整简历','完整简历','{}'::jsonb,$3,$3)`, resumeID, userID, now)
	summary := `{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","durationMinutes":45,"focus":"系统设计"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"zh-CN","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
	mustExecReplyState(t, ctx, db, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','zh-CN','完整 JD',$4::jsonb,'{}'::jsonb,$5,$5)`, targetID, userID, resumeID, summary, now)
	for _, planID := range []string{plan1ID, plan2ID} {
		mustExecReplyState(t, ctx, db, `
insert into practice_plans (
  id,user_id,target_job_id,goal,interviewer_persona,difficulty,language,time_budget_minutes,
  resume_id,focus_dimension_codes,status,round_id,round_sequence,created_at,updated_at
) values ($1,$2,$3,'baseline','hiring_manager','standard','zh-CN',45,$4,'{}'::text[],'ready','round-1-technical',1,$5,$5)`, planID, userID, targetID, resumeID, now)
	}
	for _, item := range []struct{ sessionID, planID, openingID string }{
		{session1ID, plan1ID, opening1ID},
		{session2ID, plan2ID, opening2ID},
	} {
		mustExecReplyState(t, ctx, db, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,started_at,created_at,updated_at)
values ($1,$2,$3,$4,'running','zh-CN',$5,$5,$5)`, item.sessionID, userID, item.planID, targetID, now)
		mustExecReplyState(t, ctx, db, `
insert into practice_messages (id,session_id,seq_no,role,content,created_at)
values ($1,$2,1,'assistant','请介绍一个项目。',$3)`, item.openingID, item.sessionID, now)
	}

	repo := NewSQLRepository(db)
	first, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: user1ID, UserID: userID, SessionID: session1ID, ClientMessageID: clientID, Text: "我负责了迁移。", Now: now,
	})
	if err != nil {
		t.Fatalf("reserve first message: %v", err)
	}
	if first.UserMessage.ReplyStatus != domain.PracticeReplyStatusPending || first.UserMessage.ClientMessageID != clientID {
		t.Fatalf("first reservation = %+v", first)
	}
	readPending, err := repo.GetSession(ctx, userID, session1ID, now)
	if err != nil || len(readPending.Messages) != 2 || readPending.Messages[1].ReplyStatus != domain.PracticeReplyStatusPending {
		t.Fatalf("pending readback=%+v err=%v", readPending, err)
	}
	if _, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5b70-0000-7000-8000-00000000000e", UserID: userID, SessionID: session1ID,
		ClientMessageID: clientID, Text: "我负责了迁移。", Now: now,
	}); !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("pending same-ID error=%v want ErrSessionConflict", err)
	}
	if err := repo.FailPracticeMessage(ctx, domain.FailPracticeMessageInput{
		UserID: userID, SessionID: session1ID, UserMessageID: user1ID,
		ExpectedReplyGeneration: first.ReplyGeneration, ReplyStatus: domain.PracticeReplyStatusRetryableFailed,
	}); err != nil {
		t.Fatalf("persist retryable failure: %v", err)
	}
	readFailed, err := repo.GetSession(ctx, userID, session1ID, now)
	if err != nil || readFailed.Messages[1].ReplyStatus != domain.PracticeReplyStatusRetryableFailed {
		t.Fatalf("retryable readback=%+v err=%v", readFailed, err)
	}
	if _, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5b70-0000-7000-8000-00000000000f", UserID: userID, SessionID: session1ID,
		ClientMessageID: clientID, Text: "不同正文", Now: now,
	}); !errors.Is(err, domain.ErrClientEventMismatch) {
		t.Fatalf("same-ID mismatch error=%v want ErrClientEventMismatch", err)
	}
	retry, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5b70-0000-7000-8000-000000000010", UserID: userID, SessionID: session1ID,
		ClientMessageID: clientID, Text: "我负责了迁移。", Now: now,
	})
	if err != nil || retry.UserMessage.ID != user1ID || retry.UserMessage.ReplyStatus != domain.PracticeReplyStatusPending {
		t.Fatalf("retry reservation=%+v err=%v", retry, err)
	}
	committed, err := repo.CommitPracticeMessage(ctx, domain.CommitPracticeMessageInput{
		UserID: userID, SessionID: session1ID, UserMessageID: user1ID,
		ExpectedReplyGeneration: retry.ReplyGeneration,
		AssistantMessageID:      assistant, AssistantText: "请说明取舍。", Now: now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("commit reply: %v", err)
	}
	if committed.UserMessage.ReplyStatus != domain.PracticeReplyStatusComplete || committed.AssistantMessage.ReplyStatus != "" {
		t.Fatalf("committed pair = %+v", committed)
	}
	replay, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5b70-0000-7000-8000-000000000011", UserID: userID, SessionID: session1ID,
		ClientMessageID: clientID, Text: "我负责了迁移。", Now: now.Add(2 * time.Second),
	})
	if err != nil || replay.Replay == nil || replay.Replay.AssistantMessage.ID != assistant {
		t.Fatalf("complete replay=%+v err=%v", replay, err)
	}
	var userCount, assistantCount int
	if err := db.QueryRowContext(ctx, `
select count(*) filter (where role='user'), count(*) filter (where role='assistant' and reply_to_message_id=$2)
from practice_messages where session_id=$1`, session1ID, user1ID).Scan(&userCount, &assistantCount); err != nil {
		t.Fatalf("count converged messages: %v", err)
	}
	if userCount != 1 || assistantCount != 1 {
		t.Fatalf("converged counts user=%d assistant=%d", userCount, assistantCount)
	}

	second, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: user2ID, UserID: userID, SessionID: session2ID, ClientMessageID: clientID, Text: "第二场回答。", Now: now,
	})
	if err != nil || second.UserMessage.ClientMessageID != clientID {
		t.Fatalf("cross-session same client ID=%+v err=%v", second, err)
	}
	if err := repo.FailPracticeMessage(ctx, domain.FailPracticeMessageInput{
		UserID: userID, SessionID: session2ID, UserMessageID: user2ID,
		ExpectedReplyGeneration: second.ReplyGeneration, ReplyStatus: domain.PracticeReplyStatusTerminalFailed,
	}); err != nil {
		t.Fatalf("persist terminal failure: %v", err)
	}
	if _, err := repo.ReservePracticeMessage(ctx, domain.ReservePracticeMessageInput{
		UserMessageID: "019f5b70-0000-7000-8000-000000000012", UserID: userID, SessionID: session2ID,
		ClientMessageID: clientID, Text: "第二场回答。", Now: now,
	}); !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("terminal same-ID error=%v want ErrSessionConflict", err)
	}
	terminal, err := repo.GetSession(ctx, userID, session2ID, now)
	if err != nil || terminal.Messages[1].ReplyStatus != domain.PracticeReplyStatusTerminalFailed {
		t.Fatalf("terminal readback=%+v err=%v", terminal, err)
	}
	if _, err := repo.GetSession(ctx, "019f5b70-0000-7000-8000-000000000099", session2ID, now); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("cross-user read error=%v want ErrSessionNotFound", err)
	}

	if _, err := db.ExecContext(ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,reply_status,created_at)
values ('019f5b70-0000-7000-8000-000000000020',$1,4,'assistant','invalid','pending',$2)`, session1ID, now); err == nil {
		t.Fatal("assistant reply_status must be rejected by the database")
	}
	if _, err := db.ExecContext(ctx, `
insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,created_at)
values ('019f5b70-0000-7000-8000-000000000021',$1,4,'user','invalid','019f5b70-0000-7000-8000-000000000022',$2)`, session1ID, now); err == nil {
		t.Fatal("user without reply_status must be rejected by the database")
	}
	for _, probe := range []struct {
		name  string
		query string
		args  []any
	}{
		{
			name: "non-positive generation",
			query: `insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,reply_lease_expires_at,created_at)
values ('019f5b70-0000-7000-8000-000000000023',$1,4,'user','invalid','019f5b70-0000-7000-8000-000000000024','pending',0,$2,$3)`,
			args: []any{session1ID, now.Add(domain.PracticeReplyLeaseDuration), now},
		},
		{
			name: "missing generation",
			query: `insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_lease_expires_at,created_at)
values ('019f5b70-0000-7000-8000-00000000002a',$1,4,'user','invalid','019f5b70-0000-7000-8000-00000000002b','pending',$2,$3)`,
			args: []any{session1ID, now.Add(domain.PracticeReplyLeaseDuration), now},
		},
		{
			name: "pending without lease",
			query: `insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,created_at)
values ('019f5b70-0000-7000-8000-000000000025',$1,4,'user','invalid','019f5b70-0000-7000-8000-000000000026','pending',1,$2)`,
			args: []any{session1ID, now},
		},
		{
			name: "failed with lease",
			query: `insert into practice_messages (id,session_id,seq_no,role,content,client_message_id,reply_status,reply_generation,reply_lease_expires_at,created_at)
values ('019f5b70-0000-7000-8000-000000000027',$1,4,'user','invalid','019f5b70-0000-7000-8000-000000000028','retryable_failed',1,$2,$3)`,
			args: []any{session1ID, now.Add(domain.PracticeReplyLeaseDuration), now},
		},
		{
			name: "assistant with generation and lease",
			query: `insert into practice_messages (id,session_id,seq_no,role,content,reply_generation,reply_lease_expires_at,created_at)
values ('019f5b70-0000-7000-8000-000000000029',$1,4,'assistant','invalid',1,$2,$3)`,
			args: []any{session1ID, now.Add(domain.PracticeReplyLeaseDuration), now},
		},
	} {
		if _, err := db.ExecContext(ctx, probe.query, probe.args...); err == nil {
			t.Fatalf("%s must be rejected by the database", probe.name)
		}
	}

	mustExecReplyState(t, ctx, db, `delete from users where id=$1`, userID)
	var remaining int
	if err := db.QueryRowContext(ctx, `select count(*) from practice_messages where session_id in ($1,$2)`, session1ID, session2ID).Scan(&remaining); err != nil {
		t.Fatalf("count privacy-deleted messages: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("privacy cascade left %d practice messages", remaining)
	}
	t.Log("PRACTICE_REPLY_STATE_RECOVERY_PASS")
}

func mustExecReplyState(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec reply-state fixture: %v", err)
	}
}
