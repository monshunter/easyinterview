package practice

import (
	"encoding/json"
	"strings"
	"testing"

	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestBuildPracticeSessionStartedPayloadMatchesB3Schema(t *testing.T) {
	payload, err := BuildPracticeSessionStartedPayload(PracticeSessionStartedInput{
		Goal:        sharedtypes.PracticeGoalBaseline,
		Language:    "zh-CN",
		Mode:        sharedtypes.PracticeModeAssisted,
		PlanID:      "plan-1",
		SessionID:   "session-1",
		TargetJobID: "target-1",
	})
	if err != nil {
		t.Fatalf("BuildPracticeSessionStartedPayload returned error: %v", err)
	}
	want := sharedevents.PracticeSessionStartedPayload{
		Goal:        sharedtypes.PracticeGoalBaseline,
		Language:    "zh-CN",
		Mode:        sharedtypes.PracticeModeAssisted,
		PlanID:      "plan-1",
		SessionID:   "session-1",
		TargetJobID: "target-1",
	}
	wantRaw, _ := json.Marshal(want)
	gotRaw, _ := json.Marshal(payload)
	if string(gotRaw) != string(wantRaw) {
		t.Fatalf("payload mismatch\nwant: %s\n got: %s", wantRaw, gotRaw)
	}
}

func TestBuildPracticeSessionStartedPayloadRejectsPIIBoundary(t *testing.T) {
	_, err := BuildPracticeSessionStartedPayload(PracticeSessionStartedInput{
		Goal:         sharedtypes.PracticeGoalBaseline,
		Language:     "zh-CN",
		Mode:         sharedtypes.PracticeModeAssisted,
		PlanID:       "plan-1",
		SessionID:    "session-1",
		TargetJobID:  "target-1",
		QuestionText: "raw first question must never enter outbox",
	})
	if err == nil {
		t.Fatalf("expected PII boundary error")
	}
	if !strings.Contains(err.Error(), "questionText") {
		t.Fatalf("error should identify forbidden field, got %v", err)
	}
}

func TestBuildPracticeTurnCompletedPayloadMatchesB3Schema(t *testing.T) {
	payload, err := BuildPracticeTurnCompletedPayload(PracticeTurnCompletedInput{
		SessionID:        "session-1",
		TurnID:           "turn-1",
		TurnIndex:        2,
		QuestionIntent:   "behavioral.leadership",
		FollowUpCount:    1,
		AnswerCharLength: 48,
	})
	if err != nil {
		t.Fatalf("BuildPracticeTurnCompletedPayload returned error: %v", err)
	}
	want := sharedevents.PracticeTurnCompletedPayload{
		SessionID:        "session-1",
		TurnID:           "turn-1",
		TurnIndex:        2,
		QuestionIntent:   "behavioral.leadership",
		FollowUpCount:    1,
		AnswerCharLength: 48,
	}
	wantRaw, _ := json.Marshal(want)
	gotRaw, _ := json.Marshal(payload)
	if string(gotRaw) != string(wantRaw) {
		t.Fatalf("payload mismatch\nwant: %s\n got: %s", wantRaw, gotRaw)
	}
}

func TestBuildPracticeTurnCompletedPayloadRejectsPIIBoundary(t *testing.T) {
	_, err := BuildPracticeTurnCompletedPayload(PracticeTurnCompletedInput{
		SessionID:        "session-1",
		TurnID:           "turn-1",
		TurnIndex:        2,
		QuestionIntent:   "answer_text must never enter outbox",
		FollowUpCount:    1,
		AnswerCharLength: 48,
	})
	if err == nil {
		t.Fatalf("expected PII boundary error")
	}
	if !strings.Contains(err.Error(), "answer_text") {
		t.Fatalf("error should identify forbidden field, got %v", err)
	}
}

func TestBuildPracticeSessionCompletedPayloadMatchesB3Schema(t *testing.T) {
	payload, err := BuildPracticeSessionCompletedPayload(PracticeSessionCompletedInput{
		Language:    "zh-CN",
		PlanID:      "plan-1",
		SessionID:   "session-1",
		TargetJobID: "target-1",
		TurnCount:   3,
	})
	if err != nil {
		t.Fatalf("BuildPracticeSessionCompletedPayload returned error: %v", err)
	}
	want := sharedevents.PracticeSessionCompletedPayload{
		Language:    "zh-CN",
		PlanID:      "plan-1",
		SessionID:   "session-1",
		TargetJobID: "target-1",
		TurnCount:   3,
	}
	wantRaw, _ := json.Marshal(want)
	gotRaw, _ := json.Marshal(payload)
	if string(gotRaw) != string(wantRaw) {
		t.Fatalf("payload mismatch\nwant: %s\n got: %s", wantRaw, gotRaw)
	}
}

func TestBuildPracticeSessionCompletedPayloadRejectsPIIBoundary(t *testing.T) {
	_, err := BuildPracticeSessionCompletedPayload(PracticeSessionCompletedInput{
		Language:     "zh-CN",
		PlanID:       "plan-1",
		SessionID:    "session-1",
		TargetJobID:  "target-1",
		TurnCount:    3,
		QuestionText: "raw question must never enter outbox",
	})
	if err == nil {
		t.Fatalf("expected PII boundary error")
	}
	if !strings.Contains(err.Error(), "questionText") {
		t.Fatalf("error should identify forbidden field, got %v", err)
	}
}
