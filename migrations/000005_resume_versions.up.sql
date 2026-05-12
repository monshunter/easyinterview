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

ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS source_type text CHECK (source_type IS NULL OR source_type IN ('upload', 'paste', 'guided'));
ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS original_text text;
ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS guided_answers jsonb;
ALTER TABLE resume_assets ADD COLUMN IF NOT EXISTS parsed_text_snapshot text;
