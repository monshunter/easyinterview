package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/debrief"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/generators"
	jdmatchhandler "github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	jdmatchjobs "github.com/monshunter/easyinterview/backend/internal/jdmatch/jobs"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
	jdmatchstore "github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/platform/config"
	"github.com/monshunter/easyinterview/backend/internal/practice"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	profilestore "github.com/monshunter/easyinterview/backend/internal/profile/store"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	sharedjobs "github.com/monshunter/easyinterview/backend/internal/shared/jobs"
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

type jdmatchRuntime struct {
	Routes  jdmatchRoutes
	Drainer *targetjob.Drainer
}

func (r *jdmatchRuntime) Start(ctx context.Context) {
	if r == nil || r.Drainer == nil {
		return
	}
	r.Drainer.Start(ctx)
}

func (r *jdmatchRuntime) Shutdown(ctx context.Context) error {
	if r == nil || r.Drainer == nil {
		return nil
	}
	return r.Drainer.Shutdown(ctx)
}

func buildJDMatchRuntime(loader *config.Loader, db *sql.DB, logger *slog.Logger, searchAI jdmatchAI, generatorAI generators.AIClient) (*jdmatchRuntime, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if searchAI == nil {
		searchAI = stubJDMatchAI{}
	}
	if generatorAI == nil {
		generatorAI = generatorAIAdapter{}
	}
	routes := buildJDMatchRoutes(loader, db, searchAI)
	repo := jdmatchstore.NewRepository(db, func() time.Time { return time.Now().UTC() })
	runAgentScan := func(ctx context.Context, userID string) error {
		return runJDMatchAgentScan(ctx, db, repo, userID, generatorAI)
	}
	routes.AgentScanRunOnce = runAgentScan
	drainer := targetjob.NewDrainer(targetjob.DrainerOptions{
		Store: targetjob.NewSQLStore(db),
		Handlers: map[string]targetjob.JobHandler{
			string(sharedjobs.JobTypeJdMatchAgentScan): targetjob.JobHandlerFunc(func(ctx context.Context, job targetjob.ClaimedJob) targetjob.JobOutcome {
				userID, err := jdMatchAgentScanUserID(job)
				if err != nil {
					return targetjob.JobOutcome{
						ErrorCode:    sharederrors.CodeValidationFailed,
						ErrorMessage: err.Error(),
					}
				}
				if err := runAgentScan(ctx, userID); err != nil {
					return jdMatchAgentScanOutcome(err)
				}
				return targetjob.JobOutcome{Succeeded: true}
			}),
		},
		Logger: logger,
	})
	return &jdmatchRuntime{Routes: routes, Drainer: drainer}, nil
}

func buildJDMatchAIAdapters(loader *config.Loader, db *sql.DB, ai aiclient.AIClient) (jdmatchAI, generators.AIClient, error) {
	if ai == nil {
		return nil, nil, fmt.Errorf("jdmatch AI client is required")
	}
	registryClient, err := registry.NewRegistryClient(registry.RegistryOptions{
		PromptsDir: registryDirOrDefault(loader, "ai.promptsDir", "config/prompts"),
		RubricsDir: registryDirOrDefault(loader, "ai.rubricsDir", "config/rubrics"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build jdmatch prompt registry: %w", err)
	}
	adapter := jdMatchA3F3Adapter{
		registry:        registryClient,
		ai:              ai,
		recommendations: jdmatchstore.NewRepository(db, func() time.Time { return time.Now().UTC() }),
	}
	return adapter, adapter, nil
}

func runJDMatchAgentScan(ctx context.Context, db *sql.DB, repo *jdmatchstore.Repository, userID string, ai generators.AIClient) error {
	return jdmatchjobs.Run(ctx, userID, jdmatchjobs.AgentScanDeps{
		AgentScans: repo,
		CandidateProfile: func(ctx context.Context, userID string) (json.RawMessage, error) {
			return buildJDMatchProfileJSON(ctx, db, userID)
		},
		JobsPool: func(ctx context.Context, userID string) (json.RawMessage, error) {
			return buildJDMatchJobsPoolJSON(ctx, repo, userID)
		},
		Generator: func(ctx context.Context, in generators.RunRecommendationGeneratorInput) (generators.RunRecommendationGeneratorResult, error) {
			return generators.RunRecommendationGenerator(ctx, ai, repo, in)
		},
		NewID:            idx.NewID,
		NextScanInterval: 24 * time.Hour,
		OutboxEmit: func(ctx context.Context, event generators.RecommendationCompletedEvent) error {
			return writeJDMatchRecommendationCompletedOutbox(ctx, db, event)
		},
	})
}

func jdMatchAgentScanUserID(job targetjob.ClaimedJob) (string, error) {
	var payload struct {
		UserID string `json:"userId"`
	}
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return "", fmt.Errorf("jd_match_agent_scan payload is invalid JSON: %w", err)
		}
	}
	userID := strings.TrimSpace(payload.UserID)
	if userID == "" && strings.TrimSpace(job.ResourceType) == "user" {
		userID = strings.TrimSpace(job.ResourceID)
	}
	if userID == "" {
		return "", errors.New("jd_match_agent_scan requires payload.userId or resource_type=user resource_id")
	}
	return userID, nil
}

