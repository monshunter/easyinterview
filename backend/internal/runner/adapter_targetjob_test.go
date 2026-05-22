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
	legacy := &recordingTargetjobHandler{
		outcome: targetjob.JobOutcome{
			Succeeded:    false,
			Retryable:    true,
			ErrorCode:    "TRANSIENT",
			ErrorMessage: "retry me",
		},
	}
	adapted := FromTargetjobHandler(legacy)
	got := adapted.Handle(context.Background(), ClaimedJob{
		JobID:       "job-1",
		JobType:     "target_import",
		ResourceID:  "res-1",
		Payload:     []byte(`{"k":"v"}`),
		Attempts:    2,
		MaxAttempts: 5,
	})
	if legacy.saw.JobID != "job-1" || legacy.saw.JobType != "target_import" {
		t.Fatalf("legacy handler saw %+v, want job-1/target_import", legacy.saw)
	}
	if legacy.saw.Attempts != 2 || string(legacy.saw.Payload) != `{"k":"v"}` {
		t.Fatalf("claimed-job fields not preserved: %+v", legacy.saw)
	}
	want := JobOutcome{Succeeded: false, Retryable: true, ErrorCode: "TRANSIENT", ErrorMessage: "retry me"}
	if got != want {
		t.Fatalf("outcome = %+v, want %+v", got, want)
	}
}
