ALTER TABLE resume_tailor_runs
  ADD COLUMN IF NOT EXISTS language text NOT NULL DEFAULT 'en',
  ADD COLUMN IF NOT EXISTS feature_flag text NOT NULL DEFAULT 'none',
  ADD COLUMN IF NOT EXISTS data_source_version text NOT NULL DEFAULT 'not_applicable';
