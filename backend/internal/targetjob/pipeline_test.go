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
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
	"github.com/monshunter/easyinterview/backend/internal/targetjob/urlfetch"
)

// ----- pipeline test fakes -----

type pipelineFakeStore struct {
	queuedJobs []targetjob.ClaimedJob
	finalize   []targetjob.JobOutcome
	finalizeID []string

	// parse executor side
	target              targetjob.TargetJobRecord
	sources             []targetjob.SourceRecord
	getErr              error
	applyResultIn       *targetjob.ApplyParseResultInput
	applyResultErr      error
	completeSuccessIn   *targetjob.CompleteParseSuccessInput
	completeSuccessErr  error
	completeFailureIn   *targetjob.CompleteParseFailureInput
	completeFailureErr  error
	parsedOutboxPayload []byte
	failedOutboxPayload []byte
	sourceRefreshCalled bool
	sourceFreshnessUpd  string
	sourceSnapshotURL   string
	sourceSnapshotText  string
	sourceSnapshotAt    *time.Time
	updateAnalysisFail  int
	pollMu              chan struct{}
}

func (s *pipelineFakeStore) ImportTargetJob(context.Context, targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
	return targetjob.ImportTargetJobResult{}, nil
}
func (s *pipelineFakeStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	return nil
}
func (s *pipelineFakeStore) InsertTargetJobSource(context.Context, targetjob.SourceRecord) error {
	return nil
}
func (s *pipelineFakeStore) GetTargetJobByUser(context.Context, string, string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, []targetjob.SourceRecord, error) {
	return targetjob.TargetJobRecord{}, nil, nil, nil
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
	if in.SourceRefreshJobID != "" {
		s.sourceRefreshCalled = true
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
func (s *pipelineFakeStore) UpdateSourceFreshness(_ context.Context, _ string, status targetjob.FreshnessStatus, _ time.Time) error {
	s.sourceFreshnessUpd = string(status)
	return nil
}
func (s *pipelineFakeStore) UpdateSourceSnapshot(_ context.Context, _ string, sanitizedURL string, snapshotText string, fetchedAt time.Time, _ time.Time) error {
	s.sourceSnapshotURL = sanitizedURL
	s.sourceSnapshotText = snapshotText
	cp := fetchedAt
	s.sourceSnapshotAt = &cp
	return nil
}
func (s *pipelineFakeStore) LookupFileAttachmentForUser(context.Context, string, string) (targetjob.FileAttachmentRecord, error) {
	return targetjob.FileAttachmentRecord{}, nil
}
func (s *pipelineFakeStore) ClaimNextAsyncJob(_ context.Context, _ []string, _ time.Time) (targetjob.ClaimedJob, bool, error) {
	if s.pollMu != nil {
		<-s.pollMu
	}
	if len(s.queuedJobs) == 0 {
		return targetjob.ClaimedJob{}, false, nil
	}
	job := s.queuedJobs[0]
	s.queuedJobs = s.queuedJobs[1:]
	return job, true, nil
}
func (s *pipelineFakeStore) FinalizeAsyncJob(_ context.Context, jobID string, outcome targetjob.JobOutcome, _ time.Time) error {
	s.finalizeID = append(s.finalizeID, jobID)
	s.finalize = append(s.finalize, outcome)
	return nil
}
func (s *pipelineFakeStore) EnqueueSourceRefresh(context.Context, string, string, time.Time) error {
	s.sourceRefreshCalled = true
	return nil
}
func (s *pipelineFakeStore) WriteParseFailedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.failedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) WriteTargetParsedOutbox(_ context.Context, _ string, _ string, payload []byte, _ time.Time) error {
	s.parsedOutboxPayload = append([]byte{}, payload...)
	return nil
}
func (s *pipelineFakeStore) GetTargetJobForParse(context.Context, string) (targetjob.TargetJobRecord, []targetjob.SourceRecord, error) {
	if s.getErr != nil {
		return targetjob.TargetJobRecord{}, nil, s.getErr
	}
	return s.target, s.sources, nil
}
func (s *pipelineFakeStore) UpdateTargetJobAnalysisFailure(context.Context, string, time.Time) error {
	s.updateAnalysisFail++
	return nil
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
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}
func (f *fakeAIClient) Stream(context.Context, string, aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeAIClient) Synthesize(context.Context, string, aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, errors.New("not implemented")
}

type fakeFetcher struct {
	res urlfetch.FetchResult
	err error
}

func (f *fakeFetcher) Fetch(context.Context, string) (urlfetch.FetchResult, error) {
	if f.err != nil {
		return urlfetch.FetchResult{}, f.err
	}
	return f.res, nil
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

func newParseExecutorWithFakes(t *testing.T) (*targetjob.ParseExecutor, *pipelineFakeStore, *fakeRegistry, *fakeAIClient, *fakeFetcher) {
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
	fetcher := &fakeFetcher{}
	exec := targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       ai,
		Fetcher:  fetcher,
		NewID:    idSeq("id"),
		Now:      func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})
	return exec, store, registry, ai, fetcher
}

