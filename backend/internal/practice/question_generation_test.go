package practice

import (
	"context"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestRenderQuestionTemplateBindsCanonicalMarkers(t *testing.T) {
	template := strings.Join([]string{
		"language={{language}}",
		"goal={{practice_goal}}",
		"mode={{practice_mode}}",
		"status={{turn_status}}",
		"target={{target_job_id}}",
		"kind={{generation_kind}}",
		"attempt={{attempt_mode}}",
		"question={{last_question}}",
		"intent={{question_intent}}",
		"answer={{last_answer}}",
		"count={{follow_up_count}}",
		"covered={{covered_dimensions}}",
		"remaining={{remaining_dimensions}}",
		"committed={{committed_context}}",
	}, "\n")

	got, err := renderQuestionTemplate(template, questionTemplateData{
		Language:            "zh-CN",
		PracticeGoal:        "baseline",
		PracticeMode:        "assisted",
		TurnStatus:          "follow_up_requested",
		TargetJobID:         "target-1",
		GenerationKind:      questionGenerationFollowUp,
		AttemptMode:         questionAttemptInitial,
		LastQuestion:        "请介绍一次系统拆分经历。",
		QuestionIntent:      "system_design",
		LastAnswer:          "我先建立基线，再按依赖关系拆分。",
		FollowUpCount:       1,
		CoveredDimensions:   []string{"system_design"},
		RemainingDimensions: []string{"evidence"},
		CommittedContext:    "我已说明回滚边界。",
	})
	if err != nil {
		t.Fatalf("renderQuestionTemplate returned error: %v", err)
	}
	for _, want := range []string{
		"language=zh-CN",
		"goal=baseline",
		"mode=assisted",
		"status=follow_up_requested",
		"target=target-1",
		"kind=follow_up",
		"attempt=initial",
		"question=请介绍一次系统拆分经历。",
		"intent=system_design",
		"answer=我先建立基线，再按依赖关系拆分。",
		"count=1",
		"covered=system_design",
		"remaining=evidence",
		"committed=我已说明回滚边界。",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "{{") {
		t.Fatalf("rendered prompt still contains an unresolved marker:\n%s", got)
	}
}

func TestRenderQuestionTemplateRejectsUnresolvedMarker(t *testing.T) {
	_, err := renderQuestionTemplate("known={{language}} unknown={{future_marker}}", questionTemplateData{Language: "en"})
	if err == nil {
		t.Fatal("expected unresolved marker to fail closed")
	}
}

func TestRenderQuestionTemplatePreservesMarkerLikeUserContent(t *testing.T) {
	const answer = "I used {{language}} and {{future_marker}} in the deployment template."

	got, err := renderQuestionTemplate(
		"language={{language}}\nanswer={{last_answer}}",
		questionTemplateData{Language: "zh-CN", LastAnswer: answer},
	)
	if err != nil {
		t.Fatalf("renderQuestionTemplate returned error for marker-like user content: %v", err)
	}
	if !strings.Contains(got, "language=zh-CN") || !strings.Contains(got, "answer="+answer) {
		t.Fatalf("rendered prompt changed marker-like user content: %q", got)
	}
}

func TestValidateGeneratedQuestionLanguage(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		language string
		wantErr  bool
	}{
		{name: "Chinese", text: "你如何为 RAG API 设计回滚方案？", language: "zh-CN"},
		{name: "Chinese with technical names", text: "你如何在 K8s 环境中定位 React 应用的延迟问题？", language: "zh-CN"},
		{name: "Chinese rejects English", text: "How did you measure the rollout impact?", language: "zh", wantErr: true},
		{name: "Chinese rejects English-dominant mixed text", text: "What tradeoff did 你 make when using React?", language: "zh-CN", wantErr: true},
		{name: "Chinese rejects nontechnical English padding", text: "请说明 API strategy 的取舍和影响？", language: "zh-CN", wantErr: true},
		{name: "English", text: "How did you measure the rollout impact?", language: "en-US"},
		{name: "English rejects Chinese", text: "你如何衡量上线影响？", language: "en", wantErr: true},
		{name: "Unsupported", text: "Quelle mesure avez-vous utilisée ?", language: "fr", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateGeneratedQuestionLanguage(tc.text, tc.language)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateGeneratedQuestionLanguage(%q, %q) error = %v, wantErr %v", tc.text, tc.language, err, tc.wantErr)
			}
		})
	}
}

