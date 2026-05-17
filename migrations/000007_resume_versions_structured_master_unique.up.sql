CREATE UNIQUE INDEX IF NOT EXISTS uq_resume_versions_structured_master_per_asset
  ON resume_versions (resume_asset_id)
  WHERE version_type = 'structured_master' AND deleted_at IS NULL;
