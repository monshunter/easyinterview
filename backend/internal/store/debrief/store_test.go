package debrief

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/debrief"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestRepositoryPackageCompiles(t *testing.T) {
	t.Helper()
	if NewRepository(nil) == nil {
		t.Fatalf("NewRepository returned nil")
	}
}

func TestStoreCreateDebrief_HappyTransaction(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC)
	in := validCreateDebriefStoreInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`insert into debriefs`).
		WithArgs(
			in.DebriefID,
			in.UserID,
			in.TargetJobID,
			string(in.RoundType),
			string(in.InterviewerRole),
			in.Language,
			rawQuestionsArg{t: t, wantCount: len(in.Questions)},
			in.Notes,
			in.Now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.DebriefID))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(
			in.JobID,
			string(sharedjobs.JobTypeDebriefGenerate),
			in.DebriefID,
			in.DebriefID,
			string(sharedtypes.JobStatusQueued),
			debriefJobPayloadArg{t: t, in: in},
			in.Now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			in.OutboxEventID,
			string(sharedevents.EventNameDebriefCreated),
			in.DebriefID,
			debriefCreatedPayloadArg{t: t, in: in},
			in.Now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.CreateDebrief(context.Background(), in)
	if err != nil {
		t.Fatalf("CreateDebrief returned error: %v", err)
	}
	if result.DebriefID != in.DebriefID || result.Job.ID != in.JobID || result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreCreateDebrief_RollbackOnOutboxFailure(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC)
	in := validCreateDebriefStoreInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`insert into debriefs`).
		WithArgs(
			in.DebriefID,
			in.UserID,
			in.TargetJobID,
			string(in.RoundType),
			string(in.InterviewerRole),
			in.Language,
			rawQuestionsArg{t: t, wantCount: len(in.Questions)},
			in.Notes,
			in.Now,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(in.DebriefID))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(
			in.JobID,
			string(sharedjobs.JobTypeDebriefGenerate),
			in.DebriefID,
			in.DebriefID,
			string(sharedtypes.JobStatusQueued),
			debriefJobPayloadArg{t: t, in: in},
			in.Now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			in.OutboxEventID,
			string(sharedevents.EventNameDebriefCreated),
			in.DebriefID,
			debriefCreatedPayloadArg{t: t, in: in},
			in.Now,
		).
		WillReturnError(errors.New("outbox unavailable"))
	mock.ExpectRollback()

	_, err := repo.CreateDebrief(context.Background(), in)
	if err == nil {
		t.Fatalf("expected CreateDebrief to fail when outbox insert fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRepositoryRecordDebriefAuditEvent_WritesAllowedMetadata(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC)
	event := domain.DebriefAuditEvent{
		AuditEventID: "01918fa0-0000-7000-8000-00000000d013",
		UserID:       "01918fa0-0000-7000-8000-000000000001",
		Action:       domain.AuditActionCreateDebrief,
		ResourceType: "debrief",
		ResourceID:   "01918fa0-0000-7000-8000-00000000d010",
		Result:       domain.AuditResultSuccess,
		Metadata: map[string]any{
			"debrief_id":     "01918fa0-0000-7000-8000-00000000d010",
			"target_job_id":  "01918fa0-0000-7000-8000-00000000c001",
			"language":       "zh-CN",
			"question_count": 1,
			"status":         string(sharedtypes.DebriefStatusDraft),
		},
		CreatedAt: now,
	}

	mock.ExpectExec(`insert into audit_events`).
		WithArgs(
			event.AuditEventID,
			event.UserID,
			event.UserID,
			event.Action,
			event.ResourceType,
			event.ResourceID,
			event.Result,
			auditMetadataArg{t: t, expectedKeys: []string{"debrief_id", "target_job_id", "language", "question_count", "status"}},
			event.CreatedAt,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.RecordDebriefAuditEvent(context.Background(), event); err != nil {
		t.Fatalf("RecordDebriefAuditEvent returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateDebrief_OutboxPayloadSchema(t *testing.T) {
	in := validCreateDebriefStoreInput(time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC))
	raw, err := json.Marshal(buildDebriefCreatedPayload(in))
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	for _, forbidden := range []string{"questionText", "myAnswerSummary", "interviewerReaction", in.Notes, in.Questions[0].QuestionText, in.Questions[0].MyAnswerSummary, in.Questions[0].InterviewerReaction} {
		if forbidden != "" && strings.Contains(string(raw), forbidden) {
			t.Fatalf("debrief.created payload leaked %q: %s", forbidden, string(raw))
		}
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode payload: %v raw=%s", err, string(raw))
	}
	wantKeys := []string{"debriefId", "targetJobId", "roundType", "questionCount"}
	if len(got) != len(wantKeys) {
		t.Fatalf("payload keys drifted: got=%v want=%v", got, wantKeys)
	}
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Fatalf("payload missing %s: %v", key, got)
		}
	}
}

func TestOutboxPayload_NoRawText(t *testing.T) {
	now := time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC)
	createInput := validCreateDebriefStoreInput(now)
	createInput.Notes = "__SECRET_RAW_TEXT__ notes"
	createInput.Questions[0].QuestionText = "__SECRET_RAW_TEXT__ question"
	createInput.Questions[0].MyAnswerSummary = "__SECRET_RAW_TEXT__ answer"
	createInput.Questions[0].InterviewerReaction = "__SECRET_RAW_TEXT__ reaction"
	createdPayload, err := json.Marshal(buildDebriefCreatedPayload(createInput))
	if err != nil {
		t.Fatalf("marshal created payload: %v", err)
	}
	completedInput := validUpdateDebriefCompletedInput(now)
	completedInput.Questions[0].QuestionText = "__SECRET_RAW_TEXT__ completed question"
	completedInput.Questions[0].MyAnswerSummary = "__SECRET_RAW_TEXT__ completed answer"
	completedInput.Questions[0].InterviewerReaction = "__SECRET_RAW_TEXT__ completed reaction"
	completedInput.RiskItems[0].Label = "__SECRET_RAW_TEXT__ risk"
	completedPayload, err := json.Marshal(buildDebriefCompletedPayload(completedInput, createInput.TargetJobID))
	if err != nil {
		t.Fatalf("marshal completed payload: %v", err)
	}
	for label, raw := range map[string]string{
		"debrief.created":   string(createdPayload),
		"debrief.completed": string(completedPayload),
	} {
		for _, forbidden := range []string{"__SECRET_RAW_TEXT__", "questionText", "myAnswerSummary", "interviewerReaction", "notes", "risk_items"} {
			if strings.Contains(raw, forbidden) {
				t.Fatalf("%s payload leaked %q: %s", label, forbidden, raw)
			}
		}
	}
}

