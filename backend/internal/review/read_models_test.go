package review

import (
	"reflect"
	"testing"

	practicedomain "github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestProjectFrozenReportContextProjectionIsMinimalAndLossless(t *testing.T) {
	snapshot := practicedomain.ReportContextSnapshot{
		TargetJob:    practicedomain.ReportTargetJobSnapshot{ID: "target-1", Title: "平台工程师", Company: "Acme", RawJD: "private JD"},
		Resume:       practicedomain.ReportResumeSnapshot{ID: "resume-1", DisplayName: "平台简历", SourceSnapshot: "private resume"},
		Round:        practicedomain.ReportRoundSnapshot{ID: "round-1-technical", Sequence: 1, Name: "技术面", Type: "technical", Focus: "private focus"},
		Plan:         practicedomain.ReportPlanSnapshot{ID: "plan-1", Goal: "private goal"},
		Conversation: practicedomain.ReportConversationCoordinate{SessionID: "session-1", Language: "zh-CN", MessageCount: 3, LastMessageSeqNo: 3},
		HasNextRound: true,
	}
	want := ReportContextProjection{
		SourcePlanID: "plan-1", TargetJobTitle: "平台工程师", TargetJobCompany: "Acme",
		ResumeID: "resume-1", ResumeDisplayName: "平台简历",
		RoundID: "round-1-technical", RoundSequence: 1, RoundName: "技术面", RoundType: "technical",
		Language: "zh-CN", HasNextRound: true,
	}
	if got := ProjectFrozenReportContext(snapshot); !reflect.DeepEqual(got, want) {
		t.Fatalf("projection mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}
