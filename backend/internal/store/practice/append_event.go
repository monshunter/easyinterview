package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type appendSessionContext struct {
	session    domain.SessionRecord
	plan       domain.PlanRecord
	latestTurn domain.TurnRecord
}

func (r *SQLRepository) ReserveSessionEvent(ctx context.Context, in domain.SessionEventReservationInput) (domain.SessionEventReservation, error) {
	if r == nil || r.db == nil {
		return domain.SessionEventReservation{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionEventReservation{}, fmt.Errorf("begin reserve practice session event: %w", err)
	}
	defer tx.Rollback()

	state, err := selectAppendSessionContext(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.SessionEventReservation{}, err
	}
	replay, replayErr, hit, pending, err := selectSessionEventReplay(ctx, tx, in.SessionID, in.ClientEventID, in.RequestFingerprint)
	if err != nil {
		return domain.SessionEventReservation{}, err
	}
	if hit && pending {
		return domain.SessionEventReservation{}, domain.ErrSessionConflict
	}
	if !hit {
		if err := validateSessionEventReservation(in, state); err != nil {
			return domain.SessionEventReservation{}, err
		}
		payload, err := marshalPendingAppendEventPayload(in)
		if err != nil {
			return domain.SessionEventReservation{}, err
		}
		seqNo, err := nextSessionEventSeq(ctx, tx, in.SessionID)
		if err != nil {
			return domain.SessionEventReservation{}, err
		}
		if _, err := tx.ExecContext(ctx, `
insert into practice_session_events (
  id, session_id, seq_no, event_type, client_event_id, payload, created_at
) values ($1,$2,$3,$4,$5,$6,$7)`,
			in.EventID,
			in.SessionID,
			seqNo,
			in.Kind,
			in.ClientEventID,
			payload,
			in.Now.UTC(),
		); err != nil {
			return domain.SessionEventReservation{}, fmt.Errorf("insert pending practice session event: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return domain.SessionEventReservation{}, fmt.Errorf("commit reserve practice session event: %w", err)
	}
	return domain.SessionEventReservation{
		UserID:       strings.TrimSpace(in.UserID),
		Session:      state.session,
		Plan:         state.plan,
		LatestTurn:   state.latestTurn,
		ReplayResult: replayIfHit(replay, hit),
		ReplayError:  replayErrorIfHit(replayErr, hit),
	}, nil
}

func (r *SQLRepository) FinalizeSessionEventError(ctx context.Context, in domain.FinalizeSessionEventErrorInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin finalize practice session event error: %w", err)
	}
	defer tx.Rollback()
	payload, err := marshalAppendEventErrorPayload(in)
	if err != nil {
		return err
	}
	if err := updateReservedSessionEventErrorPayload(ctx, tx, in, payload); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit finalize practice session event error: %w", err)
	}
	return nil
}

func (r *SQLRepository) AppendSessionEvent(ctx context.Context, in domain.AppendSessionEventStoreInput) (domain.AppendSessionEventResult, error) {
	if r == nil || r.db == nil {
		return domain.AppendSessionEventResult{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.AppendSessionEventResult{}, fmt.Errorf("begin append practice session event: %w", err)
	}
	defer tx.Rollback()

	state, err := selectAppendSessionContext(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	replay, replayErr, hit, pending, err := selectSessionEventReplay(ctx, tx, in.SessionID, in.ClientEventID, in.RequestFingerprint)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	if hit && !pending {
		if err := tx.Commit(); err != nil {
			return domain.AppendSessionEventResult{}, fmt.Errorf("commit append practice session event replay: %w", err)
		}
		if replayErr != nil {
			return domain.AppendSessionEventResult{}, replayErr
		}
		replay.Replay = true
		return replay, nil
	}
	if !hit {
		return domain.AppendSessionEventResult{}, domain.ErrSessionConflict
	}
	if in.Outcome.NextTurn != nil && strings.TrimSpace(in.Outcome.NextTurn.ID) != "" && in.Outcome.NextTurn.ID != state.latestTurn.ID {
		return domain.AppendSessionEventResult{}, domain.ErrSessionConflict
	}

	if in.Outcome.NextTurn != nil {
		if err := updateLatestTurn(ctx, tx, in, state.latestTurn); err != nil {
			return domain.AppendSessionEventResult{}, err
		}
	}
	if in.Outcome.AssistantAction.Type == "show_hint" && strings.TrimSpace(in.Outcome.AssistantAction.Hint) != "" {
		if err := updateTurnHintText(ctx, tx, in, state.latestTurn); err != nil {
			return domain.AppendSessionEventResult{}, err
		}
	}
	session := state.session
	session.Status = in.Outcome.NextSessionStatus
	session.UpdatedAt = in.OccurredAt.UTC()
	if in.NextQuestion != nil {
		if err := insertNextTurn(ctx, tx, in.NextQuestion, in.SessionID, state.plan.InterviewerPersona); err != nil {
			return domain.AppendSessionEventResult{}, err
		}
		session.TurnCount = in.NextQuestion.TurnIndex
		session.CurrentTurn = in.NextQuestion
	} else if in.Outcome.NextTurn != nil {
		next := *in.Outcome.NextTurn
		session.CurrentTurn = &next
	} else {
		next := state.latestTurn
		session.CurrentTurn = &next
	}
	if err := updateSessionAfterAppend(ctx, tx, in.SessionID, in.UserID, session.Status, session.TurnCount, in.OccurredAt.UTC()); err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	if in.Outcome.OutboxRecord != nil {
		if err := insertTurnCompletedOutbox(ctx, tx, in, state.latestTurn); err != nil {
			return domain.AppendSessionEventResult{}, err
		}
	}
	result := domain.AppendSessionEventResult{
		Acknowledged:    in.Outcome.Acknowledged,
		Session:         session,
		AssistantAction: in.Outcome.AssistantAction,
	}
	payload, err := marshalAppendEventPayload(in, result)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	replayPayload, err := marshalAppendEventReplayPayload(in.RequestFingerprint, result)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	if err := updateReservedSessionEventPayload(ctx, tx, in, payload, replayPayload); err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.AppendSessionEventResult{}, fmt.Errorf("commit append practice session event: %w", err)
	}
	return result, nil
}

func selectAppendSessionContext(ctx context.Context, tx *sql.Tx, userID, sessionID string) (appendSessionContext, error) {
	var state appendSessionContext
	var turnID sql.NullString
	var turnIndex sql.NullInt64
	var questionText sql.NullString
	var questionIntent sql.NullString
	var turnStatus sql.NullString
	var followUpCount sql.NullInt64
	var askedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
select s.id, s.plan_id, s.target_job_id, s.status, s.language, s.hints_enabled,
       s.turn_count, s.created_at, s.updated_at,
       p.id, p.target_job_id, p.goal, p.mode, p.interviewer_persona, p.difficulty,
       p.language, p.time_budget_minutes, p.question_budget, p.status, p.created_at,
       t.id, t.turn_index, t.question_text, t.question_intent, t.status, t.follow_up_count, t.asked_at
from practice_sessions s
join practice_plans p on p.id = s.plan_id
left join lateral (
  select id, turn_index, question_text, question_intent, status, follow_up_count, asked_at
  from practice_turns
  where session_id = s.id
  order by turn_index desc
  limit 1
) t on true
where s.user_id = $1
  and s.id = $2
for update of s`,
		userID,
		sessionID,
	).Scan(
		&state.session.ID,
		&state.session.PlanID,
		&state.session.TargetJobID,
		&state.session.Status,
		&state.session.Language,
		&state.session.HintsEnabled,
		&state.session.TurnCount,
		&state.session.CreatedAt,
		&state.session.UpdatedAt,
		&state.plan.ID,
		&state.plan.TargetJobID,
		&state.plan.Goal,
		&state.plan.Mode,
		&state.plan.InterviewerPersona,
		&state.plan.Difficulty,
		&state.plan.Language,
		&state.plan.TimeBudgetMinutes,
		&state.plan.QuestionBudget,
		&state.plan.Status,
		&state.plan.CreatedAt,
		&turnID,
		&turnIndex,
		&questionText,
		&questionIntent,
		&turnStatus,
		&followUpCount,
		&askedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return appendSessionContext{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return appendSessionContext{}, fmt.Errorf("select practice session for append: %w", err)
	}
	if !turnID.Valid {
		return appendSessionContext{}, domain.ErrSessionConflict
	}
	state.latestTurn = domain.TurnRecord{
		ID:             turnID.String,
		TurnIndex:      int32(turnIndex.Int64),
		QuestionText:   questionText.String,
		QuestionIntent: questionIntent.String,
		Status:         turnStatus.String,
		FollowUpCount:  int(followUpCount.Int64),
		AskedAt:        askedAt.Time,
	}
	turn := state.latestTurn
	state.session.CurrentTurn = &turn
	return state, nil
}

func selectSessionEventReplay(ctx context.Context, tx *sql.Tx, sessionID, clientEventID, fingerprint string) (domain.AppendSessionEventResult, *domain.ServiceError, bool, bool, error) {
	var raw []byte
	var replayRaw sql.NullString
	err := tx.QueryRowContext(ctx, `
	select payload, replay_payload
	from practice_session_events
	where session_id = $1
	  and client_event_id = $2`,
		sessionID,
		clientEventID,
	).Scan(&raw, &replayRaw)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.AppendSessionEventResult{}, nil, false, false, nil
	}
	if err != nil {
		return domain.AppendSessionEventResult{}, nil, false, false, fmt.Errorf("select practice session event replay: %w", err)
	}
	storedFingerprint, result, resultErr, pending, err := unmarshalAppendEventPayload(raw)
	if err != nil {
		return domain.AppendSessionEventResult{}, nil, false, false, err
	}
	if storedFingerprint != fingerprint {
		return domain.AppendSessionEventResult{}, nil, true, pending, domain.ErrClientEventMismatch
	}
	if !pending && resultErr == nil && strings.TrimSpace(replayRaw.String) != "" {
		replayFingerprint, replayResult, replayResultErr, replayPending, err := unmarshalAppendEventPayload([]byte(replayRaw.String))
		if err != nil {
			return domain.AppendSessionEventResult{}, nil, false, false, err
		}
		if replayFingerprint != "" && replayFingerprint != fingerprint {
			return domain.AppendSessionEventResult{}, nil, true, pending, domain.ErrClientEventMismatch
		}
		if replayPending {
			return domain.AppendSessionEventResult{}, nil, true, true, nil
		}
		if replayResultErr != nil {
			return domain.AppendSessionEventResult{}, replayResultErr, true, pending, nil
		}
		result = replayResult
	}
	return result, resultErr, true, pending, nil
}

func replayIfHit(result domain.AppendSessionEventResult, hit bool) *domain.AppendSessionEventResult {
	if !hit {
		return nil
	}
	return &result
}

func replayErrorIfHit(resultErr *domain.ServiceError, hit bool) *domain.ServiceError {
	if !hit || resultErr == nil {
		return nil
	}
	return resultErr
}

func validateSessionEventReservation(in domain.SessionEventReservationInput, state appendSessionContext) error {
	if isClosedSessionStatus(state.session.Status) {
		return domain.ErrSessionConflict
	}
	if strings.TrimSpace(in.CurrentTurnID) == "" {
		return nil
	}
	if strings.TrimSpace(in.CurrentTurnID) != state.latestTurn.ID {
		return domain.ErrSessionConflict
	}
	if isClosedTurnStatus(state.latestTurn.Status) {
		return domain.ErrSessionConflict
	}
	return nil
}

func isClosedSessionStatus(status sharedtypes.SessionStatus) bool {
	switch status {
	case sharedtypes.SessionStatusCompleting,
		sharedtypes.SessionStatusCompleted,
		sharedtypes.SessionStatusFailed,
		sharedtypes.SessionStatusCancelled:
		return true
	default:
		return false
	}
}

func isClosedTurnStatus(status string) bool {
	switch domain.TurnStatus(strings.TrimSpace(status)) {
	case domain.TurnStatusAnswered, domain.TurnStatusAssessed:
		return true
	default:
		return false
	}
}

func updateLatestTurn(ctx context.Context, tx *sql.Tx, in domain.AppendSessionEventStoreInput, latest domain.TurnRecord) error {
	answerText := payloadString(in.RequestPayload, "answerText")
	followUpCount := in.Outcome.NextTurn.FollowUpCount
	completedAt := any(nil)
	if in.Outcome.NextTurn.Status == string(domain.TurnStatusAssessed) {
		completedAt = in.OccurredAt.UTC()
	}
	answeredAt := any(nil)
	if strings.TrimSpace(answerText) != "" {
		answeredAt = in.OccurredAt.UTC()
	}
	_, err := tx.ExecContext(ctx, `
update practice_turns
set status = $1,
    answer_text = coalesce($2, answer_text),
    answer_summary = coalesce($3, answer_summary),
    follow_up_count = $4,
    answered_at = coalesce($5, answered_at),
    completed_at = coalesce($6, completed_at),
    updated_at = $7
where session_id = $8
  and id = $9`,
		in.Outcome.NextTurn.Status,
		nullableString(answerText),
		nullableString(in.Outcome.AnswerSummary),
		followUpCount,
		answeredAt,
		completedAt,
		in.OccurredAt.UTC(),
		in.SessionID,
		latest.ID,
	)
	if err != nil {
		return fmt.Errorf("update practice turn after event: %w", err)
	}
	return nil
}

func updateTurnHintText(ctx context.Context, tx *sql.Tx, in domain.AppendSessionEventStoreInput, latest domain.TurnRecord) error {
	turnID := strings.TrimSpace(in.Outcome.AssistantAction.TurnID)
	if turnID == "" {
		turnID = latest.ID
	}
	_, err := tx.ExecContext(ctx, `
update practice_turns
set hint_text = $1,
    updated_at = $2
where session_id = $3
  and id = $4`,
		strings.TrimSpace(in.Outcome.AssistantAction.Hint),
		in.OccurredAt.UTC(),
		in.SessionID,
		turnID,
	)
	if err != nil {
		return fmt.Errorf("update practice turn hint after event: %w", err)
	}
	return nil
}

func insertNextTurn(ctx context.Context, tx *sql.Tx, turn *domain.TurnRecord, sessionID string, persona sharedtypes.InterviewerRole) error {
	if turn == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
insert into practice_turns (
  id, session_id, turn_index, question_text, question_intent,
  interviewer_persona, status, asked_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$8,$8)`,
		turn.ID,
		sessionID,
		turn.TurnIndex,
		turn.QuestionText,
		nullableString(turn.QuestionIntent),
		string(persona),
		turn.Status,
		turn.AskedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert next practice turn: %w", err)
	}
	return nil
}

func updateSessionAfterAppend(ctx context.Context, tx *sql.Tx, sessionID, userID string, status sharedtypes.SessionStatus, turnCount int32, updatedAt time.Time) error {
	res, err := tx.ExecContext(ctx, `
update practice_sessions
set status = $1,
    turn_count = $2,
    updated_at = $3
where id = $4
  and user_id = $5`,
		string(status),
		turnCount,
		updatedAt.UTC(),
		sessionID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update practice session after event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update practice session after event rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func insertTurnCompletedOutbox(ctx context.Context, tx *sql.Tx, in domain.AppendSessionEventStoreInput, latest domain.TurnRecord) error {
	payload, err := BuildPracticeTurnCompletedPayload(PracticeTurnCompletedInput{
		SessionID:        in.SessionID,
		TurnID:           latest.ID,
		TurnIndex:        int(latest.TurnIndex),
		QuestionIntent:   latest.QuestionIntent,
		FollowUpCount:    in.Outcome.OutboxRecord.FollowUpCount,
		AnswerCharLength: in.Outcome.OutboxRecord.AnswerCharLength,
	})
	if err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal practice turn completed payload: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,1,'practice_turn',$3,$4,'pending',$5)`,
		in.OutboxEventID,
		string(sharedevents.EventNamePracticeTurnCompleted),
		latest.ID,
		raw,
		in.OccurredAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert practice turn completed outbox event: %w", err)
	}
	return nil
}

func nextSessionEventSeq(ctx context.Context, tx *sql.Tx, sessionID string) (int, error) {
	var seq int
	err := tx.QueryRowContext(ctx, `
select coalesce(max(seq_no), 0) + 1
from practice_session_events
where session_id = $1`,
		sessionID,
	).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("select next practice session event seq: %w", err)
	}
	return seq, nil
}

func updateReservedSessionEventPayload(ctx context.Context, tx *sql.Tx, in domain.AppendSessionEventStoreInput, payload, replayPayload []byte) error {
	res, err := tx.ExecContext(ctx, `
	update practice_session_events
	set payload = $1,
	    replay_payload = $2
	where session_id = $3
	  and client_event_id = $4
	  and id = $5`,
		payload,
		replayPayload,
		in.SessionID,
		in.ClientEventID,
		in.EventID,
	)
	if err != nil {
		return fmt.Errorf("update reserved practice session event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update reserved practice session event rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionConflict
	}
	return nil
}

func updateReservedSessionEventErrorPayload(ctx context.Context, tx *sql.Tx, in domain.FinalizeSessionEventErrorInput, payload []byte) error {
	res, err := tx.ExecContext(ctx, `
	update practice_session_events
	set payload = $1,
	    replay_payload = null
	where session_id = $2
	  and client_event_id = $3
	  and id = $4`,
		payload,
		in.SessionID,
		in.ClientEventID,
		in.EventID,
	)
	if err != nil {
		return fmt.Errorf("update reserved practice session event error: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update reserved practice session event error rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionConflict
	}
	return nil
}

type appendEventPayload struct {
	RequestFingerprint string                   `json:"requestFingerprint"`
	RequestPayload     map[string]any           `json:"requestPayload"`
	Pending            bool                     `json:"pending,omitempty"`
	Result             appendEventResultPayload `json:"result"`
	Error              *appendEventErrorPayload `json:"error,omitempty"`
}

type appendEventFinalizedErrorPayload struct {
	RequestFingerprint string                   `json:"requestFingerprint"`
	Error              *appendEventErrorPayload `json:"error,omitempty"`
}

type appendEventErrorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type appendEventResultPayload struct {
	Acknowledged    bool                       `json:"acknowledged"`
	Session         appendEventSessionPayload  `json:"session"`
	AssistantAction appendEventAssistantAction `json:"assistantAction"`
}

type appendEventSessionPayload struct {
	ID           string                    `json:"id"`
	PlanID       string                    `json:"planId"`
	TargetJobID  string                    `json:"targetJobId"`
	Status       sharedtypes.SessionStatus `json:"status"`
	Language     string                    `json:"language"`
	HintsEnabled bool                      `json:"hintsEnabled"`
	TurnCount    int32                     `json:"turnCount"`
	CurrentTurn  *appendEventTurnPayload   `json:"currentTurn,omitempty"`
	CreatedAt    string                    `json:"createdAt"`
	UpdatedAt    string                    `json:"updatedAt"`
}

type appendEventTurnPayload struct {
	ID             string `json:"id"`
	TurnIndex      int32  `json:"turnIndex"`
	QuestionText   string `json:"questionText"`
	QuestionIntent string `json:"questionIntent"`
	Status         string `json:"status"`
	AskedAt        string `json:"askedAt"`
}

type appendEventAssistantAction struct {
	Type          string                           `json:"type"`
	TurnID        string                           `json:"turnId,omitempty"`
	QuestionText  string                           `json:"questionText,omitempty"`
	Hint          string                           `json:"hint,omitempty"`
	SessionStatus sharedtypes.SessionStatus        `json:"sessionStatus"`
	Provenance    domain.AssistantActionProvenance `json:"provenance"`
}

func marshalAppendEventPayload(in domain.AppendSessionEventStoreInput, result domain.AppendSessionEventResult) ([]byte, error) {
	raw, err := json.Marshal(appendEventPayload{
		RequestFingerprint: in.RequestFingerprint,
		RequestPayload:     in.RequestPayload,
		Result:             appendEventResultFromDomain(result),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal append session event payload: %w", err)
	}
	return raw, nil
}

func marshalAppendEventReplayPayload(requestFingerprint string, result domain.AppendSessionEventResult) ([]byte, error) {
	raw, err := json.Marshal(appendEventPayload{
		RequestFingerprint: requestFingerprint,
		Result:             appendEventReplayResultFromDomain(result),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal append session event replay payload: %w", err)
	}
	return raw, nil
}

func marshalPendingAppendEventPayload(in domain.SessionEventReservationInput) ([]byte, error) {
	raw, err := json.Marshal(appendEventPayload{
		RequestFingerprint: in.RequestFingerprint,
		Pending:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal pending append session event payload: %w", err)
	}
	return raw, nil
}

func marshalAppendEventErrorPayload(in domain.FinalizeSessionEventErrorInput) ([]byte, error) {
	payload := appendEventFinalizedErrorPayload{
		RequestFingerprint: in.RequestFingerprint,
	}
	if in.Error != nil {
		payload.Error = &appendEventErrorPayload{
			Code:    in.Error.Code,
			Message: in.Error.Message,
			Details: in.Error.Details,
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal append session event error payload: %w", err)
	}
	return raw, nil
}

func unmarshalAppendEventPayload(raw []byte) (string, domain.AppendSessionEventResult, *domain.ServiceError, bool, error) {
	var decoded appendEventPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", domain.AppendSessionEventResult{}, nil, false, fmt.Errorf("decode append session event payload: %w", err)
	}
	var resultErr *domain.ServiceError
	if decoded.Error != nil {
		resultErr = &domain.ServiceError{
			Code:    decoded.Error.Code,
			Message: decoded.Error.Message,
			Details: decoded.Error.Details,
		}
	}
	return decoded.RequestFingerprint, appendEventResultToDomain(decoded.Result), resultErr, decoded.Pending, nil
}

func appendEventResultFromDomain(result domain.AppendSessionEventResult) appendEventResultPayload {
	return appendEventResultPayload{
		Acknowledged:    result.Acknowledged,
		Session:         appendEventSessionFromDomain(result.Session),
		AssistantAction: appendEventAssistantActionFromDomain(result.AssistantAction, true),
	}
}

func appendEventReplayResultFromDomain(result domain.AppendSessionEventResult) appendEventResultPayload {
	return appendEventResultPayload{
		Acknowledged:    result.Acknowledged,
		Session:         appendEventSessionFromDomain(result.Session),
		AssistantAction: appendEventAssistantActionFromDomain(result.AssistantAction, false),
	}
}

func appendEventSessionFromDomain(session domain.SessionRecord) appendEventSessionPayload {
	var turn *appendEventTurnPayload
	if session.CurrentTurn != nil {
		turn = &appendEventTurnPayload{
			ID:             session.CurrentTurn.ID,
			TurnIndex:      session.CurrentTurn.TurnIndex,
			QuestionText:   session.CurrentTurn.QuestionText,
			QuestionIntent: session.CurrentTurn.QuestionIntent,
			Status:         session.CurrentTurn.Status,
			AskedAt:        session.CurrentTurn.AskedAt.UTC().Format(time.RFC3339),
		}
	}
	return appendEventSessionPayload{
		ID:           session.ID,
		PlanID:       session.PlanID,
		TargetJobID:  session.TargetJobID,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn:  turn,
		CreatedAt:    session.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    session.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func appendEventAssistantActionFromDomain(action domain.AssistantActionRecord, redactHint bool) appendEventAssistantAction {
	hint := action.Hint
	if redactHint && action.Type == "show_hint" {
		hint = ""
	}
	return appendEventAssistantAction{
		Type:          action.Type,
		TurnID:        action.TurnID,
		QuestionText:  action.QuestionText,
		Hint:          hint,
		SessionStatus: action.SessionStatus,
		Provenance:    action.Provenance,
	}
}

func appendEventResultToDomain(result appendEventResultPayload) domain.AppendSessionEventResult {
	return domain.AppendSessionEventResult{
		Acknowledged:    result.Acknowledged,
		Session:         appendEventSessionToDomain(result.Session),
		AssistantAction: appendEventAssistantActionToDomain(result.AssistantAction),
	}
}

func appendEventSessionToDomain(session appendEventSessionPayload) domain.SessionRecord {
	createdAt, _ := time.Parse(time.RFC3339, session.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, session.UpdatedAt)
	var turn *domain.TurnRecord
	if session.CurrentTurn != nil {
		askedAt, _ := time.Parse(time.RFC3339, session.CurrentTurn.AskedAt)
		turn = &domain.TurnRecord{
			ID:             session.CurrentTurn.ID,
			TurnIndex:      session.CurrentTurn.TurnIndex,
			QuestionText:   session.CurrentTurn.QuestionText,
			QuestionIntent: session.CurrentTurn.QuestionIntent,
			Status:         session.CurrentTurn.Status,
			AskedAt:        askedAt,
		}
	}
	return domain.SessionRecord{
		ID:           session.ID,
		PlanID:       session.PlanID,
		TargetJobID:  session.TargetJobID,
		Status:       session.Status,
		Language:     session.Language,
		HintsEnabled: session.HintsEnabled,
		TurnCount:    session.TurnCount,
		CurrentTurn:  turn,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func appendEventAssistantActionToDomain(action appendEventAssistantAction) domain.AssistantActionRecord {
	return domain.AssistantActionRecord{
		Type:          action.Type,
		TurnID:        action.TurnID,
		QuestionText:  action.QuestionText,
		Hint:          action.Hint,
		SessionStatus: action.SessionStatus,
		Provenance:    action.Provenance,
	}
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}
