package review

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	platformconfig "github.com/monshunter/easyinterview/backend/internal/platform/config"
	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

type PromptResolver interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type AIClient interface {
	Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error)
}

type ReportRepository interface {
	LoadReportContext(ctx context.Context, reportID string) (ReportContext, error)
	AssertCurrentReportJobLease(ctx context.Context, jobID string, claimedAttempts int32) error
	PersistReportResult(ctx context.Context, in ReportResultPersistence) error
	PersistReportFailure(ctx context.Context, in ReportFailurePersistence) error
}

var (
	ErrReportContextInvalid = errors.New("review: frozen report context invalid")
	ErrReportContextMissing = fmt.Errorf("%w: generation context is missing", ErrReportContextInvalid)
)

type ServiceOptions struct {
	Registry            PromptResolver
	AI                  AIClient
	AITaskRuns          aiclient.AITaskRunWriter
	Repository          ReportRepository
	WaitBeforeRetry     func(context.Context, time.Duration) error
	Now                 func() time.Time
	NewID               func() string
	MaxFramedInputBytes int64
}

type Service struct {
	registry            PromptResolver
	ai                  AIClient
	aiTaskRuns          aiclient.AITaskRunWriter
	repository          ReportRepository
	waitBeforeRetry     func(context.Context, time.Duration) error
	now                 func() time.Time
	newID               func() string
	maxFramedInputBytes int64
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
	if o.WaitBeforeRetry == nil {
		o.WaitBeforeRetry = waitForReportRetry
	}
	if o.MaxFramedInputBytes <= 0 {
		o.MaxFramedInputBytes = platformconfig.DefaultContentLimits().ReportMaxFramedInputBytes
	}
	return &Service{
		registry: o.Registry, ai: o.AI, aiTaskRuns: o.AITaskRuns, repository: o.Repository,
		waitBeforeRetry: o.WaitBeforeRetry, now: o.Now, newID: o.NewID, maxFramedInputBytes: o.MaxFramedInputBytes,
	}
}

func waitForReportRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Service) GenerateReport(ctx context.Context, job AsyncJob) ReportOutcome {
	if s == nil || s.repository == nil || s.registry == nil || s.ai == nil {
		return safeReportFailureOutcome(sharederrors.CodeAiProviderConfigInvalid, false)
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		failure := classifyReportGenerationError(ctxErr)
		return safeReportFailureOutcome(failure.Code, failure.Retryable)
	}
	reportCtx, err := s.repository.LoadReportContext(ctx, job.ResourceID)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			failure := classifyReportGenerationError(ctxErr)
			return safeReportFailureOutcome(failure.Code, failure.Retryable)
		}
		if !errors.Is(err, ErrReportContextInvalid) {
			return safeReportFailureOutcome(sharederrors.CodeAiOutputInvalid, true)
		}
		if reportCtx.Session.ReportID != "" {
			if persistErr := s.persistFailure(ctx, reportCtx, job, sharederrors.CodeAiOutputInvalid, false); persistErr != nil {
				return safeReportFailureOutcome(sharederrors.CodeAiOutputInvalid, true)
			}
		}
		return safeReportFailureOutcome(sharederrors.CodeAiOutputInvalid, false)
	}
	result, actionCallCount, err := s.generateReportWithActionRetries(ctx, reportCtx, job)
	if err != nil {
		failureErr := err
		if ctxErr := ctx.Err(); errors.Is(ctxErr, context.Canceled) {
			failureErr = ctxErr
		}
		failure := classifyReportGenerationError(failureErr)
		if failure.Retryable && actionCallCount >= reportMaxCallsPerAction {
			failure.Retryable = false
		}
		_ = s.writeExplicitFailureTaskRun(ctx, reportCtx, result, aiclient.AITaskRunTaskReportGenerate, reportGenerateFeatureKey, failure)
		failureCtx, cancelFailurePersist := reportPersistenceContext(ctx)
		defer cancelFailurePersist()
		if persistErr := s.persistFailure(failureCtx, reportCtx, job, failure.Code, failure.Retryable); persistErr != nil {
			return safeReportFailureOutcome(sharederrors.CodeAiOutputInvalid, true)
		}
		return safeReportFailureOutcome(failure.Code, failure.Retryable)
	}
	content := result.Content
	persist := ReportResultPersistence{
		UserID:            reportCtx.Session.UserID,
		ReportID:          reportCtx.Session.ReportID,
		SessionID:         reportCtx.Session.SessionID,
		TargetJobID:       reportCtx.Session.TargetJobID,
		AsyncJobID:        job.JobID,
		ClaimedAttempts:   job.Attempts,
		OutboxEventID:     s.newID(),
		AuditEventID:      s.newID(),
		PromptVersion:     result.Meta.PromptVersion,
		RubricVersion:     result.Meta.RubricVersion,
		ModelID:           result.Meta.ModelID,
		Provider:          result.Meta.Provider,
		Language:          result.Meta.Language,
		FeatureFlag:       result.Meta.FeatureFlag,
		DataSourceVersion: result.Meta.DataSourceVersion,
		Now:               s.now(),
		Content:           content,
	}
	persistCtx, cancelPersist := reportPersistenceContext(ctx)
	defer cancelPersist()
	if err := s.repository.PersistReportResult(persistCtx, persist); err != nil {
		return safeReportFailureOutcome(sharederrors.CodeAiOutputInvalid, true)
	}
	return ReportOutcome{Succeeded: true, AsyncJobFinalized: true}
}

func reportPersistenceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx.Err() == nil {
		return ctx, func() {}
	}
	return context.WithTimeout(context.WithoutCancel(ctx), reportPersistenceTimeout)
}

type reportAttemptScope string

const (
	reportPersistenceTimeout                          = 5 * time.Second
	reportAttemptScopeWholeReport  reportAttemptScope = "whole_report"
	reportAttemptScopeActionLabels reportAttemptScope = "action_labels"
)

type reportAttemptRequest struct {
	scope   reportAttemptScope
	payload aiclient.CompletePayload
	base    ReportGenerationResult
	issues  []ReportValidationIssue
}

func (s *Service) generateReportWithActionRetries(ctx context.Context, reportCtx ReportContext, job AsyncJob) (ReportGenerationResult, int, error) {
	if err := ctx.Err(); err != nil {
		return ReportGenerationResult{}, 0, err
	}
	resolution, payload, err := s.prepareReportGeneration(ctx, reportCtx, nil)
	if err != nil {
		return ReportGenerationResult{Resolution: resolution}, 0, err
	}
	request := reportAttemptRequest{
		scope:   reportAttemptScopeWholeReport,
		payload: payload,
		base:    ReportGenerationResult{Resolution: resolution},
	}
	var (
		aggregateMeta aiclient.AICallMeta
		hasMeta       bool
		lastAttempt   int
	)
	for attempt := 1; attempt <= reportMaxCallsPerAction; attempt++ {
		if err := ctx.Err(); err != nil {
			return request.base, lastAttempt, err
		}
		if err := s.repository.AssertCurrentReportJobLease(ctx, job.JobID, job.Attempts); err != nil {
			return request.base, lastAttempt, fmt.Errorf("%w: %v", ErrReportJobLeaseInvalid, err)
		}
		lastAttempt = attempt

		var result ReportGenerationResult
		switch request.scope {
		case reportAttemptScopeActionLabels:
			result, err = s.completeReportActionLabelRepair(ctx, reportCtx, request.base, request.issues, request.payload)
		default:
			result, err = s.completeReportGeneration(ctx, reportCtx, resolution, request.payload)
		}
		if hasMeta {
			result.Meta = AggregateReportRepairMeta(aggregateMeta, result.Meta)
		} else {
			hasMeta = true
		}
		aggregateMeta = result.Meta
		if err == nil {
			return result, lastAttempt, nil
		}
		issues, invalidOutput := reportInvalidIssues(err)
		if lastAttempt >= reportMaxCallsPerAction {
			return result, lastAttempt, err
		}
		if invalidOutput {
			request, err = buildNextReportAttempt(reportCtx, result, issues, s.maxFramedInputBytes)
			if err != nil {
				return result, lastAttempt, err
			}
		} else if !classifyReportGenerationError(err).Retryable {
			return result, lastAttempt, err
		}
		if err := s.waitBeforeRetry(ctx, reportActionRetryDelays[lastAttempt-1]); err != nil {
			return result, lastAttempt, err
		}
	}
	return request.base, lastAttempt, ErrReviewAIOutputInvalid
}

func buildNextReportAttempt(reportCtx ReportContext, previous ReportGenerationResult, issues []ReportValidationIssue, maxFramedInputBytes int64) (reportAttemptRequest, error) {
	if _, targeted := actionLabelRepairIndices(previous.Content, reportCtx.Session.Language, issues); targeted {
		payload, err := BuildReportActionLabelRepairPayload(
			previous.Resolution,
			reportCtx.Session.Language,
			previous.Content,
			issues,
			aiclient.AITaskRunContext{
				UserID: reportCtx.Session.UserID, Capability: aiclient.AITaskRunTaskReportGenerate,
				ResourceType: aiclient.AITaskRunResourceFeedbackReport, ResourceID: reportCtx.Session.ReportID,
			},
		)
		if err != nil {
			return reportAttemptRequest{}, err
		}
		return reportAttemptRequest{scope: reportAttemptScopeActionLabels, payload: payload, base: previous, issues: append([]ReportValidationIssue(nil), issues...)}, nil
	}
	payload, err := reportCompletePayload(previous.Resolution, reportCtx, issues)
	if err != nil {
		return reportAttemptRequest{}, err
	}
	framed, err := frameReportMessages(payload.Messages)
	if err != nil {
		return reportAttemptRequest{}, err
	}
	if int64(len(framed)) > maxFramedInputBytes {
		return reportAttemptRequest{}, fmt.Errorf("%w: repair framed payload is %d bytes", ErrReportContextTooLarge, len(framed))
	}
	return reportAttemptRequest{scope: reportAttemptScopeWholeReport, payload: payload, base: previous}, nil
}

