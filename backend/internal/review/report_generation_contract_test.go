package review

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGenerateReportPersistsValidatedReadyContentAndProvenance(t *testing.T) {
	reportCtx := validGenerationReportContext("en")
	meta := validReportCallMeta("en")
	ai := &conversationReportAI{results: []conversationAIResult{{
		response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"},
		meta:     meta,
	}}}
	repo := &conversationReportRepository{ctx: reportCtx}
	outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
		JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
	})
	if !outcome.Succeeded || !outcome.AsyncJobFinalized || len(ai.payloads) != 1 {
		t.Fatalf("direct report generation did not finish once: succeeded=%t finalized=%t code=%s calls=%d", outcome.Succeeded, outcome.AsyncJobFinalized, outcome.ErrorCode, len(ai.payloads))
	}
	got := repo.persisted
	if got.Content.Summary == "" || got.Content.PreparednessLevel != sharedtypes.ReadinessTierNeedsPractice || len(got.Content.DimensionAssessments) != 1 || len(got.Content.Issues) != 1 || len(got.Content.NextActions) != 1 {
		t.Fatal("direct model fields were not persisted losslessly")
	}
	dimension := got.Content.DimensionAssessments[0]
	issue := got.Content.Issues[0]
	if dimension.Code != "technical_depth" || dimension.Label != "Technical depth" || dimension.Status != sharedtypes.DimensionStatusNeedsWork ||
		issue.DimensionCode != dimension.Code || !reflect.DeepEqual(issue.SourceMessageSeqNos, []int32{2}) ||
		!reflect.DeepEqual(got.Content.RetryFocusDimensionCodes, []string{dimension.Code}) {
		t.Fatal("direct semantic relationships drifted")
	}
	if got.PromptVersion != meta.PromptVersion || got.RubricVersion != meta.RubricVersion || got.ModelID != meta.ModelID || got.Provider != meta.Provider || got.Language != meta.Language || got.FeatureFlag != meta.FeatureFlag || got.DataSourceVersion != meta.DataSourceVersion {
		t.Fatal("actual provider provenance was not persisted")
	}
	if issues := validateReportContent(got.Content, reportCtx.FrozenContext, reportCtx.Messages); len(issues) != 0 {
		t.Fatalf("persisted direct report no longer satisfies the business validator: %v", issues)
	}
	if len(ai.payloads[0].Messages) != 2 || ai.payloads[0].Messages[0].Role != "system" || ai.payloads[0].Messages[1].Role != "user" {
		t.Fatal("report prompt trust boundary drifted")
	}
	t.Log("REPORT_DIRECT_READY_PASS")

	legacyCount := countActiveReportLegacyIdentifiers(t)
	if legacyCount != 0 {
		t.Fatalf("active report implementation still contains %d legacy identifier occurrences", legacyCount)
	}
	t.Log("legacy_identifier_count=0")
	t.Log("REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS")

}

