package resume

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

var (
	ErrValidationFailed = errors.New("resume validation failed")
	ErrNotFound         = errors.New("resume not found")
	ErrInvalidCursor    = errors.New("invalid resume cursor")
)

type RegisterInput struct {
	UserID         string
	IdempotencyKey string
	SourceType     string
	FileObjectID   string
	RawText        string
	Title          string
	Language       string
}

type RegisterStore interface {
	CreateWithParseJob(ctx context.Context, in resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error)
}

type ReadStore interface {
	Get(ctx context.Context, userID string, resumeID string) (resumestore.ResumeRecord, error)
	List(ctx context.Context, userID string, filter resumestore.ListFilter) (resumestore.ListResult, error)
}

type UpdateStore interface {
	UpdateResume(ctx context.Context, in resumestore.UpdateResumeInput) (resumestore.ResumeRecord, error)
}

type DuplicateStore interface {
	DuplicateResume(ctx context.Context, in resumestore.DuplicateResumeInput) (resumestore.ResumeRecord, error)
}

type TailorRunStore interface {
	CreateTailorRun(ctx context.Context, in resumestore.CreateTailorRunInput) (resumestore.CreateTailorRunResult, error)
	GetTailorRun(ctx context.Context, userID string, tailorRunID string) (resumestore.TailorRunRecord, error)
}

type UploadRegistrar interface {
	RegisterFileObject(ctx context.Context, in uploadservice.RegisterFileObjectInput) (uploadstore.FileObject, error)
}

type ServiceOptions struct {
	Store          RegisterStore
	UploadRegister UploadRegistrar
	Now            func() time.Time
	NewID          func() string
	DedupePepper   string
}

type Service struct {
	store          RegisterStore
	uploadRegister UploadRegistrar
	now            func() time.Time
	newID          func() string
	dedupePepper   string
}

func NewService(opts ServiceOptions) *Service {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	return &Service{
		store:          opts.Store,
		uploadRegister: opts.UploadRegister,
		now:            now,
		newID:          newID,
		dedupePepper:   opts.DedupePepper,
	}
}

