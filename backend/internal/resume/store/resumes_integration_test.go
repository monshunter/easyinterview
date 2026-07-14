//go:build integration

package store_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/lib/pq"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func ptrString(v string) *string {
	return &v
}

func TestResumesIntegrationCRUDStateIsolationPaginationAndRollback(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureResumesErrorCodeColumn(t, ctx, db)

	repo := resumestore.NewRepository(db)
	now := time.Date(2026, 5, 13, 6, 0, 0, 0, time.UTC)
	userA := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a001"
	userB := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a002"
	fileID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a003"
	assetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a004"
	jobID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a005"
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userA, userB) })

	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-store-a@example.com', 'active'), ($2, 'resume-store-b@example.com', 'active')`, userA, userB)
	mustExec(t, ctx, db, `insert into file_objects(id, user_id, purpose, object_key, original_file_name, content_type, byte_size, upload_status) values ($1, $2, 'resume', 'user-a/resume/file.pdf', 'file.pdf', 'application/pdf', 128, 'uploaded')`, fileID, userA)

	created, err := repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
		AssetID:      assetID,
		UserID:       userA,
		JobID:        jobID,
		DedupeKey:    "resume-integration-dedupe",
		SourceType:   "upload",
		FileObjectID: &fileID,
		Title:        "Integration Resume",
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		JobStatus:    sharedtypes.JobStatusQueued,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("CreateWithParseJob: %v", err)
	}
	if created.AssetID != assetID || created.JobID != jobID {
		t.Fatalf("created = %+v", created)
	}

	got, err := repo.Get(ctx, userA, assetID)
	if err != nil {
		t.Fatalf("Get owner: %v", err)
	}
	if got.FileObjectID == nil || *got.FileObjectID != fileID || got.SourceType == nil || *got.SourceType != "upload" {
		t.Fatalf("asset file/source fields = %+v", got)
	}
	if _, err := repo.Get(ctx, userB, assetID); !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("cross-user Get err = %v, want ErrAssetNotFound", err)
	}

	if err := repo.MarkParsing(ctx, resumestore.StatusUpdateInput{UserID: userA, AssetID: assetID, Now: now.Add(time.Minute)}); err != nil {
		t.Fatalf("MarkParsing: %v", err)
	}
	if err := repo.CompleteParseSuccess(ctx, resumestore.CompleteParseSuccessInput{
		UserID:             userA,
		AssetID:            assetID,
		ParsedSummary:      []byte(`{"basics":{"name":"Alice"}}`),
		StructuredProfile:  []byte(`{"basics":{"name":"Alice"}}`),
		ParsedTextSnapshot: "parsed",
		DisplayName:        ptrString("Alice - Staff Engineer"),
		OutboxEventID:      "0195f2d0-4a44-7fc2-8f77-1f9c4cf1a006",
		OutboxEventPayload: []byte(`{"resumeId":"0195f2d0-4a44-7fc2-8f77-1f9c4cf1a004","userId":"0195f2d0-4a44-7fc2-8f77-1f9c4cf1a001","parseStatus":"ready"}`),
		Now:                now.Add(2 * time.Minute),
	}); err != nil {
		t.Fatalf("CompleteParseSuccess: %v", err)
	}
	ready, err := repo.Get(ctx, userA, assetID)
	if err != nil {
		t.Fatalf("Get ready: %v", err)
	}
	if ready.ParseStatus != sharedtypes.TargetJobParseStatusReady || ready.ParsedTextSnapshot == nil || *ready.ParsedTextSnapshot != "parsed" {
		t.Fatalf("ready resume = %+v", ready)
	}
	if ready.DisplayName == nil || *ready.DisplayName != "Alice - Staff Engineer" {
		t.Fatalf("ready display name = %#v", ready.DisplayName)
	}
	assertJSONEqual(t, ready.StructuredProfile, []byte(`{"basics":{"name":"Alice"}}`), "ready structured_profile")
	var count int
	if err := db.QueryRowContext(ctx, `select count(*) from outbox_events where aggregate_id = $1 and event_name = 'resume.parse.completed' and aggregate_type = 'resume'`, assetID).Scan(&count); err != nil {
		t.Fatalf("count completed outbox: %v", err)
	}
	if count != 1 {
		t.Fatalf("resume.parse.completed outbox count = %d, want 1", count)
	}

	for i := 0; i < 24; i++ {
		id := resumeIntegrationID(i)
		job := resumeIntegrationJobID(i)
		_, err := repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
			AssetID:     id,
			UserID:      userA,
			JobID:       job,
			DedupeKey:   "resume-integration-dedupe-" + id,
			SourceType:  "paste",
			Title:       "Paged Resume",
			Language:    "en",
			RawText:     "resume text",
			ParseStatus: sharedtypes.TargetJobParseStatusQueued,
			JobStatus:   sharedtypes.JobStatusQueued,
			Now:         now.Add(time.Duration(i+3) * time.Minute),
		})
		if err != nil {
			t.Fatalf("CreateWithParseJob page row %d: %v", i, err)
		}
	}
	first, err := repo.List(ctx, userA, resumestore.ListFilter{PageSize: 20})
	if err != nil {
		t.Fatalf("List first page: %v", err)
	}
	if len(first.Items) != 20 || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("first page len=%d hasMore=%v cursor=%q", len(first.Items), first.HasMore, first.NextCursor)
	}
	second, err := repo.List(ctx, userA, resumestore.ListFilter{PageSize: 20, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("List second page: %v", err)
	}
	if len(second.Items) != 5 || second.HasMore {
		t.Fatalf("second page len=%d hasMore=%v", len(second.Items), second.HasMore)
	}

	badAssetID := "0195f2d0-4a44-7fc2-8f77-1f9c4cf1aff1"
	_, err = repo.CreateWithParseJob(ctx, resumestore.CreateAssetInput{
		AssetID:     badAssetID,
		UserID:      userA,
		JobID:       "0195f2d0-4a44-7fc2-8f77-1f9c4cf1aff2",
		DedupeKey:   "resume-integration-bad-job",
		SourceType:  "paste",
		Title:       "Rollback Resume",
		Language:    "en",
		RawText:     "resume text",
		ParseStatus: sharedtypes.TargetJobParseStatusQueued,
		JobStatus:   "not_allowed",
		Now:         now,
	})
	if err == nil {
		t.Fatal("expected rollback error for invalid async job status")
	}
	if err := db.QueryRowContext(ctx, `select count(*) from resumes where id = $1`, badAssetID).Scan(&count); err != nil {
		t.Fatalf("count rollback resume: %v", err)
	}
	if count != 0 {
		t.Fatalf("rollback resume count = %d, want 0", count)
	}
}

