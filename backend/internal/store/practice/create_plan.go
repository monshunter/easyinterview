package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrs "errors"
	"fmt"
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

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

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
	err = tx.QueryRowContext(ctx, `
insert into practice_plans (
  id, user_id, target_job_id, source_report_id, goal, mode,
  interviewer_persona, difficulty, language, time_budget_minutes,
  question_budget, resume_asset_id, focus_competency_codes, status,
  created_at, updated_at
)
select $1, $2, tj.id, null, $4, $5, $6, $7, $8, $9, $10,
       ra.id, $12, 'ready', $13, $13
from target_jobs tj
join resume_assets ra
  on ra.id = $11
 and ra.user_id = $2
 and ra.deleted_at is null
where tj.id = $3
  and tj.user_id = $2
  and tj.deleted_at is null
returning id, target_job_id, goal, mode, interviewer_persona, difficulty,
          language, time_budget_minutes, question_budget, status, created_at`,
		in.PlanID,
		in.UserID,
		in.TargetJobID,
		string(in.Goal),
		string(in.Mode),
		string(in.InterviewerPersona),
		in.Difficulty,
		in.Language,
		in.TimeBudgetMinutes,
		in.QuestionBudget,
		in.ResumeAssetID,
		pq.Array(in.FocusCompetencyCodes),
		in.Now,
	).Scan(
		&plan.ID,
		&plan.TargetJobID,
		&plan.Goal,
		&plan.Mode,
		&plan.InterviewerPersona,
		&plan.Difficulty,
		&plan.Language,
		&plan.TimeBudgetMinutes,
		&plan.QuestionBudget,
		&plan.Status,
		&plan.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("insert practice plan: %w", err)
	}

	metadata, err := json.Marshal(map[string]any{
		"plan_id":       in.PlanID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	})
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("marshal practice plan audit metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type,
  resource_id, result, metadata, created_at
) values ($1,$2,'user',$3,'practice.plan.create','practice_plan',$4,'success',$5,$6)`,
		in.AuditEventID,
		in.UserID,
		in.UserID,
		in.PlanID,
		metadata,
		in.Now,
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
	err := r.db.QueryRowContext(ctx, `
select id, target_job_id, goal, mode, interviewer_persona, difficulty,
       language, time_budget_minutes, question_budget, status, created_at
from practice_plans
where user_id = $1
  and id = $2`,
		userID,
		planID,
	).Scan(
		&plan.ID,
		&plan.TargetJobID,
		&plan.Goal,
		&plan.Mode,
		&plan.InterviewerPersona,
		&plan.Difficulty,
		&plan.Language,
		&plan.TimeBudgetMinutes,
		&plan.QuestionBudget,
		&plan.Status,
		&plan.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("select practice plan: %w", err)
	}
	return plan, nil
}

func (r *SQLRepository) GetSession(ctx context.Context, userID, sessionID string) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	return selectSessionForUser(ctx, r.db, userID, sessionID)
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

	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, strings.Join([]string{in.UserID, "practice", "startPracticeSession", in.IdempotencyKeyHash}, "\x00")); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("lock start session idempotency reservation: %w", err)
	}

	recordID := in.IdempotencyRecordID
	existing, hit, err := selectStartSessionIdempotency(ctx, tx, in)
	if err != nil {
		return domain.SessionReservation{}, err
	}
	if hit {
		if existing.fingerprint != in.RequestFingerprint {
			return domain.SessionReservation{}, domain.ErrSessionConflict
		}
		switch existing.status {
		case idempotency.StatusPending:
			return domain.SessionReservation{}, domain.ErrSessionConflict
		case idempotency.StatusSucceeded:
			if strings.TrimSpace(existing.resourceID) == "" {
				return domain.SessionReservation{}, domain.ErrSessionConflict
			}
			if len(existing.responseBody) == 0 {
				return domain.SessionReservation{}, fmt.Errorf("start session idempotency response body is empty")
			}
			session, err := sessionRecordFromStoredResponse(existing.responseBody)
			if err != nil {
				return domain.SessionReservation{}, err
			}
			if session.ID != existing.resourceID {
				return domain.SessionReservation{}, fmt.Errorf("start session idempotency response resource mismatch")
			}
			if err := tx.Commit(); err != nil {
				return domain.SessionReservation{}, fmt.Errorf("commit start session idempotency replay: %w", err)
			}
			return domain.SessionReservation{ReplaySession: &session}, nil
		case idempotency.StatusFailedRetry:
			recordID = existing.id
			if _, err := tx.ExecContext(ctx, `
