package observability_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

const rawCompleteRecordVersion = "ai.complete.raw.v1"

type rawCompleteRecord struct {
	RecordVersion string         `json:"recordVersion"`
	Type          string         `json:"type"`
	CallID        string         `json:"callId"`
	CapturedAt    string         `json:"capturedAt"`
	ProfileName   string         `json:"profileName"`
	Payload       map[string]any `json:"payload"`
}

type syncCapture struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	syncCount int
}

type failingRawCapture struct {
	mu    sync.Mutex
	calls int
}

func (c *failingRawCapture) RecordRawComplete(observability.RawCompleteRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return errors.New("/private/raw/path: provider response and credentials must not reach logs")
}

func (c *failingRawCapture) Calls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func (w *syncCapture) RecordRawComplete(record observability.RawCompleteRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	record.RecordVersion = observability.RawCompleteRecordVersion
	if record.CapturedAt.IsZero() {
		record.CapturedAt = time.Now().UTC()
	}
	line, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if _, err := w.buf.Write(append(line, '\n')); err != nil {
		return err
	}
	w.syncCount++
	return nil
}

func TestOpenRawIOCaptureRejectsVolumeRootBeforeAnyPermissionMutation(t *testing.T) {
	canonical := canonicalTempDir(t)
	volumeRoot := filepath.VolumeName(canonical) + string(filepath.Separator)
	before, err := os.Stat(volumeRoot)
	if err != nil {
		t.Fatalf("stat volume root: %v", err)
	}
	for _, path := range []string{
		volumeRoot,
		filepath.Join(volumeRoot, "easyinterview-raw-capture-must-not-open.ndjson"),
	} {
		if capture, err := observability.OpenRawIOCapture(path); err == nil {
			_ = capture.Close()
			t.Fatalf("OpenRawIOCapture(%q) accepted a volume-root target/parent", path)
		}
	}
	after, err := os.Stat(volumeRoot)
	if err != nil {
		t.Fatalf("stat volume root after rejection: %v", err)
	}
	if before.Mode().Perm() != after.Mode().Perm() {
		t.Fatalf("volume root mode changed from %#o to %#o", before.Mode().Perm(), after.Mode().Perm())
	}
}

func (w *syncCapture) Snapshot() ([]byte, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return bytes.Clone(w.buf.Bytes()), w.syncCount
}

type rawCaptureInner struct {
	mu             sync.Mutex
	beforeComplete func()
	response       aiclient.CompleteResponse
	meta           aiclient.AICallMeta
	err            error
	seenTaskRunID  string
}

func (c *rawCaptureInner) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.mu.Lock()
	c.seenTaskRunID = payload.Metadata.TaskRun.ID
	c.mu.Unlock()
	if c.beforeComplete != nil {
		c.beforeComplete()
	}
	return c.response, c.meta, c.err
}

func (c *rawCaptureInner) SeenTaskRunID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.seenTaskRunID
}

func (c *rawCaptureInner) Transcribe(_ context.Context, _ string, _ aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{Text: "transcript must not enter Complete raw capture"}, rawSuccessMeta(), nil
}

func (c *rawCaptureInner) Stream(_ context.Context, _ string, _ aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	ch := make(chan aiclient.AIStreamEvent)
	close(ch)
	return ch, nil
}

func (c *rawCaptureInner) Synthesize(_ context.Context, _ string, _ aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{Audio: []byte("audio must not enter Complete raw capture"), ContentType: "audio/mpeg"}, rawSuccessMeta(), nil
}

func rawSuccessMeta() aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider:            "unit-test-provider",
		ModelFamily:         "unit-test-model",
		ModelID:             "unit-test-model-v1",
		Capability:          aiclient.CapabilityChat,
		ModelProfileName:    "practice.chat.default",
		ModelProfileVersion: "1.0.0",
		InputTokens:         8,
		OutputTokens:        13,
		ValidationStatus:    aiclient.ValidationStatusOK,
	}
}

func rawCapturePayload() aiclient.CompletePayload {
	payload := samplePayload()
	payload.Messages = []aiclient.Message{
		{Role: "system", Content: "system raw capture contract"},
		{Role: "user", Content: "user raw capture contract"},
	}
	payload.Tools = []aiclient.Tool{{
		Name:        "lookup_interview_note",
		Description: "Look up one note.",
		Parameters:  json.RawMessage(`{"type":"object","required":["noteId"],"properties":{"noteId":{"type":"string"}}}`),
	}}
	payload.ToolChoice = &aiclient.ToolChoice{Mode: aiclient.ToolChoiceModeTool, Name: "lookup_interview_note"}
	payload.Metadata.OutputSchema = json.RawMessage(`{"type":"object","required":["answer"],"properties":{"answer":{"type":"string"}}}`)
	return payload
}

