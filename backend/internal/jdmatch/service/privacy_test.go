package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
)

func TestDeleteJobMatchDataForUserOrder(t *testing.T) {
	order := []string{}
	auditCalled := false
	var auditCounts service.PrivacyDeleteCounts
	deps := service.PrivacyDeleter{
		DeleteWatchlist:       func(ctx context.Context, _ string) (int64, error) { order = append(order, "watchlist"); return 3, nil },
		DeleteSavedSearches:   func(ctx context.Context, _ string) (int64, error) { order = append(order, "saved"); return 2, nil },
		DeleteSearchRuns:      func(ctx context.Context, _ string) (int64, error) { order = append(order, "runs"); return 5, nil },
		DeleteRecommendations: func(ctx context.Context, _ string) (int64, error) { order = append(order, "recommendations"); return 10, nil },
		DeleteAgentScans:      func(ctx context.Context, _ string) (int64, error) { order = append(order, "agent_scans"); return 4, nil },
		WriteAuditTombstone: func(_ context.Context, _ string, c service.PrivacyDeleteCounts) error {
			auditCalled = true
			auditCounts = c
			return nil
		},
	}
	counts, err := service.DeleteJobMatchDataForUser(context.Background(), "user-A", deps)
	if err != nil {
		t.Fatalf("DeleteJobMatchDataForUser: %v", err)
	}
	want := []string{"watchlist", "saved", "runs", "recommendations", "agent_scans"}
	for i, w := range want {
		if order[i] != w {
			t.Fatalf("step %d = %s, want %s", i, order[i], w)
		}
	}
	if counts.WatchlistCount != 3 || counts.SavedSearchCount != 2 || counts.SearchRunCount != 5 || counts.RecommendationCount != 10 || counts.AgentScanCount != 4 {
		t.Fatalf("counts = %#v", counts)
	}
	if !auditCalled {
		t.Fatalf("audit tombstone must be written on success")
	}
	if auditCounts.RecommendationCount != 10 {
		t.Fatalf("audit counts = %#v", auditCounts)
	}
}

func TestDeleteJobMatchDataForUserFailureWritesPartialTombstone(t *testing.T) {
	var tombstone service.PrivacyDeleteCounts
	deps := service.PrivacyDeleter{
		DeleteWatchlist:       func(ctx context.Context, _ string) (int64, error) { return 3, nil },
		DeleteSavedSearches:   func(ctx context.Context, _ string) (int64, error) { return 2, nil },
		DeleteSearchRuns:      func(ctx context.Context, _ string) (int64, error) { return 0, errors.New("transient") },
		DeleteRecommendations: func(ctx context.Context, _ string) (int64, error) { t.Fatal("must not run after failure"); return 0, nil },
		DeleteAgentScans:      func(ctx context.Context, _ string) (int64, error) { t.Fatal("must not run after failure"); return 0, nil },
		WriteAuditTombstone: func(_ context.Context, _ string, c service.PrivacyDeleteCounts) error {
			tombstone = c
			return nil
		},
	}
	_, err := service.DeleteJobMatchDataForUser(context.Background(), "user-A", deps)
	if err == nil {
		t.Fatalf("expected error on transient failure")
	}
	if tombstone.WatchlistCount != 3 || tombstone.SavedSearchCount != 2 || tombstone.SearchRunCount != 0 {
		t.Fatalf("partial tombstone counts = %#v", tombstone)
	}
}

func TestDeleteJobMatchDataForUserRejectsEmpty(t *testing.T) {
	_, err := service.DeleteJobMatchDataForUser(context.Background(), "  ", service.PrivacyDeleter{})
	if err == nil {
		t.Fatalf("expected error for empty userID")
	}
}
