package practice

import (
	"encoding/json"
	"fmt"
	"strings"

	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type PracticeSessionStartedInput struct {
	Goal         sharedtypes.PracticeGoal
	Language     string
	Mode         sharedtypes.PracticeMode
	PlanID       string
	SessionID    string
	TargetJobID  string
	QuestionText string
}

func BuildPracticeSessionStartedPayload(in PracticeSessionStartedInput) (sharedevents.PracticeSessionStartedPayload, error) {
	if strings.TrimSpace(in.QuestionText) != "" {
		return sharedevents.PracticeSessionStartedPayload{}, fmt.Errorf("%s: questionText is forbidden in outbox payload", sharedevents.EventNamePracticeSessionStarted)
	}
	if strings.TrimSpace(in.PlanID) == "" {
		return sharedevents.PracticeSessionStartedPayload{}, fmt.Errorf("%s: planId is required", sharedevents.EventNamePracticeSessionStarted)
	}
	if strings.TrimSpace(in.SessionID) == "" {
		return sharedevents.PracticeSessionStartedPayload{}, fmt.Errorf("%s: sessionId is required", sharedevents.EventNamePracticeSessionStarted)
	}
	if strings.TrimSpace(in.TargetJobID) == "" {
		return sharedevents.PracticeSessionStartedPayload{}, fmt.Errorf("%s: targetJobId is required", sharedevents.EventNamePracticeSessionStarted)
	}
	if strings.TrimSpace(in.Language) == "" {
		return sharedevents.PracticeSessionStartedPayload{}, fmt.Errorf("%s: language is required", sharedevents.EventNamePracticeSessionStarted)
	}
	payload := sharedevents.PracticeSessionStartedPayload{
		Goal:        in.Goal,
		Language:    strings.TrimSpace(in.Language),
		Mode:        in.Mode,
		PlanID:      strings.TrimSpace(in.PlanID),
		SessionID:   strings.TrimSpace(in.SessionID),
		TargetJobID: strings.TrimSpace(in.TargetJobID),
	}
	if err := assertNoPracticeOutboxPII(payload); err != nil {
		return sharedevents.PracticeSessionStartedPayload{}, err
	}
	return payload, nil
}

func assertNoPracticeOutboxPII(payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	lower := strings.ToLower(string(raw))
	for _, forbidden := range []string{"question_text", "questiontext", "answer_text", "answertext", "hint_text", "hinttext", "prompt_body", "promptbody", "response_body", "responsebody", "provider_secret", "providersecret"} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("practice outbox payload contains forbidden field %q", forbidden)
		}
	}
	return nil
}
