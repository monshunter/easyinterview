ALTER TABLE practice_plans
  DROP CONSTRAINT IF EXISTS practice_plans_mode_check;
ALTER TABLE practice_plans
  ADD CONSTRAINT practice_plans_mode_check CHECK (mode IN ('assisted', 'strict'));

DROP INDEX IF EXISTS idx_idempotency_records_expires_at;
DROP TABLE IF EXISTS idempotency_records CASCADE;
