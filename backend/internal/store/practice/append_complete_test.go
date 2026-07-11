package practice

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSQLRepositoryAppendSessionEventWritesEventTurnSessionOutboxWithoutAudit(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.AppendSessionEventStoreInput{
		EventID:            "event-1",
		OutboxEventID:      "outbox-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		OccurredAt:         now,
		RequestFingerprint: "fingerprint-1",
		RequestPayload: map[string]any{
			"answerText":    "answer",
			"followUpCount": 99,
		},
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusCompleted,
			AnswerSummary:     "Candidate explained the service boundary and rollback path.",
			NextTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusAssessed),
				FollowUpCount:  1,
				AskedAt:        now.Add(-time.Minute),
			},
			AssistantAction: domain.AssistantActionRecord{
				Type:          "session_completed",
				SessionStatus: sharedtypes.SessionStatusCompleted,
				Provenance:    domain.AssistantActionProvenance{PromptVersion: "not_applicable", RubricVersion: "not_applicable", ModelID: "model-profile:static", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "static"},
			},
			OutboxRecord: &domain.PracticeTurnCompletedRecord{
				SessionID:        "session-1",
				TurnID:           "turn-1",
				FollowUpCount:    1,
				AnswerCharLength: 6,
				CompletedAt:      now,
			},
		},
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "replay_payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","pending":true}`), nil))
	mock.ExpectExec(`update practice_turns`).
		WithArgs(string(domain.TurnStatusAssessed), nil, nil, "answer", "Candidate explained the service boundary and rollback path.", 1, now, now, now, in.SessionID, "turn-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusCompleted), int32(1), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeTurnCompleted), "turn-1", sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_session_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), in.SessionID, in.ClientEventID, in.EventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.AppendSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if !result.Acknowledged || result.Session.Status != sharedtypes.SessionStatusCompleted {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryAppendSessionEventPersistsGeneratedFollowUpOnCurrentTurn(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.AppendSessionEventStoreInput{
		EventID:            "event-1",
		OutboxEventID:      "outbox-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		OccurredAt:         now,
		RequestFingerprint: "fingerprint-1",
		RequestPayload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "answer",
		},
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusRunning,
			AnswerSummary:     "Candidate explained the migration impact.",
			NextTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusFollowUpRequested),
				FollowUpCount:  1,
				AskedAt:        now.Add(-time.Minute),
			},
			AssistantAction: domain.AssistantActionRecord{
				Type:           "ask_follow_up",
				TurnID:         "turn-1",
				QuestionText:   "服务器生成的追问是什么？",
				QuestionIntent: "evidence.follow_up",
				SessionStatus:  sharedtypes.SessionStatusRunning,
				Provenance:     domain.AssistantActionProvenance{PromptVersion: "p", RubricVersion: "r", ModelID: "model-profile:followup", Language: "zh-CN", FeatureFlag: "none", DataSourceVersion: "registry.v1"},
			},
		},
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "replay_payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","pending":true}`), nil))
	mock.ExpectExec(`update practice_turns\s+set status = \$1,\s+question_text = coalesce\(\$2, question_text\),\s+question_intent = coalesce\(\$3, question_intent\)`).
		WithArgs(
			string(domain.TurnStatusFollowUpRequested),
			"服务器生成的追问是什么？",
			"evidence.follow_up",
			"answer",
			"Candidate explained the migration impact.",
			1,
			now,
			nil,
			now,
			in.SessionID,
			"turn-1",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusRunning), int32(1), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_session_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), in.SessionID, in.ClientEventID, in.EventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.AppendSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.Session.CurrentTurn == nil ||
		result.Session.CurrentTurn.QuestionText != "服务器生成的追问是什么？" ||
		result.Session.CurrentTurn.QuestionIntent != "evidence.follow_up" {
		t.Fatalf("generated follow-up was not reflected in the current turn: %+v", result.Session.CurrentTurn)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryAppendSessionEventWritesHintTextForAssistedSuccess(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	in := domain.AppendSessionEventStoreInput{
		EventID:            "event-1",
		OutboxEventID:      "outbox-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "hint_requested",
		OccurredAt:         now,
		RequestFingerprint: "fingerprint-1",
		RequestPayload:     map[string]any{"turnId": "turn-1"},
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusRunning,
			AssistantAction: domain.AssistantActionRecord{
				Type:          "show_hint",
				TurnID:        "turn-1",
				Hint:          "Use one measurable tradeoff.",
				SessionStatus: sharedtypes.SessionStatusRunning,
				Provenance:    domain.AssistantActionProvenance{PromptVersion: "p", RubricVersion: "not_applicable", ModelID: "model-profile:practice.turn_observe.default", Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1"},
			},
		},
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "replay_payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","pending":true}`), nil))
	mock.ExpectExec(`update practice_turns\s+set hint_text`).
		WithArgs("Use one measurable tradeoff.", now, in.SessionID, "turn-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusRunning), int32(1), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_session_events`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), in.SessionID, in.ClientEventID, in.EventID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.AppendSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("AppendSessionEvent returned error: %v", err)
	}
	if result.AssistantAction.Hint != "Use one measurable tradeoff." || result.Session.TurnCount != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryRecordPracticeVoiceTurnWritesBusinessEventWithoutAudioBytes(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 17, 9, 12, 0, 0, time.UTC)
	in := domain.PracticeVoiceTurnStoreInput{
		EventID:             "event-voice-1",
		UserID:              "user-1",
		SessionID:           "session-1",
		ClientVoiceTurnID:   "client-voice-turn-1",
		TurnID:              "turn-1",
		VoiceTurnID:         "voice-turn-1",
		UserTranscriptFinal: "candidate transcript persisted as business text",
		AssistantTextDraft:  "assistant text persisted as business text",
		AudioByteLength:     23,
		AudioDurationMs:     900,
		TTSChunks: []domain.PracticeVoiceTTSChunk{{
			ChunkID:     "chunk-1",
			Sequence:    1,
			ContentType: "audio/mpeg",
			DurationMs:  880,
			ByteLength:  9,
			TextHash:    "sha256:text",
			AudioRef:    "voice-turn://voice-turn-1/chunks/chunk-1",
		}},
		ProviderMetaSummary: domain.PracticeVoiceProviderMetaSummary{
			STTProfile:   "practice.voice.stt.default",
			STTProvider:  "fixture-stt",
			ChatProfile:  "practice.followup.default",
			ChatProvider: "fixture-chat",
			TTSProfile:   "practice.voice.tts.default",
			TTSProvider:  "fixture-tts",
		},
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusRunning,
			AnswerSummary:     "Candidate explained the migration goal, validation, and rollback boundary.",
			NextTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusFollowUpRequested),
				FollowUpCount:  1,
			},
			AssistantAction: domain.AssistantActionRecord{
				Type:           "ask_follow_up",
				TurnID:         "turn-1",
				QuestionText:   "assistant text persisted as business text",
				QuestionIntent: "evidence.follow_up",
				SessionStatus:  sharedtypes.SessionStatusRunning,
			},
		},
		Session:    domain.SessionRecord{ID: "session-1", Status: sharedtypes.SessionStatusRunning},
		OccurredAt: now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"next_seq"}).AddRow(4))
	mock.ExpectExec(`update practice_turns\s+set status = \$1,\s+question_text = coalesce\(\$2, question_text\),\s+question_intent = coalesce\(\$3, question_intent\)`).
		WithArgs(
			string(domain.TurnStatusFollowUpRequested),
			in.AssistantTextDraft,
			"evidence.follow_up",
			in.UserTranscriptFinal,
			in.Outcome.AnswerSummary,
			1,
			now,
			nil,
			now,
			in.SessionID,
			in.TurnID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusRunning), int32(1), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.EventID, in.SessionID, 4, "follow_up_generated", in.ClientVoiceTurnID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	session, err := repo.RecordPracticeVoiceTurn(context.Background(), in)
	if err != nil {
		t.Fatalf("RecordPracticeVoiceTurn returned error: %v", err)
	}
	if session.CurrentTurn == nil ||
		session.CurrentTurn.Status != string(domain.TurnStatusFollowUpRequested) ||
		session.CurrentTurn.QuestionText != in.AssistantTextDraft ||
		session.CurrentTurn.QuestionIntent != "evidence.follow_up" {
		t.Fatalf("voice turn should persist generated same-turn follow-up: %+v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}

	payload, err := marshalPracticeVoiceTurnEventPayload(in)
	if err != nil {
		t.Fatalf("marshalPracticeVoiceTurnEventPayload: %v", err)
	}
	payloadText := string(payload)
	for _, want := range []string{in.UserTranscriptFinal, in.AssistantTextDraft, `"audioByteLength":23`, `"textHash":"sha256:text"`} {
		if !strings.Contains(payloadText, want) {
			t.Fatalf("voice event payload missing %q: %s", want, payloadText)
		}
	}
	for _, forbidden := range []string{"raw-audio-privacy-token", "tts-audio-privacy-token"} {
		if strings.Contains(payloadText, forbidden) {
			t.Fatalf("voice event payload leaked forbidden bytes %q: %s", forbidden, payloadText)
		}
	}
}

func TestSQLRepositoryRecordPracticeVoiceTurnAdvancesNextQuestionAndWritesCompletionOutboxAtomically(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 17, 9, 15, 0, 0, time.UTC)
	in := domain.PracticeVoiceTurnStoreInput{
		EventID:             "event-voice-2",
		OutboxEventID:       "outbox-voice-2",
		UserID:              "user-1",
		SessionID:           "session-1",
		ClientVoiceTurnID:   "client-voice-turn-2",
		TurnID:              "turn-1",
		VoiceTurnID:         "voice-turn-2",
		UserTranscriptFinal: "我用双写校验和回放验证了拆分结果。",
		AssistantTextDraft:  "请再介绍一个你主导的系统取舍案例。",
		AudioByteLength:     24,
		AudioDurationMs:     920,
		Outcome: domain.SessionEventOutcome{
			Acknowledged:      true,
			NextSessionStatus: sharedtypes.SessionStatusRunning,
			NextTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusAssessed),
				FollowUpCount:  1,
			},
			AssistantAction: domain.AssistantActionRecord{
				Type:           "ask_question",
				TurnID:         "turn-2",
				QuestionText:   "请再介绍一个你主导的系统取舍案例。",
				QuestionIntent: "system_design.tradeoff",
				SessionStatus:  sharedtypes.SessionStatusRunning,
			},
			OutboxRecord: &domain.PracticeTurnCompletedRecord{
				SessionID:        "session-1",
				TurnID:           "turn-1",
				FollowUpCount:    1,
				AnswerCharLength: 18,
				CompletedAt:      now,
			},
		},
		NextQuestion: &domain.TurnRecord{
			ID:             "turn-2",
			TurnIndex:      2,
			QuestionText:   "请再介绍一个你主导的系统取舍案例。",
			QuestionIntent: "system_design.tradeoff",
			Status:         string(domain.TurnStatusAsked),
			AskedAt:        now,
		},
		Session:    domain.SessionRecord{ID: "session-1", Status: sharedtypes.SessionStatusRunning},
		OccurredAt: now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"next_seq"}).AddRow(5))
	mock.ExpectExec(`update practice_turns`).
		WithArgs(
			string(domain.TurnStatusAssessed),
			nil,
			nil,
			in.UserTranscriptFinal,
			nil,
			1,
			now,
			now,
			now,
			in.SessionID,
			in.TurnID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into practice_turns`).
		WithArgs("turn-2", in.SessionID, int32(2), in.AssistantTextDraft, "system_design.tradeoff", string(sharedtypes.InterviewerRoleHiringManager), string(domain.TurnStatusAsked), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusRunning), int32(2), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeTurnCompleted), in.TurnID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.EventID, in.SessionID, 5, "follow_up_generated", in.ClientVoiceTurnID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	session, err := repo.RecordPracticeVoiceTurn(context.Background(), in)
	if err != nil {
		t.Fatalf("RecordPracticeVoiceTurn returned error: %v", err)
	}
	if session.TurnCount != 2 || session.CurrentTurn == nil || session.CurrentTurn.ID != "turn-2" || session.CurrentTurn.QuestionText != in.AssistantTextDraft {
		t.Fatalf("voice turn did not atomically advance current turn: %+v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMarshalAppendEventErrorPayloadSanitizesRequestPayload(t *testing.T) {
	raw, err := marshalAppendEventErrorPayload(domain.FinalizeSessionEventErrorInput{
		RequestFingerprint: "fingerprint-1",
		RequestPayload: map[string]any{
			"turnId":           "turn-secret",
			"answerText":       "confidential answer",
			"playedTextHash":   "I led the migration across twelve teams.",
			"voiceTurnId":      "voice-turn-secret",
			"playbackOffsetMs": 2840,
		},
		Error: &domain.ServiceError{
			Code:    sharederrors.CodeValidationFailed,
			Message: "voice playback event payload field is invalid",
			Details: map[string]any{
				"field": "payload.playedTextHash",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshalAppendEventErrorPayload returned error: %v", err)
	}
	payload := string(raw)
	for _, forbidden := range []string{"requestPayload", "turn-secret", "confidential answer", "I led the migration across twelve teams.", "voice-turn-secret", `"result"`} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("payload leaked %q: %s", forbidden, payload)
		}
	}

	var decoded appendEventPayload
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode sanitized payload: %v", err)
	}
	if decoded.RequestFingerprint != "fingerprint-1" {
		t.Fatalf("requestFingerprint = %q", decoded.RequestFingerprint)
	}
	if decoded.Error == nil || decoded.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("unexpected error envelope: %+v", decoded.Error)
	}
	if decoded.Error.Details["field"] != "payload.playedTextHash" {
		t.Fatalf("unexpected error details: %+v", decoded.Error.Details)
	}
}

func TestMarshalAppendEventPayloadRedactsHintButReplayPayloadKeepsSnapshot(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	result := domain.AppendSessionEventResult{
		Acknowledged: true,
		Session: domain.SessionRecord{
			ID:           "session-1",
			PlanID:       "plan-1",
			TargetJobID:  "target-1",
			Status:       sharedtypes.SessionStatusRunning,
			Language:     "en",
			HintsEnabled: true,
			TurnCount:    1,
			CurrentTurn: &domain.TurnRecord{
				ID:             "turn-1",
				TurnIndex:      1,
				QuestionText:   "Question?",
				QuestionIntent: "behavioral",
				Status:         string(domain.TurnStatusAsked),
				AskedAt:        now.Add(-time.Minute),
			},
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now,
		},
		AssistantAction: domain.AssistantActionRecord{
			Type:          "show_hint",
			TurnID:        "turn-1",
			Hint:          "Original per-event hint.",
			SessionStatus: sharedtypes.SessionStatusRunning,
			Provenance:    domain.AssistantActionProvenance{PromptVersion: "p", RubricVersion: "not_applicable", ModelID: "model-profile:practice.turn_observe.default", Language: "en", FeatureFlag: "none", DataSourceVersion: "registry.v1"},
		},
	}
	in := domain.AppendSessionEventStoreInput{
		RequestFingerprint: "fingerprint-1",
		RequestPayload: map[string]any{
			"turnId":     "turn-1",
			"answerText": "confidential answer body",
		},
	}

	eventPayload, err := marshalAppendEventPayload(in, result)
	if err != nil {
		t.Fatalf("marshalAppendEventPayload returned error: %v", err)
	}
	replayPayload, err := marshalAppendEventReplayPayload(in.RequestFingerprint, result)
	if err != nil {
		t.Fatalf("marshalAppendEventReplayPayload returned error: %v", err)
	}
	if strings.Contains(string(eventPayload), "Original per-event hint.") {
		t.Fatalf("event payload leaked hint snapshot: %s", eventPayload)
	}
	if strings.Contains(string(eventPayload), "confidential answer body") || !strings.Contains(string(eventPayload), `"turnId":"turn-1"`) {
		t.Fatalf("event payload must keep routing IDs without duplicating answer text: %s", eventPayload)
	}
	if !strings.Contains(string(replayPayload), "Original per-event hint.") {
		t.Fatalf("replay payload lost hint snapshot: %s", replayPayload)
	}
}

func TestMarshalAppendEventPayloadScopesVoiceMetadataToEventKind(t *testing.T) {
	result := domain.AppendSessionEventResult{
		Acknowledged: true,
		Session: domain.SessionRecord{
			ID: "session-1", Status: sharedtypes.SessionStatusRunning,
		},
		AssistantAction: domain.AssistantActionRecord{
			Type: "session_wait", SessionStatus: sharedtypes.SessionStatusRunning,
		},
	}
	allMetadata := map[string]any{
		"turnId":              "turn-1",
		"voiceTurnId":         "voice-turn-1",
		"chunkId":             "chunk-1",
		"playedTextHash":      "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"playedTextLength":    12,
		"playbackOffsetMs":    1200,
		"userSpeechStartedAt": "2026-05-17T08:51:05Z",
		"committedTextHash":   "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"committedTextLength": 8,
	}

	answerRaw, err := marshalAppendEventPayload(domain.AppendSessionEventStoreInput{
		Kind: "answer_submitted", RequestFingerprint: "answer-fingerprint", RequestPayload: allMetadata,
	}, result)
	if err != nil {
		t.Fatalf("marshal answer event payload: %v", err)
	}
	var answer appendEventPayload
	if err := json.Unmarshal(answerRaw, &answer); err != nil {
		t.Fatalf("decode answer event payload: %v", err)
	}
	if len(answer.RequestPayload) != 1 || answer.RequestPayload["turnId"] != "turn-1" {
		t.Fatalf("answer event persisted cross-kind voice metadata: %+v", answer.RequestPayload)
	}

	playedRaw, err := marshalAppendEventPayload(domain.AppendSessionEventStoreInput{
		Kind: "tts_chunk_played", RequestFingerprint: "played-fingerprint", RequestPayload: allMetadata,
	}, result)
	if err != nil {
		t.Fatalf("marshal played event payload: %v", err)
	}
	var played appendEventPayload
	if err := json.Unmarshal(playedRaw, &played); err != nil {
		t.Fatalf("decode played event payload: %v", err)
	}
	for _, key := range []string{"voiceTurnId", "chunkId", "playedTextHash", "playedTextLength", "playbackOffsetMs"} {
		if _, ok := played.RequestPayload[key]; !ok {
			t.Fatalf("played event lost allowed metadata %q: %+v", key, played.RequestPayload)
		}
	}
	for _, key := range []string{"turnId", "userSpeechStartedAt", "committedTextHash", "committedTextLength"} {
		if _, ok := played.RequestPayload[key]; ok {
			t.Fatalf("played event persisted cross-kind metadata %q: %+v", key, played.RequestPayload)
		}
	}
}

func TestSQLRepositoryReserveSessionEventReplaysOriginalHintSnapshot(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 47, 32, 0, time.UTC)
	in := domain.SessionEventReservationInput{
		EventID:            "event-2",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "hint_requested",
		CurrentTurnID:      "turn-1",
		RequestFingerprint: "fingerprint-1",
		Now:                now,
	}
	eventPayload := []byte(`{"requestFingerprint":"fingerprint-1","requestPayload":{"turnId":"turn-1"},"result":{"acknowledged":true,"session":{"id":"session-1","planId":"plan-1","targetJobId":"target-1","status":"running","language":"en","hintsEnabled":true,"turnCount":1,"currentTurn":{"id":"turn-1","turnIndex":1,"questionText":"Question?","questionIntent":"behavioral","status":"asked","askedAt":"2026-04-28T13:46:32Z"},"createdAt":"2026-04-28T12:47:32Z","updatedAt":"2026-04-28T13:47:32Z"},"assistantAction":{"type":"show_hint","turnId":"turn-1","sessionStatus":"running","provenance":{"promptVersion":"p","rubricVersion":"not_applicable","modelId":"model-profile:practice.turn_observe.default","language":"en","featureFlag":"none","dataSourceVersion":"registry.v1"}}}}`)
	replayPayload := []byte(`{"requestFingerprint":"fingerprint-1","result":{"acknowledged":true,"session":{"id":"session-1","planId":"plan-1","targetJobId":"target-1","status":"running","language":"en","hintsEnabled":true,"turnCount":1,"currentTurn":{"id":"turn-1","turnIndex":1,"questionText":"Question?","questionIntent":"behavioral","status":"asked","askedAt":"2026-04-28T13:46:32Z"},"createdAt":"2026-04-28T12:47:32Z","updatedAt":"2026-04-28T13:47:32Z"},"assistantAction":{"type":"show_hint","turnId":"turn-1","hint":"Original per-event hint.","sessionStatus":"running","provenance":{"promptVersion":"p","rubricVersion":"not_applicable","modelId":"model-profile:practice.turn_observe.default","language":"en","featureFlag":"none","dataSourceVersion":"registry.v1"}}}}`)

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload, replay_payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "replay_payload"}).AddRow(eventPayload, replayPayload))
	mock.ExpectCommit()

	result, err := repo.ReserveSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionEvent returned error: %v", err)
	}
	if result.ReplayResult == nil || result.ReplayResult.AssistantAction.Hint != "Original per-event hint." {
		t.Fatalf("replay should use stored per-event hint snapshot, got %+v", result.ReplayResult)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionEventCreatesPendingReservation(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.SessionEventReservationInput{
		EventID:            "event-1",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		RequestFingerprint: "fingerprint-1",
		Now:                now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(2))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.EventID, in.SessionID, 2, in.Kind, in.ClientEventID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.ReserveSessionEvent(context.Background(), in)
	if err != nil {
		t.Fatalf("ReserveSessionEvent returned error: %v", err)
	}
	if result.ReplayResult != nil || result.Session.ID != in.SessionID || result.LatestTurn.ID != "turn-1" {
		t.Fatalf("unexpected reservation: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryReserveSessionEventRejectsPendingReservationReplay(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	in := domain.SessionEventReservationInput{
		EventID:            "event-2",
		UserID:             "user-1",
		SessionID:          "session-1",
		ClientEventID:      "client-event-1",
		Kind:               "answer_submitted",
		RequestFingerprint: "fingerprint-1",
		Now:                now,
	}

	mock.ExpectBegin()
	expectAppendContext(mock, now)
	mock.ExpectQuery(`select payload`).
		WithArgs(in.SessionID, in.ClientEventID).
		WillReturnRows(sqlmock.NewRows([]string{"payload", "replay_payload"}).AddRow([]byte(`{"requestFingerprint":"fingerprint-1","requestPayload":{"answerText":"answer"},"pending":true}`), nil))
	mock.ExpectRollback()

	_, err := repo.ReserveSessionEvent(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionWritesReportJobOutboxAndAudit(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-1",
		JobID:             "job-1",
		SessionEventID:    "event-1",
		OutboxEventID:     "outbox-1",
		AuditEventID:      "audit-1",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`update practice_sessions`).
		WithArgs(string(sharedtypes.SessionStatusCompleting), now, in.SessionID, in.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`select coalesce\(max\(seq_no\), 0\) \+ 1`).
		WithArgs(in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"seq_no"}).AddRow(4))
	mock.ExpectExec(`insert into practice_session_events`).
		WithArgs(in.SessionEventID, in.SessionID, 4, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into feedback_reports`).
		WithArgs(in.ReportID, in.UserID, in.SessionID, "target-1", string(sharedtypes.ReportStatusQueued), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into async_jobs`).
		WithArgs(in.JobID, "report_generate", in.ReportID, in.SessionID, string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into outbox_events`).
		WithArgs(in.OutboxEventID, string(sharedevents.EventNamePracticeSessionCompleted), in.SessionID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`insert into audit_events`).
		WithArgs(in.AuditEventID, in.UserID, in.UserID, in.SessionID, sqlmock.AnyArg(), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession returned error: %v", err)
	}
	if result.ReportID != in.ReportID || result.Job.ID != in.JobID || result.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionRejectsIllegalStatusWithoutReport(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-1",
		JobID:             "job-1",
		SessionEventID:    "event-1",
		OutboxEventID:     "outbox-1",
		AuditEventID:      "audit-1",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusFailed), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`select fr.id`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.CompleteSession(context.Background(), in)
	if !errors.Is(err, domain.ErrSessionConflict) {
		t.Fatalf("error = %v, want ErrSessionConflict", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSQLRepositoryCompleteSessionReplaysExistingReportBeforeStatusGuard(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC)
	in := domain.CompleteSessionStoreInput{
		UserID:            "user-1",
		SessionID:         "session-1",
		ReportID:          "report-new",
		JobID:             "job-new",
		SessionEventID:    "event-new",
		OutboxEventID:     "outbox-new",
		AuditEventID:      "audit-new",
		ClientCompletedAt: now,
		Now:               now,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`select id, plan_id, target_job_id, status, language, hints_enabled`).
		WithArgs(in.UserID, in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "target_job_id", "status", "language", "hints_enabled", "turn_count", "created_at", "updated_at"}).
			AddRow(in.SessionID, "plan-1", "target-1", string(sharedtypes.SessionStatusFailed), "zh-CN", true, 3, now.Add(-time.Hour), now.Add(-time.Minute)))
	mock.ExpectQuery(`j\.dedupe_key = \$4`).
		WithArgs(in.UserID, in.SessionID, "report_generate", in.SessionID).
		WillReturnRows(sqlmock.NewRows([]string{"report_id", "job_id", "job_type", "resource_type", "resource_id", "status", "error_code", "created_at", "updated_at"}).
			AddRow("report-existing", "job-existing", "report_generate", "feedback_report", "report-existing", string(sharedtypes.JobStatusQueued), nil, now.Add(-time.Minute), now.Add(-time.Minute)))
	mock.ExpectCommit()

	result, err := repo.CompleteSession(context.Background(), in)
	if err != nil {
		t.Fatalf("CompleteSession returned error: %v", err)
	}
	if !result.Replay || result.ReportID != "report-existing" || result.Job.ID != "job-existing" {
		t.Fatalf("unexpected replay result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCanCompletePracticeSessionStatusAllowsRunningWaitingAndCompleted(t *testing.T) {
	allowed := map[sharedtypes.SessionStatus]bool{
		sharedtypes.SessionStatusRunning:          true,
		sharedtypes.SessionStatusWaitingUserInput: true,
		sharedtypes.SessionStatusCompleted:        true,
	}
	for _, status := range sharedtypes.AllSessionStatuses {
		if got := canCompletePracticeSessionStatus(status); got != allowed[status] {
			t.Fatalf("canCompletePracticeSessionStatus(%q) = %v, want %v", status, got, allowed[status])
		}
	}
}

func expectAppendContext(mock sqlmock.Sqlmock, now time.Time) {
	mock.ExpectQuery(`select s.id, s.plan_id`).
		WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "plan_id", "target_job_id", "status", "language", "hints_enabled",
			"turn_count", "created_at", "updated_at",
			"id", "target_job_id", "goal", "mode", "interviewer_persona", "difficulty",
			"language", "time_budget_minutes", "question_budget", "status", "created_at",
			"id", "turn_index", "question_text", "question_intent", "status", "follow_up_count", "asked_at",
		}).AddRow(
			"session-1", "plan-1", "target-1", string(sharedtypes.SessionStatusRunning), "zh-CN", true,
			1, now.Add(-time.Hour), now.Add(-time.Minute),
			"plan-1", "target-1", string(sharedtypes.PracticeGoalBaseline), string(sharedtypes.PracticeModeAssisted), string(sharedtypes.InterviewerRoleHiringManager), "standard",
			"zh-CN", 30, 3, "ready", now.Add(-2*time.Hour),
			"turn-1", 1, "Question?", "behavioral", string(domain.TurnStatusAsked), 1, now.Add(-time.Minute),
		))
}
