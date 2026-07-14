package doubaospeech

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/internal/responsebody"
	platformconfig "github.com/monshunter/easyinterview/backend/internal/platform/config"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const (
	PathTTSSynthesize = "/v1/tts/synthesize"
	PathSTTRecognize  = "/v1/audio/recognize"
)

const (
	HeaderRequestID = "X-Request-ID"
)

// Options configures the adapter.
type Options struct {
	Provider             providerregistry.ResolvedProvider
	HTTPClient           *http.Client
	MaxResponseBodyBytes int64
}

// Adapter implements aiclient.Provider for Doubao speech services.
type Adapter struct {
	providerRef          string
	baseURL              string
	apiKey               string
	client               *http.Client
	maxResponseBodyBytes int64
}

// New constructs an Adapter.
func New(opts Options) (*Adapter, error) {
	if opts.Provider.Entry.Name == "" {
		return nil, errors.New("doubao_speech: resolved provider is required")
	}
	if opts.Provider.Entry.Protocol != aiclient.ProviderProtocolDoubaoSpeech {
		return nil, fmt.Errorf("doubao_speech: provider %q protocol must be %q", opts.Provider.Entry.Name, aiclient.ProviderProtocolDoubaoSpeech)
	}
	if opts.Provider.BaseURL == "" {
		return nil, errors.New("doubao_speech: resolved provider BaseURL is required")
	}
	if opts.Provider.APIKey == "" {
		return nil, errors.New("doubao_speech: resolved provider APIKey is required")
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}
	if opts.MaxResponseBodyBytes <= 0 {
		opts.MaxResponseBodyBytes = platformconfig.DefaultContentLimits().AIProviderMaxResponseBodyBytes
	}
	return &Adapter{
		providerRef:          opts.Provider.Entry.Name,
		baseURL:              strings.TrimRight(opts.Provider.BaseURL, "/"),
		apiKey:               opts.Provider.APIKey,
		client:               hc,
		maxResponseBodyBytes: opts.MaxResponseBodyBytes,
	}, nil
}

// Name implements aiclient.Provider.
func (a *Adapter) Name() string { return a.providerRef }

// Complete implements aiclient.Provider. Doubao speech does not support chat.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return aiclient.CompleteResponse{}, a.errMeta(profile, sharederrors.CodeAiUnsupportedCapability), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "doubao_speech does not support chat", false)
}

// Stream implements aiclient.Provider. Doubao speech does not support streaming chat.
func (a *Adapter) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "doubao_speech does not support streaming chat", false)
}

// Synthesize implements aiclient.Provider using the Doubao TTS endpoint.
func (a *Adapter) Synthesize(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("doubao_speech: profile is nil")
	}

	req := ttsSynthesizeRequest{
		Text:         input.Text,
		Voice:        input.Voice,
		Format:       input.Format,
		SpeakingRate: input.SpeakingRate,
		Language:     input.Language,
		Model:        profile.Default.Model,
	}

	respBody, status, err := a.postJSON(ctx, profile.TimeoutMs, PathTTSSynthesize, req)
	if err != nil {
		return aiclient.SynthesisResponse{}, a.errMeta(profile, errorCodeOf(err)), err
	}
	if status >= 400 {
		err := mapHTTPError(status, respBody)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, errorCodeOf(err)), err
	}

	var wire ttsSynthesizeResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "doubao_speech: parse tts response: "+err.Error(), false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid), err
	}
	if wire.Audio == "" {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "doubao_speech: tts response missing audio", false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid), err
	}

	audio, err := base64.StdEncoding.DecodeString(wire.Audio)
	if err != nil {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "doubao_speech: decode tts audio: "+err.Error(), false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid), err
	}

	meta := a.buildMeta(profile, wire.DurationMs, wire.CharCount)
	return aiclient.SynthesisResponse{
		Audio:       audio,
		ContentType: wire.ContentType,
		DurationMs:  wire.DurationMs,
		CharCount:   wire.CharCount,
	}, meta, nil
}

