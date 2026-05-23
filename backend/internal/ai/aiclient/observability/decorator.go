package observability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
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
// capability and route labels for failures that happen before the
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
	completed := w.now()
	latencyMs := completed.Sub(start).Milliseconds()
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
	meta = enrichErrorMeta(meta, err)

	recordErr := w.recordCompleteCall(ctx, profileName, payload, resp.Content, meta, start, completed, err)
	return resp, meta, joinRecordError(err, recordErr)
}

// Transcribe implements aiclient.AIClient.
func (w *Wrap) Transcribe(ctx context.Context, profileName string, input aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	start := w.now()
	resp, meta, err := w.inner.Transcribe(ctx, profileName, input)
	completed := w.now()
	latencyMs := completed.Sub(start).Milliseconds()
	if meta.LatencyMs == 0 {
		meta.LatencyMs = latencyMs
	}
	meta = enrichErrorMeta(meta, err)
	recordErr := w.recordTranscribeCall(ctx, profileName, input, resp, meta, start, completed, err)
	return resp, meta, joinRecordError(err, recordErr)
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
				now := w.now()
				_ = w.recordCompleteCall(ctx, profileName, payload, "", *ev.Meta, now, now, nil)
			}
			if ev.Type == aiclient.StreamEventError {
				now := w.now()
				_ = w.recordCompleteCall(ctx, profileName, payload, "", aiclient.AICallMeta{ErrorCode: ev.ErrorCode, ModelProfileName: profileName}, now, now, fmt.Errorf("stream error: %s", ev.ErrorCode))
			}
		}
	}()
	return out, nil
}

// Synthesize implements aiclient.AIClient.
func (w *Wrap) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	start := w.now()
	resp, meta, err := w.inner.Synthesize(ctx, profileName, input)
	completed := w.now()
	latencyMs := completed.Sub(start).Milliseconds()
	if meta.LatencyMs == 0 {
		meta.LatencyMs = latencyMs
	}
	meta = enrichErrorMeta(meta, err)
	recordErr := w.recordSynthesizeCall(ctx, profileName, input, resp, meta, start, completed, err)
	return resp, meta, joinRecordError(err, recordErr)
}

func enrichErrorMeta(meta aiclient.AICallMeta, err error) aiclient.AICallMeta {
	if err == nil || meta.ErrorCode != "" {
		return meta
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		meta.ErrorCode = apiErr.Code
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
	}
	return meta
}

func (w *Wrap) recordCompleteCall(ctx context.Context, profileName string, payload aiclient.CompletePayload, responseContent string, meta aiclient.AICallMeta, start, completed time.Time, err error) error {
	meta = w.enrichMeta(profileName, meta, payload.Metadata)
	w.recordMetricsAndLog(meta, err)
	auditRow := w.buildAuditRow(profileName, joinMessages(payload.Messages), responseContent)
	return errors.Join(
		w.writeTaskRun(ctx, meta, payload.Metadata.TaskRun, auditRow.Metadata, start, completed, err),
		w.writeAuditEvent(ctx, auditRow),
	)
}

func (w *Wrap) recordTranscribeCall(ctx context.Context, profileName string, input aiclient.TranscriptionInput, resp aiclient.TranscriptionResponse, meta aiclient.AICallMeta, start, completed time.Time, err error) error {
	meta = w.enrichMeta(profileName, meta, input.Metadata)
	w.recordMetricsAndLog(meta, err)
	auditRow := w.buildAuditRow(profileName, audioAuditSummary(input), resp.Text)
	auditRow.Metadata.PromptCharLength = len(input.Audio)
	auditRow.Metadata.ResponseCharLength = len(resp.Text)
	return errors.Join(
		w.writeTaskRun(ctx, meta, input.Metadata.TaskRun, auditRow.Metadata, start, completed, err),
		w.writeAuditEvent(ctx, auditRow),
	)
}

func (w *Wrap) recordSynthesizeCall(ctx context.Context, profileName string, input aiclient.SynthesisInput, resp aiclient.SynthesisResponse, meta aiclient.AICallMeta, start, completed time.Time, err error) error {
	meta = w.enrichMeta(profileName, meta, input.Metadata)
	w.recordMetricsAndLog(meta, err)
	auditRow := w.buildAuditRow(profileName, synthesisAuditSummary(input), synthesisResponseAuditSummary(resp))
	auditRow.Metadata.PromptCharLength = len(input.Text)
	auditRow.Metadata.ResponseCharLength = len(resp.Audio)
	return errors.Join(
		w.writeTaskRun(ctx, meta, input.Metadata.TaskRun, auditRow.Metadata, start, completed, err),
		w.writeAuditEvent(ctx, auditRow),
	)
}

func (w *Wrap) writeTaskRun(ctx context.Context, meta aiclient.AICallMeta, taskCtx aiclient.AITaskRunContext, metadata aiclient.AuditMetadata, start, completed time.Time, callErr error) error {
	row, err := AITaskRunRowFromMeta(meta, taskCtx, metadata, start, completed, callErr)
	if err != nil {
		return err
	}
	if err := w.runWriter.WriteAITaskRun(ctx, row); err != nil {
		return fmt.Errorf("observability: write ai_task_runs: %w", err)
	}
	return nil
}

