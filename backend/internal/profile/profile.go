package profile

import (
	"context"
	stderrs "errors"
	"time"
)

// Source type literals for experience_cards.source_type. backend-profile spec D-6
// forces "manual" on the public HTTP create path; other values are reserved for
// future cross-owner AppendExperienceCardEvidence internal API.
const (
	SourceTypeManual          = "manual"
	SourceTypeResumeParse     = "resume_parse"
	SourceTypePracticeReport  = "practice_report"
	SourceTypeDebrief         = "debrief"
	ConfidenceDefaultMedium   = "medium"
	DefaultExperienceCardSize = 20
	MaxExperienceCardSize     = 100
)

// SourceTypes enumerates all valid experience_cards.source_type values, in
// stable order so service-layer count maps remain deterministic.
var SourceTypes = []string{
	SourceTypeManual,
	SourceTypeResumeParse,
	SourceTypePracticeReport,
	SourceTypeDebrief,
}

// CandidateProfileRecord mirrors the candidate_profiles row shape backing the
// HTTP CandidateProfile response. Null DB columns (headline / currentRole /
// region / yearsOfExperience) are represented as Optional* values so handlers
// can surface JSON null per spec D-1.
type CandidateProfileRecord struct {
	UserID                    string
	Headline                  *string
	YearsOfExperience         *int32
	CurrentRole               *string
	PreferredPracticeLanguage string
	UiLanguage                string
	Region                    *string
	ProfileVersion            int32
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// ExperienceCardRecord mirrors the experience_cards row shape backing the
// HTTP ExperienceCard response.
type ExperienceCardRecord struct {
	ID          string
	UserID      string
	ProfileID   string
	Title       string
	CompanyName string
	Situation   string
	Task        string
	Action      string
	Result      string
	Skills      []string
	Language    string
	SourceType  string
	SourceRefID *string
	Confidence  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserSettings carries the seed defaults read from user_settings during the
// first GetMyProfile call (spec D-1).
type UserSettings struct {
	PreferredPracticeLanguage string
	UiLanguage                string
	Region                    *string
}

// ProfilePatch carries the optional update fields for UpdateMyProfile (spec
// D-2). nil fields are not updated; non-nil *string fields are written even if
// empty (empty string clears the column).
type ProfilePatch struct {
	Headline                  *string
	YearsOfExperience         *int32
	CurrentRole               *string
	PreferredPracticeLanguage *string
	UiLanguage                *string
	Region                    *string
}

// ExperienceCardAttrs carries the mutable card attributes shared by Create /
// Update store calls.
type ExperienceCardAttrs struct {
	Title       string
	CompanyName string
	Situation   string
	Task        string
	Action      string
	Result      string
	Skills      []string
	Language    string
}

// ExperienceCardSource carries the provenance fields (manual creation always
// pins SourceType=manual / SourceRefID=nil).
type ExperienceCardSource struct {
	SourceType  string
	SourceRefID *string
	Confidence  string
}

// ExperienceCardPatch carries partial-update fields for UpdateExperienceCard
// (spec D-2 semantics — only supplied fields update). nil pointer = leave
// column untouched; non-nil *string with empty value = write empty string.
type ExperienceCardPatch struct {
	Title       *string
	CompanyName *string
	Situation   *string
	Task        *string
	Action      *string
	Result      *string
	Skills      *[]string
	Language    *string
}

// ListCardsCursor is the parsed cursor for cursor pagination.
type ListCardsCursor struct {
	UpdatedAt time.Time
	ID        string
}

// ListCardsResult wraps page contents and pagination metadata for
// ListExperienceCards.
type ListCardsResult struct {
	Items      []ExperienceCardRecord
	NextCursor string
	HasMore    bool
	PageSize   int32
}

// SourceCounts is the canonical return shape for CountExperienceCardsBySource
// (spec D-11). Map keys are taken from SourceTypes; absent keys default to 0.
type SourceCounts map[string]int64

// Domain-level error sentinels. Handler / service / store layers all map their
// implementation-specific errors to one of these.
var (
	ErrNotFound         = stderrs.New("profile: resource not found")
	ErrCrossUser        = stderrs.New("profile: cross-user access denied")
	ErrInvalidCursor    = stderrs.New("profile: invalid cursor")
	ErrValidationFailed = stderrs.New("profile: validation failed")
)

// Store is the canonical persistence boundary for candidate_profiles and
// experience_cards. Handlers / services depend on this interface only;
// concrete implementations live under backend/internal/profile/store.
type Store interface {
	GetCandidateProfileByUser(ctx context.Context, userID string) (*CandidateProfileRecord, error)
	UpsertLite(ctx context.Context, userID string, patch ProfilePatch, defaults UserSettings) (*CandidateProfileRecord, error)
	SeedCandidateProfile(ctx context.Context, userID string, defaults UserSettings) (*CandidateProfileRecord, error)
	DeleteCandidateProfileForUser(ctx context.Context, userID string) (int64, error)
	ListExperienceCardsByUser(ctx context.Context, userID string, cursor *ListCardsCursor, pageSize int32) (ListCardsResult, error)
	CreateExperienceCard(ctx context.Context, id string, userID string, attrs ExperienceCardAttrs, source ExperienceCardSource) (*ExperienceCardRecord, error)
	UpdateExperienceCard(ctx context.Context, cardID string, userID string, patch ExperienceCardPatch) (*ExperienceCardRecord, error)
	DeleteExperienceCardsForUser(ctx context.Context, userID string) (int64, error)
	CountExperienceCardsBySource(ctx context.Context, userID string) (SourceCounts, error)
}

// SettingsReader resolves the per-user user_settings row used to seed first
// GetMyProfile call (spec D-1). Implementations may shim to backend-auth's
// user_settings store; the interface keeps backend-profile independent of the
// concrete auth wiring.
type SettingsReader interface {
	GetUserSettings(ctx context.Context, userID string) (UserSettings, error)
}

// AuditTombstoneWriter records a privacy-delete tombstone after a successful
// DeleteCandidateProfileForUser run (spec D-9 audit_events tombstone payload —
// no raw card content).
type AuditTombstoneWriter interface {
	WriteCandidateProfileDeleteTombstone(ctx context.Context, in CandidateProfileDeleteTombstone) error
}

// CandidateProfileDeleteTombstone is the audit payload written after a
// successful privacy delete. Caller fills UserID + ExperienceCardCount +
// DeletedAt + JobID; resource_type / action / metadata schema are owned by the
// writer implementation (spec §4.3 PII redline — no title / situation / task /
// action / result / metrics / skills / headline / currentRole / region).
type CandidateProfileDeleteTombstone struct {
	UserID              string
	ExperienceCardCount int64
	DeletedAt           time.Time
	JobID               string
}