const happyResponseJSON = `{
  "title": "Senior Backend Engineer",
  "companyName": "Acme",
  "coreThemes": ["api"],
  "interviewHypotheses": ["microservices"],
  "strengths": ["Go"],
  "gaps": ["k8s"],
  "riskSignals": [],
  "requirements": [
    {"kind":"must_have","label":"Go","evidenceLevel":"explicit"},
    {"kind":"interview_focus","label":"system design"}
  ]
}`

func TestParseExecutor_HappyPath(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})
	if !outcome.Succeeded {
		t.Fatalf("happy path must succeed, got %+v", outcome)
	}
	if store.completeSuccessIn == nil || len(store.completeSuccessIn.Requirements) != 2 {
		t.Fatalf("expected 2 requirements applied atomically, got %+v", store.completeSuccessIn)
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
	fitProv := decodeProvenanceForTest(t, store.completeSuccessIn.FitSummary)
	if fitProv["modelId"] != "fixture-model:target-import-parse" {
		t.Fatalf("fitSummary provenance modelId must use A3 resolved model id, got %+v", fitProv)
	}
	if store.parsedOutboxPayload == nil {
		t.Fatal("target.parsed outbox payload missing")
	}
	if !store.sourceRefreshCalled {
		t.Fatal("source_refresh placeholder not enqueued")
	}
}

func TestParseExecutor_HappyPathAcceptsFencedJSON(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: "```json\n" + happyResponseJSON + "\n```"}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "zh-CN",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if !outcome.Succeeded {
		t.Fatalf("fenced JSON should be normalized and parsed, got %+v", outcome)
	}
	if store.completeSuccessIn == nil || store.completeSuccessIn.Title != "Senior Backend Engineer" ||
		store.completeSuccessIn.CompanyName != "Acme" || len(store.completeSuccessIn.Requirements) != 2 {
		t.Fatalf("fenced JSON parse result was not committed: %+v", store.completeSuccessIn)
	}
}

func TestParseExecutor_HappyPathCoalescesMissingCompanyName(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: strings.Replace(happyResponseJSON, `"companyName": "Acme"`, `"companyName": ""`, 1)}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "zh-CN",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
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
	exec, store, registry, ai, _ := newParseExecutorWithFakes(t)
	registry.resolution.SystemMessage = "registry-owned target import system prompt"
	registry.resolution.UserMessageTemplate = "registry-owned target import user prompt: {{jd_text}}"
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "en",
		RawJDText:      "JD text from caller",
	}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})

	if !outcome.Succeeded {
		t.Fatalf("parse should succeed, got %+v", outcome)
	}
	if len(ai.lastPayload.Messages) != 2 {
		t.Fatalf("messages = %+v", ai.lastPayload.Messages)
	}
	if ai.lastPayload.Messages[0].Role != "system" || ai.lastPayload.Messages[0].Content != "registry-owned target import system prompt" {
		t.Fatalf("system message not sourced from registry resolution: %+v", ai.lastPayload.Messages)
	}
	if ai.lastPayload.Messages[1].Role != "user" || ai.lastPayload.Messages[1].Content != "registry-owned target import user prompt: JD text from caller" {
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
			exec, store, _, ai, _ := newParseExecutorWithFakes(t)
			ai.resp = aiclient.CompleteResponse{Content: tc.content}
			store.target = targetjob.TargetJobRecord{
				ID:             "tgt-1",
				UserID:         "user-1",
				SourceType:     targetjob.SourceTypeManualText,
				TargetLanguage: "en",
				RawJDText:      "JD text",
			}

			outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})

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

