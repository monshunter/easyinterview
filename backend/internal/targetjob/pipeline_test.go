package targetjob_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// ----- pipeline test fakes -----

type pipelineFakeStore struct {
	importIn *targetjob.ImportTargetJobInput

	// parse executor side
	target              targetjob.TargetJobRecord
	getErr              error
	applyResultIn       *targetjob.ApplyParseResultInput
	applyResultErr      error
	completeSuccessIn   *targetjob.CompleteParseSuccessInput
	completeSuccessErr  error
	completeFailureIn   *targetjob.CompleteParseFailureInput
	completeFailureErr  error
	parsedOutboxPayload []byte
	failedOutboxPayload []byte
}

func (s *pipelineFakeStore) ImportTargetJob(_ context.Context, in targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
	cp := in
	s.importIn = &cp
	return targetjob.ImportTargetJobResult{}, nil
}
func (s *pipelineFakeStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	return nil
}
func (s *pipelineFakeStore) GetTargetJobByUser(context.Context, string, string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, error) {
	return targetjob.TargetJobRecord{}, nil, nil
}
func (s *pipelineFakeStore) ListTargetJobsForUser(context.Context, string, targetjob.ListFilter) (targetjob.ListResult, error) {
	return targetjob.ListResult{}, nil
}
func (s *pipelineFakeStore) LookupUpdateDedupe(context.Context, string, string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, bool, error) {
	return targetjob.TargetJobRecord{}, nil, false, nil
}
func (s *pipelineFakeStore) UpdateTargetJobLifecycle(context.Context, string, string, targetjob.UpdateLifecycleFields, time.Time) (targetjob.TargetJobRecord, error) {
	return targetjob.TargetJobRecord{}, nil
}
func (s *pipelineFakeStore) ArchiveTargetJob(context.Context, targetjob.ArchiveTargetJobInput) (targetjob.TargetJobRecord, error) {
	return targetjob.TargetJobRecord{}, nil
}
func (s *pipelineFakeStore) ApplyParseResult(_ context.Context, in targetjob.ApplyParseResultInput) error {
	cp := in
	s.applyResultIn = &cp
	return s.applyResultErr
}
func (s *pipelineFakeStore) CompleteParseSuccess(_ context.Context, in targetjob.CompleteParseSuccessInput) error {
	cp := in
	s.completeSuccessIn = &cp
	if len(in.ParsedEventPayload) > 0 {
		s.parsedOutboxPayload = append([]byte{}, in.ParsedEventPayload...)
	}
	return s.completeSuccessErr
}
func (s *pipelineFakeStore) CompleteParseFailure(_ context.Context, in targetjob.CompleteParseFailureInput) error {
	cp := in
	s.completeFailureIn = &cp
	if len(in.FailedEventPayload) > 0 {
		s.failedOutboxPayload = append([]byte{}, in.FailedEventPayload...)
	}
	return s.completeFailureErr
}
func (s *pipelineFakeStore) WriteParseFailedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.failedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) WriteTargetParsedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.parsedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) GetTargetJobForParse(context.Context, string) (targetjob.TargetJobRecord, error) {
	if s.getErr != nil {
		return targetjob.TargetJobRecord{}, s.getErr
	}
	return s.target, nil
}

type fakeRegistry struct {
	resolution targetjob.PromptResolution
	err        error
}

func (f *fakeRegistry) Resolve(_ context.Context, _ string, _ string) (targetjob.PromptResolution, error) {
	if f.err != nil {
		return targetjob.PromptResolution{}, f.err
	}
	return f.resolution, nil
}

type fakeAIClient struct {
	resp            aiclient.CompleteResponse
	meta            aiclient.AICallMeta
	err             error
	echoMetadata    bool
	lastProfileName string
	lastPayload     aiclient.CompletePayload
}

func (f *fakeAIClient) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.lastProfileName = profileName
	f.lastPayload = payload
	if f.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, f.err
	}
	meta := f.meta
	if f.echoMetadata {
		meta.ModelProfileName = profileName
		meta.PromptVersion = payload.Metadata.PromptVersion
		meta.RubricVersion = payload.Metadata.RubricVersion
		meta.Language = payload.Metadata.Language
		meta.FeatureKey = payload.Metadata.FeatureKey
		meta.FeatureFlag = payload.Metadata.FeatureFlag
		meta.DataSourceVersion = payload.Metadata.DataSourceVersion
	}
	return f.resp, meta, nil
}
func (f *fakeAIClient) Transcribe(context.Context, string, aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Transcribe call in targetjob fakeAIClient")
}
func (f *fakeAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("unexpected Stream call in targetjob fakeAIClient")
}
func (f *fakeAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("unexpected Synthesize call in targetjob fakeAIClient")
}

