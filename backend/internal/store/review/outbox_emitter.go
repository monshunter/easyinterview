package review

import (
	"encoding/json"
	"fmt"
	"strings"

	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type ReportGeneratedInput struct {
	ReportID          string
	SessionID         string
	TargetJobID       string
	PreparednessLevel sharedtypes.ReadinessTier
	PromptVersion     string
	RubricVersion     string
	ModelID           string
}

type ReportGenerationFailedInput struct {
	ReportID  string
	SessionID string
	ErrorCode string
	Retryable bool
}

func BuildReportGeneratedPayload(in ReportGeneratedInput) (sharedevents.ReportGeneratedPayload, error) {
	if strings.TrimSpace(in.ReportID) == "" {
		return sharedevents.ReportGeneratedPayload{}, fmt.Errorf("%s: reportId is required", sharedevents.EventNameReportGenerated)
	}
	if strings.TrimSpace(in.SessionID) == "" {
		return sharedevents.ReportGeneratedPayload{}, fmt.Errorf("%s: sessionId is required", sharedevents.EventNameReportGenerated)
	}
	if strings.TrimSpace(in.TargetJobID) == "" {
		return sharedevents.ReportGeneratedPayload{}, fmt.Errorf("%s: targetJobId is required", sharedevents.EventNameReportGenerated)
	}
	if strings.TrimSpace(string(in.PreparednessLevel)) == "" {
		return sharedevents.ReportGeneratedPayload{}, fmt.Errorf("%s: preparednessLevel is required", sharedevents.EventNameReportGenerated)
	}
	if strings.TrimSpace(in.PromptVersion) == "" || strings.TrimSpace(in.RubricVersion) == "" || strings.TrimSpace(in.ModelID) == "" {
		return sharedevents.ReportGeneratedPayload{}, fmt.Errorf("%s: provenance fields are required", sharedevents.EventNameReportGenerated)
	}
	payload := sharedevents.ReportGeneratedPayload{
		ModelID: strings.TrimSpace(in.ModelID), PreparednessLevel: in.PreparednessLevel,
		PromptVersion: strings.TrimSpace(in.PromptVersion), ReportID: strings.TrimSpace(in.ReportID),
		RubricVersion: strings.TrimSpace(in.RubricVersion), SessionID: strings.TrimSpace(in.SessionID),
		TargetJobID: strings.TrimSpace(in.TargetJobID),
	}
	if err := assertNoReviewOutboxPII(payload); err != nil {
		return sharedevents.ReportGeneratedPayload{}, err
	}
	return payload, nil
}

func BuildReportGenerationFailedPayload(in ReportGenerationFailedInput) (sharedevents.ReportGenerationFailedPayload, error) {
	if strings.TrimSpace(in.ReportID) == "" {
		return sharedevents.ReportGenerationFailedPayload{}, fmt.Errorf("%s: reportId is required", sharedevents.EventNameReportGenerationFailed)
	}
	if strings.TrimSpace(in.SessionID) == "" {
		return sharedevents.ReportGenerationFailedPayload{}, fmt.Errorf("%s: sessionId is required", sharedevents.EventNameReportGenerationFailed)
	}
	if strings.TrimSpace(in.ErrorCode) == "" {
		return sharedevents.ReportGenerationFailedPayload{}, fmt.Errorf("%s: errorCode is required", sharedevents.EventNameReportGenerationFailed)
	}
	payload := sharedevents.ReportGenerationFailedPayload{
		ErrorCode: strings.TrimSpace(in.ErrorCode),
		ReportID:  strings.TrimSpace(in.ReportID),
		Retryable: in.Retryable,
		SessionID: strings.TrimSpace(in.SessionID),
	}
	if err := assertNoReviewOutboxPII(payload); err != nil {
		return sharedevents.ReportGenerationFailedPayload{}, err
	}
	return payload, nil
}

func assertNoReviewOutboxPII(payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{"question_text", "questiontext", "answer_text", "answertext", "hint_text", "hinttext", "prompt_body", "promptbody", "response_body", "responsebody", "provider_secret", "providersecret"} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("review outbox payload contains forbidden field %q", forbidden)
		}
	}
	return nil
}
