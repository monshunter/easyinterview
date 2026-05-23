package config_test

import (
	"path/filepath"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/platform/config"
)

func TestAsyncSection(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
  leaseTimeoutSeconds: 300
  shutdownGraceSeconds: 10
  reaperIntervalSeconds: 60
  scanIntervalSeconds: 5
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cfg, err := loader.AsyncConfig()
	if err != nil {
		t.Fatalf("AsyncConfig: %v", err)
	}
	if cfg.QueueWeights != (config.AsyncQueueWeights{Critical: 6, Default: 3, Low: 1}) {
		t.Fatalf("queue weights = %+v", cfg.QueueWeights)
	}
	if cfg.LeaseTimeoutSeconds != 300 || cfg.ShutdownGraceSeconds != 10 ||
		cfg.ReaperIntervalSeconds != 60 || cfg.ScanIntervalSeconds != 5 {
		t.Fatalf("async timings = %+v", cfg)
	}
}

func TestAsyncSection_FailsFastOnMissingTimings(t *testing.T) {
	dir := t.TempDir()
	// queueWeights present but the D-14 timings missing -> must fail fast,
	// never silently fall back to code constants.
	writeYAML(t, filepath.Join(dir, "config.yaml"), `
async:
  queueWeights:
    critical: 6
    default: 3
    low: 1
`)
	loader, err := config.Load(config.Options{ConfigDir: dir})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := loader.AsyncConfig(); err == nil {
		t.Fatalf("expected AsyncConfig to fail fast on missing timing nodes")
	}
}