func newRawCaptureDecorator(t *testing.T, inner aiclient.AIClient, capture observability.RawIOCapture) (*observability.Wrap, *memTaskRunWriter, *memAuditWriter, *observability.MemoryLogger) {
	t.Helper()
	runs := &memTaskRunWriter{}
	audit := &memAuditWriter{}
	logger := observability.NewMemoryLogger()
	wrap, err := observability.New(inner,
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runs),
		observability.WithAuditEventWriter(audit),
		observability.WithProfileResolver(routeAwareResolver()),
		observability.WithRawIOCapture(capture),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	return wrap, runs, audit, logger
}

func TestRawCompleteCapture_WritesSynchronousVersionedPairAndCorrelatesTaskRun(t *testing.T) {
	capture := &syncCapture{}
	providerSawRequest := false
	inner := &rawCaptureInner{
		response: aiclient.CompleteResponse{
			Content:      `{"answer":"captured"}`,
			FinishReason: "tool_calls",
			ToolCalls: []aiclient.ToolCall{{
				ID:        "call-1",
				Name:      "lookup_interview_note",
				Arguments: json.RawMessage(`{"noteId":"note-1"}`),
			}},
		},
		meta: rawSuccessMeta(),
	}
	inner.beforeComplete = func() {
		raw, syncCount := capture.Snapshot()
		records, err := parseRawCompleteRecords(raw)
		providerSawRequest = err == nil && syncCount > 0 && len(records) == 1 && records[0].Type == "ai.complete.request"
	}
	wrap, runs, _, _ := newRawCaptureDecorator(t, inner, capture)

	if _, _, err := wrap.Complete(context.Background(), "practice.chat.default", rawCapturePayload()); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if !providerSawRequest {
		t.Fatal("provider was called before the request NDJSON line was written and synced")
	}

	raw, _ := capture.Snapshot()
	records := mustParseRawCompleteRecords(t, raw)
	if len(records) != 2 {
		t.Fatalf("raw records = %d, want request+response pair; raw=%q", len(records), raw)
	}
	request, response := records[0], records[1]
	if request.Type != "ai.complete.request" || response.Type != "ai.complete.response" {
		t.Fatalf("record order/types = %q, %q", request.Type, response.Type)
	}
	for _, record := range records {
		assertRawRecordEnvelope(t, record)
	}
	if request.CallID != response.CallID {
		t.Fatalf("request callId %q != response callId %q", request.CallID, response.CallID)
	}
	if inner.SeenTaskRunID() != request.CallID {
		t.Fatalf("provider payload taskRun.id %q != pre-call raw callId %q", inner.SeenTaskRunID(), request.CallID)
	}
	rows := runs.Rows()
	if len(rows) != 1 {
		t.Fatalf("ai_task_runs rows = %d, want 1", len(rows))
	}
	if rows[0].ID != request.CallID {
		t.Fatalf("raw callId %q != ai_task_runs.id %q", request.CallID, rows[0].ID)
	}
	if got := request.Payload["messages"]; got == nil {
		t.Fatalf("request payload omitted messages: %#v", request.Payload)
	}
	if got := request.Payload["tools"]; got == nil {
		t.Fatalf("request payload omitted tools: %#v", request.Payload)
	}
	if got := request.Payload["toolChoice"]; got == nil {
		t.Fatalf("request payload omitted toolChoice: %#v", request.Payload)
	}
	if got := request.Payload["outputSchema"]; got == nil {
		t.Fatalf("request payload omitted outputSchema: %#v", request.Payload)
	}
	if response.Payload["content"] != `{"answer":"captured"}` || response.Payload["finishReason"] != "tool_calls" {
		t.Fatalf("response payload drift: %#v", response.Payload)
	}
}