// idSeq returns a deterministic IDGenerator for a fixed list. After the
// list is exhausted, it returns indexed strings.
func idSeq(prefix string) targetjob.IDGenerator {
	var n int32
	return func() string {
		next := atomic.AddInt32(&n, 1)
		return fmt.Sprintf("%s-%d", prefix, next)
	}
}

func newParseExecutorWithFakes(t *testing.T) (*targetjob.ParseExecutor, *pipelineFakeStore, *fakeRegistry, *fakeAIClient) {
	t.Helper()
	store := &pipelineFakeStore{}
	registry := &fakeRegistry{
		resolution: targetjob.PromptResolution{
			PromptVersion:     "v1.0.0",
			RubricVersion:     "v1.0.0",
			ModelProfileName:  "target.import.default",
			DataSourceVersion: "v1",
			FeatureFlag:       "rollout-f3-target-import",
		},
	}
	ai := &fakeAIClient{
		meta: aiclient.AICallMeta{
			Provider:            "unit-test-provider",
			ModelFamily:         "fixture",
			ModelID:             "fixture-model:target-import-parse",
			ModelProfileVersion: "1.1.0",
			Capability:          aiclient.CapabilityChat,
			ValidationStatus:    aiclient.ValidationStatusOK,
		},
		echoMetadata: true,
	}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       ai,
		NewID:    idSeq("id"),
		Now:      func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})
	return exec, store, registry, ai
}

const happyResponseJSON = `{
  "title": "Senior Backend Engineer",
  "companyName": "Acme",
  "coreThemes": ["api"],
  "interviewRounds": [
    {
      "sequence": 1,
      "type": "technical",
      "name": "Backend architecture deep dive",
      "durationMinutes": 45,
      "focus": "Probe API design and async pipelines."
    },
    {
      "sequence": 2,
      "type": "manager",
      "name": "Hiring manager ownership interview",
      "durationMinutes": 40,
      "focus": "Assess ownership, incident handling, and collaboration."
    }
  ],
  "strengths": ["Go"],
  "gaps": ["k8s"],
  "riskSignals": ["The JD implies on-call ownership without naming support coverage."],
  "requirements": [
    {"kind":"must_have","label":"Go","evidenceLevel":"explicit"},
    {"kind":"hidden_signal","label":"On-call ownership ambiguity","evidenceLevel":"inferred"},
    {"kind":"interview_focus","label":"system design"}
  ]
}`

