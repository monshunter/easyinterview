package practice

import (
	"context"
	"database/sql"
	stderrs "errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *SQLRepository) ReservePracticeMessage(ctx context.Context, in domain.ReservePracticeMessageInput) (domain.PracticeMessageReservation, error) {
	if r == nil || r.db == nil {
		return domain.PracticeMessageReservation{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PracticeMessageReservation{}, fmt.Errorf("begin reserve practice message: %w", err)
	}
	defer tx.Rollback()

	reservation, err := loadMessageReservationContext(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.PracticeMessageReservation{}, err
	}
	var existing domain.MessageRecord
	var assistantID, assistantContent sql.NullString
	var assistantSeq sql.NullInt64
	var assistantCreated sql.NullTime
	err = tx.QueryRowContext(ctx, `
select u.id, u.role, u.content, u.seq_no, u.created_at,
       a.id, a.content, a.seq_no, a.created_at
from practice_messages u
left join practice_messages a on a.reply_to_message_id = u.id
where u.session_id=$1 and u.client_message_id=$2`, in.SessionID, in.ClientMessageID).Scan(
		&existing.ID, &existing.Role, &existing.Content, &existing.SeqNo, &existing.CreatedAt,
		&assistantID, &assistantContent, &assistantSeq, &assistantCreated,
	)
	if err == nil {
		if existing.Content != in.Text {
			return domain.PracticeMessageReservation{}, domain.ErrClientEventMismatch
		}
		if !assistantID.Valid {
			rows, queryErr := queryMessages(ctx, tx, in.SessionID)
			if queryErr != nil {
				return domain.PracticeMessageReservation{}, queryErr
			}
			messages, scanErr := scanMessages(rows)
			rows.Close()
			if scanErr != nil {
				return domain.PracticeMessageReservation{}, scanErr
			}
			if len(messages) == 0 || messages[len(messages)-1].ID != existing.ID {
				return domain.PracticeMessageReservation{}, domain.ErrSessionConflict
			}
			if commitErr := tx.Commit(); commitErr != nil {
				return domain.PracticeMessageReservation{}, fmt.Errorf("commit pending practice message retry: %w", commitErr)
			}
			reservation.UserMessage = existing
			reservation.History = messages[:len(messages)-1]
			return reservation, nil
		}
		assistant := domain.MessageRecord{ID: assistantID.String, Role: "assistant", Content: assistantContent.String, SeqNo: int32(assistantSeq.Int64), CreatedAt: assistantCreated.Time}
		session, err := selectSessionForUser(ctx, tx, in.UserID, in.SessionID)
		if err != nil {
			return domain.PracticeMessageReservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return domain.PracticeMessageReservation{}, fmt.Errorf("commit practice message replay: %w", err)
		}
		return domain.PracticeMessageReservation{Replay: &domain.SendPracticeMessageResult{
			Acknowledged: true, UserMessage: existing, AssistantMessage: assistant, Session: session,
		}}, nil
	}
	if !stderrs.Is(err, sql.ErrNoRows) {
		return domain.PracticeMessageReservation{}, fmt.Errorf("select practice message replay: %w", err)
	}
	var hasPending bool
	if err := tx.QueryRowContext(ctx, `
select exists(
  select 1 from practice_messages u
  left join practice_messages a on a.reply_to_message_id=u.id
  where u.session_id=$1 and u.role='user' and a.id is null
)`, in.SessionID).Scan(&hasPending); err != nil {
		return domain.PracticeMessageReservation{}, fmt.Errorf("select pending practice message: %w", err)
	}
	if hasPending {
		return domain.PracticeMessageReservation{}, domain.ErrSessionConflict
	}

	var nextSeq int32
	if err := tx.QueryRowContext(ctx, `select coalesce(max(seq_no),0)+1 from practice_messages where session_id=$1`, in.SessionID).Scan(&nextSeq); err != nil {
		return domain.PracticeMessageReservation{}, fmt.Errorf("select next practice message sequence: %w", err)
	}
	userMessage := domain.MessageRecord{ID: in.UserMessageID, Role: "user", Content: in.Text, SeqNo: nextSeq, CreatedAt: in.Now}
	if _, err := tx.ExecContext(ctx, `
insert into practice_messages (id, session_id, seq_no, role, content, client_message_id, created_at)
values ($1,$2,$3,'user',$4,$5,$6)`, in.UserMessageID, in.SessionID, nextSeq, in.Text, in.ClientMessageID, in.Now); err != nil {
		if isUniqueViolation(err) {
			return domain.PracticeMessageReservation{}, domain.ErrSessionConflict
		}
		return domain.PracticeMessageReservation{}, fmt.Errorf("insert practice user message: %w", err)
	}
	rows, err := queryMessages(ctx, tx, in.SessionID)
	if err != nil {
		return domain.PracticeMessageReservation{}, err
	}
	history, err := scanMessages(rows)
	rows.Close()
	if err != nil {
		return domain.PracticeMessageReservation{}, err
	}
	if len(history) > 0 {
		history = history[:len(history)-1]
	}
	if err := tx.Commit(); err != nil {
		return domain.PracticeMessageReservation{}, fmt.Errorf("commit practice message reservation: %w", err)
	}
	reservation.UserMessage = userMessage
	reservation.History = history
	return reservation, nil
}

func loadMessageReservationContext(ctx context.Context, tx *sql.Tx, userID, sessionID string) (domain.PracticeMessageReservation, error) {
	var out domain.PracticeMessageReservation
	var topSkills string
	var focusDimensionCodes pq.StringArray
	var semanticDimensions, semanticIssues []byte
	err := tx.QueryRowContext(ctx, `
select s.id, s.plan_id, s.target_job_id, p.goal, p.interviewer_persona, s.language,
       coalesce(nullif(tj.title,''),'target role'), coalesce(nullif(tj.seniority_level,''),'not specified'),
       coalesce(nullif(array_to_string(array(
         select req.label from target_job_requirements req where req.target_job_id=s.target_job_id
         order by req.display_order asc, req.created_at asc limit 6
       ), ', '), ''), 'target job requirements'),
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
	       ), p.focus_dimension_codes,
	       case when p.goal='retry_current_round' and cardinality(p.focus_dimension_codes)>0 then fr.dimension_assessments end,
	       case when p.goal='retry_current_round' and cardinality(p.focus_dimension_codes)>0 then fr.issues end,
	       p.round_id, p.round_sequence,
       round_context.round_type, round_context.round_name, round_context.round_focus,
       s.created_at, s.updated_at
from practice_sessions s
join practice_plans p on p.id=s.plan_id and p.user_id=s.user_id
join target_jobs tj on tj.id=s.target_job_id and tj.user_id=s.user_id and tj.resume_id=p.resume_id and tj.deleted_at is null
join resumes r on r.id=p.resume_id and r.user_id=s.user_id and r.deleted_at is null
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
where s.id=$1 and s.user_id=$2 and s.status in ('running','waiting_user_input')
  and p.round_id is not null and p.round_sequence is not null
  and round_context.round_sequence = p.round_sequence
  and p.round_id = 'round-' || round_context.round_sequence::text || '-' || round_context.round_type
  and nullif(round_context.round_name, '') is not null
  and nullif(round_context.round_focus, '') is not null
	and (p.goal='retry_current_round' or cardinality(p.focus_dimension_codes)=0)
	and (cardinality(p.focus_dimension_codes)=0 or fr.id is not null)
for update of s`, sessionID, userID).Scan(
		&out.Session.SessionID, &out.Session.PlanID, &out.Session.TargetJobID, &out.Session.Goal,
		&out.Session.InterviewerPersona, &out.Session.Language, &out.Session.RoleTitle,
		&out.Session.Seniority, &topSkills, &out.Session.ResumeContext, &focusDimensionCodes,
		&semanticDimensions, &semanticIssues,
		&out.Session.RoundID, &out.Session.RoundSequence,
		&out.Session.RoundType, &out.Session.RoundName, &out.Session.RoundFocus,
		&out.Session.CreatedAt, &out.Session.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return out, domain.ErrSessionNotFound
	}
	if err != nil {
		return out, fmt.Errorf("load practice message reservation context: %w", err)
	}
	out.Session.UserID = userID
	out.Session.TopSkills = splitCommaList(topSkills)
	out.Session.SemanticFocus, err = resolveDerivedSemanticFocus(focusDimensionCodes, semanticDimensions, semanticIssues)
	if err != nil {
		return out, domain.ErrSessionNotFound
	}
	return out, nil
}

func (r *SQLRepository) CommitPracticeMessage(ctx context.Context, in domain.CommitPracticeMessageInput) (domain.SendPracticeMessageResult, error) {
	if r == nil || r.db == nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("begin commit practice message: %w", err)
	}
	defer tx.Rollback()
	var userMessage domain.MessageRecord
	err = tx.QueryRowContext(ctx, `
select m.id, m.role, m.content, m.seq_no, m.created_at
from practice_messages m join practice_sessions s on s.id=m.session_id
where m.id=$1 and m.session_id=$2 and s.user_id=$3 and m.role='user'
for update of m`, in.UserMessageID, in.SessionID, in.UserID).Scan(
		&userMessage.ID, &userMessage.Role, &userMessage.Content, &userMessage.SeqNo, &userMessage.CreatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.SendPracticeMessageResult{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("select reserved practice message: %w", err)
	}
	assistant := domain.MessageRecord{ID: in.AssistantMessageID, Role: "assistant", Content: strings.TrimSpace(in.AssistantText), SeqNo: userMessage.SeqNo + 1, CreatedAt: in.Now}
	if _, err := tx.ExecContext(ctx, `
insert into practice_messages (id, session_id, seq_no, role, content, reply_to_message_id, created_at)
values ($1,$2,$3,'assistant',$4,$5,$6)`, assistant.ID, in.SessionID, assistant.SeqNo, assistant.Content, userMessage.ID, in.Now); err != nil {
		if isUniqueViolation(err) {
			return domain.SendPracticeMessageResult{}, domain.ErrSessionConflict
		}
		return domain.SendPracticeMessageResult{}, fmt.Errorf("insert practice assistant message: %w", err)
	}
	result, err := tx.ExecContext(ctx, `update practice_sessions set status=$1, updated_at=$2 where id=$3 and user_id=$4 and status in ($5,$6)`,
		string(sharedtypes.SessionStatusRunning), in.Now, in.SessionID, in.UserID,
		string(sharedtypes.SessionStatusRunning), string(sharedtypes.SessionStatusWaitingUserInput))
	if err != nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("update practice session after message: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("count updated practice sessions after message: %w", err)
	}
	if updated != 1 {
		return domain.SendPracticeMessageResult{}, domain.ErrSessionConflict
	}
	session, err := selectSessionForUser(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.SendPracticeMessageResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.SendPracticeMessageResult{}, fmt.Errorf("commit practice message: %w", err)
	}
	return domain.SendPracticeMessageResult{Acknowledged: true, UserMessage: userMessage, AssistantMessage: assistant, Session: session}, nil
}