func TestGenerateReportRepairsInvalidOutputAndKeepsRetriesActionLocal(t *testing.T) {
	t.Run("one path-code-only repair succeeds", func(t *testing.T) {
		reportCtx := validGenerationReportContext("en")
		invalid := strings.Replace(validDirectReportJSON("en"), `"sourceMessageSeqNos":[2]`, `"sourceMessageSeqNos":[1]`, 1)
		ai := &conversationReportAI{results: []conversationAIResult{
			{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
		}}
		repo := &conversationReportRepository{ctx: reportCtx}
		outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
			JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
		})
		if !outcome.Succeeded || len(ai.payloads) != 2 || repo.providerAdmissionCount != 2 || repo.persisted.ReportID == "" {
			t.Fatalf("single repair did not produce a direct ready report: succeeded=%t code=%s providerCalls=%d attemptCount=%d", outcome.Succeeded, outcome.ErrorCode, len(ai.payloads), repo.providerAdmissionCount)
		}
		if ai.payloads[0].Messages[1].Content != ai.payloads[1].Messages[1].Content || strings.Contains(ai.payloads[1].Messages[0].Content, invalid) || !strings.Contains(ai.payloads[1].Messages[0].Content, `"path"`) || !strings.Contains(ai.payloads[1].Messages[0].Content, `"code"`) {
			t.Fatal("repair did not retain the frozen user message with path/code-only trusted guidance")
		}
		t.Log("output_retry_provider_calls=2")
		t.Log("REPORT_OUTPUT_RETRY_PASS")
	})

	t.Run("assistant anchor repair exposes only trusted user sequence coordinates", func(t *testing.T) {
		reportCtx := validGenerationReportContext("en")
		reportCtx.FrozenContext.Conversation.MessageCount = 7
		reportCtx.FrozenContext.Conversation.LastMessageSeqNo = 7
		reportCtx.Messages = []MessageSnapshot{
			{Role: "assistant", Content: "PRIVATE-QUESTION-1", SeqNo: 1},
			{Role: "user", Content: "PRIVATE-ANSWER-2", SeqNo: 2},
			{Role: "assistant", Content: "PRIVATE-QUESTION-3", SeqNo: 3},
			{Role: "user", Content: "PRIVATE-ANSWER-4", SeqNo: 4},
			{Role: "assistant", Content: "PRIVATE-QUESTION-5", SeqNo: 5},
			{Role: "user", Content: "PRIVATE-ANSWER-6", SeqNo: 6},
			{Role: "assistant", Content: "PRIVATE-QUESTION-7", SeqNo: 7},
		}
		invalid := strings.Replace(validDirectReportJSON("en"), `"sourceMessageSeqNos":[2]`, `"sourceMessageSeqNos":[7]`, 1)
		invalid = strings.Replace(invalid, `"preparednessLevel":"needs_practice"`, `"preparednessLevel":"basically_ready"`, 1)
		ai := &conversationReportAI{results: []conversationAIResult{
			{response: aiclient.CompleteResponse{Content: invalid, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
		}}
		repo := &conversationReportRepository{ctx: reportCtx}
		outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
			JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
		})
		if !outcome.Succeeded || len(ai.payloads) != 2 {
			t.Fatalf("assistant-anchor repair outcome=%+v calls=%d", outcome, len(ai.payloads))
		}
		initial, repair := ai.payloads[0].Messages, ai.payloads[1].Messages
		if repair[1].Content != initial[1].Content {
			t.Fatal("assistant-anchor repair changed the untrusted context payload")
		}
		const allowlist = "Valid candidate user message sequence numbers for evidence anchors: [2,4,6]."
		if !strings.Contains(repair[0].Content, allowlist) {
			t.Fatalf("assistant-anchor repair guidance missing trusted allowlist %q: %s", allowlist, repair[0].Content)
		}
		const readinessIntent = "A basically_ready report cannot contain any needs_work dimension."
		if !strings.Contains(repair[0].Content, readinessIntent) {
			t.Fatalf("assistant-anchor repair guidance missing readiness intent %q: %s", readinessIntent, repair[0].Content)
		}
		for _, forbidden := range []string{
			"PRIVATE-QUESTION-1", "PRIVATE-ANSWER-2", "PRIVATE-QUESTION-3", "PRIVATE-ANSWER-4",
			"PRIVATE-QUESTION-5", "PRIVATE-ANSWER-6", "PRIVATE-QUESTION-7",
		} {
			if strings.Contains(repair[0].Content, forbidden) {
				t.Fatalf("assistant-anchor trusted guidance leaked message content %q", forbidden)
			}
		}
	})

	t.Run("four invalid outputs fail without partial ready", func(t *testing.T) {
		reportCtx := validGenerationReportContext("en")
		ai := &conversationReportAI{results: []conversationAIResult{
			{response: aiclient.CompleteResponse{Content: `{"summary":"bad"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"still bad"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"bad again"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"still invalid"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
		}}
		repo := &conversationReportRepository{ctx: reportCtx}
		outcome := newConversationReportService(ai, repo).GenerateReport(context.Background(), AsyncJob{
			JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4,
		})
		if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid || len(ai.payloads) != 4 || repo.providerAdmissionCount != 4 || repo.persisted.ReportID != "" || repo.failed.ReportID == "" {
			t.Fatalf("four invalid outputs did not fail closed: succeeded=%t retryable=%t code=%s calls=%d attemptCount=%d", outcome.Succeeded, outcome.Retryable, outcome.ErrorCode, len(ai.payloads), repo.providerAdmissionCount)
		}
		t.Log("four_invalid_provider_calls=4")
		t.Log("REPORT_FOUR_INVALID_FAIL_CLOSED_PASS")
	})

	t.Run("new user action starts a fresh action-local retry session", func(t *testing.T) {
		reportCtx := validGenerationReportContext("en")
		ai := &conversationReportAI{results: []conversationAIResult{
			{response: aiclient.CompleteResponse{Content: `{"summary":"bad"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"still bad"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"bad again"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: `{"summary":"last bad"}`, FinishReason: "stop"}, meta: validReportCallMeta("en")},
			{response: aiclient.CompleteResponse{Content: validDirectReportJSON("en"), FinishReason: "stop"}, meta: validReportCallMeta("en")},
		}}
		repo := &conversationReportRepository{ctx: reportCtx}
		var waits []time.Duration
		svc := newConversationReportServiceWithWait(ai, repo, func(_ context.Context, delay time.Duration) error {
			waits = append(waits, delay)
			return nil
		})
		first := svc.GenerateReport(context.Background(), AsyncJob{JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 1, MaxAttempts: 4})
		second := svc.GenerateReport(context.Background(), AsyncJob{JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 3, MaxAttempts: 5})
		if first.Succeeded || first.ErrorCode != sharederrors.CodeAiOutputInvalid || !second.Succeeded || len(ai.payloads) != 5 || repo.providerAdmissionCount != 5 {
			t.Fatalf("action reset failed: firstCode=%s secondSucceeded=%t calls=%d admissions=%d", first.ErrorCode, second.Succeeded, len(ai.payloads), repo.providerAdmissionCount)
		}
		if !reflect.DeepEqual(waits, []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}) {
			t.Fatalf("action retry waits=%v", waits)
		}
		t.Log("first_action_call_count=4")
		t.Log("second_action_initial_attempt=1")
		t.Log("retry_state_destroyed_after_action=true")
		t.Log("REPORT_ACTION_RETRY_RESET_PASS")
	})

	t.Run("provider retries stay in one action with 10 20 40 and attempt four terminal", func(t *testing.T) {
		reportCtx := validGenerationReportContext("en")
		providerErr := sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "provider timeout", true)
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
			JobID: testUUID(8), ResourceID: reportCtx.Session.ReportID, Attempts: 3, MaxAttempts: 5,
		})
		if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
			t.Fatalf("provider retry outcome=%+v", outcome)
		}
		if len(ai.payloads) != 4 || repo.providerAdmissionCount != 4 {
			t.Fatalf("providerCalls=%d admissions=%d", len(ai.payloads), repo.providerAdmissionCount)
		}
		if !reflect.DeepEqual(waits, []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second}) {
			t.Fatalf("provider retry waits=%v", waits)
		}
		t.Log("action_retry_schedule=10s,20s,40s")
		t.Log("async_attempts_affect_product_attempt=false")
		t.Log("attempt_four_terminal=true")
		t.Log("REPORT_RETRY_LAYER_SEPARATION_PASS")
	})
}

