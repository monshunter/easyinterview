package reports

import (
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func toAPIFeedbackReport(report reviewdomain.FeedbackReportRecord) api.FeedbackReport {
	out := api.FeedbackReport{
		Id:                  report.ID,
		SessionId:           report.SessionID,
		TargetJobId:         report.TargetJobID,
		Status:              report.Status,
		PreparednessLevel:   report.PreparednessLevel,
		Highlights:          make([]api.ReportHighlight, 0, len(report.Highlights)),
		Issues:              make([]api.ReportIssue, 0, len(report.Issues)),
		NextActions:         make([]api.ReportNextAction, 0, len(report.NextActions)),
		QuestionAssessments: make([]api.QuestionAssessment, 0, len(report.QuestionAssessments)),
		RetryFocusTurnIds:   append([]string(nil), report.RetryFocusTurnIDs...),
		CreatedAt:           report.CreatedAt.UTC().Format(timeFormatRFC3339),
		UpdatedAt:           report.UpdatedAt.UTC().Format(timeFormatRFC3339),
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
			Dimension:  highlight.Dimension,
			Evidence:   highlight.Evidence,
			Confidence: highlight.Confidence,
		})
	}
	for _, issue := range report.Issues {
		out.Issues = append(out.Issues, api.ReportIssue{
			Dimension:  issue.Dimension,
			Evidence:   issue.Evidence,
			Confidence: issue.Confidence,
		})
	}
	for _, action := range report.NextActions {
		out.NextActions = append(out.NextActions, api.ReportNextAction{Type: action.Type, Label: action.Label})
	}
	for _, assessment := range report.QuestionAssessments {
		out.QuestionAssessments = append(out.QuestionAssessments, toAPIQuestionAssessment(assessment))
	}
	return out
}

func toAPIQuestionAssessment(assessment reviewdomain.QuestionAssessmentRecord) api.QuestionAssessment {
	dimensions := make(map[string]any, len(assessment.DimensionResults))
	for name, result := range assessment.DimensionResults {
		dimensions[name] = api.DimensionResult{Status: result.Status, Confidence: result.Confidence}
	}
	return api.QuestionAssessment{
		TurnId:              assessment.TurnID,
		QuestionIntent:      assessment.QuestionIntent,
		DimensionResults:    dimensions,
		ReviewStatus:        assessment.ReviewStatus,
		IncludedInRetryPlan: assessment.IncludedInRetryPlan,
	}
}
