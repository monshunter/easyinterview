package reports

import (
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func toAPIFeedbackReport(report reviewdomain.FeedbackReportRecord) api.FeedbackReport {
	out := api.FeedbackReport{
		Id:                       report.ID,
		SessionId:                report.SessionID,
		TargetJobId:              report.TargetJobID,
		Status:                   report.Status,
		Summary:                  report.Summary,
		Context:                  toAPIReportContext(report.Context),
		PreparednessLevel:        report.PreparednessLevel,
		Highlights:               make([]api.ReportHighlight, 0, len(report.Highlights)),
		Issues:                   make([]api.ReportIssue, 0, len(report.Issues)),
		NextActions:              make([]api.ReportNextAction, 0, len(report.NextActions)),
		DimensionAssessments:     make([]api.DimensionAssessment, 0, len(report.DimensionAssessments)),
		RetryFocusDimensionCodes: append([]string{}, report.RetryFocusDimensionCodes...),
		CreatedAt:                report.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:                report.UpdatedAt.UTC().Format(timeFormatRFC3339),
	}
	if report.ErrorCode != nil && *report.ErrorCode != "" {
		code := api.ApiErrorCode(*report.ErrorCode)
		out.ErrorCode = &code
	}
	if report.Provenance != nil {
		out.Provenance = &api.GenerationProvenance{
			PromptVersion:     report.Provenance.PromptVersion,
			RubricVersion:     report.Provenance.RubricVersion,
			ModelId:           report.Provenance.ModelID,
			Language:          report.Provenance.Language,
			FeatureFlag:       report.Provenance.FeatureFlag,
			DataSourceVersion: report.Provenance.DataSourceVersion,
		}
	}
	for _, highlight := range report.Highlights {
		out.Highlights = append(out.Highlights, api.ReportHighlight{
			DimensionCode: highlight.DimensionCode,
			Evidence:      highlight.Evidence,
			Confidence:    highlight.Confidence,
		})
	}
	for _, issue := range report.Issues {
		out.Issues = append(out.Issues, api.ReportIssue{
			DimensionCode: issue.DimensionCode,
			Evidence:      issue.Evidence,
			Confidence:    issue.Confidence,
		})
	}
	for _, action := range report.NextActions {
		out.NextActions = append(out.NextActions, api.ReportNextAction{Type: action.Type, Label: action.Label})
	}
	for _, assessment := range report.DimensionAssessments {
		out.DimensionAssessments = append(out.DimensionAssessments, api.DimensionAssessment{
			Code: assessment.Code, Label: assessment.Label, Status: assessment.Status, Confidence: assessment.Confidence,
		})
	}
	return out
}

func toAPIReportContext(context reviewdomain.ReportContextProjection) api.ReportContextSnapshot {
	return api.ReportContextSnapshot{
		SourcePlanId:      context.SourcePlanID,
		TargetJobTitle:    context.TargetJobTitle,
		TargetJobCompany:  context.TargetJobCompany,
		ResumeId:          context.ResumeID,
		ResumeDisplayName: context.ResumeDisplayName,
		RoundId:           context.RoundID,
		RoundSequence:     context.RoundSequence,
		RoundName:         context.RoundName,
		RoundType:         context.RoundType,
		Language:          context.Language,
		HasNextRound:      context.HasNextRound,
	}
}
