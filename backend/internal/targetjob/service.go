package targetjob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	platformconfig "github.com/monshunter/easyinterview/backend/internal/platform/config"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// ErrIdempotencyKeyRequired is returned when the handler did not pass an
// `Idempotency-Key` header. Per spec D-6 the TargetJob mutating
// operations all require this header.
var ErrIdempotencyKeyRequired = errors.New("idempotency key required")

// ServiceImportError wraps a B1 error code so the handler can map it to the
// generated ApiError envelope without leaking internal SQL / provider
// details. The handler decides the HTTP status from ErrorCode.
type ServiceImportError struct {
	Code    string
	Message string
}

func (e *ServiceImportError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

// Service is the handler-facing TargetJob orchestrator. Phase 2.1 lands the
// synchronous import path: it validates the paste-only request, derives a
// user-scoped dedupe key, and hands a fully-shaped
// ImportTargetJobInput to the Store. The runner side, F3+A3 calls, and
// outbox publish flow are layered in Phase 3 / Phase 4.
type Service struct {
	store           Store
	newID           IDGenerator
	now             func() time.Time
	dedupePepper    string
	maxRawTextBytes int64
}

// IDGenerator returns a UUIDv7-shaped string. The service uses this for
// every ID it persists so tests can deterministically inject IDs.
type IDGenerator func() string

// ServiceOptions wires the Service constructor.
type ServiceOptions struct {
	Store           Store
	NewID           IDGenerator
	Now             func() time.Time
	DedupePepper    string
	MaxRawTextBytes int64
}

// NewService constructs a Service. Now defaults to time.Now().UTC().
func NewService(opts ServiceOptions) *Service {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.MaxRawTextBytes <= 0 {
		opts.MaxRawTextBytes = platformconfig.DefaultContentLimits().TargetJobMaxRawTextBytes
	}
	return &Service{
		store:           opts.Store,
		newID:           opts.NewID,
		now:             opts.Now,
		dedupePepper:    opts.DedupePepper,
		maxRawTextBytes: opts.MaxRawTextBytes,
	}
}

// ImportRequest is the service-layer command produced by the handler after
// it parses an `importTargetJob` request body and headers.
type ImportRequest struct {
	UserID         string
	IdempotencyKey string
	TargetLanguage string
	ResumeID       string
	RawText        string
}

// ImportResponse is what the handler renders into the generated
// TargetJobWithJob HTTP envelope.
type ImportResponse struct {
	TargetJobID string
	Job         api.Job
}

// ImportTargetJob performs the synchronous portion of `POST /targets/import`:
// validating the paste-only request, deriving the dedupe key, building the
// runner-bound store input, and translating the
// store result back into a generated `Job` shape.
func (s *Service) ImportTargetJob(ctx context.Context, in ImportRequest) (ImportResponse, error) {
	if s == nil || s.store == nil || s.newID == nil {
		return ImportResponse{}, fmt.Errorf("targetjob service is not initialised")
	}
	if in.UserID == "" {
		return ImportResponse{}, fmt.Errorf("userId is required")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return ImportResponse{}, ErrIdempotencyKeyRequired
	}
	if strings.TrimSpace(in.TargetLanguage) == "" {
		return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "targetLanguage is required"}
	}
	resumeID := strings.TrimSpace(in.ResumeID)
	if resumeID == "" {
		return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "resumeId is required"}
	}

	rawText := strings.TrimSpace(in.RawText)
	if rawText == "" {
		return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "rawText is required"}
	}
	if int64(len(rawText)) > s.maxRawTextBytes {
		return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "rawText is too large"}
	}

	now := s.now()
	targetJobID := s.newID()
	jobID := s.newID()
	dedupeKey := s.importDedupeKey(in.UserID, in.IdempotencyKey)

	storeIn := ImportTargetJobInput{
		UserID:                 in.UserID,
		DedupeKey:              dedupeKey,
		TargetJobID:            targetJobID,
		TargetLanguage:         strings.TrimSpace(in.TargetLanguage),
		ResumeID:               resumeID,
		RawJDText:              rawText,
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		InitialAnalysisStatus:  sharedtypes.TargetJobParseStatusQueued,
		JobID:                  jobID,
		Now:                    now,
	}
	if err := s.attachRunnerEnvelope(&storeIn); err != nil {
		return ImportResponse{}, err
	}

	res, err := s.store.ImportTargetJob(ctx, storeIn)
	if err != nil {
		if errors.Is(err, ErrTargetJobNotFound) {
			return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeTargetJobNotFound, Message: "resume not found"}
		}
		return ImportResponse{}, err
	}

	return ImportResponse{
		TargetJobID: res.TargetJobID,
		Job: api.Job{
			Id:           res.JobID,
			JobType:      api.JobTypeTargetImport,
			ResourceType: api.ResourceTypeTargetJob,
			ResourceId:   res.TargetJobID,
			Status:       res.JobStatus,
			CreatedAt:    res.JobCreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    res.JobUpdatedAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *Service) attachRunnerEnvelope(in *ImportTargetJobInput) error {
	in.OutboxEventID = s.newID()
	outboxPayload, err := BuildTargetImportRequestedPayload(TargetImportRequestedInput{
		TargetJobID:    in.TargetJobID,
		TargetLanguage: in.TargetLanguage,
		UserID:         in.UserID,
	})
	if err != nil {
		return err
	}
	rawOutbox, err := json.Marshal(outboxPayload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}
	in.OutboxEventPayload = rawOutbox

	jobPayload, err := BuildTargetImportJobPayload(TargetImportJobPayload{
		TargetJobID:    in.TargetJobID,
		UserID:         in.UserID,
		TargetLanguage: in.TargetLanguage,
	})
	if err != nil {
		return err
	}
	in.JobPayload = jobPayload
	return nil
}

func (s *Service) importDedupeKey(userID, idempotencyKey string) string {
	return s.dedupeKey("targetjob.import.v1", userID, idempotencyKey)
}

func (s *Service) updateDedupeKey(userID, idempotencyKey string) string {
	return s.dedupeKey("targetjob.update.v1", userID, idempotencyKey)
}

func (s *Service) archiveDedupeKey(userID, idempotencyKey string) string {
	return s.dedupeKey("targetjob.archive.v1", userID, idempotencyKey)
}

func (s *Service) dedupeKey(namespace, userID, idempotencyKey string) string {
	h := sha256.New()
	h.Write([]byte(namespace))
	if s.dedupePepper != "" {
		h.Write([]byte("|"))
		h.Write([]byte(s.dedupePepper))
	}
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(userID)))
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(idempotencyKey)))
	return "target_import:" + hex.EncodeToString(h.Sum(nil))
}

