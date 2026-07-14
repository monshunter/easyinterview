-- Down migration for the D-17 JD-Match module removal.
--
-- The module is deleted from the product (product-scope v2.1 D-17). Rolling
-- back only restores the 000009 schema shells (verbatim replay) and the
-- wider async_jobs job_type check so a sequential `migrate down` through
-- 000010 / 000009 keeps working on dev databases. The 000010 registry seed
-- rows are restored by replaying 000010 in a full down/up cycle, not here.

CREATE TABLE jd_match_recommendations (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title text NOT NULL,
  company text NOT NULL,
  company_tag text,
  level text,
  location text,
  comp text,
  posted_label text,
  score smallint NOT NULL CHECK (score >= 0 AND score <= 100),
  fit jsonb NOT NULL DEFAULT '{}'::jsonb,
  reasons text[] NOT NULL DEFAULT '{}'::text[],
  risks text[] NOT NULL DEFAULT '{}'::text[],
  highlights text[] NOT NULL DEFAULT '{}'::text[],
  seen boolean NOT NULL DEFAULT false,
  dismissed_at timestamptz,
  dismiss_reason text,
  dismiss_free_note text,
  source_url text,
  source_label text,
  network_note text,
  similar_interviewers integer,
  interview_hypotheses text[] NOT NULL DEFAULT '{}'::text[],
  prompt_version text,
  rubric_version text,
  model_id text,
  language text NOT NULL DEFAULT 'zh-CN',
  feature_flag text NOT NULL DEFAULT 'none',
  data_source_version text NOT NULL DEFAULT 'jd_match.v1',
  recommended_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX idx_jd_match_recommendations_user_active ON jd_match_recommendations (user_id, score DESC, recommended_at DESC, id DESC) WHERE dismissed_at IS NULL AND deleted_at IS NULL;
CREATE INDEX idx_jd_match_recommendations_user_recommended_at ON jd_match_recommendations (user_id, recommended_at DESC);

CREATE TABLE watchlist_items (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  linked_job_match_id uuid NOT NULL REFERENCES jd_match_recommendations(id) ON DELETE CASCADE,
  label text,
  tone text NOT NULL DEFAULT 'ok' CHECK (tone IN ('ok', 'warn', 'muted')),
  change_note text,
  added_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, linked_job_match_id)
);
CREATE INDEX idx_watchlist_items_user_added_at ON watchlist_items (user_id, added_at DESC);

CREATE TABLE saved_searches (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  label text NOT NULL,
  query text NOT NULL,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  new_jobs_count integer,
  last_run_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_saved_searches_user_created_at ON saved_searches (user_id, created_at DESC);

CREATE TABLE agent_scans (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'idle' CHECK (status IN ('idle', 'scanning', 'error')),
  started_at timestamptz,
  finished_at timestamptz,
  last_scan_at timestamptz,
  next_scan_at timestamptz,
  error_message text,
  recommendation_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_agent_scans_user_created_at ON agent_scans (user_id, created_at DESC);

CREATE TABLE jd_match_search_runs (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  search_run_id uuid NOT NULL UNIQUE,
  query text NOT NULL,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  result_count integer NOT NULL DEFAULT 0,
  prompt_version text,
  rubric_version text,
  model_id text,
  data_source_version text NOT NULL DEFAULT 'jd_match.v1',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_jd_match_search_runs_user_created_at ON jd_match_search_runs (user_id, created_at DESC);

ALTER TABLE async_jobs DROP CONSTRAINT async_jobs_job_type_check;
ALTER TABLE async_jobs ADD CONSTRAINT async_jobs_job_type_check CHECK (job_type IN ('target_import', 'resume_parse', 'report_generate', 'resume_tailor', 'privacy_export', 'privacy_delete', 'email_dispatch', 'jd_match_agent_scan', 'jd_match_search'));
