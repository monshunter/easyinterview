package jdmatch

import (
	stderrs "errors"
	"time"
)

// RecommendationRecord is the on-disk projection of a
// `jd_match_recommendations` row. Stored fields cover both the list
// and detail projections; handlers project the subset they need.
type RecommendationRecord struct {
	ID                  string
	UserID              string
	Title               string
	Company             string
	CompanyTag          *string
	Level               *string
	Location            string
	Comp                *string
	PostedLabel         *string
	Score               int
	FitMust             int
	FitTotal            int
	FitPlus             int
	FitTotalPlus        int
	Reasons             []string
	Risks               []string
	Highlights          []string
	Seen                bool
	DismissedAt         *time.Time
	DismissReason       *string
	DismissFreeNote     *string
	SourceURL           *string
	SourceLabel         *string
	NetworkNote         *string
	SimilarInterviewers *int
	InterviewHypotheses []string
	PromptVersion       *string
	RubricVersion       *string
	ModelID             *string
	Language            string
	FeatureFlag         string
	DataSourceVersion   string
	RecommendedAt       time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

// ErrAlreadyDismissed is returned when MarkDismissed targets a row
// whose dismissed_at is already set; handlers map it to a 400
// VALIDATION_FAILED envelope rather than a generic 500.
var ErrAlreadyDismissed = stderrs.New("jdmatch: recommendation already dismissed")
