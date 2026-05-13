package store

import (
	"encoding/json"
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

type AssetRecord struct {
	ID                 string
	UserID             string
	FileObjectID       *string
	Title              string
	Language           string
	ParseStatus        sharedtypes.TargetJobParseStatus
	ParsedSummary      json.RawMessage
	OriginalText       *string
	GuidedAnswers      json.RawMessage
	ParsedTextSnapshot *string
	SourceType         *string
	ErrorCode          *string
	LatestParseJobID   *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type ListFilter struct {
	Cursor   string
	PageSize int
}

type ListResult struct {
	Items      []AssetRecord
	NextCursor string
	HasMore    bool
	PageSize   int
}

type StatusUpdateInput struct {
	UserID  string
	AssetID string
	Now     time.Time
}

type MarkReadyInput struct {
	UserID             string
	AssetID            string
	ParsedSummary      json.RawMessage
	ParsedTextSnapshot string
	Now                time.Time
}

type MarkFailedInput struct {
	UserID    string
	AssetID   string
	ErrorCode string
	Now       time.Time
}

type ParseAssetRecord struct {
	ID            string
	UserID        string
	Language      string
	ParseStatus   sharedtypes.TargetJobParseStatus
	SourceType    string
	OriginalText  string
	GuidedAnswers json.RawMessage
	FileObjectID  string
	FileObjectKey string
}

type CompleteParseSuccessInput struct {
	UserID             string
	AssetID            string
	ParsedSummary      json.RawMessage
	ParsedTextSnapshot string
	OutboxEventID      string
	OutboxEventPayload []byte
	Now                time.Time
}

type CompleteParseFailureInput struct {
	UserID    string
	AssetID   string
	ErrorCode string
	Now       time.Time
}
