package practice

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestE2EP0047FreezesReportContext(t *testing.T) {
	summary := json.RawMessage(`{
  "interviewRounds": [
    {"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},
    {"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}
  ],
  "provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","featureFlag":"none","dataSourceVersion":"target-job.v1"}
}`)
	snapshot, err := BuildReportContextSnapshot(ReportContextSnapshotInput{
		TargetJob: ReportTargetJobSnapshot{
			ID: "target-1", Title: "Platform Engineer", Company: "Acme", Seniority: "senior",
			Language: "en", RawJD: "complete jd", Summary: summary,
			Requirements: []ReportRequirementSnapshot{{Kind: "must_have", Label: "Go", Description: "production Go", EvidenceLevel: "explicit", DisplayOrder: 1}},
		},
		Resume:       ReportResumeSnapshot{ID: "resume-1", DisplayName: "Backend Resume", Language: "en", SourceSnapshot: "complete resume", StructuredProfile: json.RawMessage(`{"skills":["Go"]}`)},
		Plan:         ReportPlanSnapshot{ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en", TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1, FocusDimensionCodes: []string{"system_design"}},
		Conversation: ReportConversationCoordinate{SessionID: "session-1", Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
	})
	if err != nil {
		t.Fatalf("BuildReportContextSnapshot: %v", err)
	}
	if ReportContextSchemaVersion != "report-context.v1" {
		t.Fatalf("contract schemaVersion=%q", ReportContextSchemaVersion)
	}
	if snapshot.SchemaVersion != ReportContextSchemaVersion {
		t.Fatalf("schemaVersion=%q", snapshot.SchemaVersion)
	}
	if snapshot.Round.ID != "round-1-technical" || snapshot.Round.Focus != "system design" || snapshot.Round.DurationMinutes != 45 {
		t.Fatalf("current round=%+v", snapshot.Round)
	}
	if len(snapshot.CanonicalRounds) != 2 || !snapshot.HasNextRound || snapshot.CanonicalRounds[1].ID != "round-2-manager" {
		t.Fatalf("canonical rounds=%+v hasNext=%v", snapshot.CanonicalRounds, snapshot.HasNextRound)
	}
	if snapshot.TargetJob.RawJD != "complete jd" || snapshot.Resume.SourceSnapshot != "complete resume" || snapshot.Conversation.LastMessageSeqNo != 3 {
		t.Fatalf("snapshot lost full frozen content: %+v", snapshot)
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip ReportContextSnapshot
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReportContextSnapshot(roundTrip); err != nil {
		t.Fatalf("round-trip snapshot invalid: %v", err)
	}
	t.Log("REPORT_CONTEXT_SNAPSHOT_PASS")
}

func TestBuildReportContextSnapshotFailsClosed(t *testing.T) {
	validSummary := json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","dataSourceVersion":"target-job.v1"}}`)
	base := ReportContextSnapshotInput{
		TargetJob:    ReportTargetJobSnapshot{ID: "target-1", Title: "Role", Language: "en", RawJD: "jd", Summary: validSummary},
		Resume:       ReportResumeSnapshot{ID: "resume-1", DisplayName: "Resume", Language: "en", SourceSnapshot: "resume", StructuredProfile: json.RawMessage(`{}`)},
		Plan:         ReportPlanSnapshot{ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en", TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1},
		Conversation: ReportConversationCoordinate{SessionID: "session-1", Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
	}
	tests := []struct {
		name   string
		mutate func(*ReportContextSnapshotInput)
	}{
		{name: "missing raw jd", mutate: func(in *ReportContextSnapshotInput) { in.TargetJob.RawJD = "" }},
		{name: "missing resume", mutate: func(in *ReportContextSnapshotInput) { in.Resume.SourceSnapshot = "" }},
		{name: "wrong resume binding", mutate: func(in *ReportContextSnapshotInput) { in.Plan.ResumeID = "resume-2" }},
		{name: "wrong round pair", mutate: func(in *ReportContextSnapshotInput) { in.Plan.RoundID = "round-2-manager" }},
		{name: "missing provenance", mutate: func(in *ReportContextSnapshotInput) { in.TargetJob.Summary = json.RawMessage(`{"interviewRounds":[]}`) }},
		{name: "no terminal messages", mutate: func(in *ReportContextSnapshotInput) {
			in.Conversation.MessageCount = 0
			in.Conversation.LastMessageSeqNo = 0
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := base
			tc.mutate(&input)
			if _, err := BuildReportContextSnapshot(input); err == nil {
				t.Fatal("expected fail-closed snapshot error")
			}
		})
	}
}

func TestValidateReportContextSnapshotRequiresFrozenCanonicalCatalogConsistency(t *testing.T) {
	validSummary := json.RawMessage(`{"interviewRounds":[{"sequence":1,"type":"technical","name":"Technical","durationMinutes":45,"focus":"system design"},{"sequence":2,"type":"manager","name":"Manager","durationMinutes":30,"focus":"ownership"}],"provenance":{"promptVersion":"v0.1.0","rubricVersion":"v0.1.0","modelId":"fixture-model","language":"en","dataSourceVersion":"target-job.v1"}}`)
	build := func(t *testing.T) ReportContextSnapshot {
		t.Helper()
		snapshot, err := BuildReportContextSnapshot(ReportContextSnapshotInput{
			TargetJob: ReportTargetJobSnapshot{ID: "target-1", Title: "Role", Language: "en", RawJD: "jd", Summary: validSummary},
			Resume:    ReportResumeSnapshot{ID: "resume-1", DisplayName: "Resume", Language: "en", SourceSnapshot: "resume", StructuredProfile: json.RawMessage(`{}`)},
			Plan: ReportPlanSnapshot{
				ID: "plan-1", Goal: "baseline", InterviewerPersona: "hiring_manager", Difficulty: "standard", Language: "en",
				TimeBudgetMinutes: 45, ResumeID: "resume-1", RoundID: "round-1-technical", RoundSequence: 1,
			},
			Conversation: ReportConversationCoordinate{SessionID: "session-1", Language: "en", MessageCount: 3, LastMessageSeqNo: 3},
		})
		if err != nil {
			t.Fatalf("BuildReportContextSnapshot: %v", err)
		}
		return snapshot
	}

	tests := []struct {
		name   string
		mutate func(*ReportContextSnapshot)
		want   string
	}{
		{
			name: "summary and frozen catalog differ by name",
			mutate: func(snapshot *ReportContextSnapshot) {
				snapshot.CanonicalRounds[0].Name = "Changed"
			},
			want: "canonical round catalog",
		},
		{
			name: "summary and frozen catalog differ by focus",
			mutate: func(snapshot *ReportContextSnapshot) {
				snapshot.CanonicalRounds[1].Focus = "changed"
			},
			want: "canonical round catalog",
		},
		{
			name: "current round and frozen catalog differ by duration",
			mutate: func(snapshot *ReportContextSnapshot) {
				snapshot.Round.DurationMinutes++
			},
			want: "current round",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := build(t)
			tc.mutate(&snapshot)
			err := ValidateReportContextSnapshot(snapshot)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ValidateReportContextSnapshot error=%v, want containing %q", err, tc.want)
			}
		})
	}
}
