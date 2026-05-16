package idempotency

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMiddlewareRejectsConcurrentPendingLock(t *testing.T) {
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	entered := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	var nextCalls atomic.Int32
	handler := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls.Add(1)
		close(entered)
		<-release
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"id": "plan-1"})
	}))

	rec1 := httptest.NewRecorder()
	go func() {
		defer close(done)
		handler.ServeHTTP(rec1, newJSONRequest("user-1", "key-1", `{"goal":"baseline"}`))
	}()

	<-entered
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, newJSONRequest("user-1", "key-1", `{"goal":"baseline"}`))
	if rec2.Code != http.StatusConflict {
		t.Fatalf("pending replay status: want %d, got %d body=%s", http.StatusConflict, rec2.Code, rec2.Body.String())
	}
	if nextCalls.Load() != 1 {
		t.Fatalf("pending request executed side effect: nextCalls=%d", nextCalls.Load())
	}

	close(release)
	<-done
	if rec1.Code != http.StatusCreated {
		t.Fatalf("first request status: want %d, got %d", http.StatusCreated, rec1.Code)
	}
}

func TestMiddlewareReplaysSucceededResponse(t *testing.T) {
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	var nextCalls atomic.Int32
	handler := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls.Add(1)
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"id": "plan-1"})
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, newJSONRequest("user-1", "key-1", `{"goal":"baseline"}`))
	if first.Code != http.StatusCreated {
		t.Fatalf("first status: want %d, got %d", http.StatusCreated, first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, newJSONRequest("user-1", "key-1", `{"goal":"baseline"}`))
	if second.Code != http.StatusCreated {
		t.Fatalf("replay status: want %d, got %d", http.StatusCreated, second.Code)
	}
	if strings.TrimSpace(second.Body.String()) != `{"id":"plan-1"}` {
		t.Fatalf("replay body mismatch: %s", second.Body.String())
	}
	if second.Header().Get(ReplayHeader) != "true" {
		t.Fatalf("expected replay header")
	}
	if nextCalls.Load() != 1 {
		t.Fatalf("success replay executed side effect: nextCalls=%d", nextCalls.Load())
	}
}

func TestMiddlewareRejectsFingerprintMismatchWithoutSideEffect(t *testing.T) {
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	var nextCalls atomic.Int32
	handler := mw.Handler("practice", "startPracticeSession", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls.Add(1)
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"id": "session-1"})
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, newJSONRequest("user-1", "same-key", `{"planId":"plan-1"}`))
	if first.Code != http.StatusCreated {
		t.Fatalf("first status: want %d, got %d", http.StatusCreated, first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, newJSONRequest("user-1", "same-key", `{"planId":"plan-2"}`))
	if second.Code != http.StatusConflict {
		t.Fatalf("mismatch status: want %d, got %d body=%s", http.StatusConflict, second.Code, second.Body.String())
	}
	if !strings.Contains(second.Body.String(), "IDEMPOTENCY_KEY_MISMATCH") {
		t.Fatalf("mismatch error code missing: %s", second.Body.String())
	}
	if strings.Contains(second.Body.String(), "session-1") {
		t.Fatalf("mismatch response leaked first resource: %s", second.Body.String())
	}
	if nextCalls.Load() != 1 {
		t.Fatalf("fingerprint mismatch executed second side effect: calls=%d", nextCalls.Load())
	}
}

func TestMiddlewareIsolatesRecordsPerUser(t *testing.T) {
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	handler := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"user": r.Header.Get("X-Test-User")})
	}))

	userOne := httptest.NewRecorder()
	handler.ServeHTTP(userOne, newJSONRequest("user-1", "shared-key", `{"goal":"baseline"}`))
	userTwo := httptest.NewRecorder()
	handler.ServeHTTP(userTwo, newJSONRequest("user-2", "shared-key", `{"goal":"baseline"}`))

	if userOne.Code != http.StatusCreated || userTwo.Code != http.StatusCreated {
		t.Fatalf("unexpected statuses: user1=%d user2=%d", userOne.Code, userTwo.Code)
	}
	if !strings.Contains(userOne.Body.String(), `"user":"user-1"`) {
		t.Fatalf("user1 body mismatch: %s", userOne.Body.String())
	}
	if !strings.Contains(userTwo.Body.String(), `"user":"user-2"`) {
		t.Fatalf("user2 body mismatch: %s", userTwo.Body.String())
	}
	if store.recordCount() != 2 {
		t.Fatalf("per-user isolation should keep two records, got %d", store.recordCount())
	}
}

func TestMiddlewareExpiresRecordsByTTL(t *testing.T) {
	current := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return current })

	responseID := "plan-1"
	handler := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"id": responseID})
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, newJSONRequest("user-1", "ttl-key", `{"goal":"baseline"}`))
	if first.Code != http.StatusCreated {
		t.Fatalf("first status: want %d, got %d", http.StatusCreated, first.Code)
	}

	current = current.Add(DefaultTTL + time.Second)
	responseID = "plan-2"
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, newJSONRequest("user-1", "ttl-key", `{"goal":"baseline"}`))
	if second.Code != http.StatusCreated {
		t.Fatalf("second status: want %d, got %d", http.StatusCreated, second.Code)
	}
	if !strings.Contains(second.Body.String(), `"id":"plan-2"`) {
		t.Fatalf("expired record should execute again, got %s", second.Body.String())
	}
}

