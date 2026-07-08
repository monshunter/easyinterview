ALTER TABLE target_jobs
  ADD COLUMN IF NOT EXISTS resume_id uuid REFERENCES resumes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_target_jobs_user_resume
  ON target_jobs (user_id, resume_id)
  WHERE resume_id IS NOT NULL AND deleted_at IS NULL;