func TestActiveReportLegacyIdentifierSurfaceIsClean(t *testing.T) {
	if count := countActiveReportLegacyIdentifiers(t); count != 0 {
		t.Fatalf("active report cross-layer surface still contains %d non-allowlisted legacy identifier occurrences", count)
	}
}

func countActiveReportLegacyIdentifiers(t *testing.T) int {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve evidence test path")
	}
	backendInternal := filepath.Dir(filepath.Dir(currentFile))
	repoRoot := filepath.Dir(filepath.Dir(backendInternal))
	roots := []string{
		filepath.Join(backendInternal, "review"),
		filepath.Join(backendInternal, "store", "review"),
		filepath.Join(backendInternal, "api", "reports"),
		filepath.Join(backendInternal, "api", "generated"),
		filepath.Join(repoRoot, "frontend", "src", "api", "generated"),
		filepath.Join(repoRoot, "openapi", "openapi.yaml"),
		filepath.Join(repoRoot, "openapi", "fixtures"),
	}
	legacy := []string{
		"dimension_scores", "dimensionScores", "retry_round", "retryFocusCompetencyCodes", "retry_focus_competency_codes",
		"focusCompetencyCodes", "focus_competency_codes", "retryFocusTurnIds", "retry_focus_turn_ids",
		"questionAssessments", "question_assessments", "DimensionResult",
	}
	allowed := map[string]map[string]int{
		"openapi/fixtures/PROTOTYPE_MAPPING.md": {
			"retryFocusCompetencyCodes": 1,
		},
	}
	allowedExtensions := map[string]bool{
		".go": true, ".json": true, ".md": true, ".sh": true,
		".ts": true, ".tsx": true, ".yaml": true, ".yml": true,
	}
	count := 0
	scannedFiles := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || strings.HasSuffix(path, "_test.go") || !allowedExtensions[filepath.Ext(path)] {
				return nil
			}
			scannedFiles++
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			for _, identifier := range legacy {
				occurrences := bytes.Count(raw, []byte(identifier))
				if occurrences == 0 {
					continue
				}
				allowedCount := allowed[relative][identifier]
				if occurrences > allowedCount {
					count += occurrences - allowedCount
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan active report legacy surface %s: %v", root, err)
		}
	}
	if scannedFiles < len(roots) {
		t.Fatalf("active report legacy scan covered only %d files for %d roots", scannedFiles, len(roots))
	}
	return count
}
