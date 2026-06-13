package store

import (
	"encoding/json"
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type RegisterRequestPayload struct {
	SourceType   string `json:"sourceType"`
	Title        string `json:"title"`
	Language     string `json:"language"`
	FileObjectID string `json:"fileObjectId,omitempty"`
	RawTextHash  string `json:"rawTextHash,omitempty"`
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

// ResumeRecord is a single flat resume asset (D-20 flatten): a read-only source
// snapshot plus the editable structured_profile / display_name merged in from
// the retired structured_master version.
type ResumeRecord struct {
	ID                 string
	UserID             string
	FileObjectID       *string
	Title              string
	DisplayName        *string
	Language           string
	ParseStatus        sharedtypes.TargetJobParseStatus
	ParsedSummary      json.RawMessage
	OriginalText       *string
	StructuredProfile  json.RawMessage
	ParsedTextSnapshot *string
	SourceType         *string
	ErrorCode          *string
	LatestParseJobID   *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

type UpdateResumeInput struct {
	UserID            string
	ResumeID          string
	DisplayName       *string
	StructuredProfile json.RawMessage
	Now               time.Time
}

type DuplicateResumeInput struct {
	NewResumeID       string
	UserID            string
	SourceResumeID    string
	DisplayName       *string
	StructuredProfile json.RawMessage
	Now               time.Time
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

type CreateTailorRunInput struct {
	TailorRunID string
	JobID       string
	UserID      string
	TargetJobID string
	ResumeID    string
	Mode        string
	DedupeKey   string
	Now         time.Time
}

type CreateTailorRunResult struct {
	TailorRunID  string
	JobID        string
	JobStatus    sharedtypes.JobStatus
	JobCreatedAt time.Time
	JobUpdatedAt time.Time
}

// TailorRunRecord reconstructs a tailor run from the async_jobs row that drives
// it (D-20 dropped the dedicated resume_tailor_runs table). The result jsonb
// carries the ephemeral match summary, bullet suggestions, and provenance.
type TailorRunRecord struct {
	ID           string
	UserID       string
	TargetJobID  string
	ResumeID     string
	Mode         string
	Status       string
	MatchSummary json.RawMessage
	Suggestions  json.RawMessage
	Provenance   VersionProvenance
	ErrorCode    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TailorJobContext struct {
	TailorRunID       string
	UserID            string
	ResumeID          string
	TargetJobID       string
	Mode              string
	Language          string
	ResumeSummary     json.RawMessage
	StructuredProfile json.RawMessage
	TargetSummary     json.RawMessage
	TargetTitle       string
	TargetCompany     string
	TargetSeniority   string
	RawJDText         string
	OriginalBullet    string
}

type TailorSuggestionInput struct {
	ID              string
	OriginalBullet  string
	SuggestedBullet string
	Reason          string
}

type CompleteTailorRunSuccessInput struct {
	TailorRunID        string
	ResumeID           string
	TargetJobID        string
	Mode               string
	MatchSummary       json.RawMessage
	Suggestions        []TailorSuggestionInput
	Provenance         VersionProvenance
	OutboxEventID      string
	OutboxEventPayload []byte
	Now                time.Time
}

type ListFilter struct {
	Cursor   string
	PageSize int
}

type ListResult struct {
	Items      []ResumeRecord
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
	StructuredProfile  json.RawMessage
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
	FileObjectID  string
	FileObjectKey string
}

type CompleteParseSuccessInput struct {
	UserID             string
	AssetID            string
	ParsedSummary      json.RawMessage
	StructuredProfile  json.RawMessage
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
