package main

import (
	"github.com/monshunter/easyinterview/backend/internal/runner"
)

// testRunnerConfig returns a kernel Config suitable for cmd/api lifecycle and
// integration tests: a short scan interval, the reaper loop disabled (Reaper
// interval 0), and the spec D-9 default queue weights.
func testRunnerConfig() runner.Config {
	return runner.ConfigFromSeconds(5, 300, 0, 10, runner.QueueWeights{Critical: 6, Default: 3, Low: 1})
}
