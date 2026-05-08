package targetjob_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

type fakeStore struct {
	captured  targetjob.ImportTargetJobInput
	result    targetjob.ImportTargetJobResult
	err       error
	callCount int

	listResult        targetjob.ListResult
	getRecord         targetjob.TargetJobRecord
	getRequirements   []targetjob.RequirementRecord
	getSources        []targetjob.SourceRecord
	getErr            error
	updateResult      targetjob.TargetJobRecord
	updateErr         error
	capturedListFilter targetjob.ListFilter
	capturedUpdateFields targetjob.UpdateLifecycleFields
	capturedUpdateUser   string
	capturedUpdateTarget string
	getCallCount         int
}

func (f *fakeStore) ImportTargetJob(_ context.Context, in targetjob.ImportTargetJobInput) (targetjob.ImportTargetJobResult, error) {
	f.callCount++
	f.captured = in
	if f.err != nil {
		return targetjob.ImportTargetJobResult{}, f.err
	}
	res := f.result
	if res.TargetJobID == "" {
		res = targetjob.ImportTargetJobResult{
			TargetJobID:  in.TargetJobID,
			JobID:        in.JobID,
			JobStatus:    sharedtypes.JobStatusQueued,
			JobCreatedAt: in.Now,
			JobUpdatedAt: in.Now,
		}
		if in.OutboxEventID == "" {
			res.JobStatus = sharedtypes.JobStatusSucceeded
		}
	}
	return res, nil
}

func (f *fakeStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	panic("not used")
}

func (f *fakeStore) InsertTargetJobSource(context.Context, targetjob.SourceRecord) error {
	panic("not used")
}

func (f *fakeStore) GetTargetJobByUser(_ context.Context, _ string, _ string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, []targetjob.SourceRecord, error) {
	f.getCallCount++
	return f.getRecord, f.getRequirements, f.getSources, f.getErr
}

func (f *fakeStore) ListTargetJobsForUser(_ context.Context, _ string, filter targetjob.ListFilter) (targetjob.ListResult, error) {
	f.capturedListFilter = filter
	return f.listResult, nil
}

func (f *fakeStore) UpdateTargetJobLifecycle(_ context.Context, userID string, targetID string, fields targetjob.UpdateLifecycleFields, _ time.Time) (targetjob.TargetJobRecord, error) {
	f.capturedUpdateUser = userID
	f.capturedUpdateTarget = targetID
	f.capturedUpdateFields = fields
	if f.updateErr != nil {
		return targetjob.TargetJobRecord{}, f.updateErr
	}
	return f.updateResult, nil
}

func (f *fakeStore) ApplyParseResult(context.Context, targetjob.ApplyParseResultInput) error {
	panic("not used")
}

func (f *fakeStore) UpdateSourceFreshness(context.Context, string, targetjob.FreshnessStatus, time.Time) error {
	panic("not used")
}

func newServiceWithFake(ids ...string) (*targetjob.Service, *fakeStore) {
	store := &fakeStore{}
	idx := 0
	gen := func() string {
		if idx >= len(ids) {
			panic("ran out of ids in test")
		}
		v := ids[idx]
		idx++
		return v
	}
	now := time.Date(2026, 5, 9, 19, 0, 0, 0, time.UTC)
	svc := targetjob.NewService(targetjob.ServiceOptions{
		Store:        store,
		NewID:        gen,
		Now:          func() time.Time { return now },
		DedupePepper: "test-pepper",
	})
	return svc, store
}

