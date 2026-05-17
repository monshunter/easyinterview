package jobs

import (
	"context"
	"database/sql"
	stderrs "errors"
	"fmt"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	domain "github.com/monshunter/easyinterview/backend/internal/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetJob(ctx context.Context, userID, jobID string) (domain.JobRecord, error) {
	if r == nil || r.db == nil {
		return domain.JobRecord{}, fmt.Errorf("jobs repository db is nil")
	}
	var out domain.JobRecord
	var jobType string
	var resourceType string
	var status string
	var errorCode sql.NullString
	err := r.db.QueryRowContext(ctx, `
select j.id::text,
       j.job_type,
       j.resource_type,
       j.resource_id::text,
       j.status,
       j.error_code,
       j.created_at,
       j.updated_at
from async_jobs j
where j.id = $1
  and (
    (j.resource_type = 'target_job' and exists (
      select 1 from target_jobs tj
      where tj.id = j.resource_id
        and tj.user_id = $2
        and tj.deleted_at is null
    ))
    or (j.resource_type = 'resume_asset' and exists (
      select 1 from resume_assets ra
      where ra.id = j.resource_id
        and ra.user_id = $2
        and ra.deleted_at is null
    ))
    or (j.resource_type = 'feedback_report' and exists (
      select 1 from feedback_reports fr
      where fr.id = j.resource_id
        and fr.user_id = $2
    ))
    or (j.resource_type = 'resume_tailor_run' and exists (
      select 1 from resume_tailor_runs rr
      where rr.id = j.resource_id
        and rr.user_id = $2
    ))
    or (j.resource_type = 'debrief' and exists (
      select 1 from debriefs d
      where d.id = j.resource_id
        and d.user_id = $2
    ))
    or (j.resource_type = 'privacy_request' and exists (
      select 1 from privacy_requests pr
      where pr.id = j.resource_id
        and pr.user_id = $2
    ))
  )`,
		jobID,
		userID,
	).Scan(
		&out.ID,
		&jobType,
		&resourceType,
		&out.ResourceID,
		&status,
		&errorCode,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if stderrs.Is(err, sql.ErrNoRows) {
		return domain.JobRecord{}, domain.ErrJobNotFound
	}
	if err != nil {
		return domain.JobRecord{}, fmt.Errorf("get async job: %w", err)
	}
	out.JobType = api.JobType(jobType)
	out.ResourceType = api.ResourceType(resourceType)
	out.Status = sharedtypes.JobStatus(status)
	out.ErrorCode = errorCode.String
	return out, nil
}
