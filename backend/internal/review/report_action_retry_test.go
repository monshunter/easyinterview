package review

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestGenerateReportUsesInitialCallPlusThreeOutputRetriesAcrossRepairScopes(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	validLabel := "Add executable rollback steps and replay this round"
	invalidWhole := strings.Replace(validDirectReportJSON("en"), `"sourceMessageSeqNos":[2]`, `"sourceMessageSeqNos":[1]`, 1)
	invalidLabel := strings.TrimSpace(strings.Repeat("word ", 25))
	labelOnlyInvalid := replaceReportActionLabel(t, validDirectReportJSON("en"), validLabel, invalidLabel)

	metas := []aiclient.AICallMeta{
		validReportCallMeta("en"),
		validReportCallMeta("en"),
		validReportCallMeta("en"),
		validReportCallMeta("en"),
	}
	for index := range metas {
		metas[index].InputTokens = 100 + index
		metas[index].OutputTokens = 50 + index
		metas[index].CostUSDMicros = int64(1000 + index)
		metas[index].LatencyMs = int64(500 + index)
	}
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: invalidWhole, FinishReason: "stop"}, meta: metas[0]},
		{response: aiclient.CompleteResponse{Content: labelOnlyInvalid, FinishReason: "stop"}, meta: metas[1]},
		{response: aiclient.CompleteResponse{Content: actionLabelRepairJSON(t, 0, invalidLabel), FinishReason: "stop"}, meta: metas[2]},
		{response: aiclient.CompleteResponse{Content: actionLabelRepairJSON(t, 0, validLabel), FinishReason: "stop"}, meta: metas[3]},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})

	if !outcome.Succeeded || !outcome.AsyncJobFinalized || len(ai.payloads) != 4 {
		t.Fatalf("outcome=%+v providerCalls=%d", outcome, len(ai.payloads))
	}
	if repo.providerAdmissionCount != 4 || repo.persisted.Content.NextActions[0].Label != validLabel {
		t.Fatalf("attemptCount=%d persisted=%+v", repo.providerAdmissionCount, repo.persisted.Content)
	}
	if strings.Contains(ai.payloads[1].Messages[1].Content, actionLabelRepairStartMarker) {
		t.Fatal("first semantic violation must use whole-report repair")
	}
	for _, index := range []int{2, 3} {
		if !strings.Contains(ai.payloads[index].Messages[1].Content, actionLabelRepairStartMarker) {
			t.Fatalf("attempt %d must use action-label repair", index+1)
		}
	}
}

func TestGenerateReportRetrySessionIsScopedToOneUserAction(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-1"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-2"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-3"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-4"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	var waits []time.Duration
	svc := newConversationReportServiceWithWait(ai, repo, func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	})
	job := AsyncJob{JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 5}

	first := svc.GenerateReport(context.Background(), job)
	second := svc.GenerateReport(context.Background(), job)
	if first.Succeeded || first.Retryable || first.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("first action outcome=%+v", first)
	}
	if !second.Succeeded || !second.AsyncJobFinalized {
		t.Fatalf("second action outcome=%+v", second)
	}
	if len(ai.payloads) != 5 {
		t.Fatalf("provider calls=%d want first-action 4 plus reset action 1", len(ai.payloads))
	}
	if want := []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}; !reflect.DeepEqual(waits, want) {
		t.Fatalf("retry waits=%v want=%v", waits, want)
	}
}

func TestGenerateReportRetryableProviderFailuresStayInsideActionWithExactBackoff(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	providerErr := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "redacted transient failure", true)
	ai := &conversationReportAI{results: []conversationAIResult{
		{err: providerErr, meta: validReportCallMeta("en")},
		{err: providerErr, meta: validReportCallMeta("en")},
		{err: providerErr, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	var waits []time.Duration
	svc := newConversationReportServiceWithWait(ai, repo, func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	})
	outcome := svc.GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 3, MaxAttempts: 5,
	})

	if !outcome.Succeeded || !outcome.AsyncJobFinalized || len(ai.payloads) != 4 {
		t.Fatalf("outcome=%+v providerCalls=%d", outcome, len(ai.payloads))
	}
	if want := []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}; !reflect.DeepEqual(waits, want) {
		t.Fatalf("retry waits=%v want=%v", waits, want)
	}
}

func TestWaitForReportRetryHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if err := waitForReportRetry(ctx, 10*time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("waitForReportRetry error=%v want context.Canceled", err)
	}
	if elapsed := time.Since(started); elapsed >= time.Second {
		t.Fatalf("canceled wait returned after %s", elapsed)
	}
}

