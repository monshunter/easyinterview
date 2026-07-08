DROP INDEX IF EXISTS idx_target_jobs_user_resume;

ALTER TABLE target_jobs
  DROP COLUMN IF EXISTS resume_id;
