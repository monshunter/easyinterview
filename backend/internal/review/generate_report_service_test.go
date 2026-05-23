package review

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestGenerateReportFailedMatrix(t *testing.T) {
	baseCtx := ReportContext{
		Session: sampleSession(),
		Plan:    samplePlan(),
		Turns:   []TurnSnapshot{{ID: "turn-1", TurnIndex: 1, QuestionIntent: "architecture", QuestionContext: "service boundary", AnswerSummary: "clear"}},
	}
	successContent := `{"summary":"ok","highlights":[],"issues":[],"next_actions":[{"type":"next_round","label":"Next"}]}`
	successAssessment := `{"dimension_results":{"depth":{"status":"meets_bar","confidence":0.8,"score":0.8}},"overall_status":"meets_bar","confidence":0.8,"strengths":["clear"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`

	cases := []struct {
		name             string
		resolverErrs     map[string]error
		aiResponses      []string
		aiErrs           []error
		wantCode         string
		wantRetryable    bool
		wantAICalls      int
		wantExplicitRuns int
		wantCapability   aiclient.AITaskRunCapability
		wantValidation   aiclient.ValidationStatus
	}{
		{
			name:             "f3_prompt_unsupported",
			resolverErrs:     map[string]error{reportGenerateFeatureKey: registry.ErrPromptUnsupported},
			wantCode:         sharederrors.CodeAiProviderConfigInvalid,
			wantAICalls:      0,
			wantExplicitRuns: 1,
			wantCapability:   aiclient.AITaskRunTaskReportGenerate,
		},
		{
			name:             "f3_language_unsupported_assessment",
			resolverErrs:     map[string]error{reportQuestionAssessmentFeatureKey: registry.ErrLanguageUnsupported},
			aiResponses:      []string{successContent},
			wantCode:         sharederrors.CodeAiProviderConfigInvalid,
			wantAICalls:      1,
			wantExplicitRuns: 1,
			wantCapability:   aiclient.AITaskRunTaskReportAssessment,
		},
		{
			name:          "a3_secret_missing",
			aiErrs:        []error{sharederrors.Wrap(sharederrors.CodeAiProviderSecretMissing, "missing secret", false)},
			wantCode:      sharederrors.CodeAiProviderSecretMissing,
			wantAICalls:   1,
			wantRetryable: false,
		},
		{
			name:          "a3_timeout",
			aiErrs:        []error{sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "timeout", true)},
			wantCode:      sharederrors.CodeAiProviderTimeout,
			wantAICalls:   1,
			wantRetryable: true,
		},
		{
			name:          "a3_invalid_output",
			aiErrs:        []error{sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "invalid", false)},
			wantCode:      sharederrors.CodeAiOutputInvalid,
			wantAICalls:   1,
			wantRetryable: false,
		},
		{
			name:             "parsed_empty",
			aiResponses:      []string{`{}`},
			wantCode:         sharederrors.CodeAiOutputInvalid,
			wantAICalls:      1,
			wantRetryable:    false,
			wantExplicitRuns: 1,
			wantCapability:   aiclient.AITaskRunTaskReportGenerate,
			wantValidation:   aiclient.ValidationStatusInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			responses := tc.aiResponses
			if len(responses) == 0 && len(tc.aiErrs) == 0 {
				responses = []string{successContent, successAssessment}
			}
			ai := &fakeReviewAI{responses: responses, errs: tc.aiErrs}
			repo := &fakeReportRepository{ctx: baseCtx}
			runs := &fakeTaskRunWriter{}
			svc := NewService(ServiceOptions{
				Registry: fakePromptResolver{
					resolutions: map[string]registry.PromptResolution{
						reportGenerateFeatureKey:           reportResolution(reportGenerateFeatureKey, "report.generate.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nReturn strict JSON."),
						reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Question context: {{question_context}}\nAnswer summary: {{answer_summary}}\nReturn strict JSON."),
					},
					errs: tc.resolverErrs,
				},
				AI:         ai,
				AITaskRuns: runs,
				Repository: repo,
				Now:        func() time.Time { return time.Date(2026, 5, 15, 20, 0, 0, 0, time.UTC) },
				NewID:      fixedIDs("run-1", "outbox-1", "audit-1"),
			})

			outcome := svc.GenerateReport(context.Background(), AsyncJob{JobID: "job-1", ResourceID: sampleSession().ReportID, Attempts: 1, MaxAttempts: 5})
			if outcome.Succeeded || outcome.ErrorCode != tc.wantCode || outcome.Retryable != tc.wantRetryable {
				t.Fatalf("outcome = %+v, want code=%s retryable=%v", outcome, tc.wantCode, tc.wantRetryable)
			}
			if outcome.AsyncJobFinalized {
				t.Fatalf("failure outcome must leave async job finalization to the runner kernel")
			}
			if repo.failed.ErrorCode != tc.wantCode || repo.failed.Retryable != tc.wantRetryable {
				t.Fatalf("persisted failure = %+v", repo.failed)
			}
			if len(ai.payloads) != tc.wantAICalls {
				t.Fatalf("AI calls = %d, want %d", len(ai.payloads), tc.wantAICalls)
			}
			if len(runs.rows) != tc.wantExplicitRuns {
				t.Fatalf("explicit ai_task_runs = %d, want %d rows=%+v", len(runs.rows), tc.wantExplicitRuns, runs.rows)
			}
			if tc.wantExplicitRuns > 0 {
				row := runs.rows[0]
				if row.Capability != tc.wantCapability || row.Status != aiclient.AITaskRunStatusFailed || row.ErrorCode != tc.wantCode {
					t.Fatalf("ai_task_run row = %+v", row)
				}
				if row.ValidationStatus != tc.wantValidation {
					t.Fatalf("validation status = %q, want %q", row.ValidationStatus, tc.wantValidation)
				}
			}
		})
	}
}

type fakeTaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (f *fakeTaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	f.rows = append(f.rows, row)
	return nil
}

var _ = errors.Is
