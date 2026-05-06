package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestBackgroundMailDispatcherEnqueueReturnsBeforeDrainAndShutdownDrains(t *testing.T) {
	writer := &blockingDeliveryWriter{
		block: make(chan struct{}),
		done:  make(chan struct{}),
	}
	dispatcher := auth.NewBackgroundMailDispatcher(auth.BackgroundMailDispatcherOptions{
		Writer:    writer,
		QueueSize: 1,
	})
	defer dispatcher.Shutdown(context.Background())

	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   "challenge-1",
		"templateKey":       "auth_magic_link",
		"locale":            "en",
		"deliverySecretRef": "auth_challenge:challenge-1",
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := dispatcher.Enqueue(context.Background(), payload); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	select {
	case <-writer.done:
		t.Fatal("dispatcher drained synchronously; handler would wait for provider")
	case <-time.After(20 * time.Millisecond):
	}

	close(writer.block)
	if err := dispatcher.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if writer.count != 1 {
		t.Fatalf("writer count = %d", writer.count)
	}
}

func TestBackgroundMailDispatcherFailureSummaryIsRedacted(t *testing.T) {
	writer := &failingDeliveryWriter{
		err: errors.New("provider failed for raw-token candidate@example.com http://api.test/verify?token=raw-token"),
	}
	dispatcher := auth.NewBackgroundMailDispatcher(auth.BackgroundMailDispatcherOptions{
		Writer:    writer,
		QueueSize: 1,
	})
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   "challenge-2",
		"templateKey":       "auth_magic_link",
		"locale":            "en",
		"deliverySecretRef": "auth_challenge:challenge-2",
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatcher.Enqueue(context.Background(), payload); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if err := dispatcher.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	summaries := dispatcher.ErrorSummaries()
	if len(summaries) != 1 {
		t.Fatalf("summaries = %#v", summaries)
	}
	for _, forbidden := range []string{"raw-token", "candidate@example.com", "http://api.test"} {
		if contains(summaries[0], forbidden) {
			t.Fatalf("failure summary leaked %s: %s", forbidden, summaries[0])
		}
	}
}

type blockingDeliveryWriter struct {
	block chan struct{}
	done  chan struct{}
	count int
}

func (w *blockingDeliveryWriter) Write(jobs.EmailDispatchPayload) error {
	<-w.block
	w.count++
	close(w.done)
	return nil
}

type failingDeliveryWriter struct {
	err error
}

func (w *failingDeliveryWriter) Write(jobs.EmailDispatchPayload) error {
	return w.err
}
