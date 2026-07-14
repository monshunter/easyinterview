package targetjob_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

const testResumeID = "018f2a40-0000-7000-9000-0000000000r1"

type fakeStore struct {
	captured  targetjob.ImportTargetJobInput
	result    targetjob.ImportTargetJobResult
	err       error
	callCount int

	listResult           targetjob.ListResult
	getRecord            targetjob.TargetJobRecord
	getRequirements      []targetjob.RequirementRecord
	getErr               error
	updateResult         targetjob.TargetJobRecord
	updateErr            error
	capturedListFilter   targetjob.ListFilter
	capturedUpdateFields targetjob.UpdateLifecycleFields
	capturedUpdateUser   string
	capturedUpdateTarget string
	getCallCount         int
	updateDedupeHit      bool
	updateDedupeRecord   targetjob.TargetJobRecord
	updateDedupeReqs     []targetjob.RequirementRecord
	updateDedupeErr      error
	archiveResult        targetjob.TargetJobRecord
	archiveErr           error
	capturedArchiveInput targetjob.ArchiveTargetJobInput
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
	}
	return res, nil
}

func (f *fakeStore) InsertTargetJob(context.Context, targetjob.TargetJobRecord) error {
	panic("not used")
}

func (f *fakeStore) GetTargetJobByUser(_ context.Context, _ string, _ string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, error) {
	f.getCallCount++
	return f.getRecord, f.getRequirements, f.getErr
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

func (f *fakeStore) LookupUpdateDedupe(_ context.Context, _ string, _ string) (targetjob.TargetJobRecord, []targetjob.RequirementRecord, bool, error) {
	if f.updateDedupeErr != nil {
		return targetjob.TargetJobRecord{}, nil, false, f.updateDedupeErr
	}
	return f.updateDedupeRecord, f.updateDedupeReqs, f.updateDedupeHit, nil
}

func (f *fakeStore) ArchiveTargetJob(_ context.Context, in targetjob.ArchiveTargetJobInput) (targetjob.TargetJobRecord, error) {
	f.capturedArchiveInput = in
	if f.archiveErr != nil {
		return targetjob.TargetJobRecord{}, f.archiveErr
	}
	return f.archiveResult, nil
}

func (f *fakeStore) ApplyParseResult(context.Context, targetjob.ApplyParseResultInput) error {
	panic("not used")
}
func (f *fakeStore) CompleteParseSuccess(context.Context, targetjob.CompleteParseSuccessInput) error {
	panic("not used")
}
func (f *fakeStore) CompleteParseFailure(context.Context, targetjob.CompleteParseFailureInput) error {
	panic("not used")
}

func (f *fakeStore) WriteParseFailedOutbox(context.Context, string, string, []byte, time.Time) error {
	return nil
}
func (f *fakeStore) WriteTargetParsedOutbox(context.Context, string, string, []byte, time.Time) error {
	return nil
}
func (f *fakeStore) GetTargetJobForParse(context.Context, string) (targetjob.TargetJobRecord, error) {
	return targetjob.TargetJobRecord{}, nil
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

func TestService_ImportTargetJob_DirectRawTextRunnerPath(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a1", // target_job id
		"018f2a40-0000-7000-9000-0000000000f1", // job id
		"018f2a40-0000-7000-9000-0000000000e1", // outbox id
	)

	resp, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-1",
		TargetLanguage: "en",
		ResumeID:       testResumeID,
		RawText:        "  We are hiring a Backend Engineer with strong Go experience.  ",
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
	if store.captured.OutboxEventID == "" || store.captured.JobPayload == nil || store.captured.OutboxEventPayload == nil {
		t.Fatal("runner envelope must be attached for direct rawText import")
	}
	if store.captured.RawJDText != "We are hiring a Backend Engineer with strong Go experience." {
		t.Fatalf("rawText was not trimmed and persisted exactly: %q", store.captured.RawJDText)
	}
	var outbox map[string]any
	if err := json.Unmarshal(store.captured.OutboxEventPayload, &outbox); err != nil {
		t.Fatalf("unmarshal outbox: %v", err)
	}
	if len(outbox) != 3 || outbox["targetJobId"] == nil || outbox["userId"] == nil || outbox["targetLanguage"] == nil {
		t.Fatalf("outbox payload must contain only targetJobId/userId/targetLanguage: %v", outbox)
	}
	if _, ok := outbox["sourceType"]; ok {
		t.Fatalf("outbox payload must be source-free: %v", outbox)
	}
	var jobPayload map[string]any
	if err := json.Unmarshal(store.captured.JobPayload, &jobPayload); err != nil {
		t.Fatalf("unmarshal job payload: %v", err)
	}
	if len(jobPayload) != 3 || jobPayload["targetJobId"] == nil || jobPayload["userId"] == nil || jobPayload["targetLanguage"] == nil {
		t.Fatalf("job payload must contain only targetJobId/userId/targetLanguage: %v", jobPayload)
	}
	if _, ok := jobPayload["sourceType"]; ok {
		t.Fatalf("job payload must be source-free: %v", jobPayload)
	}
	if store.captured.ResumeID != testResumeID {
		t.Fatalf("resume binding was not passed to store: %+v", store.captured)
	}
}

func TestService_ImportTargetJob_RejectsBlankRawText(t *testing.T) {
	svc, store := newServiceWithFake()
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-blank",
		TargetLanguage: "en",
		ResumeID:       testResumeID,
		RawText:        " \n\t ",
	})
	var svcErr *targetjob.ServiceImportError
	if !errors.As(err, &svcErr) || svcErr.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("expected VALIDATION_FAILED for blank rawText, got %v", err)
	}
	if store.callCount != 0 {
		t.Fatalf("blank rawText must fail before persistence, calls=%d", store.callCount)
	}
}