func TestGenerateReportAssertsLeaseBeforeCallingProvider(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	repo := &conversationReportRepository{ctx: reportCtx}
	ai := &attemptObservingReportAI{repo: repo}
	job := AsyncJob{JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), job)
	if !outcome.Succeeded || len(ai.observedCounts) != 1 || ai.observedCounts[0] != 1 {
		t.Fatalf("outcome=%+v observedCounts=%v", outcome, ai.observedCounts)
	}
	if repo.assertedJobID != job.JobID || repo.assertedAttempts != job.Attempts {
		t.Fatalf("asserted lease job=%s attempts=%d", repo.assertedJobID, repo.assertedAttempts)
	}
}

func TestGenerateReportAggregatesAllFourCallMeta(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	metas := make([]aiclient.AICallMeta, 4)
	for index := range metas {
		metas[index] = validReportCallMeta("en")
		metas[index].InputTokens = (index + 1) * 10
		metas[index].OutputTokens = index + 1
		metas[index].CostUSDMicros = int64((index + 1) * 100)
		metas[index].LatencyMs = int64((index + 1) * 50)
	}
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-1"}`, FinishReason: "stop"}, meta: metas[0]},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-2"}`, FinishReason: "stop"}, meta: metas[1]},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-3"}`, FinishReason: "stop"}, meta: metas[2]},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: metas[3]},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	result, attemptCount, err := newConversationReportService(ai, repo).generateReportWithActionRetries(context.Background(), reportCtx, AsyncJob{JobID: testUUID(8), Attempts: 1, MaxAttempts: 4})
	if err != nil {
		t.Fatalf("generateReportWithActionRetries: %v", err)
	}
	if attemptCount != 4 || repo.providerAdmissionCount != 4 || len(ai.payloads) != 4 {
		t.Fatalf("attemptCount=%d admissions=%d calls=%d", attemptCount, repo.providerAdmissionCount, len(ai.payloads))
	}
	if result.Meta.InputTokens != 100 || result.Meta.OutputTokens != 10 || result.Meta.CostUSDMicros != 1000 || result.Meta.LatencyMs != 500 {
		t.Fatalf("aggregate meta=%+v", result.Meta)
	}
}

func TestGenerateReportCanSucceedOnThirdAttempt(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-1"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-2"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || !outcome.AsyncJobFinalized || len(ai.payloads) != 3 || repo.providerAdmissionCount != 3 {
		t.Fatalf("outcome=%+v calls=%d attemptCount=%d", outcome, len(ai.payloads), repo.providerAdmissionCount)
	}
}

func TestGenerateReportRetriesDecoratedGenericSchemaFailureAsWholeReport(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	invalidMeta := validReportCallMeta("en")
	invalidMeta.ValidationStatus = aiclient.ValidationStatusInvalid
	invalidMeta.ErrorCode = sharederrors.CodeAiOutputInvalid
	ai := &conversationReportAI{results: []conversationAIResult{
		{
			response: aiclient.CompleteResponse{Content: `{"summary":"bad"}`, FinishReason: "stop"},
			meta:     invalidMeta,
			err:      sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "output failed schema validation: $.preparednessLevel missing required field", false),
		},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || len(ai.payloads) != 2 || repo.providerAdmissionCount != 2 {
		t.Fatalf("outcome=%+v calls=%d attemptCount=%d", outcome, len(ai.payloads), repo.providerAdmissionCount)
	}
	if strings.Contains(ai.payloads[1].Messages[1].Content, actionLabelRepairStartMarker) {
		t.Fatal("generic decorated schema error must use whole-report retry")
	}
}

func TestGenerateReportRecomputesTargetedFailureIntoWholeReportRetry(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	validLabel := "Add executable rollback steps and replay this round"
	invalidLabel := strings.TrimSpace(strings.Repeat("word ", 25))
	initial := replaceReportActionLabel(t, validDirectReportJSON("en"), validLabel, invalidLabel)
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: initial, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"labels":[]}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || len(ai.payloads) != 3 || repo.providerAdmissionCount != 3 {
		t.Fatalf("outcome=%+v calls=%d attemptCount=%d", outcome, len(ai.payloads), repo.providerAdmissionCount)
	}
	if !strings.Contains(ai.payloads[1].Messages[1].Content, actionLabelRepairStartMarker) {
		t.Fatal("second attempt must be targeted")
	}
	if strings.Contains(ai.payloads[2].Messages[1].Content, actionLabelRepairStartMarker) {
		t.Fatal("invalid targeted envelope must recompute into whole-report repair")
	}
}

