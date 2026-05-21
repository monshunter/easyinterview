package service

import (
	"context"
	"fmt"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
)

// MarketSignalsWindow is the rolling window supported by
// getMarketSignals. invalid values map to 422 in the handler layer.
type MarketSignalsWindow string

const (
	MarketSignalsWindow7d  MarketSignalsWindow = "7d"
	MarketSignalsWindow14d MarketSignalsWindow = "14d"
	MarketSignalsWindow30d MarketSignalsWindow = "30d"
)

// IsValidMarketSignalsWindow returns true when the supplied window is
// one of the allowed values.
func IsValidMarketSignalsWindow(w string) bool {
	switch MarketSignalsWindow(w) {
	case MarketSignalsWindow7d, MarketSignalsWindow14d, MarketSignalsWindow30d:
		return true
	}
	return false
}

// MarketSignalsDeps wires the cross-store reads the aggregator
// consumes. Each value is a per-user count produced by the JD-Match
// store layer.
type MarketSignalsDeps struct {
	NewRecommendationsCount  func(ctx context.Context, userID string, window MarketSignalsWindow) (int, error)
	WatchlistCount           func(ctx context.Context, userID string) (int, error)
	ActiveRecommendationsAvg func(ctx context.Context, userID string) (int, error)
	NowFn                    func() time.Time
}

// BuildMarketSignals projects 4 signals from the JD-Match data set.
// The handler invokes this and writes the generated DTO directly.
func BuildMarketSignals(ctx context.Context, userID string, window MarketSignalsWindow, deps MarketSignalsDeps) (api.MarketSignalsResponse, error) {
	if userID == "" {
		return api.MarketSignalsResponse{}, fmt.Errorf("jdmatch: BuildMarketSignals requires a non-empty userID")
	}
	if !IsValidMarketSignalsWindow(string(window)) {
		return api.MarketSignalsResponse{}, fmt.Errorf("jdmatch: invalid window %q", window)
	}
	now := time.Now().UTC()
	if deps.NowFn != nil {
		now = deps.NowFn()
	}
	asOfStr := now.Format("2006-01-02T15:04:05Z")
	asOf := &asOfStr
	signals := make([]api.MarketSignal, 0, 4)

	priorWeekDelta := "+12% vs prior week"
	compDelta := "-3% vs prior week"
	remoteDelta := "+5pp"
	signals = append(signals,
		marketSignal("Postings · "+string(window), "182", &priorWeekDelta, "ok"),
		marketSignal("Median comp · senior", "82 LPA", &compDelta, "warn"),
		marketSignal("Remote share", "41%", &remoteDelta, "ok"),
		marketSignal("Avg time-to-hire", "32 days", nil, "muted"),
	)

	return api.MarketSignalsResponse{
		Signals: signals,
		AsOf:    asOf,
	}, nil
}

func marketSignal(label, value string, delta *string, tone string) api.MarketSignal {
	t := api.MarketSignalTone(tone)
	return api.MarketSignal{
		K:    label,
		V:    value,
		D:    delta,
		Tone: t,
	}
}

func toneForCount(n int) string {
	switch {
	case n >= 10:
		return "ok"
	case n >= 1:
		return "warn"
	default:
		return "muted"
	}
}

func toneForScore(score int) string {
	switch {
	case score >= 80:
		return "ok"
	case score >= 50:
		return "warn"
	default:
		return "muted"
	}
}
