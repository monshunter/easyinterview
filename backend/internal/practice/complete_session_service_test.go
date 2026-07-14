package practice

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type completionTestStore struct {
	Store
	err    error
	result CompleteSessionResult
	calls  int
}

func (s *completionTestStore) CompleteSession(context.Context, CompleteSessionStoreInput) (CompleteSessionResult, error) {
	s.calls++
	if s.err != nil {
		return CompleteSessionResult{}, s.err
	}
	if s.result.ReportID != "" {
		return s.result, nil
	}
	return CompleteSessionResult{ReportID: "report-1"}, nil
}

func TestCompleteSessionRejectsZeroAnswer(t *testing.T) {
	store := &completionTestStore{err: ErrSessionNotReportable}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return time.Unix(9, 0).UTC() },
		NewID: func() string { return "id-1" },
	})

	_, err := service.CompletePracticeSession(context.Background(), CompletePracticeSessionRequest{
		UserID: "user-1", SessionID: "session-1", ClientCompletedAt: time.Unix(8, 0).UTC(),
	})
	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) || serviceErr.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("error=%v want VALIDATION_FAILED", err)
	}
	t.Log("ZERO_ANSWER_COMPLETION_REJECTED_PASS")
}

func TestCompletePracticeSessionPreservesFrozenReportContext(t *testing.T) {
	snapshot := completionServiceTestReportContext(t)
	store := &completionTestStore{result: CompleteSessionResult{ReportID: "report-1", GenerationContext: snapshot}}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return time.Unix(9, 0).UTC() },
		NewID: func() string { return "id-1" },
	})

	result, err := service.CompletePracticeSession(context.Background(), CompletePracticeSessionRequest{
		UserID: "user-1", SessionID: "session-1", ClientCompletedAt: time.Unix(8, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("CompletePracticeSession: %v", err)
	}
	if store.calls != 1 || !reflect.DeepEqual(result.GenerationContext, snapshot) {
		t.Fatalf("completion context changed across service boundary: calls=%d result=%+v", store.calls, result.GenerationContext)
	}
	t.Log("REPORT_CONTEXT_SNAPSHOT_PASS")
}

func TestCompleteSessionReplayPreservesReportContext(t *testing.T) {
	snapshot := completionServiceTestReportContext(t)
	store := &completionTestStore{result: CompleteSessionResult{ReportID: "report-1", Replay: true, GenerationContext: snapshot}}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return time.Unix(9, 0).UTC() },
		NewID: func() string { return "id-1" },
	})

	result, err := service.CompletePracticeSession(context.Background(), CompletePracticeSessionRequest{
		UserID: "user-1", SessionID: "session-1", ClientCompletedAt: time.Unix(8, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("CompletePracticeSession replay: %v", err)
	}
	if !result.Replay || store.calls != 1 || !reflect.DeepEqual(result.GenerationContext, snapshot) {
		t.Fatalf("completion replay changed frozen context: calls=%d result=%+v", store.calls, result)
	}
	t.Log("REPORT_CONTEXT_REPLAY_PASS")
}

func completionServiceTestReportContext(t *testing.T) ReportContextSnapshot {
	t.Helper()
	snapshot, err := BuildReportContextSnapshot(ReportContextSnapshotInput{
		TargetJob: ReportTargetJobSnapshot{
			ID: "target-1", Title: "Platform Engineer", Language: "en", RawJD: "complete jd",
			Summary: json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","dataSourceVersion":"target-job.v1"}}`),
		},
		Resume: ReportResumeSnapshot{
			ID: "resume-1", DisplayName: "Resume", Language: "en", SourceSnapshot: "complete resume", StructuredProfile: json.RawMessage(`{}`),
		},
		Plan: ReportPlanSnapshot{
			ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en",
			TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1,
		},
		Conversation: ReportConversationCoordinate{SessionID: "session-1", Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