func TestService_ImportTargetJob_ManualTextRunnerPath(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a1", // target_job id
		"018f2a40-0000-7000-9000-0000000000f1", // job id
		"018f2a40-0000-7000-9000-0000000000c1", // source id
		"018f2a40-0000-7000-9000-0000000000e1", // outbox id
	)

	resp, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-1",
		TargetLanguage: "en",
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": "We are hiring a Backend Engineer with strong Go experience.",
		},
	})
	if err != nil {
		t.Fatalf("ImportTargetJob: %v", err)
	}
	if resp.Job.JobType != api.JobTypeTargetImport || resp.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("unexpected job: %+v", resp.Job)
	}
	if resp.TargetJobID != "018f2a40-0000-7000-9000-0000000000a1" {
		t.Fatalf("target id: %s", resp.TargetJobID)
	}
	if store.captured.APISourceType != targetjob.SourceTypeManualText {
		t.Fatalf("captured source type: %s", store.captured.APISourceType)
	}
	if store.captured.OutboxEventID == "" || store.captured.JobPayload == nil || store.captured.OutboxEventPayload == nil {
		t.Fatal("runner envelope must be attached for manual_text")
	}
	// Verify outbox payload sourceType is the event-local "text"
	var outbox map[string]any
	if err := json.Unmarshal(store.captured.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("unmarshal outbox: %v", err)
	}
	if outbox["sourceType"] != "text" {
		t.Fatalf("manual_text must map to event sourceType=text, got %v", outbox["sourceType"])
	}
	if !strings.Contains(string(store.captured.JobPayload), `"sourceType":"manual_text"`) {
		t.Fatalf("job payload missing manual_text source: %s", string(store.captured.JobPayload))
	}
	if store.captured.SourceSnapshotText == "" {
		t.Fatal("manual_text snapshot_text must be set to rawText")
	}
}

func TestService_ImportTargetJob_URLRunnerPathValidatesHttps(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a2",
		"018f2a40-0000-7000-9000-0000000000f2",
		"018f2a40-0000-7000-9000-0000000000c2",
		"018f2a40-0000-7000-9000-0000000000e2",
	)

	resp, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-2",
		TargetLanguage: "zh-CN",
		Source: map[string]any{
			"type": "url",
			"url":  "https://jobs.example.com/role/123#share",
		},
	})
	if err != nil {
		t.Fatalf("ImportTargetJob URL: %v", err)
	}
	if resp.Job.Status != sharedtypes.JobStatusQueued {
		t.Fatalf("URL path must yield queued, got %s", resp.Job.Status)
	}
	if !strings.HasPrefix(store.captured.SourceURL, "https://jobs.example.com") {
		t.Fatalf("URL not preserved: %s", store.captured.SourceURL)
	}
	if strings.Contains(store.captured.SourceURL, "#share") {
		t.Fatal("fragment must be stripped per Phase 2.1 sanitisation")
	}
}

func TestService_ImportTargetJob_URLRejectsHTTP(t *testing.T) {
	svc, _ := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a3",
		"018f2a40-0000-7000-9000-0000000000f3",
	)
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-3",
		TargetLanguage: "en",
		Source: map[string]any{
			"type": "url",
			"url":  "http://insecure.example.com",
		},
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_IMPORT_SOURCE_INVALID" {
		t.Fatalf("expected TARGET_IMPORT_SOURCE_INVALID, got %v", err)
	}
}

func TestService_ImportTargetJob_FilePath(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a4",
		"018f2a40-0000-7000-9000-0000000000f4",
		"018f2a40-0000-7000-9000-0000000000c4",
		"018f2a40-0000-7000-9000-0000000000e4",
	)
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-4",
		TargetLanguage: "en",
		Source: map[string]any{
			"type":         "file",
			"fileObjectId": "018f2a40-0000-7000-9000-0000000000ff",
		},
	})
	if err != nil {
		t.Fatalf("file path: %v", err)
	}
	if store.captured.SourceFileObjectID != "018f2a40-0000-7000-9000-0000000000ff" {
		t.Fatalf("file object id not propagated: %s", store.captured.SourceFileObjectID)
	}
	if store.captured.OutboxEventID == "" {
		t.Fatal("file path must attach runner envelope")
	}
}

