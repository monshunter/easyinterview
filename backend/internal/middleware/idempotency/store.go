package idempotency

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"strings"
	"time"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) Reserve(ctx context.Context, in ReservationInput) (Reservation, error) {
	if err := s.checkDB(); err != nil {
		return Reservation{}, err
	}
	if err := validateReservationInput(in); err != nil {
		return Reservation{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Reservation{}, fmt.Errorf("begin idempotency reservation: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, reservationLockKey(in)); err != nil {
		return Reservation{}, fmt.Errorf("lock idempotency reservation: %w", err)
	}

	rec, hit, err := selectReservation(ctx, tx, in)
	if err != nil {
		return Reservation{}, err
	}
	if !hit {
		if err := insertPendingReservation(ctx, tx, in); err != nil {
			return Reservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return Reservation{}, fmt.Errorf("commit idempotency reservation insert: %w", err)
		}
		return Reservation{State: StateExecute, RecordID: in.RecordID}, nil
	}

	if !in.Now.Before(rec.expiresAt) {
		if err := resetPendingReservation(ctx, tx, rec.id, in); err != nil {
			return Reservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return Reservation{}, fmt.Errorf("commit expired idempotency reservation: %w", err)
		}
		return Reservation{State: StateExecute, RecordID: rec.id}, nil
	}

	if rec.status == StatusFailedTerminal {
		if err := resetPendingReservation(ctx, tx, rec.id, in); err != nil {
			return Reservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return Reservation{}, fmt.Errorf("commit failed idempotency reservation reset: %w", err)
		}
		return Reservation{State: StateExecute, RecordID: rec.id}, nil
	}

	if rec.fingerprint != in.RequestFingerprint {
		return Reservation{}, ErrFingerprintMismatch
	}

	switch rec.status {
	case StatusPending:
		return Reservation{}, ErrPending
	case StatusSucceeded:
		status, body := decodeStoredResponse(rec.responseBody)
		if err := tx.Commit(); err != nil {
			return Reservation{}, fmt.Errorf("commit idempotency replay: %w", err)
		}
		return Reservation{
			State:          StateReplay,
			RecordID:       rec.id,
			ResponseStatus: status,
			ResponseBody:   body,
			ResourceType:   rec.resourceType,
			ResourceID:     rec.resourceID,
		}, nil
	case StatusFailedRetry:
		if err := resetPendingReservation(ctx, tx, rec.id, in); err != nil {
			return Reservation{}, err
		}
		if err := tx.Commit(); err != nil {
			return Reservation{}, fmt.Errorf("commit retryable idempotency reservation: %w", err)
		}
		return Reservation{State: StateExecute, RecordID: rec.id}, nil
	default:
		return Reservation{}, ErrUnexpectedStatus
	}
}

func (s *SQLStore) MarkSucceeded(ctx context.Context, in CompletionInput) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	if strings.TrimSpace(in.RecordID) == "" || strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.Domain) == "" || strings.TrimSpace(in.Operation) == "" {
		return fmt.Errorf("complete idempotency reservation requires recordId, userId, domain, operation")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	responseBody, err := encodeStoredResponse(in.ResponseStatus, in.ResponseBody)
	if err != nil {
		return fmt.Errorf("encode idempotency response: %w", err)
	}
	res, err := s.db.ExecContext(ctx, `
update idempotency_records
set status = $1,
    response_body = $2,
    resource_type = $3,
    resource_id = $4,
    error_code = null,
    updated_at = $5
where id = $6
  and user_id = $7
  and domain = $8
  and operation = $9`,
		string(StatusSucceeded),
		responseBody,
		nullableString(in.ResourceType),
		nullableUUID(in.ResourceID),
		now,
		in.RecordID,
		in.UserID,
		in.Domain,
		in.Operation,
	)
	if err != nil {
		return fmt.Errorf("mark idempotency reservation succeeded: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark idempotency reservation rows affected: %w", err)
	}
	if rows == 0 {
		return ErrReservationNotFound
	}
	return nil
}

func (s *SQLStore) MarkFailed(ctx context.Context, in CompletionInput) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	if strings.TrimSpace(in.RecordID) == "" || strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.Domain) == "" || strings.TrimSpace(in.Operation) == "" {
		return fmt.Errorf("complete failed idempotency reservation requires recordId, userId, domain, operation")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	responseBody, err := encodeStoredResponse(in.ResponseStatus, in.ResponseBody)
	if err != nil {
		return fmt.Errorf("encode failed idempotency response: %w", err)
	}
	res, err := s.db.ExecContext(ctx, `
update idempotency_records
set status = $1,
    response_body = $2,
    resource_type = null,
    resource_id = null,
    error_code = $3,
    updated_at = $4
where id = $5
  and user_id = $6
  and domain = $7
  and operation = $8`,
		string(StatusFailedTerminal),
		responseBody,
		nullableString(errorCodeFromResponseBody(in.ResponseBody)),
		now,
		in.RecordID,
		in.UserID,
		in.Domain,
		in.Operation,
	)
	if err != nil {
		return fmt.Errorf("mark idempotency reservation failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("mark idempotency reservation failed rows affected: %w", err)
	}
	if rows == 0 {
		return ErrReservationNotFound
	}
	return nil
}

func (s *SQLStore) checkDB() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("idempotency SQL store is not configured")
	}
	return nil
}

