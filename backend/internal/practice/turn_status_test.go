package practice

import "testing"

func TestTurnStatusCurrentValues(t *testing.T) {
	tests := []struct {
		name string
		got  TurnStatus
		want string
	}{
		{name: "asked", got: TurnStatusAsked, want: "asked"},
		{name: "answered", got: TurnStatusAnswered, want: "answered"},
		{name: "follow up requested", got: TurnStatusFollowUpRequested, want: "follow_up_requested"},
		{name: "assessed", got: TurnStatusAssessed, want: "assessed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.got); got != tt.want {
				t.Fatalf("TurnStatus = %q, want %q", got, tt.want)
			}
		})
	}
}
