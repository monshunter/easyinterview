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
	ErrClientEventMismatch      = stderrs.New("practice session client event mismatch")
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
	GetSession(ctx context.Context, userID, sessionID string) (SessionRecord, error)
	ReserveSessionEvent(ctx context.Context, in SessionEventReservationInput) (SessionEventReservation, error)
	FinalizeSessionEventError(ctx context.Context, in FinalizeSessionEventErrorInput) error
	AppendSessionEvent(ctx context.Context, in AppendSessionEventStoreInput) (AppendSessionEventResult, error)
	CompleteSession(ctx context.Context, in CompleteSessionStoreInput) (CompleteSessionResult, error)
	ReserveSessionStart(ctx context.Context, in StartSessionReservationInput) (SessionReservation, error)
	CommitSessionStart(ctx context.Context, in CommitSessionStartInput) (SessionRecord, error)
	FailSessionStart(ctx context.Context, in FailSessionStartInput) error
}

type PromptResolver interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type ServiceOptions struct {
	Store         Store
	Registry      PromptResolver
	AI            aiclient.AIClient
	AITaskRuns    aiclient.AITaskRunWriter
	Now           func() time.Time
	NewID         func() string
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
	UserID               string
	TargetJobID          string
	ResumeAssetID        string
	Goal                 sharedtypes.PracticeGoal
	Mode                 sharedtypes.PracticeMode
	InterviewerPersona   sharedtypes.InterviewerRole
	Difficulty           string
	Language             string
	TimeBudgetMinutes    int32
	QuestionBudget       int32
	FocusCompetencyCodes []string
}

type CreatePlanStoreInput struct {
	PlanID               string
	AuditEventID         string
	UserID               string
	TargetJobID          string
	ResumeAssetID        string
	Goal                 sharedtypes.PracticeGoal
	Mode                 sharedtypes.PracticeMode
	InterviewerPersona   sharedtypes.InterviewerRole
	Difficulty           string
	Language             string
	TimeBudgetMinutes    int32
	QuestionBudget       int32
	FocusCompetencyCodes []string
	Now                  time.Time
}

type PlanRecord struct {
	ID                 string
	TargetJobID        string
	Goal               sharedtypes.PracticeGoal
	Mode               sharedtypes.PracticeMode
	InterviewerPersona sharedtypes.InterviewerRole
	Difficulty         string
	Language           string
	TimeBudgetMinutes  int32
	QuestionBudget     int32
	Status             string
	CreatedAt          time.Time
}

func (s *Service) CreatePracticePlan(ctx context.Context, in CreatePlanRequest) (PlanRecord, error) {
	if s == nil || s.store == nil {
		return PlanRecord{}, fmt.Errorf("practice service is not initialised")
	}
	if strings.TrimSpace(in.UserID) == "" {
		return PlanRecord{}, fmt.Errorf("userId is required")
	}
	if strings.TrimSpace(in.TargetJobID) == "" {
		return PlanRecord{}, validationError("targetJobId is required", map[string]any{"field": "targetJobId"})
	}
	if in.Goal != sharedtypes.PracticeGoalBaseline {
		return PlanRecord{}, validationError("practice goal is owned by a future plan", map[string]any{
			"goal":  string(in.Goal),
			"owner": "004-derived-plans-debrief",
		})
	}
	if strings.TrimSpace(in.ResumeAssetID) == "" {
		return PlanRecord{}, validationError("A resume asset must be bound before creating this practice plan.", map[string]any{"field": "resumeAssetId"})
	}
	if !validPracticeMode(in.Mode) {
		return PlanRecord{}, validationError("mode is invalid", map[string]any{"field": "mode"})
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
	if in.QuestionBudget < 1 {
		return PlanRecord{}, validationError("questionBudget must be positive", map[string]any{"field": "questionBudget"})
	}
	if in.TimeBudgetMinutes < 1 {
		return PlanRecord{}, validationError("timeBudgetMinutes must be positive", map[string]any{"field": "timeBudgetMinutes"})
	}

	now := s.now().UTC()
	plan, err := s.store.CreatePlan(ctx, CreatePlanStoreInput{
		PlanID:               s.newID(),
		AuditEventID:         s.newID(),
		UserID:               strings.TrimSpace(in.UserID),
		TargetJobID:          strings.TrimSpace(in.TargetJobID),
		ResumeAssetID:        strings.TrimSpace(in.ResumeAssetID),
		Goal:                 in.Goal,
		Mode:                 in.Mode,
		InterviewerPersona:   in.InterviewerPersona,
		Difficulty:           strings.TrimSpace(in.Difficulty),
		Language:             strings.TrimSpace(in.Language),
		TimeBudgetMinutes:    in.TimeBudgetMinutes,
		QuestionBudget:       in.QuestionBudget,
		FocusCompetencyCodes: append([]string(nil), in.FocusCompetencyCodes...),
		Now:                  now,
	})
	if stderrs.Is(err, ErrPlanPrerequisiteNotFound) {
		return PlanRecord{}, validationError("target job or resume asset is not available", map[string]any{
			"targetJobId":   strings.TrimSpace(in.TargetJobID),
			"resumeAssetId": strings.TrimSpace(in.ResumeAssetID),
		})
	}
	if err != nil {
		return PlanRecord{}, err
	}
	return plan, nil
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
	session, err := s.store.GetSession(ctx, userID, sessionID)
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

func validPracticeMode(mode sharedtypes.PracticeMode) bool {
	for _, allowed := range sharedtypes.AllPracticeModes {
		if mode == allowed {
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

func validDifficulty(difficulty string) bool {
	switch strings.TrimSpace(difficulty) {
	case "easy", "standard", "stretch":
		return true
	default:
		return false
	}
}
