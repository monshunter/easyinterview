package minimax_speech

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providerregistry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

const PathTTSSynthesize = "/v1/tts/synthesize"

const HeaderRequestID = "X-Request-ID"

// Options configures the adapter.
type Options struct {
	Provider   providerregistry.ResolvedProvider
	HTTPClient *http.Client
}

// Adapter implements aiclient.Provider for MiniMax TTS services.
type Adapter struct {
	providerRef string
	baseURL     string
	apiKey      string
	client      *http.Client
}

// New constructs an Adapter.
func New(opts Options) (*Adapter, error) {
	if opts.Provider.Entry.Name == "" {
		return nil, errors.New("minimax_speech: resolved provider is required")
	}
	if opts.Provider.Entry.Protocol != aiclient.ProviderProtocolMinimaxSpeech {
		return nil, fmt.Errorf("minimax_speech: provider %q protocol must be %q", opts.Provider.Entry.Name, aiclient.ProviderProtocolMinimaxSpeech)
	}
	if opts.Provider.BaseURL == "" {
		return nil, errors.New("minimax_speech: resolved provider BaseURL is required")
	}
	if opts.Provider.APIKey == "" {
		return nil, errors.New("minimax_speech: resolved provider APIKey is required")
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{}
	}
	return &Adapter{
		providerRef: opts.Provider.Entry.Name,
		baseURL:     strings.TrimRight(opts.Provider.BaseURL, "/"),
		apiKey:      opts.Provider.APIKey,
		client:      hc,
	}, nil
}

// Name implements aiclient.Provider.
func (a *Adapter) Name() string { return a.providerRef }

// Complete implements aiclient.Provider. MiniMax speech does not support chat.
func (a *Adapter) Complete(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	return aiclient.CompleteResponse{}, a.errMeta(profile, sharederrors.CodeAiUnsupportedCapability, "minimax_speech does not support chat"), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "minimax_speech does not support chat", false)
}

// Stream implements aiclient.Provider.
func (a *Adapter) Stream(ctx context.Context, profile *aiclient.ModelProfile, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "minimax_speech does not support streaming", false)
}

// Transcribe implements aiclient.Provider. MiniMax STT is not confirmed.
func (a *Adapter) Transcribe(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, a.errMeta(profile, sharederrors.CodeAiUnsupportedCapability, "minimax_speech does not support STT transcription"), sharederrors.Wrap(sharederrors.CodeAiUnsupportedCapability, "minimax_speech STT is not confirmed per plan 004", false)
}

// Synthesize implements aiclient.Provider using the MiniMax TTS endpoint.
func (a *Adapter) Synthesize(ctx context.Context, profile *aiclient.ModelProfile, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	if profile == nil {
		return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("minimax_speech: profile is nil")
	}

	req := ttsSynthesizeRequest{
		Text:         input.Text,
		Voice:        input.Voice,
		Format:       input.Format,
		SpeakingRate: input.SpeakingRate,
		Language:     input.Language,
		Model:        profile.Default.Model,
	}

	respBody, status, headers, err := a.postJSON(ctx, profile.TimeoutMs, PathTTSSynthesize, req)
	if err != nil {
		return aiclient.SynthesisResponse{}, a.errMeta(profile, errorCodeOf(err), err.Error()), err
	}
	if status >= 400 {
		err := mapHTTPError(status, respBody)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, errorCodeOf(err), err.Error()), err
	}

	var wire ttsSynthesizeResponse
	if err := json.Unmarshal(respBody, &wire); err != nil {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "minimax_speech: parse tts response: "+err.Error(), false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid, err.Error()), err
	}
	if wire.Audio == "" {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "minimax_speech: tts response missing audio", false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid, err.Error()), err
	}

	audio, err := decodeBase64Audio(wire.Audio)
	if err != nil {
		err := sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "minimax_speech: decode tts audio: "+err.Error(), false)
		return aiclient.SynthesisResponse{}, a.errMeta(profile, sharederrors.CodeAiOutputInvalid, err.Error()), err
	}

	_ = headers
	meta := aiclient.AICallMeta{
		Provider:     a.providerRef,
		ModelFamily:  "minimax_speech",
		ModelID:      profile.Default.Model,
		InputTokens:  wire.CharCount,
		OutputTokens: wire.DurationMs,
	}
	return aiclient.SynthesisResponse{
		Audio:       audio,
		ContentType: wire.ContentType,
		DurationMs:  wire.DurationMs,
		CharCount:   wire.CharCount,
	}, meta, nil
}

func (a *Adapter) postJSON(ctx context.Context, timeoutMs int, path string, body any) ([]byte, int, http.Header, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("minimax_speech: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+path, strings.NewReader(string(buf)))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("minimax_speech: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	if rid := aiclient.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set(HeaderRequestID, rid)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, 0, nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "minimax_speech: transport error: "+err.Error(), true)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "minimax_speech: read response: "+err.Error(), true)
	}
	return respBody, resp.StatusCode, resp.Header, nil
}

func (a *Adapter) errMeta(profile *aiclient.ModelProfile, code string, msg string) aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider:            a.providerRef,
		ModelID:             profile.Default.Model,
		ValidationStatus:    aiclient.ValidationStatusInvalid,
		ErrorCode:           code,
		Capability:          profile.Capability,
		ModelProfileName:    profile.Name,
		ModelProfileVersion: profile.Version,
		Route:               profile.Route,
	}
}

func mapHTTPError(status int, body []byte) error {
	if status >= 500 {
		return sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, fmt.Sprintf("minimax_speech: upstream %d", status), true)
	}
	var env errorEnvelope
	if json.Unmarshal(body, &env) == nil && env.Error.Code != "" {
		if meta, ok := sharederrors.CodeRegistry[env.Error.Code]; ok {
			return sharederrors.Wrap(env.Error.Code, env.Error.Message, meta.Retryable)
		}
	}
	return sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, fmt.Sprintf("minimax_speech: upstream %d", status), false)
}

func errorCodeOf(err error) string {
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code
	}
	return sharederrors.CodeAiOutputInvalid
}

func encodeBase64Audio(audio []byte) string {
	return base64.StdEncoding.EncodeToString(audio)
}

func decodeBase64Audio(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}