// ListRequest is the service-layer command produced by the handler from the
// `listTargetJobs` query string.
type ListRequest struct {
	UserID         string
	Status         *sharedtypes.TargetJobStatus
	AnalysisStatus *sharedtypes.TargetJobParseStatus
	SearchQuery    string
	Cursor         string
	PageSize       int32
}

// ListTargetJobs returns one page of TargetJob rows for the user. Cursor and
// pageSize handling delegate to the store; the service maps results to the
// generated paginated wire shape.
func (s *Service) ListTargetJobs(ctx context.Context, in ListRequest) (api.PaginatedTargetJob, error) {
	if in.UserID == "" {
		return api.PaginatedTargetJob{}, fmt.Errorf("userId is required")
	}
	res, err := s.store.ListTargetJobsForUser(ctx, in.UserID, ListFilter{
		Status:         in.Status,
		AnalysisStatus: in.AnalysisStatus,
		SearchQuery:    in.SearchQuery,
		Cursor:         in.Cursor,
		PageSize:       in.PageSize,
	})
	if err != nil {
		return api.PaginatedTargetJob{}, err
	}
	out := api.PaginatedTargetJob{Items: make([]api.TargetJob, 0, len(res.Items))}
	for _, r := range res.Items {
		// Phase 2.2 returns rows without requirements (those come from the
		// detail endpoint); generated `TargetJob.Requirements` is required
		// so we surface an empty array rather than nil.
		out.Items = append(out.Items, recordToAPI(r, nil))
	}
	out.PageInfo = api.PageInfo{
		NextCursor: optionalString(res.NextCursor),
		PageSize:   effectiveListPageSize(in.PageSize),
		HasMore:    res.HasMore,
	}
	return out, nil
}

func effectiveListPageSize(pageSize int32) int {
	if pageSize <= 0 {
		return sharedtypes.DefaultPageSize
	}
	if pageSize > ListMaxPageSize {
		return ListMaxPageSize
	}
	return int(pageSize)
}

