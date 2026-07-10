package registry

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/testsupport"
)

// TestStartupBudget asserts spec §4.3: full reload of the canonical multi
// baseline coordinates completes within 1 second on typical hardware.
func TestStartupBudget(t *testing.T) {
	t.Parallel()
	prompts, rubrics := testsupport.ConfigRoots(t)

	start := time.Now()
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("NewRegistryClient: %v", err)
	}
	_ = client.SnapshotSize()

	if elapsed > time.Second {
		t.Fatalf("startup budget exceeded: %v > 1s", elapsed)
	}
}

// BenchmarkResolve and TestResolveP95Budget jointly enforce spec §4.3
// Resolve P95 <= 5ms. Benchmark provides reproducible timing via go test
// -bench, while the unit test asserts the budget without requiring -bench.
func BenchmarkResolve(b *testing.B) {
	prompts, rubrics := testsupport.ConfigRoots(b)
	client, err := NewRegistryClient(RegistryOptions{
		PromptsDir: prompts,
		RubricsDir: rubrics,
	})
	if err != nil {
		b.Fatalf("NewRegistryClient: %v", err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.ResolveActive(ctx, "practice.session.follow_up", "en"); err != nil {
			b.Fatalf("ResolveActive: %v", err)
		}
	}
}

func TestResolveP95Budget(t *testing.T) {
	t.Parallel()
	client := newTestClient(t)
	ctx := context.Background()

	const samples = 1000
	measurements := make([]time.Duration, samples)
	for i := 0; i < samples; i++ {
		start := time.Now()
		if _, err := client.ResolveActive(ctx, "practice.session.follow_up", "en"); err != nil {
			t.Fatalf("ResolveActive: %v", err)
		}
		measurements[i] = time.Since(start)
	}
	sort.Slice(measurements, func(i, j int) bool {
		return measurements[i] < measurements[j]
	})
	p95 := measurements[int(float64(samples)*0.95)]
	if p95 > 5*time.Millisecond {
		t.Fatalf("Resolve P95 budget exceeded: %v > 5ms", p95)
	}
}