func TestParseExecutor_HappyPath(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})
	if !outcome.Succeeded {
		t.Fatalf("happy path must succeed, got %+v", outcome)
	}
	if store.completeSuccessIn == nil || len(store.completeSuccessIn.Requirements) != 3 {
		t.Fatalf("expected 3 requirements applied atomically, got %+v", store.completeSuccessIn)
	}
	if store.completeSuccessIn.Title != "Senior Backend Engineer" || store.completeSuccessIn.CompanyName != "Acme" {
		t.Fatalf("parse success must persist title/company from AI output, got title=%q company=%q", store.completeSuccessIn.Title, store.completeSuccessIn.CompanyName)
	}
	if store.applyResultIn != nil {
		t.Fatalf("parse executor must not apply results before atomic outbox write: %+v", store.applyResultIn)
	}
	if ai.lastProfileName != "target.import.default" {
		t.Fatalf("profileName = %q", ai.lastProfileName)
	}
	if got := ai.lastPayload.Metadata; got.FeatureKey != "target.import.parse" ||
		got.PromptVersion != "v1.0.0" ||
		got.RubricVersion != "v1.0.0" ||
		got.Language != "en" ||
		got.FeatureFlag != "rollout-f3-target-import" ||
		got.DataSourceVersion != "v1" {
		t.Fatalf("AI metadata did not carry F3 resolution: %+v", got)
	}
	if got := ai.lastPayload.Metadata.TaskRun; got.Capability != aiclient.AITaskRunTaskJDParse ||
		got.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		got.ResourceID != "tgt-1" {
		t.Fatalf("AI task run context did not carry B4 targetjob scope: %+v", got)
	}
	summaryProv := decodeProvenanceForTest(t, store.completeSuccessIn.Summary)
	if summaryProv["modelId"] != "fixture-model:target-import-parse" {
		t.Fatalf("summary provenance modelId must use A3 resolved model id, got %+v", summaryProv)
	}
	if summaryProv["featureFlag"] != "rollout-f3-target-import" {
		t.Fatalf("summary provenance featureFlag drift: %+v", summaryProv)
	}
	summary := decodeSummaryForTest(t, store.completeSuccessIn.Summary)
	if _, ok := summary["interviewHypotheses"]; ok {
		t.Fatalf("summary must not persist out-of-scope interviewHypotheses: %s", string(store.completeSuccessIn.Summary))
	}
	rounds, ok := summary["interviewRounds"].([]any)
	if !ok || len(rounds) != 2 {
		t.Fatalf("summary must persist structured interviewRounds, got %s", string(store.completeSuccessIn.Summary))
	}
	firstRound, ok := rounds[0].(map[string]any)
	if !ok ||
		firstRound["sequence"] != float64(1) ||
		firstRound["type"] != "technical" ||
		firstRound["name"] != "Backend architecture deep dive" ||
		firstRound["durationMinutes"] != float64(45) ||
		firstRound["focus"] != "Probe API design and async pipelines." {
		t.Fatalf("first persisted interview round drift: %+v", firstRound)
	}
	fitProv := decodeProvenanceForTest(t, store.completeSuccessIn.FitSummary)
	if fitProv["modelId"] != "fixture-model:target-import-parse" {
		t.Fatalf("fitSummary provenance modelId must use A3 resolved model id, got %+v", fitProv)
	}
	if store.parsedOutboxPayload == nil {
		t.Fatal("target.parsed outbox payload missing")
	}
}

func TestParseExecutor_HappyPathAcceptsFencedJSON(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: "```json\n" + happyResponseJSON + "\n```"}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "zh-CN",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if !outcome.Succeeded {
		t.Fatalf("fenced JSON should be normalized and parsed, got %+v", outcome)
	}
	if store.completeSuccessIn == nil || store.completeSuccessIn.Title != "Senior Backend Engineer" ||
		store.completeSuccessIn.CompanyName != "Acme" || len(store.completeSuccessIn.Requirements) != 3 {
		t.Fatalf("fenced JSON parse result was not committed: %+v", store.completeSuccessIn)
	}
}

func TestParseExecutor_BackfillsHiddenSignalRequirementsFromRiskSignals(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: `{
  "title": "AI Application Engineer",
  "companyName": "Acme Logistics",
  "coreThemes": ["AI agents", "RAG"],
  "interviewRounds": [
    {"sequence":1,"type":"technical","name":"Technical foundation","durationMinutes":60,"focus":"Probe Go/Python and AI agent implementation."},
    {"sequence":2,"type":"manager","name":"Business landing interview","durationMinutes":45,"focus":"Assess logistics scenario delivery and cross-team collaboration."}
  ],
  "strengths": ["AI agent delivery"],
  "gaps": ["Logistics domain depth"],
  "riskSignals": ["The JD emphasizes logistics landing, so domain implementation evidence may be an unstated screen."],
  "requirements": [
    {"kind":"must_have","label":"AI agent implementation","evidenceLevel":"explicit"}
  ]
}`}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if !outcome.Succeeded {
		t.Fatalf("riskSignals-derived hidden signal path must succeed, got %+v", outcome)
	}
	var foundHidden bool
	for _, req := range store.completeSuccessIn.Requirements {
		if req.Kind == targetjob.RequirementHiddenSignal {
			foundHidden = true
			if req.Label != "The JD emphasizes logistics landing, so domain implementation evidence may be an unstated screen." ||
				req.EvidenceLevel != targetjob.EvidenceInferred {
				t.Fatalf("derived hidden_signal drift: %+v", req)
			}
		}
	}
	if !foundHidden {
		t.Fatalf("expected hidden_signal derived from riskSignals, got %+v", store.completeSuccessIn.Requirements)
	}
}

