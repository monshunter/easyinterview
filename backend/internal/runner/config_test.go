package runner

import (
	"testing"
	"time"
)

func TestRuntimeConfig_AsyncTimings(t *testing.T) {
	weights := QueueWeights{Critical: 6, Default: 3, Low: 1}
	cfg := ConfigFromSeconds(5, 300, 60, 10, weights)

	if cfg.ScanInterval != 5*time.Second {
		t.Fatalf("ScanInterval = %s, want 5s", cfg.ScanInterval)
	}
	if cfg.LeaseTimeout != 300*time.Second {
		t.Fatalf("LeaseTimeout = %s, want 300s", cfg.LeaseTimeout)
	}
	if cfg.ReaperInterval != 60*time.Second {
		t.Fatalf("ReaperInterval = %s, want 60s", cfg.ReaperInterval)
	}
	if cfg.ShutdownGrace != 10*time.Second {
		t.Fatalf("ShutdownGrace = %s, want 10s", cfg.ShutdownGrace)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestRuntimeConfig_FailsFastOnNonPositive(t *testing.T) {
	weights := QueueWeights{Critical: 6, Default: 3, Low: 1}
	cases := map[string]Config{
		"scan":   ConfigFromSeconds(0, 300, 60, 10, weights),
		"lease":  ConfigFromSeconds(5, 0, 60, 10, weights),
		"reaper": ConfigFromSeconds(5, 300, 0, 10, weights),
		"grace":  ConfigFromSeconds(5, 300, 60, 0, weights),
		"weight": ConfigFromSeconds(5, 300, 60, 10, QueueWeights{Critical: 0, Default: 3, Low: 1}),
	}
	for name, cfg := range cases {
		if err := cfg.Validate(); err == nil {
			t.Fatalf("%s: expected Validate to reject non-positive value", name)
		}
	}
}
