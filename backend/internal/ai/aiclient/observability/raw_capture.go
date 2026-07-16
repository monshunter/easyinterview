package observability

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

const (
	// RawCompleteRecordVersion is the stable local Complete capture schema.
	RawCompleteRecordVersion = "ai.complete.raw.v1"

	RawCompleteRequestType  = "ai.complete.request"
	RawCompleteResponseType = "ai.complete.response"
)

// RawCompleteRecord is one append-only NDJSON envelope. Payload is always one
// of RawCompleteRequestPayload or RawCompleteResponsePayload. The envelope
// deliberately has no user/resource identity, headers, endpoint or credentials.
type RawCompleteRecord struct {
	RecordVersion string    `json:"recordVersion"`
	Type          string    `json:"type"`
	CallID        string    `json:"callId"`
	CapturedAt    time.Time `json:"capturedAt"`
	ProfileName   string    `json:"profileName"`
	Payload       any       `json:"payload"`
}

// RawCompleteRequestPayload is the provider-neutral Complete input allowed in
// the local recorder. CallMetadata is projected instead of marshalled wholesale
// so TaskRun identity and future internal fields cannot leak by accident.
type RawCompleteRequestPayload struct {
	Messages     []aiclient.Message   `json:"messages"`
	Tools        []aiclient.Tool      `json:"tools,omitempty"`
	ToolChoice   *aiclient.ToolChoice `json:"toolChoice,omitempty"`
	OutputSchema json.RawMessage      `json:"outputSchema,omitempty"`
	Routing      *RawCompleteRouting  `json:"routing,omitempty"`
}

