package observability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

// Wrap decorates an inner aiclient.AIClient with the spec §2.1 / §4.3
// observability obligations: 7 metric families, 4 log event names,
// ai_task_runs row, audit_events row. Spec D-6 forbids business code
// from calling the inner client directly; the decorator is the
// enforcement seam.
type Wrap struct {
	inner       aiclient.AIClient
	resolver    aiclient.ProfileResolver
	registry    Registerer
	metrics     metricSet
	logger      Logger
	runWriter   aiclient.AITaskRunWriter
	auditWriter aiclient.AuditEventWriter
	now         func() time.Time
}

// Options configure New.
type options struct {
	resolver    aiclient.ProfileResolver
	registry    Registerer
	logger      Logger
	runWriter   aiclient.AITaskRunWriter
	auditWriter aiclient.AuditEventWriter
	now         func() time.Time
}

// Option mutates Wrap construction.
type Option func(*options)

// WithRegisterer wires a metrics registry. Required.
func WithRegisterer(r Registerer) Option { return func(o *options) { o.registry = r } }

// WithLogger wires a structured logger. Required.
func WithLogger(l Logger) Option { return func(o *options) { o.logger = l } }

// WithAITaskRunWriter wires the ai_task_runs row writer.
func WithAITaskRunWriter(w aiclient.AITaskRunWriter) Option {
	return func(o *options) { o.runWriter = w }
}

// WithAuditEventWriter wires the audit_events row writer.
func WithAuditEventWriter(w aiclient.AuditEventWriter) Option {
	return func(o *options) { o.auditWriter = w }
}

// WithProfileResolver wires the profile resolver used to derive
// task_type and route labels for failures that happen before the
// inner client could populate the meta. When unset, label values fall
// back to "unknown".
func WithProfileResolver(r aiclient.ProfileResolver) Option {
	return func(o *options) { o.resolver = r }
}

// WithNow injects a clock for deterministic latency tests.
func WithNow(now func() time.Time) Option { return func(o *options) { o.now = now } }

// New constructs a decorator. inner, registry, logger, runWriter,
// auditWriter must all be non-nil; the decorator panics during boot
// rather than silently dropping observability obligations.
func New(inner aiclient.AIClient, opts ...Option) (*Wrap, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if inner == nil {
		return nil, errors.New("observability: inner AIClient is nil")
	}
	if o.registry == nil {
		return nil, errors.New("observability: Registerer is required")
	}
	if o.logger == nil {
		return nil, errors.New("observability: Logger is required")
	}
	if o.runWriter == nil {
		return nil, errors.New("observability: AITaskRunWriter is required")
	}
	if o.auditWriter == nil {
		return nil, errors.New("observability: AuditEventWriter is required")
	}
	if o.now == nil {
		o.now = time.Now
	}
	return &Wrap{
		inner:       inner,
		resolver:    o.resolver,
		registry:    o.registry,
		metrics:     registerMetrics(o.registry),
		logger:      o.logger,
		runWriter:   o.runWriter,
		auditWriter: o.auditWriter,
		now:         o.now,
	}, nil
}

// Complete implements aiclient.AIClient.
func (w *Wrap) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	start := w.now()
	resp, meta, err := w.inner.Complete(ctx, profileName, payload)
	latencyMs := w.now().Sub(start).Milliseconds()
	if meta.LatencyMs == 0 {
		meta.LatencyMs = latencyMs
	}

	// Apply validateOutput when the caller supplied a JSON schema. Plan 001
	// validates the baseline type / required / properties subset and leaves
	// full JSON Schema support to future plans.
	if err == nil && len(payload.Metadata.OutputSchema) > 0 {
		if vErr := validateOutputSchema(payload.Metadata.OutputSchema, resp.Content); vErr != nil {
			err = sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "output failed schema validation: "+vErr.Error(), false)
			meta.ValidationStatus = aiclient.ValidationStatusInvalid
			meta.ErrorCode = sharederrors.CodeAiOutputInvalid
		}
	}

	w.recordCompleteCall(ctx, profileName, payload, resp.Content, meta, err)
	return resp, meta, err
}

// Embed implements aiclient.AIClient.
func (w *Wrap) Embed(ctx context.Context, profileName string, input aiclient.EmbedInput) (aiclient.EmbedResponse, aiclient.AICallMeta, error) {
	start := w.now()
	resp, meta, err := w.inner.Embed(ctx, profileName, input)
	latencyMs := w.now().Sub(start).Milliseconds()
	if meta.LatencyMs == 0 {
		meta.LatencyMs = latencyMs
	}
	w.recordEmbedCall(ctx, profileName, input, resp, meta, err)
	return resp, meta, err
}

