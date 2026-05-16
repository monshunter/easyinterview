package registry

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// TestStartupBudget asserts spec §4.3: full reload of 11 feature_keys ×
// >=2 language coordinates completes within 1 second on typical hardware.
func TestStartupBudget(t *testing.T) {
	t.Parallel()
	prompts, rubrics := repoConfigRoots(t)

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
	prompts, rubrics := benchConfigRoots(b)
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

// benchConfigRoots mirrors repoConfigRoots for benchmarks (which receive
// *testing.B rather than *testing.T).
func benchConfigRoots(b *testing.B) (string, string) {
	b.Helper()
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			b.Skipf("could not locate backend go.mod from %s", wd)
			return "", ""
		}
		dir = parent
	}
	repoRoot := filepath.Dir(dir)
	return filepath.Join(repoRoot, "config", "prompts"),
		filepath.Join(repoRoot, "config", "rubrics")
}
