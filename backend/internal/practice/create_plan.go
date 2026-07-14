package practice

import (
	"context"
	stderrs "errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

var (
	ErrPlanPrerequisiteNotFound = stderrs.New("practice plan prerequisite not found")
	ErrPlanNotFound             = stderrs.New("practice plan not found")
	ErrSessionNotFound          = stderrs.New("practice session not found")
	ErrSessionConflict          = stderrs.New("practice session conflict")
	ErrSessionNotReportable     = stderrs.New("practice session is not reportable")
	ErrClientEventMismatch      = stderrs.New("practice session client event mismatch")
	ErrInvalidCursor            = stderrs.New("practice session invalid cursor")
)

type ServiceError struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	return e.Code + ": " + e.Message
}

type Store interface {
	CreatePlan(ctx context.Context, in CreatePlanStoreInput) (PlanRecord, error)
	GetPlan(ctx context.Context, userID, planID string) (PlanRecord, error)
	ListSessions(ctx context.Context, in ListSessionsInput) (ListSessionsResult, error)
	GetSession(ctx context.Context, userID, sessionID string, now time.Time) (SessionRecord, error)
	ReservePracticeMessage(ctx context.Context, in ReservePracticeMessageInput) (PracticeMessageReservation, error)
	FailPracticeMessage(ctx context.Context, in FailPracticeMessageInput) error
	CommitPracticeMessage(ctx context.Context, in CommitPracticeMessageInput) (SendPracticeMessageResult, error)
	CompleteSession(ctx context.Context, in CompleteSessionStoreInput) (CompleteSessionResult, error)
	ReserveSessionStart(ctx context.Context, in StartSessionReservationInput) (SessionReservation, error)
	CommitSessionStart(ctx context.Context, in CommitSessionStartInput) (SessionRecord, error)
	FailSessionStart(ctx context.Context, in FailSessionStartInput) error
}

type PromptResolver interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type ServiceOptions struct {
	Store      Store
	Registry   PromptResolver
	AI         aiclient.AIClient
	AITaskRuns aiclient.AITaskRunWriter
	Now        func() time.Time
	NewID      func() string
}

type Service struct {
	store      Store
	registry   PromptResolver
	ai         aiclient.AIClient
	aiTaskRuns aiclient.AITaskRunWriter
	now        func() time.Time
	newID      func() string
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
	return &Service{store: opts.Store, registry: opts.Registry, ai: opts.AI, aiTaskRuns: opts.AITaskRuns, now: now, newID: newID}
}

type CreatePlanRequest struct {
	UserID             string
	TargetJobID        string
	ResumeID           string
	SourceReportID     string
	RoundID            string
	Goal               sharedtypes.PracticeGoal
	InterviewerPersona sharedtypes.InterviewerRole
	Difficulty         string
	Language           string
	TimeBudgetMinutes  int32
}

type CreatePlanStoreInput struct {
	PlanID             string
	AuditEventID       string
	UserID             string
	TargetJobID        string
	ResumeID           string
	SourceReportID     string
	RoundID            string
	Goal               sharedtypes.PracticeGoal
	InterviewerPersona sharedtypes.InterviewerRole
	Difficulty         string
	Language           string
	TimeBudgetMinutes  int32
	Now                time.Time
}

type PlanRecord struct {
	ID                  string
	TargetJobID         string
	SourceReportID      string
	ResumeID            string
	RoundID             string
	RoundSequence       int32
	Goal                sharedtypes.PracticeGoal
	InterviewerPersona  sharedtypes.InterviewerRole
	Difficulty          string
	Language            string
	TimeBudgetMinutes   int32
	FocusDimensionCodes []string
	Status              string
	CreatedAt           time.Time
}

type ListSessionsRequest struct {
	UserID      string
	TargetJobID string
	Status      sharedtypes.SessionStatus
	Cursor      string
	PageSize    int
}

type ListSessionsInput struct {
	UserID      string
	TargetJobID string
	Status      sharedtypes.SessionStatus
	Cursor      string
	PageSize    int
}

type ListSessionsResult struct {
	Items      []SessionRecord
	NextCursor string
	HasMore    bool
	PageSize   int
}