// GetTargetJob returns the full detail view for a single TargetJob.
func (s *Service) GetTargetJob(ctx context.Context, userID, targetJobID string) (api.TargetJob, error) {
	if userID == "" || targetJobID == "" {
		return api.TargetJob{}, fmt.Errorf("userId and targetJobId are required")
	}
	rec, reqs, err := s.store.GetTargetJobByUser(ctx, userID, targetJobID)
	if err != nil {
		if errors.Is(err, ErrTargetJobNotFound) {
			return api.TargetJob{}, &ServiceImportError{Code: sharederrors.CodeTargetJobNotFound, Message: "target job not found"}
		}
		return api.TargetJob{}, err
	}
	return recordToAPI(rec, reqs), nil
}

// UpdateRequest is the service-layer command for `updateTargetJob`. The
// handler validates `Idempotency-Key` is present before calling.
type UpdateRequest struct {
	UserID          string
	TargetJobID     string
	IdempotencyKey  string
	Status          *sharedtypes.TargetJobStatus
	LocationText    *string
	Notes           *string
	TitleHint       *string
	CompanyNameHint *string
}

// UpdateTargetJob applies the requested mutation via the store. Spec D-6 /
// plan 2.4 require an Idempotency-Key header; the handler enforces
// presence, and the SQL store validates status transitions inside the update
// transaction after locking the current row.
func (s *Service) UpdateTargetJob(ctx context.Context, in UpdateRequest) (api.TargetJob, error) {
	if in.UserID == "" || in.TargetJobID == "" {
		return api.TargetJob{}, fmt.Errorf("userId and targetJobId are required")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return api.TargetJob{}, ErrIdempotencyKeyRequired
	}
	dedupeKey := s.updateDedupeKey(in.UserID, in.IdempotencyKey)
	if rec, reqs, hit, err := s.store.LookupUpdateDedupe(ctx, in.UserID, dedupeKey); err != nil {
		return api.TargetJob{}, err
	} else if hit {
		return recordToAPI(rec, reqs), nil
	}
	updated, err := s.store.UpdateTargetJobLifecycle(ctx, in.UserID, in.TargetJobID, UpdateLifecycleFields{
		Status:          in.Status,
		LocationText:    in.LocationText,
		Notes:           in.Notes,
		TitleHint:       in.TitleHint,
		CompanyNameHint: in.CompanyNameHint,
		DedupeKey:       dedupeKey,
		DedupeMarkerID:  s.newID(),
	}, s.now())
	if err != nil {
		if errors.Is(err, ErrTargetJobNotFound) {
			return api.TargetJob{}, &ServiceImportError{Code: sharederrors.CodeTargetJobNotFound, Message: "target job not found"}
		}
		return api.TargetJob{}, err
	}
	reloaded, reqs, err := s.store.GetTargetJobByUser(ctx, in.UserID, in.TargetJobID)
	if err != nil {
		// The mutation succeeded, but the read-side reload failed. Return the
		// locked update row without derived facts rather than invent progress.
		return recordToAPI(updated, nil), nil
	}
	// The transaction-returned row is authoritative for the committed
	// mutation. Attach only read-side facts from the reload so a stale test
	// double/read replica cannot overwrite the just-committed lifecycle fields.
	updated.PracticeFactsLoaded = reloaded.PracticeFactsLoaded
	updated.CompletedRoundFacts = append([]PracticeRoundFact(nil), reloaded.CompletedRoundFacts...)
	updated.ReadyPlanFacts = append([]ReadyPracticePlanFact(nil), reloaded.ReadyPlanFacts...)
	return recordToAPI(updated, reqs), nil
}

// ArchiveRequest is the service-layer command for `archiveTargetJob`.
type ArchiveRequest struct {
	UserID         string
	TargetJobID    string
	IdempotencyKey string
}

// ArchiveTargetJob soft-hides a TargetJob for the current user.
func (s *Service) ArchiveTargetJob(ctx context.Context, in ArchiveRequest) (api.TargetJob, error) {
	if in.UserID == "" || in.TargetJobID == "" {
		return api.TargetJob{}, fmt.Errorf("userId and targetJobId are required")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return api.TargetJob{}, ErrIdempotencyKeyRequired
	}
	rec, err := s.store.ArchiveTargetJob(ctx, ArchiveTargetJobInput{
		UserID:         in.UserID,
		TargetJobID:    in.TargetJobID,
		DedupeKey:      s.archiveDedupeKey(in.UserID, in.IdempotencyKey),
		DedupeMarkerID: s.newID(),
		Now:            s.now(),
	})
	if err != nil {
		if errors.Is(err, ErrTargetJobNotFound) {
			return api.TargetJob{}, &ServiceImportError{Code: sharederrors.CodeTargetJobNotFound, Message: "target job not found"}
		}
		return api.TargetJob{}, err
	}
	return recordToAPI(rec, nil), nil
}

