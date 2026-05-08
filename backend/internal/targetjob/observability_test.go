package targetjob_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// Phase 0.3 contract: backend-targetjob 001 plan / spec C-9 / observability-stack
// spec §3.1.1 lock the F1 metric dictionary for the TargetJob domain. Three
// metrics must be registered with bounded labels only; high-cardinality or
// free-text labels are forbidden so log/metric storage cannot leak target id,
// user id, source URL, prompt version, or any unbounded payload.

func TestTargetJobMetricsAreRegisteredInF1Spec(t *testing.T) {
	specPath := filepath.Join("..", "..", "..", "docs", "spec", "observability-stack", "spec.md")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read F1 spec: %v", err)
	}
	specText := string(raw)
	for _, name := range []string{
		targetjob.MetricTargetJobImportsTotal,
		targetjob.MetricTargetJobParseDurationSeconds,
		targetjob.MetricTargetJobParseFailuresTotal,
	} {
		if !strings.Contains(specText, "`"+name+"`") {
			t.Errorf("F1 spec does not register TargetJob metric %s", name)
		}
	}
}

func TestTargetJobMetricsLabelsArePerF1Allowlist(t *testing.T) {
	cases := map[string][]string{
		targetjob.MetricTargetJobImportsTotal:         targetjob.TargetJobImportsLabelKeys,
		targetjob.MetricTargetJobParseDurationSeconds: targetjob.TargetJobParseDurationLabelKeys,
		targetjob.MetricTargetJobParseFailuresTotal:   targetjob.TargetJobParseFailuresLabelKeys,
	}
	for name, keys := range cases {
		if len(keys) == 0 {
			t.Errorf("metric %s: label keys must not be empty", name)
			continue
		}
		seen := map[string]bool{}
		for _, key := range keys {
			if seen[key] {
				t.Errorf("metric %s: duplicate label %q", name, key)
			}
			seen[key] = true
			if !targetjob.IsF1AllowedTargetJobMetricLabel(key) {
				t.Errorf("metric %s uses non-F1 label %q", name, key)
			}
		}
	}
}

func TestTargetJobMetricsRejectForbiddenLabels(t *testing.T) {
	forbidden := []string{
		"target_id",
		"target_job_id",
		"user_id",
		"user",
		"email",
		"source_url",
		"url",
		"prompt_version",
		"prompt",
		"response_body",
		"raw_jd_text",
		"ip",
		"trace_id",
		"request_id",
	}
	for _, label := range forbidden {
		if targetjob.IsF1AllowedTargetJobMetricLabel(label) {
			t.Errorf("forbidden label %q must not be in TargetJob allowlist", label)
		}
	}
}

func TestTargetJobMetricsAllowedLabelSet(t *testing.T) {
	want := map[string]bool{
		"service":     true,
		"operation":   true,
		"job_type":    true,
		"source_type": true,
		"language":    true,
		"result":      true,
		"error_code":  true,
	}
	got := targetjob.AllowedTargetJobMetricLabels()
	if len(got) != len(want) {
		t.Errorf("AllowedTargetJobMetricLabels count = %d, want %d", len(got), len(want))
	}
	for label := range want {
		if !targetjob.IsF1AllowedTargetJobMetricLabel(label) {
			t.Errorf("expected allowlist to include %q", label)
		}
	}
}