func (w *Wrap) writeAuditEvent(ctx context.Context, row aiclient.AuditEventRow) error {
	if err := w.auditWriter.WriteAuditEvent(ctx, row); err != nil {
		return fmt.Errorf("observability: write audit_events: %w", err)
	}
	return nil
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
		fromProvider, fromModelFamily := fallbackHopLabels(meta.FallbackChain[0])
		toProvider, toModelFamily := fallbackHopLabels(meta.FallbackChain[len(meta.FallbackChain)-1])
		w.metrics.fallback.Inc(append(labels, emptyOrUnknown(fromProvider), emptyOrUnknown(fromModelFamily), emptyOrUnknown(toProvider), emptyOrUnknown(toModelFamily))...)
		w.logger.Log(EventTaskFallback, w.buildLogFields(meta))
	}

	switch {
	case err != nil && meta.ErrorCode != sharederrors.CodeAiOutputInvalid:
		w.logger.Log(EventTaskFailed, w.buildLogFields(meta))
	case err == nil && meta.ValidationStatus != aiclient.ValidationStatusInvalid:
		w.logger.Log(EventTaskCompleted, w.buildLogFields(meta))
	}
}

func (w *Wrap) enrichMeta(profileName string, meta aiclient.AICallMeta, callMeta aiclient.CallMetadata) aiclient.AICallMeta {
	if meta.ModelProfileName == "" {
		meta.ModelProfileName = profileName
	}
	if meta.PromptVersion == "" {
		meta.PromptVersion = callMeta.PromptVersion
	}
	if meta.RubricVersion == "" {
		meta.RubricVersion = callMeta.RubricVersion
	}
	if meta.Language == "" {
		meta.Language = callMeta.Language
	}
	if strings.TrimSpace(meta.FeatureKey) == "" {
		meta.FeatureKey = strings.TrimSpace(callMeta.FeatureKey)
	}
	if strings.TrimSpace(meta.FeatureFlag) == "" {
		meta.FeatureFlag = strings.TrimSpace(callMeta.FeatureFlag)
	}
	if strings.TrimSpace(meta.FeatureFlag) == "" {
		meta.FeatureFlag = "none"
	}
	if strings.TrimSpace(meta.DataSourceVersion) == "" {
		meta.DataSourceVersion = strings.TrimSpace(callMeta.DataSourceVersion)
	}
	if strings.TrimSpace(meta.DataSourceVersion) == "" {
		meta.DataSourceVersion = "not_applicable"
	}
	if meta.ErrorCode != "" && meta.ValidationStatus == "" {
		meta.ValidationStatus = aiclient.ValidationStatusInvalid
	}
	if w.resolver == nil || profileName == "" {
		return meta
	}
	profile, err := w.resolver.Resolve(profileName)
	if err != nil || profile == nil {
		return meta
	}
	if meta.Capability == "" {
		meta.Capability = profile.Capability
	}
	if meta.Route == "" {
		meta.Route = profile.Route
	}
	if meta.ModelProfileVersion == "" {
		meta.ModelProfileVersion = profile.Version
	}
	if meta.Provider == "" {
		meta.Provider = profile.Default.ProviderRef
	}
	if meta.ModelID == "" {
		meta.ModelID = profile.Default.Model
	}
	if meta.ModelFamily == "" {
		meta.ModelFamily = modelFamily(meta.ModelID)
	}
	return meta
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
		emptyOrUnknown(string(meta.Capability)),
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
		Capability:          string(meta.Capability),
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
func AITaskRunRowFromMeta(meta aiclient.AICallMeta, taskCtx aiclient.AITaskRunContext, metadata aiclient.AuditMetadata, start, completed time.Time, callErr error) (aiclient.AITaskRunRow, error) {
	if taskCtx.ID == "" {
		taskCtx.ID = idx.NewID()
	}
	if err := taskCtx.Validate(); err != nil {
		return aiclient.AITaskRunRow{}, fmt.Errorf("observability: build ai_task_runs row: %w", err)
	}
	featureKey := strings.TrimSpace(meta.FeatureKey)
	if featureKey == "" {
		return aiclient.AITaskRunRow{}, fmt.Errorf("observability: build ai_task_runs row: feature_key is required")
	}
	featureFlag := strings.TrimSpace(meta.FeatureFlag)
	if featureFlag == "" {
		featureFlag = "none"
	}
	dataSourceVersion := strings.TrimSpace(meta.DataSourceVersion)
	if dataSourceVersion == "" {
		dataSourceVersion = "not_applicable"
	}
	return aiclient.AITaskRunRow{
		ID:                   taskCtx.ID,
		UserID:               taskCtx.UserID,
		Capability:           taskCtx.Capability,
		ResourceType:         taskCtx.ResourceType,
		ResourceID:           taskCtx.ResourceID,
		Provider:             meta.Provider,
		ModelFamily:          meta.ModelFamily,
		ModelID:              meta.ModelID,
		PromptVersion:        meta.PromptVersion,
		RubricVersion:        meta.RubricVersion,
		ModelProfileName:     meta.ModelProfileName,
		ModelProfileVersion:  meta.ModelProfileVersion,
		FeatureKey:           featureKey,
		FeatureFlag:          featureFlag,
		DataSourceVersion:    dataSourceVersion,
		Language:             meta.Language,
		InputTokens:          meta.InputTokens,
		OutputTokens:         meta.OutputTokens,
		CostUSDMicros:        meta.CostUSDMicros,
		LatencyMs:            meta.LatencyMs,
		FallbackChain:        meta.FallbackChain,
		Route:                meta.Route,
		Status:               taskRunStatus(meta, callErr),
		ValidationStatus:     meta.ValidationStatus,
		OutputSchemaVersion:  taskCtx.OutputSchemaVersion,
		ErrorCode:            meta.ErrorCode,
		RawResponseObjectKey: taskCtx.RawResponseObjectKey,
		Metadata:             metadata,
		StartedAt:            start,
		CompletedAt:          completed,
	}, nil
}

func taskRunStatus(meta aiclient.AICallMeta, err error) aiclient.AITaskRunStatus {
	if err != nil {
		if meta.ErrorCode == sharederrors.CodeAiProviderTimeout {
			return aiclient.AITaskRunStatusTimeout
		}
		return aiclient.AITaskRunStatusFailed
	}
	if len(meta.FallbackChain) > 1 {
		return aiclient.AITaskRunStatusFallback
	}
	return aiclient.AITaskRunStatusSuccess
}

func joinRecordError(callErr, recordErr error) error {
	if callErr == nil {
		return recordErr
	}
	if recordErr == nil {
		return callErr
	}
	return errors.Join(callErr, recordErr)
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

func hashBytesHex(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func audioAuditSummary(input aiclient.TranscriptionInput) string {
	return fmt.Sprintf("audio:%d:%s", len(input.Audio), hashBytesHex(input.Audio))
}

func synthesisAuditSummary(input aiclient.SynthesisInput) string {
	return fmt.Sprintf("tts-text:%d:%s", len(input.Text), hashBytesHex([]byte(input.Text)))
}

func synthesisResponseAuditSummary(resp aiclient.SynthesisResponse) string {
	return fmt.Sprintf("tts-audio:%d:%s:%s:%dms", len(resp.Audio), hashBytesHex(resp.Audio), resp.ContentType, resp.DurationMs)
}

func joinMessages(messages []aiclient.Message) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.Role + ":" + m.Content
	}
	return strings.Join(parts, "\n")
}

type outputSchema struct {
	Type       string                  `json:"type"`
	Required   []string                `json:"required"`
	Properties map[string]outputSchema `json:"properties"`
	Items      *outputSchema           `json:"items"`
	Enum       []any                   `json:"enum"`
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
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return errors.New("multiple JSON values in output")
	}
	return validateAgainstSchema(schema, v, "$")
}