func TestRawCompleteCapture_PreservesSchemaInvalidResponseAndProviderFailureRequest(t *testing.T) {
	t.Run("schema invalid response", func(t *testing.T) {
		capture := &syncCapture{}
		inner := &rawCaptureInner{
			response: aiclient.CompleteResponse{Content: `{"summary":"answer missing"}`, FinishReason: "stop"},
			meta:     rawSuccessMeta(),
		}
		wrap, _, _, _ := newRawCaptureDecorator(t, inner, capture)
		_, meta, err := wrap.Complete(context.Background(), "practice.chat.default", rawCapturePayload())
		if err == nil || meta.ErrorCode != sharederrors.CodeAiOutputInvalid {
			t.Fatalf("schema-invalid Complete error/meta = %v / %+v", err, meta)
		}
		raw, _ := capture.Snapshot()
		records := mustParseRawCompleteRecords(t, raw)
		if len(records) != 2 || records[1].Type != "ai.complete.response" {
			t.Fatalf("schema-invalid raw records = %#v", records)
		}
		if records[1].Payload["content"] != `{"summary":"answer missing"}` ||
			records[1].Payload["validationStatus"] != string(aiclient.ValidationStatusInvalid) ||
			records[1].Payload["errorCode"] != sharederrors.CodeAiOutputInvalid {
			t.Fatalf("schema-invalid response evidence drift: %#v", records[1].Payload)
		}
	})

	t.Run("provider error keeps pre-call request", func(t *testing.T) {
		capture := &syncCapture{}
		inner := &rawCaptureInner{
			meta: rawSuccessMeta(),
			err:  sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "provider diagnostic must not enter raw record", true),
		}
		wrap, _, _, _ := newRawCaptureDecorator(t, inner, capture)
		if _, _, err := wrap.Complete(context.Background(), "practice.chat.default", rawCapturePayload()); err == nil {
			t.Fatal("expected provider failure")
		}
		raw, _ := capture.Snapshot()
		records := mustParseRawCompleteRecords(t, raw)
		if len(records) == 0 || records[0].Type != "ai.complete.request" {
			t.Fatalf("provider failure lost synchronous request evidence: %#v", records)
		}
		if strings.Contains(string(raw), "provider diagnostic must not enter raw record") {
			t.Fatalf("raw capture leaked Go/provider error text: %s", raw)
		}
	})
}

func TestRawCompleteCapture_WriteFailureDoesNotChangeCompleteResultAndLogsEmptyStableEvent(t *testing.T) {
	capture := &failingRawCapture{}
	inner := &rawCaptureInner{
		response: aiclient.CompleteResponse{Content: `{"answer":"business result"}`, FinishReason: "stop"},
		meta:     rawSuccessMeta(),
	}
	wrap, runs, audit, logger := newRawCaptureDecorator(t, inner, capture)

	resp, meta, err := wrap.Complete(context.Background(), "practice.chat.default", rawCapturePayload())
	if err != nil {
		t.Fatalf("diagnostic recorder failure changed Complete error: %v", err)
	}
	if resp.Content != `{"answer":"business result"}` || resp.FinishReason != "stop" {
		t.Fatalf("diagnostic recorder failure changed Complete response: %+v", resp)
	}
	if meta.Provider != "unit-test-provider" || meta.ErrorCode != "" || meta.ValidationStatus != aiclient.ValidationStatusOK {
		t.Fatalf("diagnostic recorder failure changed Complete metadata: %+v", meta)
	}
	if inner.SeenTaskRunID() == "" || len(runs.Rows()) != 1 || len(audit.Rows()) != 1 {
		t.Fatalf("business observability did not complete: taskRun=%q runs=%d audits=%d", inner.SeenTaskRunID(), len(runs.Rows()), len(audit.Rows()))
	}
	if capture.Calls() != 2 {
		t.Fatalf("capture calls = %d, want request+response failures", capture.Calls())
	}

	failureEvents := 0
	for _, entry := range logger.Entries() {
		if entry.Event != observability.EventRawCaptureWriteFailed {
			continue
		}
		failureEvents++
		if !reflect.DeepEqual(entry.Fields, observability.LogFields{}) {
			t.Fatalf("raw capture failure event leaked structured fields: %+v", entry.Fields)
		}
	}
	if failureEvents != 2 {
		t.Fatalf("stable %q events = %d, want 2", observability.EventRawCaptureWriteFailed, failureEvents)
	}
}