func TestParseExecutor_AIOutputInvalid_WhenHiddenSignalsAreMissing(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: `{
  "title": "Backend Engineer",
  "companyName": "Acme",
  "coreThemes": ["api"],
  "interviewRounds": [
    {"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"Probe APIs."},
    {"sequence":2,"type":"manager","name":"Manager","durationMinutes":40,"focus":"Assess ownership."}
  ],
  "strengths": ["Go"],
  "gaps": ["k8s"],
  "riskSignals": [],
  "requirements": [
    {"kind":"must_have","label":"Go","evidenceLevel":"explicit"}
  ]
}`}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})

	if outcome.Succeeded || outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
		t.Fatalf("missing hidden signal must fail non-retryable AI_OUTPUT_INVALID, got %+v", outcome)
	}
	if store.completeSuccessIn != nil {
		t.Fatalf("missing hidden signal must not mark target ready: %+v", store.completeSuccessIn)
	}
}

func TestParseExecutor_HappyPathCoalescesMissingCompanyName(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: strings.Replace(happyResponseJSON, `"companyName": "Acme"`, `"companyName": ""`, 1)}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "zh-CN",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if !outcome.Succeeded {
		t.Fatalf("missing companyName should not fail a JD without company evidence, got %+v", outcome)
	}
	if store.completeSuccessIn == nil || store.completeSuccessIn.CompanyName != "未提供" {
		t.Fatalf("missing companyName fallback drift: %+v", store.completeSuccessIn)
	}
}

func TestParseExecutor_UsesPromptMessagesFromRegistryResolution(t *testing.T) {
	exec, store, registry, ai := newParseExecutorWithFakes(t)
	registry.resolution.SystemMessage = "registry-owned target import system prompt"
	registry.resolution.UserMessageTemplate = "language={{language}}\ntext={{jd_text}}"
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD text from caller",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})

	if !outcome.Succeeded {
		t.Fatalf("parse should succeed, got %+v", outcome)
	}
	if len(ai.lastPayload.Messages) != 2 {
		t.Fatalf("messages = %+v", ai.lastPayload.Messages)
	}
	if ai.lastPayload.Messages[0].Role != "system" || ai.lastPayload.Messages[0].Content != "registry-owned target import system prompt" {
		t.Fatalf("system message not sourced from registry resolution: %+v", ai.lastPayload.Messages)
	}
	if ai.lastPayload.Messages[1].Role != "user" || ai.lastPayload.Messages[1].Content != "language=en\ntext=JD text from caller" {
		t.Fatalf("user message not sourced from registry template: %+v", ai.lastPayload.Messages)
	}
	for _, msg := range ai.lastPayload.Messages {
		if strings.Contains(msg.Content, "extraction assistant") {
			t.Fatalf("ParseExecutor must not carry hardcoded prompt text: %+v", ai.lastPayload.Messages)
		}
	}
}

func TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name: "all invalid kind",
			content: `{"requirements":[
				{"kind":"prior_signal","label":"Prior signal","evidenceLevel":"explicit"}
			]}`,
		},
		{
			name: "empty label",
			content: `{"requirements":[
				{"kind":"must_have","label":"   ","evidenceLevel":"explicit"}
			]}`,
		},
		{
			name: "invalid evidence level",
			content: `{"requirements":[
				{"kind":"must_have","label":"Go","evidenceLevel":"rumor"}
			]}`,
		},
		{
			name: "mixed valid and invalid",
			content: `{"requirements":[
				{"kind":"must_have","label":"Go","evidenceLevel":"explicit"},
				{"kind":"prior_signal","label":"Prior signal","evidenceLevel":"explicit"}
			]}`,
		},
		{
			name:    "valid json followed by prose",
			content: happyResponseJSON + "\nDone.",
		},
		{
			name:    "fenced json with leading prose",
			content: "Here is the JSON:\n```json\n" + happyResponseJSON + "\n```",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, store, _, ai := newParseExecutorWithFakes(t)
			ai.resp = aiclient.CompleteResponse{Content: tc.content}
			store.target = targetjob.TargetJobRecord{
				ID:             "tgt-1",
				UserID:         "user-1",
				TargetLanguage: "en",
				RawJDText:      "JD text",
			}

			outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})

			if outcome.Succeeded || outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
				t.Fatalf("invalid semantic requirements must fail non-retryable AI_OUTPUT_INVALID, got %+v", outcome)
			}
			if store.completeSuccessIn != nil {
				t.Fatalf("invalid AI output must not mark target ready: %+v", store.completeSuccessIn)
			}
			if store.completeFailureIn == nil {
				t.Fatal("invalid AI output must write target.analysis.failed")
			}
		})
	}
}

