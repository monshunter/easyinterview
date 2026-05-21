package generators_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

type stubAI struct {
	body []byte
	err  error
}

func (s *stubAI) Complete(ctx context.Context, featureKey string, payload map[string]any) (generators.CompleteResult, error) {
	if s.err != nil {
		return generators.CompleteResult{}, s.err
	}
	return generators.CompleteResult{
		Body:              s.body,
		PromptVersion:     "jd_match_recommendation.v1",
		RubricVersion:     "jd_match_recommendation_rubric.v1",
		ModelProfileName:  "jd_match.recommendation.default",
		Language:          "zh-CN",
		FeatureFlag:       "none",
		DataSourceVersion: "jd_match.v1",
	}, nil
}

type stubStore struct {
	calls    []store.UpsertRecommendationInput
	respond  func(in store.UpsertRecommendationInput) (jdmatch.RecommendationRecord, error)
}

func (s *stubStore) UpsertRecommendation(ctx context.Context, in store.UpsertRecommendationInput) (jdmatch.RecommendationRecord, error) {
	s.calls = append(s.calls, in)
	if s.respond != nil {
		return s.respond(in)
	}
	return jdmatch.RecommendationRecord{ID: in.ID, UserID: in.UserID, Title: in.Title, Score: in.Score}, nil
}

func TestRunRecommendationGeneratorHappyPath(t *testing.T) {
	body := mustJSON(t, []map[string]any{
		{"jobMatchId": "rec-1", "title": "T1", "company": "Acme", "location": "Shanghai", "score": 92, "fit": map[string]int{"must": 4, "total": 5, "plus": 3, "totalPlus": 4}, "reasons": []string{"r1"}, "risks": []string{}, "highlights": []string{}},
		{"jobMatchId": "rec-2", "title": "T2", "company": "Lumen", "location": "Remote", "score": 78, "fit": map[string]int{"must": 3, "total": 5, "plus": 2, "totalPlus": 4}, "reasons": []string{"r2"}, "risks": []string{}, "highlights": []string{}},
	})
	ai := &stubAI{body: body}
	st := &stubStore{}
	res, err := generators.RunRecommendationGenerator(context.Background(), ai, st, generators.RunRecommendationGeneratorInput{
		UserID: "user-A", AgentScanID: "scan-1",
	})
	if err != nil {
		t.Fatalf("RunRecommendationGenerator: %v", err)
	}
	if len(res.Recommendations) != 2 {
		t.Fatalf("len=%d, want 2", len(res.Recommendations))
	}
	if res.CompletedEvent.RecommendationCount != 2 {
		t.Fatalf("recommendationCount = %d", res.CompletedEvent.RecommendationCount)
	}
	if len(st.calls) != 2 {
		t.Fatalf("upsert calls = %d", len(st.calls))
	}
	for _, c := range st.calls {
		if c.PromptVersion == "" || c.ModelID == "" {
			t.Fatalf("provenance missing on upsert: %+v", c)
		}
	}
}

func TestRunRecommendationGeneratorOutputInvalid(t *testing.T) {
	ai := &stubAI{body: []byte(`{"not":"array"}`)}
	st := &stubStore{}
	_, err := generators.RunRecommendationGenerator(context.Background(), ai, st, generators.RunRecommendationGeneratorInput{
		UserID: "user-A", AgentScanID: "scan-1",
	})
	if !errors.Is(err, generators.ErrInvalidLLMOutput) {
		t.Fatalf("err = %v, want ErrInvalidLLMOutput", err)
	}
	if len(st.calls) != 0 {
		t.Fatalf("upsert must not be called on invalid LLM output, got %d calls", len(st.calls))
	}
}

func TestRunRecommendationGeneratorEmptyArrayInvalid(t *testing.T) {
	ai := &stubAI{body: []byte(`[]`)}
	st := &stubStore{}
	_, err := generators.RunRecommendationGenerator(context.Background(), ai, st, generators.RunRecommendationGeneratorInput{
		UserID: "user-A", AgentScanID: "scan-1",
	})
	if !errors.Is(err, generators.ErrInvalidLLMOutput) {
		t.Fatalf("err = %v, want ErrInvalidLLMOutput", err)
	}
}

func TestRunRecommendationGeneratorAIFailure(t *testing.T) {
	ai := &stubAI{err: errors.New("timeout")}
	st := &stubStore{}
	_, err := generators.RunRecommendationGenerator(context.Background(), ai, st, generators.RunRecommendationGeneratorInput{
		UserID: "user-A", AgentScanID: "scan-1",
	})
	if err == nil || errors.Is(err, generators.ErrInvalidLLMOutput) {
		t.Fatalf("err = %v, want non-nil non-ErrInvalidLLMOutput", err)
	}
	if len(st.calls) != 0 {
		t.Fatalf("upsert must not be called on AI failure")
	}
}

func TestRunRecommendationGeneratorRejectsEmpty(t *testing.T) {
	_, err := generators.RunRecommendationGenerator(context.Background(), &stubAI{}, &stubStore{}, generators.RunRecommendationGeneratorInput{})
	if err == nil {
		t.Fatalf("expected error for missing userID / agentScanID")
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return raw
}
