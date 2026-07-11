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

type voicePlayedEvidence struct {
	TextHash         string
	TextLength       int32
	PlaybackOffsetMs int64
}

func BuildCommittedVoiceContext(source PracticeVoiceTurnContextSource, events []VoicePlaybackEventRecord) CommittedVoiceContext {
	voiceTurnID := strings.TrimSpace(source.VoiceTurnID)
	out := CommittedVoiceContext{VoiceTurnID: voiceTurnID}
	if voiceTurnID == "" || len(events) == 0 {
		return out
	}
	sourceHash := canonicalVoiceTextHash(source.AssistantTextHash)
	if sourceHash == "" {
		return out
	}
	sourceLength := source.AssistantTextLength
	if sourceLength <= 0 {
		sourceLength = int32(len([]rune(source.AssistantTextDraft)))
	}
	if sourceLength <= 0 {
		return out
	}
	ordered := append([]VoicePlaybackEventRecord(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].OccurredAt.Before(ordered[j].OccurredAt)
	})

	playedByChunk := make(map[string]voicePlayedEvidence)
	var committedLength int32
	var committedHash string
	var committedOffset int64
	var interruptedPlayed *voicePlayedEvidence

	for _, event := range ordered {
		if strings.TrimSpace(payloadString(event.Payload, "voiceTurnId")) != voiceTurnID {
			continue
		}
		chunkID := strings.TrimSpace(payloadString(event.Payload, "chunkId"))
		if chunkID == "" {
			continue
		}
		switch strings.TrimSpace(event.Kind) {
		case sessionEventKindTTSChunkPlayed:
			length := int32(payloadInt(event.Payload, "playedTextLength"))
			hash := strings.TrimSpace(payloadString(event.Payload, "playedTextHash"))
			if length <= 0 || length > sourceLength || canonicalVoiceTextHash(hash) != sourceHash {
				continue
			}
			if previous, ok := playedByChunk[chunkID]; !ok || length > previous.TextLength {
				playedByChunk[chunkID] = voicePlayedEvidence{
					TextHash:         hash,
					TextLength:       length,
					PlaybackOffsetMs: payloadInt(event.Payload, "playbackOffsetMs"),
				}
			}
		case sessionEventKindContextCommitted:
			length := int32(payloadInt(event.Payload, "committedTextLength"))
			hash := strings.TrimSpace(payloadString(event.Payload, "committedTextHash"))
			played, ok := playedByChunk[chunkID]
			if !ok || length <= 0 || length > played.TextLength || canonicalVoiceTextHash(hash) != sourceHash {
				continue
			}
			if length > committedLength {
				committedLength = length
				committedHash = hash
				committedOffset = payloadInt(event.Payload, "playbackOffsetMs")
			}
		case sessionEventKindBargeInDetected:
			out.Interrupted = true
			out.InterruptionOffsetMs = payloadInt(event.Payload, "playbackOffsetMs")
			out.UserSpeechStartedAt = strings.TrimSpace(payloadString(event.Payload, "userSpeechStartedAt"))
			interruptedPlayed = nil
			if played, ok := playedByChunk[chunkID]; ok {
				copy := played
				interruptedPlayed = &copy
			}
		}
	}

	length := committedLength
	hash := committedHash
	offset := committedOffset
	if length == 0 && out.Interrupted && interruptedPlayed != nil {
		length = interruptedPlayed.TextLength
		hash = interruptedPlayed.TextHash
		offset = interruptedPlayed.PlaybackOffsetMs
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

func canonicalVoiceTextHash(value string) string {
	digest := strings.TrimSpace(value)
	if !validSHA256Digest(digest) {
		return ""
	}
	digest = strings.TrimPrefix(digest, "sha256:")
	return strings.ToLower(digest)
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
