CREATE TABLE schema_backfills (
  id bigserial PRIMARY KEY,
  version integer NOT NULL,
  name text NOT NULL,
  mode text NOT NULL CHECK (mode IN ('dry_run', 'apply')),
  status text NOT NULL CHECK (status IN ('running', 'succeeded', 'failed', 'skipped')),
  checksum text NOT NULL,
  started_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  error_message text,
  UNIQUE (version, mode, checksum, status)
);

CREATE TABLE users (
  id uuid PRIMARY KEY,
  email text NOT NULL UNIQUE,
  display_name text,
  auth_provider text NOT NULL DEFAULT 'passwordless',
  auth_provider_user_id text,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled', 'deleted')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE user_settings (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  ui_language text NOT NULL DEFAULT 'zh-CN',
  preferred_practice_language text NOT NULL DEFAULT 'en',
  region text,
  timezone text NOT NULL DEFAULT 'UTC',
  analytics_opt_in boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE file_objects (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose text NOT NULL CHECK (purpose IN ('resume', 'target_job_attachment', 'privacy_export', 'source_snapshot', 'audio', 'video')),
  object_key text NOT NULL UNIQUE,
  original_file_name text NOT NULL,
  content_type text NOT NULL,
  byte_size bigint NOT NULL,
  sha256_hex text,
  retention_policy text NOT NULL DEFAULT 'user_owned' CHECK (retention_policy IN ('user_owned', 'short_lived', 'legal_hold')),
  upload_status text NOT NULL DEFAULT 'pending' CHECK (upload_status IN ('pending', 'uploaded', 'scan_failed', 'deleted')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX idx_file_objects_user_created ON file_objects (user_id, created_at DESC);

CREATE TABLE resume_assets (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  file_object_id uuid REFERENCES file_objects(id) ON DELETE SET NULL,
  title text NOT NULL,
  language text NOT NULL DEFAULT 'en',
  parse_status text NOT NULL DEFAULT 'queued' CHECK (parse_status IN ('queued', 'processing', 'ready', 'failed')),
  parsed_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  raw_text text,
  latest_parse_job_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX idx_resume_assets_user_updated_at ON resume_assets (user_id, updated_at DESC);

CREATE TABLE target_jobs (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'preparing', 'applied', 'interviewing', 'offer', 'rejected', 'archived')),
  analysis_status text NOT NULL DEFAULT 'queued' CHECK (analysis_status IN ('queued', 'processing', 'ready', 'failed')),
  title text,
  company_name text,
  location_text text,
  employment_type text,
  seniority_level text,
  target_language text NOT NULL DEFAULT 'en',
  source_type text NOT NULL CHECK (source_type IN ('manual_text', 'url', 'file', 'manual_form')),
  source_url text,
  source_file_object_id uuid REFERENCES file_objects(id) ON DELETE SET NULL,
  raw_jd_text text,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  fit_summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  notes text,
  latest_report_id uuid,
  open_question_issue_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);
CREATE INDEX idx_target_jobs_user_status_updated ON target_jobs (user_id, status, updated_at DESC);
CREATE INDEX idx_target_jobs_user_analysis_updated ON target_jobs (user_id, analysis_status, updated_at DESC);
CREATE INDEX idx_target_jobs_fts ON target_jobs USING gin (to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(company_name, '')));

CREATE TABLE target_job_requirements (
  id uuid PRIMARY KEY,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('must_have', 'nice_to_have', 'hidden_signal', 'interview_focus')),
  label text NOT NULL,
  description text,
  evidence_level text NOT NULL DEFAULT 'explicit' CHECK (evidence_level IN ('explicit', 'inferred')),
  display_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_target_job_requirements_target_job ON target_job_requirements (target_job_id, display_order);

CREATE TABLE target_job_sources (
  id uuid PRIMARY KEY,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  source_type text NOT NULL CHECK (source_type IN ('url', 'file', 'manual_text', 'manual_form')),
  url text,
  file_object_id uuid REFERENCES file_objects(id) ON DELETE SET NULL,
  snapshot_text text,
  fetched_at timestamptz,
  freshness_status text NOT NULL DEFAULT 'fresh' CHECK (freshness_status IN ('fresh', 'stale', 'expired')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_target_job_sources_target_job ON target_job_sources (target_job_id, created_at DESC);

CREATE TABLE practice_plans (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  source_report_id uuid,
  goal text NOT NULL CHECK (goal IN ('baseline', 'retry_current_round', 'next_round')),
  interviewer_persona text NOT NULL CHECK (interviewer_persona IN ('generalist', 'hr', 'hiring_manager', 'technical_manager', 'peer')),
  difficulty text NOT NULL DEFAULT 'standard' CHECK (difficulty IN ('easy', 'standard', 'stretch')),
  language text NOT NULL DEFAULT 'en',
  time_budget_minutes smallint NOT NULL,
  resume_asset_id uuid REFERENCES resume_assets(id) ON DELETE SET NULL,
  focus_competency_codes text[] NOT NULL DEFAULT '{}'::text[],
  status text NOT NULL DEFAULT 'ready' CHECK (status IN ('draft', 'ready', 'archived')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT practice_plans_source_goal_check CHECK (
    (goal = 'baseline' AND source_report_id IS NULL)
    OR (goal IN ('retry_current_round', 'next_round') AND source_report_id IS NOT NULL)
  )
);
CREATE INDEX idx_practice_plans_target_job_created ON practice_plans (target_job_id, created_at DESC);

CREATE TABLE idempotency_records (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  domain text NOT NULL,
  operation text NOT NULL,
  idempotency_key_hash text NOT NULL,
  request_fingerprint text NOT NULL,
  status text NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed_retryable', 'failed_terminal')),
  resource_type text,
  resource_id uuid,
  response_body jsonb,
  error_code text,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, domain, operation, idempotency_key_hash)
);
CREATE INDEX idx_idempotency_records_expires_at ON idempotency_records (expires_at);

CREATE TABLE practice_sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  plan_id uuid NOT NULL REFERENCES practice_plans(id) ON DELETE CASCADE,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('queued', 'running', 'waiting_user_input', 'completing', 'completed', 'failed', 'cancelled')),
  language text NOT NULL DEFAULT 'en',
  started_at timestamptz,
  completed_at timestamptz,
  cancelled_at timestamptz,
  failure_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_practice_sessions_target_job_created ON practice_sessions (target_job_id, created_at DESC);
CREATE INDEX idx_practice_sessions_user_status_updated ON practice_sessions (user_id, status, updated_at DESC);

CREATE TABLE practice_session_events (
  id uuid PRIMARY KEY,
  session_id uuid NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  seq_no integer NOT NULL,
  event_type text NOT NULL CHECK (event_type IN ('session_started', 'session_completed')),
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (session_id, seq_no)
);
CREATE INDEX idx_practice_session_events_session_seq ON practice_session_events (session_id, seq_no);

CREATE TABLE practice_messages (
  id uuid PRIMARY KEY,
  session_id uuid NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  seq_no integer NOT NULL,
  role text NOT NULL CHECK (role IN ('user', 'assistant')),
  content text NOT NULL,
  client_message_id uuid,
  reply_to_message_id uuid REFERENCES practice_messages(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (session_id, seq_no),
  UNIQUE (session_id, client_message_id),
  UNIQUE (reply_to_message_id),
  CONSTRAINT practice_messages_client_id_check CHECK (
    (role = 'user' AND client_message_id IS NOT NULL AND reply_to_message_id IS NULL)
    OR (role = 'assistant' AND client_message_id IS NULL)
  )
);
CREATE INDEX idx_practice_messages_session_seq ON practice_messages (session_id, seq_no);

CREATE TABLE feedback_reports (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_id uuid NOT NULL REFERENCES practice_sessions(id) ON DELETE CASCADE,
  target_job_id uuid NOT NULL REFERENCES target_jobs(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('queued', 'generating', 'ready', 'failed')),
  preparedness_level text CHECK (preparedness_level IN ('not_ready', 'needs_practice', 'basically_ready', 'well_prepared')),
  highlights jsonb NOT NULL DEFAULT '[]'::jsonb,
  issues jsonb NOT NULL DEFAULT '[]'::jsonb,
  next_actions jsonb NOT NULL DEFAULT '[]'::jsonb,
  prompt_version text,
  rubric_version text,
  model_id text,
  provider text,
  language text NOT NULL DEFAULT 'en',
  feature_flag text NOT NULL DEFAULT 'none',
  data_source_version text NOT NULL DEFAULT 'not_applicable',
  dimension_assessments jsonb NOT NULL DEFAULT '[]'::jsonb,
  retry_focus_competency_codes text[] NOT NULL DEFAULT '{}'::text[],
  error_code text,
  generated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_feedback_reports_session_unique ON feedback_reports (session_id);
CREATE INDEX idx_feedback_reports_target_job_created ON feedback_reports (target_job_id, created_at DESC);

ALTER TABLE practice_plans
  ADD CONSTRAINT fk_practice_plans_source_report FOREIGN KEY (source_report_id) REFERENCES feedback_reports(id) ON DELETE SET NULL;

CREATE TABLE resume_tailor_runs (
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
CREATE INDEX idx_resume_tailor_runs_target_job_created ON resume_tailor_runs (target_job_id, created_at DESC);

CREATE TABLE source_records (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  owner_type text NOT NULL CHECK (owner_type IN ('target_job', 'intelligence_item')),
  owner_id uuid NOT NULL,
  source_type text NOT NULL CHECK (source_type IN ('jd_url', 'company_page', 'manual_text', 'news', 'upload')),
  title text,
  url text,
  summary jsonb NOT NULL DEFAULT '{}'::jsonb,
  snapshot_file_object_id uuid REFERENCES file_objects(id) ON DELETE SET NULL,
  fetched_at timestamptz,
  expires_at timestamptz,
  freshness_status text NOT NULL DEFAULT 'fresh' CHECK (freshness_status IN ('fresh', 'stale', 'expired')),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_source_records_owner ON source_records (owner_type, owner_id, created_at DESC);

CREATE TABLE prompt_versions (
  id uuid PRIMARY KEY,
  feature_key text NOT NULL,
  version text NOT NULL,
  language text NOT NULL DEFAULT 'multi',
  template_hash text NOT NULL,
  template_body text NOT NULL,
  is_active boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (feature_key, version, language)
);

CREATE TABLE rubric_versions (
  id uuid PRIMARY KEY,
  feature_key text NOT NULL,
  version text NOT NULL,
  language text NOT NULL DEFAULT 'multi',
  schema_json jsonb NOT NULL,
  is_active boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (feature_key, version, language)
);

CREATE TABLE ai_task_runs (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  task_type text NOT NULL CHECK (task_type IN ('jd_parse', 'resume_parse', 'practice_chat', 'report_generate', 'resume_tailor')),
  resource_type text NOT NULL,
  resource_id uuid NOT NULL,
  provider text NOT NULL,
  model_family text,
  model_id text NOT NULL,
  prompt_version text,
  rubric_version text,
  model_profile_name text,
  model_profile_version text,
  feature_key text NOT NULL,
  feature_flag text NOT NULL DEFAULT 'none',
  data_source_version text NOT NULL DEFAULT 'not_applicable',
  language text NOT NULL DEFAULT 'en',
  input_tokens integer NOT NULL DEFAULT 0,
  output_tokens integer NOT NULL DEFAULT 0,
  latency_ms integer NOT NULL DEFAULT 0,
  cost_usd_micros bigint NOT NULL DEFAULT 0,
  status text NOT NULL CHECK (status IN ('success', 'failed', 'timeout', 'fallback')),
  route text,
  validation_status text,
  output_schema_version text,
  error_code text,
  fallback_chain jsonb NOT NULL DEFAULT '[]'::jsonb,
  raw_response_object_key text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  started_at timestamptz NOT NULL,
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_task_runs_resource ON ai_task_runs (resource_type, resource_id, created_at DESC);
CREATE INDEX idx_ai_task_runs_task_started ON ai_task_runs (task_type, started_at DESC);
CREATE INDEX idx_ai_task_runs_dashboard ON ai_task_runs (model_profile_name, validation_status, created_at DESC);

CREATE TABLE async_jobs (
  id uuid PRIMARY KEY,
  job_type text NOT NULL CHECK (job_type IN ('target_import', 'resume_parse', 'report_generate', 'resume_tailor', 'source_refresh', 'privacy_export', 'privacy_delete', 'email_dispatch')),
  resource_type text NOT NULL,
  resource_id uuid NOT NULL,
  dedupe_key text,
  status text NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed', 'cancelled', 'dead')),
  attempts integer NOT NULL DEFAULT 0,
  max_attempts integer NOT NULL DEFAULT 5,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  result jsonb NOT NULL DEFAULT '{}'::jsonb,
  error_code text,
  error_message text,
  available_at timestamptz NOT NULL DEFAULT now(),
  locked_at timestamptz,
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_async_jobs_active_dedupe ON async_jobs (job_type, dedupe_key) WHERE dedupe_key IS NOT NULL AND status IN ('queued', 'running');
CREATE INDEX idx_async_jobs_status_available ON async_jobs (status, available_at);

CREATE TABLE outbox_events (
  id uuid PRIMARY KEY,
  event_name text NOT NULL,
  event_version integer NOT NULL DEFAULT 1,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  payload jsonb NOT NULL,
  publish_status text NOT NULL DEFAULT 'pending' CHECK (publish_status IN ('pending', 'published', 'failed')),
  publish_attempts integer NOT NULL DEFAULT 0,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  locked_at timestamptz,
  last_error_code text,
  last_error_message text,
  published_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_outbox_events_publish_status_created ON outbox_events (publish_status, created_at);
CREATE INDEX idx_outbox_events_pending_due ON outbox_events (publish_status, next_attempt_at, created_at);

CREATE TABLE privacy_requests (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  request_type text NOT NULL CHECK (request_type IN ('export', 'delete')),
  status text NOT NULL CHECK (status IN ('queued', 'processing', 'completed', 'failed', 'cancelled')),
  export_file_object_id uuid REFERENCES file_objects(id) ON DELETE SET NULL,
  requested_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  error_code text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);
CREATE INDEX idx_privacy_requests_user_requested ON privacy_requests (user_id, requested_at DESC);

CREATE TABLE audit_events (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  actor_type text NOT NULL CHECK (actor_type IN ('user', 'system', 'admin')),
  actor_id uuid,
  action text NOT NULL,
  resource_type text NOT NULL,
  resource_id uuid,
  result text NOT NULL CHECK (result IN ('success', 'failure')),
  ip_hash text,
  user_agent_hash text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_events_user_created ON audit_events (user_id, created_at DESC);
CREATE INDEX idx_audit_events_action_created ON audit_events (action, created_at DESC);

CREATE TABLE auth_challenges (
  id uuid PRIMARY KEY,
  user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  email text NOT NULL,
  challenge_token_hash text NOT NULL,
  purpose text NOT NULL DEFAULT 'login' CHECK (purpose IN ('login', 'signup')),
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'consumed', 'expired', 'revoked')),
  ip_hash text,
  user_agent_hash text,
  expires_at timestamptz NOT NULL,
  consumed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_auth_challenges_email_status ON auth_challenges (email, status, created_at DESC);
CREATE INDEX idx_auth_challenges_token_hash ON auth_challenges (challenge_token_hash);

CREATE TABLE sessions (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  session_hash text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
  ip_hash text,
  user_agent_hash text,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user_status ON sessions (user_id, status, updated_at DESC);

CREATE TABLE external_identities (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider text NOT NULL,
  provider_subject_hash text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  linked_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_subject_hash)
);
CREATE INDEX idx_external_identities_user ON external_identities (user_id, linked_at DESC);
