package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Purpose string

const (
	PurposeResume              Purpose = "resume"
	PurposeTargetJobAttachment Purpose = "target_job_attachment"
	PurposePrivacyExport       Purpose = "privacy_export"
)

type UploadStatus string

const (
	StatusPending    UploadStatus = "pending"
	StatusUploaded   UploadStatus = "uploaded"
	StatusScanFailed UploadStatus = "scan_failed"
	StatusDeleted    UploadStatus = "deleted"
)

type RetentionPolicy string

const (
	RetentionUserOwned RetentionPolicy = "user_owned"
)

var (
	ErrFileObjectNotFound     = errors.New("file object not found")
	ErrInvalidStateTransition = errors.New("invalid file object state transition")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateInput struct {
	ID               string
	UserID           string
	Purpose          Purpose
	ObjectKey        string
	OriginalFileName string
	ContentType      string
	ByteSize         int64
	Now              time.Time
}

type FileObject struct {
	ID               string
	UserID           string
	Purpose          Purpose
	ObjectKey        string
	OriginalFileName string
	ContentType      string
	ByteSize         int64
	SHA256Hex        string
	RetentionPolicy  RetentionPolicy
	Status           UploadStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

type DeletedFileObject struct {
	ID        string
	ObjectKey string
	Purpose   Purpose
}

type AuditTombstoneInput struct {
	AuditEventID string
	UserID       string
	FileObjectID string
	Purpose      Purpose
	DeletedAt    time.Time
	ObjectKey    string
}

func (r *Repository) Create(ctx context.Context, in CreateInput) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
insert into file_objects (
  id, user_id, purpose, object_key, original_file_name, content_type, byte_size,
  retention_policy, upload_status, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		in.ID,
		in.UserID,
		string(in.Purpose),
		in.ObjectKey,
		in.OriginalFileName,
		in.ContentType,
		in.ByteSize,
		string(RetentionUserOwned),
		string(StatusPending),
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert file object: %w", err)
	}
	return nil
}

func (r *Repository) MarkUploaded(ctx context.Context, fileObjectID string, now time.Time) error {
	return r.MarkStatus(ctx, fileObjectID, StatusUploaded, now)
}

func (r *Repository) MarkScanFailed(ctx context.Context, fileObjectID string, now time.Time) error {
	return r.MarkStatus(ctx, fileObjectID, StatusScanFailed, now)
}

func (r *Repository) MarkDeleted(ctx context.Context, fileObjectID string, now time.Time) error {
	return r.MarkStatus(ctx, fileObjectID, StatusDeleted, now)
}

func (r *Repository) MarkStatus(ctx context.Context, fileObjectID string, next UploadStatus, now time.Time) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin file object status update: %w", err)
	}
	defer tx.Rollback()

	var currentStr string
	err = tx.QueryRowContext(ctx, `select upload_status from file_objects where id = $1 for update`, fileObjectID).Scan(&currentStr)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrFileObjectNotFound
	}
	if err != nil {
		return fmt.Errorf("select file object status: %w", err)
	}
	current := UploadStatus(currentStr)
	if !validTransition(current, next) {
		return ErrInvalidStateTransition
	}
	res, err := tx.ExecContext(ctx, `update file_objects set upload_status = $1, updated_at = $2 where id = $3`, string(next), now, fileObjectID)
	if err != nil {
		return fmt.Errorf("update file object status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update file object status rows affected: %w", err)
	}
	if rows == 0 {
		return ErrFileObjectNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit file object status update: %w", err)
	}
	return nil
}

func (r *Repository) LockForRegister(ctx context.Context, fileObjectID, ownerUserID string, expectedPurpose Purpose) (FileObject, error) {
	if err := r.checkDB(); err != nil {
		return FileObject{}, err
	}
	row := r.db.QueryRowContext(ctx, `
select id, user_id, purpose, object_key, original_file_name, content_type, byte_size,
       sha256_hex, retention_policy, upload_status, created_at, updated_at, deleted_at
from file_objects
where id = $1 and user_id = $2 and purpose = $3 and deleted_at is null
for update`,
		fileObjectID,
		ownerUserID,
		string(expectedPurpose),
	)
	rec, err := scanFileObject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return FileObject{}, ErrFileObjectNotFound
	}
	if err != nil {
		return FileObject{}, err
	}
	return rec, nil
}