func TestGenerateQuestionRepairsWrongLanguageExactlyOnce(t *testing.T) {
	ai := &fakeAIClient{
		contents: []string{
			firstQuestionJSON(t, "How did you measure the rollout impact?", "evidence"),
			firstQuestionJSON(t, "你如何衡量这次上线的实际影响？", "evidence"),
		},
		meta: aiclient.AICallMeta{ModelID: "chat-model-2"},
	}
	service := NewService(ServiceOptions{AI: ai})

	question, meta, err := service.generateQuestion(context.Background(), questionGenerationRequest{
		Resolution:   questionTestResolution(),
		TemplateData: questionTestTemplateData(),
		TaskRun: aiclient.AITaskRunContext{
			UserID:       "user-1",
			Capability:   aiclient.AITaskRunTaskFollowupGenerate,
			ResourceType: aiclient.AITaskRunResourceTargetJob,
			ResourceID:   "target-1",
		},
	})
	if err != nil {
		t.Fatalf("generateQuestion returned error: %v", err)
	}
	if question.Text != "你如何衡量这次上线的实际影响？" || question.Intent != "evidence" {
		t.Fatalf("unexpected repaired question: %+v", question)
	}
	if meta.ModelID != "chat-model-2" || len(ai.payloads) != 2 {
		t.Fatalf("repair call count/meta mismatch: calls=%d meta=%+v", len(ai.payloads), meta)
	}
	initial := ai.payloads[0].Messages[len(ai.payloads[0].Messages)-1].Content
	repair := ai.payloads[1].Messages[len(ai.payloads[1].Messages)-1].Content
	if !strings.Contains(initial, "attempt=initial") || !strings.Contains(repair, "attempt=repair") {
		t.Fatalf("attempt markers missing: initial=%q repair=%q", initial, repair)
	}
	for _, stable := range []string{"language=zh-CN", "question=请介绍一次系统拆分经历。", "answer=我先建立基线。"} {
		if !strings.Contains(initial, stable) || !strings.Contains(repair, stable) {
			t.Fatalf("repair changed canonical context %q: initial=%q repair=%q", stable, initial, repair)
		}
	}
}

func TestGenerateQuestionProviderTimeoutDoesNotRepair(t *testing.T) {
	ai := &fakeAIClient{err: sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "timeout", true)}
	service := NewService(ServiceOptions{AI: ai})

	_, _, err := service.generateQuestion(context.Background(), questionGenerationRequest{
		Resolution:   questionTestResolution(),
		TemplateData: questionTestTemplateData(),
	})
	if code, ok := aiErrorCode(err); !ok || code != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("expected AI_PROVIDER_TIMEOUT, got %T %v", err, err)
	}
	if len(ai.payloads) != 1 {
		t.Fatalf("provider timeout must not repair, calls=%d", len(ai.payloads))
	}
}

func TestGenerateQuestionSecondLanguageMismatchReturnsAIOutputInvalid(t *testing.T) {
	ai := &fakeAIClient{contents: []string{
		firstQuestionJSON(t, "How did you measure the rollout impact?", "evidence"),
		firstQuestionJSON(t, "What tradeoff did you make?", "tradeoff"),
	}}
	service := NewService(ServiceOptions{AI: ai})

	_, _, err := service.generateQuestion(context.Background(), questionGenerationRequest{
		Resolution:   questionTestResolution(),
		TemplateData: questionTestTemplateData(),
	})
	if code, ok := aiErrorCode(err); !ok || code != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("expected AI_OUTPUT_INVALID, got %T %v", err, err)
	}
	if len(ai.payloads) != 2 {
		t.Fatalf("invalid output must repair exactly once, calls=%d", len(ai.payloads))
	}
}

func questionTestResolution() registry.PromptResolution {
	return registry.PromptResolution{
		FeatureKey:          followUpFeatureKey,
		PromptVersion:       "followup.prompt.v1",
		RubricVersion:       "followup.rubric.v1",
		ModelProfileName:    "practice.followup.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		SystemMessage:       "Return one structured question.",
		OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
		UserMessageTemplate: "language={{language}} kind={{generation_kind}} attempt={{attempt_mode}} question={{last_question}} answer={{last_answer}} intent={{question_intent}} goal={{practice_goal}} mode={{practice_mode}} status={{turn_status}} target={{target_job_id}} count={{follow_up_count}} covered={{covered_dimensions}} remaining={{remaining_dimensions}} committed={{committed_context}}",
	}
}

func questionTestTemplateData() questionTemplateData {
	return questionTemplateData{
		Language:          "zh-CN",
		PracticeGoal:      "baseline",
		PracticeMode:      "assisted",
		TurnStatus:        "follow_up_requested",
		TargetJobID:       "target-1",
		GenerationKind:    questionGenerationFollowUp,
		LastQuestion:      "请介绍一次系统拆分经历。",
		QuestionIntent:    "system_design",
		LastAnswer:        "我先建立基线。",
		FollowUpCount:     1,
		CoveredDimensions: []string{"system_design"},
	}
}
