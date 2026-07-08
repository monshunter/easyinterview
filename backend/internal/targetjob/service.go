package targetjob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// ErrIdempotencyKeyRequired is returned when the handler did not pass an
// `Idempotency-Key` header. Per spec D-6 the four TargetJob mutating
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
// synchronous import path: it decodes the B2 oneOf source variant, derives
// a user-scoped dedupe key, validates inputs, and hands a fully-shaped
// ImportTargetJobInput to the Store. The runner side, F3+A3 calls, and
// outbox publish flow are layered in Phase 3 / Phase 4.
type Service struct {
	store        Store
	newID        IDGenerator
	now          func() time.Time
	dedupePepper string
}

// IDGenerator returns a UUIDv7-shaped string. The service uses this for
// every ID it persists so tests can deterministically inject IDs.
type IDGenerator func() string

// ServiceOptions wires the Service constructor.
type ServiceOptions struct {
	Store        Store
	NewID        IDGenerator
	Now          func() time.Time
	DedupePepper string
}

// NewService constructs a Service. Now defaults to time.Now().UTC().
func NewService(opts ServiceOptions) *Service {
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		store:        opts.Store,
		newID:        opts.NewID,
		now:          opts.Now,
		dedupePepper: opts.DedupePepper,
	}
}

// ImportRequest is the service-layer command produced by the handler after
// it parses an `importTargetJob` request body and headers.
type ImportRequest struct {
	UserID          string
	IdempotencyKey  string
	TargetLanguage  string
	ResumeID        string
	TitleHint       string
	CompanyNameHint string
	Source          any // B2 TargetJobImportSource oneOf, decoded as map[string]any
}

// ImportResponse is what the handler renders into the generated
// TargetJobWithJob HTTP envelope.
type ImportResponse struct {
	TargetJobID string
	Job         api.Job
}

// ImportTargetJob performs the synchronous portion of `POST /targets/import`:
// decoding the source variant, deriving the dedupe key, validating inputs,
// building the runner-bound or manual_form store input, and translating the
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

	decoded, err := decodeImportSource(in.Source)
	if err != nil {
		return ImportResponse{}, err
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
		APISourceType:          decoded.SourceType,
		InitialLifecycleStatus: sharedtypes.TargetJobStatusDraft,
		Title:                  pickHint(in.TitleHint, decoded.Title),
		CompanyName:            pickHint(in.CompanyNameHint, decoded.CompanyName),
		JobID:                  jobID,
		Now:                    now,
	}

	switch decoded.SourceType {
	case SourceTypeURL:
		sanitized, err := sanitizeJDURL(decoded.URL)
		if err != nil {
			return ImportResponse{}, err
		}
		storeIn.InitialAnalysisStatus = sharedtypes.TargetJobParseStatusQueued
		storeIn.SourceURL = sanitized
		storeIn.SourceID = s.newID()
		// Phase 4 drainer fills SourceSnapshotText after fetch.
		if err := s.attachRunnerEnvelope(&storeIn); err != nil {
			return ImportResponse{}, err
		}
	case SourceTypeManualText:
		storeIn.InitialAnalysisStatus = sharedtypes.TargetJobParseStatusQueued
		storeIn.RawJDText = decoded.RawText
		storeIn.SourceID = s.newID()
		storeIn.SourceSnapshotText = decoded.RawText
		if err := s.attachRunnerEnvelope(&storeIn); err != nil {
			return ImportResponse{}, err
		}
	case SourceTypeFile:
		// Spec D-9 / plan 3.2: confirm the referenced upload belongs to the
		// caller and was uploaded with the target_job_attachment purpose.
		// Cross-user / soft-deleted IDs surface as TARGET_JOB_NOT_FOUND so
		// the handler returns 404 without leaking existence; mismatched
		// purpose surfaces as TARGET_IMPORT_SOURCE_INVALID.
		attachment, err := s.store.LookupFileAttachmentForUser(ctx, in.UserID, decoded.FileObjectID)
		if err != nil {
			if errors.Is(err, ErrTargetJobNotFound) {
				return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeTargetJobNotFound, Message: "file attachment not found"}
			}
			return ImportResponse{}, err
		}
		if attachment.Purpose != "target_job_attachment" {
			return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "file attachment purpose is not target_job_attachment"}
		}
		storeIn.InitialAnalysisStatus = sharedtypes.TargetJobParseStatusQueued
		storeIn.SourceFileObjectID = decoded.FileObjectID
		storeIn.SourceID = s.newID()
		if err := s.attachRunnerEnvelope(&storeIn); err != nil {
			return ImportResponse{}, err
		}
	case SourceTypeManualForm:
		storeIn.InitialAnalysisStatus = sharedtypes.TargetJobParseStatusReady
		storeIn.RawJDText = decoded.RawDescription
		// Spec D-11 / plan 3.1: at least one draft must_have requirement.
		storeIn.DraftRequirements = []RequirementRecord{
			{
				ID:           s.newID(),
				Kind:         RequirementMustHave,
				Label:        defaultDraftRequirementLabel(decoded.RawDescription),
				DisplayOrder: 1,
			},
		}
		// no SourceID, no OutboxEventID -> store treats as manual_form path
	default:
		return ImportResponse{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: fmt.Sprintf("unsupported source type %q", decoded.SourceType)}
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
		APISourceType:  in.APISourceType,
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
		SourceType:     string(in.APISourceType),
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