// Transcribe implements aiclient.Provider using the Doubao STT endpoint.
func (a *Adapter) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("doubao_speech: profile is nil")
	}

	req := map[string]any{
		"audio":        base64.StdEncoding.EncodeToString(input.Audio),
		"content_type": input.ContentType,
		"language":     input.Language,
		"model":        profile.Default.Model,
	}

	respBody, status, err := a.postJSON(ctx, profile.TimeoutMs, PathSTTRecognize, req)
	if err != nil {
		return aiclient.TranscriptionResponse{}, a.errMeta(profile, errorCodeOf(err)), err
	}
	if status >= 400 {
		err := mapHTTPError(status, respBody)
		return aiclient.TranscriptionResponse{}, a.errMeta(profile, errorCodeOf(err)), err
	}

	var wire sttTranscribeResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "doubao_speech: parse stt response: "+err.Error(), false)
		return aiclient.TranscriptionResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid), err
	}
	if wire.Text == "" {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "doubao_speech: stt response missing text", false)
		return aiclient.TranscriptionResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid), err
	}

	meta := aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelFamily:  "doubao_speech",
		ModelID:      profile.Default.Model,
		OutputTokens: len(wire.Text),
	}
	return aiclient.TranscriptionResponse{Text: wire.Text}, meta, nil
}

// postJSON sends a JSON POST request.
func (a *Adapter) postJSON(ctx context.Context, timeoutMs int, path string, body any) ([]byte, int, error) {
	if timeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
	}
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, 0, fmt.Errorf("doubao_speech: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, strings.NewReader(string(buf)))
	if err != nil {
		return nil, 0, fmt.Errorf("doubao_speech: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, 0, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "doubao_speech: timeout", true)
		}
		return nil, 0, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "doubao_speech: transport error: "+err.Error(), true)
	}
	defer resp.Body.Close()

	respBody, err := responsebody.Read(resp, a.maxResponseBodyBytes)
	if err != nil {
		code := sharederrors.CodeAiOutputInvalid
		if errors.Is(err, responsebody.ErrRead) {
			code = sharederrors.CodeAiProviderTimeout
		}
		meta := sharederrors.CodeRegistry[code]
		return nil, 0, sharederrors.Wrap(code, meta.Message, meta.Retryable)
	}
	return respBody, resp.StatusCode, nil
}

func (a *Adapter) errMeta(profile *aiclient.ModelProfile, code string) aiclient.AICallMeta {
	meta := aiclient.AICallMeta{
		Provider:            a.providerRef,
		ModelID:             profile.Default.Model,
		ValidationStatus:    aiclient.ValidationStatusInvalid,
		ErrorCode:           code,
		Capability:          profile.Capability,
		ModelProfileName:    profile.Name,
		ModelProfileVersion: profile.Version,
		Route:               profile.Route,
	}
	if code == sharederrors.CodeAiProviderTimeout {
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
	}
	return meta
}

func (a *Adapter) buildMeta(profile *aiclient.ModelProfile, durationMs, charCount int) aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelFamily:  "doubao_speech",
		ModelID:      profile.Default.Model,
		InputTokens:  charCount,
		OutputTokens: durationMs,
	}
}

func mapHTTPError(status int, body []byte) error {
	if status >= 500 {
		return sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, fmt.Sprintf("doubao_speech: upstream %d", status), true)
	}
	var env errorEnvelope
	if json.Unmarshal(body, &env) == nil && env.Error.Code != "" {
		if meta, ok := sharederrors.CodeRegistry[env.Error.Code]; ok {
			return sharederrors.Wrap(env.Error.Code, env.Error.Message, meta.Retryable)
		}
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, fmt.Sprintf("doubao_speech: upstream %d", status), false)
}

func errorCodeOf(err error) string {
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return sharederrors.CodeAiOutputInvalid
}