// Stream implements aiclient.AIClient. Plan 001 emits one decorator
// record on the terminal `done` event; plan 002's full SSE/chunked
// consumer will call recordCompleteCall on the consolidated meta.
func (w *Wrap) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	innerCh, err := w.inner.Stream(ctx, profileName, payload)
	if err != nil {
		return nil, err
	}
	out := make(chan aiclient.AIStreamEvent, 1)
	go func() {
		defer close(out)
		for ev := range innerCh {
			out <- ev
			if ev.Type == aiclient.StreamEventDone && ev.Meta != nil {
				w.recordCompleteCall(ctx, profileName, payload, "", *ev.Meta, nil)
			}
			if ev.Type == aiclient.StreamEventError {
				w.recordCompleteCall(ctx, profileName, payload, "", aiclient.AICallMeta{ErrorCode: ev.ErrorCode, ModelProfileName: profileName}, fmt.Errorf("stream error: %s", ev.ErrorCode))
			}
		}
	}()
	return out, nil
}

func (w *Wrap) recordCompleteCall(ctx context.Context, profileName string, payload aiclient.CompletePayload, responseContent string, meta aiclient.AICallMeta, err error) {
	w.recordMetricsAndLog(meta, err)
	_ = w.runWriter.WriteAITaskRun(ctx, AITaskRunRowFromMeta(meta))
	_ = w.auditWriter.WriteAuditEvent(ctx, w.buildAuditRow(profileName, joinMessages(payload.Messages), responseContent))
}

func (w *Wrap) recordEmbedCall(ctx context.Context, profileName string, input aiclient.EmbedInput, resp aiclient.EmbedResponse, meta aiclient.AICallMeta, err error) {
	w.recordMetricsAndLog(meta, err)
	_ = w.runWriter.WriteAITaskRun(ctx, AITaskRunRowFromMeta(meta))
	_ = w.auditWriter.WriteAuditEvent(ctx, w.buildAuditRow(profileName, strings.Join(input.Texts, "\n"), summarizeVectors(resp.Vectors)))
}

func (w *Wrap) recordMetricsAndLog(meta aiclient.AICallMeta, err error) {
	labels := w.standardLabels(meta, err)

	w.metrics.runs.Inc(labels...)
	w.metrics.latency.Observe(float64(meta.LatencyMs)/1000.0, labels...)
	if meta.InputTokens > 0 {
		w.metrics.inputTokens.Add(float64(meta.InputTokens), labels...)
	}
	if meta.OutputTokens > 0 {
		w.metrics.outputTokens.Add(float64(meta.OutputTokens), labels...)
	}
	if meta.CostUSDMicros > 0 {
		w.metrics.cost.Add(float64(meta.CostUSDMicros), labels...)
	}

	if err == nil && meta.ValidationStatus == aiclient.ValidationStatusInvalid && meta.ErrorCode == sharederrors.CodeAiOutputInvalid {
		// validateOutput failure recorded as success-with-failure on the call.
	}

	if meta.ValidationStatus == aiclient.ValidationStatusInvalid && meta.ErrorCode == sharederrors.CodeAiOutputInvalid {
		w.metrics.validationFailure.Inc(labels...)
		w.logger.Log(EventOutputValidationFailed, w.buildLogFields(meta))
	}

	if len(meta.FallbackChain) > 1 {
		w.metrics.fallback.Inc(append(labels, modelFamily(meta.FallbackChain[0]), modelFamily(meta.FallbackChain[len(meta.FallbackChain)-1]))...)
		w.logger.Log(EventTaskFallback, w.buildLogFields(meta))
	}

	switch {
	case err != nil && meta.ErrorCode != sharederrors.CodeAiOutputInvalid:
		w.logger.Log(EventTaskFailed, w.buildLogFields(meta))
	case err == nil && meta.ValidationStatus != aiclient.ValidationStatusInvalid:
		w.logger.Log(EventTaskCompleted, w.buildLogFields(meta))
	}
}

func (w *Wrap) standardLabels(meta aiclient.AICallMeta, err error) []string {
	result := "success"
	switch {
	case err != nil:
		result = "failure"
	case meta.ValidationStatus == aiclient.ValidationStatusInvalid:
		result = "failure"
	case len(meta.FallbackChain) > 1:
		result = "fallback"
	}
	return []string{
		emptyOrUnknown(meta.Provider),
		emptyOrUnknown(meta.ModelFamily),
		emptyOrUnknown(meta.ModelProfileName),
		emptyOrUnknown(meta.Route),
		emptyOrUnknown(string(meta.TaskType)),
		emptyOrUnknown(meta.Language),
		result,
	}
}

