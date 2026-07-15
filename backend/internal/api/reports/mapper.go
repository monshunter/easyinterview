package reports

import (
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	reviewdomain "github.com/monshunter/easyinterview/backend/internal/review"
)

func toAPITargetJobReportsOverview(overview reviewdomain.TargetJobReportsOverviewRecord) api.TargetJobReportsOverview {
	out := api.TargetJobReportsOverview{
		TargetJobId: overview.TargetJobID,
		Rounds:      make([]api.TargetJobReportRoundOverview, 0, len(overview.Rounds)),
	}
	for _, round := range overview.Rounds {
		item := api.TargetJobReportRoundOverview{
			Round: api.PracticeRoundRef{
				RoundId:       round.Round.RoundID,
				RoundSequence: round.Round.RoundSequence,
			},
		}
		if round.CurrentReport != nil {
			item.CurrentReport = &api.TargetJobCurrentReportSummary{
				Id:          round.CurrentReport.ID,
				GeneratedAt: round.CurrentReport.GeneratedAt.UTC().Format(timeFormatRFC3339),
			}
		}
		if round.LatestAttempt != nil {
			item.LatestAttempt = &api.TargetJobReportAttemptSummary{
				Id:        round.LatestAttempt.ID,
				Status:    round.LatestAttempt.Status,
				CreatedAt: round.LatestAttempt.CreatedAt.UTC().Format(timeFormatRFC3339),
			}
			if round.LatestAttempt.ErrorCode != nil {
				code := api.ApiErrorCode(*round.LatestAttempt.ErrorCode)
				item.LatestAttempt.ErrorCode = &code
			}
		}
		out.Rounds = append(out.Rounds, item)
	}
	return out
}

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

func toAPIReportConversation(conversation reviewdomain.ReportConversationRecord) api.ReportConversation {
	out := api.ReportConversation{
		ReportId:     conversation.ReportID,
		ReportStatus: conversation.Status,
		Context:      toAPIReportContext(conversation.Context),
		Messages:     make([]api.ReportConversationMessage, 0, len(conversation.Messages)),
	}
	for _, message := range conversation.Messages {
		out.Messages = append(out.Messages, api.ReportConversationMessage{
			Sequence:  message.Sequence,
			Role:      message.Role,
			Content:   message.Content,
			CreatedAt: message.CreatedAt.UTC().Format(time.RFC3339Nano),
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