func TestStoreLoadGenerateContext_Happy(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	payload := domain.GenerateJobPayload{
		DebriefID:     "01918fa0-0000-7000-8000-00000000d010",
		TargetJobID:   "01918fa0-0000-7000-8000-00000000c001",
		Language:      "zh-CN",
		QuestionCount: 1,
	}
	rawQuestions, err := json.Marshal(rawQuestionPayloads([]domain.QuestionInput{{
		QuestionText:        "Tell me about a migration.",
		MyAnswerSummary:     "I led the rollout.",
		InterviewerReaction: "Asked for metrics.",
	}}))
	if err != nil {
		t.Fatalf("marshal raw questions: %v", err)
	}

	mock.ExpectQuery(`(?s)select d\.user_id::text.*from debriefs d.*join target_jobs tj`).
		WithArgs(payload.DebriefID, payload.TargetJobID).
		WillReturnRows(sqlmock.NewRows([]string{
			"user_id",
			"debrief_id",
			"target_job_id",
			"language",
			"target_title",
			"target_summary",
			"raw_questions",
		}).AddRow(
			"01918fa0-0000-7000-8000-000000000001",
			payload.DebriefID,
			payload.TargetJobID,
			"zh-CN",
			"Backend Engineer",
			[]byte(`{"mustHave":["Go"]}`),
			rawQuestions,
		))

	got, err := repo.LoadGenerateContext(context.Background(), payload)
	if err != nil {
		t.Fatalf("LoadGenerateContext returned error: %v", err)
	}
	if got.UserID != "01918fa0-0000-7000-8000-000000000001" ||
		got.DebriefID != payload.DebriefID ||
		got.TargetJobID != payload.TargetJobID ||
		got.TargetTitle != "Backend Engineer" ||
		got.TargetSummary != `{"mustHave":["Go"]}` ||
		len(got.Questions) != 1 ||
		got.Questions[0].QuestionText != "Tell me about a migration." {
		t.Fatalf("unexpected generate context: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreLoadGenerateContext_NotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	payload := domain.GenerateJobPayload{
		DebriefID:   "01918fa0-0000-7000-8000-00000000d010",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
	}

	mock.ExpectQuery(`(?s)select d\.user_id::text.*from debriefs d.*join target_jobs tj`).
		WithArgs(payload.DebriefID, payload.TargetJobID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.LoadGenerateContext(context.Background(), payload)
	if !errors.Is(err, domain.ErrDebriefNotFound) {
		t.Fatalf("error=%v, want ErrDebriefNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGenerateHandler_OutboxPayloadSchema(t *testing.T) {
	in := validUpdateDebriefCompletedInput(time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC))
	targetJobID := "01918fa0-0000-7000-8000-00000000c001"
	raw, err := json.Marshal(buildDebriefCompletedPayload(in, targetJobID))
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	for _, forbidden := range []string{"questionText", "myAnswerSummary", "interviewerReaction", "risk_items", "Tell me about the migration", "Metrics missing"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("debrief.completed payload leaked %q: %s", forbidden, string(raw))
		}
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode payload: %v raw=%s", err, string(raw))
	}
	wantKeys := []string{"debriefId", "targetJobId", "riskItemCount", "practiceFocusCount"}
	if len(got) != len(wantKeys) {
		t.Fatalf("payload keys drifted: got=%v want=%v", got, wantKeys)
	}
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Fatalf("payload missing %s: %v", key, got)
		}
	}
	if got["riskItemCount"] != float64(len(in.RiskItems)) || got["practiceFocusCount"] != float64(len(in.RiskItems)) {
		t.Fatalf("payload counts drifted: %v", got)
	}
}

