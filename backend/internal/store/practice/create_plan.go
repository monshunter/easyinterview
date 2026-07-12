package practice

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository { return &SQLRepository{db: db} }

func (r *SQLRepository) CreatePlan(ctx context.Context, in domain.CreatePlanStoreInput) (domain.PlanRecord, error) {
	if r == nil || r.db == nil {
		return domain.PlanRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("begin create practice plan: %w", err)
	}
	defer tx.Rollback()

	var plan domain.PlanRecord
	var sourceReportID sql.NullString
	err = tx.QueryRowContext(ctx, `
insert into practice_plans (
  id, user_id, target_job_id, source_report_id, goal,
  interviewer_persona, difficulty, language, time_budget_minutes,
  resume_id, focus_competency_codes, status, created_at, updated_at
)
select $1, $2, tj.id, nullif($4, '')::uuid, $5, $6, $7, $8, $9,
       r.id, $11, 'ready', $12, $12
from target_jobs tj
join resumes r on r.id = $10 and r.user_id = $2 and r.deleted_at is null
left join feedback_reports fr
  on fr.id = nullif($4, '')::uuid
 and fr.user_id = $2
 and fr.target_job_id = tj.id
 and fr.status = 'ready'
where tj.id = $3 and tj.user_id = $2 and tj.deleted_at is null
  and (($5 = 'baseline' and nullif($4, '') is null)
    or ($5 in ('retry_current_round', 'next_round') and fr.id is not null))
returning id, target_job_id, source_report_id::text, goal,
          interviewer_persona, difficulty, language, time_budget_minutes,
          resume_id::text, status, created_at`,
		in.PlanID, in.UserID, in.TargetJobID, in.SourceReportID, string(in.Goal),
		string(in.InterviewerPersona), in.Difficulty, in.Language, in.TimeBudgetMinutes,
		in.ResumeID, pq.Array(in.FocusCompetencyCodes), in.Now,
	).Scan(
		&plan.ID, &plan.TargetJobID, &sourceReportID, &plan.Goal,
		&plan.InterviewerPersona, &plan.Difficulty, &plan.Language,
		&plan.TimeBudgetMinutes, &plan.ResumeID, &plan.Status, &plan.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("insert practice plan: %w", err)
	}
	plan.SourceReportID = sourceReportID.String

	metadata, err := json.Marshal(map[string]any{
		"plan_id": in.PlanID, "goal": string(in.Goal), "language": in.Language,
		"target_job_id": in.TargetJobID, "source_report_id": in.SourceReportID,
	})
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("marshal practice plan audit metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type,
  resource_id, result, metadata, created_at
) values ($1,$2,'user',$3,'practice.plan.create','practice_plan',$4,'success',$5,$6)`,
		in.AuditEventID, in.UserID, in.UserID, in.PlanID, metadata, in.Now,
	); err != nil {
		return domain.PlanRecord{}, fmt.Errorf("insert practice plan audit event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.PlanRecord{}, fmt.Errorf("commit create practice plan: %w", err)
	}
	return plan, nil
}

func (r *SQLRepository) GetPlan(ctx context.Context, userID, planID string) (domain.PlanRecord, error) {
	if r == nil || r.db == nil {
		return domain.PlanRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	var plan domain.PlanRecord
	var sourceReportID sql.NullString
	err := r.db.QueryRowContext(ctx, `
select id, target_job_id, source_report_id::text, goal,
       interviewer_persona, difficulty, language, time_budget_minutes,
       resume_id::text, status, created_at
from practice_plans where user_id = $1 and id = $2`, userID, planID).Scan(
		&plan.ID, &plan.TargetJobID, &sourceReportID, &plan.Goal,
		&plan.InterviewerPersona, &plan.Difficulty, &plan.Language,
		&plan.TimeBudgetMinutes, &plan.ResumeID, &plan.Status, &plan.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("select practice plan: %w", err)
	}
	plan.SourceReportID = sourceReportID.String
	return plan, nil
}

func (r *SQLRepository) GetSession(ctx context.Context, userID, sessionID string) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	return selectSessionForUser(ctx, r.db, userID, sessionID)
}

func (r *SQLRepository) ListSessions(ctx context.Context, in domain.ListSessionsInput) (domain.ListSessionsResult, error) {
	if r == nil || r.db == nil {
		return domain.ListSessionsResult{}, fmt.Errorf("practice SQL repository is not configured")
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = sharedtypes.DefaultPageSize
	}
	if pageSize > sharedtypes.MaxPageSize {
		pageSize = sharedtypes.MaxPageSize
	}
	args := []any{in.UserID}
	query := `select s.id, s.plan_id, s.target_job_id, s.status, s.language, s.created_at, s.updated_at
from practice_sessions s where s.user_id = $1`
	if targetJobID := strings.TrimSpace(in.TargetJobID); targetJobID != "" {
		args = append(args, targetJobID)
		query += fmt.Sprintf(" and s.target_job_id = $%d", len(args))
	}
	if in.Status != "" {
		args = append(args, string(in.Status))
		query += fmt.Sprintf(" and s.status = $%d", len(args))
	}
	if cursor := strings.TrimSpace(in.Cursor); cursor != "" {
		updatedAt, id, err := decodeSessionCursor(cursor)
		if err != nil {
			return domain.ListSessionsResult{}, domain.ErrInvalidCursor
		}
		args = append(args, updatedAt, id)
		query += fmt.Sprintf(" and (s.updated_at, s.id) < ($%d, $%d)", len(args)-1, len(args))
	}
	args = append(args, pageSize+1)
	query += fmt.Sprintf(" order by s.updated_at desc, s.id desc limit $%d", len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.ListSessionsResult{}, fmt.Errorf("list practice sessions: %w", err)
	}
	defer rows.Close()
	items := make([]domain.SessionRecord, 0, pageSize)
	for rows.Next() {
		var session domain.SessionRecord
		if err := rows.Scan(&session.ID, &session.PlanID, &session.TargetJobID, &session.Status,
			&session.Language, &session.CreatedAt, &session.UpdatedAt); err != nil {
			return domain.ListSessionsResult{}, fmt.Errorf("scan practice session: %w", err)
		}
		session.Messages = []domain.MessageRecord{}
		items = append(items, session)
	}
	if err := rows.Err(); err != nil {
		return domain.ListSessionsResult{}, fmt.Errorf("iterate practice sessions: %w", err)
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	nextCursor := ""
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeSessionCursor(last.UpdatedAt, last.ID)
	}
	return domain.ListSessionsResult{Items: items, NextCursor: nextCursor, HasMore: hasMore, PageSize: pageSize}, nil
}

func (r *SQLRepository) ReserveSessionStart(ctx context.Context, in domain.StartSessionReservationInput) (domain.SessionReservation, error) {
	if r == nil || r.db == nil {
		return domain.SessionReservation{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionReservation{}, fmt.Errorf("begin reserve practice session: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, startSessionAdvisoryLockKey(in)); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("lock start session idempotency reservation: %w", err)
	}

	recordID := in.IdempotencyRecordID
	existing, hit, err := selectStartSessionIdempotency(ctx, tx, in)
	if err != nil {
		return domain.SessionReservation{}, err
	}
	if hit {
		recordID = existing.id
		if existing.fingerprint != in.RequestFingerprint && in.Now.Before(existing.expiresAt) {
			return domain.SessionReservation{}, domain.ErrSessionConflict
		}
		if in.Now.Before(existing.expiresAt) {
			switch existing.status {
			case idempotency.StatusSucceeded:
				replay, err := sessionRecordFromStoredResponse(existing.responseBody)
				if err != nil {
					return domain.SessionReservation{}, err
				}
				if err := tx.Commit(); err != nil {
					return domain.SessionReservation{}, fmt.Errorf("commit start session replay: %w", err)
				}
				return domain.SessionReservation{ReplaySession: &replay}, nil
			case idempotency.StatusPending, idempotency.StatusFailedTerminal:
				return domain.SessionReservation{}, domain.ErrSessionConflict
			}
		}
		if err := resetStartSessionIdempotency(ctx, tx, recordID, in); err != nil {
			return domain.SessionReservation{}, err
		}
	} else if _, err := tx.ExecContext(ctx, `
insert into idempotency_records (
  id, user_id, domain, operation, idempotency_key_hash, request_fingerprint,
  status, expires_at, created_at, updated_at
) values ($1,$2,'practice','startPracticeSession',$3,$4,$5,$6,$7,$7)`,
		in.IdempotencyRecordID, in.UserID, in.IdempotencyKeyHash, in.RequestFingerprint,
		string(idempotency.StatusPending), in.ExpiresAt, in.Now,
	); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("insert start session idempotency reservation: %w", err)
	}

	var reservation domain.SessionReservation
	var topSkills string
	var focusCompetencies pq.StringArray
	err = tx.QueryRowContext(ctx, `
with selected_plan as (
  select p.id, p.target_job_id, p.goal, p.interviewer_persona, p.language,
         p.focus_competency_codes,
         coalesce(nullif(tj.title, ''), 'target role') role_title,
         coalesce(nullif(tj.seniority_level, ''), 'not specified') seniority,
         coalesce(nullif(array_to_string(array(
           select req.label from target_job_requirements req
           where req.target_job_id = p.target_job_id
           order by req.display_order asc, req.created_at asc limit 6
         ), ', '), ''), 'target job requirements') top_skills,
         r.structured_profile
  from practice_plans p
  join target_jobs tj on tj.id = p.target_job_id and tj.user_id = p.user_id and tj.deleted_at is null
  join resumes r on r.id = p.resume_id and r.user_id = p.user_id and r.deleted_at is null
  where p.id = $3 and p.user_id = $2 and p.status = 'ready'
), inserted as (
  insert into practice_sessions (id, user_id, plan_id, target_job_id, status, language, created_at, updated_at)
  select $1, $2, id, target_job_id, 'queued', language, $4, $4 from selected_plan
  returning id, plan_id, target_job_id, language, created_at, updated_at
)
select inserted.id, inserted.plan_id, inserted.target_job_id, selected_plan.goal,
       selected_plan.interviewer_persona, inserted.language, selected_plan.role_title,
       selected_plan.seniority, selected_plan.top_skills,
       coalesce(nullif(selected_plan.structured_profile::text, '{}'::text), ''),
       selected_plan.focus_competency_codes, inserted.created_at, inserted.updated_at
from inserted join selected_plan on selected_plan.id = inserted.plan_id`,
		in.SessionID, in.UserID, in.PlanID, in.Now,
	).Scan(
		&reservation.SessionID, &reservation.PlanID, &reservation.TargetJobID, &reservation.Goal,
		&reservation.InterviewerPersona, &reservation.Language, &reservation.RoleTitle,
		&reservation.Seniority, &topSkills, &reservation.ResumeProfile, &focusCompetencies,
		&reservation.CreatedAt, &reservation.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionReservation{}, domain.ErrPlanNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return domain.SessionReservation{}, domain.ErrSessionConflict
		}
		return domain.SessionReservation{}, fmt.Errorf("reserve practice session: %w", err)
	}
	reservation.TopSkills = splitCommaList(topSkills)
	reservation.FocusCompetencies = append([]string(nil), focusCompetencies...)
	reservation.IdempotencyRecordID = recordID
	reservation.UserID = in.UserID
	if err := tx.Commit(); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("commit reserve practice session: %w", err)
	}
	return reservation, nil
}

func (r *SQLRepository) CommitSessionStart(ctx context.Context, in domain.CommitSessionStartInput) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("begin commit practice session start: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
insert into practice_messages (id, session_id, seq_no, role, content, created_at)
values ($1,$2,1,'assistant',$3,$4)`, in.MessageID, in.SessionID, in.MessageText, in.StartedAt); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert opening practice message: %w", err)
	}
	eventPayload, _ := json.Marshal(map[string]any{"sessionId": in.SessionID})
	if _, err := tx.ExecContext(ctx, `
insert into practice_session_events (id, session_id, seq_no, event_type, payload, created_at)
values ($1,$2,1,'session_started',$3,$4)`, in.SessionEventID, in.SessionID, eventPayload, in.StartedAt); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session started fact: %w", err)
	}
	startedPayload, err := BuildPracticeSessionStartedPayload(PracticeSessionStartedInput{
		Goal: in.Goal, Language: in.Language, PlanID: in.PlanID, SessionID: in.SessionID, TargetJobID: in.TargetJobID,
	})
	if err != nil {
		return domain.SessionRecord{}, err
	}
	outboxPayload, _ := json.Marshal(startedPayload)
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at)
values ($1,$2,1,'practice_session',$3,$4,'pending',$5)`,
		in.OutboxEventID, string(sharedevents.EventNamePracticeSessionStarted), in.SessionID, outboxPayload, in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session started outbox event: %w", err)
	}
	auditMetadata, _ := json.Marshal(map[string]any{
		"plan_id": in.PlanID, "session_id": in.SessionID, "goal": string(in.Goal),
		"language": in.Language, "target_job_id": in.TargetJobID,
	})
	if _, err := tx.ExecContext(ctx, `
