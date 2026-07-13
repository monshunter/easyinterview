package runner

import (
	"context"
	"sort"
	"sync"
	"time"
)

// fakeRow mirrors the async_jobs columns the kernel reads/writes.
type fakeRow struct {
	id          string
	jobType     string
	resourceID  string
	payload     []byte
	status      string
	attempts    int32
	maxAttempts int32
	availableAt time.Time
	createdAt   time.Time
	lockedAt    *time.Time
	completedAt *time.Time
	errorCode   string
	errorMsg    string
}

// fakeStore is an in-memory LeaseStore that mimics the spec §4.2 SQL
// semantics. It is the deterministic unit-test substitute for SQLStore.
type fakeStore struct {
	mu   sync.Mutex
	rows map[string]*fakeRow
	seq  int
}

func newFakeStore() *fakeStore {
	return &fakeStore{rows: map[string]*fakeRow{}}
}

// enqueue inserts a queued row with the given attempts already recorded.
func (s *fakeStore) enqueue(id, jobType string, attempts int32, availableAt, createdAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	s.rows[id] = &fakeRow{
		id:          id,
		jobType:     jobType,
		resourceID:  id,
		payload:     []byte("{}"),
		status:      "queued",
		attempts:    attempts,
		maxAttempts: MaxAttempts,
		availableAt: availableAt,
		createdAt:   createdAt,
	}
}

func (s *fakeStore) enqueueRunning(id, jobType string, lockedAt time.Time, attempts int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	locked := lockedAt
	s.rows[id] = &fakeRow{
		id:          id,
		jobType:     jobType,
		resourceID:  id,
		payload:     []byte("{}"),
		status:      "running",
		attempts:    attempts,
		maxAttempts: MaxAttempts,
		availableAt: lockedAt,
		createdAt:   lockedAt,
		lockedAt:    &locked,
	}
}

func (s *fakeStore) setPayload(id string, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if row, ok := s.rows[id]; ok {
		row.payload = payload
	}
}

func (s *fakeStore) setMaxAttempts(id string, maxAttempts int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if row, ok := s.rows[id]; ok {
		row.maxAttempts = maxAttempts
	}
}

func (s *fakeStore) get(id string) *fakeRow {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return nil
	}
	clone := *row
	return &clone
}

func (s *fakeStore) LeaseAsyncJob(_ context.Context, jobTypes []string, now time.Time) (ClaimedJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allowed := map[string]struct{}{}
	for _, jt := range jobTypes {
		allowed[jt] = struct{}{}
	}
	candidates := make([]*fakeRow, 0)
	for _, row := range s.rows {
		if row.status != "queued" {
			continue
		}
		if _, ok := allowed[row.jobType]; !ok {
			continue
		}
		if row.availableAt.After(now) {
			continue
		}
		candidates = append(candidates, row)
	}
	if len(candidates) == 0 {
		return ClaimedJob{}, false, nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if !candidates[i].availableAt.Equal(candidates[j].availableAt) {
			return candidates[i].availableAt.Before(candidates[j].availableAt)
		}
		return candidates[i].createdAt.Before(candidates[j].createdAt)
	})
	row := candidates[0]
	row.status = "running"
	row.attempts++
	locked := now
	row.lockedAt = &locked
	return ClaimedJob{
		JobID:        row.id,
		JobType:      row.jobType,
		ResourceType: "",
		ResourceID:   row.resourceID,
		Payload:      append([]byte{}, row.payload...),
		Attempts:     row.attempts,
		MaxAttempts:  row.maxAttempts,
		AvailableAt:  row.availableAt,
	}, true, nil
}

func (s *fakeStore) FinalizeAsyncJob(_ context.Context, jobID string, claimedAttempts int32, outcome JobOutcome, availableAt time.Time, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[jobID]
	if !ok || row.status != "running" || row.attempts != claimedAttempts {
		return ErrStaleLease
	}
	row.lockedAt = nil
	row.errorCode = outcome.ErrorCode
	row.errorMsg = outcome.ErrorMessage
	switch {
	case outcome.Succeeded:
		row.status = "succeeded"
		completed := now
		row.completedAt = &completed
		row.errorCode = ""
		row.errorMsg = ""
	case outcome.Retryable:
		if row.attempts >= row.maxAttempts {
			row.status = "dead"
			completed := now
			row.completedAt = &completed
		} else {
			row.availableAt = availableAt
			row.status = "queued"
			row.completedAt = nil
		}
	default:
		row.status = "failed"
		completed := now
		row.completedAt = &completed
	}
	return nil
}

func (s *fakeStore) ReclaimExpiredLeases(_ context.Context, jobTypes []string, olderThan time.Time, now time.Time) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	allowed := map[string]struct{}{}
	for _, jt := range jobTypes {
		allowed[jt] = struct{}{}
	}
	var count int64
	for _, row := range s.rows {
		if row.status != "running" || row.lockedAt == nil {
			continue
		}
		if _, ok := allowed[row.jobType]; !ok {
			continue
		}
		if row.lockedAt.After(olderThan) {
			continue
		}
		row.status = "queued"
		row.lockedAt = nil
		row.availableAt = now
		count++
	}
	return count, nil
}