func jdMatchAgentScanOutcome(err error) targetjob.JobOutcome {
	out := targetjob.JobOutcome{
		ErrorCode:    sharederrors.CodeValidationFailed,
		ErrorMessage: "jd_match_agent_scan failed",
	}
	switch {
	case errors.Is(err, generators.ErrInvalidLLMOutput):
		out.ErrorCode = sharederrors.CodeAiOutputInvalid
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, jdmatchhandler.SearchTimeoutErr):
		out.ErrorCode = sharederrors.CodeAiProviderTimeout
		out.Retryable = true
	default:
		out.Retryable = false
	}
	return out
}

func writeJDMatchRecommendationCompletedOutbox(ctx context.Context, db *sql.DB, event generators.RecommendationCompletedEvent) error {
	if db == nil {
		return fmt.Errorf("jdmatch outbox writer db is nil")
	}
	payload := sharedevents.JdMatchRecommendationCompletedPayload{
		UserID:              event.UserID,
		AgentScanID:         event.AgentScanID,
		RecommendationCount: event.RecommendationCount,
		CompletedAt:         time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal jdmatch recommendation completed event: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values (
  $1, $2, 1, 'agent_scan', $3, $4::jsonb, 'pending', now()
)`,
		idx.NewID(),
		string(sharedevents.EventNameJdMatchRecommendationCompleted),
		event.AgentScanID,
		string(raw),
	); err != nil {
		return fmt.Errorf("insert jdmatch recommendation completed outbox: %w", err)
	}
	return nil
}

func buildJDMatchProfileJSON(ctx context.Context, db *sql.DB, userID string) (json.RawMessage, error) {
	if db == nil {
		return nil, fmt.Errorf("jdmatch profile db is nil")
	}
	profileRepo := profilestore.NewRepository(db)
	res, err := service.BuildJobMatchProfile(ctx, userID, jdMatchProfileDeps(db, profileRepo))
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(res.Profile)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func jdMatchProfileDeps(db *sql.DB, profileRepo *profilestore.Repository) service.ProfileDeps {
	return service.ProfileDeps{
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
			cp.UiLanguage = rec.UILanguage
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
	}
}

type jdMatchSearchCompletedEmitter struct {
	db *sql.DB
}

func (e jdMatchSearchCompletedEmitter) EmitSearchCompleted(ctx context.Context, event jdmatchhandler.SearchCompletedEvent) error {
	return writeJDMatchSearchCompletedOutbox(ctx, e.db, event)
}

func writeJDMatchSearchCompletedOutbox(ctx context.Context, db *sql.DB, event jdmatchhandler.SearchCompletedEvent) error {
	if db == nil {
		return fmt.Errorf("jdmatch search completed writer db is nil")
	}
	payload := sharedevents.JdMatchSearchCompletedPayload{
		UserID:      event.UserID,
		SearchRunID: event.SearchRunID,
		ResultCount: event.ResultCount,
		CompletedAt: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal jdmatch search completed event: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values (
  $1, $2, 1, 'search_run', $3, $4::jsonb, 'pending', now()
)`,
		idx.NewID(),
		string(sharedevents.EventNameJdMatchSearchCompleted),
		event.SearchRunID,
		string(raw),
	); err != nil {
		return fmt.Errorf("insert jdmatch search completed outbox: %w", err)
	}
	return nil
}

func writeJDMatchPrivacyAuditTombstone(ctx context.Context, db *sql.DB, userID string, counts service.PrivacyDeleteCounts) error {
	return writeJDMatchPrivacyAuditTombstoneWithExec(ctx, db, userID, counts)
}

type jdMatchSQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func writeJDMatchPrivacyAuditTombstoneWithExec(ctx context.Context, exec jdMatchSQLExecutor, userID string, counts service.PrivacyDeleteCounts) error {
	if exec == nil {
		return fmt.Errorf("jdmatch privacy audit writer db is nil")
	}
	metadata := map[string]any{
		"watchlistCount":      counts.WatchlistCount,
		"savedSearchCount":    counts.SavedSearchCount,
		"searchRunCount":      counts.SearchRunCount,
		"recommendationCount": counts.RecommendationCount,
		"agentScanCount":      counts.AgentScanCount,
		"deletedAt":           time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal jdmatch privacy audit metadata: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type, resource_id,
  result, ip_hash, user_agent_hash, metadata, created_at
) values (
  $1, $2, 'system', null, 'jd_match.privacy_delete', 'jd_match', null,
  'success', null, null, $3::jsonb, now()
)`,
		idx.NewID(),
		userID,
		string(raw),
	); err != nil {
		return fmt.Errorf("insert jdmatch privacy audit tombstone: %w", err)
	}
	return nil
}

func deleteJDMatchDataForUserInTx(ctx context.Context, db *sql.DB, userID string) (service.PrivacyDeleteCounts, error) {
	if db == nil {
		return service.PrivacyDeleteCounts{}, fmt.Errorf("jdmatch privacy delete db is nil")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return service.PrivacyDeleteCounts{}, fmt.Errorf("begin jdmatch privacy delete tx: %w", err)
	}
	deleteForUser := func(table string) func(context.Context, string) (int64, error) {
		return func(ctx context.Context, userID string) (int64, error) {
			res, err := tx.ExecContext(ctx, "delete from "+table+" where user_id = $1", userID)
			if err != nil {
				return 0, err
			}
			n, err := res.RowsAffected()
			if err != nil {
				return 0, err
			}
			return n, nil
		}
	}
	counts, err := service.DeleteJobMatchDataForUser(ctx, userID, service.PrivacyDeleter{
		DeleteWatchlist:       deleteForUser("watchlist_items"),
		DeleteSavedSearches:   deleteForUser("saved_searches"),
		DeleteSearchRuns:      deleteForUser("jd_match_search_runs"),
		DeleteRecommendations: deleteForUser("jd_match_recommendations"),
		DeleteAgentScans:      deleteForUser("agent_scans"),
		WriteAuditTombstone: func(ctx context.Context, userID string, counts service.PrivacyDeleteCounts) error {
			return writeJDMatchPrivacyAuditTombstoneWithExec(ctx, tx, userID, counts)
		},
	})
	if err != nil {
		_ = tx.Rollback()
		return counts, err
	}
	if err := tx.Commit(); err != nil {
		return counts, fmt.Errorf("commit jdmatch privacy delete tx: %w", err)
	}
	return counts, nil
}

type jdmatchRouteOptions struct {
	Now   func() time.Time
	NewID func() string
}

// buildJDMatchRoutes composes the backend-jobs-recommendations
// runtime: store layer + cross-owner counters / identity / candidate
// profile + market signals + agent_scan job + privacy delete. AI is
// passed in by the caller so the same stub used elsewhere in cmd/api
// tests covers JD-Match recommendation / search too.
func buildJDMatchRoutes(loader *config.Loader, db *sql.DB, jdMatchAI jdmatchAI) jdmatchRoutes {
	return buildJDMatchRoutesWithOptions(loader, db, jdMatchAI, jdmatchRouteOptions{})
}

func buildJDMatchRoutesWithOptions(loader *config.Loader, db *sql.DB, jdMatchAI jdmatchAI, opts jdmatchRouteOptions) jdmatchRoutes {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	repo := jdmatchstore.NewRepository(db, now)
	profileRepo := profilestore.NewRepository(db)
	h := jdmatchhandler.New(jdmatchhandler.Options{
		Session:    currentUserFromContext,
		AgentScans: repo,
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.BuildJobMatchProfile(ctx, userID, jdMatchProfileDeps(db, profileRepo))
		},
	})
	h.SetRecommendations(repo, repo)
	h.SetWatchlist(repo, newID)
	h.SetSearch(repo, repo, jdMatchAI)
	h.SetSearchCompleted(jdMatchSearchCompletedEmitter{db: db})
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
			NowFn: now,
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
	return jdmatchRoutes{
		Handler:     h,
		Idempotency: ik,
		AgentScanRunOnce: func(ctx context.Context, userID string) error {
			return nil // wired by drainer registration below
		},
		PrivacyDeleteFunc: func(ctx context.Context, userID string) (service.PrivacyDeleteCounts, error) {
			return deleteJDMatchDataForUserInTx(ctx, db, userID)
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

type jdMatchPromptRegistry interface {
	ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error)
}

type jdMatchRecommendationPool interface {
	ListRecommendationsByUser(ctx context.Context, userID string, filter jdmatchstore.ListRecommendationsFilter) (jdmatchstore.ListRecommendationsResult, error)
}

type jdMatchA3F3Adapter struct {
	registry        jdMatchPromptRegistry
	ai              aiclient.AIClient
	recommendations jdMatchRecommendationPool
}

func (a jdMatchA3F3Adapter) Search(ctx context.Context, userID, query string, filters json.RawMessage) (jdmatchhandler.SearchAIResult, error) {
	resolution, err := a.resolve(ctx, featurekeys.JdMatchSearch.String(), jdMatchDefaultLanguage)
	if err != nil {
		return jdmatchhandler.SearchAIResult{}, err
	}
	jobsPool, err := a.searchJobsPool(ctx, userID)
	if err != nil {
		return jdmatchhandler.SearchAIResult{}, err
	}
	rendered := renderJDMatchPrompt(resolution.UserMessageTemplate, map[string]string{
		"query":             query,
		"filters":           rawJSONOrEmptyObject(filters),
		"candidate_profile": "{}",
		"jobs_pool":         jobsPool,
		"language":          jdMatchDefaultLanguage,
	})
	resp, meta, err := a.ai.Complete(ctx, resolution.ModelProfileName, jdMatchPayload(resolution, jdMatchDefaultLanguage, rendered))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return jdmatchhandler.SearchAIResult{}, jdmatchhandler.SearchTimeoutErr
		}
		return jdmatchhandler.SearchAIResult{}, err
	}
	ids, err := parseJDMatchSearchIDs(resp.Content)
	if err != nil {
		return jdmatchhandler.SearchAIResult{}, fmt.Errorf("%w: %v", jdmatchhandler.SearchInvalidOutputErr, err)
	}
	return jdmatchhandler.SearchAIResult{
		MatchedJobMatchIDs: ids,
		PromptVersion:      firstNonEmptyString(meta.PromptVersion, resolution.PromptVersion),
		RubricVersion:      firstNonEmptyString(meta.RubricVersion, resolution.RubricVersion),
		ModelProfileName:   firstNonEmptyString(meta.ModelProfileName, resolution.ModelProfileName),
		Language:           firstNonEmptyString(meta.Language, jdMatchDefaultLanguage),
		FeatureFlag:        firstNonEmptyString(meta.FeatureFlag, resolution.FeatureFlag, "none"),
		DataSourceVersion:  firstNonEmptyString(meta.DataSourceVersion, resolution.DataSourceVersion, "registry.v1"),
	}, nil
}

func (a jdMatchA3F3Adapter) Complete(ctx context.Context, featureKey string, payload map[string]any) (generators.CompleteResult, error) {
	resolution, err := a.resolve(ctx, featureKey, jdMatchDefaultLanguage)
	if err != nil {
		return generators.CompleteResult{}, err
	}
	rendered := renderJDMatchPrompt(resolution.UserMessageTemplate, map[string]string{
		"candidate_profile": jsonStringValue(payload["candidateProfile"]),
		"jobs_pool":         jsonStringValue(payload["jobsPool"]),
		"language":          jdMatchDefaultLanguage,
	})
	resp, meta, err := a.ai.Complete(ctx, resolution.ModelProfileName, jdMatchPayload(resolution, jdMatchDefaultLanguage, rendered))
	if err != nil {
		return generators.CompleteResult{}, err
	}
	return generators.CompleteResult{
		Body:              json.RawMessage(resp.Content),
		PromptVersion:     firstNonEmptyString(meta.PromptVersion, resolution.PromptVersion),
		RubricVersion:     firstNonEmptyString(meta.RubricVersion, resolution.RubricVersion),
		ModelProfileName:  firstNonEmptyString(meta.ModelProfileName, resolution.ModelProfileName),
		Language:          firstNonEmptyString(meta.Language, jdMatchDefaultLanguage),
		FeatureFlag:       firstNonEmptyString(meta.FeatureFlag, resolution.FeatureFlag, "none"),
		DataSourceVersion: firstNonEmptyString(meta.DataSourceVersion, resolution.DataSourceVersion, "registry.v1"),
	}, nil
}

func (a jdMatchA3F3Adapter) resolve(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	if a.registry == nil || a.ai == nil {
		return registry.PromptResolution{}, fmt.Errorf("jdmatch A3/F3 adapter is not configured")
	}
	return a.registry.ResolveActive(ctx, featureKey, language)
}

func (a jdMatchA3F3Adapter) searchJobsPool(ctx context.Context, userID string) (string, error) {
	raw, err := buildJDMatchJobsPoolJSON(ctx, a.recommendations, userID)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func buildJDMatchJobsPoolJSON(ctx context.Context, recommendations jdMatchRecommendationPool, userID string) (json.RawMessage, error) {
	if recommendations == nil {
		return json.RawMessage("[]"), nil
	}
	res, err := recommendations.ListRecommendationsByUser(ctx, userID, jdmatchstore.ListRecommendationsFilter{PageSize: 100})
	if err != nil {
		return nil, err
	}
	raw, err := compactJDMatchJobsPool(res.Items)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

const jdMatchDefaultLanguage = "zh-CN"

func jdMatchPayload(resolution registry.PromptResolution, language, renderedPrompt string) aiclient.CompletePayload {
	metadata := aiclient.CallMetadata{
		FeatureKey:        resolution.FeatureKey,
		PromptVersion:     resolution.PromptVersion,
		RubricVersion:     resolution.RubricVersion,
		Language:          language,
		FeatureFlag:       resolution.FeatureFlag,
		DataSourceVersion: resolution.DataSourceVersion,
	}
	if resolution.OutputSchema != nil {
		metadata.OutputSchema = *resolution.OutputSchema
	}
	messages := make([]aiclient.Message, 0, 2)
	if strings.TrimSpace(resolution.SystemMessage) != "" {
		messages = append(messages, aiclient.Message{Role: "system", Content: resolution.SystemMessage})
	}
	messages = append(messages, aiclient.Message{Role: "user", Content: renderedPrompt})
	return aiclient.CompletePayload{
		Messages: messages,
		Metadata: metadata,
		Tools:    jdMatchTools(resolution.Tools),
	}
}

func jdMatchTools(tools []registry.ToolDescriptor) []aiclient.Tool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]aiclient.Tool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, aiclient.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		})
	}
	return out
}

func renderJDMatchPrompt(template string, values map[string]string) string {
	rendered := template
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered
}

func rawJSONOrEmptyObject(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	if !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

func jsonStringValue(value any) string {
	switch v := value.(type) {
	case nil:
		return "{}"
	case json.RawMessage:
		if len(v) == 0 || !json.Valid(v) {
			return "{}"
		}
		return string(v)
	case []byte:
		if len(v) == 0 || !json.Valid(v) {
			return "{}"
		}
		return string(v)
	case string:
		if strings.TrimSpace(v) == "" {
			return "{}"
		}
		return v
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return "{}"
		}
		return string(raw)
	}
}

func compactJDMatchJobsPool(items []jdmatch.RecommendationRecord) (string, error) {
	pool := make([]map[string]any, 0, len(items))
	for _, item := range items {
		pool = append(pool, map[string]any{
			"jobMatchId": item.ID,
			"title":      item.Title,
			"company":    item.Company,
			"level":      item.Level,
			"location":   item.Location,
			"score":      item.Score,
			"highlights": item.Highlights,
		})
	}
	raw, err := json.Marshal(pool)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func parseJDMatchSearchIDs(content string) ([]string, error) {
	var rawItems []json.RawMessage
	if err := json.Unmarshal([]byte(content), &rawItems); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rawItems))
	ids := make([]string, 0, len(rawItems))
	for _, raw := range rawItems {
		var id string
		if err := json.Unmarshal(raw, &id); err == nil {
			id = strings.TrimSpace(id)
		} else {
			var item struct {
				JobMatchID string `json:"jobMatchId"`
				ID         string `json:"id"`
			}
			if err := json.Unmarshal(raw, &item); err != nil {
				return nil, err
			}
			id = strings.TrimSpace(firstNonEmptyString(item.JobMatchID, item.ID))
		}
		if id == "" {
			return nil, fmt.Errorf("search item missing jobMatchId")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// stubJDMatchAI is the placeholder adapter used by unit smoke tests.
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
