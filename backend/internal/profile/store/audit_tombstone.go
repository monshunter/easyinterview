package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/profile"
	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
)

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
	metadata := map[string]any{
		"experienceCardCount": in.ExperienceCardCount,
		"deletedAt":           in.DeletedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
	if strings.TrimSpace(in.JobID) != "" {
		metadata["jobId"] = in.JobID
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	if _, err := w.db.ExecContext(ctx, `
insert into audit_events (
  id, user_id, actor_type, actor_id, action, resource_type, resource_id,
  result, ip_hash, user_agent_hash, metadata, created_at
) values (
  $1, $2, 'system', null, 'profile.privacy_delete', 'candidate_profile', null,
  'success', null, null, $3::jsonb, now()
)`,
		w.newID(),
		in.UserID,
		string(raw),
	); err != nil {
		return fmt.Errorf("insert audit tombstone: %w", err)
	}
	return nil
}