update idempotency_records
set request_fingerprint = $1,
    status = $2,
    resource_type = null,
    resource_id = null,
    response_body = null,
    error_code = null,
    expires_at = $3,
    updated_at = $4
where id = $5
  and user_id = $6
  and domain = 'practice'
  and operation = 'startPracticeSession'`,
				in.RequestFingerprint,
				string(idempotency.StatusPending),
				in.ExpiresAt,
				in.Now,
				existing.id,
				in.UserID,
			); err != nil {
				return domain.SessionReservation{}, fmt.Errorf("reset retryable start session idempotency reservation: %w", err)
			}
		default:
			return domain.SessionReservation{}, domain.ErrSessionConflict
		}
	} else if _, err := tx.ExecContext(ctx, `
insert into idempotency_records (
  id, user_id, domain, operation, idempotency_key_hash,
  request_fingerprint, status, expires_at, created_at, updated_at
) values ($1,$2,'practice','startPracticeSession',$3,$4,$5,$6,$7,$7)`,
		recordID,
		in.UserID,
		in.IdempotencyKeyHash,
		in.RequestFingerprint,
		string(idempotency.StatusPending),
		in.ExpiresAt,
		in.Now,
	); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("insert start session idempotency reservation: %w", err)
	}

	var reservation domain.SessionReservation
	err = tx.QueryRowContext(ctx, `
with selected_plan as (
  select id, target_job_id, goal, mode, interviewer_persona, language
  from practice_plans
  where id = $3
    and user_id = $2
    and status = 'ready'
),
inserted as (
  insert into practice_sessions (
    id, user_id, plan_id, target_job_id, status, language,
    hints_enabled, turn_count, created_at, updated_at
  )
  select $1, $2, id, target_job_id, 'queued', language, $4, 0, $5, $5
  from selected_plan
  returning id, plan_id, target_job_id, language, hints_enabled, created_at, updated_at
)
select inserted.id, inserted.plan_id, inserted.target_job_id,
       selected_plan.goal, selected_plan.mode, selected_plan.interviewer_persona,
       inserted.language, inserted.hints_enabled, inserted.created_at, inserted.updated_at
from inserted
join selected_plan on selected_plan.id = inserted.plan_id`,
		in.SessionID,
		in.UserID,
		in.PlanID,
		in.HintsEnabled,
		in.Now,
	).Scan(
		&reservation.SessionID,
		&reservation.PlanID,
		&reservation.TargetJobID,
		&reservation.Goal,
		&reservation.Mode,
		&reservation.InterviewerPersona,
		&reservation.Language,
		&reservation.HintsEnabled,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
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
	if err := tx.Commit(); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("commit reserve practice session: %w", err)
	}
	reservation.IdempotencyRecordID = recordID
	reservation.UserID = in.UserID
	return reservation, nil
}

type selectedStartSessionIdempotency struct {
	id           string
	fingerprint  string
	status       idempotency.Status
	resourceID   string
	responseBody []byte
}

func selectStartSessionIdempotency(ctx context.Context, tx *sql.Tx, in domain.StartSessionReservationInput) (selectedStartSessionIdempotency, bool, error) {
	var rec selectedStartSessionIdempotency
	var status string
	var resourceID sql.NullString
	var responseBody sql.NullString
	err := tx.QueryRowContext(ctx, `