func (w *Wrap) buildLogFields(meta aiclient.AICallMeta) LogFields {
	return LogFields{
		Provider:            meta.Provider,
		ModelID:             meta.ModelID,
		ModelProfileName:    meta.ModelProfileName,
		ModelProfileVersion: meta.ModelProfileVersion,
		PromptVersion:       meta.PromptVersion,
		RubricVersion:       meta.RubricVersion,
		TaskType:            string(meta.TaskType),
		Language:            meta.Language,
		InputTokens:         meta.InputTokens,
		OutputTokens:        meta.OutputTokens,
		CostUSDMicros:       meta.CostUSDMicros,
		LatencyMs:           meta.LatencyMs,
		FallbackChain:       meta.FallbackChain,
		Route:               meta.Route,
		ValidationStatus:    string(meta.ValidationStatus),
		ErrorCode:           meta.ErrorCode,
	}
}

// AITaskRunRowFromMeta builds the typed ai_task_runs row directly from
// AICallMeta. The mapping is intentionally exposed so future plans can
// reuse it.
func AITaskRunRowFromMeta(meta aiclient.AICallMeta) aiclient.AITaskRunRow {
	return aiclient.AITaskRunRow{
		Provider:            meta.Provider,
		ModelFamily:         meta.ModelFamily,
		ModelID:             meta.ModelID,
		TaskType:            meta.TaskType,
		PromptVersion:       meta.PromptVersion,
		RubricVersion:       meta.RubricVersion,
		ModelProfileName:    meta.ModelProfileName,
		ModelProfileVersion: meta.ModelProfileVersion,
		Language:            meta.Language,
		InputTokens:         meta.InputTokens,
		OutputTokens:        meta.OutputTokens,
		CostUSDMicros:       meta.CostUSDMicros,
		LatencyMs:           meta.LatencyMs,
		FallbackChain:       meta.FallbackChain,
		Route:               meta.Route,
		ValidationStatus:    meta.ValidationStatus,
		ErrorCode:           meta.ErrorCode,
	}
}

func (w *Wrap) buildAuditRow(profileName, prompt, response string) aiclient.AuditEventRow {
	return aiclient.AuditEventRow{
		Action: "ai.call",
		Metadata: aiclient.AuditMetadata{
			PromptHash:         hashHex(prompt),
			ResponseHash:       hashHex(response),
			PromptCharLength:   len(prompt),
			ResponseCharLength: len(response),
			ProfileName:        profileName,
		},
	}
}

func emptyOrUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

func hashHex(s string) string {
	if s == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func joinMessages(messages []aiclient.Message) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.Role + ":" + m.Content
	}
	return strings.Join(parts, "\n")
}

func summarizeVectors(vectors [][]float64) string {
	// We only feed the audit pipeline an opaque length-derived summary;
	// the actual vector values must not be hashed/logged because they
	// can be inverted to reveal the input.
	return fmt.Sprintf("vectors:%d", len(vectors))
}

type outputSchema struct {
	Type       string                  `json:"type"`
	Required   []string                `json:"required"`
	Properties map[string]outputSchema `json:"properties"`
	Items      *outputSchema           `json:"items"`
}

func validateOutputSchema(schemaRaw json.RawMessage, content string) error {
	if content == "" {
		return errors.New("empty content")
	}
	var schema outputSchema
	if err := json.Unmarshal(schemaRaw, &schema); err != nil {
		return fmt.Errorf("parse output_schema: %w", err)
	}
	var v any
	dec := json.NewDecoder(strings.NewReader(content))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return err
	}
	return validateAgainstSchema(schema, v, "$")
}

func validateAgainstSchema(schema outputSchema, value any, path string) error {
	if schema.Type != "" && !matchesSchemaType(schema.Type, value) {
		return fmt.Errorf("%s expected %s", path, schema.Type)
	}

	if len(schema.Required) > 0 || len(schema.Properties) > 0 {
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s expected object", path)
		}
		for _, key := range schema.Required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("%s missing required field %q", path, key)
			}
		}
		for key, childSchema := range schema.Properties {
			child, ok := obj[key]
			if !ok {
				continue
			}
			if err := validateAgainstSchema(childSchema, child, path+"."+key); err != nil {
				return err
			}
		}
	}

	if schema.Items != nil {
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s expected array", path)
		}
		for i, item := range items {
			if err := validateAgainstSchema(*schema.Items, item, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func matchesSchemaType(schemaType string, value any) bool {
	switch schemaType {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(json.Number)
		return ok
	case "integer":
		n, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := n.Int64()
		return err == nil
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

func modelFamily(provider string) string {
	if provider == "" {
		return ""
	}
	if i := strings.LastIndex(provider, "/"); i > 0 {
		return provider[:i]
	}
	if i := strings.LastIndex(provider, "-"); i > 0 {
		return provider[:i]
	}
	return provider
}