insert into audit_events (id, user_id, actor_type, actor_id, action, resource_type, resource_id, result, metadata, created_at)
values ($1,$2,'user',$3,'practice.session.start','practice_session',$4,'success',$5,$6)`,
		in.AuditEventID, in.UserID, in.UserID, in.SessionID, auditMetadata, in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session start audit event: %w", err)
	}
	var session domain.SessionRecord
	err = tx.QueryRowContext(ctx, `
update practice_sessions set status = $1, started_at = $2, updated_at = $2 where id = $3
returning id, plan_id, target_job_id, status, language, created_at, updated_at`,
		string(sharedtypes.SessionStatusRunning), in.StartedAt, in.SessionID,
	).Scan(&session.ID, &session.PlanID, &session.TargetJobID, &session.Status,
		&session.Language, &session.CreatedAt, &session.UpdatedAt)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionRecord{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("update practice session running: %w", err)
	}
	session.Messages = []domain.MessageRecord{{ID: in.MessageID, Role: "assistant", Content: in.MessageText, SeqNo: 1, CreatedAt: in.StartedAt}}
	responseBody, err := marshalSessionResponseBody(session)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	result, err := tx.ExecContext(ctx, `
update idempotency_records set status = $1, resource_type = 'practice_session', resource_id = $2,
response_body = $3, error_code = null, updated_at = $4 where id = $5`,
		string(idempotency.StatusSucceeded), in.SessionID, responseBody, in.StartedAt, in.IdempotencyRecordID,
	)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("mark start session idempotency succeeded: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return domain.SessionRecord{}, fmt.Errorf("start session idempotency reservation not found")
	}
	if err := tx.Commit(); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("commit practice session start: %w", err)
	}
	return session, nil
}

func (r *SQLRepository) FailSessionStart(ctx context.Context, in domain.FailSessionStartInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin fail practice session start: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `update practice_sessions set status=$1, failure_code=$2, updated_at=$3 where id=$4 and user_id=$5`,
		string(sharedtypes.SessionStatusFailed), in.ErrorCode, in.FailedAt, in.SessionID, in.UserID); err != nil {
		return fmt.Errorf("mark practice session failed: %w", err)
	}
	status := idempotency.StatusFailedTerminal
	if in.Retryable {
		status = idempotency.StatusFailedRetry
	}
	if _, err := tx.ExecContext(ctx, `
