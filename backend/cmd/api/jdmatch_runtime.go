package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/debrief"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	jdmatchhandler "github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
	jdmatchstore "github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/practice"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	profilestore "github.com/monshunter/easyinterview/backend/internal/profile/store"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// jdmatchRoutes bundles the JD-Match HTTP handler, idempotency
// middleware, and the agent_scan drainer entry that cmd/api wires up
// for the 12 JobMatch routes plus the single jd_match_agent_scan
// background job (spec D-12).
type jdmatchRoutes struct {
	Handler           *jdmatchhandler.Handler
	Idempotency       *idempotency.Middleware
	AgentScanRunOnce  func(ctx context.Context, userID string) error
	PrivacyDeleteFunc func(ctx context.Context, userID string) (service.PrivacyDeleteCounts, error)
}

// buildJDMatchRoutes composes the backend-jobs-recommendations
// runtime: store layer + cross-owner counters / identity / candidate
// profile + market signals + agent_scan job + privacy delete. AI is
// passed in by the caller so the same stub used elsewhere in cmd/api
// tests covers JD-Match recommendation / search too.
func buildJDMatchRoutes(loader *config.Loader, db *sql.DB, jdMatchAI jdmatchAI) jdmatchRoutes {
	repo := jdmatchstore.NewRepository(db, func() time.Time { return time.Now().UTC() })
	profileRepo := profilestore.NewRepository(db)
	h := jdmatchhandler.New(jdmatchhandler.Options{
		Session:    currentUserFromContext,
		AgentScans: repo,
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.BuildJobMatchProfile(ctx, userID, service.ProfileDeps{
				GetUserIdentity: func(ctx context.Context, userID string) (auth.UserIdentity, error) {
					return auth.GetUserIdentityForUser(ctx, db, userID)
				},
				GetCandidateProfile: func(ctx context.Context, userID string) (*api.CandidateProfile, error) {
					rec, err := profileRepo.GetCandidateProfileByUser(ctx, userID)
					if err != nil || rec == nil {
						return nil, err
					}
					cp := api.CandidateProfile{}
					cp.Headline = rec.Headline
					cp.YearsOfExperience = rec.YearsOfExperience
					cp.PreferredPracticeLanguage = rec.PreferredPracticeLanguage
					cp.UiLanguage = rec.UiLanguage
					return &cp, nil
				},
				CountExperienceCardsBySource: profileRepo.CountExperienceCardsBySource,
				CountResumes: func(ctx context.Context, userID string) (int, error) {
					return resume.CountResumesForUser(ctx, db, userID)
				},
				CountTargetJobs: func(ctx context.Context, userID string) (int, error) {
					return targetjob.CountTargetJobsForUser(ctx, db, userID)
				},
				CountPracticeSessions: func(ctx context.Context, userID string) (int, error) {
					return practice.CountPracticeSessionsForUser(ctx, db, userID)
				},
				CountDebriefs: func(ctx context.Context, userID string) (int, error) {
					return debrief.CountDebriefsForUser(ctx, db, userID)
				},
			})
		},
	})
	h.SetRecommendations(repo, repo)
	h.SetWatchlist(repo, idx.NewID)
	h.SetSearch(repo, repo, jdMatchAI)
	h.SetMarketSignals(func(ctx context.Context, userID string, window service.MarketSignalsWindow) (api.MarketSignalsResponse, error) {
		return service.BuildMarketSignals(ctx, userID, window, service.MarketSignalsDeps{
			NewRecommendationsCount: func(ctx context.Context, userID string, window service.MarketSignalsWindow) (int, error) {
				return repo.CountActiveRecommendationsByUser(ctx, userID)
			},
			WatchlistCount: func(ctx context.Context, userID string) (int, error) {
				items, err := repo.ListWatchlistByUser(ctx, userID)
				return len(items), err
			},
			ActiveRecommendationsAvg: func(ctx context.Context, userID string) (int, error) {
				return repo.CountActiveRecommendationsByUser(ctx, userID)
			},
		})
	})
	ttl := time.Duration(sharedtypes.IdempotencyKeyTTLSeconds) * time.Second
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	ik := idempotency.New(idempotency.MiddlewareOptions{
		Store:     idempotency.NewSQLStore(db),
		KeyPepper: loader.GetSecret("auth.challengeTokenPepper").Reveal(),
		TTL:       ttl,
	})
	privacy := service.PrivacyDeleter{
		DeleteWatchlist:       repo.DeleteWatchlistForUser,
		DeleteSavedSearches:   repo.DeleteSavedSearchesForUser,
		DeleteSearchRuns:      repo.DeleteSearchRunsForUser,
		DeleteRecommendations: repo.DeleteRecommendationsForUser,
		DeleteAgentScans:      repo.DeleteAgentScansForUser,
	}
	return jdmatchRoutes{
		Handler:     h,
		Idempotency: ik,
		AgentScanRunOnce: func(ctx context.Context, userID string) error {
			return nil // wired by drainer registration below
		},
		PrivacyDeleteFunc: func(ctx context.Context, userID string) (service.PrivacyDeleteCounts, error) {
			return service.DeleteJobMatchDataForUser(ctx, userID, privacy)
		},
	}
}

// jdmatchAI is the cmd/api-side adapter the AIClient implements via
// stubAIClient or a real provider. Phase 0 wiring keeps it minimal:
// the search path returns no matches when the AIClient is unavailable
// so the route stays callable for smoke / parity tests.
type jdmatchAI interface {
	Search(ctx context.Context, userID, query string, filters json.RawMessage) (jdmatchhandler.SearchAIResult, error)
}

// stubJDMatchAI is the placeholder adapter cmd/api uses when no real
// AIClient is wired. It is also the default for unit smoke tests.
type stubJDMatchAI struct{}

func (stubJDMatchAI) Search(_ context.Context, _, _ string, _ json.RawMessage) (jdmatchhandler.SearchAIResult, error) {
	return jdmatchhandler.SearchAIResult{
		PromptVersion:     "jd_match_search.v1",
		RubricVersion:     "jd_match_search_rubric.v1",
		ModelProfileName:  "jd_match.search.default",
		Language:          "zh-CN",
		FeatureFlag:       "none",
		DataSourceVersion: "jd_match.v1",
	}, nil
}

// generatorAIAdapter wraps the AIClient.Complete signature exposed by
// jdmatch/generators so cmd/api can hand the same AIClient instance
// to the agent_scan job.
type generatorAIAdapter struct{}

func (generatorAIAdapter) Complete(_ context.Context, _ string, _ map[string]any) (generators.CompleteResult, error) {
	return generators.CompleteResult{Body: []byte("[]")}, generators.ErrInvalidLLMOutput
}

var _ profile.Store = (*profilestore.Repository)(nil)
