package targetjob_test

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func reflectStoreType() reflect.Type {
	return reflect.TypeFor[targetjob.Store]()
}

func storeHasMethod(t reflect.Type, name string) bool {
	for i := 0; i < t.NumMethod(); i++ {
		if t.Method(i).Name == name {
			return true
		}
	}
	return false
}

func newMockStore(t *testing.T) (*targetjob.SQLStore, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return targetjob.NewSQLStore(db), mock, func() { db.Close() }
}

func TestSQLStore_InsertTargetJob_WritesAllColumnsAndDefaultsJSON(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	mock.ExpectExec("insert into target_jobs").
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000a1", // id
			"018f2a40-0000-7000-9000-0000000000b1", // user_id
			nil,                                    // profile_id (empty)
			"draft",                                // status
			"queued",                               // analysis_status
			"Backend Engineer",                     // title
			"Acme",                                 // company_name
			nil,                                    // location_text
			nil,                                    // employment_type
			nil,                                    // seniority_level
			"en",                                   // target_language
			"manual_text",                          // source_type
			nil,                                    // source_url
			nil,                                    // source_file_object_id
			"raw jd",                               // raw_jd_text
			[]byte(`{}`),                           // summary
			[]byte(`{}`),                           // fit_summary
			nil,                                    // notes
			int32(0),                               // open_question_issue_count
			now,                                    // created_at
			now,                                    // updated_at
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		Status:         sharedtypes.TargetJobStatusDraft,
		AnalysisStatus: sharedtypes.TargetJobParseStatusQueued,
		Title:          "Backend Engineer",
		CompanyName:    "Acme",
		TargetLanguage: "en",
		SourceType:     targetjob.SourceTypeManualText,
		RawJDText:      "raw jd",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := store.InsertTargetJob(context.Background(), rec); err != nil {
		t.Fatalf("InsertTargetJob: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_InsertTargetJobSource_DefaultsFreshAndPicksFetchedAt(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 10, 5, 0, 0, time.UTC)
	fetchedAt := now.Add(-time.Minute)
	mock.ExpectExec("insert into target_job_sources").
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000c1",
			"018f2a40-0000-7000-9000-0000000000a1",
			"url",
			"https://jobs.example.com/123",
			nil,
			"snapshot text",
			fetchedAt,
			"fresh",
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rec := targetjob.SourceRecord{
		ID:           "018f2a40-0000-7000-9000-0000000000c1",
		TargetJobID:  "018f2a40-0000-7000-9000-0000000000a1",
		SourceType:   targetjob.SourceTypeURL,
		URL:          "https://jobs.example.com/123",
		SnapshotText: "snapshot text",
		FetchedAt:    &fetchedAt,
		CreatedAt:    now,
	}
	if err := store.InsertTargetJobSource(context.Background(), rec); err != nil {
		t.Fatalf("InsertTargetJobSource: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_GetTargetJobByUser_ReturnsRecordWithRequirementsAndSources(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 11, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	}).AddRow(
		"018f2a40-0000-7000-9000-0000000000a1",
		"018f2a40-0000-7000-9000-0000000000b1",
		nil, "draft", "ready",
		"Backend Engineer", "Acme", nil, nil, nil,
		"en", "manual_text", nil, nil,
		"raw jd",
		[]byte(`{"coreThemes":["api"]}`),
		[]byte(`{}`),
		nil, nil, int32(0),
		now, now,
	)
	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1", "018f2a40-0000-7000-9000-0000000000b1").
		WillReturnRows(rows)

	mock.ExpectQuery(`from target_job_requirements`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "kind", "label", "description", "evidence_level", "display_order", "created_at",
		}).AddRow(
			"018f2a40-0000-7000-9000-0000000000d1",
			"018f2a40-0000-7000-9000-0000000000a1",
			"must_have", "Go", nil, "explicit", int32(1), now,
		))

	mock.ExpectQuery(`from target_job_sources`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "source_type", "url", "file_object_id", "snapshot_text", "fetched_at", "freshness_status", "created_at",
		}).AddRow(
			"018f2a40-0000-7000-9000-0000000000c1",
			"018f2a40-0000-7000-9000-0000000000a1",
			"manual_text", nil, nil, "raw jd", nil, "fresh", now,
		))

	got, reqs, sources, err := store.GetTargetJobByUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		"018f2a40-0000-7000-9000-0000000000a1",
	)
	if err != nil {
		t.Fatalf("GetTargetJobByUser: %v", err)
	}
	if got.Title != "Backend Engineer" || got.Status != sharedtypes.TargetJobStatusDraft || got.AnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("unexpected record: %+v", got)
	}
	if string(got.Summary) != `{"coreThemes":["api"]}` {
		t.Fatalf("summary not preserved: %q", got.Summary)
	}
	if len(reqs) != 1 || reqs[0].Kind != targetjob.RequirementMustHave || reqs[0].Label != "Go" {
		t.Fatalf("requirements unexpected: %+v", reqs)
	}
	if len(sources) != 1 || sources[0].SourceType != targetjob.SourceTypeManualText {
		t.Fatalf("sources unexpected: %+v", sources)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_GetTargetJobByUser_NotFoundForCrossUserOrSoftDeleted(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1", "018f2a40-0000-7000-9000-0000000000b9").
		WillReturnError(sql.ErrNoRows)

	_, _, _, err := store.GetTargetJobByUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b9", // a different user
		"018f2a40-0000-7000-9000-0000000000a1",
	)
	if !errors.Is(err, targetjob.ErrTargetJobNotFound) {
		t.Fatalf("expected ErrTargetJobNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ListTargetJobsForUser_AppliesFiltersAndClampsPageSize(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	status := sharedtypes.TargetJobStatusPreparing
	analysis := sharedtypes.TargetJobParseStatusReady

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	})

	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	for range 1 {
		rows.AddRow(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
			nil, "preparing", "ready",
			"Backend Engineer", "Acme", nil, nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), nil, nil, int32(0),
			now, now,
		)
	}

	mock.ExpectQuery(`from target_jobs\s+where user_id = \$1 and deleted_at is null and status = \$2 and analysis_status = \$3 and to_tsvector\('simple'.*plainto_tsquery\('simple', \$4\)\s+order by updated_at desc, id desc\s+limit \$5`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000b1",
			"preparing",
			"ready",
			"backend",
			int(targetjob.ListMaxPageSize+1), // pageSize 1000 clamped to 100, +1 sentinel
		).
		WillReturnRows(rows)

	res, err := store.ListTargetJobsForUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		targetjob.ListFilter{
			Status:         &status,
			AnalysisStatus: &analysis,
			SearchQuery:    "  backend  ",
			PageSize:       1000,
		})
	if err != nil {
		t.Fatalf("ListTargetJobsForUser: %v", err)
	}
	if res.HasMore || res.NextCursor != "" {
		t.Fatalf("expected no next cursor for short page, got %+v", res)
	}
	if len(res.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(res.Items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ListTargetJobsForUser_PaginationCursorOnOverflow(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 13, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	})
	for i := range 3 {
		rows.AddRow(
			"018f2a40-0000-7000-9000-00000000000"+string(rune('a'+i)),
			"018f2a40-0000-7000-9000-0000000000b1",
			nil, "draft", "ready",
			"Backend", "Acme", nil, nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), nil, nil, int32(0),
			now.Add(-time.Duration(i)*time.Minute),
			now.Add(-time.Duration(i)*time.Minute),
		)
	}

	mock.ExpectQuery(`from target_jobs\s+where user_id = \$1 and deleted_at is null\s+order by updated_at desc, id desc\s+limit \$2`).
		WithArgs("018f2a40-0000-7000-9000-0000000000b1", int(3)).
		WillReturnRows(rows)

	res, err := store.ListTargetJobsForUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		targetjob.ListFilter{PageSize: 2})
	if err != nil {
		t.Fatalf("ListTargetJobsForUser: %v", err)
	}
	if !res.HasMore || res.NextCursor == "" {
		t.Fatalf("expected next cursor on overflow, got %+v", res)
	}
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 items kept after overflow trim, got %d", len(res.Items))
	}
	// cursor decodes to (updated_at_of_last_item, id_of_last_item)
	raw, err := base64.RawURLEncoding.DecodeString(res.NextCursor)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if !strings.Contains(string(raw), res.Items[1].ID) {
		t.Fatalf("cursor missing id of last kept row: %q", string(raw))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ListTargetJobsForUser_RejectsBadCursor(t *testing.T) {
	store, _, cleanup := newMockStore(t)
	defer cleanup()
	_, err := store.ListTargetJobsForUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		targetjob.ListFilter{Cursor: "not-base64-/!"},
	)
	if err == nil {
		t.Fatal("expected cursor decode error, got nil")
	}
}

