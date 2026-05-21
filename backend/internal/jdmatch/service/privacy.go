package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
)

// PrivacyDeleter is the cross-table privacy delete dependency set
// (spec §5.4). Each Delete*ForUser is a count-returning operation
// implemented by the store layer. cmd/api wires the full Repository.
type PrivacyDeleter struct {
	DeleteWatchlist        func(ctx context.Context, userID string) (int64, error)
	DeleteSavedSearches    func(ctx context.Context, userID string) (int64, error)
	DeleteSearchRuns       func(ctx context.Context, userID string) (int64, error)
	DeleteRecommendations  func(ctx context.Context, userID string) (int64, error)
	DeleteAgentScans       func(ctx context.Context, userID string) (int64, error)
	WriteAuditTombstone    func(ctx context.Context, userID string, counts PrivacyDeleteCounts) error
}

// PrivacyDeleteCounts is the aggregate the privacy_delete job tombstone
// records. No field carries user content text per spec §3.1.2.
type PrivacyDeleteCounts struct {
	WatchlistCount       int64
	SavedSearchCount     int64
	SearchRunCount       int64
	RecommendationCount  int64
	AgentScanCount       int64
}

// DeleteJobMatchDataForUser executes the privacy delete cascade in
// the canonical order watchlist_items -> saved_searches ->
// jd_match_search_runs -> jd_match_recommendations -> agent_scans
// (D-11 / D-15 derived). Any deletion error short-circuits the rest
// and the audit tombstone records the partial counts so the privacy
// runner can retry / surface the failure.
func DeleteJobMatchDataForUser(ctx context.Context, userID string, deps PrivacyDeleter) (PrivacyDeleteCounts, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return PrivacyDeleteCounts{}, jdmatch.ErrUserIDRequired
	}
	counts := PrivacyDeleteCounts{}
	deletes := []struct {
		name string
		fn   func(ctx context.Context, userID string) (int64, error)
		out  *int64
	}{
		{"watchlist_items", deps.DeleteWatchlist, &counts.WatchlistCount},
		{"saved_searches", deps.DeleteSavedSearches, &counts.SavedSearchCount},
		{"jd_match_search_runs", deps.DeleteSearchRuns, &counts.SearchRunCount},
		{"jd_match_recommendations", deps.DeleteRecommendations, &counts.RecommendationCount},
		{"agent_scans", deps.DeleteAgentScans, &counts.AgentScanCount},
	}
	for _, step := range deletes {
		if step.fn == nil {
			continue
		}
		n, err := step.fn(ctx, uid)
		if err != nil {
			if deps.WriteAuditTombstone != nil {
				_ = deps.WriteAuditTombstone(ctx, uid, counts)
			}
			return counts, fmt.Errorf("jdmatch privacy delete %s: %w", step.name, err)
		}
		*step.out = n
	}
	if deps.WriteAuditTombstone != nil {
		if err := deps.WriteAuditTombstone(ctx, uid, counts); err != nil {
			return counts, fmt.Errorf("jdmatch privacy delete audit tombstone: %w", err)
		}
	}
	return counts, nil
}
