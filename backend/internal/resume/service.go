package resume

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	ErrNotFound         = errors.New("resume asset not found")
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
