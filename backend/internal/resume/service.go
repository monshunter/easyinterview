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
	ErrValidationFailed              = errors.New("resume validation failed")
	ErrNotFound                      = errors.New("resume asset not found")
	ErrInvalidCursor                 = errors.New("invalid resume cursor")
	ErrAssetParseNotReady            = errors.New("resume asset parse is not ready")
	ErrStructuredMasterAlreadyExists = errors.New("structured master resume version already exists")
)

type RegisterInput struct {
	UserID         string
	IdempotencyKey string
	SourceType     string
	FileObjectID   string
	RawText        string
	GuidedAnswers  map[string]any
	Title          string
	Language       string
}

type RegisterStore interface {
	CreateWithParseJob(ctx context.Context, in resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error)
}

type ReadStore interface {
	Get(ctx context.Context, userID string, assetID string) (resumestore.AssetRecord, error)
	List(ctx context.Context, userID string, filter resumestore.ListFilter) (resumestore.ListResult, error)
}

type StructuredMasterStore interface {
	CreateStructuredMasterFromAsset(ctx context.Context, in resumestore.CreateStructuredMasterInput) (resumestore.VersionRecord, error)
}

type VersionReadStore interface {
	GetVersionByID(ctx context.Context, userID string, versionID string) (resumestore.VersionRecord, error)
	ListVersionsByAsset(ctx context.Context, userID string, assetID string, filter resumestore.VersionListFilter) (resumestore.VersionListResult, error)
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

func (s *Service) RegisterResume(ctx context.Context, in RegisterInput) (api.ResumeAssetWithJob, error) {
	if s == nil || s.store == nil || s.newID == nil {
		return api.ResumeAssetWithJob{}, fmt.Errorf("resume service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	sourceType := strings.TrimSpace(in.SourceType)
	if userID == "" || strings.TrimSpace(in.IdempotencyKey) == "" || sourceType == "" || strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Language) == "" {
		return api.ResumeAssetWithJob{}, ErrValidationFailed
	}
	now := s.now()
	assetID := s.newID()
	jobID := s.newID()
	storeIn := resumestore.CreateAssetInput{
		AssetID:        assetID,
		UserID:         userID,
		JobID:          jobID,
		DedupeKey:      s.dedupeKey(userID, in.IdempotencyKey),
		SourceType:     sourceType,
		Title:          strings.TrimSpace(in.Title),
		Language:       strings.TrimSpace(in.Language),
		RawText:        strings.TrimSpace(in.RawText),
		GuidedAnswers:  cloneMap(in.GuidedAnswers),
		ParseStatus:    sharedtypes.TargetJobParseStatusQueued,
		JobStatus:      sharedtypes.JobStatusQueued,
		Now:            now,
		RequestPayload: resumestore.RegisterRequestPayload{SourceType: sourceType, Title: strings.TrimSpace(in.Title), Language: strings.TrimSpace(in.Language)},
	}
	switch sourceType {
	case "upload":
		if s.uploadRegister == nil {
			return api.ResumeAssetWithJob{}, fmt.Errorf("resume upload register service is not configured")
		}
		file, err := s.uploadRegister.RegisterFileObject(ctx, uploadservice.RegisterFileObjectInput{
			FileObjectID:    strings.TrimSpace(in.FileObjectID),
			ExpectedPurpose: uploadstore.PurposeResume,
			OwnerUserID:     userID,
		})
		if errors.Is(err, uploadservice.ErrValidationFailed) {
			return api.ResumeAssetWithJob{}, ErrValidationFailed
		}
		if err != nil {
			return api.ResumeAssetWithJob{}, err
		}
		storeIn.FileObjectID = &file.ID
		storeIn.RequestPayload.FileObjectID = strings.TrimSpace(in.FileObjectID)
	case "paste":
		storeIn.RequestPayload.RawTextHash = contentHash(in.RawText)
	case "guided":
		storeIn.RequestPayload.GuidedAnswersHash = contentHashMap(in.GuidedAnswers)
	default:
		return api.ResumeAssetWithJob{}, ErrValidationFailed
	}
	res, err := s.store.CreateWithParseJob(ctx, storeIn)
	if err != nil {
		return api.ResumeAssetWithJob{}, err
	}
	return api.ResumeAssetWithJob{
		ResumeAssetId: res.AssetID,
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

func (s *Service) GetResume(ctx context.Context, userID string, resumeAssetID string) (api.ResumeAsset, error) {
	if s == nil {
		return api.ResumeAsset{}, fmt.Errorf("resume read store is not configured")
	}
	reader, ok := s.store.(ReadStore)
	if !ok {
		return api.ResumeAsset{}, fmt.Errorf("resume read store is not configured")
	}
	rec, err := reader.Get(ctx, strings.TrimSpace(userID), strings.TrimSpace(resumeAssetID))
	if errors.Is(err, resumestore.ErrAssetNotFound) {
		return api.ResumeAsset{}, ErrNotFound
	}
	if err != nil {
		return api.ResumeAsset{}, err
	}
	return assetRecordToAPI(rec), nil
}

type ListRequest struct {
	UserID   string
	Cursor   string
	PageSize int
}

func (s *Service) ListResumes(ctx context.Context, in ListRequest) (api.PaginatedResumeAsset, error) {
	if s == nil {
		return api.PaginatedResumeAsset{}, fmt.Errorf("resume read store is not configured")
	}
	reader, ok := s.store.(ReadStore)
	if !ok {
		return api.PaginatedResumeAsset{}, fmt.Errorf("resume read store is not configured")
	}
	res, err := reader.List(ctx, strings.TrimSpace(in.UserID), resumestore.ListFilter{Cursor: in.Cursor, PageSize: in.PageSize})
	if err != nil {
		return api.PaginatedResumeAsset{}, err
	}
	out := api.PaginatedResumeAsset{Items: make([]api.ResumeAsset, 0, len(res.Items))}
	for _, item := range res.Items {
		out.Items = append(out.Items, assetRecordToAPI(item))
	}
	out.PageInfo = api.PageInfo{
		NextCursor: optionalString(res.NextCursor),
		PageSize:   res.PageSize,
		HasMore:    res.HasMore,
	}
	return out, nil
}

type ConfirmStructuredMasterInput struct {
	UserID            string
	ResumeAssetID     string
	DisplayName       string
	Language          string
	StructuredProfile map[string]any
}

func (s *Service) ConfirmStructuredMaster(ctx context.Context, in ConfirmStructuredMasterInput) (api.ResumeVersion, error) {
	if s == nil {
		return api.ResumeVersion{}, fmt.Errorf("resume structured master store is not configured")
	}
	store, ok := s.store.(StructuredMasterStore)
	if !ok {
		return api.ResumeVersion{}, fmt.Errorf("resume structured master store is not configured")
	}
	userID := strings.TrimSpace(in.UserID)
	assetID := strings.TrimSpace(in.ResumeAssetID)
	displayName := strings.TrimSpace(in.DisplayName)
	if userID == "" || assetID == "" || displayName == "" || len(in.StructuredProfile) == 0 {
		return api.ResumeVersion{}, ErrValidationFailed
	}
	provenance, err := extractVersionProvenance(in.StructuredProfile)
	if err != nil {
		return api.ResumeVersion{}, ErrValidationFailed
	}
	if provenance.Language == "" {
		provenance.Language = strings.TrimSpace(in.Language)
	}
	if provenance.Language == "" {
		return api.ResumeVersion{}, ErrValidationFailed
	}
	profile, err := json.Marshal(in.StructuredProfile)
	if err != nil {
		return api.ResumeVersion{}, ErrValidationFailed
	}
	rec, err := store.CreateStructuredMasterFromAsset(ctx, resumestore.CreateStructuredMasterInput{
		VersionID:         s.newID(),
		UserID:            userID,
		ResumeAssetID:     assetID,
		DisplayName:       displayName,
		StructuredProfile: profile,
		Provenance:        provenance,
		Now:               s.now(),
	})
	switch {
	case errors.Is(err, resumestore.ErrAssetNotFound):
		return api.ResumeVersion{}, ErrNotFound
	case errors.Is(err, resumestore.ErrAssetParseNotReady):
		return api.ResumeVersion{}, ErrAssetParseNotReady
	case errors.Is(err, resumestore.ErrStructuredMasterAlreadyExists):
		return api.ResumeVersion{}, ErrStructuredMasterAlreadyExists
	case err != nil:
		return api.ResumeVersion{}, err
	}
	return versionRecordToAPI(rec), nil
}

func (s *Service) GetResumeVersion(ctx context.Context, userID string, versionID string) (api.ResumeVersion, error) {
	if s == nil {
		return api.ResumeVersion{}, fmt.Errorf("resume version read store is not configured")
	}
	reader, ok := s.store.(VersionReadStore)
	if !ok {
		return api.ResumeVersion{}, fmt.Errorf("resume version read store is not configured")
	}
	rec, err := reader.GetVersionByID(ctx, strings.TrimSpace(userID), strings.TrimSpace(versionID))
	if errors.Is(err, resumestore.ErrVersionNotFound) {
		return api.ResumeVersion{}, ErrNotFound
	}
	if err != nil {
		return api.ResumeVersion{}, err
	}
	return versionRecordToAPI(rec), nil
}

type ListVersionRequest struct {
	UserID        string
	ResumeAssetID string
	Cursor        string
	PageSize      int
}

func (s *Service) ListResumeVersions(ctx context.Context, in ListVersionRequest) (api.PaginatedResumeVersion, error) {
	if s == nil {
		return api.PaginatedResumeVersion{}, fmt.Errorf("resume version read store is not configured")
	}
	reader, ok := s.store.(VersionReadStore)
	if !ok {
		return api.PaginatedResumeVersion{}, fmt.Errorf("resume version read store is not configured")
	}
	res, err := reader.ListVersionsByAsset(ctx, strings.TrimSpace(in.UserID), strings.TrimSpace(in.ResumeAssetID), resumestore.VersionListFilter{Cursor: in.Cursor, PageSize: in.PageSize})
	switch {
	case errors.Is(err, resumestore.ErrAssetNotFound):
		return api.PaginatedResumeVersion{}, ErrNotFound
	case errors.Is(err, resumestore.ErrInvalidCursor):
		return api.PaginatedResumeVersion{}, ErrInvalidCursor
	case err != nil:
		return api.PaginatedResumeVersion{}, err
	}
	out := api.PaginatedResumeVersion{Items: make([]api.ResumeVersion, 0, len(res.Items))}
	for _, item := range res.Items {
		out.Items = append(out.Items, versionRecordToAPI(item))
	}
	out.PageInfo = api.PageInfo{
		NextCursor: optionalString(res.NextCursor),
		PageSize:   res.PageSize,
		HasMore:    res.HasMore,
	}
	return out, nil
}

func (s *Service) dedupeKey(userID, idempotencyKey string) string {
	h := sha256.New()
	h.Write([]byte("resume.register.v1"))
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

func contentHashMap(in map[string]any) string {
	return contentHash(fmt.Sprintf("%#v", in))
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

func assetRecordToAPI(rec resumestore.AssetRecord) api.ResumeAsset {
	status := "active"
	out := api.ResumeAsset{
		Id:            rec.ID,
		Title:         rec.Title,
		Language:      rec.Language,
		ParseStatus:   rec.ParseStatus,
		Status:        &status,
		FileObjectId:  cloneStringPtr(rec.FileObjectID),
		OriginalText:  cloneStringPtr(rec.OriginalText),
		SourceType:    cloneStringPtr(rec.SourceType),
		CreatedAt:     rec.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     rec.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:     timePtrToString(rec.DeletedAt),
		ParsedSummary: rawJSONMapPtr(rec.ParsedSummary),
		GuidedAnswers: rawJSONMapPtr(rec.GuidedAnswers),
	}
	if rec.ParsedTextSnapshot != nil {
		out.ParsedTextSnapshot = cloneStringPtr(rec.ParsedTextSnapshot)
	}
	return out
}

func extractVersionProvenance(profile map[string]any) (resumestore.VersionProvenance, error) {
	raw, ok := profile["provenance"].(map[string]any)
	if !ok || len(raw) == 0 {
		return resumestore.VersionProvenance{}, fmt.Errorf("structured profile provenance is required")
	}
	out := resumestore.VersionProvenance{
		PromptVersion:     stringValue(raw["promptVersion"]),
		RubricVersion:     stringValue(raw["rubricVersion"]),
		ModelID:           stringValue(raw["modelId"]),
		Provider:          stringValue(raw["provider"]),
		Language:          stringValue(raw["language"]),
		FeatureFlag:       stringValue(raw["featureFlag"]),
		DataSourceVersion: stringValue(raw["dataSourceVersion"]),
	}
	if out.PromptVersion == "" || out.RubricVersion == "" || out.ModelID == "" || out.FeatureFlag == "" || out.DataSourceVersion == "" {
		return resumestore.VersionProvenance{}, fmt.Errorf("structured profile provenance is incomplete")
	}
	return out, nil
}

func stringValue(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func versionRecordToAPI(rec resumestore.VersionRecord) api.ResumeVersion {
	profile := rawJSONAny(rec.StructuredProfile)
	provenance := rec.Provenance
	if profileMap, ok := profile.(map[string]any); ok {
		if profileProvenance, err := extractVersionProvenance(profileMap); err == nil {
			fillProvenance(&provenance, profileProvenance)
		}
	}
	if provenance.PromptVersion == "" && rec.PromptVersion != nil {
		provenance.PromptVersion = *rec.PromptVersion
	}
	if provenance.RubricVersion == "" && rec.RubricVersion != nil {
		provenance.RubricVersion = *rec.RubricVersion
	}
	if provenance.ModelID == "" && rec.ModelID != nil {
		provenance.ModelID = *rec.ModelID
	}
	if provenance.Provider == "" && rec.Provider != nil {
		provenance.Provider = *rec.Provider
	}
	promptVersion := cloneStringPtr(rec.PromptVersion)
	if promptVersion == nil && provenance.PromptVersion != "" {
		promptVersion = pointerString(provenance.PromptVersion)
	}
	rubricVersion := cloneStringPtr(rec.RubricVersion)
	if rubricVersion == nil && provenance.RubricVersion != "" {
		rubricVersion = pointerString(provenance.RubricVersion)
	}
	modelID := cloneStringPtr(rec.ModelID)
	if modelID == nil && provenance.ModelID != "" {
		modelID = pointerString(provenance.ModelID)
	}
	provider := cloneStringPtr(rec.Provider)
	if provider == nil && provenance.Provider != "" {
		provider = pointerString(provenance.Provider)
	}
	out := api.ResumeVersion{
		Id:                rec.ID,
		ResumeAssetId:     rec.ResumeAssetID,
		ParentVersionId:   cloneStringPtr(rec.ParentVersionID),
		VersionType:       rec.VersionType,
		TargetJobId:       cloneStringPtr(rec.TargetJobID),
		DisplayName:       rec.DisplayName,
		SeedStrategy:      cloneSeedStrategyPtr(rec.SeedStrategy),
		FocusAngle:        cloneStringPtr(rec.FocusAngle),
		StructuredProfile: profile,
		MatchScore:        cloneFloatPtr(rec.MatchScore),
		PromptVersion:     promptVersion,
		RubricVersion:     rubricVersion,
		ModelId:           modelID,
		Provider:          provider,
		Provenance: api.GenerationProvenance{
			PromptVersion:     provenance.PromptVersion,
			RubricVersion:     provenance.RubricVersion,
			ModelId:           provenance.ModelID,
			Language:          provenance.Language,
			FeatureFlag:       provenance.FeatureFlag,
			DataSourceVersion: provenance.DataSourceVersion,
		},
		Suggestions: cloneAnySlice(rec.Suggestions),
		CreatedAt:   rec.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   rec.UpdatedAt.UTC().Format(time.RFC3339),
		DeletedAt:   timePtrToString(rec.DeletedAt),
	}
	return out
}

func cloneAnySlice(in []any) []any {
	if in == nil {
		return []any{}
	}
	out := make([]any, len(in))
	copy(out, in)
	return out
}

func fillProvenance(target *resumestore.VersionProvenance, fallback resumestore.VersionProvenance) {
	if target.PromptVersion == "" {
		target.PromptVersion = fallback.PromptVersion
	}
	if target.RubricVersion == "" {
		target.RubricVersion = fallback.RubricVersion
	}
	if target.ModelID == "" {
		target.ModelID = fallback.ModelID
	}
	if target.Provider == "" {
		target.Provider = fallback.Provider
	}
	if target.Language == "" {
		target.Language = fallback.Language
	}
	if target.FeatureFlag == "" {
		target.FeatureFlag = fallback.FeatureFlag
	}
	if target.DataSourceVersion == "" {
		target.DataSourceVersion = fallback.DataSourceVersion
	}
}

func pointerString(in string) *string {
	v := in
	return &v
}

func rawJSONAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func cloneSeedStrategyPtr(in *sharedtypes.ResumeSeedStrategy) *sharedtypes.ResumeSeedStrategy {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func cloneFloatPtr(in *float64) *float64 {
	if in == nil {
		return nil
	}
	v := *in
	return &v
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
