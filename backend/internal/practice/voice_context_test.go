package practice

import (
	"testing"
	"time"
)

func TestBuildCommittedVoiceContextCompleteChunk(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "请继续说明高风险团队试点。",
		AssistantTextHash:   textSHA256("请继续说明高风险团队试点。"),
		AssistantTextLength: int32(len([]rune("请继续说明高风险团队试点。"))),
	}
	events := []VoicePlaybackEventRecord{
		voiceEvent(sessionEventKindTTSChunkStarted, 0, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playbackOffsetMs": 0}),
		voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": source.AssistantTextLength, "playbackOffsetMs": 2840}),
		voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": source.AssistantTextLength, "playbackOffsetMs": 2840}),
	}

	got := BuildCommittedVoiceContext(source, events)
	if got.CommittedAssistantText != source.AssistantTextDraft ||
		got.CommittedTextHash != source.AssistantTextHash ||
		got.CommittedTextLength != source.AssistantTextLength ||
		got.Interrupted {
		t.Fatalf("complete committed context drift: %+v", got)
	}
}

func TestBuildCommittedVoiceContextPartialBargeIn(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "请继续说明高风险团队试点。",
		AssistantTextHash:   textSHA256("请继续说明高风险团队试点。"),
		AssistantTextLength: int32(len([]rune("请继续说明高风险团队试点。"))),
	}
	events := []VoicePlaybackEventRecord{
		voiceEvent(sessionEventKindTTSChunkStarted, 0, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playbackOffsetMs": 0}),
		voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 4, "playbackOffsetMs": 1100}),
		voiceEvent(sessionEventKindBargeInDetected, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playbackOffsetMs": 1480, "userSpeechStartedAt": "2026-05-17T08:51:05Z"}),
	}

	got := BuildCommittedVoiceContext(source, events)
	if got.CommittedAssistantText != "请继续说" ||
		got.CommittedTextLength != 4 ||
		!got.Interrupted ||
		got.InterruptionOffsetMs != 1480 ||
		got.InterruptionNote == "" {
		t.Fatalf("partial committed context drift: %+v", got)
	}
}

func TestBuildCommittedVoiceContextNoPlayback(t *testing.T) {
	got := BuildCommittedVoiceContext(PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "unplayed draft",
		AssistantTextLength: 13,
	}, nil)
	if got.HasCommittedContext || got.CommittedAssistantText != "" || got.InterruptionNote != "" {
		t.Fatalf("no playback must not commit draft text: %+v", got)
	}
}

func TestBuildCommittedVoiceContextDeduplicatesAndSortsEvents(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "abcdef",
		AssistantTextHash:   textSHA256("abcdef"),
		AssistantTextLength: 6,
	}
	events := []VoicePlaybackEventRecord{
		voiceEvent(sessionEventKindContextCommitted, 3, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 6, "playbackOffsetMs": 600}),
		voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 3, "playbackOffsetMs": 300}),
		voiceEvent(sessionEventKindTTSChunkPlayed, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 6, "playbackOffsetMs": 600}),
	}

	got := BuildCommittedVoiceContext(source, events)
	if got.CommittedAssistantText != "abcdef" ||
		got.CommittedTextLength != 6 ||
		got.PlaybackOffsetMs != 600 {
		t.Fatalf("out-of-order/duplicate context drift: %+v", got)
	}
}

func TestBuildCommittedVoiceContextRequiresMatchingPlayedEvidence(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "abcdef",
		AssistantTextHash:   textSHA256("abcdef"),
		AssistantTextLength: 6,
	}
	tests := []struct {
		name   string
		events []VoicePlaybackEventRecord
	}{
		{
			name: "committed only",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindContextCommitted, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 3, "playbackOffsetMs": 300}),
			},
		},
		{
			name: "commit precedes played evidence",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindContextCommitted, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 3, "playbackOffsetMs": 300}),
				voiceEvent(sessionEventKindTTSChunkPlayed, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 3, "playbackOffsetMs": 300}),
			},
		},
		{
			name: "chunk does not match",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 4, "playbackOffsetMs": 400}),
				voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-2", "committedTextHash": source.AssistantTextHash, "committedTextLength": 4, "playbackOffsetMs": 400}),
			},
		},
		{
			name: "commit exceeds played length",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 3, "playbackOffsetMs": 300}),
				voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 4, "playbackOffsetMs": 400}),
			},
		},
		{
			name: "played hash does not match source",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": textSHA256("different"), "playedTextLength": 4, "playbackOffsetMs": 400}),
				voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 4, "playbackOffsetMs": 400}),
			},
		},
		{
			name: "committed hash does not match source",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 4, "playbackOffsetMs": 400}),
				voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": textSHA256("different"), "committedTextLength": 4, "playbackOffsetMs": 400}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCommittedVoiceContext(source, tt.events)
			if got.HasCommittedContext || got.CommittedAssistantText != "" || got.CommittedTextLength != 0 {
				t.Fatalf("unproven playback must not commit assistant context: %+v", got)
			}
		})
	}
}

func TestBuildCommittedVoiceContextAllowsCommitWithinMatchingPlayedLength(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "abcdef",
		AssistantTextHash:   textSHA256("abcdef"),
		AssistantTextLength: 6,
	}
	events := []VoicePlaybackEventRecord{
		voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 5, "playbackOffsetMs": 500}),
		voiceEvent(sessionEventKindContextCommitted, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "committedTextHash": source.AssistantTextHash, "committedTextLength": 3, "playbackOffsetMs": 300}),
	}

	got := BuildCommittedVoiceContext(source, events)
	if !got.HasCommittedContext || got.CommittedAssistantText != "abc" || got.CommittedTextLength != 3 {
		t.Fatalf("matching played evidence should allow a bounded commit: %+v", got)
	}
}

func TestBuildCommittedVoiceContextRequiresPlayedEvidenceBeforeMatchingBargeIn(t *testing.T) {
	source := PracticeVoiceTurnContextSource{
		VoiceTurnID:         "voice-turn-1",
		AssistantTextDraft:  "abcdef",
		AssistantTextHash:   textSHA256("abcdef"),
		AssistantTextLength: 6,
	}
	tests := []struct {
		name   string
		events []VoicePlaybackEventRecord
	}{
		{
			name: "barge in precedes played evidence",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindBargeInDetected, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playbackOffsetMs": 300, "userSpeechStartedAt": "2026-05-17T08:51:01Z"}),
				voiceEvent(sessionEventKindTTSChunkPlayed, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 3, "playbackOffsetMs": 300}),
			},
		},
		{
			name: "barge in chunk does not match",
			events: []VoicePlaybackEventRecord{
				voiceEvent(sessionEventKindTTSChunkPlayed, 1, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-1", "playedTextHash": source.AssistantTextHash, "playedTextLength": 3, "playbackOffsetMs": 300}),
				voiceEvent(sessionEventKindBargeInDetected, 2, map[string]any{"voiceTurnId": "voice-turn-1", "chunkId": "chunk-2", "playbackOffsetMs": 300, "userSpeechStartedAt": "2026-05-17T08:51:02Z"}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCommittedVoiceContext(source, tt.events)
			if got.HasCommittedContext || got.CommittedAssistantText != "" {
				t.Fatalf("barge-in without prior matching playback must not commit context: %+v", got)
			}
		})
	}
}

func voiceEvent(kind string, seconds int, payload map[string]any) VoicePlaybackEventRecord {
	return VoicePlaybackEventRecord{
		Kind:       kind,
		OccurredAt: time.Date(2026, 5, 17, 8, 51, seconds, 0, time.UTC),
		Payload:    payload,
	}
}
