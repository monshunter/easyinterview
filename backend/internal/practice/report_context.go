package practice

import (
	"encoding/json"
	"fmt"
	"strings"
)

const ReportContextSchemaVersion = "report-context.v1"

type ReportRequirementSnapshot struct {
	Kind          string `json:"kind"`
	Label         string `json:"label"`
	Description   string `json:"description,omitempty"`
	EvidenceLevel string `json:"evidenceLevel"`
	DisplayOrder  int32  `json:"displayOrder"`
}

type ReportTargetJobSnapshot struct {
	ID           string                      `json:"id"`
	Title        string                      `json:"title"`
	Company      string                      `json:"company"`
	Seniority    string                      `json:"seniority,omitempty"`
	Language     string                      `json:"language"`
	RawJD        string                      `json:"rawJd"`
	Summary      json.RawMessage             `json:"summary"`
	Requirements []ReportRequirementSnapshot `json:"requirements"`
}

type ReportResumeSnapshot struct {
	ID                string          `json:"id"`
	DisplayName       string          `json:"displayName"`
	Language          string          `json:"language"`
	SourceSnapshot    string          `json:"sourceSnapshot"`
	StructuredProfile json.RawMessage `json:"structuredProfile"`
}

type ReportRoundSnapshot struct {
	ID              string `json:"id"`
	Sequence        int32  `json:"sequence"`
	Type            string `json:"type"`
	Name            string `json:"name"`
	Focus           string `json:"focus"`
	DurationMinutes int32  `json:"durationMinutes"`
}

type ReportPlanSnapshot struct {
	ID                  string   `json:"id"`
	Goal                string   `json:"goal"`
	InterviewerPersona  string   `json:"interviewerPersona"`
	Difficulty          string   `json:"difficulty"`
	Language            string   `json:"language"`
	TimeBudgetMinutes   int32    `json:"timeBudgetMinutes"`
	ResumeID            string   `json:"resumeId"`
	RoundID             string   `json:"roundId"`
	RoundSequence       int32    `json:"roundSequence"`
	FocusDimensionCodes []string `json:"focusDimensionCodes"`
}

type ReportConversationCoordinate struct {
	SessionID        string `json:"sessionId"`
	Language         string `json:"language"`
	MessageCount     int32  `json:"messageCount"`
	LastMessageSeqNo int32  `json:"lastMessageSeqNo"`
}

type ReportContextSnapshot struct {
	SchemaVersion   string                       `json:"schemaVersion"`
	TargetJob       ReportTargetJobSnapshot      `json:"targetJob"`
	Resume          ReportResumeSnapshot         `json:"resume"`
	Round           ReportRoundSnapshot          `json:"round"`
	CanonicalRounds []ReportRoundSnapshot        `json:"canonicalRounds"`
	Plan            ReportPlanSnapshot           `json:"plan"`
	Conversation    ReportConversationCoordinate `json:"conversation"`
	HasNextRound    bool                         `json:"hasNextRound"`
}

type ReportContextSnapshotInput struct {
	TargetJob    ReportTargetJobSnapshot
	Resume       ReportResumeSnapshot
	Plan         ReportPlanSnapshot
	Conversation ReportConversationCoordinate
}

type reportTargetSummary struct {
	InterviewRounds []struct {
		Sequence        int32  `json:"sequence"`
		Type            string `json:"type"`
		Name            string `json:"name"`
		Focus           string `json:"focus"`
		DurationMinutes int32  `json:"durationMinutes"`
	} `json:"interviewRounds"`
	Provenance struct {
		PromptVersion     string `json:"promptVersion"`
		RubricVersion     string `json:"rubricVersion"`
		ModelID           string `json:"modelId"`
		Language          string `json:"language"`
		DataSourceVersion string `json:"dataSourceVersion"`
	} `json:"provenance"`
}