func TestStoreGetDebrief_DraftPartial(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	userID := "01918fa0-0000-7000-8000-000000000001"
	debriefID := "01918fa0-0000-7000-8000-00000000d010"
	createdAt := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	mock.ExpectQuery(`(?s)select d\.id::text.*from debriefs d.*where d\.id = \$1.*d\.user_id = \$2`).
		WithArgs(debriefID, userID).
		WillReturnRows(sqlmock.NewRows(debriefRecordColumns()).AddRow(
			debriefID,
			"01918fa0-0000-7000-8000-00000000c001",
			string(sharedtypes.DebriefStatusDraft),
			string(sharedtypes.DebriefRoundTypeBehavioral),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			[]byte(`[{"questionText":"Tell me about scope.","myAnswerSummary":"I explained ownership.","interviewerReaction":"Asked for metrics."}]`),
			[]byte(`[{"label":"must not leak in draft","severity":"high"}]`),
			[]byte(`[{"label":"must not leak in draft"}]`),
			"draft thank-you should not leak",
			"",
			"",
			"",
			createdAt,
			updatedAt,
		))

	got, err := repo.GetDebrief(context.Background(), userID, debriefID)
	if err != nil {
		t.Fatalf("GetDebrief returned error: %v", err)
	}
	if got.ID != debriefID || got.Status != sharedtypes.DebriefStatusDraft || got.RoundType != sharedtypes.DebriefRoundTypeBehavioral {
		t.Fatalf("draft record drifted: %+v", got)
	}
	if len(got.Questions) != 1 || got.Questions[0].AIAnalysis != "" {
		t.Fatalf("draft questions should be partial without aiAnalysis: %+v", got.Questions)
	}
	if len(got.RiskItems) != 0 || len(got.NextRoundChecklist) != 0 || got.ThankYouDraft != "" || got.Provenance != nil {
		t.Fatalf("draft response leaked completed fields: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetDebrief_CompletedFull(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	userID := "01918fa0-0000-7000-8000-000000000001"
	debriefID := "01918fa0-0000-7000-8000-00000000d010"
	createdAt := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	mock.ExpectQuery(`(?s)select d\.id::text.*from debriefs d.*where d\.id = \$1.*d\.user_id = \$2`).
		WithArgs(debriefID, userID).
		WillReturnRows(sqlmock.NewRows(debriefRecordColumns()).AddRow(
			debriefID,
			"01918fa0-0000-7000-8000-00000000c001",
			string(sharedtypes.DebriefStatusCompleted),
			string(sharedtypes.DebriefRoundTypeBehavioral),
			string(sharedtypes.InterviewerRoleHiringManager),
			"zh-CN",
			[]byte(`[{"questionText":"Tell me about scope.","myAnswerSummary":"I explained ownership.","interviewerReaction":"Asked for metrics.","aiAnalysis":"Add outcome numbers."}]`),
			[]byte(`[{"label":"Metrics missing","severity":"medium"}]`),
			[]byte(`[{"label":"Prepare launch metrics","rationale":"Interviewer asked follow-up."}]`),
			"Thanks for the conversation.",
			"v0.1.0",
			"v0.1.0",
			"stub-model",
			createdAt,
			updatedAt,
		))

	got, err := repo.GetDebrief(context.Background(), userID, debriefID)
	if err != nil {
		t.Fatalf("GetDebrief returned error: %v", err)
	}
	if got.Status != sharedtypes.DebriefStatusCompleted ||
		len(got.Questions) != 1 ||
		got.Questions[0].AIAnalysis != "Add outcome numbers." ||
		len(got.RiskItems) != 1 ||
		len(got.NextRoundChecklist) != 1 ||
		got.ThankYouDraft == "" ||
		got.Provenance == nil ||
		got.Provenance.PromptVersion != "v0.1.0" ||
		got.Provenance.ModelID != "stub-model" {
		t.Fatalf("completed record drifted: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetDebrief_CrossUserNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	userID := "01918fa0-0000-7000-8000-000000000001"
	debriefID := "01918fa0-0000-7000-8000-00000000d010"

	mock.ExpectQuery(`(?s)select d\.id::text.*from debriefs d.*where d\.id = \$1.*d\.user_id = \$2`).
		WithArgs(debriefID, userID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetDebrief(context.Background(), userID, debriefID)
	if !errors.Is(err, domain.ErrDebriefNotFound) {
		t.Fatalf("error=%v, want ErrDebriefNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetSuggestionContext_TargetJobScoped(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	in := domain.SuggestionContextRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
	}

	mock.ExpectQuery(`(?s)select tj\.id::text.*from target_jobs tj.*where tj\.id = \$1.*tj\.user_id = \$2`).
		WithArgs(in.TargetJobID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id", "title", "company_name", "summary"}).AddRow(
			in.TargetJobID,
			"Backend Engineer",
			"Example Co",
			[]byte(`{"mustHave":["Go"]}`),
		))

	got, err := repo.GetSuggestionContext(context.Background(), in)
	if err != nil {
		t.Fatalf("GetSuggestionContext returned error: %v", err)
	}
	if got.TargetJobID != in.TargetJobID || got.Title != "Backend Engineer" || got.CompanyName != "Example Co" || got.Summary != `{"mustHave":["Go"]}` {
		t.Fatalf("context drifted: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetSuggestionContext_LoadsResumeStructuredProfile(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	in := domain.SuggestionContextRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		ResumeID:    "01918fa0-0000-7000-8000-00000000a001",
	}

	mock.ExpectQuery(`(?s)select tj\.id::text.*from target_jobs tj.*where tj\.id = \$1.*tj\.user_id = \$2`).
		WithArgs(in.TargetJobID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id", "title", "company_name", "summary"}).AddRow(
			in.TargetJobID,
			"Backend Engineer",
			"Example Co",
			[]byte(`{"mustHave":["Go"]}`),
		))
	mock.ExpectQuery(`(?s)select coalesce\(structured_profile, '\{\}'::jsonb\).*from resumes.*where id = \$1.*user_id = \$2.*deleted_at is null`).
		WithArgs(in.ResumeID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"structured_profile"}).AddRow([]byte(`{"basics":{"headline":"Platform engineer"},"skills":["Go","PostgreSQL"]}`)))

	got, err := repo.GetSuggestionContext(context.Background(), in)
	if err != nil {
		t.Fatalf("GetSuggestionContext returned error: %v", err)
	}
	if got.ResumeSummary != `{"basics":{"headline":"Platform engineer"},"skills":["Go","PostgreSQL"]}` {
		t.Fatalf("resume structured profile not loaded into context: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetSuggestionContext_LoadsPracticeSessionSummary(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	in := domain.SuggestionContextRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		SessionID:   "01918fa0-0000-7000-8000-000000005000",
	}

	mock.ExpectQuery(`(?s)select tj\.id::text.*from target_jobs tj.*where tj\.id = \$1.*tj\.user_id = \$2`).
		WithArgs(in.TargetJobID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id", "title", "company_name", "summary"}).AddRow(
			in.TargetJobID,
			"Backend Engineer",
			"Example Co",
			[]byte(`{"mustHave":["Go"]}`),
		))
	mock.ExpectQuery(`(?s)select jsonb_build_object.*from practice_sessions ps.*where ps\.id = \$1.*ps\.user_id = \$2.*ps\.target_job_id = \$3.*ps\.status = 'completed'`).
		WithArgs(in.SessionID, in.UserID, in.TargetJobID).
		WillReturnRows(sqlmock.NewRows([]string{"session_summary"}).AddRow([]byte(`{"sessionId":"01918fa0-0000-7000-8000-000000005000","status":"completed","turns":[{"questionText":"How did you measure adoption?","answerSummary":"I cited rollout metrics."}],"report":{"preparednessLevel":"basically_ready","issues":[{"title":"Needs sharper metrics"}]}}`)))

	got, err := repo.GetSuggestionContext(context.Background(), in)
	if err != nil {
		t.Fatalf("GetSuggestionContext returned error: %v", err)
	}
	if got.SessionSummary == "" || !strings.Contains(got.SessionSummary, `"sessionId":"01918fa0-0000-7000-8000-000000005000"`) || !strings.Contains(got.SessionSummary, `"Needs sharper metrics"`) {
		t.Fatalf("practice session summary not loaded into context: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetSuggestionContext_CrossUserSessionNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	in := domain.SuggestionContextRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		SessionID:   "01918fa0-0000-7000-8000-000000005000",
	}

	mock.ExpectQuery(`(?s)select tj\.id::text.*from target_jobs tj.*where tj\.id = \$1.*tj\.user_id = \$2`).
		WithArgs(in.TargetJobID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id", "title", "company_name", "summary"}).AddRow(
			in.TargetJobID,
			"Backend Engineer",
			"Example Co",
			[]byte(`{"mustHave":["Go"]}`),
		))
	mock.ExpectQuery(`(?s)select jsonb_build_object.*from practice_sessions ps.*where ps\.id = \$1.*ps\.user_id = \$2.*ps\.target_job_id = \$3.*ps\.status = 'completed'`).
		WithArgs(in.SessionID, in.UserID, in.TargetJobID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetSuggestionContext(context.Background(), in)
	if !errors.Is(err, domain.ErrDebriefPrerequisite) {
		t.Fatalf("error=%v, want ErrDebriefPrerequisite", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreGetSuggestionContext_CrossUserResumeNotFound(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	in := domain.SuggestionContextRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		ResumeID:    "01918fa0-0000-7000-8000-00000000a001",
	}

	mock.ExpectQuery(`(?s)select tj\.id::text.*from target_jobs tj.*where tj\.id = \$1.*tj\.user_id = \$2`).
		WithArgs(in.TargetJobID, in.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id", "title", "company_name", "summary"}).AddRow(
			in.TargetJobID,
			"Backend Engineer",
			"Example Co",
			[]byte(`{"mustHave":["Go"]}`),
		))
	mock.ExpectQuery(`(?s)select coalesce\(structured_profile, '\{\}'::jsonb\).*from resumes.*where id = \$1.*user_id = \$2.*deleted_at is null`).
		WithArgs(in.ResumeID, in.UserID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetSuggestionContext(context.Background(), in)
	if !errors.Is(err, domain.ErrDebriefPrerequisite) {
		t.Fatalf("error=%v, want ErrDebriefPrerequisite", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreUpdateDebriefCompleted_HappyTransaction(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	in := validUpdateDebriefCompletedInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`update debriefs`).
		WithArgs(
			completedQuestionsArg{t: t, wantCount: len(in.Questions)},
			riskItemsArg{t: t, wantCount: len(in.RiskItems)},
			in.Provenance.PromptVersion,
			in.Provenance.RubricVersion,
			in.Provenance.ModelID,
			in.Now,
			in.DebriefID,
			in.UserID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id"}).AddRow("01918fa0-0000-7000-8000-00000000c001"))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			in.OutboxEventID,
			string(sharedevents.EventNameDebriefCompleted),
			in.DebriefID,
			debriefCompletedPayloadArg{t: t, debriefID: in.DebriefID, targetJobID: "01918fa0-0000-7000-8000-00000000c001", riskCount: len(in.RiskItems)},
			in.Now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := repo.UpdateDebriefCompleted(context.Background(), in)
	if err != nil {
		t.Fatalf("UpdateDebriefCompleted returned error: %v", err)
	}
	if got.ID != in.DebriefID || got.Status != sharedtypes.DebriefStatusCompleted || got.TargetJobID != "01918fa0-0000-7000-8000-00000000c001" {
		t.Fatalf("unexpected record: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreUpdateDebriefCompleted_CASRejectsCompleted(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	in := validUpdateDebriefCompletedInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`update debriefs`).
		WithArgs(
			completedQuestionsArg{t: t, wantCount: len(in.Questions)},
			riskItemsArg{t: t, wantCount: len(in.RiskItems)},
			in.Provenance.PromptVersion,
			in.Provenance.RubricVersion,
			in.Provenance.ModelID,
			in.Now,
			in.DebriefID,
			in.UserID,
		).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.UpdateDebriefCompleted(context.Background(), in)
	if !errors.Is(err, domain.ErrDebriefIllegalState) {
		t.Fatalf("error=%v, want ErrDebriefIllegalState", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestStoreUpdateDebriefCompleted_OutboxRollback(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewRepository(db)
	now := time.Date(2026, 5, 16, 15, 0, 0, 0, time.UTC)
	in := validUpdateDebriefCompletedInput(now)

	mock.ExpectBegin()
	mock.ExpectQuery(`update debriefs`).
		WithArgs(
			completedQuestionsArg{t: t, wantCount: len(in.Questions)},
			riskItemsArg{t: t, wantCount: len(in.RiskItems)},
			in.Provenance.PromptVersion,
			in.Provenance.RubricVersion,
			in.Provenance.ModelID,
			in.Now,
			in.DebriefID,
			in.UserID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"target_job_id"}).AddRow("01918fa0-0000-7000-8000-00000000c001"))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(
			in.OutboxEventID,
			string(sharedevents.EventNameDebriefCompleted),
			in.DebriefID,
			debriefCompletedPayloadArg{t: t, debriefID: in.DebriefID, targetJobID: "01918fa0-0000-7000-8000-00000000c001", riskCount: len(in.RiskItems)},
			in.Now,
		).
		WillReturnError(errors.New("outbox unavailable"))
	mock.ExpectRollback()

	_, err := repo.UpdateDebriefCompleted(context.Background(), in)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return db, mock, func() { _ = db.Close() }
}

func debriefRecordColumns() []string {
	return []string{
		"id",
		"target_job_id",
		"status",
		"round_type",
		"interviewer_role",
		"language",
		"raw_questions",
		"risk_items",
		"next_round_checklist",
		"thank_you_draft",
		"prompt_version",
		"rubric_version",
		"model_id",
		"created_at",
		"updated_at",
	}
}

func validUpdateDebriefCompletedInput(now time.Time) domain.UpdateDebriefCompletedInput {
	return domain.UpdateDebriefCompletedInput{
		UserID:        "01918fa0-0000-7000-8000-000000000001",
		DebriefID:     "01918fa0-0000-7000-8000-00000000d010",
		OutboxEventID: "01918fa0-0000-7000-8000-00000000d020",
		Questions: []domain.QuestionRecord{{
			QuestionText:        "Tell me about the migration.",
			MyAnswerSummary:     "I led the rollout.",
			InterviewerReaction: "Asked for metrics.",
			AIAnalysis:          "Add adoption metrics.",
		}},
		RiskItems: []domain.RiskItem{{Label: "Metrics missing", Severity: "medium"}},
		Provenance: domain.Provenance{
			PromptVersion:     "v0.1.0",
			RubricVersion:     "v0.1.0",
			ModelID:           "stub-model",
			Language:          "zh-CN",
			FeatureFlag:       "none",
			DataSourceVersion: "debrief/01918fa0-0000-7000-8000-00000000d010@v1",
		},
		Now: now,
	}
}

func validCreateDebriefStoreInput(now time.Time) domain.CreateDebriefStoreInput {
	return domain.CreateDebriefStoreInput{
		DebriefID:       "01918fa0-0000-7000-8000-00000000d010",
		JobID:           "01918fa0-0000-7000-8000-00000000d011",
		OutboxEventID:   "01918fa0-0000-7000-8000-00000000d012",
		UserID:          "01918fa0-0000-7000-8000-000000000001",
		TargetJobID:     "01918fa0-0000-7000-8000-00000000c001",
		RoundType:       sharedtypes.DebriefRoundTypeBehavioral,
		InterviewerRole: sharedtypes.InterviewerRoleHiringManager,
		Language:        "zh-CN",
		Notes:           "Use concise STAR structure next time.",
		Questions: []domain.QuestionInput{{
			QuestionText:        "Tell me about a cross-functional project.",
			MyAnswerSummary:     "I described a design-system migration.",
			InterviewerReaction: "The interviewer asked for metrics.",
		}},
		Now: now,
	}
}

type rawQuestionsArg struct {
	t         *testing.T
	wantCount int
}

func (a rawQuestionsArg) Match(value driver.Value) bool {
	raw, ok := driverValueBytes(a.t, "raw questions", value)
	if !ok {
		return false
	}
	var got []map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		a.t.Errorf("raw questions is not JSON array: %v; raw=%s", err, string(raw))
		return false
	}
	if len(got) != a.wantCount {
		a.t.Errorf("raw question count=%d want %d; got=%v", len(got), a.wantCount, got)
		return false
	}
	for _, key := range []string{"questionText", "myAnswerSummary", "interviewerReaction"} {
		if _, ok := got[0][key]; !ok {
			a.t.Errorf("raw question missing key %s: %v", key, got[0])
			return false
		}
	}
	return true
}

type debriefJobPayloadArg struct {
	t  *testing.T
	in domain.CreateDebriefStoreInput
}

func (a debriefJobPayloadArg) Match(value driver.Value) bool {
	got, ok := exactJSONMap(a.t, "async job payload", value, []string{"debriefId", "targetJobId", "language", "questionCount"})
	if !ok {
		return false
	}
	return assertPayloadFields(a.t, got, map[string]any{
		"debriefId":     a.in.DebriefID,
		"targetJobId":   a.in.TargetJobID,
		"language":      a.in.Language,
		"questionCount": float64(len(a.in.Questions)),
	})
}

type debriefCreatedPayloadArg struct {
	t  *testing.T
	in domain.CreateDebriefStoreInput
}

func (a debriefCreatedPayloadArg) Match(value driver.Value) bool {
	got, ok := exactJSONMap(a.t, "debrief.created payload", value, []string{"debriefId", "targetJobId", "roundType", "questionCount"})
	if !ok {
		return false
	}
	return assertPayloadFields(a.t, got, map[string]any{
		"debriefId":     a.in.DebriefID,
		"targetJobId":   a.in.TargetJobID,
		"roundType":     string(a.in.RoundType),
		"questionCount": float64(len(a.in.Questions)),
	})
}

type debriefCompletedPayloadArg struct {
	t           *testing.T
	debriefID   string
	targetJobID string
	riskCount   int
}

func (a debriefCompletedPayloadArg) Match(value driver.Value) bool {
	got, ok := exactJSONMap(a.t, "debrief.completed payload", value, []string{"debriefId", "targetJobId", "riskItemCount", "practiceFocusCount"})
	if !ok {
		return false
	}
	return assertPayloadFields(a.t, got, map[string]any{
		"debriefId":          a.debriefID,
		"targetJobId":        a.targetJobID,
		"riskItemCount":      float64(a.riskCount),
		"practiceFocusCount": float64(a.riskCount),
	})
}

type completedQuestionsArg struct {
	t         *testing.T
	wantCount int
}

func (a completedQuestionsArg) Match(value driver.Value) bool {
	raw, ok := driverValueBytes(a.t, "completed questions", value)
	if !ok {
		return false
	}
	var got []map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		a.t.Errorf("completed questions is not JSON array: %v; raw=%s", err, string(raw))
		return false
	}
	if len(got) != a.wantCount || got[0]["aiAnalysis"] == "" {
		a.t.Errorf("completed questions drifted: %v", got)
		return false
	}
	return true
}

type riskItemsArg struct {
	t         *testing.T
	wantCount int
}

func (a riskItemsArg) Match(value driver.Value) bool {
	raw, ok := driverValueBytes(a.t, "risk items", value)
	if !ok {
		return false
	}
	var got []map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		a.t.Errorf("risk items is not JSON array: %v; raw=%s", err, string(raw))
		return false
	}
	if len(got) != a.wantCount || got[0]["label"] == "" {
		a.t.Errorf("risk items drifted: %v", got)
		return false
	}
	return true
}

func exactJSONMap(t *testing.T, label string, value driver.Value, keys []string) (map[string]any, bool) {
	t.Helper()
	raw, ok := driverValueBytes(t, label, value)
	if !ok {
		return nil, false
	}
	for _, forbidden := range []string{"questionText", "myAnswerSummary", "interviewerReaction", "Use concise STAR"} {
		if strings.Contains(string(raw), forbidden) {
			t.Errorf("%s leaked raw text token %q: %s", label, forbidden, string(raw))
			return nil, false
		}
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Errorf("%s is not JSON object: %v; raw=%s", label, err, string(raw))
		return nil, false
	}
	if len(got) != len(keys) {
		t.Errorf("%s keys drifted: got=%v want=%v", label, got, keys)
		return nil, false
	}
	for _, key := range keys {
		if _, ok := got[key]; !ok {
			t.Errorf("%s missing key %s: %v", label, key, got)
			return nil, false
		}
	}
	return got, true
}

func assertPayloadFields(t *testing.T, got map[string]any, want map[string]any) bool {
	t.Helper()
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Errorf("payload[%s]=%v want %v; full=%v", key, got[key], wantValue, got)
			return false
		}
	}
	return true
}

type auditMetadataArg struct {
	t            *testing.T
	expectedKeys []string
}

func (a auditMetadataArg) Match(value driver.Value) bool {
	got, ok := exactJSONMap(a.t, "debrief audit metadata", value, a.expectedKeys)
	if !ok {
		return false
	}
	if got["question_count"] != float64(1) {
		a.t.Errorf("question_count=%v want 1; full=%v", got["question_count"], got)
		return false
	}
	return true
}

func driverValueBytes(t *testing.T, label string, value driver.Value) ([]byte, bool) {
	t.Helper()
	switch v := value.(type) {
	case []byte:
		return append([]byte{}, v...), true
	case string:
		return []byte(v), true
	default:
		t.Errorf("%s has unexpected type %T", label, value)
		return nil, false
	}
}
