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
	focusDimensionCodes := []string{}

	var plan domain.PlanRecord
	var sourceReportID sql.NullString
	if in.Goal == sharedtypes.PracticeGoalRetryCurrentRound || in.Goal == sharedtypes.PracticeGoalNextRound {
		plan, err = createDerivedPlan(ctx, tx, in)
	} else {
		err = tx.QueryRowContext(ctx, `
with selected_target as (
  select tj.id,
         case when jsonb_typeof(tj.summary->'interviewRounds') = 'array'
              then tj.summary->'interviewRounds' else '[]'::jsonb end rounds
  from target_jobs tj
  where tj.id = $3
    and tj.user_id = $2
    and tj.resume_id = $10
    and tj.deleted_at is null
    and jsonb_typeof(tj.summary->'provenance') = 'object'
    and nullif(btrim(tj.summary#>>'{provenance,promptVersion}'), '') is not null
    and nullif(btrim(tj.summary#>>'{provenance,rubricVersion}'), '') is not null
    and nullif(btrim(tj.summary#>>'{provenance,modelId}'), '') is not null
    and nullif(btrim(tj.summary#>>'{provenance,language}'), '') is not null
    and nullif(btrim(tj.summary#>>'{provenance,dataSourceVersion}'), '') is not null
), raw_rounds as (
  select st.id target_job_id,
         case
           when value->>'sequence' ~ '^[1-9][0-9]{0,9}$'
            and (length(value->>'sequence') < 10 or value->>'sequence' <= '2147483647')
           then (value->>'sequence')::int
         end round_sequence,
         btrim(value->>'type') round_type,
         btrim(value->>'name') round_name,
         case when value->>'durationMinutes' ~ '^[0-9]{1,3}$' then (value->>'durationMinutes')::int end duration_minutes,
         btrim(value->>'focus') round_focus
  from selected_target st
  cross join lateral jsonb_array_elements(st.rounds) value
), canonical_rounds as (
  select target_job_id, round_sequence, round_type, round_name, duration_minutes, round_focus,
         'round-' || round_sequence::text || '-' || round_type round_id
  from raw_rounds
  where round_sequence > 0
    and round_type in ('hr','technical','manager','cross_functional','culture','final','other')
    and nullif(round_name, '') is not null
    and duration_minutes between 10 and 180
    and nullif(round_focus, '') is not null
), valid_rounds as (
  select cr.*
  from canonical_rounds cr
  join selected_target st on st.id = cr.target_job_id
  where jsonb_array_length(st.rounds) between 2 and 5
    and (select count(*) from canonical_rounds) = jsonb_array_length(st.rounds)
    and (select count(distinct round_sequence) from canonical_rounds) = jsonb_array_length(st.rounds)
    and (select count(distinct round_id) from canonical_rounds) = jsonb_array_length(st.rounds)
), completed_rounds as (
  select distinct pp.round_id, pp.round_sequence
  from practice_plans pp
  join practice_sessions ps
    on ps.plan_id = pp.id
   and ps.user_id = pp.user_id
   and ps.target_job_id = pp.target_job_id
  where pp.user_id = $2 and pp.target_job_id = $3
    and pp.resume_id = $10
    and pp.round_id is not null and pp.round_sequence is not null
    and exists (
      select 1 from practice_session_events pse
      where pse.session_id = ps.id and pse.event_type = 'session_completed'
    )
), current_round as (
  select vr.*
  from valid_rounds vr
  where not exists (
    select 1 from completed_rounds done
    where done.round_id = vr.round_id and done.round_sequence = vr.round_sequence
  )
  order by vr.round_sequence, vr.round_id
  limit 1
), source_round as (
  select pp.round_id, pp.round_sequence
  from feedback_reports fr
  join practice_sessions ps
    on ps.id = fr.session_id and ps.user_id = fr.user_id and ps.target_job_id = fr.target_job_id
  join practice_plans pp
    on pp.id = ps.plan_id and pp.user_id = ps.user_id and pp.target_job_id = ps.target_job_id
  where fr.id = nullif($4, '')::uuid
    and fr.user_id = $2 and fr.target_job_id = $3 and fr.status = 'ready'
    and pp.resume_id = $10
    and pp.round_id is not null and pp.round_sequence is not null
    and exists (
      select 1 from practice_session_events pse
      where pse.session_id = ps.id and pse.event_type = 'session_completed'
    )
), candidate_round as (
  select current.* from current_round current
  where $5 = 'baseline' and nullif($4, '') is null
  union all
  select vr.*
  from source_round source
  join valid_rounds vr
    on vr.round_id = source.round_id and vr.round_sequence = source.round_sequence
  where $5 = 'retry_current_round'
  union all
  select successor.*
  from source_round source
  cross join lateral (
    select vr.* from valid_rounds vr
    where vr.round_sequence > source.round_sequence
    order by vr.round_sequence, vr.round_id
    limit 1
  ) successor
  join current_round current
    on current.round_id = successor.round_id and current.round_sequence = successor.round_sequence
  where $5 = 'next_round'
), validated_round as (
  select * from candidate_round
  where duration_minutes = $9
    and (nullif(btrim($13), '') is null or round_id = btrim($13))
)
insert into practice_plans (
  id, user_id, target_job_id, source_report_id, goal, round_id, round_sequence,
  interviewer_persona, difficulty, language, time_budget_minutes,
	  resume_id, focus_dimension_codes, status, created_at, updated_at
)
select $1, $2, selected.target_job_id, nullif($4, '')::uuid, $5, selected.round_id, selected.round_sequence,
       $6, $7, $8, $9, r.id, $11, 'ready', $12, $12
from validated_round selected
join resumes r on r.id = $10 and r.user_id = $2 and r.deleted_at is null
returning id, target_job_id, source_report_id::text, goal, round_id, round_sequence,
          interviewer_persona, difficulty, language, time_budget_minutes,
          resume_id::text, focus_dimension_codes, status, created_at`,
			in.PlanID, in.UserID, in.TargetJobID, in.SourceReportID, string(in.Goal),
			string(in.InterviewerPersona), in.Difficulty, in.Language, in.TimeBudgetMinutes,
			in.ResumeID, pq.Array(focusDimensionCodes), in.Now, in.RoundID,
		).Scan(
			&plan.ID, &plan.TargetJobID, &sourceReportID, &plan.Goal,
			&plan.RoundID, &plan.RoundSequence,
			&plan.InterviewerPersona, &plan.Difficulty, &plan.Language,
			&plan.TimeBudgetMinutes, &plan.ResumeID, pq.Array(&plan.FocusDimensionCodes), &plan.Status, &plan.CreatedAt,
		)
	}
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanPrerequisiteNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("insert practice plan: %w", err)
	}
	if sourceReportID.Valid {
		plan.SourceReportID = sourceReportID.String
	}

	metadata, err := json.Marshal(map[string]any{
		"plan_id": in.PlanID, "goal": string(in.Goal), "language": plan.Language,
		"target_job_id": plan.TargetJobID, "source_report_id": plan.SourceReportID,
		"round_id": plan.RoundID, "round_sequence": plan.RoundSequence,
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
	var roundID sql.NullString
	var roundSequence sql.NullInt32
	err := r.db.QueryRowContext(ctx, `
select id, target_job_id, source_report_id::text, goal, round_id, round_sequence,
       interviewer_persona, difficulty, language, time_budget_minutes,
       resume_id::text, focus_dimension_codes, status, created_at
from practice_plans where user_id = $1 and id = $2`, userID, planID).Scan(
		&plan.ID, &plan.TargetJobID, &sourceReportID, &plan.Goal,
		&roundID, &roundSequence,
		&plan.InterviewerPersona, &plan.Difficulty, &plan.Language,
		&plan.TimeBudgetMinutes, &plan.ResumeID, pq.Array(&plan.FocusDimensionCodes), &plan.Status, &plan.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.PlanRecord{}, domain.ErrPlanNotFound
	}
	if err != nil {
		return domain.PlanRecord{}, fmt.Errorf("select practice plan: %w", err)
	}
	plan.SourceReportID = sourceReportID.String
	plan.RoundID = roundID.String
	if roundSequence.Valid {
		plan.RoundSequence = roundSequence.Int32
	}
	return plan, nil
}

func (r *SQLRepository) GetSession(ctx context.Context, userID, sessionID string, now time.Time) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("begin get practice session: %w", err)
	}
	defer tx.Rollback()
	if err := lockPracticeSessionForUser(ctx, tx, userID, sessionID); err != nil {
		return domain.SessionRecord{}, err
	}
	if _, err := tx.ExecContext(ctx, `
update practice_messages
set reply_status=$1, reply_lease_expires_at=null
where session_id=$2 and role='user' and reply_lease_expires_at <= $3 and reply_status=$4`,
		string(domain.PracticeReplyStatusRetryableFailed), sessionID, now, string(domain.PracticeReplyStatusPending)); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("expire pending practice message replies: %w", err)
	}
	session, err := selectSessionForUser(ctx, tx, userID, sessionID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("commit get practice session: %w", err)
	}
	return session, nil
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
	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, startSessionPlanAdvisoryLockKey(in)); err != nil {
		return domain.SessionReservation{}, fmt.Errorf("lock start session plan reservation: %w", err)
	}

	var reservation domain.SessionReservation
	var topSkills string
	var focusDimensionCodes pq.StringArray
	var semanticDimensions, semanticIssues []byte
	var recoverActive bool
	err = tx.QueryRowContext(ctx, `
with selected_plan as (
  select p.id, p.target_job_id, p.goal, p.interviewer_persona, p.language,
	         p.focus_dimension_codes,
	         case when p.goal='retry_current_round' and cardinality(p.focus_dimension_codes)>0 then fr.dimension_assessments end semantic_dimensions,
	         case when p.goal='retry_current_round' and cardinality(p.focus_dimension_codes)>0 then fr.issues end semantic_issues,
	         p.round_id, p.round_sequence,
         round_context.round_type, round_context.round_name, round_context.round_focus,
         coalesce(nullif(tj.title, ''), 'target role') role_title,
         coalesce(nullif(tj.seniority_level, ''), 'not specified') seniority,
         coalesce(nullif(array_to_string(array(
           select req.label from target_job_requirements req
           where req.target_job_id = p.target_job_id
           order by req.display_order asc, req.created_at asc limit 6
         ), ', '), ''), 'target job requirements') top_skills,
         coalesce(
           nullif(btrim(r.parsed_text_snapshot), ''),
           nullif(btrim(r.original_text), ''),
           case
             when r.structured_profile is not null
               and r.structured_profile <> '{}'::jsonb
               and r.structured_profile <> 'null'::jsonb
             then r.structured_profile::text
           end,
           ''
         ) resume_context
  from practice_plans p
	  join target_jobs tj on tj.id = p.target_job_id and tj.user_id = p.user_id and tj.resume_id = p.resume_id and tj.deleted_at is null
	  join resumes r on r.id = p.resume_id and r.user_id = p.user_id and r.deleted_at is null
	  left join feedback_reports fr
	    on fr.id=p.source_report_id and fr.user_id=p.user_id and fr.target_job_id=p.target_job_id and fr.status='ready'
  cross join lateral (
    select btrim(entry.value->>'type') round_type,
           btrim(entry.value->>'name') round_name,
           btrim(entry.value->>'focus') round_focus,
           case when entry.value->>'sequence' ~ '^[1-9][0-9]{0,9}$'
                 and (length(entry.value->>'sequence') < 10 or entry.value->>'sequence' <= '2147483647')
                then (entry.value->>'sequence')::int end round_sequence
    from jsonb_array_elements(
      case when jsonb_typeof(tj.summary->'interviewRounds') = 'array'
           then tj.summary->'interviewRounds' else '[]'::jsonb end
    ) entry(value)
  ) round_context
  where p.id = $3 and p.user_id = $2 and p.status = 'ready'
    and p.round_id is not null and p.round_sequence is not null
    and round_context.round_sequence = p.round_sequence
    and p.round_id = 'round-' || round_context.round_sequence::text || '-' || round_context.round_type
	    and nullif(round_context.round_name, '') is not null
	    and nullif(round_context.round_focus, '') is not null
	    and (p.goal='retry_current_round' or cardinality(p.focus_dimension_codes)=0)
	    and (cardinality(p.focus_dimension_codes)=0 or fr.id is not null)
), active_session as (
  select s.id, s.plan_id, s.target_job_id, s.language, s.created_at, s.updated_at
  from practice_sessions s
  join selected_plan on selected_plan.id = s.plan_id
  where s.user_id = $2 and s.status in ('queued', 'running')
  limit 1
  for update of s
), inserted as (
  insert into practice_sessions (id, user_id, plan_id, target_job_id, status, language, created_at, updated_at)
  select $1, $2, id, target_job_id, 'queued', language, $4, $4 from selected_plan
  where not exists (select 1 from active_session)
  returning id, plan_id, target_job_id, language, created_at, updated_at
), selected_session as (
  select id, plan_id, target_job_id, language, created_at, updated_at, true recover_active
  from active_session
  union all
  select id, plan_id, target_job_id, language, created_at, updated_at, false recover_active
  from inserted
)
select selected_session.id, selected_session.plan_id, selected_session.target_job_id, selected_plan.goal,
	       selected_plan.interviewer_persona, selected_session.language, selected_plan.role_title,
	       selected_plan.seniority, selected_plan.top_skills,
		       selected_plan.resume_context,
		       selected_plan.focus_dimension_codes, selected_plan.semantic_dimensions, selected_plan.semantic_issues,
	       selected_plan.round_id, selected_plan.round_sequence, selected_plan.round_type,
	       selected_plan.round_name, selected_plan.round_focus,
	       selected_session.created_at, selected_session.updated_at, selected_session.recover_active
from selected_session join selected_plan on selected_plan.id = selected_session.plan_id`,
		in.SessionID, in.UserID, in.PlanID, in.Now,
	).Scan(
		&reservation.SessionID, &reservation.PlanID, &reservation.TargetJobID, &reservation.Goal,
		&reservation.InterviewerPersona, &reservation.Language, &reservation.RoleTitle,
		&reservation.Seniority, &topSkills, &reservation.ResumeContext, &focusDimensionCodes,
		&semanticDimensions, &semanticIssues,
		&reservation.RoundID, &reservation.RoundSequence, &reservation.RoundType,
		&reservation.RoundName, &reservation.RoundFocus,
		&reservation.CreatedAt, &reservation.UpdatedAt, &recoverActive,
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
	reservation.SemanticFocus, err = resolveDerivedSemanticFocus(focusDimensionCodes, semanticDimensions, semanticIssues)
	if err != nil {
		return domain.SessionReservation{}, domain.ErrPlanNotFound
	}
	reservation.IdempotencyRecordID = recordID
	reservation.UserID = in.UserID
	if recoverActive {
		session, err := selectSessionForUser(ctx, tx, in.UserID, reservation.SessionID)
		if err != nil {
			return domain.SessionReservation{}, err
		}
		reservation.RecoverSession = &session
	}
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

func (r *SQLRepository) CommitSessionStartRecovery(ctx context.Context, in domain.CommitSessionStartRecoveryInput) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("begin commit practice session start recovery: %w", err)
	}
	defer tx.Rollback()
	session, err := selectSessionForUser(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if session.Status != sharedtypes.SessionStatusRunning {
		return domain.SessionRecord{}, domain.ErrSessionConflict
	}
	responseBody, err := marshalSessionResponseBody(session)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	result, err := tx.ExecContext(ctx, `
update idempotency_records set status = $1, resource_type = 'practice_session', resource_id = $2,
response_body = $3, error_code = null, updated_at = $4
where id = $5 and user_id = $6 and domain = 'practice' and operation = 'startPracticeSession' and status = $7`,
		string(idempotency.StatusSucceeded), session.ID, responseBody, in.RecoveredAt,
		in.IdempotencyRecordID, in.UserID, string(idempotency.StatusPending),
	)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("mark recovered start session idempotency succeeded: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return domain.SessionRecord{}, domain.ErrSessionConflict
	}
	if err := tx.Commit(); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("commit practice session start recovery: %w", err)
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
select id, role, content, seq_no, client_message_id::text, reply_status, created_at from practice_messages
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
		var clientMessageID, replyStatus sql.NullString
		if err := rows.Scan(&message.ID, &message.Role, &message.Content, &message.SeqNo, &clientMessageID, &replyStatus, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan practice message: %w", err)
		}
		message.ClientMessageID = clientMessageID.String
		message.ReplyStatus = domain.PracticeReplyStatus(replyStatus.String)
		switch message.Role {
		case "user":
			if !clientMessageID.Valid || !replyStatus.Valid || !validPracticeReplyStatus(message.ReplyStatus) {
				return nil, fmt.Errorf("scan practice message: user recovery state is invalid")
			}
		case "assistant":
			if clientMessageID.Valid || replyStatus.Valid {
				return nil, fmt.Errorf("scan practice message: assistant recovery state is invalid")
			}
		default:
			return nil, fmt.Errorf("scan practice message: role is invalid")
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate practice messages: %w", err)
	}
	return messages, nil
}

func validPracticeReplyStatus(status domain.PracticeReplyStatus) bool {
	switch status {
	case domain.PracticeReplyStatusPending,
		domain.PracticeReplyStatusRetryableFailed,
		domain.PracticeReplyStatusTerminalFailed,
		domain.PracticeReplyStatusComplete:
		return true
	default:
		return false
	}
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

func startSessionPlanAdvisoryLockKey(in domain.StartSessionReservationInput) string {
	return strings.Join([]string{in.UserID, "practice", "startPracticeSession", "plan", in.PlanID}, "\x1f")
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