func TestParseExecutorAITaskRuns(t *testing.T) {
	exec, store, registry, ai, _ := newParseExecutorWithFakes(t)
	store.target = targetjob.TargetJobRecord{
		ID:             "01918fa0-0000-7000-8000-000000002000",
		UserID:         "01918fa0-0000-7000-8000-000000001000",
		SourceType:     targetjob.SourceTypeManualText,
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
	exec = targetjob.NewParseExecutor(targetjob.ParseExecutorOptions{
		Store:    store,
		Registry: registry,
		AI:       wrappedAI,
		Fetcher:  &fakeFetcher{},
		NewID:    idSeq("id"),
		Now:      func() time.Time { return time.Date(2026, 5, 9, 22, 0, 0, 0, time.UTC) },
	})

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
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

	delegated, _, err := client.Complete(context.Background(), "practice.followup.default", aiclient.CompletePayload{
		Messages: []aiclient.Message{{Role: "user", Content: "other"}},
		Metadata: aiclient.CallMetadata{FeatureKey: "practice.followup"},
	})
	if err != nil {
		t.Fatalf("delegated Complete: %v", err)
	}
	if delegated.Content != "delegated" || inner.lastProfileName != "practice.followup.default" {
		t.Fatalf("non target import calls must delegate, got resp=%+v profile=%q", delegated, inner.lastProfileName)
	}
}

func TestParseExecutor_SuccessCommitFailureDoesNotWriteFailureAfterPartialSuccess(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	store.completeSuccessErr = errors.New("outbox unavailable")
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeManualText,
		TargetLanguage: "en",
		RawJDText:      "JD text",
	}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{
		JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1",
	})

	if outcome.Succeeded || outcome.ErrorCode != "TARGET_IMPORT_FAILED" || !outcome.Retryable {
		t.Fatalf("atomic success commit failure must be retryable TARGET_IMPORT_FAILED, got %+v", outcome)
	}
	if store.completeSuccessIn == nil {
		t.Fatal("success side effects were not delegated to the atomic store method")
	}
	if store.updateAnalysisFail != 0 || store.completeFailureIn != nil {
		t.Fatalf("must not write analysis.failed after a failed success transaction: updateFail=%d failure=%+v", store.updateAnalysisFail, store.completeFailureIn)
	}
	if store.applyResultIn != nil {
		t.Fatalf("parse result was applied outside the atomic success transaction: %+v", store.applyResultIn)
	}
}

func TestParseExecutor_F3FailClosed(t *testing.T) {
	exec, store, registry, _, _ := newParseExecutorWithFakes(t)
	registry.err = targetjob.ErrPromptUnsupported
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != "AI_PROVIDER_CONFIG_INVALID" {
		t.Fatalf("F3 failure must map to AI_PROVIDER_CONFIG_INVALID non-retryable, got %+v", outcome)
	}
	if store.failedOutboxPayload == nil {
		t.Fatal("target.analysis.failed outbox payload missing")
	}
}

func TestParseExecutor_AIProviderTimeout_Retryable(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.err = errors.New("provider error: AI_PROVIDER_TIMEOUT context deadline")
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_PROVIDER_TIMEOUT" || !outcome.Retryable {
		t.Fatalf("AI timeout must be retryable, got %+v", outcome)
	}
}

