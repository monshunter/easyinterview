package review

import (
	"context"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type PromptResolver interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type AIClient interface {
	Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error)
}

type ReportRepository interface {
	LoadReportContext(ctx context.Context, reportID string) (ReportContext, error)
	PersistReportResult(ctx context.Context, in ReportResultPersistence) error
	PersistReportFailure(ctx context.Context, in ReportFailurePersistence) error
}

type ServiceOptions struct {
	Registry   PromptResolver
	AI         AIClient
	AITaskRuns aiclient.AITaskRunWriter
	Repository ReportRepository
	Now        func() time.Time
	NewID      func() string
}

type Service struct {
	registry   PromptResolver
	ai         AIClient
	aiTaskRuns aiclient.AITaskRunWriter
	repository ReportRepository
	now        func() time.Time
	newID      func() string
}

func NewService(opts ...ServiceOptions) *Service {
	var o ServiceOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Now == nil {
		o.Now = func() time.Time { return time.Now().UTC() }
	}
	if o.NewID == nil {
		o.NewID = idx.NewID
	}
	return &Service{registry: o.Registry, ai: o.AI, aiTaskRuns: o.AITaskRuns, repository: o.Repository, now: o.Now, newID: o.NewID}
}

func (s *Service) GenerateReport(ctx context.Context, job AsyncJob) ReportOutcome {
	if s == nil || s.repository == nil || s.registry == nil || s.ai == nil {
		return ReportOutcome{Succeeded: true}
	}
	reportCtx, err := s.repository.LoadReportContext(ctx, job.ResourceID)
	if err != nil {
		return ReportOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: true}
	}
	content, err := s.generateReportContent(ctx, reportCtx.Session, reportCtx.Plan, reportCtx.Messages, reportCtx.Rubric)
	if err != nil {
		failure := classifyReportGenerationError(err)
		_ = s.writeExplicitFailureTaskRun(ctx, reportCtx, aiclient.AITaskRunTaskReportGenerate, reportGenerateFeatureKey, failure)
		_ = s.persistFailure(ctx, reportCtx, job, failure.Code, failure.Retryable)
		return ReportOutcome{ErrorCode: failure.Code, ErrorMessage: err.Error(), Retryable: failure.Retryable}
	}
	assessments := dimensionAssessments(content)
	readiness := readinessFromContent(content)
	retryFocus := append([]string(nil), content.RetryFocusCompetencyCodes...)
	nextAction := decideNextAction(readiness, len(retryFocus))
	if len(content.NextActions) == 0 {
		content.NextActions = []ReportNextActionDraft{{Type: string(nextAction), Label: string(nextAction)}}
	}
	persist := ReportResultPersistence{
		UserID:                    reportCtx.Session.UserID,
		ReportID:                  reportCtx.Session.ReportID,
		SessionID:                 reportCtx.Session.SessionID,
		TargetJobID:               reportCtx.Session.TargetJobID,
		AsyncJobID:                job.JobID,
		OutboxEventID:             s.newID(),
		AuditEventID:              s.newID(),
		PreparednessLevel:         readiness,
		PromptVersion:             firstNonEmpty(reportCtx.ReportPromptVersion, "v0.1.0"),
		RubricVersion:             firstNonEmpty(reportCtx.ReportRubricVersion, "v0.1.0"),
		ModelID:                   firstNonEmpty(reportCtx.ModelID, "model-profile:report.generate.default"),
		Provider:                  reportCtx.Provider,
		Language:                  fallbackLanguage(reportCtx.Session.Language),
		FeatureFlag:               firstNonEmpty(reportCtx.FeatureFlag, "none"),
		DataSourceVersion:         firstNonEmpty(reportCtx.DataSourceVersion, "registry.v1"),
		RetryFocusCompetencyCodes: retryFocus,
		Now:                       s.now(),
		Content:                   content,
		DimensionAssessments:      assessments,
	}
	if err := s.repository.PersistReportResult(ctx, persist); err != nil {
		return ReportOutcome{ErrorCode: sharederrors.CodeAiOutputInvalid, ErrorMessage: err.Error(), Retryable: true}
	}
	return ReportOutcome{Succeeded: true, AsyncJobFinalized: true}
}

func (s *Service) persistFailure(ctx context.Context, reportCtx ReportContext, job AsyncJob, code string, retryable bool) error {
	return s.repository.PersistReportFailure(ctx, ReportFailurePersistence{
		UserID:        reportCtx.Session.UserID,
		ReportID:      reportCtx.Session.ReportID,
		SessionID:     reportCtx.Session.SessionID,
		AsyncJobID:    job.JobID,
		OutboxEventID: s.newID(),
		AuditEventID:  s.newID(),
		ErrorCode:     code,
		Retryable:     retryable,
		Attempts:      job.Attempts,
		MaxAttempts:   job.MaxAttempts,
		Now:           s.now(),
	})
}

type ReportContext struct {
	Session             SessionSnapshot
	Plan                PracticePlanSnapshot
	Messages            []MessageSnapshot
	Rubric              registry.RubricSchema
	ReportPromptVersion string
	ReportRubricVersion string
	ModelID             string
	Provider            string
	FeatureFlag         string
	DataSourceVersion   string
}

type ReportResultPersistence struct {
	UserID                    string
	ReportID                  string
	SessionID                 string
	TargetJobID               string
	AsyncJobID                string
	OutboxEventID             string
	AuditEventID              string
	PreparednessLevel         sharedtypes.ReadinessTier
	PromptVersion             string
	RubricVersion             string
	ModelID                   string
	Provider                  string
	Language                  string
	FeatureFlag               string
	DataSourceVersion         string
	RetryFocusCompetencyCodes []string
	Now                       time.Time
	Content                   ReportContentDraft
	DimensionAssessments      []DimensionAssessmentDraft
}

type ReportFailurePersistence struct {
	UserID        string
	ReportID      string
	SessionID     string
	AsyncJobID    string
	OutboxEventID string
	AuditEventID  string
	ErrorCode     string
	Retryable     bool
	Attempts      int32
	MaxAttempts   int32
	Now           time.Time
}

func firstNonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
