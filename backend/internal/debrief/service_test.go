package debrief

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestServicePackageCompiles(t *testing.T) {
	t.Helper()
	if NewService(ServiceOptions{}) == nil {
		t.Fatalf("NewService returned nil")
	}
}

func TestStoreInterface_Compiles(t *testing.T) {
	t.Helper()
	var _ Store = (*compileStore)(nil)
}

func TestServiceCreateDebrief_AuditEmitted(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	store := &recordingStore{result: CreateDebriefResult{
		DebriefID: "debrief-id",
		Job: JobRecord{
			ID:           "job-id",
			JobType:      api.JobTypeDebriefGenerate,
			ResourceType: api.ResourceTypeDebrief,
			ResourceID:   "debrief-id",
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}}
	audit := &recordingAudit{}
	service := NewService(ServiceOptions{
		Store: store,
		Audit: audit,
		Now:   func() time.Time { return now },
		NewID: sequenceID("debrief-id", "job-id", "outbox-id", "audit-id"),
	})
	req := CreateDebriefRequest{
		UserID:          "user-id",
		TargetJobID:     "target-job-id",
		RoundType:       sharedtypes.DebriefRoundTypeBehavioral,
		InterviewerRole: sharedtypes.InterviewerRoleHiringManager,
		Language:        "zh-CN",
		Notes:           "private note",
		Questions: []QuestionInput{{
			QuestionText:        "raw question text",
			MyAnswerSummary:     "raw answer text",
			InterviewerReaction: "raw reaction text",
		}},
	}

	result, err := service.CreateDebrief(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateDebrief returned error: %v", err)
	}
	if result.DebriefID != "debrief-id" || result.Job.ID != "job-id" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if store.calls != 1 || store.last.DebriefID != "debrief-id" || store.last.JobID != "job-id" || store.last.OutboxEventID != "outbox-id" {
		t.Fatalf("store input drifted: calls=%d last=%+v", store.calls, store.last)
	}
	if len(audit.events) != 1 {
		t.Fatalf("audit calls=%d want 1", len(audit.events))
	}
	event := audit.events[0]
	if event.AuditEventID != "audit-id" || event.Action != AuditActionCreateDebrief || event.ResourceType != string(api.ResourceTypeDebrief) || event.ResourceID != "debrief-id" {
		t.Fatalf("audit event drifted: %+v", event)
	}
	wantMetadata := map[string]any{
		"debrief_id":     "debrief-id",
		"target_job_id":  "target-job-id",
		"language":       "zh-CN",
		"question_count": 1,
		"status":         string(sharedtypes.DebriefStatusDraft),
	}
	if !reflect.DeepEqual(event.Metadata, wantMetadata) {
		t.Fatalf("audit metadata = %+v, want %+v", event.Metadata, wantMetadata)
	}
	for _, forbidden := range []string{"raw question text", "raw answer text", "raw reaction text", "private note"} {
		if strings.Contains(metadataString(event.Metadata), forbidden) {
			t.Fatalf("audit metadata leaked forbidden text %q: %+v", forbidden, event.Metadata)
		}
	}
}

func TestServiceSuggestQuestions_Happy(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	targetJobID := "01918fa0-0000-7000-8000-00000000c001"
	contexts := &recordingSuggestionContextStore{result: SuggestionContext{
		TargetJobID: targetJobID,
		Title:       "Staff Frontend Engineer",
		CompanyName: "Example Co",
		Summary:     "Design systems and cross-functional leadership.",
	}}
	reg := &recordingPromptResolver{resolution: registry.PromptResolution{
		FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.suggest_questions.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
		SystemMessage:       "You generate debrief questions.",
		OutputSchema:        testOutputSchema(`{"type":"object","required":["suggestions"],"properties":{"suggestions":{"type":"array"}}}`),
		UserMessageTemplate: "Target: {{targetTitle}}",
	}}
	ai := &recordingAIClient{
		response: aiclient.CompleteResponse{Content: `{"suggestions":[{"stage":"opening","questionText":"What changed after the system migration?","whyLikelyAsked":"The JD stresses change leadership.","source":"jd"},{"questionText":"How did you measure adoption?","whyLikelyAsked":"Metrics are missing from the summary.","source":"manual"}]}`},
		meta: aiclient.AICallMeta{
			Provider:            "stub",
			ModelFamily:         "stub-family",
			ModelID:             "stub-model",
			PromptVersion:       "v0.1.0",
			RubricVersion:       "v0.1.0",
			ModelProfileName:    "debrief.suggest_questions.default",
			ModelProfileVersion: "2026-05-16",
			FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
			FeatureFlag:         "none",
			DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
			Language:            "zh-CN",
			ValidationStatus:    aiclient.ValidationStatusOK,
		},
	}
	taskRuns := &recordingTaskRunWriter{}
	audit := &recordingAudit{}
	service := NewService(ServiceOptions{
		SuggestionContext: contexts,
		Registry:          reg,
		AI:                ai,
		AITaskRuns:        taskRuns,
		Audit:             audit,
		Now:               func() time.Time { return now },
		NewID:             sequenceID("01918fa0-0000-7000-8000-00000000e001", "01918fa0-0000-7000-8000-00000000e002"),
	})

	result, err := service.SuggestQuestions(context.Background(), SuggestQuestionsRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: targetJobID,
		Language:    "zh-CN",
		Count:       2,
	})
	if err != nil {
		t.Fatalf("SuggestQuestions returned error: %v", err)
	}
	if len(result.Suggestions) != 2 || result.Suggestions[0].QuestionText == "" || result.Suggestions[0].Source != sharedtypes.DebriefQuestionSourceJd {
		t.Fatalf("suggestions not parsed: %+v", result.Suggestions)
	}
	if contexts.calls != 1 || contexts.last.TargetJobID != targetJobID {
		t.Fatalf("context lookup drifted: calls=%d last=%+v", contexts.calls, contexts.last)
	}
	if reg.calls != 1 || reg.featureKey != featurekeys.DebriefSuggestQuestions.String() || reg.language != "zh-CN" {
		t.Fatalf("registry call drifted: %+v", reg)
	}
	if ai.calls != 1 || ai.profileName != "debrief.suggest_questions.default" || len(ai.payload.Messages) == 0 {
		t.Fatalf("AI call drifted: calls=%d profile=%s payload=%+v", ai.calls, ai.profileName, ai.payload)
	}
	if ai.payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskDebriefSuggestQuestions ||
		ai.payload.Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		ai.payload.Metadata.TaskRun.ResourceID != targetJobID {
		t.Fatalf("AI task metadata drifted: %+v", ai.payload.Metadata.TaskRun)
	}
	if len(ai.payload.Metadata.OutputSchema) == 0 {
		t.Fatalf("AI metadata OutputSchema must be populated")
	}
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Capability != aiclient.AITaskRunTaskDebriefSuggestQuestions || taskRuns.rows[0].Status != aiclient.AITaskRunStatusSuccess {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
	if len(audit.events) != 1 || audit.events[0].Action != AuditActionSuggestDebriefQuestions || audit.events[0].ResourceID != targetJobID {
		t.Fatalf("audit event drifted: %+v", audit.events)
	}
}

func TestServiceSuggestQuestions_ResumeContextInPrompt(t *testing.T) {
	targetJobID := "01918fa0-0000-7000-8000-00000000c001"
	resumeID := "01918fa0-0000-7000-8000-00000000a001"
	contexts := &recordingSuggestionContextStore{result: SuggestionContext{
		TargetJobID:   targetJobID,
		Title:         "Staff Backend Engineer",
		ResumeSummary: `{"basics":{"headline":"Platform engineer"},"skills":["Go","PostgreSQL"]}`,
	}}
	reg := &recordingPromptResolver{resolution: registry.PromptResolution{
		FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.suggest_questions.default",
		DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
		UserMessageTemplate: "Target: {{targetTitle}}\nResume: {{resumeSummary}}",
	}}
	ai := &recordingAIClient{
		response: aiclient.CompleteResponse{Content: `{"suggestions":[{"questionText":"How did you scale the platform?","whyLikelyAsked":"The resume profile highlights platform engineering.","source":"resume"}]}`},
		meta: aiclient.AICallMeta{
			ValidationStatus: aiclient.ValidationStatusOK,
		},
	}
	service := NewService(ServiceOptions{
		SuggestionContext: contexts,
		Registry:          reg,
		AI:                ai,
		AITaskRuns:        &recordingTaskRunWriter{},
		Now:               func() time.Time { return time.Date(2026, 6, 14, 10, 0, 0, 0, time.UTC) },
		NewID:             sequenceID("01918fa0-0000-7000-8000-00000000e001"),
	})

	result, err := service.SuggestQuestions(context.Background(), SuggestQuestionsRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: targetJobID,
		ResumeID:    resumeID,
		Language:    "zh-CN",
		Count:       1,
	})
	if err != nil {
		t.Fatalf("SuggestQuestions returned error: %v", err)
	}
	if contexts.last.ResumeID != resumeID {
		t.Fatalf("resumeId not passed to context store: %+v", contexts.last)
	}
	if len(result.Suggestions) != 1 || result.Suggestions[0].Source != sharedtypes.DebriefQuestionSourceResume {
		t.Fatalf("suggestions drifted: %+v", result.Suggestions)
	}
	if !strings.Contains(ai.payload.Messages[len(ai.payload.Messages)-1].Content, `"headline":"Platform engineer"`) {
		t.Fatalf("resume structured profile missing from AI prompt: %+v", ai.payload.Messages)
	}
}