func (s *Service) CreatePracticePlan(ctx context.Context, in CreatePlanRequest) (PlanRecord, error) {
	if s == nil || s.store == nil {
		return PlanRecord{}, fmt.Errorf("practice service is not initialised")
	}
	if strings.TrimSpace(in.UserID) == "" {
		return PlanRecord{}, fmt.Errorf("userId is required")
	}
	if !validPracticeGoal(in.Goal) {
		return PlanRecord{}, validationError("goal is invalid", map[string]any{"field": "goal", "goal": string(in.Goal)})
	}
	sourceReportID := strings.TrimSpace(in.SourceReportID)
	if err := validateCreatePlanSources(in.Goal, sourceReportID); err != nil {
		return PlanRecord{}, err
	}
	language := ""
	if isReportDerivedGoal(in.Goal) {
		if field := copiedDerivedPlanField(in); field != "" {
			return PlanRecord{}, validationError("report-derived plan fields are server-owned", map[string]any{"field": field})
		}
	} else {
		if strings.TrimSpace(in.TargetJobID) == "" {
			return PlanRecord{}, validationError("targetJobId is required", map[string]any{"field": "targetJobId"})
		}
		if strings.TrimSpace(in.ResumeID) == "" {
			return PlanRecord{}, validationError("A resume asset must be bound before creating this practice plan.", map[string]any{"field": "resumeId"})
		}
		if !validInterviewerRole(in.InterviewerPersona) {
			return PlanRecord{}, validationError("interviewerPersona is invalid", map[string]any{"field": "interviewerPersona"})
		}
		if !validDifficulty(in.Difficulty) {
			return PlanRecord{}, validationError("difficulty is invalid", map[string]any{"field": "difficulty"})
		}
		if strings.TrimSpace(in.Language) == "" {
			return PlanRecord{}, validationError("language is required", map[string]any{"field": "language"})
		}
		var ok bool
		language, ok = canonicalPracticeLanguage(in.Language)
		if !ok {
			return PlanRecord{}, validationError("language is invalid", map[string]any{"field": "language"})
		}
		if in.TimeBudgetMinutes < 1 {
			return PlanRecord{}, validationError("timeBudgetMinutes must be positive", map[string]any{"field": "timeBudgetMinutes"})
		}
	}

	now := s.now().UTC()
	plan, err := s.store.CreatePlan(ctx, CreatePlanStoreInput{
		PlanID:             s.newID(),
		AuditEventID:       s.newID(),
		UserID:             strings.TrimSpace(in.UserID),
		TargetJobID:        strings.TrimSpace(in.TargetJobID),
		ResumeID:           strings.TrimSpace(in.ResumeID),
		SourceReportID:     sourceReportID,
		RoundID:            strings.TrimSpace(in.RoundID),
		Goal:               in.Goal,
		InterviewerPersona: in.InterviewerPersona,
		Difficulty:         strings.TrimSpace(in.Difficulty),
		Language:           language,
		TimeBudgetMinutes:  in.TimeBudgetMinutes,
		Now:                now,
	})
	if stderrs.Is(err, ErrPlanPrerequisiteNotFound) {
		return PlanRecord{}, validationError("target job or resume asset is not available", map[string]any{
			"targetJobId": strings.TrimSpace(in.TargetJobID),
			"resumeId":    strings.TrimSpace(in.ResumeID),
		})
	}
	if err != nil {
		return PlanRecord{}, err
	}
	return plan, nil
}

func isReportDerivedGoal(goal sharedtypes.PracticeGoal) bool {
	return goal == sharedtypes.PracticeGoalRetryCurrentRound || goal == sharedtypes.PracticeGoalNextRound
}

func copiedDerivedPlanField(in CreatePlanRequest) string {
	for _, field := range []struct {
		name    string
		present bool
	}{
		{name: "targetJobId", present: strings.TrimSpace(in.TargetJobID) != ""},
		{name: "resumeId", present: strings.TrimSpace(in.ResumeID) != ""},
		{name: "roundId", present: strings.TrimSpace(in.RoundID) != ""},
		{name: "interviewerPersona", present: in.InterviewerPersona != ""},
		{name: "difficulty", present: strings.TrimSpace(in.Difficulty) != ""},
		{name: "language", present: strings.TrimSpace(in.Language) != ""},
		{name: "timeBudgetMinutes", present: in.TimeBudgetMinutes != 0},
	} {
		if field.present {
			return field.name
		}
	}
	return ""
}

func validateCreatePlanSources(goal sharedtypes.PracticeGoal, sourceReportID string) error {
	switch goal {
	case sharedtypes.PracticeGoalBaseline:
		if sourceReportID != "" {
			return validationError("baseline practice plans cannot use a report source", map[string]any{"field": "sourceReportId", "goal": string(goal)})
		}
	case sharedtypes.PracticeGoalRetryCurrentRound, sharedtypes.PracticeGoalNextRound:
		if sourceReportID == "" {
			return validationError("sourceReportId is required for this practice goal", map[string]any{"field": "sourceReportId", "goal": string(goal)})
		}
	}
	return nil
}

