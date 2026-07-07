package practice

import (
	"context"
	"errors"
	"testing"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestServiceCreatePracticePlanCreatesBaselinePlan(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)
	store := &recordingPlanStore{}
	service := NewService(ServiceOptions{
		Store: store,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("plan-1", "audit-1"),
	})

	plan, err := service.CreatePracticePlan(context.Background(), CreatePlanRequest{
		UserID:               "user-1",
		TargetJobID:          "target-1",
		ResumeID:             "resume-1",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		QuestionBudget:       6,
		TimeBudgetMinutes:    30,
		FocusCompetencyCodes: []string{"communication", "design-systems"},
	})
	if err != nil {
		t.Fatalf("CreatePracticePlan returned error: %v", err)
	}
	if plan.ID != "plan-1" || plan.Status != "ready" || plan.CreatedAt != now {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if store.last.PlanID != "plan-1" || store.last.AuditEventID != "audit-1" {
		t.Fatalf("ids not propagated to store: %+v", store.last)
	}
	if store.last.UserID != "user-1" || store.last.TargetJobID != "target-1" || store.last.ResumeID != "resume-1" {
		t.Fatalf("ownership inputs not propagated: %+v", store.last)
	}
	if store.last.Goal != sharedtypes.PracticeGoalBaseline || store.last.Mode != sharedtypes.PracticeModeAssisted {
		t.Fatalf("plan enum inputs not propagated: %+v", store.last)
	}
}

func TestServiceCreatePracticePlanCreatesDerivedPlans(t *testing.T) {
	tests := []struct {
		name           string
		goal           sharedtypes.PracticeGoal
		sourceReportID string
	}{
		{
			name:           "retry current round report source",
			goal:           sharedtypes.PracticeGoalRetryCurrentRound,
			sourceReportID: "report-1",
		},
		{
			name:           "next round report source",
			goal:           sharedtypes.PracticeGoalNextRound,
			sourceReportID: "report-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingPlanStore{}
			service := NewService(ServiceOptions{Store: store, NewID: sequenceIDs("plan-1", "audit-1")})

			plan, err := service.CreatePracticePlan(context.Background(), validCreatePlanRequest(func(in *CreatePlanRequest) {
				in.Goal = tc.goal
				in.SourceReportID = " " + tc.sourceReportID + " "
			}))
			if err != nil {
				t.Fatalf("CreatePracticePlan returned error: %v", err)
			}
			if plan.Goal != tc.goal || plan.SourceReportID != tc.sourceReportID {
				t.Fatalf("unexpected plan source fields: %+v", plan)
			}
			if store.last.Goal != tc.goal || store.last.SourceReportID != tc.sourceReportID {
				t.Fatalf("store source fields not propagated: %+v", store.last)
			}
		})
	}
}

func TestServiceCreatePracticePlanRejectsInvalidSourceRules(t *testing.T) {
	tests := []struct {
		name           string
		goal           sharedtypes.PracticeGoal
		sourceReportID string
		wantField      string
	}{
		{
			name:           "baseline forbids report source",
			goal:           sharedtypes.PracticeGoalBaseline,
			sourceReportID: "report-1",
			wantField:      "sourceReportId",
		},
		{
			name:      "retry requires report source",
			goal:      sharedtypes.PracticeGoalRetryCurrentRound,
			wantField: "sourceReportId",
		},
		{
			name:      "next round requires report source",
			goal:      sharedtypes.PracticeGoalNextRound,
			wantField: "sourceReportId",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := NewService(ServiceOptions{Store: &recordingPlanStore{}, NewID: sequenceIDs("plan-1", "audit-1")})

			_, err := service.CreatePracticePlan(context.Background(), validCreatePlanRequest(func(in *CreatePlanRequest) {
				in.Goal = tc.goal
				in.SourceReportID = tc.sourceReportID
			}))
			var svcErr *ServiceError
			if !errors.As(err, &svcErr) {
				t.Fatalf("expected ServiceError, got %T: %v", err, err)
			}
			if svcErr.Code != sharederrors.CodeValidationFailed || svcErr.Details["field"] != tc.wantField {
				t.Fatalf("unexpected error for %s: %+v", tc.name, svcErr)
			}
		})
	}
}

func TestServiceCreatePracticePlanRejectsNonCurrentDebriefMode(t *testing.T) {
	service := NewService(ServiceOptions{Store: &recordingPlanStore{}, NewID: sequenceIDs("plan-1", "audit-1")})

	_, err := service.CreatePracticePlan(context.Background(), validCreatePlanRequest(func(in *CreatePlanRequest) {
		in.Mode = sharedtypes.PracticeMode("debrief")
	}))
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodeValidationFailed || svcErr.Details["field"] != "mode" {
		t.Fatalf("unexpected error: %+v", svcErr)
	}
}

