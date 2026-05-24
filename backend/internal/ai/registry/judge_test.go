package registry

import (
	"context"
	"errors"
	"testing"
)

func TestNotImplementedJudgeAlwaysFails(t *testing.T) {
	t.Parallel()
	var j Judge = NotImplementedJudge{}
	scores, reasoning, err := j.Judge(
		context.Background(),
		"practice.session.first_question",
		"v0.1.0",
		[]byte("{\"foo\":\"bar\"}"),
		"v0.1.0",
	)
	if !errors.Is(err, ErrJudgeNotImplemented) {
		t.Fatalf("want ErrJudgeNotImplemented, got %v", err)
	}
	if scores != nil {
		t.Errorf("Scores must be nil on fail-closed, got %+v", scores)
	}
	if reasoning.Summary != "" || len(reasoning.EvidenceQuotes) != 0 {
		t.Errorf("Reasoning must be zero value, got %+v", reasoning)
	}
}
