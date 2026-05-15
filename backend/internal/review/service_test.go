package review

import (
	"context"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGenerateReportServiceOrchestratesAndPersists(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{
		`{"summary":"ok","highlights":[{"dimension":"depth","evidence":"good","confidence":0.8}],"issues":[],"next_actions":[]}`,
		`{"dimension_results":{"depth":{"status":"meets_bar","confidence":0.8,"score":0.8}},"overall_status":"meets_bar","confidence":0.8,"strengths":["clear"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`,
	}}
	repo := &fakeReportRepository{ctx: ReportContext{
		Session: sampleSession(),
		Plan:    samplePlan(),
		Turns:   []TurnSnapshot{{ID: "turn-1", TurnIndex: 1, QuestionIntent: "architecture", QuestionContext: "service boundary", AnswerSummary: "clear"}},
		Rubric:  registry.RubricSchema{Dimensions: []registry.RubricDimension{{Name: "depth", Weight: 1}}},
	}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportGenerateFeatureKey:           reportResolution(reportGenerateFeatureKey, "report.generate.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nRubric: {{rubric_dimensions}}\nReturn strict JSON."),
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nQuestion context: {{question_context}}\nAnswer summary: {{answer_summary}}\nRubric: {{rubric}}\nReturn strict JSON."),
		}},
		AI:         ai,
		Repository: repo,
		Now:        func() time.Time { return time.Date(2026, 5, 15, 19, 0, 0, 0, time.UTC) },
		NewID:      fixedIDs("outbox-id", "audit-id"),
	})

	outcome := svc.GenerateReport(context.Background(), AsyncJob{JobID: "job-1", ResourceID: sampleSession().ReportID})
	if !outcome.Succeeded {
		t.Fatalf("outcome = %+v", outcome)
	}
	if !outcome.AsyncJobFinalized {
		t.Fatalf("outcome.AsyncJobFinalized = false, want true")
	}
	if repo.persisted.ReportID != sampleSession().ReportID ||
		repo.persisted.PreparednessLevel != sharedtypes.ReadinessTierWellPrepared ||
		len(repo.persisted.Assessments) != 1 {
		t.Fatalf("persisted = %+v", repo.persisted)
	}
	if repo.persisted.OutboxEventID != "outbox-id" || repo.persisted.AuditEventID != "audit-id" {
		t.Fatalf("ids = %s/%s", repo.persisted.OutboxEventID, repo.persisted.AuditEventID)
	}
	if len(ai.payloads) != 2 ||
		ai.payloads[0].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportGenerate ||
		ai.payloads[1].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportAssessment {
		t.Fatalf("payloads = %+v", ai.payloads)
	}
}

type fakeReportRepository struct {
	ctx       ReportContext
	persisted ReportResultPersistence
	failed    ReportFailurePersistence
}

func (f *fakeReportRepository) LoadReportContext(_ context.Context, _ string) (ReportContext, error) {
	return f.ctx, nil
}

func (f *fakeReportRepository) PersistReportResult(_ context.Context, in ReportResultPersistence) error {
	f.persisted = in
	return nil
}

func (f *fakeReportRepository) PersistReportFailure(_ context.Context, in ReportFailurePersistence) error {
	f.failed = in
	return nil
}

func fixedIDs(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return "extra-id"
		}
		out := ids[i]
		i++
		return out
	}
}