func validateAgainstSchema(schema outputSchema, value any, path string) error {
	if schema.Type != "" && !matchesSchemaType(schema.Type, value) {
		return fmt.Errorf("%s expected %s", path, schema.Type)
	}
	if len(schema.Enum) > 0 && !valueInEnum(value, schema.Enum) {
		return fmt.Errorf("%s value is not in enum", path)
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

func valueInEnum(value any, enum []any) bool {
	for _, candidate := range enum {
		if jsonValuesEqual(value, candidate) {
			return true
		}
	}
	return false
}

func jsonValuesEqual(a, b any) bool {
	if an, ok := a.(json.Number); ok {
		return jsonNumberEqual(an, b)
	}
	if bn, ok := b.(json.Number); ok {
		return jsonNumberEqual(bn, a)
	}
	return reflect.DeepEqual(a, b)
}

func jsonNumberEqual(n json.Number, other any) bool {
	switch v := other.(type) {
	case json.Number:
		return n.String() == v.String()
	case float64:
		nf, err := n.Float64()
		return err == nil && nf == v
	default:
		return false
	}
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

func fallbackHopLabels(hop string) (string, string) {
	if hop == "" {
		return "", ""
	}
	if provider, model, ok := strings.Cut(hop, "/"); ok {
		return provider, modelFamily(model)
	}
	return "", modelFamily(hop)
}

func modelFamily(model string) string {
	if model == "" {
		return ""
	}
	parts := strings.Split(model, "-")
	if len(parts) >= 4 && isDateSuffix(parts[len(parts)-3], parts[len(parts)-2], parts[len(parts)-1]) {
		return strings.Join(parts[:len(parts)-3], "-")
	}
	return model
}

func isDateSuffix(year, month, day string) bool {
	return len(year) == 4 && len(month) == 2 && len(day) == 2 &&
		allDigits(year) && allDigits(month) && allDigits(day)
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
