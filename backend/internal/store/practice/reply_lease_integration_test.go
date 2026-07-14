//go:build integration

package practice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
)

type replyLeaseIntegrationFixture struct {
	dsn       string
	db        *sql.DB
	ctx       context.Context
	userID    string
	sessionID string
	now       time.Time
}

func TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce(t *testing.T) {
	fixture := setupReplyLeaseIntegrationFixture(t, 1)
	repos := openReplyLeaseRepositories(t, fixture.dsn, 2)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo *SQLRepository) {
			defer wg.Done()
			<-start
			_, err := repo.ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
				UserMessageID: replyLeaseUUID(1, 20+i), UserID: fixture.userID, SessionID: fixture.sessionID,
				ClientMessageID: replyLeaseUUID(1, 30+i), Text: fmt.Sprintf("并发回答 %d", i), Now: fixture.now,
			})
			results <- err
		}(i, repo)
	}
	close(start)
	wg.Wait()
	close(results)
	assertOneReservationWinner(t, results)
	assertReplyLeaseRow(t, fixture, 1, domain.PracticeReplyStatusPending, fixture.now.Add(domain.PracticeReplyLeaseDuration), 1, 0)
}

func TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce(t *testing.T) {
	fixture := setupReplyLeaseIntegrationFixture(t, 2)
	repos := openReplyLeaseRepositories(t, fixture.dsn, 2)
	start := make(chan struct{})
	results := make(chan error, 2)
	clientMessageID := replyLeaseUUID(2, 30)
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo *SQLRepository) {
			defer wg.Done()
			<-start
			_, err := repo.ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
				UserMessageID: replyLeaseUUID(2, 20+i), UserID: fixture.userID, SessionID: fixture.sessionID,
				ClientMessageID: clientMessageID, Text: "同一回答", Now: fixture.now,
			})
			results <- err
		}(i, repo)
	}
	close(start)
	wg.Wait()
	close(results)
	assertOneReservationWinner(t, results)
	assertReplyLeaseRow(t, fixture, 1, domain.PracticeReplyStatusPending, fixture.now.Add(domain.PracticeReplyLeaseDuration), 1, 0)
}

func TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration(t *testing.T) {
	fixture := setupReplyLeaseIntegrationFixture(t, 3)
	initialRepo := openReplyLeaseRepositories(t, fixture.dsn, 1)[0]
	clientMessageID := replyLeaseUUID(3, 30)
	first, err := initialRepo.ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
		UserMessageID: replyLeaseUUID(3, 20), UserID: fixture.userID, SessionID: fixture.sessionID,
		ClientMessageID: clientMessageID, Text: "可恢复回答", Now: fixture.now,
	})
	if err != nil || first.ReplyGeneration != 1 {
		t.Fatalf("initial reservation=%+v err=%v", first, err)
	}

	retryNow := fixture.now.Add(domain.PracticeReplyLeaseDuration)
	repos := openReplyLeaseRepositories(t, fixture.dsn, 2)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo *SQLRepository) {
			defer wg.Done()
			<-start
			reservation, err := repo.ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
				UserMessageID: replyLeaseUUID(3, 21+i), UserID: fixture.userID, SessionID: fixture.sessionID,
				ClientMessageID: clientMessageID, Text: "可恢复回答", Now: retryNow,
			})
			if err == nil && reservation.ReplyGeneration != 2 {
				err = fmt.Errorf("winning generation=%d want 2", reservation.ReplyGeneration)
			}
			results <- err
		}(i, repo)
	}
	close(start)
	wg.Wait()
	close(results)
	assertOneReservationWinner(t, results)
	assertReplyLeaseRow(t, fixture, 1, domain.PracticeReplyStatusPending, retryNow.Add(domain.PracticeReplyLeaseDuration), 2, 0)
}

func TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery(t *testing.T) {
	fixture := setupReplyLeaseIntegrationFixture(t, 4)
	repos := openReplyLeaseRepositories(t, fixture.dsn, 3)
	clientMessageID := replyLeaseUUID(4, 30)
	userMessageID := replyLeaseUUID(4, 20)
	first, err := repos[0].ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
		UserMessageID: userMessageID, UserID: fixture.userID, SessionID: fixture.sessionID,
		ClientMessageID: clientMessageID, Text: "跨代回答", Now: fixture.now,
	})
	if err != nil || first.ReplyGeneration != 1 {
		t.Fatalf("initial reservation=%+v err=%v", first, err)
	}
	recoveryNow := fixture.now.Add(domain.PracticeReplyLeaseDuration)
	session, err := repos[1].GetSession(fixture.ctx, fixture.userID, fixture.sessionID, recoveryNow)
	if err != nil || session.Messages[len(session.Messages)-1].ReplyStatus != domain.PracticeReplyStatusRetryableFailed {
		t.Fatalf("GET recovery=%+v err=%v", session, err)
	}
	second, err := repos[1].ReservePracticeMessage(fixture.ctx, domain.ReservePracticeMessageInput{
		UserMessageID: replyLeaseUUID(4, 21), UserID: fixture.userID, SessionID: fixture.sessionID,
		ClientMessageID: clientMessageID, Text: "跨代回答", Now: recoveryNow,
	})
	if err != nil || second.ReplyGeneration != 2 || second.UserMessage.ID != userMessageID {
		t.Fatalf("G2 reservation=%+v err=%v", second, err)
	}

	start := make(chan struct{})
	staleResults := make(chan error, 2)
	staleAssistantID := replyLeaseUUID(4, 40)
	validAssistantID := replyLeaseUUID(4, 41)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := repos[0].CommitPracticeMessage(fixture.ctx, domain.CommitPracticeMessageInput{
			UserID: fixture.userID, SessionID: fixture.sessionID, UserMessageID: userMessageID,
			ExpectedReplyGeneration: 1, AssistantMessageID: staleAssistantID, AssistantText: "迟到 G1", Now: recoveryNow,
		})
		staleResults <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		staleResults <- repos[2].FailPracticeMessage(fixture.ctx, domain.FailPracticeMessageInput{
			UserID: fixture.userID, SessionID: fixture.sessionID, UserMessageID: userMessageID,
			ExpectedReplyGeneration: 1, ReplyStatus: domain.PracticeReplyStatusRetryableFailed,
		})
	}()
	close(start)
	wg.Wait()
	close(staleResults)
	for staleErr := range staleResults {
		if !errors.Is(staleErr, domain.ErrSessionConflict) {
			t.Fatalf("stale worker error=%v want ErrSessionConflict", staleErr)
		}
	}

	_, err = repos[1].CommitPracticeMessage(fixture.ctx, domain.CommitPracticeMessageInput{
		UserID: fixture.userID, SessionID: fixture.sessionID, UserMessageID: userMessageID,
		ExpectedReplyGeneration: 2, AssistantMessageID: validAssistantID, AssistantText: "有效 G2", Now: recoveryNow.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("commit G2: %v", err)
	}
	assertReplyLeaseRow(t, fixture, 1, domain.PracticeReplyStatusComplete, time.Time{}, 2, 1)
	var staleReplies, validReplies int
	if err := fixture.db.QueryRowContext(fixture.ctx, `
select count(*) filter (where id=$2), count(*) filter (where id=$3 and reply_to_message_id=$4)
from practice_messages where session_id=$1`, fixture.sessionID, staleAssistantID, validAssistantID, userMessageID).Scan(&staleReplies, &validReplies); err != nil {
		t.Fatalf("count fenced assistant replies: %v", err)
	}
	if staleReplies != 0 || validReplies != 1 {
		t.Fatalf("assistant replies stale=%d valid=%d want 0/1", staleReplies, validReplies)
	}
}

func setupReplyLeaseIntegrationFixture(t *testing.T, namespace int) replyLeaseIntegrationFixture {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for the practice reply lease integration gate")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}
	userID := replyLeaseUUID(namespace, 1)
	resumeID := replyLeaseUUID(namespace, 2)
	targetID := replyLeaseUUID(namespace, 3)
	planID := replyLeaseUUID(namespace, 4)
	sessionID := replyLeaseUUID(namespace, 5)
	openingID := replyLeaseUUID(namespace, 6)
	now := time.Date(2026, 7, 14, 8, 0, namespace, 123000000, time.UTC)
	cleanup := func() { _, _ = db.ExecContext(context.Background(), `delete from users where id=$1`, userID) }
	cleanup()
	t.Cleanup(cleanup)
	mustExecReplyLease(t, ctx, db, `insert into users (id,email,display_name,created_at,updated_at) values ($1,$2,$3,$4,$4)`,
		userID, fmt.Sprintf("reply-lease-%d@example.test", namespace), "Reply Lease", now)
	mustExecReplyLease(t, ctx, db, `
insert into resumes (
  id,user_id,title,display_name,language,parse_status,parsed_summary,raw_text,
  source_type,original_text,parsed_text_snapshot,structured_profile,created_at,updated_at
) values ($1,$2,'Reply Lease Resume','Reply Lease Resume','zh-CN','ready','{}'::jsonb,'完整简历','paste','完整简历','完整简历','{}'::jsonb,$3,$3)`, resumeID, userID, now)
	summary := `{"interviewRounds":[{"sequence":1,"type":"technical","name":"技术面","durationMinutes":45,"focus":"系统设计"},{"sequence":2,"type":"manager","name":"主管面","durationMinutes":45,"focus":"团队协作"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"zh-CN","featureFlag":"none","dataSourceVersion":"target-job.v1"}}`
	mustExecReplyLease(t, ctx, db, `
insert into target_jobs (
  id,user_id,resume_id,status,analysis_status,title,target_language,raw_jd_text,summary,fit_summary,created_at,updated_at
) values ($1,$2,$3,'draft','ready','Platform Engineer','zh-CN','完整 JD',$4::jsonb,'{}'::jsonb,$5,$5)`, targetID, userID, resumeID, summary, now)
	mustExecReplyLease(t, ctx, db, `
insert into practice_plans (
  id,user_id,target_job_id,goal,interviewer_persona,difficulty,language,time_budget_minutes,
  resume_id,focus_dimension_codes,status,round_id,round_sequence,created_at,updated_at
) values ($1,$2,$3,'baseline','hiring_manager','standard','zh-CN',45,$4,'{}'::text[],'ready','round-1-technical',1,$5,$5)`, planID, userID, targetID, resumeID, now)
	mustExecReplyLease(t, ctx, db, `
insert into practice_sessions (id,user_id,plan_id,target_job_id,status,language,started_at,created_at,updated_at)
values ($1,$2,$3,$4,'running','zh-CN',$5,$5,$5)`, sessionID, userID, planID, targetID, now)
	mustExecReplyLease(t, ctx, db, `
insert into practice_messages (id,session_id,seq_no,role,content,created_at)
values ($1,$2,1,'assistant','请介绍一个项目。',$3)`, openingID, sessionID, now)
	return replyLeaseIntegrationFixture{dsn: dsn, db: db, ctx: ctx, userID: userID, sessionID: sessionID, now: now}
}