func TestServiceSuggestQuestions_SessionContextInPrompt(t *testing.T) {
	targetJobID := "01918fa0-0000-7000-8000-00000000c001"
	sessionID := "01918fa0-0000-7000-8000-000000005000"
	contexts := &recordingSuggestionContextStore{result: SuggestionContext{
		TargetJobID:    targetJobID,
		Title:          "Staff Backend Engineer",
		Summary:        `{"mustHave":["platform reliability"]}`,
		SessionSummary: `{"sessionId":"01918fa0-0000-7000-8000-000000005000","turns":[{"questionText":"How did you measure adoption?","answerSummary":"I cited rollout metrics."}],"report":{"issues":[{"title":"Needs sharper metrics"}]}}`,
	}}
	reg := &recordingPromptResolver{resolution: registry.PromptResolution{
		FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.suggest_questions.default",
		DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
		UserMessageTemplate: "Target: {{role_title}}\nJob: {{job_summary}}\nMock: {{mock_report_summary}}",
	}}
	ai := &recordingAIClient{
		response: aiclient.CompleteResponse{Content: `{"suggestions":[{"questionText":"How did you measure adoption?","whyLikelyAsked":"The mock interview report shows metrics risk.","source":"mock_report"}]}`},
		meta: aiclient.AICallMeta{
			ValidationStatus: aiclient.ValidationStatusOK,
		},
	}
	service := NewService(ServiceOptions{
		SuggestionContext: contexts,
		Registry:          reg,
		AI:                ai,
		AITaskRuns:        &recordingTaskRunWriter{},
		Now:               func() time.Time { return time.Date(2026, 6, 14, 14, 0, 0, 0, time.UTC) },
		NewID:             sequenceID("01918fa0-0000-7000-8000-00000000e001"),
	})

	result, err := service.SuggestQuestions(context.Background(), SuggestQuestionsRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: targetJobID,
		SessionID:   sessionID,
		Language:    "zh-CN",
		Count:       1,
	})
	if err != nil {
		t.Fatalf("SuggestQuestions returned error: %v", err)
	}
	if contexts.last.SessionID != sessionID {
		t.Fatalf("sessionId not passed to context store: %+v", contexts.last)
	}
	if len(result.Suggestions) != 1 || result.Suggestions[0].Source != sharedtypes.DebriefQuestionSourceMockReport {
		t.Fatalf("suggestions drifted: %+v", result.Suggestions)
	}
	prompt := ai.payload.Messages[len(ai.payload.Messages)-1].Content
	if strings.Contains(prompt, "{{mock_report_summary}}") || !strings.Contains(prompt, `"Needs sharper metrics"`) {
		t.Fatalf("mock session summary missing from AI prompt: %s", prompt)
	}
	if strings.Contains(prompt, "{{role_title}}") || strings.Contains(prompt, "{{job_summary}}") {
		t.Fatalf("real prompt markers were not replaced: %s", prompt)
	}
}