func TestMiddlewareFinalizesNon2xxAndAllowsCorrectedSameKey(t *testing.T) {
	now := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	var nextCalls atomic.Int32
	handler := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalls.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !bytes.Contains(body, []byte(`"resumeAssetId"`)) {
			writeJSONForTest(t, w, http.StatusUnprocessableEntity, map[string]any{
				"error": map[string]any{"code": "VALIDATION_FAILED", "message": "resumeAssetId is required"},
			})
			return
		}
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"id": "plan-1"})
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, newJSONRequest("user-1", "recovery-key", `{"goal":"baseline"}`))
	if first.Code != http.StatusUnprocessableEntity {
		t.Fatalf("first status: want %d, got %d body=%s", http.StatusUnprocessableEntity, first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, newJSONRequest("user-1", "recovery-key", `{"goal":"baseline","resumeAssetId":"resume-1"}`))
	if second.Code != http.StatusCreated {
		t.Fatalf("corrected retry status: want %d, got %d body=%s", http.StatusCreated, second.Code, second.Body.String())
	}
	if nextCalls.Load() != 2 {
		t.Fatalf("corrected same-key retry should re-execute after non-2xx finalization, calls=%d", nextCalls.Load())
	}
}

func TestMiddlewareSeparatesDomainNamespace(t *testing.T) {
	now := time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
	store := newMemoryStore()
	mw := newTestMiddleware(store, func() time.Time { return now })

	practice := mw.Handler("practice", "createPracticePlan", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONForTest(t, w, http.StatusCreated, map[string]string{"domain": "practice"})
	}))
	targetJob := mw.Handler("targetjob", "importTargetJob", userFromHeader, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONForTest(t, w, http.StatusAccepted, map[string]string{"domain": "targetjob"})
	}))

	rec1 := httptest.NewRecorder()
	practice.ServeHTTP(rec1, newJSONRequest("user-1", "same-key", `{"source":"manual"}`))
	rec2 := httptest.NewRecorder()
	targetJob.ServeHTTP(rec2, newJSONRequest("user-1", "same-key", `{"source":"manual"}`))

	if rec1.Code != http.StatusCreated || rec2.Code != http.StatusAccepted {
		t.Fatalf("unexpected statuses: practice=%d targetjob=%d", rec1.Code, rec2.Code)
	}
	if store.recordCount() != 2 {
		t.Fatalf("domain namespace should keep two records, got %d", store.recordCount())
	}
}

func newTestMiddleware(store Store, now func() time.Time) *Middleware {
	seq := 0
	return New(MiddlewareOptions{
		Store: store,
		Now:   now,
		NewID: func() string {
			seq++
			return "record-" + strconv.Itoa(seq)
		},
		KeyPepper: "test-pepper",
	})
}

func userFromHeader(r *http.Request) (string, bool) {
	userID := strings.TrimSpace(r.Header.Get("X-Test-User"))
	return userID, userID != ""
}

func newJSONRequest(userID, key, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/practice/plans", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", userID)
	req.Header.Set(HeaderName, key)
	return req
}

func writeJSONForTest(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

type memoryStore struct {
	mu      sync.Mutex
	records map[string]memoryRecord
}

type memoryRecord struct {
	recordID    string
	fingerprint string
	status      Status
	expiresAt   time.Time
	response    []byte
	httpStatus  int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: map[string]memoryRecord{}}
}

func (s *memoryStore) Reserve(ctx context.Context, in ReservationInput) (Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(in)
	rec, ok := s.records[key]
	if ok && !in.Now.Before(rec.expiresAt) {
		ok = false
	}
	if !ok {
		s.records[key] = memoryRecord{
			recordID:    in.RecordID,
			fingerprint: in.RequestFingerprint,
			status:      StatusPending,
			expiresAt:   in.ExpiresAt,
		}
		return Reservation{State: StateExecute, RecordID: in.RecordID}, nil
	}
	if rec.status == StatusFailedTerminal {
		rec.fingerprint = in.RequestFingerprint
		rec.status = StatusPending
		rec.expiresAt = in.ExpiresAt
		rec.response = nil
		rec.httpStatus = 0
		s.records[key] = rec
		return Reservation{State: StateExecute, RecordID: rec.recordID}, nil
	}
	if rec.fingerprint != in.RequestFingerprint {
		return Reservation{}, ErrFingerprintMismatch
	}
	switch rec.status {
	case StatusPending:
		return Reservation{}, ErrPending
	case StatusSucceeded:
		return Reservation{
			State:          StateReplay,
			RecordID:       rec.recordID,
			ResponseBody:   append([]byte(nil), rec.response...),
			ResponseStatus: rec.httpStatus,
		}, nil
	default:
		return Reservation{}, ErrUnexpectedStatus
	}
}

func (s *memoryStore) MarkSucceeded(ctx context.Context, in CompletionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, rec := range s.records {
		if rec.recordID == in.RecordID && rec.status == StatusPending {
			rec.status = StatusSucceeded
			rec.response = append([]byte(nil), in.ResponseBody...)
			rec.httpStatus = in.ResponseStatus
			s.records[key] = rec
			return nil
		}
	}
	return ErrReservationNotFound
}

func (s *memoryStore) MarkFailed(ctx context.Context, in CompletionInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, rec := range s.records {
		if rec.recordID == in.RecordID && rec.status == StatusPending {
			rec.status = StatusFailedTerminal
			rec.response = append([]byte(nil), in.ResponseBody...)
			rec.httpStatus = in.ResponseStatus
			s.records[key] = rec
			return nil
		}
	}
	return ErrReservationNotFound
}

func (s *memoryStore) recordCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.records)
}

func (s *memoryStore) key(in ReservationInput) string {
	return strings.Join([]string{in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash}, "\x00")
}
