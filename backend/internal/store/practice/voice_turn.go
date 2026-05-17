package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *SQLRepository) RecordPracticeVoiceTurn(ctx context.Context, in domain.PracticeVoiceTurnStoreInput) (domain.SessionRecord, error) {
	if r == nil || r.db == nil {
		return domain.SessionRecord{}, fmt.Errorf("practice SQL repository is not configured")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.SessionRecord{}, fmt.Errorf("begin record practice voice turn: %w", err)
	}
	defer tx.Rollback()

	state, err := selectAppendSessionContext(ctx, tx, in.UserID, in.SessionID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if err := validatePracticeVoiceTurnRecord(in, state); err != nil {
		return domain.SessionRecord{}, err
	}
	seqNo, err := nextSessionEventSeq(ctx, tx, in.SessionID)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if err := updatePracticeVoiceTurn(ctx, tx, in); err != nil {
		return domain.SessionRecord{}, err
	}
	session := state.session
	session.Status = sharedtypes.SessionStatusRunning
	session.UpdatedAt = in.OccurredAt.UTC()
	if session.CurrentTurn != nil {
		turn := *session.CurrentTurn
		turn.Status = string(domain.TurnStatusFollowUpRequested)
		turn.FollowUpCount = 1
		session.CurrentTurn = &turn
	}
	if err := updateSessionAfterAppend(ctx, tx, in.SessionID, in.UserID, session.Status, session.TurnCount, in.OccurredAt.UTC()); err != nil {
		return domain.SessionRecord{}, err
	}
	payload, err := marshalPracticeVoiceTurnEventPayload(in)
	if err != nil {
		return domain.SessionRecord{}, err
	}
	if err := insertPracticeVoiceTurnEvent(ctx, tx, in, seqNo, payload); err != nil {
		return domain.SessionRecord{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.SessionRecord{}, fmt.Errorf("commit record practice voice turn: %w", err)
	}
	return session, nil
}

func (r *SQLRepository) LoadCommittedVoiceContext(ctx context.Context, userID, sessionID string) (domain.CommittedVoiceContext, error) {
	if r == nil || r.db == nil {
		return domain.CommittedVoiceContext{}, fmt.Errorf("practice SQL repository is not configured")
	}
	var seqNo int
	var raw []byte
	err := r.db.QueryRowContext(ctx, `
select e.seq_no, e.payload
from practice_session_events e
join practice_sessions s on s.id = e.session_id
where s.user_id = $1
  and e.session_id = $2
  and e.event_type = 'follow_up_generated'
  and e.payload ? 'voiceTurnId'
order by e.seq_no desc
limit 1`,
		userID,
		sessionID,
	).Scan(&seqNo, &raw)
	if err == sql.ErrNoRows {
		return domain.CommittedVoiceContext{}, nil
	}
	if err != nil {
		return domain.CommittedVoiceContext{}, fmt.Errorf("load latest practice voice turn event: %w", err)
	}
	var sourcePayload practiceVoiceTurnEventPayload
	if err := json.Unmarshal(raw, &sourcePayload); err != nil {
		return domain.CommittedVoiceContext{}, fmt.Errorf("decode latest practice voice turn event: %w", err)
	}
	source := domain.PracticeVoiceTurnContextSource{
		VoiceTurnID:         strings.TrimSpace(sourcePayload.VoiceTurnID),
		AssistantTextDraft:  strings.TrimSpace(sourcePayload.AssistantTextDraft),
		AssistantTextLength: int32(len([]rune(strings.TrimSpace(sourcePayload.AssistantTextDraft)))),
	}
	if len(sourcePayload.TTSChunks) > 0 {
		source.AssistantTextHash = strings.TrimSpace(sourcePayload.TTSChunks[0].TextHash)
	}
	rows, err := r.db.QueryContext(ctx, `
select event_type, payload, created_at
from practice_session_events
where session_id = $1
  and seq_no > $2
  and event_type in ('tts_chunk_played', 'barge_in_detected', 'assistant_context_committed')
order by seq_no asc`,
		sessionID,
		seqNo,
	)
	if err != nil {
		return domain.CommittedVoiceContext{}, fmt.Errorf("load practice voice playback events: %w", err)
	}
	defer rows.Close()
	events := make([]domain.VoicePlaybackEventRecord, 0)
	for rows.Next() {
		var kind string
		var eventRaw []byte
		var createdAt sql.NullTime
		if err := rows.Scan(&kind, &eventRaw, &createdAt); err != nil {
			return domain.CommittedVoiceContext{}, fmt.Errorf("scan practice voice playback event: %w", err)
		}
		var payload storedAppendEventPayload
		if err := json.Unmarshal(eventRaw, &payload); err != nil {
			return domain.CommittedVoiceContext{}, fmt.Errorf("decode practice voice playback event: %w", err)
		}
		events = append(events, domain.VoicePlaybackEventRecord{
			Kind:       strings.TrimSpace(kind),
			OccurredAt: createdAt.Time,
			Payload:    payload.RequestPayload,
		})
	}
	if err := rows.Err(); err != nil {
		return domain.CommittedVoiceContext{}, fmt.Errorf("iterate practice voice playback events: %w", err)
	}
	return domain.BuildCommittedVoiceContext(source, events), nil
}

func validatePracticeVoiceTurnRecord(in domain.PracticeVoiceTurnStoreInput, state appendSessionContext) error {
	if isClosedSessionStatus(state.session.Status) {
		return domain.ErrSessionConflict
	}
	if strings.TrimSpace(in.TurnID) == "" || strings.TrimSpace(in.TurnID) != state.latestTurn.ID {
		return domain.ErrSessionConflict
	}
	if isClosedTurnStatus(state.latestTurn.Status) {
		return domain.ErrSessionConflict
	}
	return nil
}

func updatePracticeVoiceTurn(ctx context.Context, tx *sql.Tx, in domain.PracticeVoiceTurnStoreInput) error {
	res, err := tx.ExecContext(ctx, `
update practice_turns
set status = $1,
    answer_text = $2,
    follow_up_count = $3,
    answered_at = coalesce(answered_at, $4),
    updated_at = $5
where session_id = $6
  and id = $7`,
		string(domain.TurnStatusFollowUpRequested),
		strings.TrimSpace(in.UserTranscriptFinal),
		1,
		in.OccurredAt.UTC(),
		in.OccurredAt.UTC(),
		in.SessionID,
		in.TurnID,
	)
	if err != nil {
		return fmt.Errorf("update practice turn after voice turn: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update practice turn after voice turn rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionConflict
	}
	return nil
}

func insertPracticeVoiceTurnEvent(ctx context.Context, tx *sql.Tx, in domain.PracticeVoiceTurnStoreInput, seqNo int, payload []byte) error {
	res, err := tx.ExecContext(ctx, `
insert into practice_session_events (
  id, session_id, seq_no, event_type, client_event_id, payload, created_at
) values ($1,$2,$3,$4,$5,$6,$7)`,
		in.EventID,
		in.SessionID,
		seqNo,
		"follow_up_generated",
		in.ClientVoiceTurnID,
		payload,
		in.OccurredAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert practice voice turn event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("insert practice voice turn event rows affected: %w", err)
	}
	if rows == 0 {
		return domain.ErrSessionConflict
	}
	return nil
}

type practiceVoiceTurnEventPayload struct {
	VoiceTurnID         string                                  `json:"voiceTurnId"`
	TurnID              string                                  `json:"turnId"`
	UserTranscriptFinal string                                  `json:"userTranscriptFinal"`
	AssistantTextDraft  string                                  `json:"assistantTextDraft"`
	AudioByteLength     int32                                   `json:"audioByteLength"`
	AudioDurationMs     int32                                   `json:"audioDurationMs"`
	TTSChunks           []practiceVoiceTTSChunkPayload          `json:"ttsChunks"`
	TTSError            *practiceVoiceTTSErrorPayload           `json:"ttsError,omitempty"`
	ProviderMetaSummary practiceVoiceProviderMetaSummaryPayload `json:"providerMetaSummary"`
}

type practiceVoiceTTSChunkPayload struct {
	ChunkID     string `json:"chunkId"`
	Sequence    int32  `json:"sequence"`
	ContentType string `json:"contentType"`
	DurationMs  int32  `json:"durationMs"`
	ByteLength  int32  `json:"byteLength"`
	TextHash    string `json:"textHash"`
	AudioRef    string `json:"audioRef"`
}

type practiceVoiceProviderMetaSummaryPayload struct {
	STTProfile    string `json:"sttProfile"`
	STTProvider   string `json:"sttProvider"`
	STTLatencyMs  *int32 `json:"sttLatencyMs,omitempty"`
	ChatProfile   string `json:"chatProfile"`
	ChatProvider  string `json:"chatProvider"`
	ChatLatencyMs *int32 `json:"chatLatencyMs,omitempty"`
	TTSProfile    string `json:"ttsProfile"`
	TTSProvider   string `json:"ttsProvider"`
	TTSLatencyMs  *int32 `json:"ttsLatencyMs,omitempty"`
}

type practiceVoiceTTSErrorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type storedAppendEventPayload struct {
	RequestPayload map[string]any `json:"requestPayload"`
}

func marshalPracticeVoiceTurnEventPayload(in domain.PracticeVoiceTurnStoreInput) ([]byte, error) {
	payload := practiceVoiceTurnEventPayload{
		VoiceTurnID:         strings.TrimSpace(in.VoiceTurnID),
		TurnID:              strings.TrimSpace(in.TurnID),
		UserTranscriptFinal: strings.TrimSpace(in.UserTranscriptFinal),
		AssistantTextDraft:  strings.TrimSpace(in.AssistantTextDraft),
		AudioByteLength:     in.AudioByteLength,
		AudioDurationMs:     in.AudioDurationMs,
		TTSChunks:           make([]practiceVoiceTTSChunkPayload, 0, len(in.TTSChunks)),
		ProviderMetaSummary: practiceVoiceProviderMetaSummaryPayload{
			STTProfile:    in.ProviderMetaSummary.STTProfile,
			STTProvider:   in.ProviderMetaSummary.STTProvider,
			STTLatencyMs:  in.ProviderMetaSummary.STTLatencyMs,
			ChatProfile:   in.ProviderMetaSummary.ChatProfile,
			ChatProvider:  in.ProviderMetaSummary.ChatProvider,
			ChatLatencyMs: in.ProviderMetaSummary.ChatLatencyMs,
			TTSProfile:    in.ProviderMetaSummary.TTSProfile,
			TTSProvider:   in.ProviderMetaSummary.TTSProvider,
			TTSLatencyMs:  in.ProviderMetaSummary.TTSLatencyMs,
		},
	}
	for _, chunk := range in.TTSChunks {
		payload.TTSChunks = append(payload.TTSChunks, practiceVoiceTTSChunkPayload{
			ChunkID:     chunk.ChunkID,
			Sequence:    chunk.Sequence,
			ContentType: chunk.ContentType,
			DurationMs:  chunk.DurationMs,
			ByteLength:  chunk.ByteLength,
			TextHash:    chunk.TextHash,
			AudioRef:    chunk.AudioRef,
		})
	}
	if in.TTSError != nil {
		payload.TTSError = &practiceVoiceTTSErrorPayload{
			Code:      in.TTSError.Code,
			Message:   in.TTSError.Message,
			Retryable: in.TTSError.Retryable,
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal practice voice turn event payload: %w", err)
	}
	return raw, nil
}
