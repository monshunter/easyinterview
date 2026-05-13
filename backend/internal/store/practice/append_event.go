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
	replay, hit, err := selectSessionEventReplay(ctx, tx, in.SessionID, in.ClientEventID, in.RequestFingerprint)
	if err != nil {
		return domain.SessionEventReservation{}, err
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
	}, nil
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
	replay, hit, err := selectSessionEventReplay(ctx, tx, in.SessionID, in.ClientEventID, in.RequestFingerprint)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
	}
	if hit {
		if err := tx.Commit(); err != nil {
			return domain.AppendSessionEventResult{}, fmt.Errorf("commit append practice session event replay: %w", err)
		}
		replay.Replay = true
		return replay, nil
	}
	if in.Outcome.NextTurn != nil && strings.TrimSpace(in.Outcome.NextTurn.ID) != "" && in.Outcome.NextTurn.ID != state.latestTurn.ID {
		return domain.AppendSessionEventResult{}, domain.ErrSessionConflict
	}

	if in.Outcome.NextTurn != nil {
		if err := updateLatestTurn(ctx, tx, in, state.latestTurn); err != nil {
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
	seqNo, err := nextSessionEventSeq(ctx, tx, in.SessionID)
	if err != nil {
		return domain.AppendSessionEventResult{}, err
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
		in.OccurredAt.UTC(),
	); err != nil {
		return domain.AppendSessionEventResult{}, fmt.Errorf("insert practice session event: %w", err)
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

func selectSessionEventReplay(ctx context.Context, tx *sql.Tx, sessionID, clientEventID, fingerprint string) (domain.AppendSessionEventResult, bool, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `
select payload
from practice_session_events
where session_id = $1
  and client_event_id = $2`,
		sessionID,
		clientEventID,
	).Scan(&raw)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.AppendSessionEventResult{}, false, nil
	}
	if err != nil {
		return domain.AppendSessionEventResult{}, false, fmt.Errorf("select practice session event replay: %w", err)
	}
	storedFingerprint, result, err := unmarshalAppendEventPayload(raw)
	if err != nil {
		return domain.AppendSessionEventResult{}, false, err
	}
	if storedFingerprint != fingerprint {
		return domain.AppendSessionEventResult{}, true, domain.ErrClientEventMismatch
	}
	return result, true, nil
}

func replayIfHit(result domain.AppendSessionEventResult, hit bool) *domain.AppendSessionEventResult {
	if !hit {
		return nil
	}
	return &result
}

func updateLatestTurn(ctx context.Context, tx *sql.Tx, in domain.AppendSessionEventStoreInput, latest domain.TurnRecord) error {
	answerText := payloadString(in.RequestPayload, "answerText")
	followUpCount := in.Outcome.NextTurn.FollowUpCount
	completedAt := any(nil)
	if in.Outcome.NextTurn.Status == string(domain.TurnStatusAssessed) || in.Outcome.NextTurn.Status == string(domain.TurnStatusSkipped) {
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
    follow_up_count = $3,
    answered_at = coalesce($4, answered_at),
    completed_at = coalesce($5, completed_at),
    updated_at = $6
where session_id = $7
  and id = $8`,
		in.Outcome.NextTurn.Status,
		nullableString(answerText),
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

type appendEventPayload struct {
	RequestFingerprint string                   `json:"requestFingerprint"`
	RequestPayload     map[string]any           `json:"requestPayload"`
	Result             appendEventResultPayload `json:"result"`
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

func unmarshalAppendEventPayload(raw []byte) (string, domain.AppendSessionEventResult, error) {
	var decoded appendEventPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", domain.AppendSessionEventResult{}, fmt.Errorf("decode append session event payload: %w", err)
	}
	return decoded.RequestFingerprint, appendEventResultToDomain(decoded.Result), nil
}

func appendEventResultFromDomain(result domain.AppendSessionEventResult) appendEventResultPayload {
	return appendEventResultPayload{
		Acknowledged:    result.Acknowledged,
		Session:         appendEventSessionFromDomain(result.Session),
		AssistantAction: appendEventAssistantActionFromDomain(result.AssistantAction),
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

func appendEventAssistantActionFromDomain(action domain.AssistantActionRecord) appendEventAssistantAction {
	return appendEventAssistantAction{
		Type:          action.Type,
		TurnID:        action.TurnID,
		QuestionText:  action.QuestionText,
		Hint:          action.Hint,
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