func TestRawCompleteCapture_AppendsAcrossReopenAndSerializesOneHundredConcurrentCalls(t *testing.T) {
	root := canonicalTempDir(t)
	path := filepath.Join(root, "raw", "ai-raw.ndjson")
	open := func() *observability.FileRawIOCapture {
		capture, err := observability.OpenRawIOCapture(path)
		if err != nil {
			t.Fatalf("open raw capture: %v", err)
		}
		return capture
	}
	newInner := func() *rawCaptureInner {
		return &rawCaptureInner{
			response: aiclient.CompleteResponse{Content: `{"answer":"concurrent"}`, FinishReason: "stop"},
			meta:     rawSuccessMeta(),
		}
	}

	firstCapture := open()
	first, _, _, _ := newRawCaptureDecorator(t, newInner(), firstCapture)
	if _, _, err := first.Complete(context.Background(), "practice.chat.default", rawCapturePayload()); err != nil {
		t.Fatalf("first process Complete: %v", err)
	}
	if err := firstCapture.Close(); err != nil {
		t.Fatalf("close first raw capture: %v", err)
	}

	secondCapture := open()
	second, _, _, _ := newRawCaptureDecorator(t, newInner(), secondCapture)
	const concurrency = 100
	var wg sync.WaitGroup
	errs := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := second.Complete(context.Background(), "practice.chat.default", rawCapturePayload())
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Complete: %v", err)
		}
	}
	if err := secondCapture.Close(); err != nil {
		t.Fatalf("close second raw capture: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read raw capture: %v", err)
	}
	records := mustParseRawCompleteRecords(t, raw)
	if len(records) != 2*(concurrency+1) {
		t.Fatalf("raw record count = %d, want %d", len(records), 2*(concurrency+1))
	}
	byCallID := make(map[string]map[string]int, concurrency+1)
	for _, record := range records {
		assertRawRecordEnvelope(t, record)
		if byCallID[record.CallID] == nil {
			byCallID[record.CallID] = map[string]int{}
		}
		byCallID[record.CallID][record.Type]++
	}
	if len(byCallID) != concurrency+1 {
		t.Fatalf("unique call IDs = %d, want %d (append/restart collision)", len(byCallID), concurrency+1)
	}
	for callID, types := range byCallID {
		if types["ai.complete.request"] != 1 || types["ai.complete.response"] != 1 {
			t.Errorf("callId %s is not one complete pair: %#v", callID, types)
		}
	}
}

func TestRawCompleteCapture_OnlyCompleteIsCaptured(t *testing.T) {
	capture := &syncCapture{}
	inner := &rawCaptureInner{
		response: aiclient.CompleteResponse{Content: `{"answer":"complete only"}`},
		meta:     rawSuccessMeta(),
	}
	wrap, _, _, _ := newRawCaptureDecorator(t, inner, capture)

	if _, _, err := wrap.Transcribe(context.Background(), "practice.voice.stt.default", sampleTranscriptionInput()); err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if _, _, err := wrap.Synthesize(context.Background(), "practice.voice.tts.default", sampleSynthesisInput()); err != nil {
		t.Fatalf("Synthesize: %v", err)
	}
	stream, err := wrap.Stream(context.Background(), "practice.chat.default", rawCapturePayload())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	for range stream {
	}
	if raw, _ := capture.Snapshot(); len(raw) != 0 {
		t.Fatalf("Transcribe/Synthesize/Stream entered Complete raw recorder: %q", raw)
	}

	if _, _, err := wrap.Complete(context.Background(), "practice.chat.default", rawCapturePayload()); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	raw, _ := capture.Snapshot()
	if records := mustParseRawCompleteRecords(t, raw); len(records) != 2 {
		t.Fatalf("Complete raw records = %d, want 2", len(records))
	}
}

