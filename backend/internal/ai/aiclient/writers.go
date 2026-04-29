package aiclient

import "context"

// AITaskRunRow is the typed payload A3 writes for every AIClient call. B4
// db-migrations-baseline owns the table schema; this struct mirrors the
// columns A3 fills (spec §2.1 / D-6).
type AITaskRunRow struct {
	Provider            string
	ModelFamily         string
	ModelID             string
	TaskType            TaskType
	PromptVersion       string
	RubricVersion       string
	ModelProfileName    string
	ModelProfileVersion string
	Language            string
	InputTokens         int
	OutputTokens        int
	CostUSDMicros       int64
	LatencyMs           int64
	FallbackChain       []string
	Route               string
	ValidationStatus    ValidationStatus
	ErrorCode           string
}

// AITaskRunWriter persists one ai_task_runs row per AIClient call. Tests
// supply an in-memory implementation; production wiring (out of plan 001
// scope) binds the real PG store.
type AITaskRunWriter interface {
	WriteAITaskRun(ctx context.Context, row AITaskRunRow) error
}

// AuditEventRow mirrors the typed audit_events columns A3 fills. Action is
// always "ai.call" (spec §4.3); Metadata is restricted to hash + length +
// profile triples — the decorator enforces the privacy red line.
type AuditEventRow struct {
	Action   string
	Metadata AuditMetadata
}

// AuditMetadata is the closed allowlist of keys A3 may write into
// audit_events.metadata. Adding a key requires a spec revision.
type AuditMetadata struct {
	PromptHash         string
	ResponseHash       string
	PromptCharLength   int
	ResponseCharLength int
	ProfileName        string
}

// AuditEventWriter persists one audit_events row per AIClient call.
type AuditEventWriter interface {
	WriteAuditEvent(ctx context.Context, row AuditEventRow) error
}
