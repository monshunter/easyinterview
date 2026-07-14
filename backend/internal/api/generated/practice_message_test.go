package generated

import (
	"encoding/json"
	"testing"
)

func TestPracticeMessageRoleUnionRejectsInvalidRecoveryShapes(t *testing.T) {
	validUser := `{
		"id":"01918fa0-0000-7000-8000-000000006001",
		"seqNo":2,
		"role":"user",
		"content":"same message",
		"createdAt":"2026-07-13T08:03:00Z",
		"clientMessageId":"01918fa0-0000-7000-8000-000000007001",
		"replyStatus":"retryable_failed"
	}`
	var userUnion PracticeMessage
	if err := json.Unmarshal([]byte(validUser), &userUnion); err != nil {
		t.Fatalf("valid user message: %v", err)
	}
	user, ok := userUnion.AsPracticeUserMessage()
	if !ok || user.ClientMessageId != "01918fa0-0000-7000-8000-000000007001" || user.ReplyStatus != PracticeReplyStatusRetryableFailed {
		t.Fatalf("unexpected user projection: %#v, ok=%v", user, ok)
	}
	encoded, err := json.Marshal(NewPracticeMessageFromPracticeUserMessage(user))
	if err != nil {
		t.Fatalf("marshal valid user message: %v", err)
	}
	if string(encoded) == "{}" {
		t.Fatal("union constructor must marshal the selected wire variant")
	}

	validAssistant := `{
		"id":"01918fa0-0000-7000-8000-000000006002",
		"seqNo":3,
		"role":"assistant",
		"content":"follow up",
		"createdAt":"2026-07-13T08:03:01Z"
	}`
	var assistantUnion PracticeMessage
	if err := json.Unmarshal([]byte(validAssistant), &assistantUnion); err != nil {
		t.Fatalf("valid assistant message: %v", err)
	}
	if _, ok := assistantUnion.AsPracticeAssistantMessage(); !ok {
		t.Fatal("expected assistant variant")
	}
	assistant, _ := assistantUnion.AsPracticeAssistantMessage()
	if _, err := json.Marshal(NewPracticeMessageFromPracticeAssistantMessage(assistant)); err != nil {
		t.Fatalf("marshal valid assistant message: %v", err)
	}

	invalid := map[string]string{
		"user missing clientMessageId": `{"id":"01918fa0-0000-7000-8000-000000006001","seqNo":2,"role":"user","content":"same","createdAt":"2026-07-13T08:03:00Z","replyStatus":"pending"}`,
		"user null clientMessageId":    `{"id":"01918fa0-0000-7000-8000-000000006001","seqNo":2,"role":"user","content":"same","createdAt":"2026-07-13T08:03:00Z","clientMessageId":null,"replyStatus":"pending"}`,
		"user missing replyStatus":     `{"id":"01918fa0-0000-7000-8000-000000006001","seqNo":2,"role":"user","content":"same","createdAt":"2026-07-13T08:03:00Z","clientMessageId":"01918fa0-0000-7000-8000-000000007001"}`,
		"user unknown replyStatus":     `{"id":"01918fa0-0000-7000-8000-000000006001","seqNo":2,"role":"user","content":"same","createdAt":"2026-07-13T08:03:00Z","clientMessageId":"01918fa0-0000-7000-8000-000000007001","replyStatus":"unknown"}`,
		"assistant recovery fields":    `{"id":"01918fa0-0000-7000-8000-000000006002","seqNo":3,"role":"assistant","content":"follow up","createdAt":"2026-07-13T08:03:01Z","clientMessageId":"01918fa0-0000-7000-8000-000000007001","replyStatus":"complete"}`,
		"unknown role":                 `{"id":"01918fa0-0000-7000-8000-000000006002","seqNo":3,"role":"system","content":"hidden","createdAt":"2026-07-13T08:03:01Z"}`,
	}
	for name, body := range invalid {
		t.Run(name, func(t *testing.T) {
			var message PracticeMessage
			if err := json.Unmarshal([]byte(body), &message); err == nil {
				t.Fatal("expected closed union validation error")
			}
		})
	}
}

func TestPracticeMessageMarshalRejectsInvalidConstructedVariants(t *testing.T) {
	validUser := PracticeUserMessage{
		Id:              "01918fa0-0000-7000-8000-000000006001",
		SeqNo:           2,
		Role:            "user",
		Content:         "same message",
		CreatedAt:       "2026-07-13T08:03:00Z",
		ClientMessageId: "01918fa0-0000-7000-8000-000000007001",
		ReplyStatus:     PracticeReplyStatusPending,
	}
	validAssistant := PracticeAssistantMessage{
		Id:        "01918fa0-0000-7000-8000-000000006002",
		SeqNo:     3,
		Role:      "assistant",
		Content:   "follow up",
		CreatedAt: "2026-07-13T08:03:01Z",
	}

	for name, message := range map[string]PracticeMessage{
		"user with assistant role": NewPracticeMessageFromPracticeUserMessage(func() PracticeUserMessage {
			invalid := validUser
			invalid.Role = "assistant"
			return invalid
		}()),
		"user with invalid reply status": NewPracticeMessageFromPracticeUserMessage(func() PracticeUserMessage {
			invalid := validUser
			invalid.ReplyStatus = PracticeReplyStatus("unknown")
			return invalid
		}()),
		"assistant with user role": NewPracticeMessageFromPracticeAssistantMessage(func() PracticeAssistantMessage {
			invalid := validAssistant
			invalid.Role = "user"
			return invalid
		}()),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := json.Marshal(message); err == nil {
				t.Fatal("expected marshal validation error")
			}
		})
	}

	for name, message := range map[string]PracticeMessage{
		"user":      NewPracticeMessageFromPracticeUserMessage(validUser),
		"assistant": NewPracticeMessageFromPracticeAssistantMessage(validAssistant),
	} {
		t.Run("valid "+name, func(t *testing.T) {
			if _, err := json.Marshal(message); err != nil {
				t.Fatalf("marshal valid %s message: %v", name, err)
			}
		})
	}
}
