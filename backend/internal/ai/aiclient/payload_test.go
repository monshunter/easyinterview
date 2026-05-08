package aiclient

import (
	"encoding/json"
	"testing"
)

func TestSynthesisInputJSONRoundTrip(t *testing.T) {
	input := SynthesisInput{
		Text:         "你好，欢迎参加面试",
		Voice:        "zh_female_qingxin",
		Format:       "mp3",
		SpeakingRate: 1.0,
		Language:     "zh-CN",
		Metadata: CallMetadata{
			FeatureKey:    "practice.voice.tts.default",
			PromptVersion: "1.0",
			RubricVersion: "1.0",
			Language:      "zh-CN",
		},
	}

	b, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("SynthesisInput json marshal: %v", err)
	}

	var got SynthesisInput
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("SynthesisInput json unmarshal: %v", err)
	}

	if got.Text != input.Text {
		t.Fatalf("Text = %q, want %q", got.Text, input.Text)
	}
	if got.Voice != input.Voice {
		t.Fatalf("Voice = %q, want %q", got.Voice, input.Voice)
	}
	if got.Format != input.Format {
		t.Fatalf("Format = %q, want %q", got.Format, input.Format)
	}
	if got.SpeakingRate != input.SpeakingRate {
		t.Fatalf("SpeakingRate = %f, want %f", got.SpeakingRate, input.SpeakingRate)
	}
	if got.Language != input.Language {
		t.Fatalf("Language = %q, want %q", got.Language, input.Language)
	}
}

func TestSynthesisResponseJSONRoundTrip(t *testing.T) {
	resp := SynthesisResponse{
		Audio:       []byte{0x01, 0x02, 0x03},
		ContentType: "audio/mpeg",
		DurationMs:  1500,
		CharCount:   10,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("SynthesisResponse json marshal: %v", err)
	}

	var got SynthesisResponse
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("SynthesisResponse json unmarshal: %v", err)
	}

	// Audio bytes must NOT appear in the JSON (privacy gate).
	if len(got.Audio) != 0 {
		t.Fatal("SynthesisResponse.Audio must not be deserialized from JSON")
	}
	if got.ContentType != resp.ContentType {
		t.Fatalf("ContentType = %q, want %q", got.ContentType, resp.ContentType)
	}
	if got.DurationMs != resp.DurationMs {
		t.Fatalf("DurationMs = %d, want %d", got.DurationMs, resp.DurationMs)
	}
	if got.CharCount != resp.CharCount {
		t.Fatalf("CharCount = %d, want %d", got.CharCount, resp.CharCount)
	}

	// Verify Audio field has json:"-" tag by checking the JSON output.
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("raw json unmarshal: %v", err)
	}
	if _, ok := raw["audio"]; ok {
		t.Fatal("SynthesisResponse JSON must not contain 'audio' key")
	}
}

func TestSynthesisInputRequiredFields(t *testing.T) {
	input := SynthesisInput{
		Text: "test",
	}
	if input.Voice != "" {
		t.Fatalf("default Voice = %q, want empty", input.Voice)
	}
	if input.Format != "" {
		t.Fatalf("default Format = %q, want empty", input.Format)
	}
	if input.SpeakingRate != 0 {
		t.Fatalf("default SpeakingRate = %f, want 0", input.SpeakingRate)
	}
}
