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

type VersionProvenance struct {
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	Provider          string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
}

type CreateStructuredMasterInput struct {
	VersionID         string
	UserID            string
	ResumeAssetID     string
	DisplayName       string
	StructuredProfile json.RawMessage
	Provenance        VersionProvenance
	Now               time.Time
}

type VersionRecord struct {
	ID                string
	UserID            string
	ResumeAssetID     string
	ParentVersionID   *string
	VersionType       sharedtypes.ResumeVersionType
	TargetJobID       *string
	DisplayName       string
	SeedStrategy      *sharedtypes.ResumeSeedStrategy
	FocusAngle        *string
	StructuredProfile json.RawMessage
	MatchScore        *float64
	PromptVersion     *string
	RubricVersion     *string
	ModelID           *string
	Provider          *string
	Provenance        VersionProvenance
	Suggestions       []any
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type VersionListFilter struct {
	Cursor   string
	PageSize int
}

type VersionListResult struct {
	Items      []VersionRecord
	NextCursor string
	HasMore    bool
	PageSize   int
}

type VersionUpdateInput struct {
	UserID                 string
	VersionID              string
	DisplayName            *string
	DisplayNameSet         bool
	FocusAngle             *string
	FocusAngleSet          bool
	MatchScore             *float64
	MatchScoreSet          bool
	StructuredProfile      json.RawMessage
	StructuredProfileSet   bool
	StructuredProfilePatch map[string]any
	Now                    time.Time
}

type BranchVersionInput struct {
	VersionID       string
	UserID          string
	ParentVersionID string
	TargetJobID     string
	SeedStrategy    sharedtypes.ResumeSeedStrategy
	DisplayName     string
	FocusAngle      *string
	Provenance      VersionProvenance
	TailorRunID     string
	JobID           string
	DedupeKey       string
	Now             time.Time
}

type BranchVersionResult struct {
	Version      VersionRecord
	TailorRunID  string
	JobID        string
	JobStatus    sharedtypes.JobStatus
	JobCreatedAt time.Time
	JobUpdatedAt time.Time
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
