package registry

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestFailClosedJudgeAlwaysFails(t *testing.T) {
	t.Parallel()
	var j Judge = FailClosedJudge{}
	scores, reasoning, err := j.Judge(
		context.Background(),
		"practice.session.first_question",
		"v0.1.0",
		[]byte("{\"foo\":\"bar\"}"),
		"v0.1.0",
	)
	if !errors.Is(err, ErrJudgeUnavailable) {
		t.Fatalf("want ErrJudgeUnavailable, got %v", err)
	}
	if strings.Contains(err.Error(), "plan ") {
		t.Fatalf("ErrJudgeUnavailable must not expose implementation plan wording, got %q", err.Error())
	}
	if scores != nil {
		t.Errorf("Scores must be nil on fail-closed, got %+v", scores)
	}
	if reasoning.Summary != "" || len(reasoning.EvidenceQuotes) != 0 {
		t.Errorf("Reasoning must be zero value, got %+v", reasoning)
	}
}
