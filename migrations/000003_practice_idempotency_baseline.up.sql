CREATE TABLE IF NOT EXISTS idempotency_records (
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
CREATE INDEX IF NOT EXISTS idx_idempotency_records_expires_at ON idempotency_records (expires_at);

ALTER TABLE practice_plans
  DROP CONSTRAINT IF EXISTS practice_plans_mode_check;
ALTER TABLE practice_plans
  ADD CONSTRAINT practice_plans_mode_check CHECK (mode IN ('assisted', 'strict'));
