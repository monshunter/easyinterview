ALTER TABLE practice_plans
  DROP CONSTRAINT IF EXISTS practice_plans_mode_check;
ALTER TABLE practice_plans
  ADD CONSTRAINT practice_plans_mode_check CHECK (mode IN ('assisted', 'strict'));

-- idempotency_records and idx_idempotency_records_expires_at are owned by
-- 000001_create_baseline after the pre-launch baseline rebase.