func TestServiceGetDebrief_ProvenanceWireOnly(t *testing.T) {
	store := &recordingStore{getResult: DebriefRecord{
		ID:          "01918fa0-0000-7000-8000-00000000d010",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		Status:      sharedtypes.DebriefStatusCompleted,
		Provenance: &Provenance{
			PromptVersion: "v0.1.0",
			RubricVersion: "v0.1.0",
			ModelID:       "stub-model",
			Language:      "zh-CN",
		},
	}}
	service := NewService(ServiceOptions{Store: store})

	got, err := service.GetDebrief(context.Background(), "user-1", "01918fa0-0000-7000-8000-00000000d010")
	if err != nil {
		t.Fatalf("GetDebrief returned error: %v", err)
	}
	if store.getCalls != 1 || store.getUserID != "user-1" || store.getDebriefID != "01918fa0-0000-7000-8000-00000000d010" {
		t.Fatalf("store call drifted: calls=%d user=%s debrief=%s", store.getCalls, store.getUserID, store.getDebriefID)
	}
	if got.Provenance == nil ||
		got.Provenance.PromptVersion != "v0.1.0" ||
		got.Provenance.RubricVersion != "v0.1.0" ||
		got.Provenance.ModelID != "stub-model" ||
		got.Provenance.Language != "zh-CN" ||
		got.Provenance.FeatureFlag != "none" ||
		got.Provenance.DataSourceVersion != "debrief/01918fa0-0000-7000-8000-00000000d010@v1" {
		t.Fatalf("provenance drifted: %+v", got.Provenance)
	}
}

