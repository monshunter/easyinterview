package aiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

// AITaskRunCapability is the business capability persisted into B4 ai_task_runs.
// It is distinct from Model Profile Capability (chat/stt), which only
// describes the provider call shape.
type AITaskRunCapability string

const (
	AITaskRunTaskJDParse          AITaskRunCapability = "jd_parse"
	AITaskRunTaskResumeParse      AITaskRunCapability = AITaskRunCapability(jobs.JobTypeResumeParse)
	AITaskRunTaskQuestionGenerate AITaskRunCapability = "question_generate"
	AITaskRunTaskFollowupGenerate AITaskRunCapability = "followup_generate"
	AITaskRunTaskHintGenerate     AITaskRunCapability = "hint_generate"
	AITaskRunTaskReportGenerate   AITaskRunCapability = AITaskRunCapability(jobs.JobTypeReportGenerate)
	AITaskRunTaskResumeTailor     AITaskRunCapability = AITaskRunCapability(jobs.JobTypeResumeTailor)
	AITaskRunTaskDebriefGenerate  AITaskRunCapability = AITaskRunCapability(jobs.JobTypeDebriefGenerate)
)

var allowedAITaskRunCapabilities = map[AITaskRunCapability]struct{}{
	AITaskRunTaskJDParse:          {},
	AITaskRunTaskResumeParse:      {},
	AITaskRunTaskQuestionGenerate: {},
	AITaskRunTaskFollowupGenerate: {},
	AITaskRunTaskHintGenerate:     {},
	AITaskRunTaskReportGenerate:   {},
	AITaskRunTaskResumeTailor:     {},
	AITaskRunTaskDebriefGenerate:  {},
}

// AITaskRunResourceType mirrors the B2 API-facing ResourceType values. B4
// allows internal extensions, but A3 tests pin the P0 values consumers use.
type AITaskRunResourceType string

const (
	AITaskRunResourceTargetJob       AITaskRunResourceType = "target_job"
	AITaskRunResourceFeedbackReport  AITaskRunResourceType = "feedback_report"
	AITaskRunResourceResumeAsset     AITaskRunResourceType = "resume_asset"
	AITaskRunResourceResumeTailorRun AITaskRunResourceType = "resume_tailor_run"
	AITaskRunResourceDebrief         AITaskRunResourceType = "debrief"
	AITaskRunResourcePrivacyRequest  AITaskRunResourceType = "privacy_request"
)

// AITaskRunStatus is the B4 ai_task_runs.status enum.
type AITaskRunStatus string

const (
	AITaskRunStatusSuccess  AITaskRunStatus = "success"
	AITaskRunStatusFailed   AITaskRunStatus = "failed"
	AITaskRunStatusTimeout  AITaskRunStatus = "timeout"
	AITaskRunStatusFallback AITaskRunStatus = "fallback"
)

// AITaskRunContext is supplied by business callers so the decorator can build
// a row that satisfies B4's required business columns.
type AITaskRunContext struct {
	ID                   string
	UserID               string
	Capability           AITaskRunCapability
	ResourceType         AITaskRunResourceType
	ResourceID           string
	OutputSchemaVersion  string
	RawResponseObjectKey string
}

// Validate checks the subset of B4 constraints A3 can enforce before the row
// reaches the real store.
func (c AITaskRunContext) Validate() error {
	if c.ID != "" {
		if _, err := uuid.Parse(c.ID); err != nil {
			return fmt.Errorf("id must be uuid: %w", err)
		}
	}
	if c.UserID != "" {
		if _, err := uuid.Parse(c.UserID); err != nil {
			return fmt.Errorf("user_id must be uuid: %w", err)
		}
	}
	if _, ok := allowedAITaskRunCapabilities[c.Capability]; !ok {
		return fmt.Errorf("capability %q is not allowed by B4 ai_task_runs", c.Capability)
	}
	if c.ResourceType == "" {
		return fmt.Errorf("resource_type is required")
	}
	if c.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}
	if _, err := uuid.Parse(c.ResourceID); err != nil {
		return fmt.Errorf("resource_id must be uuid: %w", err)
	}
	return nil
}

// AITaskRunRow is the typed payload A3 writes for every AIClient call. B4
// db-migrations-baseline owns the table schema; this struct mirrors the
// columns A3 fills (spec §2.1 / D-6).
//
// FeatureKey, FeatureFlag, and DataSourceVersion are F3 prompt-rubric
// provenance fields added by plan prompt-rubric-registry/001-baseline
// phase 4.3. The writer rejects empty FeatureKey (the F3 coordinate is
// required) and substitutes documented defaults for the other two
// (FeatureFlag = "none", DataSourceVersion = "not_applicable") when the
// caller passes empty strings.
type AITaskRunRow struct {
	ID                   string
	UserID               string
	Capability           AITaskRunCapability
	ResourceType         AITaskRunResourceType
	ResourceID           string
	Provider             string
	ModelFamily          string
	ModelID              string
	PromptVersion        string
	RubricVersion        string
	ModelProfileName     string
	ModelProfileVersion  string
	FeatureKey           string
	FeatureFlag          string
	DataSourceVersion    string
	Language             string
	InputTokens          int
	OutputTokens         int
	CostUSDMicros        int64
	LatencyMs            int64
	FallbackChain        []string
	Route                string
	Status               AITaskRunStatus
	ValidationStatus     ValidationStatus
	OutputSchemaVersion  string
	ErrorCode            string
	RawResponseObjectKey string
	Metadata             AuditMetadata
	StartedAt            time.Time
	CompletedAt          time.Time
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