func TestRawCompleteCapture_PrivacyAuditParsesFieldsWithoutRejectingAuthorizedContent(t *testing.T) {
	capture := &syncCapture{}
	inner := &rawCaptureInner{
		response: aiclient.CompleteResponse{
			Content:      `{"answer":"The words Authorization, Cookie and apiKey are allowed user-facing prose."}`,
			FinishReason: "tool_calls",
			ToolCalls: []aiclient.ToolCall{{
				Name:      "privacy_word_probe",
				Arguments: json.RawMessage(`{"verbatim":"reasoning and Authorization are ordinary argument values"}`),
			}},
		},
		meta: rawSuccessMeta(),
	}
	wrap, _, _, _ := newRawCaptureDecorator(t, inner, capture)
	payload := rawCapturePayload()
	payload.Messages[1].Content = "Authorization Cookie apiKey reasoning must remain because this is authorized message content."
	payload.Tools = []aiclient.Tool{{
		Name:        "privacy_word_probe",
		Description: "Cookie is an ordinary domain word here.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"verbatim":{"type":"string","description":"Authorization is allowed as schema prose"}}}`),
	}}
	payload.ToolChoice = &aiclient.ToolChoice{Mode: aiclient.ToolChoiceModeTool, Name: "privacy_word_probe"}
	payload.Metadata.TaskRun.UserID = "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e"
	payload.Metadata.TaskRun.ResourceID = "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9f"
	payload.Metadata.TaskRun.RawResponseObjectKey = "private/raw/provider/object"

	if _, _, err := wrap.Complete(context.Background(), "practice.chat.default", payload); err != nil {
		t.Fatalf("Complete: %v", err)
	}
	raw, _ := capture.Snapshot()
	records := mustParseRawCompleteRecords(t, raw)
	if len(records) != 2 {
		t.Fatalf("raw record count = %d, want 2", len(records))
	}

	// The recorder is an explicit local exception: message content, response
	// content and tool arguments are intentionally present. Privacy therefore
	// audits parsed field names and internal identity values, not arbitrary
	// secret-looking substrings inside those allowlisted values.
	if !strings.Contains(string(raw), "Authorization Cookie apiKey reasoning") ||
		!strings.Contains(string(raw), "ordinary argument values") {
		t.Fatalf("authorized Complete values were unexpectedly redacted: %s", raw)
	}
	for _, forbiddenValue := range []string{
		payload.Metadata.TaskRun.UserID,
		payload.Metadata.TaskRun.ResourceID,
		payload.Metadata.TaskRun.RawResponseObjectKey,
	} {
		if strings.Contains(string(raw), forbiddenValue) {
			t.Errorf("internal task-run identity leaked into raw capture: %q", forbiddenValue)
		}
	}
	for _, record := range records {
		keys := map[string]struct{}{}
		collectJSONKeys(record.Payload, keys)
		for _, forbiddenKey := range []string{
			"authorization", "headers", "apikey", "cookie", "providerurl",
			"reasoning", "reasoningcontent", "userid", "resourceid", "taskrun",
			"rawresponseobjectkey", "audio", "transcript", "goerror", "errortext",
		} {
			if _, ok := keys[forbiddenKey]; ok {
				t.Errorf("forbidden raw capture field %q present in %s payload: %#v", forbiddenKey, record.Type, record.Payload)
			}
		}
	}
}

func parseRawCompleteRecords(raw []byte) ([]rawCompleteRecord, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	// Raw Complete payloads may legitimately exceed Scanner's small default.
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	var records []rawCompleteRecord
	line := 0
	for scanner.Scan() {
		line++
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			continue
		}
		var record rawCompleteRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("line %d is not one NDJSON object: %w", line, err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func mustParseRawCompleteRecords(t *testing.T, raw []byte) []rawCompleteRecord {
	t.Helper()
	records, err := parseRawCompleteRecords(raw)
	if err != nil {
		t.Fatalf("parse raw Complete NDJSON: %v; raw=%q", err, raw)
	}
	return records
}

func assertRawRecordEnvelope(t *testing.T, record rawCompleteRecord) {
	t.Helper()
	if record.RecordVersion != rawCompleteRecordVersion {
		t.Errorf("recordVersion = %q, want %q", record.RecordVersion, rawCompleteRecordVersion)
	}
	if record.Type != "ai.complete.request" && record.Type != "ai.complete.response" {
		t.Errorf("type = %q", record.Type)
	}
	if ok, _ := regexp.MatchString(idx.UUIDv7RegexExpr, record.CallID); !ok {
		t.Errorf("callId = %q, want UUIDv7", record.CallID)
	}
	capturedAt, err := time.Parse(time.RFC3339Nano, record.CapturedAt)
	if err != nil || capturedAt.Location() != time.UTC {
		t.Errorf("capturedAt = %q, want RFC3339Nano UTC", record.CapturedAt)
	}
	if record.ProfileName != "practice.chat.default" {
		t.Errorf("profileName = %q", record.ProfileName)
	}
	if record.Payload == nil {
		t.Errorf("payload must be an object")
	}
}

func collectJSONKeys(value any, out map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", ".", "").Replace(key))
			out[normalized] = struct{}{}
			collectJSONKeys(child, out)
		}
	case []any:
		for _, child := range typed {
			collectJSONKeys(child, out)
		}
	case json.RawMessage:
		var decoded any
		if json.Unmarshal(typed, &decoded) == nil {
			collectJSONKeys(decoded, out)
		}
	}
}

func canonicalTempDir(t *testing.T) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir symlinks: %v", err)
	}
	return resolved
}
