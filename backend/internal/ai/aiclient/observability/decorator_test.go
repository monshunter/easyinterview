package observability_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/providers/stub"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type memTaskRunWriter struct {
	mu   sync.Mutex
	rows []aiclient.AITaskRunRow
}

func (m *memTaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, row)
	return nil
}

func (m *memTaskRunWriter) Rows() []aiclient.AITaskRunRow {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]aiclient.AITaskRunRow{}, m.rows...)
}

type failingTaskRunWriter struct {
	err error
}

func (f failingTaskRunWriter) WriteAITaskRun(_ context.Context, _ aiclient.AITaskRunRow) error {
	return f.err
}

type memAuditWriter struct {
	mu   sync.Mutex
	rows []aiclient.AuditEventRow
}

func (m *memAuditWriter) WriteAuditEvent(_ context.Context, row aiclient.AuditEventRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, row)
	return nil
}

func (m *memAuditWriter) Rows() []aiclient.AuditEventRow {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]aiclient.AuditEventRow{}, m.rows...)
}

type staticResolver map[string]*aiclient.ModelProfile

func (r staticResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	p, ok := r[name]
	if !ok {
		return nil, errors.New("not found: " + name)
	}
	return p, nil
}

func newTestStack(t *testing.T) (
	aiclient.AIClient,
	*observability.InMemoryRegistry,
	*observability.MemoryLogger,
	*memTaskRunWriter,
	*memAuditWriter,
) {
	t.Helper()
	stubProv, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	resolver := staticResolver{
		"practice.followup.default": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default: aiclient.ProviderConfig{
				Provider: stub.Name,
				Model:    "stub-chat-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
		"review.embed.default": {
			Name:     "review.embed.default",
			TaskType: aiclient.TaskTypeEmbed,
			Default: aiclient.ProviderConfig{
				Provider: stub.Name,
				Model:    "stub-embed-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
	}
	inner, err := aiclient.New(
		aiclient.Config{AppEnv: aiclient.AppEnvTest},
		aiclient.WithStubAllowed(true),
		aiclient.WithProfileResolver(resolver),
		aiclient.WithProvider(stubProv),
	)
	if err != nil {
		t.Fatalf("aiclient.New: %v", err)
	}
	registry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	runWriter := &memTaskRunWriter{}
	auditWriter := &memAuditWriter{}

	wrapped, err := observability.New(inner,
		observability.WithRegisterer(registry),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(auditWriter),
		observability.WithProfileResolver(resolver),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	return wrapped, registry, logger, runWriter, auditWriter
}

func samplePayload() aiclient.CompletePayload {
	return aiclient.CompletePayload{
		Messages: []aiclient.Message{
			{Role: "system", Content: "you are an interviewer."},
			{Role: "user", Content: "tell me about yourself."},
		},
		Metadata: aiclient.CallMetadata{
			FeatureKey:    "practice.followup",
			PromptVersion: "p1",
			RubricVersion: "r1",
			Language:      "en",
			TaskRun: aiclient.AITaskRunContext{
				TaskType:     aiclient.AITaskRunTaskFollowupGenerate,
				ResourceType: aiclient.AITaskRunResourceTargetJob,
				ResourceID:   "018f0d59-0f7a-7b58-9f2f-65cc4d8e8b1d",
			},
		},
	}
}

func TestDecorator_AllSevenMetricFamiliesRegistered(t *testing.T) {
	_, registry, _, _, _ := newTestStack(t)
	expectedCounters := []string{
		observability.MetricRunsTotal,
		observability.MetricInputTokensTotal,
		observability.MetricOutputTokensTotal,
		observability.MetricCostUSDTotal,
		observability.MetricOutputValidationFailures,
		observability.MetricFallbackTotal,
	}
	for _, name := range expectedCounters {
		if !registry.CounterRegistered(name) {
			t.Errorf("expected counter %q to be registered", name)
		}
	}
	if !registry.HistogramRegistered(observability.MetricLatencySeconds) {
		t.Errorf("expected histogram %q to be registered", observability.MetricLatencySeconds)
	}
}

func TestDecorator_SuccessIncrementsRunsAndLogsCompleted(t *testing.T) {
	wrap, registry, logger, runs, audit := newTestStack(t)

	_, meta, err := wrap.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if meta.Provider != stub.Name {
		t.Fatalf("expected provider=%q, got %q", stub.Name, meta.Provider)
	}

	successLabels := []string{stub.Name, "stub", "practice.followup.default", "unknown", string(aiclient.TaskTypeChat), "en", "success"}
	if got := registry.CounterValue(observability.MetricRunsTotal, successLabels...); got != 1 {
		t.Errorf("ai_task_runs_total: expected 1, got %v", got)
	}
	if got := registry.CounterValue(observability.MetricInputTokensTotal, successLabels...); got == 0 {
		t.Errorf("ai_task_input_tokens_total: expected non-zero, got %v", got)
	}
	if got := registry.CounterValue(observability.MetricOutputTokensTotal, successLabels...); got == 0 {
		t.Errorf("ai_task_output_tokens_total: expected non-zero, got %v", got)
	}
	if got := registry.CounterValue(observability.MetricOutputValidationFailures, successLabels...); got != 0 {
		t.Errorf("validation failures should be 0 on success, got %v", got)
	}
	// fallback counter uses extended labels; with FallbackChain length == 1 it shouldn't be incremented.
	for _, labels := range registry.CounterLabelValues(observability.MetricFallbackTotal) {
		t.Errorf("fallback counter unexpectedly incremented: %v", labels)
	}
	if observations := registry.HistogramObservations(observability.MetricLatencySeconds, successLabels...); len(observations) != 1 {
		t.Errorf("latency histogram: expected 1 observation, got %d", len(observations))
	}

	entries := logger.Entries()
	if len(entries) != 1 || entries[0].Event != observability.EventTaskCompleted {
		t.Fatalf("expected single ai.task.completed entry, got %+v", entries)
	}
	fields := entries[0].Fields
	if fields.Provider != stub.Name || fields.ModelProfileName != "practice.followup.default" {
		t.Errorf("log fields incomplete: %+v", fields)
	}

	rows := runs.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 ai_task_runs row, got %d", len(rows))
	}
	if rows[0].Provider != stub.Name || rows[0].ModelProfileName != "practice.followup.default" {
		t.Errorf("ai_task_runs row missing fields: %+v", rows[0])
	}
	if rows[0].ID == "" {
		t.Fatalf("ai_task_runs row missing id: %+v", rows[0])
	}
	if rows[0].TaskType != aiclient.AITaskRunTaskFollowupGenerate {
		t.Fatalf("expected B4 task_type=%q, got %q", aiclient.AITaskRunTaskFollowupGenerate, rows[0].TaskType)
	}
	if rows[0].ResourceType != aiclient.AITaskRunResourceTargetJob || rows[0].ResourceID == "" {
		t.Fatalf("ai_task_runs row missing resource identity: %+v", rows[0])
	}
	if rows[0].Status != aiclient.AITaskRunStatusSuccess {
		t.Fatalf("expected status=%q, got %q", aiclient.AITaskRunStatusSuccess, rows[0].Status)
	}
	if rows[0].StartedAt.IsZero() || rows[0].CompletedAt.IsZero() {
		t.Fatalf("ai_task_runs row missing timestamps: %+v", rows[0])
	}
	if rows[0].CompletedAt.Before(rows[0].StartedAt) {
		t.Fatalf("completed_at before started_at: %+v", rows[0])
	}
	if rows[0].Metadata.PromptHash == "" || rows[0].Metadata.ResponseHash == "" {
		t.Fatalf("ai_task_runs metadata missing hash summary: %+v", rows[0].Metadata)
	}

	auditRows := audit.Rows()
	if len(auditRows) != 1 || auditRows[0].Action != "ai.call" {
		t.Fatalf("expected 1 audit row with action=ai.call, got %+v", auditRows)
	}
	if auditRows[0].Metadata.ProfileName != "practice.followup.default" {
		t.Errorf("audit profile name mismatch: %q", auditRows[0].Metadata.ProfileName)
	}
	if auditRows[0].Metadata.PromptHash == "" || auditRows[0].Metadata.ResponseHash == "" {
		t.Errorf("audit hashes empty: %+v", auditRows[0].Metadata)
	}
	if auditRows[0].Metadata.PromptCharLength == 0 || auditRows[0].Metadata.ResponseCharLength == 0 {
		t.Errorf("audit char lengths zero: %+v", auditRows[0].Metadata)
	}
}

func TestDecorator_AITaskRunWriterFailureReturned(t *testing.T) {
	stubProv, err := stub.New(stub.WithAppEnv(aiclient.AppEnvTest))
	if err != nil {
		t.Fatalf("stub.New: %v", err)
	}
	resolver := staticResolver{
		"practice.followup.default": {
			Name:     "practice.followup.default",
			TaskType: aiclient.TaskTypeChat,
			Default: aiclient.ProviderConfig{
				Provider: stub.Name,
				Model:    "stub-chat-1",
			},
			TimeoutMs: 5000,
			Version:   "1.0.0",
		},
	}
	inner, err := aiclient.New(
		aiclient.Config{AppEnv: aiclient.AppEnvTest},
		aiclient.WithStubAllowed(true),
		aiclient.WithProfileResolver(resolver),
		aiclient.WithProvider(stubProv),
	)
	if err != nil {
		t.Fatalf("aiclient.New: %v", err)
	}
	wrap, err := observability.New(inner,
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(observability.NewMemoryLogger()),
		observability.WithAITaskRunWriter(failingTaskRunWriter{err: errors.New("db down")}),
		observability.WithAuditEventWriter(&memAuditWriter{}),
		observability.WithProfileResolver(resolver),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}

	_, _, err = wrap.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err == nil {
		t.Fatalf("expected writer failure to be returned")
	}
	if !strings.Contains(err.Error(), "write ai_task_runs") || !strings.Contains(err.Error(), "db down") {
		t.Fatalf("expected ai_task_runs write context in error, got: %v", err)
	}
}

func TestDecorator_FailureIncrementsFailureLogsFailed(t *testing.T) {
	wrap, registry, logger, _, _ := newTestStack(t)

	_, _, err := wrap.Complete(context.Background(), "practice.followup.default", aiclient.CompletePayload{})
	if err == nil {
		t.Fatalf("expected error for empty messages")
	}
	failureLabels := []string{"unknown", "unknown", "practice.followup.default", "unknown", "unknown", "unknown", "failure"}
	if got := registry.CounterValue(observability.MetricRunsTotal, failureLabels...); got != 1 {
		t.Errorf("expected runs_total=1 on failure, got %v", got)
	}
	if got := registry.CounterValue(observability.MetricOutputValidationFailures, failureLabels...); got != 1 {
		t.Errorf("expected validation_failures_total=1, got %v", got)
	}

	events := []string{}
	for _, e := range logger.Entries() {
		events = append(events, e.Event)
	}
	hasValidationFailed := false
	for _, e := range events {
		if e == observability.EventOutputValidationFailed {
			hasValidationFailed = true
		}
	}
	if !hasValidationFailed {
		t.Errorf("expected ai.output.validation_failed log; got %v", events)
	}
}

func TestDecorator_FallbackChainTriggersFallbackCounterAndLog(t *testing.T) {
	registry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	runs := &memTaskRunWriter{}
	audit := &memAuditWriter{}
	resolver := staticResolver{}

	innerStub := &fallbackInner{
		meta: aiclient.AICallMeta{
			Provider:         "openai_compatible",
			ModelFamily:      "openai/gpt-4-turbo",
			ModelID:          "gpt-4-turbo",
			TaskType:         aiclient.TaskTypeChat,
			ModelProfileName: "practice.followup.default",
			Language:         "en",
			InputTokens:      10,
			OutputTokens:     20,
			LatencyMs:        50,
			FallbackChain:    []string{"openai/gpt-4", "anthropic/claude-3"},
			Route:            "practice.followup",
			ValidationStatus: aiclient.ValidationStatusOK,
		},
	}
	wrap, err := observability.New(innerStub,
		observability.WithRegisterer(registry),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runs),
		observability.WithAuditEventWriter(audit),
		observability.WithProfileResolver(resolver),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}

	_, _, err = wrap.Complete(context.Background(), "practice.followup.default", samplePayload())
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	// fallback counter has extended labels — search by metric name.
	tuples := registry.CounterLabelValues(observability.MetricFallbackTotal)
	if len(tuples) != 1 {
		t.Fatalf("expected 1 fallback counter tuple, got %v", tuples)
	}
	if got := registry.CounterValue(observability.MetricFallbackTotal, tuples[0]...); got != 1 {
		t.Fatalf("expected fallback counter=1, got %v", got)
	}

	gotFallbackLog := false
	for _, e := range logger.Entries() {
		if e.Event == observability.EventTaskFallback {
			gotFallbackLog = true
		}
	}
	if !gotFallbackLog {
		t.Errorf("expected ai.task.fallback log entry")
	}
}

type fallbackInner struct {
	meta    aiclient.AICallMeta
	content string
}

func (f *fallbackInner) Complete(_ context.Context, _ string, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	content := f.content
	if content == "" {
		content = "ok"
	}
	return aiclient.CompleteResponse{Content: content}, f.meta, nil
}
func (f *fallbackInner) Embed(_ context.Context, _ string, _ aiclient.EmbedInput) (aiclient.EmbedResponse, aiclient.AICallMeta, error) {
	return aiclient.EmbedResponse{}, f.meta, nil
}
func (f *fallbackInner) Stream(_ context.Context, _ string, _ aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	ch := make(chan aiclient.AIStreamEvent, 1)
	close(ch)
	return ch, nil
}

func TestDecorator_OutputSchemaInvalidEmitsAIOutputInvalid(t *testing.T) {
	wrap, registry, logger, _, _ := newTestStack(t)

	payload := samplePayload()
	payload.Metadata.OutputSchema = json.RawMessage(`{"type":"object"}`)

	_, meta, err := wrap.Complete(context.Background(), "practice.followup.default", payload)
	if err == nil {
		t.Fatalf("expected error for invalid JSON content")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
	if meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected ValidationStatusInvalid, got %q", meta.ValidationStatus)
	}

	tuples := registry.CounterLabelValues(observability.MetricOutputValidationFailures)
	if len(tuples) == 0 {
		t.Fatalf("expected validation_failures_total to have at least one tuple")
	}

	gotEvent := false
	for _, e := range logger.Entries() {
		if e.Event == observability.EventOutputValidationFailed {
			gotEvent = true
		}
	}
	if !gotEvent {
		t.Errorf("expected ai.output.validation_failed log event")
	}
}

func TestDecorator_OutputSchemaRequiredFieldMismatchEmitsAIOutputInvalid(t *testing.T) {
	registry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	runs := &memTaskRunWriter{}
	audit := &memAuditWriter{}
	inner := &fallbackInner{
		content: `{"summary":"present-but-answer-missing"}`,
		meta: aiclient.AICallMeta{
			Provider:            stub.Name,
			ModelFamily:         "stub",
			ModelID:             "stub-chat-1",
			TaskType:            aiclient.TaskTypeChat,
			ModelProfileName:    "practice.followup.default",
			ModelProfileVersion: "1.0.0",
			Language:            "en",
			InputTokens:         10,
			OutputTokens:        20,
			LatencyMs:           5,
			ValidationStatus:    aiclient.ValidationStatusOK,
		},
	}
	wrap, err := observability.New(inner,
		observability.WithRegisterer(registry),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runs),
		observability.WithAuditEventWriter(audit),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}

	payload := samplePayload()
	payload.Metadata.OutputSchema = json.RawMessage(`{"type":"object","required":["answer"],"properties":{"answer":{"type":"string"}}}`)

	_, meta, err := wrap.Complete(context.Background(), "practice.followup.default", payload)
	if err == nil {
		t.Fatalf("expected schema mismatch to return error")
	}
	var apiErr *sharederrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %v", err)
	}
	if meta.ValidationStatus != aiclient.ValidationStatusInvalid {
		t.Fatalf("expected ValidationStatusInvalid, got %q", meta.ValidationStatus)
	}
	if len(registry.CounterLabelValues(observability.MetricOutputValidationFailures)) == 0 {
		t.Fatalf("expected validation failure counter to increment")
	}
}

func TestDecorator_ConstructorRequiresAllInjectables(t *testing.T) {
	registry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	cases := []struct {
		name string
		opts []observability.Option
	}{
		{"missing-registry", []observability.Option{
			observability.WithLogger(logger),
			observability.WithAITaskRunWriter(&memTaskRunWriter{}),
			observability.WithAuditEventWriter(&memAuditWriter{}),
		}},
		{"missing-logger", []observability.Option{
			observability.WithRegisterer(registry),
			observability.WithAITaskRunWriter(&memTaskRunWriter{}),
			observability.WithAuditEventWriter(&memAuditWriter{}),
		}},
		{"missing-run-writer", []observability.Option{
			observability.WithRegisterer(registry),
			observability.WithLogger(logger),
			observability.WithAuditEventWriter(&memAuditWriter{}),
		}},
		{"missing-audit-writer", []observability.Option{
			observability.WithRegisterer(registry),
			observability.WithLogger(logger),
			observability.WithAITaskRunWriter(&memTaskRunWriter{}),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := observability.New(&fallbackInner{}, tc.opts...); err == nil {
				t.Errorf("expected error for %s", tc.name)
			}
		})
	}
}