func (s *Service) RegisterResume(ctx context.Context, in RegisterInput) (api.ResumeWithJob, error) {
	if s == nil || s.store == nil || s.newID == nil {
		return api.ResumeWithJob{}, fmt.Errorf("resume service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	sourceType := strings.TrimSpace(in.SourceType)
	if userID == "" || strings.TrimSpace(in.IdempotencyKey) == "" || sourceType == "" || strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Language) == "" {
		return api.ResumeWithJob{}, ErrValidationFailed
	}
	now := s.now()
	resumeID := s.newID()
	jobID := s.newID()
	storeIn := resumestore.CreateAssetInput{
		AssetID:        resumeID,
		UserID:         userID,
		JobID:          jobID,
		DedupeKey:      s.dedupeKey(userID, in.IdempotencyKey),
		SourceType:     sourceType,
		Title:          strings.TrimSpace(in.Title),
		Language:       strings.TrimSpace(in.Language),
		RawText:        strings.TrimSpace(in.RawText),
		ParseStatus:    sharedtypes.TargetJobParseStatusQueued,
		JobStatus:      sharedtypes.JobStatusQueued,
		Now:            now,
		RequestPayload: resumestore.RegisterRequestPayload{SourceType: sourceType, Title: strings.TrimSpace(in.Title), Language: strings.TrimSpace(in.Language)},
	}
	switch sourceType {
	case "upload":
		if s.uploadRegister == nil {
			return api.ResumeWithJob{}, fmt.Errorf("resume upload register service is not configured")
		}
		file, err := s.uploadRegister.RegisterFileObject(ctx, uploadservice.RegisterFileObjectInput{
			FileObjectID:    strings.TrimSpace(in.FileObjectID),
			ExpectedPurpose: uploadstore.PurposeResume,
			OwnerUserID:     userID,
		})
		if errors.Is(err, uploadservice.ErrValidationFailed) {
			return api.ResumeWithJob{}, ErrValidationFailed
		}
		if err != nil {
			return api.ResumeWithJob{}, err
		}
		storeIn.FileObjectID = &file.ID
		storeIn.RequestPayload.FileObjectID = strings.TrimSpace(in.FileObjectID)
	case "paste":
		storeIn.RequestPayload.RawTextHash = contentHash(in.RawText)
	default:
		return api.ResumeWithJob{}, ErrValidationFailed
	}
	res, err := s.store.CreateWithParseJob(ctx, storeIn)
	if err != nil {
		return api.ResumeWithJob{}, err
	}
	return api.ResumeWithJob{
		ResumeId: res.AssetID,
		Job: api.Job{
			Id:           res.JobID,
			JobType:      api.JobTypeResumeParse,
			ResourceType: api.ResourceTypeResumeAsset,
			ResourceId:   res.AssetID,
			Status:       res.JobStatus,
			CreatedAt:    res.JobCreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    res.JobUpdatedAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *Service) GetResume(ctx context.Context, userID string, resumeID string) (api.Resume, error) {
	if s == nil {
		return api.Resume{}, fmt.Errorf("resume read store is not configured")
	}
	reader, ok := s.store.(ReadStore)
	if !ok {
		return api.Resume{}, fmt.Errorf("resume read store is not configured")
	}
	rec, err := reader.Get(ctx, strings.TrimSpace(userID), strings.TrimSpace(resumeID))
	if errors.Is(err, resumestore.ErrAssetNotFound) {
		return api.Resume{}, ErrNotFound
	}
	if err != nil {
		return api.Resume{}, err
	}
	return resumeRecordToAPI(rec), nil
}

type ListRequest struct {
	UserID   string
	Cursor   string
	PageSize int
}

func (s *Service) ListResumes(ctx context.Context, in ListRequest) (api.PaginatedResume, error) {
	if s == nil {
		return api.PaginatedResume{}, fmt.Errorf("resume read store is not configured")
	}
	reader, ok := s.store.(ReadStore)
	if !ok {
		return api.PaginatedResume{}, fmt.Errorf("resume read store is not configured")
	}
	res, err := reader.List(ctx, strings.TrimSpace(in.UserID), resumestore.ListFilter{Cursor: in.Cursor, PageSize: in.PageSize})
	if err != nil {
		return api.PaginatedResume{}, err
	}
	out := api.PaginatedResume{Items: make([]api.Resume, 0, len(res.Items))}
	for _, item := range res.Items {
		out.Items = append(out.Items, resumeRecordToAPI(item))
	}
	out.PageInfo = api.PageInfo{
		NextCursor: optionalString(res.NextCursor),
		PageSize:   res.PageSize,
		HasMore:    res.HasMore,
	}
	return out, nil
}

type UpdateResumeRequest struct {
	UserID               string
	ResumeID             string
	DisplayName          *string
	DisplayNameSet       bool
	StructuredProfile    map[string]any
	StructuredProfileSet bool
}

// UpdateResume overwrites the editable fields on an existing resume (D-20 C-17).
func (s *Service) UpdateResume(ctx context.Context, in UpdateResumeRequest) (api.Resume, error) {
	if s == nil {
		return api.Resume{}, fmt.Errorf("resume update store is not configured")
	}
	store, ok := s.store.(UpdateStore)
	if !ok {
		return api.Resume{}, fmt.Errorf("resume update store is not configured")
	}
	userID := strings.TrimSpace(in.UserID)
	resumeID := strings.TrimSpace(in.ResumeID)
	if userID == "" || resumeID == "" {
		return api.Resume{}, ErrValidationFailed
	}
	if !in.StructuredProfileSet && !in.DisplayNameSet {
		return api.Resume{}, ErrValidationFailed
	}
	update := resumestore.UpdateResumeInput{
		UserID:   userID,
		ResumeID: resumeID,
		Now:      s.now(),
	}
	if in.StructuredProfileSet {
		profile := cloneMap(in.StructuredProfile)
		delete(profile, "provenance")
		raw, err := json.Marshal(profile)
		if err != nil {
			return api.Resume{}, ErrValidationFailed
		}
		update.StructuredProfile = raw
		update.StructuredProfileSet = true
	}
	if in.DisplayNameSet {
		if in.DisplayName == nil {
			return api.Resume{}, ErrValidationFailed
		}
		displayName := strings.TrimSpace(*in.DisplayName)
		if displayName == "" {
			return api.Resume{}, ErrValidationFailed
		}
		update.DisplayName = &displayName
	}
	rec, err := store.UpdateResume(ctx, update)
	switch {
	case errors.Is(err, resumestore.ErrAssetNotFound):
		return api.Resume{}, ErrNotFound
	case err != nil:
		return api.Resume{}, err
	}
	return resumeRecordToAPI(rec), nil
}

type DuplicateResumeRequest struct {
	UserID            string
	SourceResumeID    string
	DisplayName       *string
	DisplayNameSet    bool
	StructuredProfile map[string]any
}

// DuplicateResume saves the accepted rewrites as a new resume copied from the
// source read-only snapshot (D-20 C-18).
func (s *Service) DuplicateResume(ctx context.Context, in DuplicateResumeRequest) (api.Resume, error) {
	if s == nil {
		return api.Resume{}, fmt.Errorf("resume duplicate store is not configured")
	}
	store, ok := s.store.(DuplicateStore)
	if !ok {
		return api.Resume{}, fmt.Errorf("resume duplicate store is not configured")
	}
	userID := strings.TrimSpace(in.UserID)
	sourceResumeID := strings.TrimSpace(in.SourceResumeID)
	if userID == "" || sourceResumeID == "" {
		return api.Resume{}, ErrValidationFailed
	}
	duplicate := resumestore.DuplicateResumeInput{
		NewResumeID:    s.newID(),
		UserID:         userID,
		SourceResumeID: sourceResumeID,
		Now:            s.now(),
	}
	if len(in.StructuredProfile) > 0 {
		profile := cloneMap(in.StructuredProfile)
		delete(profile, "provenance")
		raw, err := json.Marshal(profile)
		if err != nil {
			return api.Resume{}, ErrValidationFailed
		}
		duplicate.StructuredProfile = raw
	}
	if in.DisplayNameSet {
		if in.DisplayName == nil {
			return api.Resume{}, ErrValidationFailed
		}
		displayName := strings.TrimSpace(*in.DisplayName)
		if displayName == "" {
			return api.Resume{}, ErrValidationFailed
		}
		duplicate.DisplayName = &displayName
	}
	rec, err := store.DuplicateResume(ctx, duplicate)
	switch {
	case errors.Is(err, resumestore.ErrAssetNotFound):
		return api.Resume{}, ErrNotFound
	case err != nil:
		return api.Resume{}, err
	}
	return resumeRecordToAPI(rec), nil
}

// ArchiveResume marks a resume as archived (soft-hide, not privacy deletion).
// P0 archive has no dedicated persistence column yet (full archive/delete
// integration is plan 003); it enforces ownership / cross-user 404 and returns
// the resume with status archived.
func (s *Service) ArchiveResume(ctx context.Context, userID string, resumeID string) (api.Resume, error) {
	if s == nil {
		return api.Resume{}, fmt.Errorf("resume read store is not configured")
	}
	reader, ok := s.store.(ReadStore)
	if !ok {
		return api.Resume{}, fmt.Errorf("resume read store is not configured")
	}
	rec, err := reader.Get(ctx, strings.TrimSpace(userID), strings.TrimSpace(resumeID))
	if errors.Is(err, resumestore.ErrAssetNotFound) {
		return api.Resume{}, ErrNotFound
	}
	if err != nil {
		return api.Resume{}, err
	}
	out := resumeRecordToAPI(rec)
	archived := "archived"
	out.Status = &archived
	return out, nil
}

type RequestTailorRunInput struct {
	UserID         string
	TargetJobID    string
	ResumeID       string
	Mode           string
	IdempotencyKey string
}

func (s *Service) RequestResumeTailor(ctx context.Context, in RequestTailorRunInput) (api.ResumeTailorRunWithJob, error) {
	if s == nil {
		return api.ResumeTailorRunWithJob{}, fmt.Errorf("resume tailor store is not configured")
	}
	store, ok := s.store.(TailorRunStore)
	if !ok {
		return api.ResumeTailorRunWithJob{}, fmt.Errorf("resume tailor store is not configured")
	}
	userID := strings.TrimSpace(in.UserID)
	targetJobID := strings.TrimSpace(in.TargetJobID)
	resumeID := strings.TrimSpace(in.ResumeID)
	mode := strings.TrimSpace(in.Mode)
	idempotencyKey := strings.TrimSpace(in.IdempotencyKey)
	if userID == "" || resumeID == "" || idempotencyKey == "" {
		return api.ResumeTailorRunWithJob{}, ErrValidationFailed
	}
	switch mode {
	case "gap_review", "bullet_suggestions":
	default:
		return api.ResumeTailorRunWithJob{}, ErrValidationFailed
	}
	now := s.now()
	storeIn := resumestore.CreateTailorRunInput{
		TailorRunID: s.newID(),
		JobID:       s.newID(),
		UserID:      userID,
		TargetJobID: targetJobID,
		ResumeID:    resumeID,
		Mode:        mode,
		DedupeKey:   s.tailorDedupeKey(userID, idempotencyKey),
		Now:         now,
	}
	res, err := store.CreateTailorRun(ctx, storeIn)
	switch {
	case errors.Is(err, resumestore.ErrAssetNotFound), errors.Is(err, resumestore.ErrTailorRunNotFound):
		return api.ResumeTailorRunWithJob{}, ErrNotFound
	case err != nil:
		return api.ResumeTailorRunWithJob{}, err
	}
	jobStatus := res.JobStatus
	if jobStatus == "" {
		jobStatus = sharedtypes.JobStatusQueued
	}
	return api.ResumeTailorRunWithJob{
		TailorRunId: res.TailorRunID,
		Job: api.Job{
			Id:           res.JobID,
			JobType:      api.JobTypeResumeTailor,
			ResourceType: api.ResourceTypeResumeTailorRun,
			ResourceId:   res.TailorRunID,
			Status:       jobStatus,
			CreatedAt:    res.JobCreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:    res.JobUpdatedAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *Service) GetResumeTailorRun(ctx context.Context, userID string, tailorRunID string) (api.ResumeTailorRun, error) {
	if s == nil {
		return api.ResumeTailorRun{}, fmt.Errorf("resume tailor store is not configured")
	}
	store, ok := s.store.(TailorRunStore)
	if !ok {
		return api.ResumeTailorRun{}, fmt.Errorf("resume tailor store is not configured")
	}
	rec, err := store.GetTailorRun(ctx, strings.TrimSpace(userID), strings.TrimSpace(tailorRunID))
	switch {
	case errors.Is(err, resumestore.ErrTailorRunNotFound):
		return api.ResumeTailorRun{}, ErrNotFound
	case err != nil:
		return api.ResumeTailorRun{}, err
	}
	return tailorRunRecordToAPI(rec), nil
}

func (s *Service) dedupeKey(userID, idempotencyKey string) string {
	return s.namespacedDedupeKey("resume.register.v1", userID, idempotencyKey)
}

func (s *Service) tailorDedupeKey(userID, idempotencyKey string) string {
	return s.namespacedDedupeKey("resume.tailor.v1", userID, idempotencyKey)
}

func (s *Service) namespacedDedupeKey(namespace, userID, idempotencyKey string) string {
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
	return hex.EncodeToString(h.Sum(nil))
}

func contentHash(in string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(in)))
	return hex.EncodeToString(h[:])
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func resumeRecordToAPI(rec resumestore.ResumeRecord) api.Resume {
	status := "active"
	out := api.Resume{
		Id:                rec.ID,
		Title:             rec.Title,
		Language:          rec.Language,
		ParseStatus:       rec.ParseStatus,
		Status:            &status,
		FileObjectId:      cloneStringPtr(rec.FileObjectID),
		OriginalText:      cloneStringPtr(rec.OriginalText),
		SourceType:        cloneStringPtr(rec.SourceType),
		CreatedAt:         rec.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         rec.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:         timePtrToString(rec.DeletedAt),
		ParsedSummary:     rawJSONMapPtr(rec.ParsedSummary),
		StructuredProfile: rawJSONAnyOrNil(rec.StructuredProfile),
	}
	if rec.DisplayName != nil {
		out.DisplayName = *rec.DisplayName
	}
	if rec.ParsedTextSnapshot != nil {
		out.ParsedTextSnapshot = cloneStringPtr(rec.ParsedTextSnapshot)
	}
	return out
}

func tailorRunRecordToAPI(rec resumestore.TailorRunRecord) api.ResumeTailorRun {
	out := api.ResumeTailorRun{
		Id:          rec.ID,
		Status:      rec.Status,
		ResumeId:    rec.ResumeID,
		Suggestions: decodeTailorSuggestions(rec.Suggestions),
		CreatedAt:   rec.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   rec.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if strings.TrimSpace(rec.TargetJobID) != "" {
		out.TargetJobId = pointerString(rec.TargetJobID)
	}
	if rec.Status == "ready" {
		if matchSummary := decodeTailorMatchSummary(rec.MatchSummary); matchSummary != nil {
			out.MatchSummary = matchSummary
		}
		if rec.Provenance.PromptVersion != "" || rec.Provenance.ModelID != "" {
			out.Provenance = &api.GenerationProvenance{
				PromptVersion:     rec.Provenance.PromptVersion,
				RubricVersion:     rec.Provenance.RubricVersion,
				ModelId:           rec.Provenance.ModelID,
				Language:          rec.Provenance.Language,
				FeatureFlag:       rec.Provenance.FeatureFlag,
				DataSourceVersion: rec.Provenance.DataSourceVersion,
			}
		}
	}
	if out.Suggestions == nil {
		out.Suggestions = []api.ResumeTailorBulletSuggestion{}
	}
	return out
}

func decodeTailorMatchSummary(raw json.RawMessage) *api.ResumeTailorMatchSummary {
	if len(raw) == 0 || string(raw) == "{}" {
		return nil
	}
	var out api.ResumeTailorMatchSummary
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	if len(out.Strengths) == 0 && len(out.Gaps) == 0 {
		return nil
	}
	return &out
}

func decodeTailorSuggestions(raw json.RawMessage) []api.ResumeTailorBulletSuggestion {
	if len(raw) == 0 {
		return []api.ResumeTailorBulletSuggestion{}
	}
	var out []api.ResumeTailorBulletSuggestion
	if err := json.Unmarshal(raw, &out); err != nil {
		return []api.ResumeTailorBulletSuggestion{}
	}
	if out == nil {
		return []api.ResumeTailorBulletSuggestion{}
	}
	return out
}

func pointerString(in string) *string {
	v := in
	return &v
}

func rawJSONAnyOrNil(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "{}" || string(raw) == "null" {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func rawJSONMapPtr(raw json.RawMessage) *map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil || len(out) == 0 {
		return nil
	}
	return &out
}

func cloneStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func timePtrToString(in *time.Time) *string {
	if in == nil {
		return nil
	}
	v := in.UTC().Format(time.RFC3339)
	return &v
}

func optionalString(in string) *string {
	if strings.TrimSpace(in) == "" {
		return nil
	}
	v := in
	return &v
}