func (r *Repository) HardDelete(ctx context.Context, fileObjectID string) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	res, err := r.db.ExecContext(ctx, `delete from file_objects where id = $1`, fileObjectID)
	if err != nil {
		return fmt.Errorf("hard delete file object: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("hard delete file object rows affected: %w", err)
	}
	if rows == 0 {
		return ErrFileObjectNotFound
	}
	return nil
}

func (r *Repository) ListFileObjectsForUser(ctx context.Context, userID string) ([]DeletedFileObject, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
select id, object_key, purpose
from file_objects
where user_id = $1 and deleted_at is null
for update`, userID)
	if err != nil {
		return nil, fmt.Errorf("list file objects for user delete: %w", err)
	}
	defer rows.Close()
	var out []DeletedFileObject
	for rows.Next() {
		var rec DeletedFileObject
		var purpose string
		if err := rows.Scan(&rec.ID, &rec.ObjectKey, &purpose); err != nil {
			return nil, fmt.Errorf("scan file object for user delete list: %w", err)
		}
		rec.Purpose = Purpose(purpose)
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file objects for user delete list: %w", err)
	}
	return out, nil
}

func (r *Repository) DeleteFileObjectsForUser(ctx context.Context, userID string, _ time.Time) ([]DeletedFileObject, error) {
	files, err := r.ListFileObjectsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if err := r.HardDelete(ctx, file.ID); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func (r *Repository) InsertAuditTombstone(ctx context.Context, in AuditTombstoneInput) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	deletedAt := in.DeletedAt
	if deletedAt.IsZero() {
		deletedAt = time.Now().UTC()
	}
	metadata, err := json.Marshal(map[string]string{
		"fileObjectId": in.FileObjectID,
		"purpose":      string(in.Purpose),
		"deletedAt":    deletedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return fmt.Errorf("marshal file object audit tombstone: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type, resource_id,
  result, ip_hash, user_agent_hash, metadata, created_at
) values ($1, null, 'system', null, 'privacy.file_object_deleted', 'file_object', $2, 'success', null, null, $3, $4)`,
		in.AuditEventID,
		in.FileObjectID,
		metadata,
		deletedAt,
	)
	if err != nil {
		return fmt.Errorf("insert file object audit tombstone: %w", err)
	}
	return nil
}

func (r *Repository) checkDB() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("upload file object repository is not configured")
	}
	return nil
}

func validTransition(current, next UploadStatus) bool {
	switch next {
	case StatusUploaded:
		return current == StatusPending
	case StatusScanFailed:
		return current == StatusPending || current == StatusUploaded
	case StatusDeleted:
		return current == StatusPending || current == StatusUploaded || current == StatusScanFailed
	default:
		return false
	}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFileObject(row scanner) (FileObject, error) {
	var rec FileObject
	var purpose string
	var retentionPolicy string
	var uploadStatus string
	var sha256Hex sql.NullString
	var deletedAt sql.NullTime
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&purpose,
		&rec.ObjectKey,
		&rec.OriginalFileName,
		&rec.ContentType,
		&rec.ByteSize,
		&sha256Hex,
		&retentionPolicy,
		&uploadStatus,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&deletedAt,
	); err != nil {
		return FileObject{}, err
	}
	rec.Purpose = Purpose(purpose)
	rec.RetentionPolicy = RetentionPolicy(retentionPolicy)
	rec.Status = UploadStatus(uploadStatus)
	if sha256Hex.Valid {
		rec.SHA256Hex = sha256Hex.String
	}
	if deletedAt.Valid {
		rec.DeletedAt = &deletedAt.Time
	}
	return rec, nil
}
