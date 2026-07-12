ALTER TABLE practice_plans
  ADD COLUMN round_id text,
  ADD COLUMN round_sequence integer,
  ADD CONSTRAINT practice_plans_round_identity_pair_check CHECK (
    (round_id IS NULL AND round_sequence IS NULL)
    OR (
      round_id IS NOT NULL
      AND btrim(round_id) <> ''
      AND round_sequence IS NOT NULL
    )
  ),
  ADD CONSTRAINT practice_plans_round_sequence_positive_check CHECK (
    round_sequence IS NULL OR round_sequence > 0
  );

CREATE INDEX idx_practice_plans_target_job_round_created
  ON practice_plans (user_id, target_job_id, status, round_sequence, round_id, created_at DESC, id DESC)
  WHERE round_id IS NOT NULL AND round_sequence IS NOT NULL;

CREATE INDEX idx_practice_sessions_plan_user_target
  ON practice_sessions (plan_id, user_id, target_job_id);
