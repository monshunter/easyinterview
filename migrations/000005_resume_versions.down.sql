DROP INDEX IF EXISTS idx_resume_suggestions_tailor_run;
DROP INDEX IF EXISTS idx_resume_suggestions_version_status;
DROP TABLE IF EXISTS resume_version_suggestions;

DROP INDEX IF EXISTS idx_resume_versions_parent;
DROP INDEX IF EXISTS idx_resume_versions_asset_type;
DROP INDEX IF EXISTS idx_resume_versions_user_updated;
DROP TABLE IF EXISTS resume_versions;

ALTER TABLE resume_assets
  DROP COLUMN IF EXISTS parsed_text_snapshot,
  DROP COLUMN IF EXISTS guided_answers,
  DROP COLUMN IF EXISTS original_text,
  DROP COLUMN IF EXISTS source_type;