update idempotency_records set status=$1, error_code=$2, resource_type='practice_session',
resource_id=$3, response_body=null, updated_at=$4
where id=$5 and user_id=$6 and domain='practice' and operation='startPracticeSession'`,
		string(status), in.ErrorCode, in.SessionID, in.FailedAt, in.IdempotencyRecordID, in.UserID); err != nil {
		return fmt.Errorf("mark start session idempotency failed: %w", err)
	}
	return tx.Commit()
}

func selectSessionForUser(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, userID, sessionID string) (domain.SessionRecord, error) {
	var session domain.SessionRecord
	err := q.QueryRowContext(ctx, `
select id, plan_id, target_job_id, status, language, created_at, updated_at
from practice_sessions where user_id=$1 and id=$2`, userID, sessionID).Scan(
		&session.ID, &session.PlanID, &session.TargetJobID, &session.Status,
		&session.Language, &session.CreatedAt, &session.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionRecord{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("select practice session: %w", err)
	}
	rows, err := queryMessages(ctx, q, sessionID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	defer rows.Close()
	session.Messages, err = scanMessages(rows)
	return session, err
}

func queryMessages(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, sessionID string) (*sql.Rows, error) {
	rows, err := q.QueryContext(ctx, `
select id, role, content, seq_no, created_at from practice_messages
where session_id=$1 order by seq_no asc`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("select practice messages: %w", err)
	}
	return rows, nil
}

func scanMessages(rows *sql.Rows) ([]domain.MessageRecord, error) {
	messages := []domain.MessageRecord{}
	for rows.Next() {
		var message domain.MessageRecord
		if err := rows.Scan(&message.ID, &message.Role, &message.Content, &message.SeqNo, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan practice message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate practice messages: %w", err)
	}
	return messages, nil
}

type selectedStartSessionIdempotency struct {
	id, fingerprint string
	status          idempotency.Status
	responseBody    []byte
	expiresAt       time.Time
}

func selectStartSessionIdempotency(ctx context.Context, tx *sql.Tx, in domain.StartSessionReservationInput) (selectedStartSessionIdempotency, bool, error) {
	var rec selectedStartSessionIdempotency
	var status string
	var responseBody []byte
	err := tx.QueryRowContext(ctx, `
