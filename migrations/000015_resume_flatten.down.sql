-- Down migration for D-20 resume flatten (000015).
--
-- D-4 requires the down to restore the table-structure skeleton. Rolling back
-- renames resumes -> resume_assets, restores the guided source_type check and
-- guided_answers column, renames practice_plans.resume_id -> resume_asset_id,
-- and recreates the resume_versions / resume_version_suggestions /
-- resume_tailor_runs skeletons (verbatim from 000005 / 000001) so a sequential
-- `migrate down` through 000008 / 000007 / 000005 keeps working. The flattened
-- structured_profile content is not copied back into versions (data may be
-- recorded via the backfill log; structure recovery only).

ALTER TABLE resumes RENAME TO resume_assets;

ALTER TABLE resume_assets DROP COLUMN IF EXISTS display_name;
ALTER TABLE resume_assets DROP COLUMN IF EXISTS structured_profile;

ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS guided_answers jsonb;
ALTER TABLE resume_assets DROP CONSTRAINT IF EXISTS resumes_source_type_check;
ALTER TABLE resume_assets ADD CONSTRAINT resume_assets_source_type_check
  CHECK (source_type IS NULL OR source_type IN ('upload', 'paste', 'guided'));

ALTER TABLE practice_plans RENAME COLUMN resume_id TO resume_asset_id;

-- Recreate the version-tree skeletons (FK dependency order: tailor_runs and
-- versions before suggestions, which references both).
CREATE TABLE IF NOT EXISTS resume_tailor_runs (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  resume_asset_id uuid NOT NULL REFERENCES resume_assets(id) ON DELETE CASCADE,
  mode text NOT NULL CHECK (mode IN ('gap_review', 'bullet_suggestions')),
  status text NOT NULL CHECK (status IN ('queued', 'generating', 'ready', 'failed')),
  match_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  suggestions jsonb NOT NULL DEFAULT '[]'::jsonb,
  prompt_version text,
  rubric_version text,
  model_id text,
  provider text,
  error_code text,
  generated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_resume_tailor_runs_target_job_created ON resume_tailor_runs (target_job_id, created_at DESC);

CREATE TABLE IF NOT EXISTS resume_versions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id),
  resume_asset_id uuid NOT NULL REFERENCES resume_assets(id),
  parent_version_id uuid REFERENCES resume_versions(id),
  version_type text NOT NULL CHECK (version_type IN ('structured_master', 'targeted')),
  target_job_id uuid REFERENCES target_jobs(id),
  display_name text NOT NULL,
  seed_strategy text CHECK (seed_strategy IS NULL OR seed_strategy IN ('copy_master', 'blank', 'ai_select')),
  focus_angle text,
  structured_profile jsonb NOT NULL DEFAULT '{}'::jsonb,
  match_score numeric,
  prompt_version text,
  rubric_version text,
  model_id text,
  provider text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_resume_versions_user_updated ON resume_versions (user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_resume_versions_asset_type ON resume_versions (resume_asset_id, version_type);
CREATE INDEX IF NOT EXISTS idx_resume_versions_parent ON resume_versions (parent_version_id) WHERE parent_version_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS resume_version_suggestions (
  id uuid PRIMARY KEY,
  resume_version_id uuid NOT NULL REFERENCES resume_versions(id) ON DELETE CASCADE,
  tailor_run_id uuid NOT NULL REFERENCES resume_tailor_runs(id),
  original_bullet text NOT NULL,
  suggested_bullet text NOT NULL,
  reason text,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
  decided_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_resume_suggestions_version_status ON resume_version_suggestions (resume_version_id, status);
CREATE INDEX IF NOT EXISTS idx_resume_suggestions_tailor_run ON resume_version_suggestions (tailor_run_id);