func TestService_ImportTargetJob_RequiresResumeID(t *testing.T) {
	svc, _ := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000a1",
		"018f2a40-0000-7000-9000-0000000000f1",
	)
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "018f2a40-0000-7000-9000-0000000000b1",
		IdempotencyKey: "key-missing-resume",
		TargetLanguage: "en",
		RawText:        "We are hiring a Backend Engineer.",
	})
	var svcErr *targetjob.ServiceImportError
	if !errors.As(err, &svcErr) || svcErr.Code != "VALIDATION_FAILED" {
		t.Fatalf("expected VALIDATION_FAILED for missing resumeId, got %v", err)
	}
}

func TestService_ImportTargetJob_RequiresIdempotencyKey(t *testing.T) {
	svc, _ := newServiceWithFake("a", "b", "c", "d")
	_, err := svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "u",
		TargetLanguage: "en",
		ResumeID:       testResumeID,
		RawText:        "x",
	})
	if !errors.Is(err, targetjob.ErrIdempotencyKeyRequired) {
		t.Fatalf("expected ErrIdempotencyKeyRequired, got %v", err)
	}
}

func TestService_ImportTargetJob_DedupeKeyIsUserScoped(t *testing.T) {
	svc, store := newServiceWithFake("a", "b", "c")
	_, _ = svc.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "user-1",
		IdempotencyKey: "shared-key",
		TargetLanguage: "en",
		ResumeID:       testResumeID,
		RawText:        "JD A",
	})
	keyForUser1 := store.captured.DedupeKey

	svc2, store2 := newServiceWithFake("e", "f", "g")
	_, _ = svc2.ImportTargetJob(context.Background(), targetjob.ImportRequest{
		UserID:         "user-2",
		IdempotencyKey: "shared-key", // same idempotency key as user-1
		TargetLanguage: "en",
		ResumeID:       testResumeID,
		RawText:        "JD B",
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
	if res.PageInfo.PageSize != 20 {
		t.Fatalf("page size = %d, want 20", res.PageInfo.PageSize)
	}
	if store.capturedListFilter.Status == nil || *store.capturedListFilter.Status != sharedtypes.TargetJobStatusPreparing {
		t.Fatal("status filter not propagated to store")
	}
}

func TestService_ListTargetJobs_PageInfoReportsEffectivePageSize(t *testing.T) {
	cases := []struct {
		name string
		in   int32
		want int
	}{
		{name: "default", in: 0, want: 20},
		{name: "negative defaults", in: -10, want: 20},
		{name: "explicit", in: 7, want: 7},
		{name: "clamped max", in: 1000, want: 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newServiceWithFake()
			res, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{
				UserID:   "u1",
				PageSize: tc.in,
			})
			if err != nil {
				t.Fatalf("ListTargetJobs: %v", err)
			}
			if res.PageInfo.PageSize != tc.want {
				t.Fatalf("page size = %d, want %d", res.PageInfo.PageSize, tc.want)
			}
		})
	}
}

