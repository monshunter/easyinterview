package practice

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type PracticeVoiceTurnContextSource struct {
	VoiceTurnID         string
	AssistantTextDraft  string
	AssistantTextHash   string
	AssistantTextLength int32
}

type VoicePlaybackEventRecord struct {
	Kind       string
	OccurredAt time.Time
	Payload    map[string]any
}

type CommittedVoiceContext struct {
	VoiceTurnID            string
	HasCommittedContext    bool
	CommittedAssistantText string
	CommittedTextHash      string
	CommittedTextLength    int32
	PlaybackOffsetMs       int64
	Interrupted            bool
	InterruptionOffsetMs   int64
	UserSpeechStartedAt    string
	InterruptionNote       string
}

func BuildCommittedVoiceContext(source PracticeVoiceTurnContextSource, events []VoicePlaybackEventRecord) CommittedVoiceContext {
	voiceTurnID := strings.TrimSpace(source.VoiceTurnID)
	out := CommittedVoiceContext{VoiceTurnID: voiceTurnID}
	if voiceTurnID == "" || len(events) == 0 {
		return out
	}
	ordered := append([]VoicePlaybackEventRecord(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].OccurredAt.Before(ordered[j].OccurredAt)
	})

	var playedLength int32
	var playedHash string
	var playedOffset int64
	var committedLength int32
	var committedHash string
	var committedOffset int64

	for _, event := range ordered {
		if strings.TrimSpace(payloadString(event.Payload, "voiceTurnId")) != voiceTurnID {
			continue
		}
		switch strings.TrimSpace(event.Kind) {
		case sessionEventKindTTSChunkPlayed:
			length := int32(payloadInt(event.Payload, "playedTextLength"))
			if length > playedLength {
				playedLength = length
				playedHash = strings.TrimSpace(payloadString(event.Payload, "playedTextHash"))
				playedOffset = payloadInt(event.Payload, "playbackOffsetMs")
			}
		case sessionEventKindContextCommitted:
			length := int32(payloadInt(event.Payload, "committedTextLength"))
			if length > committedLength {
				committedLength = length
				committedHash = strings.TrimSpace(payloadString(event.Payload, "committedTextHash"))
				committedOffset = payloadInt(event.Payload, "playbackOffsetMs")
			}
		case sessionEventKindBargeInDetected:
			out.Interrupted = true
			out.InterruptionOffsetMs = payloadInt(event.Payload, "playbackOffsetMs")
			out.UserSpeechStartedAt = strings.TrimSpace(payloadString(event.Payload, "userSpeechStartedAt"))
		}
	}

	length := committedLength
	hash := committedHash
	offset := committedOffset
	if length == 0 && out.Interrupted && playedLength > 0 {
		length = playedLength
		hash = playedHash
		offset = playedOffset
	}
	sourceLength := source.AssistantTextLength
	if sourceLength <= 0 {
		sourceLength = int32(len([]rune(source.AssistantTextDraft)))
	}
	if length > sourceLength {
		length = sourceLength
	}
	if length > 0 {
		out.HasCommittedContext = true
		out.CommittedTextLength = length
		out.CommittedAssistantText = firstRunes(source.AssistantTextDraft, int(length))
		out.CommittedTextHash = hash
		out.PlaybackOffsetMs = offset
		if out.CommittedTextHash == "" && length == sourceLength {
			out.CommittedTextHash = source.AssistantTextHash
		}
	}
	if out.Interrupted {
		out.InterruptionNote = voiceInterruptionNote(out)
	}
	return out
}

func firstRunes(value string, count int) string {
	if count <= 0 {
		return ""
	}
	runes := []rune(value)
	if count > len(runes) {
		count = len(runes)
	}
	return string(runes[:count])
}

func voiceInterruptionNote(ctx CommittedVoiceContext) string {
	if ctx.UserSpeechStartedAt != "" {
		return fmt.Sprintf("Assistant playback was interrupted at %dms when user speech started at %s.", ctx.InterruptionOffsetMs, ctx.UserSpeechStartedAt)
	}
	return fmt.Sprintf("Assistant playback was interrupted at %dms.", ctx.InterruptionOffsetMs)
}