select id, request_fingerprint, status, resource_id::text, response_body
from idempotency_records
where user_id = $1
  and domain = 'practice'
  and operation = 'startPracticeSession'
  and idempotency_key_hash = $2
for update`,
		in.UserID,
		in.IdempotencyKeyHash,
	).Scan(&rec.id, &rec.fingerprint, &status, &resourceID, &responseBody)
	if stderrs.Is(err, sql.ErrNoRows) {
		return selectedStartSessionIdempotency{}, false, nil
	}
	if err != nil {
		return selectedStartSessionIdempotency{}, false, fmt.Errorf("select start session idempotency reservation: %w", err)
	}
	rec.status = idempotency.Status(status)
	rec.resourceID = resourceID.String
	if responseBody.Valid {
		rec.responseBody = []byte(responseBody.String)
	}
	return rec, true, nil
}

func selectSessionForUser(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userID, sessionID string) (domain.SessionRecord, error) {
	var session domain.SessionRecord
	var turnID sql.NullString
	var turnIndex sql.NullInt64
	var questionText sql.NullString
	var questionIntent sql.NullString
	var turnStatus sql.NullString
	var askedAt sql.NullTime
	err := q.QueryRowContext(ctx, `
select s.id, s.plan_id, s.target_job_id, s.status, s.language, s.hints_enabled,
       s.turn_count, s.created_at, s.updated_at,
       t.id, t.turn_index, t.question_text, t.question_intent, t.status, t.asked_at
from practice_sessions s
left join lateral (
  select id, turn_index, question_text, question_intent, status, asked_at
  from practice_turns
  where session_id = s.id
  order by turn_index desc
  limit 1
) t on true
where s.user_id = $1
  and s.id = $2`,
		userID,
		sessionID,
	).Scan(
		&session.ID,
		&session.PlanID,
		&session.TargetJobID,
		&session.Status,
		&session.Language,
		&session.HintsEnabled,
		&session.TurnCount,
		&session.CreatedAt,
		&session.UpdatedAt,
		&turnID,
		&turnIndex,
		&questionText,
		&questionIntent,
		&turnStatus,
		&askedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionRecord{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("select practice session: %w", err)
	}
	if turnID.Valid {
		session.CurrentTurn = &domain.TurnRecord{
			ID:             turnID.String,
			TurnIndex:      int32(turnIndex.Int64),
			QuestionText:   questionText.String,
			QuestionIntent: questionIntent.String,
			Status:         turnStatus.String,
			AskedAt:        askedAt.Time,
		}
	}
	return session, nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return stderrs.As(err, &pqErr) && string(pqErr.Code) == "23505"
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
insert into practice_turns (
  id, session_id, turn_index, question_text, question_intent,
  interviewer_persona, status, asked_at, created_at, updated_at
) values ($1,$2,1,$3,$4,$5,'asked',$6,$6,$6)`,
		in.TurnID,
		in.SessionID,
		in.QuestionText,
		nullableString(in.QuestionIntent),
		string(in.InterviewerPersona),
		in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert first practice turn: %w", err)
	}

	eventPayload, err := json.Marshal(map[string]any{
		"sessionId": in.SessionID,
		"turnId":    in.TurnID,
		"turnIndex": 1,
	})
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("marshal practice session event payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into practice_session_events (
  id, session_id, seq_no, event_type, client_event_id, payload, created_at
) values ($1,$2,1,'session_started',null,$3,$4)`,
		in.SessionEventID,
		in.SessionID,
		eventPayload,
		in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session event: %w", err)
	}

	startedPayload, err := BuildPracticeSessionStartedPayload(PracticeSessionStartedInput{
		Goal:        in.Goal,
		Language:    in.Language,
		Mode:        in.Mode,
		PlanID:      in.PlanID,
		SessionID:   in.SessionID,
		TargetJobID: in.TargetJobID,
	})
	if err != nil {
		return domain.SessionRecord{}, err
	}
	outboxPayload, err := json.Marshal(startedPayload)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("marshal practice session started payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,1,'practice_session',$3,$4,'pending',$5)`,
		in.OutboxEventID,
		string(sharedevents.EventNamePracticeSessionStarted),
		in.SessionID,
		outboxPayload,
		in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session started outbox event: %w", err)
	}

	auditMetadata, err := json.Marshal(map[string]any{
		"plan_id":       in.PlanID,
		"session_id":    in.SessionID,
		"goal":          string(in.Goal),
		"mode":          string(in.Mode),
		"language":      in.Language,
		"target_job_id": in.TargetJobID,
	})
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("marshal practice session start audit metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type,
  resource_id, result, metadata, created_at
) values ($1,$2,'user',$3,'practice.session.start','practice_session',$4,'success',$5,$6)`,
		in.AuditEventID,
		in.UserID,
		in.UserID,
		in.SessionID,
		auditMetadata,
		in.StartedAt,
	); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("insert practice session start audit event: %w", err)
	}

	var session domain.SessionRecord
	err = tx.QueryRowContext(ctx, `