func TestService_ListTargetJobs_ProjectsCanonicalPracticeProgressIndependentOfLifecycleStatus(t *testing.T) {
	statuses := []sharedtypes.TargetJobStatus{
		sharedtypes.TargetJobStatusDraft,
		sharedtypes.TargetJobStatusInterviewing,
		sharedtypes.TargetJobStatusOffer,
	}
	for _, lifecycleStatus := range statuses {
		t.Run(string(lifecycleStatus), func(t *testing.T) {
			svc, store := newServiceWithFake()
			created := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
			store.listResult = targetjob.ListResult{Items: []targetjob.TargetJobRecord{{
				ID:                  "018f2a40-0000-7000-9000-0000000000a1",
				UserID:              "u1",
				Status:              lifecycleStatus,
				AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
				Title:               "Backend",
				TargetLanguage:      "en",
				Summary:             threeRoundSummaryJSON(),
				PracticeFactsLoaded: true,
				CompletedRoundFacts: []targetjob.PracticeRoundFact{
					{RoundID: "round-3-manager", RoundSequence: 3},
					{RoundID: "round-1-hr", RoundSequence: 1},
					{RoundID: "round-3-manager", RoundSequence: 3},
					{RoundID: "round-99-other", RoundSequence: 99},
					{},
				},
				ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{
					{PlanID: "old-round-newer", RoundID: "round-1-hr", RoundSequence: 1, CreatedAt: created.Add(3 * time.Hour)},
					{PlanID: "current-round-older", RoundID: "round-2-technical", RoundSequence: 2, CreatedAt: created.Add(time.Hour)},
					{PlanID: "current-round-newest", RoundID: "round-2-technical", RoundSequence: 2, CreatedAt: created.Add(2 * time.Hour)},
					{PlanID: "unknown", RoundID: "round-9-other", RoundSequence: 9, CreatedAt: created.Add(4 * time.Hour)},
				},
				CreatedAt: created,
				UpdatedAt: created,
			}}}

			res, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{UserID: "u1"})
			if err != nil {
				t.Fatalf("ListTargetJobs: %v", err)
			}
			got := res.Items[0]
			if got.PracticeProgress == nil {
				t.Fatal("practiceProgress must be projected for a valid structured summary")
			}
			if got.PracticeProgress.Status != "in_progress" {
				t.Fatalf("status = %q, want in_progress", got.PracticeProgress.Status)
			}
			if len(got.PracticeProgress.CompletedRounds) != 1 ||
				got.PracticeProgress.CompletedRounds[0].RoundId != "round-1-hr" {
				t.Fatalf("completed rounds = %+v", got.PracticeProgress.CompletedRounds)
			}
			if got.PracticeProgress.CurrentRound == nil || got.PracticeProgress.CurrentRound.RoundId != "round-2-technical" {
				t.Fatalf("current round = %+v", got.PracticeProgress.CurrentRound)
			}
			if got.CurrentPracticePlanId == nil || *got.CurrentPracticePlanId != "current-round-newest" {
				t.Fatalf("current practice plan = %v, want current-round-newest", got.CurrentPracticePlanId)
			}
		})
	}
}