func TestSQLStore_UpdateTargetJobLifecycle_ScopesByUser_ReturnsRow(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	rowCols := []string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null\s+for update`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
		).
		WillReturnRows(sqlmock.NewRows(rowCols).AddRow(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
			nil, "draft", "ready",
			"Backend Engineer", "Acme", nil, nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), nil, nil, int32(0),
			now, now,
		))
	mock.ExpectQuery(`update target_jobs\s+set status = \$1, location_text = \$2, notes = \$3, updated_at = \$4\s+where id = \$5 and user_id = \$6 and deleted_at is null\s+returning`).
		WithArgs("preparing", "Remote", "applied via portal", now,
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
		).
		WillReturnRows(sqlmock.NewRows(rowCols).AddRow(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
			nil, "preparing", "ready",
			"Backend Engineer", "Acme", "Remote", nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), "applied via portal", nil, int32(0),
			now, now,
		))
	mock.ExpectCommit()

	status := sharedtypes.TargetJobStatusPreparing
	loc := "Remote"
	notes := "applied via portal"
	rec, err := store.UpdateTargetJobLifecycle(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		"018f2a40-0000-7000-9000-0000000000a1",
		targetjob.UpdateLifecycleFields{Status: &status, LocationText: &loc, Notes: &notes},
		now,
	)
	if err != nil {
		t.Fatalf("UpdateTargetJobLifecycle: %v", err)
	}
	if rec.Status != sharedtypes.TargetJobStatusPreparing || rec.LocationText != "Remote" || rec.Notes != "applied via portal" {
		t.Fatalf("unexpected returned row: %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_UpdateTargetJobLifecycle_OverwritesTitleAndCompanyHints(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 14, 5, 0, 0, time.UTC)
	rowCols := []string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	}
	mock.ExpectQuery(`update target_jobs\s+set title = \$1, company_name = \$2, updated_at = \$3\s+where id = \$4 and user_id = \$5 and deleted_at is null\s+returning`).
		WithArgs("Senior Frontend Engineer", "Acme Labs", now,
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
		).
		WillReturnRows(sqlmock.NewRows(rowCols).AddRow(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
			nil, "draft", "ready",
			"Senior Frontend Engineer", "Acme Labs", nil, nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), nil, nil, int32(0),
			now, now,
		))

	title := "Senior Frontend Engineer"
	company := "Acme Labs"
	rec, err := store.UpdateTargetJobLifecycle(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		"018f2a40-0000-7000-9000-0000000000a1",
		targetjob.UpdateLifecycleFields{TitleHint: &title, CompanyNameHint: &company},
		now,
	)
	if err != nil {
		t.Fatalf("UpdateTargetJobLifecycle: %v", err)
	}
	if rec.Title != title || rec.CompanyName != company {
		t.Fatalf("title/company hints were not persisted: %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_UpdateTargetJobLifecycle_DedupeHitReturnsExistingWithoutMutation(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 14, 15, 0, 0, time.UTC)
	dedupeKey := "targetjob:update:user1:key1"
	targetID := "018f2a40-0000-7000-9000-0000000000a1"
	userID := "018f2a40-0000-7000-9000-0000000000b1"
	rowCols := []string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnRows(sqlmock.NewRows([]string{"resource_id"}).AddRow(targetID))
	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null`).
		WithArgs(targetID, userID).
		WillReturnRows(sqlmock.NewRows(rowCols).AddRow(
			targetID,
			userID,
			nil, "preparing", "ready",
			"Backend Engineer", "Acme", "Remote", nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), "already updated", nil, int32(0),
			now, now,
		))
	mock.ExpectCommit()

	status := sharedtypes.TargetJobStatusApplied
	rec, err := store.UpdateTargetJobLifecycle(context.Background(),
		userID,
		targetID,
		targetjob.UpdateLifecycleFields{
			Status:         &status,
			DedupeKey:      dedupeKey,
			DedupeMarkerID: "018f2a40-0000-7000-9000-0000000000d1",
		},
		now,
	)
	if err != nil {
		t.Fatalf("UpdateTargetJobLifecycle dedupe hit: %v", err)
	}
	if rec.Status != sharedtypes.TargetJobStatusPreparing || rec.Notes != "already updated" {
		t.Fatalf("dedupe hit returned wrong row: %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_UpdateTargetJobLifecycle_IdempotentRejectsStaleStatusTransition(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 14, 20, 0, 0, time.UTC)
	dedupeKey := "targetjob:update:user1:key2"
	targetID := "018f2a40-0000-7000-9000-0000000000a1"
	userID := "018f2a40-0000-7000-9000-0000000000b1"
	rowCols := []string{
		"id", "user_id", "profile_id", "status", "analysis_status", "title", "company_name", "location_text",
		"employment_type", "seniority_level", "target_language", "source_type", "source_url", "source_file_object_id",
		"raw_jd_text", "summary", "fit_summary", "notes", "latest_report_id", "open_question_issue_count",
		"created_at", "updated_at",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`pg_advisory_xact_lock`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null\s+for update`).
		WithArgs(targetID, userID).
		WillReturnRows(sqlmock.NewRows(rowCols).AddRow(
			targetID,
			userID,
			nil, "applied", "ready",
			"Backend Engineer", "Acme", "Remote", nil, nil,
			"en", "manual_text", nil, nil,
			"raw jd", []byte(`{}`), []byte(`{}`), "current state changed", nil, int32(0),
			now, now,
		))
	mock.ExpectRollback()

	status := sharedtypes.TargetJobStatusOffer
	_, err := store.UpdateTargetJobLifecycle(context.Background(),
		userID,
		targetID,
		targetjob.UpdateLifecycleFields{
			Status:         &status,
			DedupeKey:      dedupeKey,
			DedupeMarkerID: "018f2a40-0000-7000-9000-0000000000d2",
		},
		now,
	)
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_INVALID_STATE_TRANSITION" {
		t.Fatalf("expected TARGET_INVALID_STATE_TRANSITION from locked current row, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_UpdateTargetJobLifecycle_NotFoundForCrossUser(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 14, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(`from target_jobs\s+where id = \$1 and user_id = \$2 and deleted_at is null\s+for update`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b9",
		).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	status := sharedtypes.TargetJobStatusPreparing
	_, err := store.UpdateTargetJobLifecycle(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b9",
		"018f2a40-0000-7000-9000-0000000000a1",
		targetjob.UpdateLifecycleFields{Status: &status},
		now,
	)
	if !errors.Is(err, targetjob.ErrTargetJobNotFound) {
		t.Fatalf("expected ErrTargetJobNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ApplyParseResult_MergesByKindLabelAndAccumulatesDisplayOrder(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 15, 0, 0, 0, time.UTC)

	mock.ExpectBegin()

	// Existing requirement: must_have + Go @ display_order=1
	mock.ExpectQuery(`from target_job_requirements`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "target_job_id", "kind", "label", "description", "evidence_level", "display_order", "created_at",
		}).AddRow(
			"018f2a40-0000-7000-9000-0000000000d1",
			"018f2a40-0000-7000-9000-0000000000a1",
			"must_have", "Go", nil, "explicit", int32(1), now,
		))

	// Duplicate "must_have / Go" must NOT be inserted; only the new "interview_focus / system design" gets inserted at display_order 2.
	mock.ExpectExec(`insert into target_job_requirements`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000d3",
			"018f2a40-0000-7000-9000-0000000000a1",
			"interview_focus",
			"system design",
			nil,
			"explicit",
			int32(2),
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`update target_jobs\s+set analysis_status = \$1,\s+summary = \$2,\s+fit_summary = \$3,\s+updated_at = \$4\s+where id = \$5 and deleted_at is null`).
		WithArgs("ready",
			[]byte(`{"coreThemes":["api"]}`),
			[]byte(`{}`),
			now,
			"018f2a40-0000-7000-9000-0000000000a1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	in := targetjob.ApplyParseResultInput{
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Summary:        []byte(`{"coreThemes":["api"]}`),
		Requirements: []targetjob.RequirementRecord{
			{ID: "018f2a40-0000-7000-9000-0000000000d2", Kind: targetjob.RequirementMustHave, Label: "Go"},
			{ID: "018f2a40-0000-7000-9000-0000000000d3", Kind: targetjob.RequirementInterviewFocus, Label: "system design"},
		},
		Now: now,
	}
	if err := store.ApplyParseResult(context.Background(), in); err != nil {
		t.Fatalf("ApplyParseResult: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_CompleteParseSuccess_WritesReadyStateParsedOutboxAndSourceRefreshAtomically(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 15, 10, 0, 0, time.UTC)
	targetID := "018f2a40-0000-7000-9000-0000000000a1"

	mock.ExpectBegin()
	mock.ExpectQuery(`from target_job_requirements`).
		WithArgs(targetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "kind", "label", "description", "evidence_level", "display_order", "created_at"}))
	mock.ExpectExec(`insert into target_job_requirements`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000d1",
			targetID,
			"must_have",
			"Go",
			nil,
			"explicit",
			int32(1),
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update target_jobs\s+set analysis_status = \$1,\s+summary = \$2,\s+fit_summary = \$3,\s+updated_at = \$4\s+where id = \$5 and deleted_at is null`).
		WithArgs("ready", []byte(`{"coreThemes":["api"]}`), []byte(`{}`), now, targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000e1",
			"target.parsed",
			targetID,
			[]byte(`{"targetJobId":"018f2a40-0000-7000-9000-0000000000a1"}`),
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000f2",
			"source_refresh",
			targetID,
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.CompleteParseSuccess(context.Background(), targetjob.CompleteParseSuccessInput{
		TargetJobID:        targetID,
		AnalysisStatus:     sharedtypes.TargetJobParseStatusReady,
		Summary:            []byte(`{"coreThemes":["api"]}`),
		FitSummary:         []byte(`{}`),
		ParsedEventID:      "018f2a40-0000-7000-9000-0000000000e1",
		ParsedEventPayload: []byte(`{"targetJobId":"018f2a40-0000-7000-9000-0000000000a1"}`),
		SourceRefreshJobID: "018f2a40-0000-7000-9000-0000000000f2",
		Requirements: []targetjob.RequirementRecord{
			{ID: "018f2a40-0000-7000-9000-0000000000d1", Kind: targetjob.RequirementMustHave, Label: "Go"},
		},
		Now: now,
	})
	if err != nil {
		t.Fatalf("CompleteParseSuccess: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_CompleteParseSuccess_RollsBackWhenParsedOutboxInsertFails(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 15, 20, 0, 0, time.UTC)
	targetID := "018f2a40-0000-7000-9000-0000000000a1"

	mock.ExpectBegin()
	mock.ExpectQuery(`from target_job_requirements`).
		WithArgs(targetID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "kind", "label", "description", "evidence_level", "display_order", "created_at"}))
	mock.ExpectExec(`update target_jobs\s+set analysis_status = \$1,\s+summary = \$2,\s+fit_summary = \$3,\s+updated_at = \$4\s+where id = \$5 and deleted_at is null`).
		WithArgs("ready", []byte(`{}`), []byte(`{}`), now, targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs("018f2a40-0000-7000-9000-0000000000e1", "target.parsed", targetID, []byte(`{}`), now).
		WillReturnError(errors.New("outbox unavailable"))
	mock.ExpectRollback()

	err := store.CompleteParseSuccess(context.Background(), targetjob.CompleteParseSuccessInput{
		TargetJobID:        targetID,
		AnalysisStatus:     sharedtypes.TargetJobParseStatusReady,
		Summary:            []byte(`{}`),
		FitSummary:         []byte(`{}`),
		ParsedEventID:      "018f2a40-0000-7000-9000-0000000000e1",
		ParsedEventPayload: []byte(`{}`),
		SourceRefreshJobID: "018f2a40-0000-7000-9000-0000000000f2",
		Now:                now,
	})
	if err == nil {
		t.Fatal("expected outbox failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_CompleteParseFailure_WritesFailedStateAndOutboxAtomically(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 15, 25, 0, 0, time.UTC)
	targetID := "018f2a40-0000-7000-9000-0000000000a1"

	mock.ExpectBegin()
	mock.ExpectExec(`update target_jobs\s+set analysis_status = 'failed', updated_at = \$1\s+where id = \$2 and deleted_at is null`).
		WithArgs(now, targetID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000e2",
			"target.analysis.failed",
			targetID,
			[]byte(`{"targetJobId":"018f2a40-0000-7000-9000-0000000000a1","errorCode":"AI_OUTPUT_INVALID"}`),
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.CompleteParseFailure(context.Background(), targetjob.CompleteParseFailureInput{
		TargetJobID:        targetID,
		FailedEventID:      "018f2a40-0000-7000-9000-0000000000e2",
		FailedEventPayload: []byte(`{"targetJobId":"018f2a40-0000-7000-9000-0000000000a1","errorCode":"AI_OUTPUT_INVALID"}`),
		Now:                now,
	})
	if err != nil {
		t.Fatalf("CompleteParseFailure: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ApplyParseResult_NotFoundForSoftDeletedTarget(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(`from target_job_requirements`).
		WithArgs("018f2a40-0000-7000-9000-0000000000a1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "target_job_id", "kind", "label", "description", "evidence_level", "display_order", "created_at"}))
	mock.ExpectExec(`update target_jobs\s+set analysis_status = \$1`).
		WithArgs("failed",
			[]byte(`{}`),
			[]byte(`{}`),
			now,
			"018f2a40-0000-7000-9000-0000000000a1",
		).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := store.ApplyParseResult(context.Background(), targetjob.ApplyParseResultInput{
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		AnalysisStatus: sharedtypes.TargetJobParseStatusFailed,
		Now:            now,
	})
	if !errors.Is(err, targetjob.ErrTargetJobNotFound) {
		t.Fatalf("expected ErrTargetJobNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_UpdateSourceFreshness_StalesAllSourcesForTarget(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 16, 0, 0, 0, time.UTC)
	mock.ExpectExec(`update target_job_sources\s+set freshness_status = \$1,\s+fetched_at = \$2\s+where target_job_id = \$3`).
		WithArgs("stale", now, "018f2a40-0000-7000-9000-0000000000a1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.UpdateSourceFreshness(context.Background(),
		"018f2a40-0000-7000-9000-0000000000a1",
		targetjob.FreshnessStale,
		now,
	); err != nil {
		t.Fatalf("UpdateSourceFreshness: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ImportTargetJob_RunnerBoundPathInsertsAllFourTables(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 17, 0, 0, 0, time.UTC)
	dedupeKey := "tj:dedupe:user1:key1"

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock\(hashtext\(\$1\)\)`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`insert into target_jobs`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000a1",
			"018f2a40-0000-7000-9000-0000000000b1",
			"draft", "queued",
			"Backend Engineer", "Acme",
			nil, nil, nil,
			"en", "manual_text",
			nil, nil,
			"raw jd",
			[]byte(`{}`), []byte(`{}`),
			int32(0), now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into target_job_sources`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000c1",
			"018f2a40-0000-7000-9000-0000000000a1",
			"manual_text", nil, nil, "raw jd", nil, "fresh", now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000e1",
			"target.import.requested", 1,
			"target_job",
			"018f2a40-0000-7000-9000-0000000000a1",
			[]byte(`{"sourceType":"text"}`),
			"pending", now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000f1",
			"target_import", "target_job",
			"018f2a40-0000-7000-9000-0000000000a1",
			dedupeKey, "queued",
			[]byte(`{"targetJobId":"x"}`),
			now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := store.ImportTargetJob(context.Background(), targetjob.ImportTargetJobInput{
		UserID:                 "018f2a40-0000-7000-9000-0000000000b1",
		DedupeKey:              dedupeKey,
		TargetJobID:            "018f2a40-0000-7000-9000-0000000000a1",
		Title:                  "Backend Engineer",
		CompanyName:            "Acme",
		TargetLanguage:         "en",
		APISourceType:          targetjob.SourceTypeManualText,
		RawJDText:              "raw jd",
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		InitialAnalysisStatus:  sharedtypes.TargetJobParseStatusQueued,
		SourceID:               "018f2a40-0000-7000-9000-0000000000c1",
		SourceSnapshotText:     "raw jd",
		JobID:                  "018f2a40-0000-7000-9000-0000000000f1",
		OutboxEventID:          "018f2a40-0000-7000-9000-0000000000e1",
		OutboxEventPayload:     []byte(`{"sourceType":"text"}`),
		JobPayload:             []byte(`{"targetJobId":"x"}`),
		Now:                    now,
	})
	if err != nil {
		t.Fatalf("ImportTargetJob: %v", err)
	}
	if res.Existing {
		t.Fatal("expected fresh import, got existing=true")
	}
	if res.JobStatus != sharedtypes.JobStatusQueued {
		t.Fatalf("expected queued status, got %q", res.JobStatus)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ImportTargetJob_ManualFormSyncSucceededAndNoOutbox(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 17, 30, 0, 0, time.UTC)
	dedupeKey := "tj:dedupe:user1:manualform"

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock\(hashtext\(\$1\)\)`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`insert into target_jobs`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000a2",
			"018f2a40-0000-7000-9000-0000000000b1",
			"draft", "ready",
			"PM", "Acme",
			nil, nil, nil,
			"en", "manual_form",
			nil, nil, "manual jd",
			[]byte(`{}`), []byte(`{}`),
			int32(0), now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into target_job_requirements`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000d1",
			"018f2a40-0000-7000-9000-0000000000a2",
			"must_have", "draft must-have", nil, "explicit", int32(1), now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(
			"018f2a40-0000-7000-9000-0000000000f2",
			"target_import", "target_job",
			"018f2a40-0000-7000-9000-0000000000a2",
			dedupeKey, "succeeded",
			[]byte(`{}`),
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := store.ImportTargetJob(context.Background(), targetjob.ImportTargetJobInput{
		UserID:                 "018f2a40-0000-7000-9000-0000000000b1",
		DedupeKey:              dedupeKey,
		TargetJobID:            "018f2a40-0000-7000-9000-0000000000a2",
		Title:                  "PM",
		CompanyName:            "Acme",
		TargetLanguage:         "en",
		APISourceType:          targetjob.SourceTypeManualForm,
		RawJDText:              "manual jd",
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		InitialAnalysisStatus:  sharedtypes.TargetJobParseStatusReady,
		JobID:                  "018f2a40-0000-7000-9000-0000000000f2",
		DraftRequirements: []targetjob.RequirementRecord{
			{ID: "018f2a40-0000-7000-9000-0000000000d1", Kind: targetjob.RequirementMustHave, Label: "draft must-have", DisplayOrder: 1},
		},
		Now: now,
	})
	if err != nil {
		t.Fatalf("ImportTargetJob: %v", err)
	}
	if res.JobStatus != sharedtypes.JobStatusSucceeded {
		t.Fatalf("manual_form must yield succeeded job, got %q", res.JobStatus)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ImportTargetJob_DedupeReturnsExistingActiveRunnerJob(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 18, 0, 0, 0, time.UTC)
	dedupeKey := "tj:dedupe:user1:keydup"

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock\(hashtext\(\$1\)\)`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "resource_id", "status", "created_at", "updated_at",
		}).AddRow(
			"018f2a40-0000-7000-9000-0000000000f9",
			"018f2a40-0000-7000-9000-0000000000a9",
			"queued", now, now,
		))
	mock.ExpectCommit()

	res, err := store.ImportTargetJob(context.Background(), targetjob.ImportTargetJobInput{
		UserID:                 "018f2a40-0000-7000-9000-0000000000b1",
		DedupeKey:              dedupeKey,
		TargetJobID:            "ignored-because-dedupe",
		TargetLanguage:         "en",
		APISourceType:          targetjob.SourceTypeManualText,
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		InitialAnalysisStatus:  sharedtypes.TargetJobParseStatusQueued,
		JobID:                  "ignored",
		OutboxEventID:          "ignored",
		Now:                    now,
	})
	if err != nil {
		t.Fatalf("dedupe: %v", err)
	}
	if !res.Existing {
		t.Fatal("expected Existing=true on dedupe hit")
	}
	if res.TargetJobID != "018f2a40-0000-7000-9000-0000000000a9" {
		t.Fatalf("expected dedupe to surface existing target id, got %q", res.TargetJobID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_ImportTargetJob_DedupeLockClosesManualFormRaceWindow(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	now := time.Date(2026, 5, 9, 18, 15, 0, 0, time.UTC)
	dedupeKey := "tj:dedupe:user1:manualform-race"

	mock.ExpectBegin()
	mock.ExpectExec(`select pg_advisory_xact_lock\(hashtext\(\$1\)\)`).
		WithArgs(dedupeKey).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`from async_jobs\s+where job_type = \$1 and dedupe_key = \$2`).
		WithArgs("target_import", dedupeKey).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "resource_id", "status", "created_at", "updated_at",
		}).AddRow(
			"018f2a40-0000-7000-9000-0000000000fa",
			"018f2a40-0000-7000-9000-0000000000aa",
			"succeeded", now, now,
		))
	mock.ExpectCommit()

	res, err := store.ImportTargetJob(context.Background(), targetjob.ImportTargetJobInput{
		UserID:                 "018f2a40-0000-7000-9000-0000000000b1",
		DedupeKey:              dedupeKey,
		TargetJobID:            "ignored-because-dedupe",
		TargetLanguage:         "en",
		APISourceType:          targetjob.SourceTypeManualForm,
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		InitialAnalysisStatus:  sharedtypes.TargetJobParseStatusReady,
		JobID:                  "ignored",
		DraftRequirements: []targetjob.RequirementRecord{
			{ID: "ignored-req", Kind: targetjob.RequirementMustHave, Label: "ignored", DisplayOrder: 1},
		},
		Now: now,
	})
	if err != nil {
		t.Fatalf("manual_form dedupe hit: %v", err)
	}
	if !res.Existing || res.TargetJobID != "018f2a40-0000-7000-9000-0000000000aa" || res.JobStatus != sharedtypes.JobStatusSucceeded {
		t.Fatalf("manual_form dedupe must return existing succeeded marker, got %+v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_LookupFileAttachmentForUser_HappyPath(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(`from file_objects\s+where id = \$1 and user_id = \$2 and deleted_at is null`).
		WithArgs("018f2a40-0000-7000-9000-0000000000ff", "018f2a40-0000-7000-9000-0000000000b1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "purpose"}).AddRow(
			"018f2a40-0000-7000-9000-0000000000ff",
			"018f2a40-0000-7000-9000-0000000000b1",
			"target_job_attachment",
		))
	rec, err := store.LookupFileAttachmentForUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b1",
		"018f2a40-0000-7000-9000-0000000000ff")
	if err != nil {
		t.Fatalf("LookupFileAttachmentForUser: %v", err)
	}
	if rec.Purpose != "target_job_attachment" {
		t.Fatalf("purpose not propagated: %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStore_LookupFileAttachmentForUser_NotFound(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()

	mock.ExpectQuery(`from file_objects`).
		WithArgs("018f2a40-0000-7000-9000-0000000000ff", "018f2a40-0000-7000-9000-0000000000b9").
		WillReturnError(sql.ErrNoRows)
	_, err := store.LookupFileAttachmentForUser(context.Background(),
		"018f2a40-0000-7000-9000-0000000000b9",
		"018f2a40-0000-7000-9000-0000000000ff")
	if !errors.Is(err, targetjob.ErrTargetJobNotFound) {
		t.Fatalf("expected ErrTargetJobNotFound, got %v", err)
	}
}

func TestSQLStore_ImportTargetJob_RequiresMandatoryIDs(t *testing.T) {
	store, _, cleanup := newMockStore(t)
	defer cleanup()
	if _, err := store.ImportTargetJob(context.Background(), targetjob.ImportTargetJobInput{}); err == nil {
		t.Fatal("empty input must be rejected")
	}
}

func TestRetryPolicy_BackoffBelowMax(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()
	now := time.Date(2026, 5, 16, 18, 0, 0, 0, time.UTC)

	mock.ExpectExec(`update async_jobs\s+set status = case when attempts >= max_attempts then 'dead' else 'queued' end`).
		WithArgs(now.Add(15*time.Second), "AI_PROVIDER_TIMEOUT", sqlmock.AnyArg(), "01918fa0-0000-7000-8000-00000000d011").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.FinalizeAsyncJob(context.Background(), "01918fa0-0000-7000-8000-00000000d011", targetjob.JobOutcome{
		ErrorCode:    "AI_PROVIDER_TIMEOUT",
		ErrorMessage: "provider timeout",
		Retryable:    true,
	}, now)
	if err != nil {
		t.Fatalf("FinalizeAsyncJob retryable: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRetryPolicy_PermanentFailAtMax(t *testing.T) {
	store, mock, cleanup := newMockStore(t)
	defer cleanup()
	now := time.Date(2026, 5, 16, 18, 0, 0, 0, time.UTC)

	mock.ExpectExec(`update async_jobs\s+set status = 'failed'`).
		WithArgs(now, "AI_PROVIDER_TIMEOUT", sqlmock.AnyArg(), "01918fa0-0000-7000-8000-00000000d011").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := store.FinalizeAsyncJob(context.Background(), "01918fa0-0000-7000-8000-00000000d011", targetjob.JobOutcome{
		ErrorCode:    "AI_PROVIDER_TIMEOUT",
		ErrorMessage: "provider timeout",
		Retryable:    false,
	}, now)
	if err != nil {
		t.Fatalf("FinalizeAsyncJob permanent failure: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

// TestStoreSurfaceRequiresUserScopeOnReadsAndWrites guards against accidental
// "GetByID" / "FindByID" methods that don't take user_id, which would break
// spec D-9 cross-user isolation.
func TestStoreSurfaceRequiresUserScopeOnReadsAndWrites(t *testing.T) {
	want := []string{"GetTargetJobByUser", "ListTargetJobsForUser", "UpdateTargetJobLifecycle"}
	storeType := reflectStoreType()
	for _, name := range want {
		if !storeHasMethod(storeType, name) {
			t.Fatalf("Store interface missing required user-scoped method %q", name)
		}
	}
	for i := 0; i < storeType.NumMethod(); i++ {
		method := storeType.Method(i).Name
		if (strings.Contains(method, "ByID") || strings.Contains(method, "FindByID")) &&
			!strings.Contains(method, "ByUser") {
			t.Fatalf("Store method %q must be user-scoped (use ByUser suffix)", method)
		}
	}
}
