package review

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/runner"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

type failingReportStatusStore struct{ err error }

func (s failingReportStatusStore) UpdateFeedbackReportStatus(context.Context, ReportStatusUpdate) error {
	return s.err
}

type recordingReportService struct{ called bool }

func (s *recordingReportService) GenerateReport(context.Context, AsyncJob) ReportOutcome {
	s.called = true
	return ReportOutcome{Succeeded: true, AsyncJobFinalized: true}
}

type fixedOutcomeReportService struct{ outcome ReportOutcome }

func (s fixedOutcomeReportService) GenerateReport(context.Context, AsyncJob) ReportOutcome {
	return s.outcome
}

func TestGenerateHandlerRedactsStatusStoreErrorBeforeRunnerPersistence(t *testing.T) {
	service := &recordingReportService{}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:   failingReportStatusStore{err: errors.New("sql failure contains secret transcript marker")},
		Service: service,
	})
	outcome := handler.Handle(context.Background(), runner.ClaimedJob{ResourceID: "report-1"})
	if !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeValidationFailed {
		t.Fatalf("outcome=%+v", outcome)
	}
	if outcome.ErrorMessage != sharederrors.CodeRegistry[sharederrors.CodeValidationFailed].Message || strings.Contains(outcome.ErrorMessage, "secret transcript") {
		t.Fatalf("status-store error was not redacted: %+v", outcome)
	}
	if service.called {
		t.Fatal("service must not run after status transition failure")
	}
}

func TestGenerateHandlerMisconfigurationUsesStableMessage(t *testing.T) {
	var handler *GenerateHandler
	outcome := handler.Handle(context.Background(), runner.ClaimedJob{ResourceID: "report-1"})
	if !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("outcome=%+v", outcome)
	}
	if outcome.ErrorMessage != sharederrors.CodeRegistry[sharederrors.CodeAiOutputInvalid].Message {
		t.Fatalf("misconfiguration message=%q", outcome.ErrorMessage)
	}
}

func TestGenerateHandlerRedactsServiceFailureAtRunnerBoundary(t *testing.T) {
	handler := NewGenerateHandler(GenerateHandlerOptions{Service: fixedOutcomeReportService{outcome: ReportOutcome{
		ErrorCode: sharederrors.CodeAiProviderTimeout, ErrorMessage: "provider leaked raw resume marker", Retryable: true,
	}}})
	outcome := handler.Handle(context.Background(), runner.ClaimedJob{ResourceID: "report-1"})
	if !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("outcome=%+v", outcome)
	}
	if outcome.ErrorMessage != sharederrors.CodeRegistry[sharederrors.CodeAiProviderTimeout].Message || strings.Contains(outcome.ErrorMessage, "raw resume") {
		t.Fatalf("service error was not redacted at runner boundary: %+v", outcome)
	}
}
