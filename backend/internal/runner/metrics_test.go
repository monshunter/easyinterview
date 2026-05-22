package runner

import (
	"reflect"
	"testing"
	"time"
)

func TestKernelMetrics_FamilyAndLabels(t *testing.T) {
	reg := NewInMemoryMetricsRegistry()
	_ = NewKernelMetrics(reg)

	wantLabels := map[string][]string{
		MetricAsyncJobDuration:      {"job_type"},
		MetricAsyncJobsProcessed:    {"job_type", "result"},
		MetricAsyncJobQueueDepth:    {"job_type"},
		MetricAsyncJobLag:           {"job_type"},
		MetricOutboxPending:         nil,
		MetricOutboxPublishDuration: {"result"},
		MetricOutboxPublishFailures: nil,
	}
	for name, labels := range wantLabels {
		if !reg.Registered(name) {
			t.Fatalf("metric family %q not registered", name)
		}
		got := reg.LabelKeys(name)
		if len(got) == 0 && len(labels) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, labels) {
			t.Fatalf("metric %q label keys = %v, want %v", name, got, labels)
		}
	}
}

func TestKernelMetrics_RecordsJobAndOutbox(t *testing.T) {
	reg := NewInMemoryMetricsRegistry()
	m := NewKernelMetrics(reg)

	m.ObserveJobProcessed("report_generate", "succeeded", 250*time.Millisecond)
	m.ObserveJobProcessed("report_generate", "succeeded", 250*time.Millisecond)
	m.ObserveReaped("target_import", 3)
	m.ObserveOutboxFailure()
	m.SetOutboxPending(7)

	if got := reg.Value(MetricAsyncJobsProcessed, "report_generate", "succeeded"); got != 2 {
		t.Fatalf("async_jobs_processed_total{succeeded} = %v, want 2", got)
	}
	if got := reg.Value(MetricAsyncJobsProcessed, "target_import", "reaped"); got != 3 {
		t.Fatalf("async_jobs_processed_total{reaped} = %v, want 3", got)
	}
	if got := reg.Value(MetricOutboxPublishFailures); got != 1 {
		t.Fatalf("outbox_publish_failures_total = %v, want 1", got)
	}
	if got := reg.Value(MetricOutboxPending); got != 7 {
		t.Fatalf("outbox_events_pending = %v, want 7", got)
	}
}
