package practice

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLRepositoryLoadCommittedVoiceContextBuildsFromLatestVoiceTurnEvents(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()
	repo := NewSQLRepository(db)
	now := time.Date(2026, 5, 17, 9, 30, 0, 0, time.UTC)

	mock.ExpectQuery(`(?s)from practice_session_events e.*event_type = 'follow_up_generated'.*order by e.seq_no desc`).
		WithArgs("user-1", "session-1").
		WillReturnRows(sqlmock.NewRows([]string{"seq_no", "payload"}).AddRow(7, []byte(`{
			"voiceTurnId":"voice-turn-1",
			"assistantTextDraft":"Please expand on your migration validation.",
				"ttsChunks":[{"chunkId":"chunk-1","textHash":"72872d4a414ea78752a0794a760c3afb6042a545fd0bef6328ad98ef0b52ea49","audioRef":"voice-turn://voice-turn-1/chunks/chunk-1"}]
			}`)))
	mock.ExpectQuery(`(?s)from practice_session_events.*seq_no > \$2.*event_type in`).
		WithArgs("session-1", 7).
		WillReturnRows(sqlmock.NewRows([]string{"event_type", "payload", "created_at"}).
			AddRow("tts_chunk_played", []byte(`{"requestPayload":{"voiceTurnId":"voice-turn-1","chunkId":"chunk-1","playedTextLength":13,"playedTextHash":"72872d4a414ea78752a0794a760c3afb6042a545fd0bef6328ad98ef0b52ea49","playbackOffsetMs":1480}}`), now).
			AddRow("barge_in_detected", []byte(`{"requestPayload":{"voiceTurnId":"voice-turn-1","chunkId":"chunk-1","playbackOffsetMs":1480,"userSpeechStartedAt":"2026-05-17T09:30:01Z"}}`), now.Add(time.Second)))

	got, err := repo.LoadCommittedVoiceContext(context.Background(), "user-1", "session-1")
	if err != nil {
		t.Fatalf("LoadCommittedVoiceContext: %v", err)
	}
	if !got.HasCommittedContext || got.CommittedAssistantText != "Please expand" || !got.Interrupted {
		t.Fatalf("committed context drifted: %+v", got)
	}
	if got.InterruptionNote == "" {
		t.Fatalf("expected interruption note, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