// allowedLifecycleTransitions captures spec §3.1 D-* state-machine rules.
// Each entry maps a current state to the set of states it may transition
// into. Archived is always permitted from any non-terminal state.
var allowedLifecycleTransitions = map[sharedtypes.TargetJobStatus]map[sharedtypes.TargetJobStatus]struct{}{
	sharedtypes.TargetJobStatusDraft: {
		sharedtypes.TargetJobStatusPreparing: {},
		sharedtypes.TargetJobStatusArchived:  {},
	},
	sharedtypes.TargetJobStatusPreparing: {
		sharedtypes.TargetJobStatusApplied:  {},
		sharedtypes.TargetJobStatusArchived: {},
	},
	sharedtypes.TargetJobStatusApplied: {
		sharedtypes.TargetJobStatusInterviewing: {},
		sharedtypes.TargetJobStatusArchived:     {},
	},
	sharedtypes.TargetJobStatusInterviewing: {
		sharedtypes.TargetJobStatusOffer:    {},
		sharedtypes.TargetJobStatusRejected: {},
		sharedtypes.TargetJobStatusArchived: {},
	},
	sharedtypes.TargetJobStatusOffer: {
		sharedtypes.TargetJobStatusArchived: {},
	},
	sharedtypes.TargetJobStatusRejected: {
		sharedtypes.TargetJobStatusArchived: {},
	},
	sharedtypes.TargetJobStatusArchived: {},
}

func validateLifecycleTransition(current, next sharedtypes.TargetJobStatus) error {
	if current == next {
		return nil // idempotent same-state update
	}
	allowed, ok := allowedLifecycleTransitions[current]
	if !ok {
		return &ServiceImportError{Code: sharederrors.CodeTargetInvalidStateTransition, Message: fmt.Sprintf("unknown current status %q", current)}
	}
	if _, ok := allowed[next]; !ok {
		return &ServiceImportError{Code: sharederrors.CodeTargetInvalidStateTransition, Message: fmt.Sprintf("transition %s -> %s is not allowed", current, next)}
	}
	return nil
}

func recordToAPI(rec TargetJobRecord, reqs []RequirementRecord) api.TargetJob {
	summary := decodeTargetJobSummary(rec.Summary)
	practiceProgress, currentPracticePlanID := projectPracticeProgress(rec, summary)
	out := api.TargetJob{
		AnalysisStatus:         rec.AnalysisStatus,
		CompanyName:            rec.CompanyName,
		CreatedAt:              rec.CreatedAt.UTC().Format(time.RFC3339),
		CurrentPracticePlanId:  currentPracticePlanID,
		Id:                     rec.ID,
		LocationText:           optionalString(rec.LocationText),
		OpenQuestionIssueCount: rec.OpenQuestionIssueCount,
		PracticeProgress:       practiceProgress,
		Requirements:           []api.TargetJobRequirement{},
		ResumeId:               optionalString(rec.ResumeID),
		Status:                 rec.Status,
		Summary:                summary,
		FitSummary:             decodeTargetJobFitSummary(rec.FitSummary),
		TargetLanguage:         rec.TargetLanguage,
		Title:                  rec.Title,
		UpdatedAt:              rec.UpdatedAt.UTC().Format(time.RFC3339),
	}
	for _, r := range reqs {
		out.Requirements = append(out.Requirements, api.TargetJobRequirement{
			Id:            r.ID,
			Kind:          string(r.Kind),
			Label:         r.Label,
			EvidenceLevel: string(r.EvidenceLevel),
		})
	}
	return out
}

type practiceRoundKey struct {
	id       string
	sequence int32
}

var canonicalPracticeRoundTypes = map[string]struct{}{
	"hr":               {},
	"technical":        {},
	"manager":          {},
	"cross_functional": {},
	"culture":          {},
	"final":            {},
	"other":            {},
}