type selectedReservation struct {
	id           string
	fingerprint  string
	status       Status
	responseBody []byte
	resourceType string
	resourceID   string
	expiresAt    time.Time
}

func selectReservation(ctx context.Context, tx *sql.Tx, in ReservationInput) (selectedReservation, bool, error) {
	var rec selectedReservation
	var status string
	var responseBody sql.NullString
	var resourceType sql.NullString
	var resourceID sql.NullString
	err := tx.QueryRowContext(ctx, `
select id, request_fingerprint, status, response_body, resource_type, resource_id, expires_at
from idempotency_records
where user_id = $1
  and domain = $2
  and operation = $3
  and idempotency_key_hash = $4
for update`,
		in.UserID,
		in.Domain,
		in.Operation,
		in.IdempotencyKeyHash,
	).Scan(
		&rec.id,
		&rec.fingerprint,
		&status,
		&responseBody,
		&resourceType,
		&resourceID,
		&rec.expiresAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return selectedReservation{}, false, nil
	}
	if err != nil {
		return selectedReservation{}, false, fmt.Errorf("select idempotency reservation: %w", err)
	}
	rec.status = Status(status)
	if responseBody.Valid {
		rec.responseBody = []byte(responseBody.String)
	}
	if resourceType.Valid {
		rec.resourceType = resourceType.String
	}
	if resourceID.Valid {
		rec.resourceID = resourceID.String
	}
	return rec, true, nil
}

func insertPendingReservation(ctx context.Context, tx *sql.Tx, in ReservationInput) error {
	_, err := tx.ExecContext(ctx, `
insert into idempotency_records (
  id, user_id, domain, operation, idempotency_key_hash,
  request_fingerprint, status, resource_type, resource_id, response_body,
  error_code, expires_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,null,null,null,null,$8,$9,$9)`,
		in.RecordID,
		in.UserID,
		in.Domain,
		in.Operation,
		in.IdempotencyKeyHash,
		in.RequestFingerprint,
		string(StatusPending),
		in.ExpiresAt,
		in.Now,
	)
	if err != nil {
		return fmt.Errorf("insert idempotency reservation: %w", err)
	}
	return nil
}

func resetPendingReservation(ctx context.Context, tx *sql.Tx, recordID string, in ReservationInput) error {
	_, err := tx.ExecContext(ctx, `
update idempotency_records
set request_fingerprint = $1,
    status = $2,
    resource_type = null,
    resource_id = null,
    response_body = null,
    error_code = null,
    expires_at = $3,
    updated_at = $4
where id = $5`,
		in.RequestFingerprint,
		string(StatusPending),
		in.ExpiresAt,
		in.Now,
		recordID,
	)
	if err != nil {
		return fmt.Errorf("reset idempotency reservation: %w", err)
	}
	return nil
}

func validateReservationInput(in ReservationInput) error {
	if strings.TrimSpace(in.RecordID) == "" ||
		strings.TrimSpace(in.UserID) == "" ||
		strings.TrimSpace(in.Domain) == "" ||
		strings.TrimSpace(in.Operation) == "" ||
		strings.TrimSpace(in.IdempotencyKeyHash) == "" ||
		strings.TrimSpace(in.RequestFingerprint) == "" {
		return fmt.Errorf("idempotency reservation requires recordId, userId, domain, operation, key hash, fingerprint")
	}
	if in.Now.IsZero() || in.ExpiresAt.IsZero() {
		return fmt.Errorf("idempotency reservation requires now and expiresAt")
	}
	return nil
}

func reservationLockKey(in ReservationInput) string {
	return strings.Join([]string{in.UserID, in.Domain, in.Operation, in.IdempotencyKeyHash}, "|")
}

type storedResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

func encodeStoredResponse(status int, body []byte) ([]byte, error) {
	if status == 0 {
		status = httpStatusOK
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		trimmed = []byte("null")
	}
	if !json.Valid(trimmed) {
		quoted, err := json.Marshal(string(body))
		if err != nil {
			return nil, err
		}
		trimmed = quoted
	}
	return json.Marshal(storedResponse{Status: status, Body: json.RawMessage(trimmed)})
}

func decodeStoredResponse(raw []byte) (int, []byte) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return httpStatusOK, nil
	}
	var resp storedResponse
	if err := json.Unmarshal(raw, &resp); err == nil && resp.Status > 0 && len(resp.Body) > 0 {
		return resp.Status, append([]byte(nil), resp.Body...)
	}
	return httpStatusOK, append([]byte(nil), raw...)
}

func errorCodeFromResponseBody(body []byte) string {
	var decoded struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(body), &decoded); err != nil {
		return ""
	}
	return strings.TrimSpace(decoded.Error.Code)
}

const httpStatusOK = 200

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func nullableUUID(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