func TestCreateWithParseJobSerializesActiveLimitPerUser(t *testing.T) {
	db := openResumeStoreTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ensureResumesErrorCodeColumn(t, ctx, db)

	const (
		userID    = "019f6000-0000-7000-8000-000000000001"
		maxActive = 1
	)
	t.Cleanup(func() { cleanupResumeStoreUsers(t, db, userID) })
	mustExec(t, ctx, db, `insert into users(id, email, status) values ($1, 'resume-limit-concurrency@example.com', 'active')`, userID)
	for i := 0; i < maxActive-1; i++ {
		mustExec(t, ctx, db, `
insert into resumes(id, user_id, title, language, parse_status, source_type, original_text)
values ($1, $2, 'Baseline Resume', 'en', 'queued', 'paste', 'resume text')`,
			fmt.Sprintf("019f6000-0000-7000-8100-%012d", i+1), userID)
	}

	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin resume insert blocker: %v", err)
	}
	defer blocker.Rollback()
	if _, err := blocker.ExecContext(ctx, `lock table resumes in share mode`); err != nil {
		t.Fatalf("lock resumes table: %v", err)
	}

	const appA = "resume-limit-concurrency-a"
	const appB = "resume-limit-concurrency-b"
	dbA := openNamedResumeStoreTestDB(t, appA)
	dbB := openNamedResumeStoreTestDB(t, appB)
	repos := []*resumestore.Repository{resumestore.NewRepository(dbA), resumestore.NewRepository(dbB)}
	inputs := []resumestore.CreateAssetInput{
		{
			AssetID:          "019f6000-0000-7000-8200-000000000001",
			UserID:           userID,
			JobID:            "019f6000-0000-7000-8300-000000000001",
			DedupeKey:        "resume-limit-concurrency-ik-a",
			SourceType:       "paste",
			Title:            "Concurrent Resume A",
			Language:         "en",
			RawText:          "resume text a",
			MaxActiveForUser: maxActive,
			Now:              time.Date(2026, 7, 14, 21, 0, 0, 0, time.UTC),
		},
		{
			AssetID:          "019f6000-0000-7000-8200-000000000002",
			UserID:           userID,
			JobID:            "019f6000-0000-7000-8300-000000000002",
			DedupeKey:        "resume-limit-concurrency-ik-b",
			SourceType:       "paste",
			Title:            "Concurrent Resume B",
			Language:         "en",
			RawText:          "resume text b",
			MaxActiveForUser: maxActive,
			Now:              time.Date(2026, 7, 14, 21, 0, 1, 0, time.UTC),
		},
	}

	start := make(chan struct{})
	results := make(chan error, len(inputs))
	for i := range inputs {
		go func(i int) {
			<-start
			_, err := repos[i].CreateWithParseJob(ctx, inputs[i])
			results <- err
		}(i)
	}
	close(start)
	waitForBlockedResumeLimitContenders(t, ctx, db, appA, appB)
	if err := blocker.Commit(); err != nil {
		t.Fatalf("release resume insert blocker: %v", err)
	}

	succeeded := 0
	limitRejected := 0
	for range inputs {
		err := <-results
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, resumestore.ErrResumeLimitExceeded):
			limitRejected++
		default:
			t.Fatalf("concurrent CreateWithParseJob error = %v", err)
		}
	}
	if succeeded != 1 || limitRejected != 1 {
		t.Fatalf("concurrent outcomes: succeeded=%d limitRejected=%d, want 1/1", succeeded, limitRejected)
	}

	var active int
	if err := db.QueryRowContext(ctx, `select count(*) from resumes where user_id = $1 and deleted_at is null`, userID).Scan(&active); err != nil {
		t.Fatalf("count active resumes: %v", err)
	}
	if active != maxActive {
		t.Fatalf("active resumes = %d, want %d", active, maxActive)
	}
	var jobs int
	if err := db.QueryRowContext(ctx, `select count(*) from async_jobs where id in ($1, $2)`, inputs[0].JobID, inputs[1].JobID).Scan(&jobs); err != nil {
		t.Fatalf("count concurrent parse jobs: %v", err)
	}
	if jobs != 1 {
		t.Fatalf("concurrent parse jobs = %d, want 1", jobs)
	}
}

func openResumeStoreTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping resumes integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	return db
}

func openNamedResumeStoreTestDB(t *testing.T, applicationName string) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for named resume store integration connection")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	query := parsed.Query()
	query.Set("application_name", applicationName)
	parsed.RawQuery = query.Encode()
	db, err := sql.Open("postgres", parsed.String())
	if err != nil {
		t.Fatalf("open named postgres connection: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping named postgres connection: %v", err)
	}
	return db
}

func waitForBlockedResumeLimitContenders(t *testing.T, ctx context.Context, db *sql.DB, appA, appB string) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked int
		if err := db.QueryRowContext(ctx, `
select count(*)
from pg_stat_activity
where application_name in ($1, $2)
  and state = 'active'
  and wait_event_type = 'Lock'`, appA, appB).Scan(&blocked); err != nil {
			t.Fatalf("inspect concurrent resume limit transactions: %v", err)
		}
		if blocked == 2 {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for concurrent resume limit transactions: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}

func ensureResumesErrorCodeColumn(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var exists bool
	if err := db.QueryRowContext(ctx, `select exists(select 1 from information_schema.columns where table_name = 'resumes' and column_name = 'error_code')`).Scan(&exists); err != nil {
		t.Fatalf("check resumes.error_code: %v", err)
	}
	if !exists {
		t.Skip("resumes.error_code is not migrated; run make migrate-up before live resume store test")
	}
}

func cleanupResumeStoreUsers(t *testing.T, db *sql.DB, userIDs ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, userID := range userIDs {
		_, _ = db.ExecContext(ctx, `delete from outbox_events where aggregate_id in (select id from resumes where user_id = $1)`, userID)
		_, _ = db.ExecContext(ctx, `delete from ai_task_runs where user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `delete from async_jobs where resource_id in (select id from resumes where user_id = $1)`, userID)
		_, _ = db.ExecContext(ctx, `delete from resumes where user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `delete from file_objects where user_id = $1`, userID)
		_, _ = db.ExecContext(ctx, `delete from users where id = $1`, userID)
	}
}

func mustExec(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func assertJSONEqual(t *testing.T, got []byte, want []byte, label string) {
	t.Helper()
	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("%s invalid JSON: %v", label, err)
	}
	var wantValue any
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("%s invalid expected JSON: %v", label, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("%s = %s, want %s", label, got, want)
	}
}

func resumeIntegrationID(i int) string {
	return fmt.Sprintf("0195f2d0-4a44-7fc2-8f77-1f9c4cf1b%03d", i)
}

func resumeIntegrationJobID(i int) string {
	return fmt.Sprintf("0195f2d0-4a44-7fc2-8f77-1f9c4cf1c%03d", i)
}
