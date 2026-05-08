package minimax_speech

// ttsSynthesizeRequest is the JSON body posted to the MiniMax TTS endpoint.
type ttsSynthesizeRequest struct {
	Text         string  `json:"text"`
	Voice        string  `json:"voice"`
	Format       string  `json:"format,omitempty"`
	SpeakingRate float64 `json:"speaking_rate,omitempty"`
	Language     string  `json:"language,omitempty"`
	Model        string  `json:"model"`
}

// ttsSynthesizeResponse is the JSON response from the MiniMax TTS endpoint.
type ttsSynthesizeResponse struct {
	Audio       string `json:"audio"` // base64-encoded audio
	ContentType string `json:"content_type"`
	DurationMs  int    `json:"duration_ms"`
	CharCount   int    `json:"char_count"`
}

// errorEnvelope mirrors the MiniMax error response shape.
type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