func TestServiceCreatePracticePlanRejectsInvalidGoal(t *testing.T) {
	service := NewService(ServiceOptions{Store: &recordingPlanStore{}, NewID: sequenceIDs("plan-1", "audit-1")})

	_, err := service.CreatePracticePlan(context.Background(), validCreatePlanRequest(func(in *CreatePlanRequest) {
		in.Goal = sharedtypes.PracticeGoal("growth")
	}))
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("code = %q", svcErr.Code)
	}
	if svcErr.Details["goal"] != "growth" || svcErr.Details["field"] != "goal" {
		t.Fatalf("unexpected details: %+v", svcErr.Details)
	}
}

func TestServiceCreatePracticePlanRejectsMissingResume(t *testing.T) {
	service := NewService(ServiceOptions{Store: &recordingPlanStore{}, NewID: sequenceIDs("plan-1", "audit-1")})

	_, err := service.CreatePracticePlan(context.Background(), validCreatePlanRequest(func(in *CreatePlanRequest) {
		in.ResumeID = ""
	}))
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != sharederrors.CodeValidationFailed || svcErr.Details["field"] != "resumeId" {
		t.Fatalf("unexpected missing resume error: %+v", svcErr)
	}
}

type recordingPlanStore struct {
	last                       CreatePlanStoreInput
	createErr                  error
	getRecord                  PlanRecord
	getErr                     error
	getUserID                  string
	getPlanID                  string
	listSessionsInput          ListSessionsInput
	listSessionsResult         ListSessionsResult
	listSessionsErr            error
	getSessionRecord           SessionRecord
	getSessionErr              error
	getSessionUserID           string
	getSessionID               string
	eventReservationInput      SessionEventReservationInput
	eventReservation           SessionEventReservation
	eventReserveErr            error
	finalizeEventError         FinalizeSessionEventErrorInput
	finalizeEventErrorErr      error
	appendEvent                AppendSessionEventStoreInput
	appendEventErr             error
	complete                   CompleteSessionStoreInput
	completeResult             CompleteSessionResult
	completeErr                error
	reservation                SessionReservation
	reserveErr                 error
	commit                     CommitSessionStartInput
	commitErr                  error
	fail                       FailSessionStartInput
	failErr                    error
	voiceTurn                  PracticeVoiceTurnStoreInput
	voiceTurnErr               error
	committedContext           CommittedVoiceContext
	loadCommittedContextCalled bool
	steps                      []string
	inTx                       bool
}

func (s *recordingPlanStore) CreatePlan(ctx context.Context, in CreatePlanStoreInput) (PlanRecord, error) {
	s.last = in
	if s.createErr != nil {
		return PlanRecord{}, s.createErr
	}
	return PlanRecord{
		ID:                 in.PlanID,
		TargetJobID:        in.TargetJobID,
		SourceReportID:     in.SourceReportID,
		Goal:               in.Goal,
		Mode:               in.Mode,
		InterviewerPersona: in.InterviewerPersona,
		Difficulty:         in.Difficulty,
		Language:           in.Language,
		TimeBudgetMinutes:  in.TimeBudgetMinutes,
		QuestionBudget:     in.QuestionBudget,
		Status:             "ready",
		CreatedAt:          in.Now,
	}, nil
}

func (s *recordingPlanStore) GetPlan(ctx context.Context, userID, planID string) (PlanRecord, error) {
	s.getUserID = userID
	s.getPlanID = planID
	if s.getErr != nil {
		return PlanRecord{}, s.getErr
	}
	return s.getRecord, nil
}

func (s *recordingPlanStore) ListSessions(ctx context.Context, in ListSessionsInput) (ListSessionsResult, error) {
	s.listSessionsInput = in
	if s.listSessionsErr != nil {
		return ListSessionsResult{}, s.listSessionsErr
	}
	return s.listSessionsResult, nil
}

func (s *recordingPlanStore) GetSession(ctx context.Context, userID, sessionID string) (SessionRecord, error) {
	s.getSessionUserID = userID
	s.getSessionID = sessionID
	if s.getSessionErr != nil {
		return SessionRecord{}, s.getSessionErr
	}
	return s.getSessionRecord, nil
}

func (s *recordingPlanStore) ReserveSessionEvent(ctx context.Context, in SessionEventReservationInput) (SessionEventReservation, error) {
	s.steps = append(s.steps, "reserve-event")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.eventReservationInput = in
	if s.eventReserveErr != nil {
		return SessionEventReservation{}, s.eventReserveErr
	}
	s.eventReservation.UserID = in.UserID
	return s.eventReservation, nil
}

func (s *recordingPlanStore) FinalizeSessionEventError(ctx context.Context, in FinalizeSessionEventErrorInput) error {
	s.steps = append(s.steps, "finalize-event-error")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.finalizeEventError = in
	return s.finalizeEventErrorErr
}