select id, request_fingerprint, status, response_body, expires_at
from idempotency_records where user_id=$1 and domain='practice'
and operation='startPracticeSession' and idempotency_key_hash=$2 for update`,
		in.UserID, in.IdempotencyKeyHash,
	).Scan(&rec.id, &rec.fingerprint, &status, &responseBody, &rec.expiresAt)
	if stderrs.Is(err, sql.ErrNoRows) {
		return rec, false, nil
	}
	if err != nil {
		return rec, false, fmt.Errorf("select start session idempotency reservation: %w", err)
	}
	rec.status = idempotency.Status(status)
	rec.responseBody = responseBody
	return rec, true, nil
}

func resetStartSessionIdempotency(ctx context.Context, tx *sql.Tx, recordID string, in domain.StartSessionReservationInput) error {
	_, err := tx.ExecContext(ctx, `
update idempotency_records set request_fingerprint=$1, status=$2, resource_type=null,
resource_id=null, response_body=null, error_code=null, expires_at=$3, updated_at=$4
where id=$5 and user_id=$6 and domain='practice' and operation='startPracticeSession'`,
		in.RequestFingerprint, string(idempotency.StatusPending), in.ExpiresAt, in.Now, recordID, in.UserID)
	if err != nil {
		return fmt.Errorf("reset start session idempotency reservation: %w", err)
	}
	return nil
}

func startSessionAdvisoryLockKey(in domain.StartSessionReservationInput) string {
	return strings.Join([]string{in.UserID, "practice", "startPracticeSession", in.IdempotencyKeyHash}, "\x1f")
}

type sessionCursorPayload struct {
	UpdatedAt string `json:"updatedAt"`
	ID        string `json:"id"`
}

func encodeSessionCursor(updatedAt time.Time, id string) string {
	raw, _ := json.Marshal(sessionCursorPayload{UpdatedAt: updatedAt.UTC().Format(time.RFC3339Nano), ID: strings.TrimSpace(id)})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeSessionCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cursor))
	if err != nil || len(raw) == 0 {
		return time.Time{}, "", domain.ErrInvalidCursor
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var payload sessionCursorPayload
	if err := dec.Decode(&payload); err != nil {
		return time.Time{}, "", domain.ErrInvalidCursor
	}
	if err := dec.Decode(&struct{}{}); !stderrs.Is(err, io.EOF) {
		return time.Time{}, "", domain.ErrInvalidCursor
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil || payload.UpdatedAt != updatedAt.UTC().Format(time.RFC3339Nano) || strings.TrimSpace(payload.ID) == "" {
		return time.Time{}, "", domain.ErrInvalidCursor
	}
	return updatedAt.UTC(), strings.TrimSpace(payload.ID), nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return stderrs.As(err, &pqErr) && string(pqErr.Code) == "23505"
}

func splitCommaList(value string) []string {
	out := []string{}
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func marshalSessionResponseBody(session domain.SessionRecord) ([]byte, error) {
	return json.Marshal(session)
}

func sessionRecordFromStoredResponse(raw []byte) (domain.SessionRecord, error) {
	var session domain.SessionRecord
	if len(raw) == 0 {
		return session, fmt.Errorf("stored session response is empty")
	}
	if err := json.Unmarshal(raw, &session); err != nil {
		return session, fmt.Errorf("unmarshal stored session response: %w", err)
	}
	if strings.TrimSpace(session.ID) == "" {
		return session, fmt.Errorf("stored session response missing id")
	}
	if session.Messages == nil {
		session.Messages = []domain.MessageRecord{}
	}
	return session, nil
}

var _ domain.Store = (*SQLRepository)(nil)