func openReplyLeaseRepositories(t *testing.T, dsn string, count int) []*SQLRepository {
	t.Helper()
	repos := make([]*SQLRepository, 0, count)
	backendPIDs := make(map[int]struct{}, count)
	for i := 0; i < count; i++ {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Fatalf("open independent postgres connection %d: %v", i, err)
		}
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		if err := db.Ping(); err != nil {
			db.Close()
			t.Fatalf("ping independent postgres connection %d: %v", i, err)
		}
		t.Cleanup(func() { db.Close() })
		var backendPID int
		if err := db.QueryRow(`select pg_backend_pid()`).Scan(&backendPID); err != nil {
			t.Fatalf("read independent postgres connection %d pid: %v", i, err)
		}
		if _, duplicate := backendPIDs[backendPID]; duplicate {
			t.Fatalf("postgres connection %d reused backend pid %d", i, backendPID)
		}
		backendPIDs[backendPID] = struct{}{}
		repos = append(repos, NewSQLRepository(db))
	}
	return repos
}

func assertOneReservationWinner(t *testing.T, results <-chan error) {
	t.Helper()
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrSessionConflict):
			conflicts++
		default:
			t.Fatalf("concurrent reservation error=%v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("reservation outcomes success=%d conflict=%d want 1/1", successes, conflicts)
	}
}

func assertReplyLeaseRow(
	t *testing.T,
	fixture replyLeaseIntegrationFixture,
	wantUsers int,
	wantStatus domain.PracticeReplyStatus,
	wantLease time.Time,
	wantGeneration int64,
	wantAssistants int,
) {
	t.Helper()
	var users, assistants int
	var status string
	var generation int64
	var lease sql.NullTime
	err := fixture.db.QueryRowContext(fixture.ctx, `
select count(*) filter (where role='user'),
       count(*) filter (where role='assistant' and reply_to_message_id is not null),
       coalesce(max(reply_status) filter (where role='user'),''),
       coalesce(max(reply_generation) filter (where role='user'),0),
       max(reply_lease_expires_at) filter (where role='user')
from practice_messages where session_id=$1`, fixture.sessionID).Scan(&users, &assistants, &status, &generation, &lease)
	if err != nil {
		t.Fatalf("read reply lease row: %v", err)
	}
	if users != wantUsers || assistants != wantAssistants || status != string(wantStatus) || generation != wantGeneration {
		t.Fatalf("reply row users=%d assistants=%d status=%s generation=%d", users, assistants, status, generation)
	}
	if wantLease.IsZero() {
		if lease.Valid {
			t.Fatalf("lease=%s want NULL", lease.Time)
		}
	} else if !lease.Valid || !lease.Time.Equal(wantLease) {
		t.Fatalf("lease=%v want %s", lease, wantLease)
	}
}

func mustExecReplyLease(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec reply lease fixture: %v", err)
	}
}

func replyLeaseUUID(namespace, offset int) string {
	return fmt.Sprintf("019f5b71-%04x-7000-8000-%012x", namespace, offset)
}
