package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/profile"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

type auditExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// AuditTombstoneWriter writes the privacy-delete tombstone row required by
// spec D-9. Only UserID + ExperienceCardCount + DeletedAt + JobID land in
// the audit_events.metadata payload (PII redline §4.3).
type AuditTombstoneWriter struct {
	db    *sql.DB
	newID func() string
}

// NewAuditTombstoneWriter wires the SQL implementation of
// profile.AuditTombstoneWriter against the supplied *sql.DB.
func NewAuditTombstoneWriter(db *sql.DB) *AuditTombstoneWriter {
	return &AuditTombstoneWriter{db: db, newID: idx.NewID}
}

// WriteCandidateProfileDeleteTombstone inserts the audit_events row.
func (w *AuditTombstoneWriter) WriteCandidateProfileDeleteTombstone(ctx context.Context, in profile.CandidateProfileDeleteTombstone) error {
	if w == nil || w.db == nil {
		return fmt.Errorf("audit tombstone writer db is nil")
	}
	metadata := successTombstoneMetadata(in)
	if err := writeCandidateProfileDeleteAudit(ctx, w.db, w.newID, in.UserID, "success", metadata); err != nil {
		return fmt.Errorf("insert audit tombstone: %w", err)
	}
	return nil
}

func successTombstoneMetadata(in profile.CandidateProfileDeleteTombstone) map[string]any {
	metadata := map[string]any{
		"experienceCardCount": in.ExperienceCardCount,
		"deletedAt":           in.DeletedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if strings.TrimSpace(in.JobID) != "" {
		metadata["jobId"] = in.JobID
	}
	return metadata
}

func failureTombstoneMetadata(jobID string, failedAt time.Time, errorStage string) map[string]any {
	metadata := map[string]any{
		"failedAt":   failedAt.UTC().Format("2006-01-02T15:04:05Z"),
		"errorStage": errorStage,
	}
	if strings.TrimSpace(jobID) != "" {
		metadata["jobId"] = jobID
	}
	return metadata
}

func writeCandidateProfileDeleteAudit(ctx context.Context, execer auditExecer, newID func() string, userID string, result string, metadata map[string]any) error {
	if execer == nil {
		return fmt.Errorf("audit execer is nil")
	}
	if newID == nil {
		newID = idx.NewID
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	if _, err := execer.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type, resource_id,
  result, ip_hash, user_agent_hash, metadata, created_at
) values (
  $1, $2, 'system', null, 'profile.privacy_delete', 'candidate_profile', null,
  $3, null, null, $4::jsonb, now()
)`,
		newID(),
		userID,
		result,
		string(raw),
	); err != nil {
		return err
	}
	return nil
}