// decodedSource is the union of typed B2 source variants after JSON
// unmarshalling. Only the fields relevant to the surfaced source variant
// are populated; the rest stay zero.
type decodedSource struct {
	SourceType     SourceType
	URL            string
	RawText        string
	FileObjectID   string
	Title          string
	CompanyName    string
	RawDescription string
}

func decodeImportSource(raw any) (decodedSource, error) {
	if raw == nil {
		return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "source is required"}
	}
	rawMap, ok := raw.(map[string]any)
	if !ok {
		// Try to round-trip via JSON for robustness against non-map shapes.
		buf, err := json.Marshal(raw)
		if err != nil {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "source must be an object"}
		}
		if err := json.Unmarshal(buf, &rawMap); err != nil || rawMap == nil {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeValidationFailed, Message: "source must be an object"}
		}
	}
	typeStr, _ := rawMap["type"].(string)
	switch SourceType(typeStr) {
	case SourceTypeURL:
		urlStr, _ := rawMap["url"].(string)
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "url is required for source type=url"}
		}
		return decodedSource{SourceType: SourceTypeURL, URL: urlStr}, nil
	case SourceTypeManualText:
		text, _ := rawMap["rawText"].(string)
		text = strings.TrimSpace(text)
		if text == "" {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "rawText is required for source type=manual_text"}
		}
		return decodedSource{SourceType: SourceTypeManualText, RawText: text}, nil
	case SourceTypeFile:
		fileObjectID, _ := rawMap["fileObjectId"].(string)
		fileObjectID = strings.TrimSpace(fileObjectID)
		if fileObjectID == "" {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "fileObjectId is required for source type=file"}
		}
		return decodedSource{SourceType: SourceTypeFile, FileObjectID: fileObjectID}, nil
	case SourceTypeManualForm:
		title, _ := rawMap["title"].(string)
		title = strings.TrimSpace(title)
		if title == "" {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "title is required for source type=manual_form"}
		}
		rawDesc, _ := rawMap["rawDescription"].(string)
		rawDesc = strings.TrimSpace(rawDesc)
		if rawDesc == "" {
			return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "rawDescription is required for source type=manual_form"}
		}
		companyPtr, _ := rawMap["companyName"].(string)
		return decodedSource{
			SourceType:     SourceTypeManualForm,
			Title:          title,
			CompanyName:    strings.TrimSpace(companyPtr),
			RawDescription: rawDesc,
		}, nil
	default:
		return decodedSource{}, &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: fmt.Sprintf("unsupported source type %q", typeStr)}
	}
}

// sanitizeJDURL performs the synchronous-stage URL hygiene checks: scheme
// must be https, no fragment, query string is preserved structurally but
// snapshot recording (Phase 3.3) strips secrets. Full SSRF guard (DNS,
// private-network rejection, redirect inspection) is the Phase 3.3 fetcher's
// responsibility — this helper only rejects the unmistakable up-front
// violations so the runner does not spin up on bad input.
func sanitizeJDURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "url is malformed"}
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return "", &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "url scheme must be https"}
	}
	if u.Host == "" {
		return "", &ServiceImportError{Code: sharederrors.CodeTargetImportSourceInvalid, Message: "url host is required"}
	}
	u.User = nil
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	return u.String(), nil
}

func pickHint(hint, decoded string) string {
	if strings.TrimSpace(hint) != "" {
		return strings.TrimSpace(hint)
	}
	return strings.TrimSpace(decoded)
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
	rec, reqs, _, err := s.store.GetTargetJobByUser(ctx, userID, targetJobID)
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
	_, reqs, _, err := s.store.GetTargetJobByUser(ctx, in.UserID, in.TargetJobID)
	if err != nil {
		// updated successfully, but failed to reload requirements — return
		// the updated row without requirements rather than fail the call.
		return recordToAPI(updated, nil), nil
	}
	return recordToAPI(updated, reqs), nil
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
	out := api.TargetJob{
		AnalysisStatus:         rec.AnalysisStatus,
		CompanyName:            rec.CompanyName,
		CreatedAt:              rec.CreatedAt.UTC().Format(time.RFC3339),
		CurrentPracticePlanId:  optionalString(rec.CurrentPracticePlanID),
		Id:                     rec.ID,
		LocationText:           optionalString(rec.LocationText),
		OpenQuestionIssueCount: rec.OpenQuestionIssueCount,
		Requirements:           []api.TargetJobRequirement{},
		ResumeId:               optionalString(rec.ResumeID),
		SourceType:             string(rec.SourceType),
		SourceUrl:              optionalString(rec.SourceURL),
		Status:                 rec.Status,
		Summary:                decodeTargetJobSummary(rec.Summary),
		FitSummary:             decodeTargetJobFitSummary(rec.FitSummary),
		TargetLanguage:         rec.TargetLanguage,
		Title:                  rec.Title,
		UpdatedAt:              rec.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if rec.LatestReportID != "" {
		v := rec.LatestReportID
		out.LatestReportId = &v
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

// defaultDraftRequirementLabel synthesises the Phase 2.1 manual_form draft
// requirement label. Phase 3.1 may refine this; for now we trim a leading
// description line and fall back to a generic placeholder.
func defaultDraftRequirementLabel(rawDescription string) string {
	for _, line := range strings.Split(rawDescription, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			if len(t) > 120 {
				t = t[:120]
			}
			return t
		}
	}
	return "Manual draft requirement"
}
