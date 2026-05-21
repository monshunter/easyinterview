// Package generators hosts the per-feature AI generation services
// invoked inline by the agent_scan / search handlers. The generators
// are NOT canonical job types (spec D-12 only registers
// jd_match_agent_scan + jd_match_search). They expose a plain Go
// function the caller composes into its own job execution loop.
package generators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
)

// AIClient is the slice of the A3 AIClient surface this generator
// needs. cmd/api wiring constructs a stub for tests and a real
// adapter for production.
type AIClient interface {
	// Complete invokes the supplied feature_key (F3 routing key) and
	// returns the raw model response body (JSON in this contract).
	Complete(ctx context.Context, featureKey string, payload map[string]any) (CompleteResult, error)
}

// CompleteResult captures the AI provenance the generator stamps onto
// each recommendation row. Provider/model identifiers stay opaque;
// the registry/profile names are the public contract per A3 D-1.
type CompleteResult struct {
	Body              json.RawMessage
	PromptVersion     string
	RubricVersion     string
	ModelProfileName  string
	Language          string
	FeatureFlag       string
	DataSourceVersion string
}

// RecommendationUpserter is the slice of the store layer the
// generator writes to. The full Repository satisfies it.
type RecommendationUpserter interface {
	UpsertRecommendation(ctx context.Context, in store.UpsertRecommendationInput) (jdmatch.RecommendationRecord, error)
}

// RunRecommendationGeneratorInput captures the agent_scan job's
// hand-off shape. CandidateProfileJSON / JobsPoolJSON are caller-owned
// JSON payloads carrying the context the generator forwards to the
// LLM; they never leak through to logs because the AIClient owns
// redaction (spec D-9 / F3 D-10).
type RunRecommendationGeneratorInput struct {
	UserID               string
	AgentScanID          string
	CandidateProfileJSON json.RawMessage
	JobsPoolJSON         json.RawMessage
}

// RunRecommendationGeneratorResult bundles the upserted recommendations
// and the completion event payload. The caller (agent_scan job
// handler) emits the outbox event after a successful return.
type RunRecommendationGeneratorResult struct {
	Recommendations []jdmatch.RecommendationRecord
	CompletedEvent  RecommendationCompletedEvent
}

// RecommendationCompletedEvent mirrors the
// jd_match.recommendation.completed envelope payload (B3 D-16 PII
// boundary: only userId / agentScanId / count / completedAt).
type RecommendationCompletedEvent struct {
	UserID              string `json:"userId"`
	AgentScanID         string `json:"agentScanId"`
	RecommendationCount int    `json:"recommendationCount"`
}

// llmRecommendation is the per-item JSON the LLM is contracted to
// return (see config/prompts/jd_match.recommendation/v0.1.0.md).
type llmRecommendation struct {
	JobMatchID          string   `json:"jobMatchId"`
	Title               string   `json:"title"`
	Company             string   `json:"company"`
	CompanyTag          *string  `json:"companyTag,omitempty"`
	Level               *string  `json:"level,omitempty"`
	Location            string   `json:"location"`
	Comp                *string  `json:"comp,omitempty"`
	Posted              *string  `json:"posted,omitempty"`
	Score               int      `json:"score"`
	Fit                 fitTuple `json:"fit"`
	Reasons             []string `json:"reasons"`
	Risks               []string `json:"risks"`
	Highlights          []string `json:"highlights"`
	SourceURL           *string  `json:"sourceUrl,omitempty"`
	SourceLabel         *string  `json:"sourceLabel,omitempty"`
	NetworkNote         *string  `json:"networkNote,omitempty"`
	SimilarInterviewers *int     `json:"similarInterviewers,omitempty"`
	InterviewHypotheses []string `json:"interviewHypotheses,omitempty"`
}

type fitTuple struct {
	Must      int `json:"must"`
	Total     int `json:"total"`
	Plus      int `json:"plus"`
	TotalPlus int `json:"totalPlus"`
}