func (s *Service) GetPracticePlan(ctx context.Context, userID, planID string) (PlanRecord, error) {
	if s == nil || s.store == nil {
		return PlanRecord{}, fmt.Errorf("practice service is not initialised")
	}
	userID = strings.TrimSpace(userID)
	planID = strings.TrimSpace(planID)
	if userID == "" {
		return PlanRecord{}, fmt.Errorf("userId is required")
	}
	if planID == "" {
		return PlanRecord{}, planNotFoundError()
	}
	plan, err := s.store.GetPlan(ctx, userID, planID)
	if stderrs.Is(err, ErrPlanNotFound) {
		return PlanRecord{}, planNotFoundError()
	}
	if err != nil {
		return PlanRecord{}, err
	}
	return plan, nil
}

func (s *Service) ListPracticeSessions(ctx context.Context, in ListSessionsRequest) (ListSessionsResult, error) {
	if s == nil || s.store == nil {
		return ListSessionsResult{}, fmt.Errorf("practice service is not initialised")
	}
	userID := strings.TrimSpace(in.UserID)
	if userID == "" {
		return ListSessionsResult{}, fmt.Errorf("userId is required")
	}
	if !validSessionStatus(in.Status) {
		return ListSessionsResult{}, validationError("status is invalid", map[string]any{"field": "status"})
	}
	result, err := s.store.ListSessions(ctx, ListSessionsInput{
		UserID:      userID,
		TargetJobID: strings.TrimSpace(in.TargetJobID),
		Status:      in.Status,
		Cursor:      strings.TrimSpace(in.Cursor),
		PageSize:    in.PageSize,
	})
	if stderrs.Is(err, ErrInvalidCursor) {
		return ListSessionsResult{}, validationError("cursor is invalid", map[string]any{"field": "cursor"})
	}
	if err != nil {
		return ListSessionsResult{}, err
	}
	return result, nil
}

func (s *Service) GetPracticeSession(ctx context.Context, userID, sessionID string) (SessionRecord, error) {
	if s == nil || s.store == nil {
		return SessionRecord{}, fmt.Errorf("practice service is not initialised")
	}
	userID = strings.TrimSpace(userID)
	sessionID = strings.TrimSpace(sessionID)
	if userID == "" {
		return SessionRecord{}, fmt.Errorf("userId is required")
	}
	if sessionID == "" {
		return SessionRecord{}, sessionNotFoundError()
	}
	session, err := s.store.GetSession(ctx, userID, sessionID, s.now().UTC())
	if stderrs.Is(err, ErrSessionNotFound) {
		return SessionRecord{}, sessionNotFoundError()
	}
	if err != nil {
		return SessionRecord{}, err
	}
	return session, nil
}

func validationError(message string, details map[string]any) *ServiceError {
	return &ServiceError{Code: sharederrors.CodeValidationFailed, Message: message, Details: details}
}

func planNotFoundError() *ServiceError {
	return &ServiceError{Code: sharederrors.CodePracticePlanNotFound, Message: "practice plan not found"}
}

func sessionNotFoundError() *ServiceError {
	return &ServiceError{Code: sharederrors.CodePracticeSessionNotFound, Message: "practice session not found"}
}

func sessionConflictError() *ServiceError {
	return &ServiceError{Code: sharederrors.CodePracticeSessionConflict, Message: "practice session is in conflicting state"}
}

func idempotencyKeyMismatchError() *ServiceError {
	return &ServiceError{Code: sharederrors.CodeIdempotencyKeyMismatch, Message: "idempotency key was reused with a different request body"}
}

func validPracticeGoal(goal sharedtypes.PracticeGoal) bool {
	for _, allowed := range sharedtypes.AllPracticeGoals {
		if goal == allowed {
			return true
		}
	}
	return false
}

func validInterviewerRole(role sharedtypes.InterviewerRole) bool {
	for _, allowed := range sharedtypes.AllInterviewerRoles {
		if role == allowed {
			return true
		}
	}
	return false
}

func validSessionStatus(status sharedtypes.SessionStatus) bool {
	if status == "" {
		return true
	}
	for _, allowed := range sharedtypes.AllSessionStatuses {
		if status == allowed {
			return true
		}
	}
	return false
}

func validDifficulty(difficulty string) bool {
	switch strings.TrimSpace(difficulty) {
	case "easy", "standard", "stretch":
		return true
	default:
		return false
	}
}

func canonicalPracticeLanguage(language string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "en":
		return "en", true
	case "zh", "zh_cn", "zh-cn":
		return "zh-CN", true
	default:
		return "", false
	}
}
