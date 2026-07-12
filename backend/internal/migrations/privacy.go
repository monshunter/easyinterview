package migrations

import (
	"fmt"
	"io"
	"strings"
)

// PrivacyMatrixEntry describes B4's P0 deletion disposition for a table group.
type PrivacyMatrixEntry struct {
	Tables      string
	Disposition string
}

// PrivacyMatrix is the executable projection of db-migrations-baseline spec §3.1.2.
var PrivacyMatrix = []PrivacyMatrixEntry{
	{Tables: "users", Disposition: "sync_soft_delete_then_hard_delete"},
	{Tables: "user_settings", Disposition: "hard_delete"},
	{Tables: "file_objects,resumes", Disposition: "hard_delete_and_object_storage_delete"},
	{Tables: "target_jobs,target_job_requirements,target_job_sources", Disposition: "cascade_or_hard_delete"},
	{Tables: "practice_plans,practice_sessions,practice_session_events,practice_messages", Disposition: "cascade_or_hard_delete"},
	{Tables: "idempotency_records", Disposition: "hard_delete"},
	{Tables: "feedback_reports", Disposition: "hard_delete"},
	{Tables: "source_records", Disposition: "hard_delete"},
	{Tables: "ai_task_runs", Disposition: "hard_delete_after_audit_summary"},
	{Tables: "async_jobs,outbox_events", Disposition: "hard_delete_or_redacted_terminal_tombstone"},
	{Tables: "privacy_requests", Disposition: "audit_tombstone"},
	{Tables: "audit_events", Disposition: "audit_tombstone_and_user_event_hard_delete"},
	{Tables: "auth_challenges,sessions,external_identities", Disposition: "hard_delete"},
	{Tables: "prompt_versions,rubric_versions", Disposition: "retain"},
	{Tables: "schema_migrations,schema_backfills", Disposition: "retain"},
}

// WritePrivacyMatrix writes a stable dry-run view for the backend internal privacy runner.
func WritePrivacyMatrix(w io.Writer) {
	for _, entry := range PrivacyMatrix {
		for _, table := range strings.Split(entry.Tables, ",") {
			fmt.Fprintf(w, "%s: %s\n", table, entry.Disposition)
		}
	}
}