func TestParseExecutor_AIOutputInvalid_NonRetryable(t *testing.T) {
	exec, store, _, ai, _ := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: "not-json"}
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "x"}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "AI_OUTPUT_INVALID" || outcome.Retryable {
		t.Fatalf("non-JSON response must map to AI_OUTPUT_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_BlankJDText_SourceInvalid(t *testing.T) {
	exec, store, _, _, _ := newParseExecutorWithFakes(t)
	store.target = targetjob.TargetJobRecord{ID: "tgt-1", SourceType: targetjob.SourceTypeManualText, RawJDText: "   "}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_INVALID" || outcome.Retryable {
		t.Fatalf("blank JD must map to TARGET_IMPORT_SOURCE_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchUnavailable_Retryable(t *testing.T) {
	exec, store, _, _, fetcher := newParseExecutorWithFakes(t)
	fetcher.err = fmt.Errorf("%w: upstream timeout", urlfetch.ErrSourceUnavailable)
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", SourceType: targetjob.SourceTypeURL, SourceURL: "https://jobs.example.com/1",
		RawJDText: "x",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_UNAVAILABLE" || !outcome.Retryable {
		t.Fatalf("upstream unavailable must be retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchInvalid_NonRetryable(t *testing.T) {
	exec, store, _, _, fetcher := newParseExecutorWithFakes(t)
	fetcher.err = fmt.Errorf("%w: bad scheme", urlfetch.ErrInvalidSource)
	store.target = targetjob.TargetJobRecord{
		ID: "tgt-1", SourceType: targetjob.SourceTypeURL, SourceURL: "https://jobs.example.com/1",
		RawJDText: "x",
	}
	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if outcome.ErrorCode != "TARGET_IMPORT_SOURCE_INVALID" || outcome.Retryable {
		t.Fatalf("invalid source must map to TARGET_IMPORT_SOURCE_INVALID non-retryable, got %+v", outcome)
	}
}

func TestParseExecutor_URLFetchBodyIsPersistedAndParsed(t *testing.T) {
	exec, store, registry, ai, fetcher := newParseExecutorWithFakes(t)
	ai.resp = aiclient.CompleteResponse{Content: happyResponseJSON}
	registry.resolution.UserMessageTemplate = "source={{jd_source_url}}\ntext={{jd_text}}"
	fetchedAt := time.Date(2026, 5, 9, 22, 30, 0, 0, time.UTC)
	fetcher.res = urlfetch.FetchResult{
		SanitizedURL: "https://jobs.example.com/role/1",
		Body:         "Fetched JD body for a Backend Engineer.",
		FetchedAt:    fetchedAt,
	}
	store.target = targetjob.TargetJobRecord{
		ID:             "tgt-1",
		UserID:         "user-1",
		SourceType:     targetjob.SourceTypeURL,
		SourceURL:      "https://jobs.example.com/role/1?token=secret",
		TargetLanguage: "en",
	}
	store.sources = []targetjob.SourceRecord{{ID: "src-1", SourceType: targetjob.SourceTypeURL}}

	outcome := exec.Handle(context.Background(), targetjob.ClaimedJob{JobID: "j-1", ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("URL fetch body should drive parse success, got %+v", outcome)
	}
	if store.sourceSnapshotText != "Fetched JD body for a Backend Engineer." || store.sourceSnapshotAt == nil {
		t.Fatalf("source snapshot not persisted: text=%q at=%v", store.sourceSnapshotText, store.sourceSnapshotAt)
	}
	if store.sourceSnapshotURL != "https://jobs.example.com/role/1" {
		t.Fatalf("sanitized source URL = %q", store.sourceSnapshotURL)
	}
	if store.completeSuccessIn == nil || len(store.completeSuccessIn.Requirements) == 0 {
		t.Fatalf("parse result was not applied from fetched URL body: %+v", store.completeSuccessIn)
	}
	if len(ai.lastPayload.Messages) == 0 {
		t.Fatalf("AI messages missing")
	}
	gotPrompt := ai.lastPayload.Messages[len(ai.lastPayload.Messages)-1].Content
	if !strings.Contains(gotPrompt, "source=https://jobs.example.com/role/1") {
		t.Fatalf("prompt did not include sanitized source URL: %q", gotPrompt)
	}
	if strings.Contains(gotPrompt, "token=secret") || strings.Contains(gotPrompt, "{{jd_source_url}}") {
		t.Fatalf("prompt leaked raw source URL or unresolved placeholder: %q", gotPrompt)
	}
}

func TestSourceRefreshHandler_MarksStale(t *testing.T) {
	store := &pipelineFakeStore{}
	h := &targetjob.SourceRefreshHandler{Store: store, Now: func() time.Time { return time.Now().UTC() }}
	outcome := h.Handle(context.Background(), targetjob.ClaimedJob{ResourceID: "tgt-1"})
	if !outcome.Succeeded {
		t.Fatalf("source refresh must succeed by default, got %+v", outcome)
	}
	if store.sourceFreshnessUpd != "stale" {
		t.Fatalf("source freshness should be stale, got %q", store.sourceFreshnessUpd)
	}
}

func TestDrainer_RunOnceProcessesQueuedJobAndFinalizes(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-1", JobType: "target_import", ResourceType: "target_job", ResourceID: "tgt-1"},
		},
	}
	called := false
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, j targetjob.ClaimedJob) targetjob.JobOutcome {
				called = true
				if j.JobID != "j-1" {
					t.Errorf("unexpected jobID: %q", j.JobID)
				}
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
	})
	processed, err := drainer.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if !processed || !called || len(store.finalize) != 1 || !store.finalize[0].Succeeded {
		t.Fatalf("expected one processed job + finalize succeeded; got processed=%v called=%v finalize=%+v", processed, called, store.finalize)
	}
}

func TestDrainer_RunOnceWithUnknownJobTypeFinalizesNonRetryableFailure(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-9", JobType: "ghost_runner", ResourceType: "target_job", ResourceID: "tgt-9"},
		},
	}
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, _ targetjob.ClaimedJob) targetjob.JobOutcome {
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
	})
	if _, err := drainer.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	// note: jobTypes derived from handlers map = ["target_import"]; the
	// fake store returns the queued job regardless of jobTypes filter to
	// exercise the unknown-handler safety path.
	if len(store.finalize) != 1 || store.finalize[0].Succeeded || store.finalize[0].ErrorCode != "TARGET_IMPORT_FAILED" {
		t.Fatalf("expected unknown-handler failure finalize, got %+v", store.finalize)
	}
}

func TestDrainer_StartShutdownDrainsCleanly(t *testing.T) {
	store := &pipelineFakeStore{
		queuedJobs: []targetjob.ClaimedJob{
			{JobID: "j-1", JobType: "target_import", ResourceID: "tgt-1"},
		},
	}
	processed := make(chan struct{}, 1)
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: store,
		Handlers: map[string]targetjob.JobHandler{
			"target_import": targetjob.JobHandlerFunc(func(_ context.Context, _ targetjob.ClaimedJob) targetjob.JobOutcome {
				processed <- struct{}{}
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
		Workers:      1,
		PollInterval: 10 * time.Millisecond,
	})
	drainer.Start(context.Background())

	select {
	case <-processed:
	case <-time.After(2 * time.Second):
		t.Fatal("worker never picked up queued job")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := drainer.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
