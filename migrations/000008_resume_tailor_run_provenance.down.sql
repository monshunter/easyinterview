ALTER TABLE resume_tailor_runs
  DROP COLUMN IF EXISTS data_source_version,
  DROP COLUMN IF EXISTS feature_flag,
  DROP COLUMN IF EXISTS language;