func TestParseExecutor_AIOutputInvalid_WhenInterviewRoundsAreSemanticallyInvalid(t *testing.T) {
	cases := []struct {
		name  string
		round string
	}{
		{
			name:  "at least two rounds required",
			round: `{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"Probe APIs"}`,
		},
		{
			name: "at most five rounds allowed",
			round: `{"sequence":1,"type":"hr","name":"Recruiter","durationMinutes":30,"focus":"Motivation"},
				{"sequence":2,"type":"technical","name":"Technical 1","durationMinutes":45,"focus":"APIs"},
				{"sequence":3,"type":"technical","name":"Technical 2","durationMinutes":60,"focus":"Architecture"},
				{"sequence":4,"type":"manager","name":"Manager","durationMinutes":45,"focus":"Ownership"},
				{"sequence":5,"type":"culture","name":"Culture","durationMinutes":30,"focus":"Collaboration"},
				{"sequence":6,"type":"final","name":"Final","durationMinutes":30,"focus":"Calibration"}`,
		},
		{
			name:  "sequence must be one based",
			round: `{"sequence":0,"type":"technical","name":"Technical","durationMinutes":45,"focus":"Probe APIs"}`,
		},
		{
			name: "sequence must fit the int32 API contract",
			round: `{"sequence":2147483648,"type":"technical","name":"Technical","durationMinutes":45,"focus":"Probe APIs"},
				{"sequence":2147483649,"type":"manager","name":"Manager","durationMinutes":45,"focus":"Ownership"}`,
		},
		{
			name:  "type must be known enum",
			round: `{"sequence":1,"type":"screen","name":"Screen","durationMinutes":30,"focus":"Probe APIs"}`,
		},
		{
			name:  "name must be present",
			round: `{"sequence":1,"type":"technical","name":"  ","durationMinutes":45,"focus":"Probe APIs"}`,
		},
		{
			name:  "duration must be positive",
			round: `{"sequence":1,"type":"technical","name":"Technical","durationMinutes":0,"focus":"Probe APIs"}`,
		},
		{
			name:  "focus must be present",
			round: `{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"  "}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exec, store, _, ai := newParseExecutorWithFakes(t)
			ai.resp = aiclient.CompleteResponse{Content: targetImportResponseWithRounds(tc.round)}
			store.target = targetjob.TargetJobRecord{
				ID:             "tgt-1",
				UserID:         "user-1",
				TargetLanguage: "en",
				RawJDText:      "JD text",
			}

			outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})

			if outcome.Succeeded || outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
				t.Fatalf("invalid interview rounds must fail non-retryable AI_OUTPUT_INVALID, got %+v", outcome)
			}
			if store.completeSuccessIn != nil {
				t.Fatalf("invalid interview rounds must not mark target ready: %+v", store.completeSuccessIn)
			}
		})
	}
}

func TestParseExecutorAITaskRuns(t *testing.T) {
	_, store, registry, ai := newParseExecutorWithFakes(t)
	store.target = targetjob.TargetJobRecord{
		ID:             "01918fa0-0000-7000-8000-000000002000",
		UserID:         "01918fa0-0000-7000-8000-000000001000",
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}
	registry.resolution = targetjob.PromptResolution{
		PromptVersion:     "v0.1.0",
		RubricVersion:     "v0.1.0",
		ModelProfileName:  "target.import.default",
		DataSourceVersion: "registry.v1",
		FeatureFlag:       "rollout-f3-target-import",
	}
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	ai.meta.ModelID = "fixture-model:target-import-parse"
	runWriter := &targetjobMemTaskRunWriter{}
	auditWriter := &targetjobMemAuditWriter{}
	wrappedAI, err := observability.New(ai,
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(observability.NewMemoryLogger()),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(auditWriter),
		observability.WithProfileResolver(targetjobStaticResolver{
			"target.import.default": {
				Name:       "target.import.default",
				Capability: aiclient.CapabilityChat,
				Status:     aiclient.ProfileStatusActive,
				Default: aiclient.ProviderConfig{
					ProviderRef: "unit-test-provider",
					Model:       "fixture-model:target-import-parse",
				},
				TimeoutMs: 15000,
				Route:     "target.import",
				Version:   "1.1.0",
			},
		}),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       wrappedAI,
		NewID:    idSeq("id"),
		Now:      func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: store.target.ID,
	})

	if !outcome.Succeeded {
		t.Fatalf("parse should succeed with observability wrapper, got %+v", outcome)
	}
	rows := runWriter.Rows()
	if len(rows) != 1 {
		t.Fatalf("expected one ai_task_runs row, got %+v", rows)
	}
	row := rows[0]
	if row.Capability != aiclient.AITaskRunTaskJDParse ||
		row.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		row.ResourceID != store.target.ID {
		t.Fatalf("ai_task_runs task context drift: %+v", row)
	}
	if row.FeatureKey != targetjob.FeatureKeyTargetImportParse ||
		row.PromptVersion != "v0.1.0" ||
		row.RubricVersion != "v0.1.0" ||
		row.ModelProfileName != "target.import.default" ||
		row.ModelID != "fixture-model:target-import-parse" ||
		row.DataSourceVersion != "registry.v1" ||
		row.FeatureFlag != "rollout-f3-target-import" {
		t.Fatalf("ai_task_runs provenance drift: %+v", row)
	}
}

func decodeProvenanceForTest(t *testing.T, raw json.RawMessage) map[string]string {
	t.Helper()
	var envelope struct {
		Provenance map[string]string `json:"provenance"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode provenance: %v", err)
	}
	return envelope.Provenance
}

func decodeSummaryForTest(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	return out
}

func targetImportResponseWithRounds(rounds string) string {
	return `{
  "title": "Senior Backend Engineer",
  "companyName": "Acme",
  "coreThemes": ["api"],
  "interviewRounds": [` + rounds + `],
  "strengths": ["Go"],
  "gaps": ["k8s"],
  "riskSignals": [],
  "requirements": [
    {"kind":"must_have","label":"Go","evidenceLevel":"explicit"}
  ]
}`
}

type targetjobMemTaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (m *targetjobMemTaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	m.rows = append(m.rows, row)
	return nil
}

func (m *targetjobMemTaskRunWriter) Rows() []aiclient.AITaskRunRow {
	return append([]aiclient.AITaskRunRow{}, m.rows...)
}

type targetjobMemAuditWriter struct {
	rows []aiclient.AuditEventRow
}

func (m *targetjobMemAuditWriter) WriteAuditEvent(_ context.Context, row aiclient.AuditEventRow) error {
	m.rows = append(m.rows, row)
	return nil
}

type targetjobStaticResolver map[string]*aiclient.ModelProfile

func (r targetjobStaticResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	profile, ok := r[name]
	if !ok {
		return nil, errors.New("not found: " + name)
	}
	return profile, nil
}

func TestDeterministicParseAIClient_OnlyInterceptsTargetImportParse(t *testing.T) {
	inner := &fakeAIClient{resp: aiclient.CompleteResponse{Content: "delegated"}}
	client := targetjob.NewDeterministicParseAIClient(inner)

	resp, _, err := client.Complete(context.Background(), "target.import.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "JD text"}},
		Metadata: aiclient.CallMetadata{
			FeatureKey: targetjob.FeatureKeyTargetImportParse,
			Language:   "en",
		},
	})
	if err != nil {
		t.Fatalf("Complete target.import.parse: %v", err)
	}
	var parsed struct {
		Requirements []struct {
			Kind  string `json:"kind"`
			Label string `json:"label"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &parsed); err != nil {
		t.Fatalf("fixture did not return JSON parse content: %v; content=%s", err, resp.Content)
	}
	if len(parsed.Requirements) == 0 || parsed.Requirements[0].Kind != string(targetjob.RequirementMustHave) || parsed.Requirements[0].Label == "" {
		t.Fatalf("fixture returned invalid requirements: %+v", parsed.Requirements)
	}

	store := &pipelineFakeStore{
		target: targetjob.TargetJobRecord{
			ID:             "tgt-1",
			UserID:         "user-1",
			TargetLanguage: "en",
			RawJDText:      "JD text",
		},
	}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store: store,
		Registry: &fakeRegistry{
			resolution: targetjob.PromptResolution{
				PromptVersion:     "v1.0.0",
				RubricVersion:     "v1.0.0",
				ModelProfileName:  "target.import.default",
				DataSourceVersion: "v1",
			},
		},
		AI:    client,
		NewID: idSeq("fixture-req"),
		Now:   func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})
	if !outcome.Succeeded {
		t.Fatalf("deterministic parse fixture must satisfy current ParseExecutor contract, got %+v", outcome)
	}
	var foundHidden bool
	for _, req := range store.completeSuccessIn.Requirements {
		if req.Kind == targetjob.RequirementHiddenSignal {
			foundHidden = true
		}
	}
	if !foundHidden {
		t.Fatalf("deterministic parse fixture must include hidden_signal coverage: %+v", store.completeSuccessIn.Requirements)
	}
	summary := decodeSummaryForTest(t, store.completeSuccessIn.Summary)
	rounds, ok := summary["interviewRounds"].([]any)
	if !ok || len(rounds) < 2 || len(rounds) > 5 {
		t.Fatalf("deterministic parse fixture must persist 2 to 5 structured rounds, got %s", string(store.completeSuccessIn.Summary))
	}

	delegated, _, err := client.Complete(context.Background(), "practice.chat.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "other"}},
		Metadata: aiclient.CallMetadata{FeatureKey: "practice.session.chat"},
	})
	if err != nil {
		t.Fatalf("delegated Complete: %v", err)
	}
	if delegated.Content != "delegated" || inner.lastProfileName != "practice.chat.default" {
		t.Fatalf("non target import calls must delegate, got resp=%+v profile=%q", delegated, inner.lastProfileName)
	}
}

func TestParseExecutor_SuccessCommitFailureDoesNotWriteFailureAfterPartialSuccess(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.completeSuccessErr = errors.New("outbox unavailable")
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if outcome.Succeeded || outcome.ErrorCode != "TARGET_IMPORT_FAILED" || !outcome.Retryable {
		t.Fatalf("atomic success commit failure must be retryable TARGET_IMPORT_FAILED, got %+v", outcome)
	}
	if store.completeSuccessIn == nil {
		t.Fatal("success side effects were not delegated to the atomic store method")
	}
	if store.completeFailureIn != nil {
		t.Fatalf("must not write analysis.failed after a failed success transaction: failure=%+v", store.completeFailureIn)
	}
	if store.applyResultIn != nil {
		t.Fatalf("parse result was applied outside the atomic success transaction: %+v", store.applyResultIn)
	}
}

func TestParseExecutor_MissingTargetIsTerminalWithoutFailureCleanup(t *testing.T) {
	exec, store, _, _ := newParseExecutorWithFakes(t)
	store.getErr = targetjob.ErrTargetJobNotFound

	outcome := exec.Handle(context.Background(), runner.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-archived",
	})

	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != "TARGET_JOB_NOT_FOUND" {
		t.Fatalf("missing or archived target must be terminal TARGET_JOB_NOT_FOUND, got %+v", outcome)
	}
	if store.completeFailureIn != nil || store.failedOutboxPayload != nil {
		t.Fatalf("missing or archived target must not run parse-failure cleanup: failure=%+v payload=%s", store.completeFailureIn, string(store.failedOutboxPayload))
	}
}

func TestParseExecutor_F3FailClosed(t *testing.T) {
	exec, store, registry, _ := newParseExecutorWithFakes(t)
	registry.err = targetjob.ErrPromptUnsupported
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", RawJDText: "x"}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != "AI_PROVIDER_CONFIG_INVALID" {
		t.Fatalf("F3 failure must map to AI_PROVIDER_CONFIG_INVALID non-retryable, got %+v", outcome)
	}
	if store.failedOutboxPayload == nil {
		t.Fatal("target.analysis.failed outbox payload missing")
	}
}

func TestParseExecutor_AIProviderTimeout_Retryable(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.err = errors.New("provider error: AI_PROVIDER_TIMEOUT context deadline")
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", RawJDText: "x"}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_PROVIDER_TIMEOUT" || !outcome.Retryable {
		t.Fatalf("AI timeout must be retryable, got %+v", outcome)
	}
}

func TestParseExecutor_AIOutputInvalid_NonRetryable(t *testing.T) {
	exec, store, _, ai := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: "not-json"}
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", RawJDText: "x"}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
		t.Fatalf("non-JSON response must map to AI_OUTPUT_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_BlankPersistedJDText_IsTerminalImportFailure(t *testing.T) {
	exec, store, _, _ := newParseExecutorWithFakes(t)
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", RawJDText: "   "}
	outcome := exec.Handle(context.Background(), runner.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_FAILED" || outcome.Retryable {
		t.Fatalf("persisted blank JD must map to TARGET_IMPORT_FAILED non-retryable, got %+v", outcome)
	}
}