func TestService_ImportTargetJob_ManualFormSyncReady(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a5", // target id
		"018f2a40-0000-7000-9000-0000000000f5", // job id
		"018f2a40-0000-7000-9000-0000000000d5", // requirement id
	)
	resp, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-5",
		TargetLanguage: "en",
		Source: map[string]any{
			"type":           "manual_form",
			"title":          "Senior PM",
			"companyName":    "Acme",
			"rawDescription": "Lead product strategy.",
		},
	})
	if err != nil {
		t.Fatalf("manual_form: %v", err)
	}
	if resp.Job.Status != sharedtypes.JobStatusSucceeded {
		t.Fatalf("manual_form must yield succeeded, got %s", resp.Job.Status)
	}
	if store.captured.OutboxEventID != "" {
		t.Fatal("manual_form must NOT attach outbox envelope")
	}
	if len(store.captured.DraftRequirements) != 1 || store.captured.DraftRequirements[0].Kind != targetjob.RequirementMustHave {
		t.Fatalf("manual_form must seed at least 1 must_have draft: %+v", store.captured.DraftRequirements)
	}
	if store.captured.InitialAnalysisStatus != sharedtypes.TargetJobParseStatusReady {
		t.Fatalf("manual_form analysis_status must be ready, got %s", store.captured.InitialAnalysisStatus)
	}
}

func TestService_ImportTargetJob_RequiresIdempotencyKey(t *testing.T) {
	svc, _ := newServiceWithFake("a", "b", "c", "d")
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "u",
		TargetLanguage: "en",
		Source:         map[string]any{"type": "manual_text", "rawText": "x"},
	})
	if !errors.Is(err, targetjob.ErrIdempotencyKeyRequired) {
		t.Fatalf("expected ErrIdempotencyKeyRequired, got %v", err)
	}
}

func TestService_ImportTargetJob_DedupeKeyIsUserScoped(t *testing.T) {
	svc, store := newServiceWithFake("a", "b", "c", "d")
	_, _ = svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "user-1",
		IdempotencyKey: "shared-key",
		TargetLanguage: "en",
		Source:         map[string]any{"type": "manual_text", "rawText": "JD A"},
	})
	keyForUser1 := store.captured.DedupeKey

	svc2, store2 := newServiceWithFake("e", "f", "g", "h")
	_, _ = svc2.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "user-2",
		IdempotencyKey: "shared-key", // same idempotency key as user-1
		TargetLanguage: "en",
		Source:         map[string]any{"type": "manual_text", "rawText": "JD B"},
	})
	keyForUser2 := store2.captured.DedupeKey

	if keyForUser1 == "" || keyForUser2 == "" {
		t.Fatal("dedupe keys must be set")
	}
	if keyForUser1 == keyForUser2 {
		t.Fatalf("dedupe key must be user-scoped, both got %s", keyForUser1)
	}
}

func TestService_ListTargetJobs_PassesFiltersAndShapesPaginated(t *testing.T) {
	svc, store := newServiceWithFake()
	created := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store.listResult = targetjob.ListResult{
		Items: []targetjob.TargetJobRecord{{
			ID:             "018f2a40-0000-7000-9000-0000000000a1",
			UserID:         "u1",
			Status:         sharedtypes.TargetJobStatusPreparing,
			AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
			Title:          "Backend",
			SourceType:     targetjob.SourceTypeManualText,
			TargetLanguage: "en",
			CreatedAt:      created,
			UpdatedAt:      created,
		}},
		NextCursor: "cursor-2",
		HasMore:    true,
	}
	status := sharedtypes.TargetJobStatusPreparing
	res, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{
		UserID:      "u1",
		Status:      &status,
		SearchQuery: "go",
		PageSize:    20,
	})
	if err != nil {
		t.Fatalf("ListTargetJobs: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].Title != "Backend" {
		t.Fatalf("unexpected list response: %+v", res)
	}
	if !res.PageInfo.HasMore || res.PageInfo.NextCursor == nil || *res.PageInfo.NextCursor != "cursor-2" {
		t.Fatalf("page info not propagated: %+v", res.PageInfo)
	}
	if store.capturedListFilter.Status == nil || *store.capturedListFilter.Status != sharedtypes.TargetJobStatusPreparing {
		t.Fatal("status filter not propagated to store")
	}
}

