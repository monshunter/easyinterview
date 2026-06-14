-- D-20 resume flatten (product-scope v2.1, B4 D-22).
--
-- Flatten the resume version tree into a single flat `resumes` asset:
--   * rename resume_assets -> resumes and merge the structured_master
--     content (structured_profile / display_name) into it,
--   * collapse source_type to {upload, paste} (drop guided onboarding),
--   * rename practice_plans.resume_asset_id -> resume_id (interview binding
--     object: ResumeVersion -> flat resume / resumeId),
--   * drop the version-tree tables resume_versions /
--     resume_version_suggestions / resume_tailor_runs.
--
-- The structured_master -> resumes.structured_profile copy is an in-migration
-- SQL UPDATE *before* the DROP. A Go backfill registry entry cannot be used
-- here because RunBackfills (backend/internal/migrations/backfill.go) runs
-- after all SQL up migrations complete, by which point resume_versions is
-- already dropped and unreadable.

ALTER TABLE resume_assets RENAME TO resumes;

ALTER TABLE resumes ADD COLUMN IF NOT EXISTS structured_profile jsonb NOT NULL DEFAULT '{}'::jsonb;
ALTER TABLE resumes ADD COLUMN IF NOT EXISTS display_name text;

-- Backfill the structured_master content into the flat resume before the
-- version tables are dropped. Resumes without a structured_master keep the
-- default empty structured_profile / NULL display_name.
UPDATE resumes
SET structured_profile = rv.structured_profile,
    display_name = rv.display_name
FROM (
  SELECT DISTINCT ON (resume_asset_id)
    resume_asset_id, structured_profile, display_name
  FROM resume_versions
  WHERE version_type = 'structured_master' AND deleted_at IS NULL
  ORDER BY resume_asset_id, updated_at DESC
) rv
WHERE resumes.id = rv.resume_asset_id;

-- D-20 create flow is upload / paste only; drop guided onboarding answers and
-- narrow the source_type check accordingly.
UPDATE resumes
SET source_type = 'paste',
    original_text = COALESCE(original_text, raw_text, guided_answers::text)
WHERE source_type = 'guided';

ALTER TABLE resumes DROP COLUMN IF EXISTS guided_answers;
ALTER TABLE resumes DROP CONSTRAINT IF EXISTS resume_assets_source_type_check;
ALTER TABLE resumes ADD CONSTRAINT resumes_source_type_check
  CHECK (source_type IS NULL OR source_type IN ('upload', 'paste'));

-- Interview binding column rename (the FK target follows the table rename
-- above; this only renames the column).
ALTER TABLE practice_plans RENAME COLUMN resume_asset_id TO resume_id;

-- Drop the version tree (reverse FK dependency order).
DROP TABLE IF EXISTS resume_version_suggestions;
DROP TABLE IF EXISTS resume_tailor_runs;
DROP TABLE IF EXISTS resume_versions;
