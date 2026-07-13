package review

import (
	"context"
	"errors"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type reportFailure struct {
	Code             string
	Retryable        bool
	ExplicitTaskRun  bool
	ValidationStatus aiclient.ValidationStatus
}

func classifyReportGenerationError(err error) reportFailure {
	switch {
	case errors.Is(err, context.Canceled):
		return reportFailure{Code: sharederrors.CodeAiProviderTimeout, Retryable: false}
	case errors.Is(err, context.DeadlineExceeded):
		return reportFailure{Code: sharederrors.CodeAiProviderTimeout, Retryable: true}
	case errors.Is(err, ErrReportContextTooLarge):
		return reportFailure{Code: sharederrors.CodeReportContextTooLarge, Retryable: false, ExplicitTaskRun: true, ValidationStatus: aiclient.ValidationStatusInvalid}
	case errors.Is(err, ErrReportJobLeaseInvalid):
		return reportFailure{Code: sharederrors.CodeAiOutputInvalid, Retryable: true}
	case errors.Is(err, ErrReportGenerationConfigInvalid):
		return reportFailure{Code: sharederrors.CodeAiProviderConfigInvalid, Retryable: false, ExplicitTaskRun: true}
	case errors.Is(err, registry.ErrPromptUnsupported), errors.Is(err, registry.ErrLanguageUnsupported):
		return reportFailure{Code: sharederrors.CodeAiProviderConfigInvalid, Retryable: false, ExplicitTaskRun: true}
	case errors.Is(err, ErrReviewAIOutputInvalid):
		return reportFailure{Code: sharederrors.CodeAiOutputInvalid, Retryable: false, ExplicitTaskRun: true, ValidationStatus: aiclient.ValidationStatusInvalid}
	}
	var apiErr *sharederrors.APIError
	if errors.As(err, &apiErr) && apiErr.Code != "" {
		meta := sharederrors.CodeRegistry[apiErr.Code]
		return reportFailure{Code: apiErr.Code, Retryable: apiErr.Retryable && meta.Retryable}
	}
	return reportFailure{Code: sharederrors.CodeAiOutputInvalid, Retryable: false, ExplicitTaskRun: true, ValidationStatus: aiclient.ValidationStatusInvalid}
}

func (s *Service) writeExplicitFailureTaskRun(ctx context.Context, reportCtx ReportContext, result ReportGenerationResult, capability aiclient.AITaskRunCapability, featureKey string, failure reportFailure) error {
	if !failure.ExplicitTaskRun || s == nil || s.aiTaskRuns == nil {
		return nil
	}
	now := s.now()
	return s.aiTaskRuns.WriteAITaskRun(ctx, aiclient.AITaskRunRow{
		ID:                s.newID(),
		UserID:            reportCtx.Session.UserID,
		Capability:        capability,
		ResourceType:      aiclient.AITaskRunResourceFeedbackReport,
		ResourceID:        reportCtx.Session.ReportID,
		ModelID:           firstNonEmpty(result.Meta.ModelID, "not_applicable"),
		PromptVersion:     firstNonEmpty(result.Resolution.PromptVersion, "not_applicable"),
		RubricVersion:     firstNonEmpty(result.Resolution.RubricVersion, "not_applicable"),
		ModelProfileName:  firstNonEmpty(result.Resolution.ModelProfileName, modelProfileForFeature(featureKey)),
		FeatureKey:        featureKey,
		FeatureFlag:       firstNonEmpty(result.Resolution.FeatureFlag, "none"),
		DataSourceVersion: firstNonEmpty(result.Resolution.DataSourceVersion, "not_applicable"),
		Language:          fallbackLanguage(reportCtx.Session.Language),
		Status:            aiclient.AITaskRunStatusFailed,
		ValidationStatus:  failure.ValidationStatus,
		ErrorCode:         failure.Code,
		StartedAt:         now,
		CompletedAt:       now,
	})
}

func modelProfileForFeature(featureKey string) string {
	switch featureKey {
	case reportGenerateFeatureKey:
		return "report.generate.default"
	default:
		return ""
	}
}
