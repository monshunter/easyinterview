package registry

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

func TestReportGenerateConversationContractPreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	resolution, err := client.ResolveActive(context.Background(), string(featurekeys.ReportGenerate), "en")
	if err != nil {
		t.Fatalf("ResolveActive: %v", err)
	}
	if resolution.ModelProfileName != "report.generate.default" || resolution.OutputSchema == nil {
		t.Fatalf("resolution = %+v", resolution)
	}
	if resolution.PromptVersion != "v0.2.0" || resolution.RubricVersion != "v0.2.0" || resolution.DataSourceVersion != "report-context.v1" {
		t.Fatalf("grounded resolution = %+v", resolution)
	}
	lower := strings.ToLower(resolution.UserMessageTemplate)
	for _, required := range []string{"conversation_messages", "dimensionassessments", "highlights", "issues", "nextactions", "retryfocusdimensioncodes"} {
		if !strings.Contains(lower, required) {
			t.Fatalf("report prompt missing %q", required)
		}
	}
	for _, stale := range []string{"question_assessment", "retry_focus_turn_ids", "turn_summaries", "dimension_scores", "retry_focus_competency_codes", "{{rubric_dimensions}}"} {
		if strings.Contains(lower, stale) {
			t.Fatalf("report prompt contains stale %q", stale)
		}
	}
}

func TestReportGenerateGroundedCandidateContractPreflight(t *testing.T) {
	prompts, rubrics := testsupport.ConfigRoots(t)
	client, err := NewRegistryClient(RegistryOptions{PromptsDir: prompts, RubricsDir: rubrics})
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	meta, body, err := client.GetPrompt(string(featurekeys.ReportGenerate), "v0.2.0", "multi")
	if err != nil {
		t.Fatalf("GetPrompt v0.2.0: %v", err)
	}
	if meta.Status != "active" || meta.OutputSchema == nil {
		t.Fatalf("candidate meta = %+v", meta)
	}
	for _, token := range []string{"{{frozen_context}}", "{{conversation_messages}}", "{{language}}", "<untrusted_report_context_json>"} {
		if !strings.Contains(body, token) {
			t.Fatalf("grounded prompt missing %q", token)
		}
	}
	normalizedBody := strings.Join(strings.Fields(strings.ToLower(body)), " ")
	for _, policy := range []string{
		"evidence being partial, rehearsed, or merely not covered is not itself a deficiency and must not lower preparedness or justify a corrective action",
		"a brief assertion that only names a mechanism without concrete supporting detail",
		"not as evidence of a topic-specific capability gap",
		"use the exact dimension code `answer_depth`",
		"may state only that the answer provides no concrete supporting detail",
		"must not enumerate unmentioned expected details",
		"do not use the assistant question or its topic to create a more specific issue or retry focus",
		"emit `retry_current_round` with an empty focus array and a generic label",
		"the generic label must not repeat the assistant question or name its topic or mechanism",
		"a corrective `retry_current_round` label may only turn the cited missing behavior",
		"`review_evidence` must only ask the user to revisit cited positive or explicitly evidence-limited content",
		"do not invent an artifact, corrective gap, new scenario, or transfer task",
		"for every selected focus code, the first retry label must name at least one directly cited missing behavior",
		"umbrella labels such as `add a backpressure mechanism`, `add a safety check`, `add detail`, or `improve the answer` are invalid",
		"the schema's 200-character bound is only the outer malformed-output safety cap",
		"one semicolon-separated cited missing behavior per selected focus code",
		"in `en`, each label has 1-24 whitespace-delimited words",
		"in `zh-cn`, each label has 1-64 unicode code points",
		"do not add an introduction such as `retry the answer by adding`",
		"retry the prioritization answer by explaining the tie-breaking rule",
		"empty focus is allowed only for exactly one `answer_depth` issue",
		"or exactly one `answer_relevance` issue",
		"must equal the ascending unique dimension codes of all issues whose declared dimension status is `needs_work`",
		"before returning json, set `i = len(issues)`",
		"if `i >= 2`, empty focus is invalid",
		"treat an explicitly stated unsafe current-round approach as blocking",
		"a `retry_current_round` label must name the concrete missing control, check, or answer detail to add",
		"use each type at most once",
		"it never emits a second retry action",
		"it must not prescribe a new mechanism, threshold, tool, sequence, framework, or example absent from the cited candidate messages",
	} {
		if !strings.Contains(normalizedBody, policy) {
			t.Fatalf("grounded prompt missing reliability policy %q", policy)
		}
	}
	for _, ambiguous := range []string{"when retry is present it may be empty", "use non-empty focus only when"} {
		if strings.Contains(normalizedBody, ambiguous) {
			t.Fatalf("grounded prompt contains ambiguous focus policy %q", ambiguous)
		}
	}
	for _, unsupported := range []string{"an action label may only turn the cited missing behavior", "with the stated tie-breaking rule"} {
		if strings.Contains(normalizedBody, unsupported) {
			t.Fatalf("grounded prompt contains unsupported advice policy %q", unsupported)
		}
	}
	for _, stale := range []string{"{{rubric_dimensions}}", "dimension_scores", "retry_focus_competency_codes", "supporting_observations"} {
		if strings.Contains(body, stale) {
			t.Fatalf("grounded prompt contains stale %q", stale)
		}
	}

	var schema struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(*meta.OutputSchema, &schema); err != nil {
		t.Fatalf("parse v0.2 schema: %v", err)
	}
	want := []string{"summary", "preparednessLevel", "dimensionAssessments", "highlights", "issues", "nextActions", "retryFocusDimensionCodes"}
	if strings.Join(schema.Required, ",") != strings.Join(want, ",") {
		t.Fatalf("required keys = %v, want %v", schema.Required, want)
	}
	if len(schema.Properties) != len(want) {
		t.Fatalf("top-level schema must be closed by exact parser contract: got %v", schema.Properties)
	}
}