// ErrInvalidLLMOutput indicates the model returned a payload that
// does not satisfy the per-item contract (missing required fields,
// wrong JSON shape, etc.). The agent_scan job maps this to a failed
// ai_task_runs row and does not emit the completed event.
var ErrInvalidLLMOutput = errors.New("generators: invalid LLM output")

// RunRecommendationGenerator drives the JD-Match recommendation
// flow per spec §4.2 / D-12. Failure paths return an error without
// upserting partial rows.
func RunRecommendationGenerator(
	ctx context.Context,
	ai AIClient,
	store RecommendationUpserter,
	in RunRecommendationGeneratorInput,
) (RunRecommendationGeneratorResult, error) {
	if ai == nil || store == nil {
		return RunRecommendationGeneratorResult{}, errors.New("generators: ai or store dependency is nil")
	}
	if strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.AgentScanID) == "" {
		return RunRecommendationGeneratorResult{}, errors.New("generators: userID and agentScanID are required")
	}
	payload := map[string]any{
		"candidateProfile": in.CandidateProfileJSON,
		"jobsPool":         in.JobsPoolJSON,
	}
	res, err := ai.Complete(ctx, featurekeys.JdMatchRecommendation.String(), payload)
	if err != nil {
		return RunRecommendationGeneratorResult{}, fmt.Errorf("generators: ai call failed: %w", err)
	}
	if len(res.Body) == 0 {
		return RunRecommendationGeneratorResult{}, ErrInvalidLLMOutput
	}
	var items []llmRecommendation
	if err := json.Unmarshal(res.Body, &items); err != nil {
		return RunRecommendationGeneratorResult{}, fmt.Errorf("%w: %v", ErrInvalidLLMOutput, err)
	}
	if len(items) == 0 {
		return RunRecommendationGeneratorResult{}, ErrInvalidLLMOutput
	}
	upserted := make([]jdmatch.RecommendationRecord, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.JobMatchID) == "" || strings.TrimSpace(it.Title) == "" || strings.TrimSpace(it.Company) == "" {
			return RunRecommendationGeneratorResult{}, ErrInvalidLLMOutput
		}
		rec, err := store.UpsertRecommendation(ctx, mapLLMToUpsert(in.UserID, it, res))
		if err != nil {
			return RunRecommendationGeneratorResult{}, fmt.Errorf("generators: upsert: %w", err)
		}
		upserted = append(upserted, rec)
	}
	return RunRecommendationGeneratorResult{
		Recommendations: upserted,
		CompletedEvent: RecommendationCompletedEvent{
			UserID:              in.UserID,
			AgentScanID:         in.AgentScanID,
			RecommendationCount: len(upserted),
		},
	}, nil
}

func mapLLMToUpsert(userID string, it llmRecommendation, res CompleteResult) store.UpsertRecommendationInput {
	return store.UpsertRecommendationInput{
		ID:                  it.JobMatchID,
		UserID:              userID,
		Title:               it.Title,
		Company:             it.Company,
		CompanyTag:          it.CompanyTag,
		Level:               it.Level,
		Location:            it.Location,
		Comp:                it.Comp,
		PostedLabel:         it.Posted,
		Score:               it.Score,
		FitMust:             it.Fit.Must,
		FitTotal:            it.Fit.Total,
		FitPlus:             it.Fit.Plus,
		FitTotalPlus:        it.Fit.TotalPlus,
		Reasons:             it.Reasons,
		Risks:               it.Risks,
		Highlights:          it.Highlights,
		SourceURL:           it.SourceURL,
		SourceLabel:         it.SourceLabel,
		NetworkNote:         it.NetworkNote,
		SimilarInterviewers: it.SimilarInterviewers,
		InterviewHypotheses: it.InterviewHypotheses,
		PromptVersion:       res.PromptVersion,
		RubricVersion:       res.RubricVersion,
		ModelID:             res.ModelProfileName,
		Language:            firstNonEmpty(res.Language, "zh-CN"),
		FeatureFlag:         firstNonEmpty(res.FeatureFlag, "none"),
		DataSourceVersion:   firstNonEmpty(res.DataSourceVersion, "jd_match.v1"),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
