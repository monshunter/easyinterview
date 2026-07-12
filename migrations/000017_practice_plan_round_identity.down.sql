DROP INDEX IF EXISTS idx_practice_plans_target_job_round_created;
DROP INDEX IF EXISTS idx_practice_sessions_plan_user_target;

ALTER TABLE practice_plans
  DROP CONSTRAINT IF EXISTS practice_plans_round_sequence_positive_check,
  DROP CONSTRAINT IF EXISTS practice_plans_round_identity_pair_check,
  DROP COLUMN IF EXISTS round_sequence,
  DROP COLUMN IF EXISTS round_id;