func safeReportFailureOutcome(code string, retryable bool) ReportOutcome {
	return ReportOutcome{ErrorCode: code, ErrorMessage: safeReportErrorMessage(code), Retryable: retryable}
}

func safeReportErrorMessage(code string) string {
	if meta, ok := sharederrors.CodeRegistry[code]; ok {
		return meta.Message
	}
	return "report generation failed"
}

func (s *Service) persistFailure(ctx context.Context, reportCtx ReportContext, job AsyncJob, code string, retryable bool) error {
	return s.repository.PersistReportFailure(ctx, ReportFailurePersistence{
		UserID:          reportCtx.Session.UserID,
		ReportID:        reportCtx.Session.ReportID,
		SessionID:       reportCtx.Session.SessionID,
		AsyncJobID:      job.JobID,
		ClaimedAttempts: job.Attempts,
		OutboxEventID:   s.newID(),
		AuditEventID:    s.newID(),
		ErrorCode:       code,
		Retryable:       retryable,
		MaxAttempts:     job.MaxAttempts,
		Now:             s.now(),
	})
}

type ReportContext struct {
	FrozenContext practicedomain.ReportContextSnapshot
	Session       SessionSnapshot
	Messages      []MessageSnapshot
}

type ReportResultPersistence struct {
	UserID            string
	ReportID          string
	SessionID         string
	TargetJobID       string
	AsyncJobID        string
	ClaimedAttempts   int32
	OutboxEventID     string
	AuditEventID      string
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	Provider          string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
	Now               time.Time
	Content           ReportContentDraft
}

type ReportFailurePersistence struct {
	UserID          string
	ReportID        string
	SessionID       string
	AsyncJobID      string
	ClaimedAttempts int32
	OutboxEventID   string
	AuditEventID    string
	ErrorCode       string
	Retryable       bool
	MaxAttempts     int32
	Now             time.Time
}

func firstNonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func reportInvalidIssues(err error) ([]ReportValidationIssue, bool) {
	if err == nil {
		return nil, false
	}
	var invalid *ReportContentInvalidError
	if errors.As(err, &invalid) {
		return append([]ReportValidationIssue(nil), invalid.Issues...), true
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) && apiErr.Code == sharederrors.CodeAiOutputInvalid {
		return []ReportValidationIssue{OutputSchemaRepairIssue(apiErr)}, true
	}
	return nil, false
}

var outputSchemaRepairPathPattern = regexp.MustCompile(`\$(?:(?:\.[A-Za-z][A-Za-z0-9_]*)|(?:\[[0-9]+\]))*`)

// OutputSchemaRepairIssue reduces a schema-validation error to one safe path
// and one bounded code. It intentionally discards values and free-form prose
// before the coordinate enters a trusted repair message.
func OutputSchemaRepairIssue(err error) ReportValidationIssue {
	issue := ReportValidationIssue{Path: "$", Code: "output_schema_invalid"}
	if err == nil {
		return issue
	}
	message := err.Error()
	if path := outputSchemaRepairPathPattern.FindString(message); path != "" {
		issue.Path = path
	}
	switch {
	case strings.Contains(message, " length ") && strings.Contains(message, " exceeds "):
		issue.Code = "max_length"
	case strings.Contains(message, " length ") && strings.Contains(message, " is below "):
		issue.Code = "min_length"
	case strings.Contains(message, " items, maximum is "):
		issue.Code = "max_items"
	case strings.Contains(message, " items, minimum is "):
		issue.Code = "min_items"
	case strings.Contains(message, " are duplicates"):
		issue.Code = "duplicate"
	case strings.Contains(message, " missing required field "):
		issue.Code = "required"
	case strings.Contains(message, " unknown field "):
		issue.Code = "unknown_field"
	case strings.Contains(message, " value is not in enum"):
		issue.Code = "invalid_enum"
	case strings.Contains(message, " does not match pattern "):
		issue.Code = "invalid_format"
	case strings.Contains(message, " expected "):
		issue.Code = "invalid_type"
	}
	return issue
}