func TestGenerateReportConfigurationFailureMakesZeroProviderCalls(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	resolution := validReportResolution()
	resolution.PromptVersion = "unsupported"
	ai := &conversationReportAI{}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportServiceWithResolution(ai, repo, resolution).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("outcome=%+v", outcome)
	}
	if len(ai.payloads) != 0 || repo.providerAdmissionCount != 0 {
		t.Fatalf("providerCalls=%d attemptCount=%d", len(ai.payloads), repo.providerAdmissionCount)
	}
}

func TestGenerateReportFourthInvalidOutputIsTerminalWithinOneAction(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-1"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-2"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-3"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: `{"summary":"bad-4"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	svc := newConversationReportService(ai, repo)
	job := AsyncJob{JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4}

	outcome := svc.GenerateReport(context.Background(), job)
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("outcome=%+v", outcome)
	}
	if len(ai.payloads) != 4 || repo.providerAdmissionCount != 4 || repo.persisted.ReportID != "" {
		t.Fatalf("providerCalls=%d admissions=%d persisted=%q", len(ai.payloads), repo.providerAdmissionCount, repo.persisted.ReportID)
	}
}

func TestGenerateReportFourthRetryableProviderFailureEndsCurrentAction(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	providerErr := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "redacted transient failure", true)
	ai := &conversationReportAI{results: []conversationAIResult{
		{err: providerErr, meta: validReportCallMeta("en")},
		{err: providerErr, meta: validReportCallMeta("en")},
		{err: providerErr, meta: validReportCallMeta("en")},
		{err: providerErr, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	var waits []time.Duration
	svc := newConversationReportServiceWithWait(ai, repo, func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	})
	outcome := svc.GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 4, MaxAttempts: 5,
	})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("outcome=%+v", outcome)
	}
	if len(ai.payloads) != 4 || repo.providerAdmissionCount != 4 {
		t.Fatalf("providerCalls=%d admissions=%d", len(ai.payloads), repo.providerAdmissionCount)
	}
	if want := []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}; !reflect.DeepEqual(waits, want) {
		t.Fatalf("retry waits=%v want=%v", waits, want)
	}
}

func TestGenerateReportNonRetryableProviderFailureConsumesNoExtraAttempt(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{{
		err:  sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "secret detail", false),
		meta: validReportCallMeta("en"),
	}}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})

	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderSecretMissing {
		t.Fatalf("outcome=%+v", outcome)
	}
	if len(ai.payloads) != 1 || repo.providerAdmissionCount != 1 {
		t.Fatalf("providerCalls=%d attemptCount=%d", len(ai.payloads), repo.providerAdmissionCount)
	}
}

func TestGenerateReportProviderTimeoutRetriesInsideActionThenPersistsSuccess(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{
		{err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "transient timeout", true), meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	var waits []time.Duration
	svc := newConversationReportServiceWithWait(ai, repo, func(_ context.Context, delay time.Duration) error {
		waits = append(waits, delay)
		return nil
	})
	outcome := svc.GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || !outcome.AsyncJobFinalized || repo.persisted.ReportID != reportCtx.Session.ReportID {
		t.Fatalf("outcome=%+v persisted=%+v", outcome, repo.persisted)
	}
	if len(ai.payloads) != 2 || repo.providerAdmissionCount != 2 {
		t.Fatalf("providerCalls=%d admissions=%d", len(ai.payloads), repo.providerAdmissionCount)
	}
	if want := []time.Duration{10 * time.Second}; !reflect.DeepEqual(waits, want) {
		t.Fatalf("retry waits=%v want=%v", waits, want)
	}
}

func TestGenerateReportDeadlineRetriesButExplicitCancellationIsTerminal(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ai := &conversationReportAI{results: []conversationAIResult{
		{err: context.DeadlineExceeded, meta: validReportCallMeta("en")},
		{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
	}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 5,
	})
	if !outcome.Succeeded || len(ai.payloads) != 2 {
		t.Fatalf("deadline retry outcome=%+v providerCalls=%d", outcome, len(ai.payloads))
	}

	cancelAI := &conversationReportAI{results: []conversationAIResult{{err: context.Canceled, meta: validReportCallMeta("en")}}}
	cancelRepo := &conversationReportRepository{ctx: reportCtx}
	cancelOutcome := newConversationReportService(cancelAI, cancelRepo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 5,
	})
	if cancelOutcome.Succeeded || cancelOutcome.Retryable || cancelOutcome.ErrorCode != sharederrors.CodeAiProviderTimeout || len(cancelAI.payloads) != 1 {
		t.Fatalf("explicit cancellation outcome=%+v providerCalls=%d", cancelOutcome, len(cancelAI.payloads))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	preCanceledRepo := &conversationReportRepository{ctx: reportCtx}
	outcome = newConversationReportService(&conversationReportAI{}, preCanceledRepo).GenerateReport(ctx, AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout || preCanceledRepo.providerAdmissionCount != 0 {
		t.Fatalf("pre-canceled outcome=%+v admissions=%d", outcome, preCanceledRepo.providerAdmissionCount)
	}
}

func TestGenerateReportPersistsTerminalCancellationWithDetachedBoundedContext(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "raw context cancellation", err: context.Canceled},
		{name: "adapter wrapped cancellation as timeout", err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "stable timeout", true)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reportCtx := validGenerationReportContext("en")
			ctx, cancel := context.WithCancel(context.Background())
			ai := cancelingReportAI{cancel: cancel, err: tc.err}
			repo := &conversationReportRepository{ctx: reportCtx}
			outcome := newConversationReportService(ai, repo).GenerateReport(ctx, AsyncJob{
				JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
			})

			if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
				t.Fatalf("outcome=%+v", outcome)
			}
			if repo.providerAdmissionCount != 1 || repo.failed.ReportID != reportCtx.Session.ReportID || repo.failureCtxErr != nil {
				t.Fatalf("attemptCount=%d failed=%+v failureCtxErr=%v", repo.providerAdmissionCount, repo.failed, repo.failureCtxErr)
			}
		})
	}
}

