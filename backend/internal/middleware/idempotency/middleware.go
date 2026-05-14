package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

const (
	HeaderName   = "Idempotency-Key"
	ReplayHeader = "X-Idempotency-Replay"
	DefaultTTL   = 24 * time.Hour

	defaultMaxRequestBodyBytes = 10 << 20
	resourceTypeHeader         = "X-Idempotency-Resource-Type"
	resourceIDHeader           = "X-Idempotency-Resource-ID"
)

var (
	ErrIdempotencyKeyRequired = stderrs.New("idempotency key required")
	ErrUnauthenticated        = stderrs.New("authenticated user required")
	ErrPending                = stderrs.New("idempotency record is pending")
	ErrFingerprintMismatch    = stderrs.New("idempotency request fingerprint mismatch")
	ErrReservationNotFound    = stderrs.New("idempotency reservation not found")
	ErrUnexpectedStatus       = stderrs.New("idempotency record has unexpected status")
)

type Status string

const (
	StatusPending        Status = "pending"
	StatusSucceeded      Status = "succeeded"
	StatusFailedRetry    Status = "failed_retryable"
	StatusFailedTerminal Status = "failed_terminal"
)

type ReservationState string

const (
	StateExecute ReservationState = "execute"
	StateReplay  ReservationState = "replay"
)

type Store interface {
	Reserve(ctx context.Context, in ReservationInput) (Reservation, error)
	MarkSucceeded(ctx context.Context, in CompletionInput) error
	MarkFailed(ctx context.Context, in CompletionInput) error
}

type ReservationInput struct {
	RecordID            string
	UserID              string
	Domain              string
	Operation           string
	IdempotencyKeyHash  string
	RequestFingerprint  string
	Now                 time.Time
	ExpiresAt           time.Time
	RawIdempotencyKey   string
	RequestBodyByteSize int
}

type Reservation struct {
	State          ReservationState
	RecordID       string
	ResponseStatus int
	ResponseBody   []byte
	ResourceType   string
	ResourceID     string
}

type CompletionInput struct {
	RecordID       string
	UserID         string
	Domain         string
	Operation      string
	ResponseStatus int
	ResponseBody   []byte
	ResourceType   string
	ResourceID     string
	Now            time.Time
}

type MiddlewareOptions struct {
	Store               Store
	Now                 func() time.Time
	NewID               func() string
	TTL                 time.Duration
	KeyPepper           string
	MaxRequestBodyBytes int64
}

type Middleware struct {
	store               Store
	now                 func() time.Time
	newID               func() string
	ttl                 time.Duration
	keyPepper           string
	maxRequestBodyBytes int64
}

type UserIDResolver func(r *http.Request) (userID string, ok bool)

func New(opts MiddlewareOptions) *Middleware {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	newID := opts.NewID
	if newID == nil {
		newID = idx.NewID
	}
	ttl := opts.TTL
	if ttl == 0 {
		ttl = DefaultTTL
	}
	maxBody := opts.MaxRequestBodyBytes
	if maxBody == 0 {
		maxBody = defaultMaxRequestBodyBytes
	}
	return &Middleware{
		store:               opts.Store,
		now:                 now,
		newID:               newID,
		ttl:                 ttl,
		keyPepper:           opts.KeyPepper,
		maxRequestBodyBytes: maxBody,
	}
}

func (m *Middleware) TTL() time.Duration {
	if m == nil {
		return 0
	}
	return m.ttl
}

