package runner

import (
	"context"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

type recordingTargetjobHandler struct {
	saw     targetjob.ClaimedJob
	outcome targetjob.JobOutcome
}

func (h *recordingTargetjobHandler) Handle(_ context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
	h.saw = job
	return h.outcome
}

func TestFromTargetjobHandler_PreservesOutcome(t *testing.T) {
	nonCurrent := &recordingTargetjobHandler{
		outcome: targetjob.JobOutcome{
			Succeeded:    false,
			Retryable:    true,
			ErrorCode:    "TRANSIENT",
			ErrorMessage: "retry me",
		},
	}
	adapted := FromTargetjobHandler(nonCurrent)
	got := adapted.Handle(context.Background(), ClaimedJob{
		JobID:       "job-1",
		JobType:     "target_import",
		ResourceID:  "res-1",
		Payload:     []byte(`{"k":"v"}`),
		Attempts:    2,
		MaxAttempts: 5,
	})
	if nonCurrent.saw.JobID != "job-1" || nonCurrent.saw.JobType != "target_import" {
		t.Fatalf("targetjob handler saw %+v, want job-1/target_import", nonCurrent.saw)
	}
	if nonCurrent.saw.Attempts != 2 || string(nonCurrent.saw.Payload) != `{"k":"v"}` {
		t.Fatalf("claimed-job fields not preserved: %+v", nonCurrent.saw)
	}
	want := JobOutcome{Succeeded: false, Retryable: true, ErrorCode: "TRANSIENT", ErrorMessage: "retry me"}
	if got != want {
		t.Fatalf("outcome = %+v, want %+v", got, want)
	}
}