func TestServiceSuggestQuestions_CrossUserTargetJob_403(t *testing.T) {
	contexts := &recordingSuggestionContextStore{err: ErrDebriefPrerequisite}
	reg := &recordingPromptResolver{}
	ai := &recordingAIClient{}
	taskRuns := &recordingTaskRunWriter{}
	audit := &recordingAudit{}
	service := NewService(ServiceOptions{
		SuggestionContext: contexts,
		Registry:          reg,
		AI:                ai,
		AITaskRuns:        taskRuns,
		Audit:             audit,
	})

	_, err := service.SuggestQuestions(context.Background(), SuggestQuestionsRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		Language:    "zh-CN",
		Count:       6,
	})
	if !errors.Is(err, ErrDebriefPrerequisite) {
		t.Fatalf("error=%v, want ErrDebriefPrerequisite", err)
	}
	if reg.calls != 0 || ai.calls != 0 || len(taskRuns.rows) != 0 || len(audit.events) != 0 {
		t.Fatalf("cross-user path should not call downstreams: registry=%d ai=%d taskRuns=%d audit=%d", reg.calls, ai.calls, len(taskRuns.rows), len(audit.events))
	}
}

func TestServiceSuggestQuestions_F3ResolveFailed(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	taskRuns := &recordingTaskRunWriter{}
	service := NewService(ServiceOptions{
		SuggestionContext: &recordingSuggestionContextStore{result: validSuggestionContext()},
		Registry:          &recordingPromptResolver{err: registry.ErrPromptUnsupported},
		AI:                &recordingAIClient{},
		AITaskRuns:        taskRuns,
		Now:               func() time.Time { return now },
		NewID:             sequenceID("01918fa0-0000-7000-8000-00000000e101"),
	})

	_, err := service.SuggestQuestions(context.Background(), validSuggestQuestionsRequest())

	assertServiceErrorCode(t, err, "AI_PROVIDER_CONFIG_INVALID")
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Status != aiclient.AITaskRunStatusFailed || taskRuns.rows[0].ErrorCode != "AI_PROVIDER_CONFIG_INVALID" {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestServiceSuggestQuestions_A3Timeout(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	taskRuns := &recordingTaskRunWriter{}
	service := NewService(ServiceOptions{
		SuggestionContext: &recordingSuggestionContextStore{result: validSuggestionContext()},
		Registry:          &recordingPromptResolver{resolution: validSuggestionResolution()},
		AI:                &recordingAIClient{err: context.DeadlineExceeded},
		AITaskRuns:        taskRuns,
		Now:               func() time.Time { return now },
		NewID:             sequenceID("01918fa0-0000-7000-8000-00000000e102"),
	})

	_, err := service.SuggestQuestions(context.Background(), validSuggestQuestionsRequest())

	assertServiceErrorCode(t, err, "AI_PROVIDER_TIMEOUT")
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Status != aiclient.AITaskRunStatusTimeout || taskRuns.rows[0].ErrorCode != "AI_PROVIDER_TIMEOUT" {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestServiceSuggestQuestions_ParseFailed(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	taskRuns := &recordingTaskRunWriter{}
	service := NewService(ServiceOptions{
		SuggestionContext: &recordingSuggestionContextStore{result: validSuggestionContext()},
		Registry:          &recordingPromptResolver{resolution: validSuggestionResolution()},
		AI: &recordingAIClient{
			response: aiclient.CompleteResponse{Content: `not json`},
			meta:     validSuggestionMeta(),
		},
		AITaskRuns: taskRuns,
		Now:        func() time.Time { return now },
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000e103"),
	})

	_, err := service.SuggestQuestions(context.Background(), validSuggestQuestionsRequest())

	assertServiceErrorCode(t, err, "AI_OUTPUT_INVALID")
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Status != aiclient.AITaskRunStatusFailed || taskRuns.rows[0].ErrorCode != "AI_OUTPUT_INVALID" {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestAITaskRunsWritten(t *testing.T) {
	_, taskRuns := exerciseObservedDebriefFlow(t, "__SECRET_RAW_TEXT__")

	seen := map[aiclient.AITaskRunCapability]bool{}
	for _, row := range taskRuns.rows {
		seen[row.Capability] = true
		if row.FeatureKey == "" || row.ModelProfileName == "" || row.Status == "" || row.ValidationStatus == "" {
			t.Fatalf("task run row missing required observability fields: %+v", row)
		}
	}
	if !seen[aiclient.AITaskRunTaskDebriefGenerate] || !seen[aiclient.AITaskRunTaskDebriefSuggestQuestions] {
		t.Fatalf("task run capabilities missing: seen=%v rows=%+v", seen, taskRuns.rows)
	}
}

func TestAuditEventsWritten(t *testing.T) {
	audit, _ := exerciseObservedDebriefFlow(t, "__SECRET_RAW_TEXT__")

	seen := map[string]bool{}
	for _, event := range audit.events {
		seen[event.Action] = true
	}
	for _, action := range []string{AuditActionCreateDebrief, AuditActionCompleteDebrief, AuditActionSuggestDebriefQuestions} {
		if !seen[action] {
			t.Fatalf("audit action %s missing: events=%+v", action, audit.events)
		}
	}
}

func TestAuditEvents_NoRawText(t *testing.T) {
	audit, _ := exerciseObservedDebriefFlow(t, "__SECRET_RAW_TEXT__")
	allowedKeys := map[string]bool{
		"debrief_id":       true,
		"target_job_id":    true,
		"status":           true,
		"language":         true,
		"error_code":       true,
		"suggestion_count": true,
		"question_count":   true,
	}
	for _, event := range audit.events {
		for key := range event.Metadata {
			if !allowedKeys[key] {
				t.Fatalf("audit metadata key %q is not allowed: event=%+v", key, event)
			}
		}
		if strings.Contains(metadataString(event.Metadata), "__SECRET_RAW_TEXT__") {
			t.Fatalf("audit metadata leaked raw text: %+v", event.Metadata)
		}
	}
}

func exerciseObservedDebriefFlow(t *testing.T, secret string) (*recordingAudit, *recordingTaskRunWriter) {
	t.Helper()
	now := time.Date(2026, 5, 16, 18, 0, 0, 0, time.UTC)
	audit := &recordingAudit{}
	taskRuns := &recordingTaskRunWriter{}

	createService := NewService(ServiceOptions{
		Store: &recordingStore{result: CreateDebriefResult{
			DebriefID: "01918fa0-0000-7000-8000-00000000d010",
			Job: JobRecord{
				ID:           "01918fa0-0000-7000-8000-00000000d011",
				JobType:      api.JobTypeDebriefGenerate,
				ResourceType: api.ResourceTypeDebrief,
				ResourceID:   "01918fa0-0000-7000-8000-00000000d010",
				Status:       sharedtypes.JobStatusQueued,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}},
		Audit: audit,
		Now:   func() time.Time { return now },
		NewID: sequenceID("01918fa0-0000-7000-8000-00000000d010", "01918fa0-0000-7000-8000-00000000d011", "01918fa0-0000-7000-8000-00000000d012", "01918fa0-0000-7000-8000-00000000d013"),
	})
	if _, err := createService.CreateDebrief(context.Background(), CreateDebriefRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		RoundType:   sharedtypes.DebriefRoundTypeBehavioral,
		Language:    "zh-CN",
		Notes:       secret + " notes",
		Questions: []QuestionInput{{
			QuestionText:        secret + " question",
			MyAnswerSummary:     secret + " answer",
			InterviewerReaction: secret + " reaction",
		}},
	}); err != nil {
		t.Fatalf("CreateDebrief: %v", err)
	}

	suggestService := NewService(ServiceOptions{
		SuggestionContext: &recordingSuggestionContextStore{result: validSuggestionContext()},
		Registry:          &recordingPromptResolver{resolution: validSuggestionResolution()},
		AI: &recordingAIClient{
			response: aiclient.CompleteResponse{Content: `{"suggestions":[{"questionText":"How did you measure adoption?","whyLikelyAsked":"The JD stresses metrics.","source":"jd"}]}`},
			meta:     validSuggestionMeta(),
		},
		AITaskRuns: taskRuns,
		Audit:      audit,
		Now:        func() time.Time { return now },
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000e001", "01918fa0-0000-7000-8000-00000000e002"),
	})
	if _, err := suggestService.SuggestQuestions(context.Background(), validSuggestQuestionsRequest()); err != nil {
		t.Fatalf("SuggestQuestions: %v", err)
	}

	generateContext := validGenerateContext()
	generateContext.Questions[0].QuestionText = secret + " generated question"
	generateContext.Questions[0].MyAnswerSummary = secret + " generated answer"
	generateContext.Questions[0].InterviewerReaction = secret + " generated reaction"
	generateHandler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      &recordingGenerateStore{context: generateContext},
		Registry:   &recordingPromptResolver{resolution: validGenerateResolution()},
		AI:         &recordingAIClient{response: aiclient.CompleteResponse{Content: `{"questions":[{"questionText":"Tell me about the migration.","myAnswerSummary":"I led the rollout.","aiAnalysis":"Add metrics."}],"riskItems":[{"label":"Metrics missing","severity":"medium"}]}`}, meta: validGenerateMeta()},
		AITaskRuns: taskRuns,
		Audit:      audit,
		Now:        func() time.Time { return now },
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f001", "01918fa0-0000-7000-8000-00000000f002", "01918fa0-0000-7000-8000-00000000f003"),
	})
	if outcome := generateHandler.Handle(context.Background(), targetjob.ClaimedJob{
		JobID:        "01918fa0-0000-7000-8000-00000000d011",
		JobType:      "debrief_generate",
		ResourceType: "debrief",
		ResourceID:   generateContext.DebriefID,
		Payload:      []byte(`{"debriefId":"01918fa0-0000-7000-8000-00000000d010","targetJobId":"01918fa0-0000-7000-8000-00000000c001","language":"zh-CN","questionCount":1}`),
		Attempts:     1,
		MaxAttempts:  5,
	}); !outcome.Succeeded {
		t.Fatalf("GenerateHandler outcome: %+v", outcome)
	}

	return audit, taskRuns
}

type compileStore struct{}

func (s *compileStore) CreateDebrief(context.Context, CreateDebriefStoreInput) (CreateDebriefResult, error) {
	return CreateDebriefResult{}, nil
}

func (s *compileStore) GetDebrief(context.Context, string, string) (DebriefRecord, error) {
	return DebriefRecord{}, nil
}

func (s *compileStore) UpdateDebriefCompleted(context.Context, UpdateDebriefCompletedInput) (DebriefRecord, error) {
	return DebriefRecord{}, nil
}

type recordingStore struct {
	calls        int
	last         CreateDebriefStoreInput
	result       CreateDebriefResult
	err          error
	getCalls     int
	getUserID    string
	getDebriefID string
	getResult    DebriefRecord
	getErr       error
}

func (s *recordingStore) CreateDebrief(_ context.Context, in CreateDebriefStoreInput) (CreateDebriefResult, error) {
	s.calls++
	s.last = in
	if s.err != nil {
		return CreateDebriefResult{}, s.err
	}
	return s.result, nil
}

func (s *recordingStore) GetDebrief(_ context.Context, userID string, debriefID string) (DebriefRecord, error) {
	s.getCalls++
	s.getUserID = userID
	s.getDebriefID = debriefID
	if s.getErr != nil {
		return DebriefRecord{}, s.getErr
	}
	return s.getResult, nil
}

func (s *recordingStore) UpdateDebriefCompleted(context.Context, UpdateDebriefCompletedInput) (DebriefRecord, error) {
	return DebriefRecord{}, errors.New("not implemented")
}

type recordingAudit struct {
	events []DebriefAuditEvent
}

func (r *recordingAudit) RecordDebriefAuditEvent(_ context.Context, event DebriefAuditEvent) error {
	r.events = append(r.events, event)
	return nil
}

type recordingSuggestionContextStore struct {
	calls  int
	last   SuggestionContextRequest
	result SuggestionContext
	err    error
}

func (s *recordingSuggestionContextStore) GetSuggestionContext(_ context.Context, in SuggestionContextRequest) (SuggestionContext, error) {
	s.calls++
	s.last = in
	if s.err != nil {
		return SuggestionContext{}, s.err
	}
	return s.result, nil
}

type recordingPromptResolver struct {
	calls      int
	featureKey string
	language   string
	resolution registry.PromptResolution
	err        error
}

func (r *recordingPromptResolver) ResolveActive(_ context.Context, featureKey, language string) (registry.PromptResolution, error) {
	r.calls++
	r.featureKey = featureKey
	r.language = language
	if r.err != nil {
		return registry.PromptResolution{}, r.err
	}
	return r.resolution, nil
}

type recordingAIClient struct {
	calls       int
	profileName string
	payload     aiclient.CompletePayload
	response    aiclient.CompleteResponse
	meta        aiclient.AICallMeta
	err         error
}

func (c *recordingAIClient) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.calls++
	c.profileName = profileName
	c.payload = payload
	if c.err != nil {
		return aiclient.CompleteResponse{}, c.meta, c.err
	}
	return c.response, c.meta, nil
}

func (c *recordingAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

func (c *recordingAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}

func (c *recordingAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

type recordingTaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (w *recordingTaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

func sequenceID(ids ...string) func() string {
	next := 0
	return func() string {
		if next >= len(ids) {
			return "unexpected-extra-id"
		}
		id := ids[next]
		next++
		return id
	}
}

func metadataString(metadata map[string]any) string {
	var b strings.Builder
	for key, value := range metadata {
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(fmt.Sprint(value))
		b.WriteString(";")
	}
	return b.String()
}

func validSuggestQuestionsRequest() SuggestQuestionsRequest {
	return SuggestQuestionsRequest{
		UserID:      "01918fa0-0000-7000-8000-000000000001",
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		Language:    "zh-CN",
		Count:       2,
	}
}

func validSuggestionContext() SuggestionContext {
	return SuggestionContext{
		TargetJobID: "01918fa0-0000-7000-8000-00000000c001",
		Title:       "Staff Frontend Engineer",
		CompanyName: "Example Co",
		Summary:     "Design systems and cross-functional leadership.",
	}
}

func validSuggestionResolution() registry.PromptResolution {
	return registry.PromptResolution{
		FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.suggest_questions.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
		SystemMessage:       "You generate debrief questions.",
		OutputSchema:        testOutputSchema(`{"type":"object","required":["suggestions"],"properties":{"suggestions":{"type":"array"}}}`),
		UserMessageTemplate: "Target: {{targetTitle}}",
	}
}

func validSuggestionMeta() aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider:            "stub",
		ModelFamily:         "stub-family",
		ModelID:             "stub-model",
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.suggest_questions.default",
		ModelProfileVersion: "2026-05-16",
		FeatureKey:          featurekeys.DebriefSuggestQuestions.String(),
		FeatureFlag:         "none",
		DataSourceVersion:   "target_job/01918fa0-0000-7000-8000-00000000c001@v1",
		Language:            "zh-CN",
		ValidationStatus:    aiclient.ValidationStatusOK,
	}
}

func assertServiceErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != wantCode {
		t.Fatalf("error=%v, want ServiceError code %s", err, wantCode)
	}
}