func TestGenerateReportPersistsRecoverableDeadlineWithDetachedBoundedContext(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(deadlineWaitingReportAI{}, repo).GenerateReport(ctx, AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if outcome.Succeeded || !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("outcome=%+v", outcome)
	}
	if repo.providerAdmissionCount != 1 || repo.failed.ReportID != reportCtx.Session.ReportID || repo.failureCtxErr != nil {
		t.Fatalf("attemptCount=%d failed=%+v failureCtxErr=%v", repo.providerAdmissionCount, repo.failed, repo.failureCtxErr)
	}
}

func TestGenerateReportPersistsValidResultWhenCallCompletesAsContextIsCanceled(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	ctx, cancel := context.WithCancel(context.Background())
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(cancelingValidReportAI{cancel: cancel}, repo).GenerateReport(ctx, AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || !outcome.AsyncJobFinalized || repo.persisted.ReportID != reportCtx.Session.ReportID || repo.resultCtxErr != nil {
		t.Fatalf("outcome=%+v persisted=%+v resultCtxErr=%v", outcome, repo.persisted, repo.resultCtxErr)
	}
	if repo.providerAdmissionCount != 1 {
		t.Fatalf("attemptCount=%d", repo.providerAdmissionCount)
	}
}

func TestClassifyReportGenerationErrorUsesTypedRetryabilityAndCancellation(t *testing.T) {
	if failure := classifyReportGenerationError(context.DeadlineExceeded); !failure.Retryable || failure.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("deadline failure=%+v", failure)
	}
	if failure := classifyReportGenerationError(context.Canceled); failure.Retryable || failure.Code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("cancel failure=%+v", failure)
	}
	wrapped := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "typed nonretryable", false)
	if failure := classifyReportGenerationError(wrapped); failure.Retryable {
		t.Fatalf("typed nonretryable failure=%+v", failure)
	}
}

type cancelingReportAI struct {
	cancel context.CancelFunc
	err    error
}

type attemptObservingReportAI struct {
	repo           *conversationReportRepository
	observedCounts []int
}

type deadlineWaitingReportAI struct{}

func (deadlineWaitingReportAI) Complete(ctx context.Context, _ string, _ aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	<-ctx.Done()
	return aiclient.CompleteResponse{}, validReportCallMeta("en"), sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "stable timeout", true)
}

type cancelingValidReportAI struct {
	cancel context.CancelFunc
}

func (f cancelingValidReportAI) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.cancel()
	return aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, validReportCallMeta("en"), nil
}

func (f *attemptObservingReportAI) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.observedCounts = append(f.observedCounts, f.repo.providerAdmissionCount)
	return aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, validReportCallMeta("en"), nil
}

func (f cancelingReportAI) Complete(context.Context, string, aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.cancel()
	return aiclient.CompleteResponse{}, validReportCallMeta("en"), f.err
}