func TestService_GetTargetJob_HidesCompletedFactsAfterFirstCanonicalGap(t *testing.T) {
	svc, store := newServiceWithFake()
	now := time.Date(2026, 7, 12, 10, 30, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:                  "target-gap",
		UserID:              "u1",
		Status:              sharedtypes.TargetJobStatusInterviewing,
		AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
		Title:               "Backend",
		TargetLanguage:      "en",
		Summary:             threeRoundSummaryJSON(),
		PracticeFactsLoaded: true,
		CompletedRoundFacts: []targetjob.PracticeRoundFact{{RoundID: "round-3-manager", RoundSequence: 3}, {RoundID: "round-2-technical", RoundSequence: 2}},
		ReadyPlanFacts:      []targetjob.ReadyPracticePlanFact{{PlanID: "round-1-plan", RoundID: "round-1-hr", RoundSequence: 1, CreatedAt: now}},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	got, err := svc.GetTargetJob(context.Background(), "u1", "target-gap")
	if err != nil {
		t.Fatalf("GetTargetJob: %v", err)
	}
	if got.PracticeProgress == nil {
		t.Fatal("practiceProgress must be projected")
	}
	if got.PracticeProgress.Status != "not_started" || len(got.PracticeProgress.CompletedRounds) != 0 {
		t.Fatalf("gap facts must remain hidden: %+v", got.PracticeProgress)
	}
	if got.PracticeProgress.CurrentRound == nil || got.PracticeProgress.CurrentRound.RoundId != "round-1-hr" {
		t.Fatalf("current round = %+v, want round-1-hr", got.PracticeProgress.CurrentRound)
	}
	if got.CurrentPracticePlanId == nil || *got.CurrentPracticePlanId != "round-1-plan" {
		t.Fatalf("current plan = %v, want round-1-plan", got.CurrentPracticePlanId)
	}
}

func TestService_GetTargetJob_ProjectsFirstRoundAndAllCompleted(t *testing.T) {
	created := time.Date(2026, 7, 12, 11, 0, 0, 0, time.UTC)
	t.Run("not started selects first exact ready plan", func(t *testing.T) {
		svc, store := newServiceWithFake()
		store.getRecord = targetjob.TargetJobRecord{
			ID:                  "target-1",
			UserID:              "u1",
			Status:              sharedtypes.TargetJobStatusDraft,
			AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
			Title:               "Backend",
			TargetLanguage:      "en",
			Summary:             threeRoundSummaryJSON(),
			PracticeFactsLoaded: true,
			ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{
				{PlanID: "round-2-plan", RoundID: "round-2-technical", RoundSequence: 2, CreatedAt: created.Add(time.Hour)},
				{PlanID: "round-1-plan", RoundID: "round-1-hr", RoundSequence: 1, CreatedAt: created},
			},
			CreatedAt: created,
			UpdatedAt: created,
		}
		got, err := svc.GetTargetJob(context.Background(), "u1", "target-1")
		if err != nil {
			t.Fatalf("GetTargetJob: %v", err)
		}
		if got.PracticeProgress == nil || got.PracticeProgress.Status != "not_started" || got.PracticeProgress.CurrentRound == nil || got.PracticeProgress.CurrentRound.RoundId != "round-1-hr" {
			t.Fatalf("unexpected initial progress: %+v", got.PracticeProgress)
		}
		if got.CurrentPracticePlanId == nil || *got.CurrentPracticePlanId != "round-1-plan" {
			t.Fatalf("initial plan = %v", got.CurrentPracticePlanId)
		}
	})

	t.Run("all complete clears current round and plan", func(t *testing.T) {
		svc, store := newServiceWithFake()
		store.getRecord = targetjob.TargetJobRecord{
			ID:                  "target-1",
			UserID:              "u1",
			Status:              sharedtypes.TargetJobStatusPreparing,
			AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
			Title:               "Backend",
			TargetLanguage:      "en",
			Summary:             threeRoundSummaryJSON(),
			PracticeFactsLoaded: true,
			CompletedRoundFacts: []targetjob.PracticeRoundFact{
				{RoundID: "round-3-manager", RoundSequence: 3},
				{RoundID: "round-1-hr", RoundSequence: 1},
				{RoundID: "round-2-technical", RoundSequence: 2},
				{RoundID: "round-1-hr", RoundSequence: 1},
			},
			ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{
				{PlanID: "late-retry", RoundID: "round-1-hr", RoundSequence: 1, CreatedAt: created.Add(time.Hour)},
			},
			CreatedAt: created,
			UpdatedAt: created,
		}
		got, err := svc.GetTargetJob(context.Background(), "u1", "target-1")
		if err != nil {
			t.Fatalf("GetTargetJob: %v", err)
		}
		if got.PracticeProgress == nil || got.PracticeProgress.Status != "completed" || got.PracticeProgress.CurrentRound != nil || len(got.PracticeProgress.CompletedRounds) != 3 {
			t.Fatalf("unexpected completed progress: %+v", got.PracticeProgress)
		}
		if got.CurrentPracticePlanId != nil {
			t.Fatalf("completed target must not expose a plan: %v", got.CurrentPracticePlanId)
		}
	})
}

func TestService_GetTargetJob_ProjectsNonContiguousCanonicalSequences(t *testing.T) {
	svc, store := newServiceWithFake()
	now := time.Date(2026, 7, 12, 11, 30, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:                  "target-non-contiguous",
		UserID:              "u1",
		Status:              sharedtypes.TargetJobStatusInterviewing,
		AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
		Title:               "Backend",
		TargetLanguage:      "en",
		Summary:             nonContiguousRoundSummaryJSON(),
		PracticeFactsLoaded: true,
		CompletedRoundFacts: []targetjob.PracticeRoundFact{
			{RoundID: "round-1-hr", RoundSequence: 1},
			{RoundID: "round-2-technical", RoundSequence: 2},
		},
		ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{
			{PlanID: "round-4-plan", RoundID: "round-4-manager", RoundSequence: 4, CreatedAt: now},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	got, err := svc.GetTargetJob(context.Background(), "u1", "target-non-contiguous")
	if err != nil {
		t.Fatalf("GetTargetJob: %v", err)
	}
	if got.PracticeProgress == nil || got.PracticeProgress.Status != "in_progress" || len(got.PracticeProgress.CompletedRounds) != 2 {
		t.Fatalf("unexpected non-contiguous progress: %+v", got.PracticeProgress)
	}
	if got.PracticeProgress.CurrentRound == nil || got.PracticeProgress.CurrentRound.RoundId != "round-4-manager" || got.PracticeProgress.CurrentRound.RoundSequence != 4 {
		t.Fatalf("current round = %+v, want round-4-manager", got.PracticeProgress.CurrentRound)
	}
	if got.CurrentPracticePlanId == nil || *got.CurrentPracticePlanId != "round-4-plan" {
		t.Fatalf("current plan = %v, want round-4-plan", got.CurrentPracticePlanId)
	}

	store.getRecord.CompletedRoundFacts = append(store.getRecord.CompletedRoundFacts, targetjob.PracticeRoundFact{RoundID: "round-4-manager", RoundSequence: 4})
	got, err = svc.GetTargetJob(context.Background(), "u1", "target-non-contiguous")
	if err != nil {
		t.Fatalf("GetTargetJob final: %v", err)
	}
	if got.PracticeProgress == nil || got.PracticeProgress.Status != "completed" || got.PracticeProgress.CurrentRound != nil || len(got.PracticeProgress.CompletedRounds) != 3 || got.CurrentPracticePlanId != nil {
		t.Fatalf("unexpected final non-contiguous progress: progress=%+v plan=%v", got.PracticeProgress, got.CurrentPracticePlanId)
	}
}

func TestService_GetAndListTargetJob_PracticeProgressFailsClosedForInvalidSummary(t *testing.T) {
	cases := []struct {
		name    string
		summary json.RawMessage
	}{
		{name: "unloaded", summary: nil},
		{name: "malformed", summary: json.RawMessage(`{"interviewRounds":`)},
		{name: "missing provenance", summary: json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":15}]}`)},
		{name: "too few rounds", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":15}]`)},
		{name: "duplicate sequence", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":15},{"sequence":1,"type":"technical","name":"Tech","focus":"code","durationMinutes":45}]`)},
		{name: "zero sequence", summary: summaryWithRoundsJSON(`[{"sequence":0,"type":"hr","name":"HR","focus":"fit","durationMinutes":15},{"sequence":2,"type":"technical","name":"Tech","focus":"code","durationMinutes":45}]`)},
		{name: "unknown type", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"sales","name":"Sales","focus":"pitch","durationMinutes":30},{"sequence":4,"type":"manager","name":"Manager","focus":"ownership","durationMinutes":30}]`)},
		{name: "duration out of bounds", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":5},{"sequence":2,"type":"technical","name":"Tech","focus":"code","durationMinutes":45}]`)},
		{name: "blank name", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"hr","name":"  ","focus":"fit","durationMinutes":15},{"sequence":2,"type":"manager","name":"Manager","focus":"ownership","durationMinutes":30}]`)},
		{name: "blank focus", summary: summaryWithRoundsJSON(`[{"sequence":1,"type":"hr","name":"HR","focus":"  ","durationMinutes":15},{"sequence":2,"type":"manager","name":"Manager","focus":"ownership","durationMinutes":30}]`)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, store := newServiceWithFake()
			rec := targetjob.TargetJobRecord{
				ID:                  "target-1",
				UserID:              "u1",
				Status:              sharedtypes.TargetJobStatusInterviewing,
				AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
				Title:               "Backend",
				TargetLanguage:      "en",
				Summary:             tc.summary,
				PracticeFactsLoaded: true,
				ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{{
					PlanID: "round-1-plan", RoundID: "round-1-hr", RoundSequence: 1,
				}},
				CreatedAt: time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC),
			}
			store.getRecord = rec
			store.listResult = targetjob.ListResult{Items: []targetjob.TargetJobRecord{rec}}

			gotDetail, err := svc.GetTargetJob(context.Background(), "u1", "target-1")
			if err != nil {
				t.Fatalf("GetTargetJob: %v", err)
			}
			gotList, err := svc.ListTargetJobs(context.Background(), targetjob.ListRequest{UserID: "u1"})
			if err != nil {
				t.Fatalf("ListTargetJobs: %v", err)
			}
			for name, got := range map[string]api.TargetJob{"get": gotDetail, "list": gotList.Items[0]} {
				if got.PracticeProgress != nil || got.CurrentPracticePlanId != nil {
					t.Fatalf("%s must fail closed, progress=%+v plan=%v", name, got.PracticeProgress, got.CurrentPracticePlanId)
				}
			}
		})
	}
}

func TestService_GetTargetJob_PracticeProgressFailsClosedWhenFactsAreUnloaded(t *testing.T) {
	svc, store := newServiceWithFake()
	now := time.Date(2026, 7, 12, 12, 30, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             "target-1",
		UserID:         "u1",
		Status:         sharedtypes.TargetJobStatusInterviewing,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Backend",
		TargetLanguage: "en",
		Summary:        threeRoundSummaryJSON(),
		ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{{
			PlanID: "round-1-plan", RoundID: "round-1-hr", RoundSequence: 1, CreatedAt: now,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	got, err := svc.GetTargetJob(context.Background(), "u1", "target-1")
	if err != nil {
		t.Fatalf("GetTargetJob: %v", err)
	}
	if got.PracticeProgress != nil || got.CurrentPracticePlanId != nil {
		t.Fatalf("unloaded facts must fail closed, progress=%+v plan=%v", got.PracticeProgress, got.CurrentPracticePlanId)
	}
}

func threeRoundSummaryJSON() json.RawMessage {
	return summaryWithRoundsJSON(`[
		{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":15},
		{"sequence":2,"type":"technical","name":"Technical","focus":"code","durationMinutes":45},
		{"sequence":3,"type":"manager","name":"Manager","focus":"ownership","durationMinutes":30}
	]`)
}

func nonContiguousRoundSummaryJSON() json.RawMessage {
	return summaryWithRoundsJSON(`[
		{"sequence":1,"type":"hr","name":"HR","focus":"fit","durationMinutes":15},
		{"sequence":2,"type":"technical","name":"Technical","focus":"code","durationMinutes":45},
		{"sequence":4,"type":"manager","name":"Manager","focus":"ownership","durationMinutes":30}
	]`)
}

func summaryWithRoundsJSON(rounds string) json.RawMessage {
	return json.RawMessage(`{"interviewRounds":` + rounds + `,"provenance":{"promptVersion":"v1","rubricVersion":"not_applicable","modelId":"fixture-model","featureFlag":"none","language":"en","dataSourceVersion":"target-summary-v1"}}`)
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
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d0")
	now := time.Date(2026, 5, 9, 21, 0, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "u1",
		Status:         sharedtypes.TargetJobStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
		TargetLanguage: "en",
	}
	store.updateResult = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "u1",
		Status:         sharedtypes.TargetJobStatusPreparing,
		CreatedAt:      now,
		UpdatedAt:      now,
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

func TestService_UpdateTargetJob_PassesUserScopedDedupeToStore(t *testing.T) {
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d1")
	now := time.Date(2026, 5, 9, 21, 0, 0, 0, time.UTC)
	store.getRecord = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "user-1",
		Status:         sharedtypes.TargetJobStatusDraft,
		CreatedAt:      now,
		UpdatedAt:      now,
		TargetLanguage: "en",
	}
	store.updateResult = store.getRecord
	store.updateResult.Status = sharedtypes.TargetJobStatusPreparing
	target := sharedtypes.TargetJobStatusPreparing

	_, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         "user-1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "same-client-key",
		Status:         &target,
	})
	if err != nil {
		t.Fatalf("UpdateTargetJob: %v", err)
	}
	if store.capturedUpdateFields.DedupeKey == "" {
		t.Fatal("update dedupe key was not passed to the store")
	}
	if strings.Contains(store.capturedUpdateFields.DedupeKey, "same-client-key") {
		t.Fatalf("dedupe key must be hashed/redacted, got %q", store.capturedUpdateFields.DedupeKey)
	}
	if store.capturedUpdateFields.DedupeMarkerID != "018f2a40-0000-7000-9000-0000000000d1" {
		t.Fatalf("marker id = %q", store.capturedUpdateFields.DedupeMarkerID)
	}
}

func TestService_UpdateTargetJob_FirstResponseMatchesIdempotentReplayPracticeProjection(t *testing.T) {
	svc, store := newServiceWithFake("update-marker")
	now := time.Date(2026, 7, 12, 14, 0, 0, 0, time.UTC)
	reloaded := targetjob.TargetJobRecord{
		ID:                  "target-update-progress",
		UserID:              "u1",
		Status:              sharedtypes.TargetJobStatusPreparing,
		AnalysisStatus:      sharedtypes.TargetJobParseStatusReady,
		Title:               "Backend",
		CompanyName:         "Acme",
		TargetLanguage:      "en",
		Summary:             nonContiguousRoundSummaryJSON(),
		PracticeFactsLoaded: true,
		CompletedRoundFacts: []targetjob.PracticeRoundFact{{RoundID: "round-1-hr", RoundSequence: 1}},
		ReadyPlanFacts: []targetjob.ReadyPracticePlanFact{{
			PlanID: "round-2-plan", RoundID: "round-2-technical", RoundSequence: 2, CreatedAt: now,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.updateResult = reloaded
	store.updateResult.PracticeFactsLoaded = false
	store.updateResult.CompletedRoundFacts = nil
	store.updateResult.ReadyPlanFacts = nil
	store.getRecord = reloaded
	status := sharedtypes.TargetJobStatusPreparing
	request := targetjob.UpdateRequest{
		UserID:         "u1",
		TargetJobID:    reloaded.ID,
		IdempotencyKey: "same-update-key",
		Status:         &status,
	}

	first, err := svc.UpdateTargetJob(context.Background(), request)
	if err != nil {
		t.Fatalf("first UpdateTargetJob: %v", err)
	}
	store.updateDedupeHit = true
	store.updateDedupeRecord = reloaded
	replay, err := svc.UpdateTargetJob(context.Background(), request)
	if err != nil {
		t.Fatalf("replay UpdateTargetJob: %v", err)
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("marshal first response: %v", err)
	}
	replayJSON, err := json.Marshal(replay)
	if err != nil {
		t.Fatalf("marshal replay response: %v", err)
	}
	if string(firstJSON) != string(replayJSON) {
		t.Fatalf("first/replay wire mismatch:\nfirst=%s\nreplay=%s", firstJSON, replayJSON)
	}
	if first.PracticeProgress == nil || first.PracticeProgress.CurrentRound == nil || first.PracticeProgress.CurrentRound.RoundId != "round-2-technical" {
		t.Fatalf("first response progress = %+v", first.PracticeProgress)
	}
	if first.CurrentPracticePlanId == nil || *first.CurrentPracticePlanId != "round-2-plan" {
		t.Fatalf("first response plan = %v", first.CurrentPracticePlanId)
	}
}

func TestService_UpdateTargetJob_DedupeHitBypassesLaterStateTransition(t *testing.T) {
	svc, store := newServiceWithFake()
	now := time.Date(2026, 5, 9, 21, 5, 0, 0, time.UTC)
	store.updateDedupeHit = true
	store.updateDedupeRecord = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "user-1",
		Status:         sharedtypes.TargetJobStatusApplied,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		CreatedAt:      now,
		UpdatedAt:      now,
		TargetLanguage: "en",
	}
	status := sharedtypes.TargetJobStatusPreparing

	out, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         "user-1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "same-client-key",
		Status:         &status,
	})
	if err != nil {
		t.Fatalf("dedupe hit should not revalidate later state: %v", err)
	}
	if out.Status != sharedtypes.TargetJobStatusApplied {
		t.Fatalf("dedupe hit returned status %s", out.Status)
	}
	if store.capturedUpdateTarget != "" || store.getCallCount != 0 {
		t.Fatalf("dedupe hit must bypass get/update, get=%d updateTarget=%q", store.getCallCount, store.capturedUpdateTarget)
	}
}

func TestService_UpdateTargetJob_RejectsIllegalTransition(t *testing.T) {
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d4")
	store.updateErr = &targetjob.ServiceImportError{Code: "TARGET_INVALID_STATE_TRANSITION", Message: "transition draft -> offer is not allowed"}
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

func TestService_UpdateTargetJob_DelegatesStatusTransitionValidationToStore(t *testing.T) {
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d3")
	store.updateErr = &targetjob.ServiceImportError{Code: "TARGET_INVALID_STATE_TRANSITION", Message: "stale transition rejected"}
	target := sharedtypes.TargetJobStatusOffer

	_, err := svc.UpdateTargetJob(context.Background(), targetjob.UpdateRequest{
		UserID:         "u1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "k",
		Status:         &target,
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != "TARGET_INVALID_STATE_TRANSITION" {
		t.Fatalf("expected store-sourced TARGET_INVALID_STATE_TRANSITION, got %v", err)
	}
	if store.getCallCount != 0 {
		t.Fatalf("service must not preflight target status outside the store transaction, getCallCount=%d", store.getCallCount)
	}
	if store.capturedUpdateTarget == "" {
		t.Fatal("service did not delegate the status update to the store")
	}
}

func TestService_UpdateTargetJob_AllowsArchivedFromAnyState(t *testing.T) {
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d2")
	store.getRecord = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "u1",
		Status:         sharedtypes.TargetJobStatusOffer,
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

func TestService_ArchiveTargetJob_PersistsWithUserScopedDedupe(t *testing.T) {
	svc, store := newServiceWithFake("018f2a40-0000-7000-9000-0000000000d9")
	now := time.Date(2026, 7, 9, 13, 45, 0, 0, time.UTC)
	store.archiveResult = targetjob.TargetJobRecord{
		ID:             "018f2a40-0000-7000-9000-0000000000a1",
		UserID:         "user-1",
		Status:         sharedtypes.TargetJobStatusArchived,
		AnalysisStatus: sharedtypes.TargetJobParseStatusReady,
		Title:          "Backend Engineer",
		CompanyName:    "Acme",
		TargetLanguage: "en",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	out, err := svc.ArchiveTargetJob(context.Background(), targetjob.ArchiveRequest{
		UserID:         "user-1",
		TargetJobID:    "018f2a40-0000-7000-9000-0000000000a1",
		IdempotencyKey: "same-client-key",
	})
	if err != nil {
		t.Fatalf("ArchiveTargetJob: %v", err)
	}
	if out.Status != sharedtypes.TargetJobStatusArchived {
		t.Fatalf("status = %s, want archived", out.Status)
	}
	if store.capturedArchiveInput.UserID != "user-1" || store.capturedArchiveInput.TargetJobID != "018f2a40-0000-7000-9000-0000000000a1" {
		t.Fatalf("archive input not scoped: %+v", store.capturedArchiveInput)
	}
	if store.capturedArchiveInput.DedupeKey == "" || strings.Contains(store.capturedArchiveInput.DedupeKey, "same-client-key") {
		t.Fatalf("dedupe key must be hashed/redacted, got %q", store.capturedArchiveInput.DedupeKey)
	}
	if store.capturedArchiveInput.DedupeMarkerID != "018f2a40-0000-7000-9000-0000000000d9" {
		t.Fatalf("dedupe marker = %q", store.capturedArchiveInput.DedupeMarkerID)
	}
}

func TestService_ArchiveTargetJob_MapsNotFoundAndAlreadyArchived(t *testing.T) {
	svc, store := newServiceWithFake(
		"018f2a40-0000-7000-9000-0000000000d8",
		"018f2a40-0000-7000-9000-0000000000d7",
	)
	store.archiveErr = targetjob.ErrTargetJobNotFound
	_, err := svc.ArchiveTargetJob(context.Background(), targetjob.ArchiveRequest{
		UserID:         "user-1",
		TargetJobID:    "target-1",
		IdempotencyKey: "k",
	})
	var apiErr *targetjob.ServiceImportError
	if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeTargetJobNotFound {
		t.Fatalf("expected TARGET_JOB_NOT_FOUND, got %v", err)
	}

	store.archiveErr = targetjob.ErrTargetJobAlreadyArchived
	_, err = svc.ArchiveTargetJob(context.Background(), targetjob.ArchiveRequest{
		UserID:         "user-1",
		TargetJobID:    "target-1",
		IdempotencyKey: "k2",
	})
	if !errors.Is(err, targetjob.ErrTargetJobAlreadyArchived) {
		t.Fatalf("expected ErrTargetJobAlreadyArchived, got %v", err)
	}
}

func TestService_ArchiveTargetJob_RequiresIdempotencyKey(t *testing.T) {
	svc, _ := newServiceWithFake()
	_, err := svc.ArchiveTargetJob(context.Background(), targetjob.ArchiveRequest{
		UserID: "u1", TargetJobID: "t1",
	})
	if !errors.Is(err, targetjob.ErrIdempotencyKeyRequired) {
		t.Fatalf("expected ErrIdempotencyKeyRequired, got %v", err)
	}
}