update practice_sessions
set status = $1,
    turn_count = 1,
    started_at = $2,
    updated_at = $2
where id = $3
returning id, plan_id, target_job_id, status, language, hints_enabled, turn_count, created_at, updated_at`,
		string(sharedtypes.SessionStatusRunning),
		in.StartedAt,
		in.SessionID,
	).Scan(
		&session.ID,
		&session.PlanID,
		&session.TargetJobID,
		&session.Status,
		&session.Language,
		&session.HintsEnabled,
		&session.TurnCount,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SessionRecord{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("update practice session running: %w", err)
	}
	session.CurrentTurn = &domain.TurnRecord{
		ID:             in.TurnID,
		TurnIndex:      1,
		QuestionText:   in.QuestionText,
		QuestionIntent: in.QuestionIntent,
		Status:         "asked",
		AskedAt:        in.StartedAt,
	}
	responseBody, err := marshalSessionResponseBody(session)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	res, err := tx.ExecContext(ctx, `
update idempotency_records
set status = $1,
    resource_type = 'practice_session',
    resource_id = $2,
    response_body = $3,
    error_code = null,
    updated_at = $4
where id = $5`,
		string(idempotency.StatusSucceeded),
		in.SessionID,
		responseBody,
		in.StartedAt,
		in.IdempotencyRecordID,
	)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("mark start session idempotency succeeded: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("mark start session idempotency rows affected: %w", err)
	}
	if rows == 0 {
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

	res, err := tx.ExecContext(ctx, `
update practice_sessions
set status = $1,
    failure_code = $2,
    updated_at = $3
where id = $4
  and user_id = $5`,
		string(sharedtypes.SessionStatusFailed),
		in.ErrorCode,
		in.FailedAt,
		in.SessionID,
		in.UserID,
	)
	if err != nil {
		return fmt.Errorf("mark practice session failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark practice session failed rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	status := idempotency.StatusFailedTerminal
	if in.Retryable {
		status = idempotency.StatusFailedRetry
	}
	res, err = tx.ExecContext(ctx, `
update idempotency_records
set status = $1,
    error_code = $2,
    resource_type = 'practice_session',
    resource_id = $3,
    response_body = null,
    updated_at = $4