func (m *Middleware) Handler(domain, operation string, resolveUser UserIDResolver, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m == nil || m.store == nil {
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "idempotency middleware is not configured")
			return
		}
		if resolveUser == nil {
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "idempotency user resolver is not configured")
			return
		}
		userID, ok := resolveUser(r)
		if !ok || strings.TrimSpace(userID) == "" {
			writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
			return
		}
		idempotencyKey := strings.TrimSpace(r.Header.Get(HeaderName))
		if idempotencyKey == "" {
			writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
			return
		}
		body, err := readAndRestoreBody(r, m.maxRequestBodyBytes)
		if err != nil {
			writeAPIError(w, http.StatusRequestEntityTooLarge, sharederrors.CodeValidationFailed, "request body is too large")
			return
		}

		now := m.now().UTC()
		reservation, err := m.store.Reserve(r.Context(), ReservationInput{
			RecordID:            m.newID(),
			UserID:              strings.TrimSpace(userID),
			Domain:              strings.TrimSpace(domain),
			Operation:           strings.TrimSpace(operation),
			IdempotencyKeyHash:  hashIdempotencyKey(idempotencyKey, m.keyPepper),
			RequestFingerprint:  fingerprintRequest(r.Method, r.URL.EscapedPath(), r.URL.RawQuery, body),
			Now:                 now,
			ExpiresAt:           now.Add(m.ttl),
			RawIdempotencyKey:   idempotencyKey,
			RequestBodyByteSize: len(body),
		})
		if err != nil {
			writeReservationError(w, err)
			return
		}
		if reservation.State == StateReplay {
			writeReplay(w, reservation)
			return
		}

		buffer := newBufferedResponseWriter()
		next.ServeHTTP(buffer, r)
		status := buffer.status
		if status == 0 {
			status = http.StatusOK
		}
		complete := CompletionInput{
			RecordID:       reservation.RecordID,
			UserID:         strings.TrimSpace(userID),
			Domain:         strings.TrimSpace(domain),
			Operation:      strings.TrimSpace(operation),
			ResponseStatus: status,
			ResponseBody:   buffer.body.Bytes(),
			ResourceType:   strings.TrimSpace(buffer.header.Get(resourceTypeHeader)),
			ResourceID:     strings.TrimSpace(buffer.header.Get(resourceIDHeader)),
			Now:            m.now().UTC(),
		}
		buffer.header.Del(resourceTypeHeader)
		buffer.header.Del(resourceIDHeader)
		var completeErr error
		if status >= 200 && status < 300 {
			completeErr = m.store.MarkSucceeded(r.Context(), complete)
		} else {
			completeErr = m.store.MarkFailed(r.Context(), complete)
		}
		if completeErr != nil {
			writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "idempotency response could not be persisted")
			return
		}
		buffer.flushTo(w)
	})
}

func SetResponseResource(w http.ResponseWriter, resourceType, resourceID string) {
	if w == nil {
		return
	}
	w.Header().Set(resourceTypeHeader, strings.TrimSpace(resourceType))
	w.Header().Set(resourceIDHeader, strings.TrimSpace(resourceID))
}

func readAndRestoreBody(r *http.Request, maxBytes int64) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("request body exceeds %d bytes", maxBytes)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func writeReservationError(w http.ResponseWriter, err error) {
	switch {
	case stderrs.Is(err, ErrPending):
		writeAPIError(w, http.StatusConflict, sharederrors.CodePracticeSessionConflict, "request with this Idempotency-Key is still pending")
	case stderrs.Is(err, ErrFingerprintMismatch):
		writeAPIError(w, http.StatusConflict, sharederrors.CodePracticeSessionConflict, "Idempotency-Key was already used with a different request")
	case stderrs.Is(err, ErrIdempotencyKeyRequired):
		writeAPIError(w, http.StatusBadRequest, sharederrors.CodeValidationFailed, "Idempotency-Key header is required")
	case stderrs.Is(err, ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, sharederrors.CodeAuthUnauthorized, "authentication required")
	default:
		writeAPIError(w, http.StatusInternalServerError, sharederrors.CodeValidationFailed, "idempotency reservation failed")
	}
}

func writeReplay(w http.ResponseWriter, reservation Reservation) {
	status := reservation.ResponseStatus
	if status == 0 {
		status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(ReplayHeader, "true")
	w.WriteHeader(status)
	_, _ = w.Write(reservation.ResponseBody)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	meta := sharederrors.CodeRegistry[code]
	raw, _ := json.Marshal(map[string]any{
		"error": sharederrors.APIError{
			Code:      code,
			Message:   message,
			RequestID: "",
			Retryable: meta.Retryable,
		},
	})
	_, _ = w.Write(raw)
}

func hashIdempotencyKey(key, pepper string) string {
	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(key)))
	if pepper != "" {
		h.Write([]byte("|"))
		h.Write([]byte(pepper))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func HashKey(key, pepper string) string {
	return hashIdempotencyKey(key, pepper)
}

func fingerprintRequest(method, path, rawQuery string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(strings.ToUpper(strings.TrimSpace(method))))
	h.Write([]byte("\n"))
	h.Write([]byte(path))
	h.Write([]byte("\n"))
	h.Write([]byte(rawQuery))
	h.Write([]byte("\n"))
	h.Write(bytes.TrimSpace(body))
	return hex.EncodeToString(h.Sum(nil))
}

func Fingerprint(method, path, rawQuery string, body []byte) string {
	return fingerprintRequest(method, path, rawQuery, body)
}

type bufferedResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{header: make(http.Header)}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) flushTo(dst http.ResponseWriter) {
	for key, values := range w.header {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}
	status := w.status
	if status == 0 {
		status = http.StatusOK
	}
	dst.WriteHeader(status)
	_, _ = dst.Write(w.body.Bytes())
}
