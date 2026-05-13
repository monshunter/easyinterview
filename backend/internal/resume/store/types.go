package store

import (
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type RegisterRequestPayload struct {
	SourceType        string `json:"sourceType"`
	Title             string `json:"title"`
	Language          string `json:"language"`
	FileObjectID      string `json:"fileObjectId,omitempty"`
	RawTextHash       string `json:"rawTextHash,omitempty"`
	GuidedAnswersHash string `json:"guidedAnswersHash,omitempty"`
}

type CreateAssetInput struct {
	AssetID        string
	UserID         string
	JobID          string
	DedupeKey      string
	SourceType     string
	FileObjectID   *string
	Title          string
	Language       string
	RawText        string
	GuidedAnswers  map[string]any
	ParseStatus    sharedtypes.TargetJobParseStatus
	JobStatus      sharedtypes.JobStatus
	RequestPayload RegisterRequestPayload
	Now            time.Time
}

type CreateAssetResult struct {
	AssetID      string
	JobID        string
	JobStatus    sharedtypes.JobStatus
	JobCreatedAt time.Time
	JobUpdatedAt time.Time
	Existing     bool
}
