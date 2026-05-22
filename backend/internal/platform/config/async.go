package config

import "fmt"

// AsyncQueueWeights is the A4 async.queueWeights typed config (spec D-9).
type AsyncQueueWeights struct {
	Critical int
	Default  int
	Low      int
}

// AsyncConfig is the typed view of the `async.*` config section consumed by the
// backend internal runner kernel. queueWeights are spec D-9; the four timing
// nodes are spec D-14 additive config-only values (no env keys). Every value is
// fail-fast positive — the kernel must not silently fall back to code
// constants.
type AsyncConfig struct {
	QueueWeights          AsyncQueueWeights
	LeaseTimeoutSeconds   int
	ShutdownGraceSeconds  int
	ReaperIntervalSeconds int
	ScanIntervalSeconds   int
}

// AsyncConfig reads and validates the `async.*` section. It returns an error
// when any queue weight or timing node is missing or non-positive.
func (l *Loader) AsyncConfig() (AsyncConfig, error) {
	if l == nil {
		return AsyncConfig{}, fmt.Errorf("config: loader is nil")
	}
	cfg := AsyncConfig{
		QueueWeights: AsyncQueueWeights{
			Critical: l.GetInt("async.queueWeights.critical"),
			Default:  l.GetInt("async.queueWeights.default"),
			Low:      l.GetInt("async.queueWeights.low"),
		},
		LeaseTimeoutSeconds:   l.GetInt("async.leaseTimeoutSeconds"),
		ShutdownGraceSeconds:  l.GetInt("async.shutdownGraceSeconds"),
		ReaperIntervalSeconds: l.GetInt("async.reaperIntervalSeconds"),
		ScanIntervalSeconds:   l.GetInt("async.scanIntervalSeconds"),
	}
	var problems []string
	if cfg.QueueWeights.Critical <= 0 || cfg.QueueWeights.Default <= 0 || cfg.QueueWeights.Low <= 0 {
		problems = append(problems, "async.queueWeights must declare positive critical/default/low values")
	}
	for path, v := range map[string]int{
		"async.leaseTimeoutSeconds":   cfg.LeaseTimeoutSeconds,
		"async.shutdownGraceSeconds":  cfg.ShutdownGraceSeconds,
		"async.reaperIntervalSeconds": cfg.ReaperIntervalSeconds,
		"async.scanIntervalSeconds":   cfg.ScanIntervalSeconds,
	} {
		if v <= 0 {
			problems = append(problems, fmt.Sprintf("%s must be positive", path))
		}
	}
	if len(problems) > 0 {
		return AsyncConfig{}, fmt.Errorf("async config invalid: %v", problems)
	}
	return cfg, nil
}