func projectPracticeProgress(rec TargetJobRecord, summary *api.TargetJobSummary) (*api.PracticeProgress, *string) {
	if !rec.PracticeFactsLoaded {
		return nil, nil
	}
	canonicalRounds, ok := canonicalPracticeRounds(summary)
	if !ok {
		return nil, nil
	}

	canonicalByKey := make(map[practiceRoundKey]struct{}, len(canonicalRounds))
	for _, round := range canonicalRounds {
		canonicalByKey[practiceRoundKey{id: round.RoundId, sequence: round.RoundSequence}] = struct{}{}
	}
	completed := make(map[practiceRoundKey]struct{}, len(rec.CompletedRoundFacts))
	for _, fact := range rec.CompletedRoundFacts {
		key := practiceRoundKey{id: fact.RoundID, sequence: fact.RoundSequence}
		if _, exists := canonicalByKey[key]; exists {
			completed[key] = struct{}{}
		}
	}

	progress := &api.PracticeProgress{
		CompletedRounds: make([]api.PracticeRoundRef, 0, len(completed)),
		Status:          "not_started",
	}
	for i := range canonicalRounds {
		round := canonicalRounds[i]
		key := practiceRoundKey{id: round.RoundId, sequence: round.RoundSequence}
		if _, done := completed[key]; done {
			progress.CompletedRounds = append(progress.CompletedRounds, round)
			continue
		}
		current := round
		progress.CurrentRound = &current
		break
	}

	switch {
	case len(progress.CompletedRounds) == len(canonicalRounds):
		progress.Status = "completed"
		progress.CurrentRound = nil
	case len(progress.CompletedRounds) > 0:
		progress.Status = "in_progress"
	}

	if progress.CurrentRound == nil {
		return progress, nil
	}
	currentKey := practiceRoundKey{
		id:       progress.CurrentRound.RoundId,
		sequence: progress.CurrentRound.RoundSequence,
	}
	var selected *ReadyPracticePlanFact
	for i := range rec.ReadyPlanFacts {
		candidate := &rec.ReadyPlanFacts[i]
		if candidate.PlanID == "" || (practiceRoundKey{id: candidate.RoundID, sequence: candidate.RoundSequence}) != currentKey {
			continue
		}
		if selected == nil || candidate.CreatedAt.After(selected.CreatedAt) ||
			(candidate.CreatedAt.Equal(selected.CreatedAt) && candidate.PlanID > selected.PlanID) {
			selected = candidate
		}
	}
	if selected == nil {
		return progress, nil
	}
	planID := selected.PlanID
	return progress, &planID
}

func canonicalPracticeRounds(summary *api.TargetJobSummary) ([]api.PracticeRoundRef, bool) {
	if summary == nil || len(summary.InterviewRounds) < 2 || len(summary.InterviewRounds) > 5 {
		return nil, false
	}
	rounds := append([]api.TargetJobInterviewRound(nil), summary.InterviewRounds...)
	sort.Slice(rounds, func(i, j int) bool {
		return rounds[i].Sequence < rounds[j].Sequence
	})
	canonical := make([]api.PracticeRoundRef, 0, len(rounds))
	var previousSequence int32
	for i, round := range rounds {
		if round.Sequence <= 0 || (i > 0 && round.Sequence <= previousSequence) {
			return nil, false
		}
		roundType := strings.TrimSpace(round.Type)
		if _, allowed := canonicalPracticeRoundTypes[roundType]; !allowed {
			return nil, false
		}
		if round.DurationMinutes < 10 || round.DurationMinutes > 180 {
			return nil, false
		}
		if strings.TrimSpace(round.Name) == "" || strings.TrimSpace(round.Focus) == "" {
			return nil, false
		}
		canonical = append(canonical, api.PracticeRoundRef{
			RoundId:       fmt.Sprintf("round-%d-%s", round.Sequence, roundType),
			RoundSequence: round.Sequence,
		})
		previousSequence = round.Sequence
	}
	return canonical, true
}

func decodeTargetJobSummary(raw json.RawMessage) *api.TargetJobSummary {
	if !hasMaterializedJSON(raw) {
		return nil
	}
	var out api.TargetJobSummary
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	if !hasGenerationProvenance(out.Provenance) {
		return nil
	}
	return &out
}

func decodeTargetJobFitSummary(raw json.RawMessage) *api.TargetJobFitSummary {
	if !hasMaterializedJSON(raw) {
		return nil
	}
	var out api.TargetJobFitSummary
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	if !hasGenerationProvenance(out.Provenance) {
		return nil
	}
	return &out
}

func hasMaterializedJSON(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && trimmed != "{}" && trimmed != "null"
}

func hasGenerationProvenance(p api.GenerationProvenance) bool {
	return p.PromptVersion != "" &&
		p.RubricVersion != "" &&
		p.ModelId != "" &&
		p.Language != "" &&
		p.DataSourceVersion != ""
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