func (s *recordingPlanStore) AppendSessionEvent(ctx context.Context, in AppendSessionEventStoreInput) (AppendSessionEventResult, error) {
	s.steps = append(s.steps, "append-event")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.appendEvent = in
	if s.appendEventErr != nil {
		return AppendSessionEventResult{}, s.appendEventErr
	}
	session := s.eventReservation.Session
	session.Status = in.Outcome.NextSessionStatus
	session.UpdatedAt = in.OccurredAt
	if in.NextQuestion != nil {
		session.CurrentTurn = in.NextQuestion
		session.TurnCount = in.NextQuestion.TurnIndex
	} else if in.Outcome.NextTurn != nil {
		next := *in.Outcome.NextTurn
		session.CurrentTurn = &next
	}
	return AppendSessionEventResult{
		Acknowledged:    in.Outcome.Acknowledged,
		Session:         session,
		AssistantAction: in.Outcome.AssistantAction,
	}, nil
}

func (s *recordingPlanStore) CompleteSession(ctx context.Context, in CompleteSessionStoreInput) (CompleteSessionResult, error) {
	s.steps = append(s.steps, "complete")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.complete = in
	if s.completeErr != nil {
		return CompleteSessionResult{}, s.completeErr
	}
	if s.completeResult.ReportID != "" {
		return s.completeResult, nil
	}
	return CompleteSessionResult{
		ReportID: in.ReportID,
		Job: JobRecord{
			ID:           in.JobID,
			JobType:      "report_generate",
			ResourceType: "feedback_report",
			ResourceID:   in.ReportID,
			Status:       sharedtypes.JobStatusQueued,
			CreatedAt:    in.Now,
			UpdatedAt:    in.Now,
		},
	}, nil
}

func (s *recordingPlanStore) ReserveSessionStart(ctx context.Context, in StartSessionReservationInput) (SessionReservation, error) {
	s.steps = append(s.steps, "reserve")
	s.inTx = true
	defer func() { s.inTx = false }()
	if s.reserveErr != nil {
		return SessionReservation{}, s.reserveErr
	}
	if s.reservation.SessionID == "" {
		s.reservation.SessionID = in.SessionID
	}
	s.reservation.IdempotencyRecordID = in.IdempotencyRecordID
	s.reservation.UserID = in.UserID
	s.reservation.HintsEnabled = in.HintsEnabled
	return s.reservation, nil
}

func (s *recordingPlanStore) CommitSessionStart(ctx context.Context, in CommitSessionStartInput) (SessionRecord, error) {
	s.steps = append(s.steps, "commit")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.commit = in
	if s.commitErr != nil {
		return SessionRecord{}, s.commitErr
	}
	return SessionRecord{
		ID:           in.SessionID,
		PlanID:       in.PlanID,
		TargetJobID:  in.TargetJobID,
		Status:       sharedtypes.SessionStatusRunning,
		Language:     in.Language,
		HintsEnabled: in.HintsEnabled,
		TurnCount:    1,
		CurrentTurn: &TurnRecord{
			ID:             in.TurnID,
			TurnIndex:      1,
			QuestionText:   in.QuestionText,
			QuestionIntent: in.QuestionIntent,
			Status:         "asked",
			AskedAt:        in.StartedAt,
		},
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.StartedAt,
	}, nil
}

func (s *recordingPlanStore) FailSessionStart(ctx context.Context, in FailSessionStartInput) error {
	s.steps = append(s.steps, "fail")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.fail = in
	return s.failErr
}

func (s *recordingPlanStore) RecordPracticeVoiceTurn(ctx context.Context, in PracticeVoiceTurnStoreInput) (SessionRecord, error) {
	s.steps = append(s.steps, "record-voice-turn")
	s.inTx = true
	defer func() { s.inTx = false }()
	s.voiceTurn = in
	if s.voiceTurnErr != nil {
		return SessionRecord{}, s.voiceTurnErr
	}
	session := s.getSessionRecord
	if in.Session.ID != "" {
		session = in.Session
	}
	session.Status = sharedtypes.SessionStatusRunning
	session.UpdatedAt = in.OccurredAt
	if session.CurrentTurn != nil {
		turn := *session.CurrentTurn
		turn.Status = string(TurnStatusFollowUpRequested)
		turn.FollowUpCount = 1
		session.CurrentTurn = &turn
	}
	return session, nil
}

func (s *recordingPlanStore) LoadCommittedVoiceContext(ctx context.Context, userID, sessionID string) (CommittedVoiceContext, error) {
	s.loadCommittedContextCalled = true
	return s.committedContext, nil
}

func validCreatePlanRequest(mutators ...func(*CreatePlanRequest)) CreatePlanRequest {
	in := CreatePlanRequest{
		UserID:               "user-1",
		TargetJobID:          "target-1",
		ResumeID:             "resume-1",
		Goal:                 sharedtypes.PracticeGoalBaseline,
		Mode:                 sharedtypes.PracticeModeAssisted,
		InterviewerPersona:   sharedtypes.InterviewerRoleHiringManager,
		Difficulty:           "standard",
		Language:             "zh-CN",
		QuestionBudget:       6,
		TimeBudgetMinutes:    30,
		FocusCompetencyCodes: []string{"communication", "design-systems"},
	}
	for _, mutate := range mutators {
		mutate(&in)
	}
	return in
}

func sequenceIDs(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return "extra-id"
		}
		id := ids[i]
		i++
		return id
	}
}