func TestService_GetTargetJob_NotFoundMaps404Code(t *testing.T) {
	svc, store := newServiceWithFake()
	store.getErr = targetjob.ErrTargetJobNotFound
	_, err := svc.GetTargetJob(context.Background(), "u1", "018f2a40-0000-7000-9000-0000000000a1")
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_JOB_NOT_FOUND" {
		t.Fatalf("expected TARGET_JOB_NOT_FOUND, got %v", err)
	}
}

func TestService_UpdateTargetJob_AllowsLegalTransition(t *testing.T) {
	svc, store := newServiceWithFake()
	now := time.Date(2026, 5, 9, 21, 0, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:        "018f2a40-0000-7000-9000-0000000000a1",
		UserID:    "u1",
		Status:    sharedtypes.TargetJobStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
		SourceType: targetjob.SourceTypeManualText,
		TargetLanguage: "en",
	}
	store.updateResult = targetjob.TargetJobRecord{
		ID:        "018f2a40-0000-7000-9000-0000000000a1",
		UserID:    "u1",
		Status:    sharedtypes.TargetJobStatusPreparing,
		CreatedAt: now,
		UpdatedAt: now,
		SourceType: targetjob.SourceTypeManualText,
		TargetLanguage: "en",
	}
	target := sharedtypes.TargetJobStatusPreparing
	out, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         "u1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "k",
		Status:         &target,
	})
	if err != nil {
		t.Fatalf("UpdateTargetJob: %v", err)
	}
	if out.Status != sharedtypes.TargetJobStatusPreparing {
		t.Fatalf("status not applied: %s", out.Status)
	}
}

func TestService_UpdateTargetJob_RejectsIllegalTransition(t *testing.T) {
	svc, store := newServiceWithFake()
	store.getRecord = targetjob.TargetJobRecord{
		ID:     "018f2a40-0000-7000-9000-0000000000a1",
		UserID: "u1",
		Status: sharedtypes.TargetJobStatusDraft,
		SourceType: targetjob.SourceTypeManualText,
		TargetLanguage: "en",
	}
	target := sharedtypes.TargetJobStatusOffer // draft -> offer is invalid
	_, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         "u1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "k",
		Status:         &target,
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_INVALID_STATE_TRANSITION" {
		t.Fatalf("expected TARGET_INVALID_STATE_TRANSITION, got %v", err)
	}
}

func TestService_UpdateTargetJob_AllowsArchivedFromAnyState(t *testing.T) {
	svc, store := newServiceWithFake()
	store.getRecord = targetjob.TargetJobRecord{
		ID: "018f2a40-0000-7000-9000-0000000000a1",
		UserID: "u1",
		Status: sharedtypes.TargetJobStatusOffer,
		SourceType: targetjob.SourceTypeManualText,
		TargetLanguage: "en",
	}
	store.updateResult = store.getRecord
	store.updateResult.Status = sharedtypes.TargetJobStatusArchived
	target := sharedtypes.TargetJobStatusArchived
	if _, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID: "u1", TargetJobID: "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "k", Status: &target,
	}); err != nil {
		t.Fatalf("offer -> archived must succeed: %v", err)
	}
}

func TestService_UpdateTargetJob_RequiresIdempotencyKey(t *testing.T) {
	svc, _ := newServiceWithFake()
	_, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID: "u1", TargetJobID: "t1",
	})
	if !errors.Is(err, targetjob.ErrIdempotencyKeyRequired) {
		t.Fatalf("expected ErrIdempotencyKeyRequired, got %v", err)
	}
}

func TestService_ImportTargetJob_RejectsUnknownSource(t *testing.T) {
	svc, _ := newServiceWithFake("a", "b")
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "u",
		IdempotencyKey: "k",
		TargetLanguage: "en",
		Source:         map[string]any{"type": "satellite_uplink"},
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_IMPORT_SOURCE_INVALID" {
		t.Fatalf("expected TARGET_IMPORT_SOURCE_INVALID, got %v", err)
	}
}