func BuildReportContextSnapshot(in ReportContextSnapshotInput) (ReportContextSnapshot, error) {
	target := in.TargetJob
	target.ID = strings.TrimSpace(target.ID)
	target.Title = strings.TrimSpace(target.Title)
	target.Company = strings.TrimSpace(target.Company)
	target.Seniority = strings.TrimSpace(target.Seniority)
	target.Language = strings.TrimSpace(target.Language)
	target.RawJD = strings.TrimSpace(target.RawJD)
	target.Summary = cloneRawMessage(target.Summary)
	target.Requirements = normalizeRequirements(target.Requirements)

	resume := in.Resume
	resume.ID = strings.TrimSpace(resume.ID)
	resume.DisplayName = strings.TrimSpace(resume.DisplayName)
	resume.Language = strings.TrimSpace(resume.Language)
	resume.SourceSnapshot = strings.TrimSpace(resume.SourceSnapshot)
	resume.StructuredProfile = cloneRawMessage(resume.StructuredProfile)

	plan := in.Plan
	plan.ID = strings.TrimSpace(plan.ID)
	plan.Goal = strings.TrimSpace(plan.Goal)
	plan.InterviewerPersona = strings.TrimSpace(plan.InterviewerPersona)
	plan.Difficulty = strings.TrimSpace(plan.Difficulty)
	plan.Language = strings.TrimSpace(plan.Language)
	plan.ResumeID = strings.TrimSpace(plan.ResumeID)
	plan.RoundID = strings.TrimSpace(plan.RoundID)
	plan.FocusDimensionCodes = normalizeStringSet(plan.FocusDimensionCodes)

	conversation := in.Conversation
	conversation.SessionID = strings.TrimSpace(conversation.SessionID)
	conversation.Language = strings.TrimSpace(conversation.Language)

	rounds, err := canonicalReportRounds(target.Summary)
	if err != nil {
		return ReportContextSnapshot{}, err
	}
	currentIndex := -1
	for i, round := range rounds {
		if round.ID == plan.RoundID && round.Sequence == plan.RoundSequence {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		return ReportContextSnapshot{}, fmt.Errorf("report context round does not match canonical plan identity")
	}

	out := ReportContextSnapshot{
		SchemaVersion:   ReportContextSchemaVersion,
		TargetJob:       target,
		Resume:          resume,
		Round:           rounds[currentIndex],
		CanonicalRounds: rounds,
		Plan:            plan,
		Conversation:    conversation,
		HasNextRound:    currentIndex+1 < len(rounds),
	}
	if err := ValidateReportContextSnapshot(out); err != nil {
		return ReportContextSnapshot{}, err
	}
	return out, nil
}

func ValidateReportContextSnapshot(snapshot ReportContextSnapshot) error {
	if snapshot.SchemaVersion != ReportContextSchemaVersion {
		return fmt.Errorf("report context schemaVersion must be %s", ReportContextSchemaVersion)
	}
	if snapshot.TargetJob.ID == "" || snapshot.TargetJob.Title == "" || snapshot.TargetJob.Language == "" || snapshot.TargetJob.RawJD == "" {
		return fmt.Errorf("report context target job is incomplete")
	}
	if len(snapshot.TargetJob.Summary) == 0 || !json.Valid(snapshot.TargetJob.Summary) {
		return fmt.Errorf("report context target summary is invalid")
	}
	if snapshot.Resume.ID == "" || snapshot.Resume.DisplayName == "" || snapshot.Resume.Language == "" || snapshot.Resume.SourceSnapshot == "" {
		return fmt.Errorf("report context resume is incomplete")
	}
	if len(snapshot.Resume.StructuredProfile) == 0 || !json.Valid(snapshot.Resume.StructuredProfile) {
		return fmt.Errorf("report context resume profile is invalid")
	}
	if snapshot.Plan.ID == "" || snapshot.Plan.ResumeID != snapshot.Resume.ID || snapshot.Plan.RoundID == "" || snapshot.Plan.RoundSequence <= 0 || snapshot.Plan.TimeBudgetMinutes <= 0 {
		return fmt.Errorf("report context plan binding is invalid")
	}
	if snapshot.Plan.Language == "" || snapshot.Plan.Language != snapshot.Conversation.Language {
		return fmt.Errorf("report context session language mismatch")
	}
	if snapshot.Conversation.SessionID == "" || snapshot.Conversation.MessageCount <= 0 || snapshot.Conversation.LastMessageSeqNo <= 0 {
		return fmt.Errorf("report context terminal conversation coordinate is invalid")
	}
	if len(snapshot.CanonicalRounds) < 2 || len(snapshot.CanonicalRounds) > 5 {
		return fmt.Errorf("report context canonical round catalog must contain 2 to 5 rounds")
	}
	currentIndex := -1
	for i, round := range snapshot.CanonicalRounds {
		if round.ID == snapshot.Round.ID && round.Sequence == snapshot.Round.Sequence {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 || snapshot.Round.ID != snapshot.Plan.RoundID || snapshot.Round.Sequence != snapshot.Plan.RoundSequence {
		return fmt.Errorf("report context current round binding is invalid")
	}
	if snapshot.HasNextRound != (currentIndex+1 < len(snapshot.CanonicalRounds)) {
		return fmt.Errorf("report context next-round projection is invalid")
	}
	return nil
}

func canonicalReportRounds(raw json.RawMessage) ([]ReportRoundSnapshot, error) {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil, fmt.Errorf("report context target summary is invalid")
	}
	var summary reportTargetSummary
	if err := json.Unmarshal(raw, &summary); err != nil {
		return nil, fmt.Errorf("decode report context target summary: %w", err)
	}
	if strings.TrimSpace(summary.Provenance.PromptVersion) == "" || strings.TrimSpace(summary.Provenance.RubricVersion) == "" || strings.TrimSpace(summary.Provenance.ModelID) == "" || strings.TrimSpace(summary.Provenance.Language) == "" || strings.TrimSpace(summary.Provenance.DataSourceVersion) == "" {
		return nil, fmt.Errorf("report context target summary provenance is incomplete")
	}
	if len(summary.InterviewRounds) < 2 || len(summary.InterviewRounds) > 5 {
		return nil, fmt.Errorf("report context target summary must contain 2 to 5 interview rounds")
	}
	allowedTypes := map[string]struct{}{"hr": {}, "technical": {}, "manager": {}, "cross_functional": {}, "culture": {}, "final": {}, "other": {}}
	rounds := make([]ReportRoundSnapshot, 0, len(summary.InterviewRounds))
	var previousSequence int32
	for _, rawRound := range summary.InterviewRounds {
		roundType := strings.TrimSpace(rawRound.Type)
		name := strings.TrimSpace(rawRound.Name)
		focus := strings.TrimSpace(rawRound.Focus)
		if rawRound.Sequence <= previousSequence {
			return nil, fmt.Errorf("report context round sequence must be strictly increasing")
		}
		if _, ok := allowedTypes[roundType]; !ok || name == "" || focus == "" || rawRound.DurationMinutes < 10 || rawRound.DurationMinutes > 180 {
			return nil, fmt.Errorf("report context round is not canonical")
		}
		rounds = append(rounds, ReportRoundSnapshot{
			ID: fmt.Sprintf("round-%d-%s", rawRound.Sequence, roundType), Sequence: rawRound.Sequence,
			Type: roundType, Name: name, Focus: focus, DurationMinutes: rawRound.DurationMinutes,
		})
		previousSequence = rawRound.Sequence
	}
	return rounds, nil
}

func normalizeRequirements(in []ReportRequirementSnapshot) []ReportRequirementSnapshot {
	out := make([]ReportRequirementSnapshot, 0, len(in))
	for _, requirement := range in {
		requirement.Kind = strings.TrimSpace(requirement.Kind)
		requirement.Label = strings.TrimSpace(requirement.Label)
		requirement.Description = strings.TrimSpace(requirement.Description)
		requirement.EvidenceLevel = strings.TrimSpace(requirement.EvidenceLevel)
		out = append(out, requirement)
	}
	return out
}

func normalizeStringSet(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, value := range in {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cloneRawMessage(in json.RawMessage) json.RawMessage {
	if len(in) == 0 {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), in...)
}