where id = $5
  and user_id = $6
  and domain = 'practice'
  and operation = 'startPracticeSession'`,
		string(status),
		in.ErrorCode,
		in.SessionID,
		in.FailedAt,
		in.IdempotencyRecordID,
		in.UserID,
	)
	if err != nil {
		return fmt.Errorf("mark start session idempotency failed: %w", err)
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark start session idempotency failed rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("start session idempotency reservation not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit fail practice session start: %w", err)
	}
	return nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func marshalSessionResponseBody(session domain.SessionRecord) ([]byte, error) {
	var currentTurn any
	if session.CurrentTurn != nil {
		currentTurn = map[string]any{
			"id":             session.CurrentTurn.ID,
			"turnIndex":      session.CurrentTurn.TurnIndex,
			"questionText":   session.CurrentTurn.QuestionText,
			"questionIntent": session.CurrentTurn.QuestionIntent,
			"status":         session.CurrentTurn.Status,
			"askedAt":        session.CurrentTurn.AskedAt.UTC().Format(time.RFC3339),
		}
	}
	raw, err := json.Marshal(map[string]any{
		"id":           session.ID,
		"planId":       session.PlanID,
		"targetJobId":  session.TargetJobID,
		"status":       string(session.Status),
		"language":     session.Language,
		"hintsEnabled": session.HintsEnabled,
		"turnCount":    session.TurnCount,
		"currentTurn":  currentTurn,
		"createdAt":    session.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt":    session.UpdatedAt.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal start session idempotency response: %w", err)
	}
	return raw, nil
}

func sessionRecordFromStoredResponse(raw []byte) (domain.SessionRecord, error) {
	var decoded struct {
		ID           string                    `json:"id"`
		PlanID       string                    `json:"planId"`
		TargetJobID  string                    `json:"targetJobId"`
		Status       sharedtypes.SessionStatus `json:"status"`
		Language     string                    `json:"language"`
		HintsEnabled bool                      `json:"hintsEnabled"`
		TurnCount    int32                     `json:"turnCount"`
		CurrentTurn  *struct {
			ID             string  `json:"id"`
			TurnIndex      int32   `json:"turnIndex"`
			QuestionText   string  `json:"questionText"`
			QuestionIntent *string `json:"questionIntent"`
			Status         string  `json:"status"`
			AskedAt        *string `json:"askedAt"`
		} `json:"currentTurn"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}
	if len(raw) == 0 {
		return domain.SessionRecord{}, fmt.Errorf("stored session response is empty")
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("unmarshal stored session response: %w", err)
	}
	if strings.TrimSpace(decoded.ID) == "" {
		return domain.SessionRecord{}, fmt.Errorf("stored session response missing id")
	}
	createdAt, err := time.Parse(time.RFC3339, decoded.CreatedAt)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("parse stored session createdAt: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, decoded.UpdatedAt)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("parse stored session updatedAt: %w", err)
	}
	session := domain.SessionRecord{
		ID:           decoded.ID,
		PlanID:       decoded.PlanID,
		TargetJobID:  decoded.TargetJobID,
		Status:       decoded.Status,
		Language:     decoded.Language,
		HintsEnabled: decoded.HintsEnabled,
		TurnCount:    decoded.TurnCount,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if decoded.CurrentTurn != nil {
		var askedAt time.Time
		if decoded.CurrentTurn.AskedAt != nil && strings.TrimSpace(*decoded.CurrentTurn.AskedAt) != "" {
			parsedAskedAt, err := time.Parse(time.RFC3339, *decoded.CurrentTurn.AskedAt)
			if err != nil {
				return domain.SessionRecord{}, fmt.Errorf("parse stored turn askedAt: %w", err)
			}
			askedAt = parsedAskedAt
		}
		intent := ""
		if decoded.CurrentTurn.QuestionIntent != nil {
			intent = *decoded.CurrentTurn.QuestionIntent
		}
		session.CurrentTurn = &domain.TurnRecord{
			ID:             decoded.CurrentTurn.ID,
			TurnIndex:      decoded.CurrentTurn.TurnIndex,
			QuestionText:   decoded.CurrentTurn.QuestionText,
			QuestionIntent: intent,
			Status:         decoded.CurrentTurn.Status,
			AskedAt:        askedAt,
		}
	}
	return session, nil
}

var _ domain.Store = (*SQLRepository)(nil)