// RawCompleteRouting contains only allowlisted model-behaviour coordinates.
// Provider endpoints and credentials are intentionally not representable.
type RawCompleteRouting struct {
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"topP,omitempty"`
	Thinking    string   `json:"thinking,omitempty"`
	MaxTokens   int      `json:"maxTokens,omitempty"`
}

// RawCompleteResponsePayload is the provider-neutral Complete output plus a
// bounded metadata projection. ErrorCode is stable; Go/provider error text is
// never accepted by this type.
type RawCompleteResponsePayload struct {
	Content          string                  `json:"content"`
	FinishReason     string                  `json:"finishReason,omitempty"`
	ToolCalls        []aiclient.ToolCall     `json:"toolCalls,omitempty"`
	ValidationStatus string                  `json:"validationStatus,omitempty"`
	ErrorCode        string                  `json:"errorCode,omitempty"`
	Meta             RawCompleteResponseMeta `json:"meta"`
}

// RawCompleteResponseMeta is a closed, identity-free projection of AICallMeta.
type RawCompleteResponseMeta struct {
	Provider            string                        `json:"provider,omitempty"`
	ModelFamily         string                        `json:"modelFamily,omitempty"`
	ModelID             string                        `json:"modelId,omitempty"`
	Capability          aiclient.Capability           `json:"capability,omitempty"`
	ModelProfileVersion string                        `json:"modelProfileVersion,omitempty"`
	InputTokens         int                           `json:"inputTokens,omitempty"`
	OutputTokens        int                           `json:"outputTokens,omitempty"`
	CostUSDMicros       int64                         `json:"costUsdMicros,omitempty"`
	LatencyMs           int64                         `json:"latencyMs,omitempty"`
	FallbackChain       []string                      `json:"fallbackChain,omitempty"`
	Route               string                        `json:"route,omitempty"`
	ToolInvocations     []aiclient.ToolInvocationMeta `json:"toolInvocations,omitempty"`
	PartialMetaReason   string                        `json:"partialMetaReason,omitempty"`
}

// RawIOCapture is the narrow seam the decorator consumes. Runtime code opens
// one FileRawIOCapture and shares it across every Complete wrapper.
type RawIOCapture interface {
	RecordRawComplete(record RawCompleteRecord) error
}

// FileRawIOCapture is a mutex-protected append-only NDJSON recorder.
type FileRawIOCapture struct {
	mu     sync.Mutex
	file   *os.File
	closed bool
	now    func() time.Time
}

// OpenRawIOCapture securely opens one dedicated raw Complete file. path must
// already be absolute (platform/config owns ConfigDir-parent resolution).
// Every existing path component is checked with Lstat; symlinks and a
// non-regular target fail closed. The dedicated parent and file are tightened
// to 0700/0600 even when they already exist.
func OpenRawIOCapture(path string) (*FileRawIOCapture, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "." || !filepath.IsAbs(clean) {
		return nil, errors.New("observability: raw capture path must be absolute")
	}
	parent := filepath.Dir(clean)
	if clean == parent || filepath.Dir(parent) == parent {
		return nil, errors.New("observability: raw capture requires a dedicated parent directory")
	}
	if err := rejectSymlinkComponents(parent); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return nil, errors.New("observability: create raw capture parent failed")
	}
	if err := rejectSymlinkComponents(parent); err != nil {
		return nil, err
	}
	if err := os.Chmod(parent, 0o700); err != nil {
		return nil, errors.New("observability: harden raw capture parent failed")
	}

	if info, err := os.Lstat(clean); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("observability: raw capture target must not be a symlink")
		}
		if !info.Mode().IsRegular() {
			return nil, errors.New("observability: raw capture target must be a regular file")
		}
		if err := os.Chmod(clean, 0o600); err != nil {
			return nil, errors.New("observability: harden raw capture file failed")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("observability: inspect raw capture target failed")
	}

	file, err := os.OpenFile(clean, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, errors.New("observability: open raw capture failed")
	}
	closeOnError := func(message string) (*FileRawIOCapture, error) {
		_ = file.Close()
		return nil, errors.New(message)
	}
	if err := file.Chmod(0o600); err != nil {
		return closeOnError("observability: harden opened raw capture failed")
	}
	openedInfo, err := file.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() {
		return closeOnError("observability: opened raw capture is not regular")
	}
	pathInfo, err := os.Lstat(clean)
	if err != nil || pathInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(openedInfo, pathInfo) {
		return closeOnError("observability: raw capture target changed during open")
	}
	return &FileRawIOCapture{file: file, now: time.Now}, nil
}

// RecordRawComplete marshals one full line while holding the process-shared
// mutex, appends it with one Write call and Syncs it before returning. Request
// durability therefore precedes the provider invocation.
func (c *FileRawIOCapture) RecordRawComplete(record RawCompleteRecord) error {
	if c == nil {
		return errors.New("observability: raw capture is nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.file == nil {
		return errors.New("observability: raw capture is closed")
	}
	if record.Type != RawCompleteRequestType && record.Type != RawCompleteResponseType {
		return errors.New("observability: raw capture record type is invalid")
	}
	if err := idx.RequireServerID(record.CallID); err != nil {
		return errors.New("observability: raw capture call ID must be UUIDv7")
	}
	if strings.TrimSpace(record.ProfileName) == "" || record.Payload == nil {
		return errors.New("observability: raw capture envelope is incomplete")
	}
	record.RecordVersion = RawCompleteRecordVersion
	if record.CapturedAt.IsZero() {
		record.CapturedAt = c.now().UTC()
	} else {
		record.CapturedAt = record.CapturedAt.UTC()
	}
	line, err := json.Marshal(record)
	if err != nil {
		return errors.New("observability: encode raw capture record failed")
	}
	line = append(line, '\n')
	n, err := c.file.Write(line)
	if err != nil {
		return errors.New("observability: append raw capture record failed")
	}
	if n != len(line) {
		return io.ErrShortWrite
	}
	if err := c.file.Sync(); err != nil {
		return errors.New("observability: sync raw capture record failed")
	}
	return nil
}

// Close flushes and closes the process-owned recorder once.
func (c *FileRawIOCapture) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if c.file == nil {
		return nil
	}
	return c.file.Close()
}

func rejectSymlinkComponents(path string) error {
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	rooted := filepath.IsAbs(clean)
	rest = strings.TrimPrefix(rest, string(filepath.Separator))
	current := volume
	if rooted {
		current += string(filepath.Separator)
	}
	for _, component := range strings.Split(rest, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return errors.New("observability: inspect raw capture path failed")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("observability: raw capture path must not contain symlinks")
		}
		if !info.IsDir() {
			return errors.New("observability: raw capture path component is not a directory")
		}
	}
	return nil
}

func rawRouting(profile *aiclient.ModelProfile) *RawCompleteRouting {
	if profile == nil {
		return nil
	}
	routing := &RawCompleteRouting{MaxTokens: profile.MaxTokens}
	if value, ok := profile.Default.Params["temperature"]; ok {
		if parsed, ok := numericFloat(value); ok {
			routing.Temperature = &parsed
		}
	}
	if value, ok := profile.Default.Params["top_p"]; ok {
		if parsed, ok := numericFloat(value); ok {
			routing.TopP = &parsed
		}
	}
	if value, ok := profile.Default.Params["thinking"].(string); ok {
		routing.Thinking = value
	}
	if routing.MaxTokens == 0 && routing.Temperature == nil && routing.TopP == nil && routing.Thinking == "" {
		return nil
	}
	return routing
}

func numericFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func rawResponseMeta(meta aiclient.AICallMeta) RawCompleteResponseMeta {
	return RawCompleteResponseMeta{
		Provider:            meta.Provider,
		ModelFamily:         meta.ModelFamily,
		ModelID:             meta.ModelID,
		Capability:          meta.Capability,
		ModelProfileVersion: meta.ModelProfileVersion,
		InputTokens:         meta.InputTokens,
		OutputTokens:        meta.OutputTokens,
		CostUSDMicros:       meta.CostUSDMicros,
		LatencyMs:           meta.LatencyMs,
		FallbackChain:       append([]string(nil), meta.FallbackChain...),
		Route:               meta.Route,
		ToolInvocations:     append([]aiclient.ToolInvocationMeta(nil), meta.ToolInvocations...),
		PartialMetaReason:   meta.PartialMetaReason,
	}
}

var _ RawIOCapture = (*FileRawIOCapture)(nil)
var _ io.Closer = (*FileRawIOCapture)(nil)
