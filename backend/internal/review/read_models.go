package review

import (
	"time"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const ReportListMaxPageSize = 50

var ErrReportNotFound = errReportNotFound{}

type errReportNotFound struct{}

func (errReportNotFound) Error() string { return "review: feedback report not found" }

type ReportEvidenceRecord struct {
	Dimension  string
	Evidence   string
	Confidence sharedtypes.Confidence
}

type ReportNextActionRecord struct {
	Type  string
	Label string
}

type DimensionResultRecord struct {
	Status     sharedtypes.DimensionStatus
	Confidence sharedtypes.Confidence
}

type QuestionAssessmentRecord struct {
	TurnID              string
	QuestionIntent      string
	DimensionResults    map[string]DimensionResultRecord
	ReviewStatus        sharedtypes.QuestionReviewStatus
	IncludedInRetryPlan bool
}

type GenerationProvenanceRecord struct {
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
}

type FeedbackReportRecord struct {
	ID                  string
	SessionID           string
	TargetJobID         string
	Status              sharedtypes.ReportStatus
	PreparednessLevel   *sharedtypes.ReadinessTier
	Highlights          []ReportEvidenceRecord
	Issues              []ReportEvidenceRecord
	NextActions         []ReportNextActionRecord
	QuestionAssessments []QuestionAssessmentRecord
	RetryFocusTurnIDs   []string
	Provenance          *GenerationProvenanceRecord
	ErrorCode           *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ListTargetJobReportsRequest struct {
	UserID      string
	TargetJobID string
	Cursor      string
	PageSize    int
}

type ListTargetJobReportsInput struct {
	UserID          string
	TargetJobID     string
	Cursor          string
	CursorCreatedAt time.Time
	CursorID        string
	PageSize        int
}

type ListTargetJobReportsResult struct {
	Items      []FeedbackReportRecord
	NextCursor string
	HasMore    bool
	PageSize   int
}

type PageInfo struct {
	NextCursor string
	PageSize   int
	HasMore    bool
}

type PaginatedFeedbackReportRecord struct {
	Items    []FeedbackReportRecord
	PageInfo PageInfo
}

func EffectiveReportPageSize(pageSize int) int {
	if pageSize <= 0 {
		return sharedtypes.DefaultPageSize
	}
	if pageSize > ReportListMaxPageSize {
		return ReportListMaxPageSize
	}
	return pageSize
}
