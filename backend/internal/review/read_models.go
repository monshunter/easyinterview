package review

import (
	"time"

	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

var ErrReportNotFound = errReportNotFound{}

type errReportNotFound struct{}

func (errReportNotFound) Error() string { return "review: feedback report not found" }

type ReportEvidenceRecord struct {
	DimensionCode       string                 `json:"dimensionCode"`
	Evidence            string                 `json:"evidence"`
	Confidence          sharedtypes.Confidence `json:"confidence"`
	SourceMessageSeqNos []int32                `json:"sourceMessageSeqNos"`
}

type ReportNextActionRecord struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

type DimensionAssessmentRecord struct {
	Code       string                      `json:"code"`
	Label      string                      `json:"label"`
	Status     sharedtypes.DimensionStatus `json:"status"`
	Confidence sharedtypes.Confidence      `json:"confidence"`
}

type GenerationProvenanceRecord struct {
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
}

type ReportContextProjection struct {
	SourcePlanID      string
	TargetJobTitle    string
	TargetJobCompany  string
	ResumeID          string
	ResumeDisplayName string
	RoundID           string
	RoundSequence     int32
	RoundName         string
	RoundType         string
	Language          string
	HasNextRound      bool
}

func ProjectFrozenReportContext(snapshot practicedomain.ReportContextSnapshot) ReportContextProjection {
	return ReportContextProjection{
		SourcePlanID:      snapshot.Plan.ID,
		TargetJobTitle:    snapshot.TargetJob.Title,
		TargetJobCompany:  snapshot.TargetJob.Company,
		ResumeID:          snapshot.Resume.ID,
		ResumeDisplayName: snapshot.Resume.DisplayName,
		RoundID:           snapshot.Round.ID,
		RoundSequence:     snapshot.Round.Sequence,
		RoundName:         snapshot.Round.Name,
		RoundType:         snapshot.Round.Type,
		Language:          snapshot.Conversation.Language,
		HasNextRound:      snapshot.HasNextRound,
	}
}

type FeedbackReportRecord struct {
	ID                       string
	SessionID                string
	TargetJobID              string
	Status                   sharedtypes.ReportStatus
	ErrorCode                *string
	Summary                  *string
	Context                  ReportContextProjection
	PreparednessLevel        *sharedtypes.ReadinessTier
	Highlights               []ReportEvidenceRecord
	Issues                   []ReportEvidenceRecord
	NextActions              []ReportNextActionRecord
	DimensionAssessments     []DimensionAssessmentRecord
	RetryFocusDimensionCodes []string
	Provenance               *GenerationProvenanceRecord
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type ListTargetJobReportsRequest struct {
	UserID      string
	TargetJobID string
}

type PracticeRoundRefRecord struct {
	RoundID       string
	RoundSequence int32
}

type TargetJobCurrentReportSummaryRecord struct {
	ID          string
	GeneratedAt time.Time
}

type TargetJobReportAttemptSummaryRecord struct {
	ID        string
	Status    sharedtypes.ReportStatus
	ErrorCode *string
	CreatedAt time.Time
}

type TargetJobReportRoundOverviewRecord struct {
	Round         PracticeRoundRefRecord
	CurrentReport *TargetJobCurrentReportSummaryRecord
	LatestAttempt *TargetJobReportAttemptSummaryRecord
}

type TargetJobReportsOverviewRecord struct {
	TargetJobID string
	Rounds      []TargetJobReportRoundOverviewRecord
}
