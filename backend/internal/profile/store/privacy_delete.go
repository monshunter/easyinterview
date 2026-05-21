package store

import (
	"context"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/profile"
)

// DeleteCandidateProfileForUserWithAudit executes the full profile privacy
// delete chain atomically for the SQL repository. Success deletes cards, profile,
// and writes the success audit tombstone in one transaction. Any delete/audit
// failure rolls back the data mutation, then writes a redacted failure tombstone.
func (r *Repository) DeleteCandidateProfileForUserWithAudit(ctx context.Context, userID string, jobID string, deletedAt time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("profile store db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin profile privacy delete: %w", err)
	}

	var cardCount int64
	if err := tx.QueryRowContext(ctx, `select count(*)::bigint from experience_cards where user_id = $1`, userID).Scan(&cardCount); err != nil {
		return r.rollbackPrivacyDeleteAndAuditFailure(ctx, tx, userID, jobID, deletedAt, "count_experience_cards", err)
	}
	if _, err := tx.ExecContext(ctx, `delete from experience_cards where user_id = $1`, userID); err != nil {
		return r.rollbackPrivacyDeleteAndAuditFailure(ctx, tx, userID, jobID, deletedAt, "delete_experience_cards", err)
	}
	if _, err := tx.ExecContext(ctx, `delete from candidate_profiles where user_id = $1`, userID); err != nil {
		return r.rollbackPrivacyDeleteAndAuditFailure(ctx, tx, userID, jobID, deletedAt, "delete_candidate_profile", err)
	}
	tombstone := profile.CandidateProfileDeleteTombstone{
		UserID:              userID,
		ExperienceCardCount: cardCount,
		DeletedAt:           deletedAt,
		JobID:               jobID,
	}
	if err := writeCandidateProfileDeleteAudit(ctx, tx, r.newID, userID, "success", successTombstoneMetadata(tombstone)); err != nil {
		return r.rollbackPrivacyDeleteAndAuditFailure(ctx, tx, userID, jobID, deletedAt, "write_success_audit", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit profile privacy delete: %w", err)
	}
	return nil
}

func (r *Repository) rollbackPrivacyDeleteAndAuditFailure(ctx context.Context, tx txRollbacker, userID string, jobID string, failedAt time.Time, errorStage string, cause error) error {
	if tx != nil {
		_ = tx.Rollback()
	}
	metadata := failureTombstoneMetadata(jobID, failedAt, errorStage)
	if err := writeCandidateProfileDeleteAudit(ctx, r.db, r.newID, userID, "failure", metadata); err != nil {
		return fmt.Errorf("%s: %w; write failure audit: %v", errorStage, cause, err)
	}
	return fmt.Errorf("%s: %w", errorStage, cause)
}

type txRollbacker interface {
	Rollback() error
}
